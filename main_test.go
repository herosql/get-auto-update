package main

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version, err := GetVersion()

	if err != nil {
		t.Fatalf("The expectation was correct, but what I received was: %v", err)
	}

	if !strings.HasPrefix(version.Version, "go") {
		t.Errorf("The output format is incorrect, resulting in: %s", version.Version)
	}

	if version.Os != "windows/amd64" {
		t.Errorf("The output format is incorrect, resulting in: %s", version.Os)
	}

}

func TestGetLatestVersion(t *testing.T) {
	version, err := GetLatestVersion()

	if err != nil {
		t.Fatalf("The expectation was correct, but what I received was: %v", err)
	}

	if !strings.HasPrefix(version, "go") {
		t.Errorf("The output format is incorrect, resulting in: %s", version)
	}

}

func TestGetInstallDir(t *testing.T) {
	dir, err := GetInstallDir()

	if err != nil {
		t.Fatalf("The expectation was correct, but what I received was: %v", err)
	}

	if dir != `D:\apply\go` {
		t.Errorf("The output format is incorrect, resulting in: %s", dir)
	}

}
