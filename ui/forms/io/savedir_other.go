//go:build !windows

package io

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var vdfPath = regexp.MustCompile(`^\s*"path"\s+"(.+)"\s*$`)

// documentsDirs returns the Proton prefix Documents folder of the game in every Steam library.
func documentsDirs(appID string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var dirs []string
	for _, lib := range steamLibraries(home) {
		dirs = append(dirs, filepath.Join(lib, "steamapps", "compatdata", appID, "pfx", "drive_c", "users", "steamuser", "Documents"))
	}
	return dirs
}

func steamLibraries(home string) []string {
	var libs []string
	seen := make(map[string]bool)
	add := func(p string) {
		if r, err := filepath.EvalSymlinks(p); err == nil && !seen[r] {
			seen[r] = true
			libs = append(libs, r)
		}
	}
	for _, root := range []string{
		filepath.Join(home, ".local", "share", "Steam"),
		filepath.Join(home, ".steam", "steam"),
		filepath.Join(home, ".steam", "root"),
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".local", "share", "Steam"),
	} {
		add(root)
		f, err := os.Open(filepath.Join(root, "steamapps", "libraryfolders.vdf"))
		if err != nil {
			continue
		}
		s := bufio.NewScanner(f)
		for s.Scan() {
			if m := vdfPath.FindStringSubmatch(s.Text()); m != nil {
				add(strings.ReplaceAll(m[1], `\\`, `\`))
			}
		}
		_ = f.Close()
	}
	return libs
}
