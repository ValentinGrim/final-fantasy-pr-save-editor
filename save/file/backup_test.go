package file

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
)

// TestMain keeps every test of the package away from the real config dir.
func TestMain(m *testing.M) {
	root, err := os.MkdirTemp("", "ffpr-backups-")
	if err != nil {
		panic(err)
	}
	backupRoot = func() string { return root }
	code := m.Run()
	_ = os.RemoveAll(root)
	os.Exit(code)
}

// useBackupRoot points the backups to a fresh dir for one test.
func useBackupRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	prev := backupRoot
	backupRoot = func() string { return root }
	t.Cleanup(func() { backupRoot = prev })
	return root
}

func useClock(t *testing.T, start time.Time) {
	t.Helper()
	cur := start
	prev := now
	now = func() time.Time { cur = cur.Add(time.Second); return cur }
	t.Cleanup(func() { now = prev })
}

func listDir(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names
}

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	for _, n := range listDir(t, dir) {
		if strings.Contains(n, ".tmp") {
			t.Fatalf("temporary file left behind: %s", n)
		}
	}
}

// pcSaveDir returns a save dir holding an encoded slot and bestiary.
func pcSaveDir(t *testing.T) (dir string, slot, bestiary []byte) {
	t.Helper()
	dir = t.TempDir()
	slot = writeEncoded(t, filepath.Join(dir, "slot"), []byte(syntheticSlot))
	bestiary = writeEncoded(t, filepath.Join(dir, bestiaryFile), syntheticBestiary(20))
	return
}

func changeBestiary(t *testing.T, data *save.Data) {
	t.Helper()
	b, err := data.Bestiary()
	if err != nil || b == nil {
		t.Fatalf("bestiary not loaded: %v", err)
	}
	b.TotalSubjugationCount++
	if err = data.SetBestiary(b); err != nil {
		t.Fatal(err)
	}
}

func TestSaveSaveBacksUp(t *testing.T) {
	t.Setenv("PR_PRINT", "false")
	root := useBackupRoot(t)
	useClock(t, time.Date(2026, 10, 18, 16, 30, 0, 0, time.Local))
	dir, slot, bestiary := pcSaveDir(t)
	if err := os.Chmod(filepath.Join(dir, "slot"), 0o640); err != nil {
		t.Fatal(err)
	}

	data, err := LoadSave(global.Two, dir, "slot", global.PC)
	if err != nil {
		t.Fatal(err)
	}
	changeBestiary(t, data)
	if err = SaveSave(global.Two, data, 1, dir, "slot", global.PC); err != nil {
		t.Fatal(err)
	}

	folders := listDir(t, filepath.Join(root, "FF2"))
	if len(folders) != 1 || folders[0] != "20261018-163001" {
		t.Fatalf("backup folders = %v", folders)
	}
	folder := filepath.Join(root, "FF2", folders[0])
	if got, _ := os.ReadFile(filepath.Join(folder, "slot")); !bytes.Equal(got, slot) {
		t.Fatal("slot backup differs from the original")
	}
	if got, _ := os.ReadFile(filepath.Join(folder, bestiaryFile)); !bytes.Equal(got, bestiary) {
		t.Fatal("bestiary backup differs from the original")
	}
	if got, _ := os.ReadFile(filepath.Join(dir, bestiaryFile)); bytes.Equal(got, bestiary) {
		t.Fatal("bestiary not written")
	}
	assertNoTempFiles(t, dir)
	if runtime.GOOS != "windows" {
		if fi, _ := os.Stat(filepath.Join(dir, "slot")); fi.Mode().Perm() != 0o640 {
			t.Fatalf("permissions not kept: %v", fi.Mode().Perm())
		}
	}
	if _, err = LoadSave(global.Two, dir, "slot", global.PC); err != nil {
		t.Fatalf("written save does not load: %v", err)
	}

	// Unchanged bestiary: only the slot is backed up.
	if err = SaveSave(global.Two, data, 1, dir, "slot", global.PC); err != nil {
		t.Fatal(err)
	}
	folders = listDir(t, filepath.Join(root, "FF2"))
	if len(folders) != 2 {
		t.Fatalf("backup folders = %v", folders)
	}
	assertNoTempFiles(t, dir)
}

