package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"pixel-remastered-save-editor/global"
)

const (
	file   = "prSaveEditor.config"
	appDir = "ffpr-save-editor"
)

var (
	config ConfigData
	path   string
)

type ConfigData struct {
	SaveDir1          string `json:"dir1"`
	SaveDir2          string `json:"dir2"`
	SaveDir3          string `json:"dir3"`
	SaveDir4          string `json:"dir4"`
	SaveDir5          string `json:"dir5"`
	SaveDir6          string `json:"dir6"`
	EnablePlayStation bool   `json:"ps"`
}

func init() {
	path = resolvePath()
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &config)
	}
}

// resolvePath returns where the config lives. A legacy config next to the
// working dir or the executable wins (portable mode), otherwise the per-user
// config dir is used, falling back to the working dir.
func resolvePath() string {
	for _, d := range legacyDirs() {
		p := filepath.Join(d, file)
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	if d, err := os.UserConfigDir(); err == nil && d != "" {
		return filepath.Join(d, appDir, file)
	}
	return filepath.Join(global.PWD, file)
}

func legacyDirs() []string {
	dirs := []string{global.PWD}
	if exe, err := os.Executable(); err == nil {
		if r, err := filepath.EvalSymlinks(exe); err == nil {
			exe = r
		}
		if d := filepath.Dir(exe); d != global.PWD {
			dirs = append(dirs, d)
		}
	}
	return dirs
}

// Path returns the config file location.
func Path() string { return path }

func saveConfig() {
	b, err := json.Marshal(&config)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, b, 0o644)
}

func Dir(saveType global.Game) (dir string) {
	switch saveType {
	case global.One:
		dir = config.SaveDir1
	case global.Two:
		dir = config.SaveDir2
	case global.Three:
		dir = config.SaveDir3
	case global.Four:
		dir = config.SaveDir4
	case global.Five:
		dir = config.SaveDir5
	case global.Six:
		dir = config.SaveDir6
	}
	if dir == "" {
		dir = "."
	}
	return
}

func SetSaveDir(saveType global.Game, dir string) {
	switch saveType {
	case global.One:
		config.SaveDir1 = dir
	case global.Two:
		config.SaveDir2 = dir
	case global.Three:
		config.SaveDir3 = dir
	case global.Four:
		config.SaveDir4 = dir
	case global.Five:
		config.SaveDir5 = dir
	default: // global.Six
		config.SaveDir6 = dir
	}
	saveConfig()
}

func EnablePlayStation() bool { return config.EnablePlayStation }

func SetEnablePlayStation(v bool) {
	config.EnablePlayStation = v
	saveConfig()
}
