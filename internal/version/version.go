package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

type Version struct {
	Version    string `json:"version"`
	Stable     bool   `json:"stable"`
	Os         string
	InstallDir string
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

func GetVersion() (*Version, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.CombinedOutput()
	localVersion := &Version{}
	if err != nil {
		return nil, err
	}

	parts := strings.Fields(string(output))
	localVersion.Version = parts[2]
	localVersion.Os = parts[3]
	return localVersion, nil
}