func TestSaveSaveNewFileNoBackup(t *testing.T) {
	t.Setenv("PR_PRINT", "false")
	root := useBackupRoot(t)
	dir := t.TempDir()
	writeEncoded(t, filepath.Join(dir, "slot"), []byte(syntheticSlot))
	data, err := LoadSave(global.One, dir, "slot", global.PC)
	if err != nil {
		t.Fatal(err)
	}
	if err = SaveSave(global.One, data, 2, dir, "other", global.PC); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(root, "FF1")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("no backup expected for a new file: %v", err)
	}
	if fi, err := os.Stat(filepath.Join(dir, "other")); err != nil || (runtime.GOOS != "windows" && fi.Mode().Perm() != 0o644) {
		t.Fatalf("new file: %v %v", fi, err)
	}
	assertNoTempFiles(t, dir)
}

func TestSaveSavePlayStation(t *testing.T) {
	root := useBackupRoot(t)
	dir := t.TempDir()
	orig := []byte(syntheticSlot)
	if err := os.WriteFile(filepath.Join(dir, "slot1.sav"), orig, 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := LoadSave(global.Three, dir, "slot1.sav", global.PS)
	if err != nil {
		t.Fatal(err)
	}
	if err = SaveSave(global.Three, data, 1, dir, "slot1.sav", global.PS); err != nil {
		t.Fatal(err)
	}
	folders := listDir(t, filepath.Join(root, "FF3"))
	if len(folders) != 1 {
		t.Fatalf("backup folders = %v", folders)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "FF3", folders[0], "slot1.sav")); !bytes.Equal(got, orig) {
		t.Fatal("PlayStation backup differs from the original")
	}
	assertNoTempFiles(t, dir)
	if _, err = LoadSave(global.Three, dir, "slot1.sav", global.PS); err != nil {
		t.Fatal(err)
	}
}

