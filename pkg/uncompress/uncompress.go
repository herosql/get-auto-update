package uncompress

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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
