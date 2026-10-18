package io

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/global"
)

// newTestSelect builds the widget without NewFileIO, which would write the user's config.
func newTestSelect(kind Kind, dir string, picked *string) *saveSelect {
	return &saveSelect{
		kind:    kind,
		game:    global.Six,
		dir:     binding.BindString(&dir),
		buttons: container.NewVBox(),
		onSelected: func(_ global.Game, _, dir, _ string, _ int, _ global.SaveFileType) {
			*picked = dir
		},
	}
}

func TestSaveButtonsFollowDirectory(t *testing.T) {
	test.NewTempApp(t)
	withSave, empty := t.TempDir(), t.TempDir()
	// One of these matches whatever save type the user's config selects.
	for _, n := range []string{saves[2].UUID, "slot1.sav"} {
		if err := os.WriteFile(filepath.Join(withSave, n), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var picked string
	w := newTestSelect(Save, withSave, &picked)
	w.DataChanged()
	if got := len(w.buttons.Objects); got != len(saves) {
		t.Fatalf("save buttons in %s: got %d, want %d", withSave, got, len(saves))
	}

	_ = w.dir.Set(empty)
	w.DataChanged()
	if got := len(w.buttons.Objects); got != len(saves) {
		t.Fatalf("save buttons in empty dir: got %d, want %d", got, len(saves))
	}
	for _, o := range w.buttons.Objects {
		picked = ""
		o.(*widget.Button).OnTapped()
		if picked != empty {
			t.Fatalf("button %q saves into %q, want %q", o.(*widget.Button).Text, picked, empty)
		}
	}

	_ = w.dir.Set(filepath.Join(empty, "missing"))
	w.DataChanged()
	if got := len(w.buttons.Objects); got != 0 {
		t.Fatalf("save buttons in missing dir: got %d, want 0", got)
	}
}

func TestLoadButtonsFollowDirectory(t *testing.T) {
	test.NewTempApp(t)
	withSave, empty := t.TempDir(), t.TempDir()
	for _, n := range []string{saves[2].UUID, "slot1.sav"} {
		if err := os.WriteFile(filepath.Join(withSave, n), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var picked string
	w := newTestSelect(Load, withSave, &picked)
	w.DataChanged()
	if got := len(w.buttons.Objects); got != 1 {
		t.Fatalf("load buttons in %s: got %d, want 1", withSave, got)
	}

	_ = w.dir.Set(empty)
	w.DataChanged()
	if got := len(w.buttons.Objects); got != 0 {
		t.Fatalf("load buttons in empty dir: got %d, want 0", got)
	}
}
