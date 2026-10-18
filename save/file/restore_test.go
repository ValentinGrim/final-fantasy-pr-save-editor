package file

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"pixel-remastered-save-editor/global"
)

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for n, c := range files {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func readManifest(t *testing.T, folder string) manifest {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(folder, manifestName))
	if err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err = json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestBackupManifest(t *testing.T) {
	t.Setenv("PR_PRINT", "false")
	root := useBackupRoot(t)
	useClock(t, time.Date(2026, 10, 18, 16, 30, 0, 0, time.Local))
	dir, _, _ := pcSaveDir(t)
	data, err := LoadSave(global.Two, dir, "slot", global.PC)
	if err != nil {
		t.Fatal(err)
	}
	changeBestiary(t, data)
	if err = SaveSave(global.Two, data, 1, dir, "slot", global.PC); err != nil {
		t.Fatal(err)
	}
	m := readManifest(t, filepath.Join(root, "FF2", "20261018-163001"))
	created, err := time.Parse(time.RFC3339, m.Created)
	if m.Version != 1 || m.Game != 2 || m.SourceDir != dir || m.SaveType != "PC" || m.Reason != ReasonSave ||
		err != nil || !created.Equal(time.Date(2026, 10, 18, 16, 30, 1, 0, time.Local)) ||
		!slices.Equal(m.Files, []string{"slot", bestiaryFile}) {
		t.Fatalf("manifest = %+v (%v)", m, err)
	}

	backups, err := ListBackups(global.Two)
	if err != nil || len(backups) != 1 {
		t.Fatalf("ListBackups = %v, %v", backups, err)
	}
	b := backups[0]
	if !slices.Equal(b.Files, []string{"dp3fS2vqP7GDj8eF72YKqbT7FIAF=e7Shy2CsTITm2E=", "slot"}) ||
		b.SourceDir != dir || !b.SaveTypeKnown || b.SaveType != global.PC || b.Reason != ReasonSave {
		t.Fatalf("backup = %+v", b)
	}
}

func TestListBackups(t *testing.T) {
	root := useBackupRoot(t)
	if b, err := ListBackups(global.Five); err != nil || b != nil {
		t.Fatalf("no backup dir: %v, %v", b, err)
	}
	gameDir := filepath.Join(root, "FF5")
	writeFiles(t, filepath.Join(gameDir, "20261017-090000"), map[string]string{"old": "1"}) // no manifest
	writeFiles(t, filepath.Join(gameDir, "20261018-120000"), map[string]string{"a": "2"})
	writeFiles(t, filepath.Join(gameDir, "20261018-120000-02"), map[string]string{"slot1.sav": "3"})
	if err := writeManifest(filepath.Join(gameDir, "20261018-120000-02"), manifest{
		Version: 1, Game: 5, SourceDir: "/somewhere", Files: []string{"slot1.sav"}, SaveType: "PS", Reason: ReasonRestore,
	}); err != nil {
		t.Fatal(err)
	}
	// Not backups: ignored.
	writeFiles(t, filepath.Join(gameDir, "keep-me"), map[string]string{"x": "x"})
	writeFiles(t, gameDir, map[string]string{"20261019-000000": "a file"})
	if err := os.Mkdir(filepath.Join(gameDir, "20261019-100000"), 0o755); err != nil { // empty
		t.Fatal(err)
	}
	writeFiles(t, filepath.Join(gameDir, "20261019-110000"), nil)
	if err := writeManifest(filepath.Join(gameDir, "20261019-110000"), manifest{Version: 1}); err != nil { // manifest only
		t.Fatal(err)
	}

	backups, err := ListBackups(global.Five)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, b := range backups {
		names = append(names, b.Name)
	}
	if strings.Join(names, " ") != "20261018-120000-02 20261018-120000 20261017-090000" {
		t.Fatalf("backups = %v", names)
	}
	if b := backups[0]; !b.Time.Equal(time.Date(2026, 10, 18, 12, 0, 0, 0, time.Local)) ||
		!slices.Equal(b.Files, []string{"slot1.sav"}) || b.SourceDir != "/somewhere" ||
		!b.SaveTypeKnown || b.SaveType != global.PS || b.Reason != ReasonRestore {
		t.Fatalf("backup with suffix = %+v", b)
	}
	if b := backups[2]; !b.Time.Equal(time.Date(2026, 10, 17, 9, 0, 0, 0, time.Local)) ||
		!slices.Equal(b.Files, []string{"old"}) || b.SourceDir != "" || b.SaveTypeKnown || b.Reason != "" {
		t.Fatalf("backup without manifest = %+v", b)
	}
}

func TestRestoreBackup(t *testing.T) {
	root := useBackupRoot(t)
	useClock(t, time.Date(2026, 10, 18, 18, 0, 0, 0, time.Local))
	gameDir := filepath.Join(root, "FF6")
	// A backup without manifest holding a slot and the bestiary.
	writeFiles(t, filepath.Join(gameDir, "20261018-100000"), map[string]string{"slot": "old slot", bestiaryFile: "old bestiary"})
	target := t.TempDir()
	writeFiles(t, target, map[string]string{"slot": "new slot", "other": "untouched"})

	backups, err := ListBackups(global.Six)
	if err != nil || len(backups) != 1 {
		t.Fatalf("ListBackups = %v, %v", backups, err)
	}
	if err = RestoreBackup(global.Six, backups[0], target); err != nil {
		t.Fatal(err)
	}
	for n, want := range map[string]string{"slot": "old slot", bestiaryFile: "old bestiary", "other": "untouched"} {
		if got, _ := os.ReadFile(filepath.Join(target, n)); string(got) != want {
			t.Fatalf("%s = %q, want %q", n, got, want)
		}
	}
	assertNoTempFiles(t, target)

	// The overwritten slot was backed up first; the bestiary did not exist.
	backups, err = ListBackups(global.Six)
	if err != nil || len(backups) != 2 {
		t.Fatalf("ListBackups = %v, %v", backups, err)
	}
	undo := backups[0]
	if undo.Name != "20261018-180001" || undo.Reason != ReasonRestore || undo.SourceDir != target ||
		!slices.Equal(undo.Files, []string{"slot"}) {
		t.Fatalf("undo backup = %+v", undo)
	}
	if got, _ := os.ReadFile(filepath.Join(gameDir, undo.Name, "slot")); string(got) != "new slot" {
		t.Fatalf("undo backup holds %q", got)
	}
	// The backup restored is untouched.
	if got := backupContent(t, filepath.Join(gameDir, "20261018-100000")); len(got) != 2 {
		t.Fatalf("restored backup = %v", got)
	}

	// Undo: back to the content before the restore. The bestiary stays as
	// restored, as the undo backup does not hold one.
	if err = RestoreBackup(global.Six, undo, target); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(target, "slot")); string(got) != "new slot" {
		t.Fatalf("slot after undo = %q", got)
	}
}

