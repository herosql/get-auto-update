package internal

import (
	"github.com/herosql/get-auto-update/internal/installer"
	"github.com/herosql/get-auto-update/internal/version"
	"github.com/herosql/get-auto-update/pkg/file"
	"github.com/herosql/get-auto-update/pkg/log"
)

func Update() error {
	local, err := version.GetVersion()

	if err != nil {
		log.Error("get version err：%s", err)
		return err
	}

	log.Info("local current version:%s", local.Version)

	official, err := version.GetLatestVersion()
	if err != nil {
		log.Error("GetLatestVersion err：%s", err)
		return err
	}

	log.Info("latest release:%s\n", official)

	if local.Version == official {
		log.Info("It is now the latest version.")
		return err
	}

	installDir, err := file.GetInstallDir()
	if err != nil {
		log.Error("install dir get error:%s", err)
		return err
	}

	log.Info("clean up old versions.")
	err = file.Clean(installDir)
	if err != nil {
		log.Error("clean up old versions error:%s", err)
		return err
	}

	local.Version = official
	local.InstallDir = installDir

	install := installer.New(local)

	log.Info("download the latest version.")
	filePath, err := install.Download()

	if err != nil {
		log.Error("download error:%s", err)
		return err
	}

	log.Info("extract the latest version.")
	err = install.Extract(filePath)

	if err != nil {
		log.Error("extract error:%s", err)
		return err
	}

	log.Info("clean up downloaded files.")
	err = file.DeleteFile(filePath)

	if err != nil {
		log.Error("deleteFile error:%s", err)
		return err
	}

	return nil
}
