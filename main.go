package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	local, _ := GetVersion()

	official, _ := GetLatestVersion()

	officialTargetOs := strings.Replace(local.Os, "/", "-", 1)

	if local.Version == official {
		fmt.Println("It is now the latest version.")
	}

	fileSuffix := ".zip"

	downloadUrl := "https://dl.google.com/go/" + local.Version + "." + officialTargetOs + fileSuffix

	installDir, _ := GetInstallDir()

	Clean(installDir)

	filePath := installDir + `\` + local.Version + fileSuffix

	DownloadFile(downloadUrl, filePath)

	UnzipWithPrefix(filePath, "go", installDir)

	DeleteFile(filePath)
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
