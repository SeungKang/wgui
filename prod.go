//go:build !dev

package main

import (
	"os"
	"path/filepath"
	"runtime"
)

func wguExePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	wguPath := filepath.Join(filepath.Dir(exePath), "wgu")

	if runtime.GOOS == "windows" {
		wguPath += ".exe"
	}

	return wguPath, nil
}
