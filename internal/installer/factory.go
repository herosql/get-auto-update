package installer

import (
	"strings"

	"github.com/herosql/get-auto-update/internal/version"
)

func New(version *version.Version) Installer {
	if strings.Contains(version.Os, "windows") {
		return &WindowsInstaller{version: version}
	}
	return &UnixInstaller{version: version}
}
