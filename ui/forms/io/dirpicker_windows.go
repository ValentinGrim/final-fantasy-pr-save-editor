//go:build windows

package io

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"github.com/sqweek/dialog"
)

// pickDirectory lets the user choose a directory with the native Windows folder browser, starting in dir, and
// calls onPicked with the choice; it is not called when the user cancels.
func pickDirectory(_ fyne.Window, title, dir string, onPicked func(string)) {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	if picked, err := dialog.Directory().Title(title).SetStartDir(dir).Browse(); err == nil && picked != "" {
		onPicked(picked)
	}
}
