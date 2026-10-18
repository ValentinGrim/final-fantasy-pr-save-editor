package io

import (
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save/file"
)

// useBackupRoot keeps the backups, and the config dir, away from the user's.
func useBackupRoot(t *testing.T) string {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	root := t.TempDir()
	t.Cleanup(file.SetBackupRoot(root))
	return root
}

func writeBackup(t *testing.T, folder string, files map[string]string, manifest map[string]any) {
	t.Helper()
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	for n, c := range files {
		if err := os.WriteFile(filepath.Join(folder, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if manifest != nil {
		b, _ := json.Marshal(manifest)
		if err := os.WriteFile(filepath.Join(folder, "backup.json"), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// newTestRestore builds the page in a 725x800 test window. Opening a folder or
// looking for the Steam save dir is never done for real.
func newTestRestore(t *testing.T, game global.Game, fallbackDir string, opened *string, back *bool) (*restorePage, fyne.Window) {
	t.Helper()
	w := test.NewTempWindow(t, nil)
	p := newRestorePage(game, w, func() { *back = true })
	p.saveDir = func(global.Game) string { return fallbackDir }
	p.openFolder = func(dir string) error { *opened = dir; return nil }
	p.reload()
	w.SetContent(p)
	w.Resize(fyne.NewSize(global.WindowWidth, global.WindowHeight))
	return p, w
}

// capture writes the window to $FFPR_CAPTURE_DIR/<name>.png when set.
func capture(t *testing.T, w fyne.Window, name string) {
	t.Helper()
	dir := os.Getenv("FFPR_CAPTURE_DIR")
	if dir == "" {
		return
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = png.Encode(f, w.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}

func assertFits(t *testing.T, w fyne.Window) {
	t.Helper()
	if s := w.Content().MinSize(); s.Width > global.WindowWidth || s.Height > global.WindowHeight {
		t.Fatalf("page needs %v, more than the %dx%d window", s, global.WindowWidth, global.WindowHeight)
	}
	if s := w.Canvas().Size(); s.Width > global.WindowWidth || s.Height > global.WindowHeight {
		t.Fatalf("window grew to %v", s)
	}
}

// objects returns o and everything it renders.
func objects(o fyne.CanvasObject) []fyne.CanvasObject {
	all := []fyne.CanvasObject{o}
	switch c := o.(type) {
	case *fyne.Container:
		for _, child := range c.Objects {
			all = append(all, objects(child)...)
		}
	case fyne.Widget:
		for _, child := range test.WidgetRenderer(c).Objects() {
			all = append(all, objects(child)...)
		}
	}
	return all
}

func findButton(t *testing.T, o fyne.CanvasObject, text string) *widget.Button {
	t.Helper()
	for _, c := range objects(o) {
		if b, ok := c.(*widget.Button); ok && b.Text == text {
			return b
		}
	}
	t.Fatalf("no %q button", text)
	return nil
}

func topDialogText(w fyne.Window) string {
	top := w.Canvas().Overlays().Top()
	if top == nil {
		return ""
	}
	var texts []string
	for _, c := range objects(top) {
		switch l := c.(type) {
		case *widget.Label:
			texts = append(texts, l.Text)
		}
	}
	return strings.Join(texts, "\n")
}

func TestRestorePage(t *testing.T) {
	a := test.NewTempApp(t)
	a.Settings().SetTheme(theme.DefaultTheme())
	root := useBackupRoot(t)
	target, fallback := t.TempDir(), t.TempDir()
	slot1, slot2, auto := saves[2].UUID, saves[3].UUID, saves[0].UUID
	gameDir := filepath.Join(root, "FF6")
	writeBackup(t, filepath.Join(gameDir, "20261017-090000"), // before manifests
		map[string]string{slot2: "old slot 2"}, nil)
	writeBackup(t, filepath.Join(gameDir, "20261018-163001"),
		map[string]string{slot1: "slot 1 v1", file.BestiaryFile: "bestiary v1"},
		map[string]any{"version": 1, "game": 6, "sourceDir": target, "files": []string{slot1, file.BestiaryFile}, "saveType": "PC", "reason": "save"})
	writeBackup(t, filepath.Join(gameDir, "20261018-163001-02"),
		map[string]string{auto: "auto", "unknown.bin": "?"},
		map[string]any{"version": 1, "game": 6, "sourceDir": "/no/such/dir/" + strings.Repeat("very-long-folder-name/", 12), "files": []string{auto, "unknown.bin"}, "saveType": "PC", "reason": "restore"})
	for n, c := range map[string]string{slot1: "slot 1 current", file.BestiaryFile: "bestiary current", "other": "untouched"} {
		if err := os.WriteFile(filepath.Join(target, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var (
		opened string
		back   bool
	)
	p, w := newTestRestore(t, global.Six, fallback, &opened, &back)
	assertFits(t, w)
	capture(t, w, "restore-list")

	if len(p.backups) != 3 || p.backups[0].Name != "20261018-163001-02" || p.backups[2].Name != "20261017-090000" {
		t.Fatalf("backups = %+v", p.backups)
	}
	if !p.details.Hidden || !p.restoreBtn.Disabled() || !p.empty.Hidden {
		t.Fatal("nothing selected: no details, Restore disabled, no empty text")
	}

	// A backup whose source dir is gone, and one without manifest, use the save dir.
	p.list.Select(0)
	if p.dir.Text != fallback {
		t.Fatalf("restore to %q, want the save dir %q", p.dir.Text, fallback)
	}
	p.list.Select(2)
	if p.dir.Text != fallback {
		t.Fatalf("restore to %q, want the save dir %q", p.dir.Text, fallback)
	}

	p.list.Select(1)
	if p.dir.Text != target || p.details.Hidden || p.restoreBtn.Disabled() {
		t.Fatalf("selected: dir %q, details hidden %v", p.dir.Text, p.details.Hidden)
	}
	if !strings.Contains(p.summary.Text, "slot 1 (+ Bestiary)") {
		t.Fatalf("summary = %q", p.summary.Text)
	}
	assertFits(t, w)
	capture(t, w, "restore-selected")

	test.Tap(p.restoreBtn)
	msg := topDialogText(w)
	if !strings.Contains(msg, "This replaces 2 file(s) in "+target+". The current files are backed up first, so you can undo this restore.") {
		t.Fatalf("confirm dialog = %q", msg)
	}
	capture(t, w, "restore-confirm")
	test.Tap(findButton(t, w.Canvas().Overlays().Top(), "Restore"))

	if msg = topDialogText(w); !strings.Contains(msg, "Restored. Make sure the game is closed") {
		t.Fatalf("after restoring, dialog = %q", msg)
	}
	capture(t, w, "restore-done")
	for n, want := range map[string]string{slot1: "slot 1 v1", file.BestiaryFile: "bestiary v1", "other": "untouched"} {
		if got, _ := os.ReadFile(filepath.Join(target, n)); string(got) != want {
			t.Fatalf("%s = %q, want %q", n, got, want)
		}
	}
	// The list now starts with the undo backup holding the replaced files.
	if len(p.backups) != 4 || p.backups[0].Reason != file.ReasonRestore || p.backups[0].SourceDir != target {
		t.Fatalf("backups after restoring = %+v", p.backups)
	}
	if got, _ := os.ReadFile(filepath.Join(gameDir, p.backups[0].Name, slot1)); string(got) != "slot 1 current" {
		t.Fatalf("undo backup holds %q", got)
	}
	if !p.details.Hidden || !p.restoreBtn.Disabled() {
		t.Fatal("the selection is cleared after restoring")
	}

	test.Tap(findButton(t, p, "Open backups folder"))
	if opened != gameDir {
		t.Fatalf("opened %q, want %q", opened, gameDir)
	}
	test.Tap(findButton(t, p, "Back"))
	if !back {
		t.Fatal("Back not called")
	}
}

func TestRestorePageErrors(t *testing.T) {
	a := test.NewTempApp(t)
	a.Settings().SetTheme(theme.DefaultTheme())
	root := useBackupRoot(t)
	writeBackup(t, filepath.Join(root, "FF2", "20261018-163001"), map[string]string{saves[2].UUID: "x"}, nil)
	var (
		opened string
		back   bool
	)
	missing := filepath.Join(t.TempDir(), "missing")
	p, w := newTestRestore(t, global.Two, missing, &opened, &back)
	p.list.Select(0)
	test.Tap(p.restoreBtn)
	test.Tap(findButton(t, w.Canvas().Overlays().Top(), "Restore"))
	if msg := topDialogText(w); !strings.Contains(msg, "annot restore to "+missing+": the directory does not exist") {
		t.Fatalf("error dialog = %q", msg)
	}
	if _, err := os.Stat(missing); err == nil {
		t.Fatal("target dir created")
	}
}

func TestRestorePageEmpty(t *testing.T) {
	a := test.NewTempApp(t)
	a.Settings().SetTheme(theme.DefaultTheme())
	useBackupRoot(t)
	var (
		opened string
		back   bool
	)
	p, w := newTestRestore(t, global.One, t.TempDir(), &opened, &back)
	if p.empty.Hidden || !p.list.Hidden || p.empty.Text != "No backups yet for FF1. A backup is made each time the editor overwrites a save." {
		t.Fatalf("empty state: %q hidden %v", p.empty.Text, p.empty.Hidden)
	}
	assertFits(t, w)
	capture(t, w, "restore-empty")
}

func TestDescribeBackupFiles(t *testing.T) {
	for _, tc := range []struct {
		files []string
		want  string
	}{
		{[]string{saves[2].UUID, file.BestiaryFile}, "slot 1 (+ Bestiary)"},
		{[]string{saves[0].UUID}, "Auto Save"},
		{[]string{"slot3.sav"}, "slot 3"},
		{[]string{file.BestiaryFile}, "Bestiary"},
		{[]string{"weird", saves[1].UUID}, "Quick Save, weird"},
	} {
		if got := describeBackupFiles(tc.files); got != tc.want {
			t.Errorf("describeBackupFiles(%v) = %q, want %q", tc.files, got, tc.want)
		}
	}
}