func TestRestoreBackupRoundTrip(t *testing.T) {
	t.Setenv("PR_PRINT", "false")
	useBackupRoot(t)
	useClock(t, time.Date(2026, 10, 18, 16, 30, 0, 0, time.Local))
	dir, slot, bestiary := pcSaveDir(t)
	data, err := LoadSave(global.Two, dir, "slot", global.PC)
	if err != nil {
		t.Fatal(err)
	}
	changeBestiary(t, data)
	if err = SaveSave(global.Two, data, 1, dir, "slot", global.PC); err != nil {
		t.Fatal(err)
	}
	saved, _ := os.ReadFile(filepath.Join(dir, "slot"))
	backups, err := ListBackups(global.Two)
	if err != nil || len(backups) != 1 {
		t.Fatalf("ListBackups = %v, %v", backups, err)
	}
	if err = RestoreBackup(global.Two, backups[0], dir); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "slot")); !bytes.Equal(got, slot) {
		t.Fatal("restored slot differs from the original")
	}
	if got, _ := os.ReadFile(filepath.Join(dir, bestiaryFile)); !bytes.Equal(got, bestiary) {
		t.Fatal("restored bestiary differs from the original")
	}
	if _, err = LoadSave(global.Two, dir, "slot", global.PC); err != nil {
		t.Fatalf("restored save does not load: %v", err)
	}
	backups, _ = ListBackups(global.Two)
	if len(backups) != 2 || backups[0].Reason != ReasonRestore {
		t.Fatalf("backups = %+v", backups)
	}
	if got, _ := os.ReadFile(filepath.Join(BackupsDir(global.Two), backups[0].Name, "slot")); !bytes.Equal(got, saved) {
		t.Fatal("undo backup does not hold the overwritten slot")
	}
	assertNoTempFiles(t, dir)
}

