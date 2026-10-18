package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save/config"
)

const (
	// BestiaryFile is the name of the bestiary file shared by every slot.
	BestiaryFile = bestiaryFile

	// manifestName is written in every backup folder. It cannot be mistaken
	// for a save: slot files are Steam UUIDs ("...=") or "slotN.sav".
	manifestName    = "backup.json"
	manifestVersion = 1

	// Why a backup was made.
	ReasonSave    = "save"
	ReasonRestore = "restore"
)

type (
	// Backup is one backup folder of a game.
	Backup struct {
		// Name is the folder name, "YYYYMMDD-HHMMSS" with an optional "-NN".
		Name string
		// Time is read from the folder name, in local time.
		Time time.Time
		// Files are the names of the backed up files, sorted.
		Files []string
		// SourceDir is the save directory the files were copied from, "" for
		// backups made before manifests existed.
		SourceDir string
		// SaveType is meaningful when SaveTypeKnown is set.
		SaveType      global.SaveFileType
		SaveTypeKnown bool
		// Reason is ReasonSave, ReasonRestore or "" when unknown.
		Reason string
	}

	manifest struct {
		Version   int      `json:"version"`
		Game      int      `json:"game"`
		SourceDir string   `json:"sourceDir"`
		Created   string   `json:"created"`
		Files     []string `json:"files"`
		SaveType  string   `json:"saveType,omitempty"`
		Reason    string   `json:"reason,omitempty"`
	}
)

func defaultBackupRoot() string {
	return filepath.Join(filepath.Dir(config.Path()), "backups")
}

// SetBackupRoot makes the backups live in dir instead of next to the config
// file, and returns a function putting the previous location back. It is
// meant for tests of other packages.
func SetBackupRoot(dir string) (restore func()) {
	prev := backupRoot
	backupRoot = func() string { return dir }
	return func() { backupRoot = prev }
}

// BackupsDir returns the folder holding the backups of game. It may not exist
// yet.
func BackupsDir(game global.Game) string {
	return filepath.Join(backupRoot(), fmt.Sprintf("FF%d", game))
}

func saveTypeName(t global.SaveFileType) string {
	if t == global.PS {
		return "PS"
	}
	return "PC"
}

func writeManifest(folder string, m manifest) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(folder, manifestName), append(b, '\n'), 0o644)
}

// ListBackups returns the backups of game, newest first. Folders that are not
// backups and backups holding no file are left out; a missing or unreadable
// manifest only leaves SourceDir, SaveType and Reason unknown.
func ListBackups(game global.Game) ([]Backup, error) {
	gameDir := BackupsDir(game)
	entries, err := os.ReadDir(gameDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var backups []Backup
	for _, e := range entries {
		if !e.IsDir() || !backupDirName.MatchString(e.Name()) {
			continue
		}
		b, err := readBackup(gameDir, e.Name())
		if err != nil || len(b.Files) == 0 {
			continue
		}
		backups = append(backups, b)
	}
	// The names sort chronologically.
	sort.Slice(backups, func(i, j int) bool { return backups[i].Name > backups[j].Name })
	return backups, nil
}

func readBackup(gameDir, name string) (b Backup, err error) {
	b.Name = name
	if b.Time, err = time.ParseInLocation(backupTimeLayout, name[:len(backupTimeLayout)], time.Local); err != nil {
		return
	}
	folder := filepath.Join(gameDir, name)
	entries, err := os.ReadDir(folder)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.Type().IsRegular() && e.Name() != manifestName && validFileName(e.Name()) == nil {
			b.Files = append(b.Files, e.Name())
		}
	}
	sort.Strings(b.Files)

	var m manifest
	if raw, rerr := os.ReadFile(filepath.Join(folder, manifestName)); rerr == nil && json.Unmarshal(raw, &m) == nil {
		b.SourceDir = m.SourceDir
		b.Reason = m.Reason
		switch m.SaveType {
		case "PC":
			b.SaveType, b.SaveTypeKnown = global.PC, true
		case "PS":
			b.SaveType, b.SaveTypeKnown = global.PS, true
		}
	}
	return b, nil
}

// validFileName refuses anything but a plain file name.
func validFileName(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\:`) ||
		filepath.Base(name) != name || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return fmt.Errorf("invalid file name %q in the backup", name)
	}
	return nil
}

// RestoreBackup copies the files of the backup b of game into targetDir.
// The files of targetDir about to be replaced are first backed up (reason
// "restore"), so that the restore can be undone; if that fails nothing is
// written. Each file is replaced atomically. Files of targetDir that are not
// in the backup are left alone. Old backups are then pruned, never b itself.
func RestoreBackup(game global.Game, b Backup, targetDir string) error {
	if !backupDirName.MatchString(b.Name) {
		return fmt.Errorf("invalid backup name %q", b.Name)
	}
	if len(b.Files) == 0 {
		return errors.New("the backup holds no file")
	}
	fi, err := os.Stat(targetDir)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot restore to %s: the directory does not exist", targetDir)
	}
	if err != nil {
		return fmt.Errorf("cannot restore to %s: %w", targetDir, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("cannot restore to %s: not a directory", targetDir)
	}

	gameDir := BackupsDir(game)
	folder := filepath.Join(gameDir, b.Name)
	// Read everything first: a broken backup must not leave a half restore.
	data := make([][]byte, len(b.Files))
	for i, n := range b.Files {
		if err = validFileName(n); err != nil {
			return err
		}
		if n == manifestName {
			return fmt.Errorf("invalid file name %q in the backup", n)
		}
		p := filepath.Join(folder, n)
		if fi, err = os.Lstat(p); err != nil {
			return fmt.Errorf("the backup is incomplete: %w", err)
		}
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("%s in the backup is not a regular file", n)
		}
		if data[i], err = os.ReadFile(p); err != nil {
			return err
		}
	}

	saveType := b.SaveType
	if !b.SaveTypeKnown {
		saveType = guessSaveType(b.Files)
	}
	undo, err := makeBackup(game, targetDir, saveType, ReasonRestore, b.Files...)
	if err != nil {
		return fmt.Errorf("could not back up the current files, nothing was restored: %w", err)
	}

	for i, n := range b.Files {
		p := filepath.Join(targetDir, n)
		if err = writeFileAtomic(p, data[i]); err != nil {
			err = fmt.Errorf("failed to write %s: %w", p, err)
			if i > 0 && undo != "" {
				err = fmt.Errorf("%w (some files were already restored; the previous ones are in %s)", err, undo)
			}
			break
		}
	}
	pruneBackups(gameDir, maxBackups, b.Name)
	return err
}

func guessSaveType(files []string) global.SaveFileType {
	for _, f := range files {
		if strings.HasSuffix(f, ".sav") {
			return global.PS
		}
	}
	return global.PC
}
