package file

import "testing"

func TestGetInstallDir(t *testing.T) {
	dir, err := GetInstallDir()

	if err != nil {
		t.Fatalf("The expectation was correct, but what I received was: %v", err)
	}

	if dir != `D:\apply\go` {
		t.Errorf("The output format is incorrect, resulting in: %s", dir)
	}

}
