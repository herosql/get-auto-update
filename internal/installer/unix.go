package installer

import (
	"fmt"
	"strings"

	"github.com/herosql/get-auto-update/internal/version"
	"github.com/herosql/get-auto-update/pkg/download"
	"github.com/herosql/get-auto-update/pkg/uncompress"
)

type UnixInstaller struct {
	version *version.Version
}

func (u *UnixInstaller) Download() (string, error) {
	officialTargetOs := strings.Replace(u.version.Os, "/", "-", 1)

	fileSuffix := ".tar.gz"

	downloadUrl := "https://dl.google.com/go/" + u.version.Version + "." + officialTargetOs + fileSuffix

	filePath := u.version.InstallDir + `\` + u.version.Version + fileSuffix

	err := download.DownloadFile(downloadUrl, filePath)

	if err != nil {
		return "", fmt.Errorf("unix download error: %w", err)
	}

	return filePath, nil
}

func (u *UnixInstaller) Extract(src string) error {
	err := uncompress.UncompressTarGz(src, "go", u.version.InstallDir)
	if err != nil {
		return fmt.Errorf("UncompressTarGz error: %w", err)
	}
	return err
}
