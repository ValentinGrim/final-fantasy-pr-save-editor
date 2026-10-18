package io

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"pixel-remastered-save-editor/global"
)

func gameInfo(game global.Game) (folder, appID string) {
	switch game {
	case global.One:
		return one, "1173770"
	case global.Two:
		return two, "1173780"
	case global.Three:
		return three, "1173790"
	case global.Four:
		return four, "1173800"
	case global.Five:
		return five, "1173810"
	default: // six
		return six, "1173820"
	}
}

// defaultSaveDir looks for <Documents>/My Games/<game>/Steam/<SteamID64> and returns "" if nothing is found.
func defaultSaveDir(game global.Game) string {
	var (
		folder, appID = gameInfo(game)
		seen          = make(map[string]bool)
		fallback      string
		best          string
		bestSave      bool
		bestTime      time.Time
	)
	for _, docs := range documentsDirs(appID) {
		steam := filepath.Join(docs, "My Games", folder, "Steam")
		resolved, err := filepath.EvalSymlinks(steam)
		if err != nil || seen[resolved] {
			continue
		}
		seen[resolved] = true
		entries, err := os.ReadDir(steam)
		if err != nil {
			continue
		}
		if fallback == "" {
			fallback = steam
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dir := filepath.Join(steam, e.Name())
			hasSave, mod := scanSaveDir(dir)
			if best == "" || (hasSave && !bestSave) || (hasSave == bestSave && mod.After(bestTime)) {
				best, bestSave, bestTime = dir, hasSave, mod
			}
		}
	}
	if best != "" {
		return best
	}
	return fallback
}

func scanSaveDir(dir string) (hasSave bool, latest time.Time) {
	if fi, err := os.Stat(dir); err == nil {
		latest = fi.ModTime()
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !isSaveFile(e.Name()) {
			continue
		}
		hasSave = true
		if fi, err := e.Info(); err == nil && fi.ModTime().After(latest) {
			latest = fi.ModTime()
		}
	}
	return
}

func isSaveFile(name string) bool {
	for _, s := range saves {
		if name == s.UUID || name == fmt.Sprintf("slot%d.sav", s.Slot) {
			return true
		}
	}
	return false
}
