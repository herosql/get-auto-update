package installer

import (
	"fmt"
	"strings"

	"github.com/herosql/get-auto-update/internal/version"
	"github.com/herosql/get-auto-update/pkg/download"
	"github.com/herosql/get-auto-update/pkg/uncompress"
)

type WindowsInstaller struct {
	version *version.Version
}

func (w *WindowsInstaller) Download() (string, error) {
	officialTargetOs := strings.Replace(w.version.Os, "/", "-", 1)

	fileSuffix := ".zip"

	downloadUrl := "https://dl.google.com/go/" + w.version.Version + "." + officialTargetOs + fileSuffix

	filePath := w.version.InstallDir + `\` + w.version.Version + fileSuffix

	err := download.DownloadFile(downloadUrl, filePath)

	if err != nil {
		return "", fmt.Errorf("unix download error: %w", err)
	}

	return filePath, nil
}

func (w *WindowsInstaller) Extract(src string) error {

	err := uncompress.UnzipWithPrefix(src, "go", w.version.InstallDir)

	if err != nil {
		return fmt.Errorf("UnzipWithPrefix error: %w", err)
	}
	return nil
}