func TestRestoreBackupRefuses(t *testing.T) {
	root := useBackupRoot(t)
	gameDir := filepath.Join(root, "FF1")
	writeFiles(t, filepath.Join(gameDir, "20261018-100000"), map[string]string{"slot": "old"})
	writeFiles(t, filepath.Join(root, "outside"), map[string]string{"secret": "s"})
	good := Backup{Name: "20261018-100000", Files: []string{"slot"}}

	target := t.TempDir()
	writeFiles(t, target, map[string]string{"slot": "current"})
	notDir := filepath.Join(target, "slot")

	for name, tc := range map[string]struct {
		b      Backup
		target string
		want   string
	}{
		"missing target":  {good, filepath.Join(target, "missing"), "cannot restore"},
		"target is file":  {good, notDir, "not a directory"},
		"parent dir":      {Backup{Name: "20261018-100000", Files: []string{"../outside/secret"}}, target, "invalid file name"},
		"dot dot":         {Backup{Name: "20261018-100000", Files: []string{".."}}, target, "invalid file name"},
		"separator":       {Backup{Name: "20261018-100000", Files: []string{`a\b`}}, target, "invalid file name"},
		"absolute":        {Backup{Name: "20261018-100000", Files: []string{"/etc/passwd"}}, target, "invalid file name"},
		"manifest":        {Backup{Name: "20261018-100000", Files: []string{manifestName}}, target, "invalid file name"},
		"bad folder":      {Backup{Name: "../outside", Files: []string{"secret"}}, target, "invalid backup name"},
		"no files":        {Backup{Name: "20261018-100000"}, target, "holds no file"},
		"missing in copy": {Backup{Name: "20261018-100000", Files: []string{"slot", "gone"}}, target, "incomplete"},
	} {
		t.Run(name, func(t *testing.T) {
			err := RestoreBackup(global.One, tc.b, tc.target)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
			if got, _ := os.ReadFile(filepath.Join(target, "slot")); string(got) != "current" {
				t.Fatalf("slot changed to %q", got)
			}
			if got := listDir(t, gameDir); len(got) != 1 {
				t.Fatalf("backups = %v, no undo backup expected", got)
			}
		})
	}
}

func TestRestoreBackupKeepsRestored(t *testing.T) {
	root := useBackupRoot(t)
	useClock(t, time.Date(2026, 10, 18, 18, 0, 0, 0, time.Local))
	gameDir := filepath.Join(root, "FF3")
	for i := 0; i < maxBackups; i++ {
		writeFiles(t, filepath.Join(gameDir, fmt.Sprintf("20250101-0000%02d", i)), map[string]string{"slot": fmt.Sprint(i)})
	}
	target := t.TempDir()
	writeFiles(t, target, map[string]string{"slot": "current"})

	backups, err := ListBackups(global.Three)
	if err != nil || len(backups) != maxBackups {
		t.Fatalf("ListBackups = %d, %v", len(backups), err)
	}
	oldest := backups[len(backups)-1]
	if oldest.Name != "20250101-000000" {
		t.Fatalf("oldest = %s", oldest.Name)
	}
	for i := 0; i < 3; i++ {
		if err = RestoreBackup(global.Three, oldest, target); err != nil {
			t.Fatal(err)
		}
	}
	if got, _ := os.ReadFile(filepath.Join(target, "slot")); string(got) != "0" {
		t.Fatalf("slot = %q", got)
	}
	names := listDir(t, gameDir)
	if len(names) != maxBackups || !slices.Contains(names, oldest.Name) {
		t.Fatalf("backups after restoring = %v", names)
	}
	// The next oldest made room for the undo backups.
	for _, gone := range []string{"20250101-000001", "20250101-000002", "20250101-000003"} {
		if slices.Contains(names, gone) {
			t.Fatalf("%s should have been pruned: %v", gone, names)
		}
	}
	if !slices.Contains(names, "20261018-180003") {
		t.Fatalf("newest undo backup missing: %v", names)
	}
}
