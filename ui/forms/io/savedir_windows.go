//go:build windows

package io

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func documentsDirs(string) []string {
	var dirs []string
	if p, err := windows.KnownFolderPath(windows.FOLDERID_Documents, 0); err == nil {
		dirs = append(dirs, p)
	}
	if p := os.Getenv("USERPROFILE"); p != "" {
		dirs = append(dirs, filepath.Join(p, "Documents"), filepath.Join(p, "OneDrive", "Documents"))
	}
	if p := os.Getenv("OneDrive"); p != "" {
		dirs = append(dirs, filepath.Join(p, "Documents"))
	}
	return dirs
}