func TestBackupPruning(t *testing.T) {
	root := useBackupRoot(t)
	useClock(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	gameDir := filepath.Join(root, "FF1")
	// Older backups, plus entries that are not backups and must be kept.
	for i := 0; i < 25; i++ {
		if err := os.MkdirAll(filepath.Join(gameDir, fmt.Sprintf("20250101-0000%02d", i)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, n := range []string{"keep-me", "20250101-000000.txt"} {
		if err := os.Mkdir(filepath.Join(gameDir, n), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "slot"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	var made []string
	for i := 0; i < 3; i++ {
		folder, err := backupFiles(global.One, dir, "slot", bestiaryFile)
		if err != nil {
			t.Fatal(err)
		}
		made = append(made, filepath.Base(folder))
	}

	var backups []string
	others := 0
	for _, n := range listDir(t, gameDir) {
		if backupDirName.MatchString(n) {
			backups = append(backups, n)
		} else {
			others++
		}
	}
	if len(backups) != maxBackups || others != 2 {
		t.Fatalf("got %d backups and %d other entries: %v", len(backups), others, backups)
	}
	// The oldest are gone, the new ones are kept.
	if backups[0] != "20250101-000008" || backups[maxBackups-1] != made[2] {
		t.Fatalf("wrong backups pruned: %v", backups)
	}
	// Only existing files are copied.
	if got := listDir(t, filepath.Join(gameDir, made[0])); len(got) != 1 || got[0] != "slot" {
		t.Fatalf("backup content = %v", got)
	}
}

func TestBackupSameSecond(t *testing.T) {
	root := useBackupRoot(t)
	fixed := time.Date(2026, 10, 18, 12, 0, 0, 0, time.Local)
	prev := now
	now = func() time.Time { return fixed }
	t.Cleanup(func() { now = prev })

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "slot"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := backupFiles(global.Four, dir, "slot"); err != nil {
			t.Fatal(err)
		}
	}
	got := strings.Join(listDir(t, filepath.Join(root, "FF4")), " ")
	if got != "20261018-120000 20261018-120000-02 20261018-120000-03" {
		t.Fatalf("backup folders = %s", got)
	}
}

func TestSaveSaveBackupFailureWritesNothing(t *testing.T) {
	t.Setenv("PR_PRINT", "false")
	// The backup root is a file: the backup folder cannot be created.
	f := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(f, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	prev := backupRoot
	backupRoot = func() string { return f }
	t.Cleanup(func() { backupRoot = prev })

	dir, slot, bestiary := pcSaveDir(t)
	data, err := LoadSave(global.Two, dir, "slot", global.PC)
	if err != nil {
		t.Fatal(err)
	}
	changeBestiary(t, data)
	err = SaveSave(global.Two, data, 1, dir, "slot", global.PC)
	if err == nil || !strings.Contains(err.Error(), "could not back up the existing save, nothing was written") {
		t.Fatalf("err = %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "slot")); !bytes.Equal(got, slot) {
		t.Fatal("slot changed")
	}
	if got, _ := os.ReadFile(filepath.Join(dir, bestiaryFile)); !bytes.Equal(got, bestiary) {
		t.Fatal("bestiary changed")
	}
	assertNoTempFiles(t, dir)
}

func TestSaveSaveWriteFailureKeepsOriginal(t *testing.T) {
	t.Setenv("PR_PRINT", "false")

	t.Run("rename fails", func(t *testing.T) {
		root := useBackupRoot(t)
		dir, slot, _ := pcSaveDir(t)
		data, err := LoadSave(global.One, dir, "slot", global.PC)
		if err != nil {
			t.Fatal(err)
		}
		prev := renameFile
		renameFile = func(string, string) error { return errors.New("disk on fire") }
		t.Cleanup(func() { renameFile = prev })

		if err = SaveSave(global.One, data, 1, dir, "slot", global.PC); err == nil || !strings.Contains(err.Error(), "disk on fire") {
			t.Fatalf("err = %v", err)
		}
		if got, _ := os.ReadFile(filepath.Join(dir, "slot")); !bytes.Equal(got, slot) {
			t.Fatal("original changed")
		}
		assertNoTempFiles(t, dir)
		// The backup was made before the failed write.
		if folders := listDir(t, filepath.Join(root, "FF1")); len(folders) != 1 {
			t.Fatalf("backup folders = %v", folders)
		}
	})

	t.Run("read-only dir", func(t *testing.T) {
		if runtime.GOOS == "windows" || os.Geteuid() == 0 {
			t.Skip("directory permissions not enforced")
		}
		useBackupRoot(t)
		dir, slot, _ := pcSaveDir(t)
		data, err := LoadSave(global.One, dir, "slot", global.PC)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.Chmod(dir, 0o555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

		if err = SaveSave(global.One, data, 1, dir, "slot", global.PC); err == nil {
			t.Fatal("write into a read-only dir succeeded")
		}
		if got, _ := os.ReadFile(filepath.Join(dir, "slot")); !bytes.Equal(got, slot) {
			t.Fatal("original changed")
		}
		assertNoTempFiles(t, dir)
	})

	t.Run("read-only file", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("permission bits not enforced")
		}
		useBackupRoot(t)
		dir, slot, _ := pcSaveDir(t)
		data, err := LoadSave(global.One, dir, "slot", global.PC)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.Chmod(filepath.Join(dir, "slot"), 0o444); err != nil {
			t.Fatal(err)
		}
		if err = SaveSave(global.One, data, 1, dir, "slot", global.PC); err == nil || !strings.Contains(err.Error(), "read-only") {
			t.Fatalf("err = %v", err)
		}
		if got, _ := os.ReadFile(filepath.Join(dir, "slot")); !bytes.Equal(got, slot) {
			t.Fatal("original changed")
		}
		assertNoTempFiles(t, dir)
	})
}

func TestWriteFileAtomicSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks need privileges")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "real")
	link := filepath.Join(dir, "link")
	if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(link, []byte("new")); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Lstat(link); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("symlink replaced")
	}
	if got, _ := os.ReadFile(target); string(got) != "new" {
		t.Fatalf("target = %q", got)
	}
	if fi, _ := os.Stat(target); fi.Mode().Perm() != 0o600 {
		t.Fatalf("permissions not kept: %v", fi.Mode().Perm())
	}
	assertNoTempFiles(t, dir)
}
