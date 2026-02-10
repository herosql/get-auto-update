package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func LogAndPrint(level slog.Level, msg string, args ...any) {
	slog.Default().Log(context.Background(), level, msg, args...)

	if level >= slog.LevelInfo {
		if level == slog.LevelError {
			LogAndPrint(slog.LevelError, "Error: %s\n", msg)
		} else {
			fmt.Printf("➜ %s\n", msg)
		}
	}
}

func main() {
	local, err := GetVersion()

	if err != nil {
		LogAndPrint(slog.LevelError, "get version err：%s\n", err)
		os.Exit(1)
	}

	LogAndPrint(slog.LevelInfo, "local current version:%s\n", local.Version)

	official, err := GetLatestVersion()
	if err != nil {
		LogAndPrint(slog.LevelError, "GetLatestVersion err：%s\n", err)
	}

	LogAndPrint(slog.LevelInfo, "latest release:%s\n", official)

	officialTargetOs := strings.Replace(local.Os, "/", "-", 1)

	if local.Version == official {
		LogAndPrint(slog.LevelInfo, "It is now the latest version.")
		os.Exit(1)
	}

	fileSuffix := ".zip"

	isWindows := strings.Contains(local.Os, "windows")

	if !isWindows {
		fileSuffix = ".tar.gz"
	}

	downloadUrl := "https://dl.google.com/go/" + local.Version + "." + officialTargetOs + fileSuffix

	installDir, _ := GetInstallDir()

	LogAndPrint(slog.LevelInfo, "clean up old versions.\n")

	err = Clean(installDir)

	if err != nil {
		LogAndPrint(slog.LevelError, "clean error：%s\n", err)
		os.Exit(1)
	}

	filePath := installDir + `\` + local.Version + fileSuffix

	LogAndPrint(slog.LevelInfo, "download the latest version.\n")

	err = DownloadFile(downloadUrl, filePath)

	if err != nil {
		LogAndPrint(slog.LevelError, "donload error:%s\n", err)
		os.Exit(1)
	}

	if !isWindows {
		err = UncompressTarGz(filePath, "go", installDir)
		if err != nil {
			LogAndPrint(slog.LevelError, "UncompressTarGz error:%s\n", err)
			os.Exit(1)
		}
	} else {
		err = UnzipWithPrefix(filePath, "go", installDir)

		if err != nil {
			LogAndPrint(slog.LevelError, "UnzipWithPrefix error:%s\n", err)
			os.Exit(1)
		}
	}

	LogAndPrint(slog.LevelInfo, "clean up downloaded files.\n")
	err = DeleteFile(filePath)

	if err != nil {
		LogAndPrint(slog.LevelError, "DeleteFile error:%s\n", err)
		os.Exit(1)
	}
}

func UncompressTarGz(srcFilePath, targetFolder, dest string) error {
	f, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	targetFolder = filepath.ToSlash(targetFolder)
	if !strings.HasSuffix(targetFolder, "/") {
		targetFolder += "/"
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		headerName := filepath.ToSlash(header.Name)
		if !strings.HasPrefix(headerName, targetFolder) {
			continue
		}

		relPath := strings.TrimPrefix(headerName, targetFolder)
		if relPath == "" {
			continue
		}

		targetPath := filepath.Join(dest, relPath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

func UnzipWithPrefix(zipPath, targetFolder, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	targetFolder = filepath.ToSlash(targetFolder)
	if !strings.HasSuffix(targetFolder, "/") {
		targetFolder += "/"
	}

	for _, f := range r.File {
		fName := filepath.ToSlash(f.Name)
		if !strings.HasPrefix(fName, targetFolder) {
			continue
		}

		relPath := strings.TrimPrefix(fName, targetFolder)
		if relPath == "" {
			continue
		}

		fpath := filepath.Join(dest, relPath)

		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}

func DownloadFile(url string, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Download failed, status code: %d", resp.StatusCode)
	}

	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func Clean(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		err := os.RemoveAll(fullPath)
		if err != nil {
			return err
		}
	}
	return nil
}

type Version struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

func GetInstallDir() (string, error) {
	cmd := exec.Command("go", "env", "GOROOT")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func GetLatestVersion() (string, error) {
	resp, err := http.Get("https://go.dev/dl/?mode=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var versions []Version

	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}

	for _, v := range versions {
		if v.Stable {
			return v.Version, nil
		}
	}
	return "", fmt.Errorf("No stable version found")
}

type LocalVersion struct {
	Version string
	Os      string
}

func GetVersion() (*LocalVersion, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.CombinedOutput()
	localVersion := &LocalVersion{}
	if err != nil {
		return nil, err
	}

	parts := strings.Fields(string(output))
	localVersion.Version = parts[2]
	localVersion.Os = parts[3]
	return localVersion, nil
}
