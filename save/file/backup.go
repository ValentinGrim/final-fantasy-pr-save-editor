package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"time"

	"pixel-remastered-save-editor/global"
)

// maxBackups is how many backup folders are kept per game.
const maxBackups = 20

const backupTimeLayout = "20060102-150405"

var (
	// backupRoot returns the folder holding the backups. It is next to the
	// config file, outside the save directory, so that neither the game nor
	// Steam Cloud ever see the copies. Tests replace it.
	backupRoot = defaultBackupRoot
	// now and renameFile are replaced by tests.
	now        = time.Now
	renameFile = os.Rename

	backupDirName = regexp.MustCompile(`^\d{8}-\d{6}(-\d+)?$`)
)

// backupFiles copies the files among names that exist in dir into a new
// folder <backupRoot>/FF<n>/<YYYYMMDD-HHMMSS>/, then removes the oldest
// folders so that only maxBackups remain. It returns the folder created, or ""
// when none of the files exists yet. On error nothing is left behind.
func backupFiles(game global.Game, dir string, saveType global.SaveFileType, reason string, names ...string) (folder string, err error) {
	if folder, err = makeBackup(game, dir, saveType, reason, names...); err == nil && folder != "" {
		pruneBackups(filepath.Dir(folder), maxBackups)
	}
	return
}

// makeBackup is backupFiles without the pruning. Along with the copies, the
// folder gets a manifest (see manifestName) recording where they came from.
func makeBackup(game global.Game, dir string, saveType global.SaveFileType, reason string, names ...string) (folder string, err error) {
	var existing []string
	for _, n := range names {
		fi, err := os.Stat(filepath.Join(dir, n))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		if !fi.Mode().IsRegular() {
			return "", fmt.Errorf("%s is not a regular file", n)
		}
		existing = append(existing, n)
	}
	if len(existing) == 0 {
		return "", nil
	}

	gameDir := BackupsDir(game)
	if err = os.MkdirAll(gameDir, 0o755); err != nil {
		return "", err
	}
	created := now()
	if folder, err = makeBackupFolder(gameDir, created); err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(folder)
			folder = ""
		}
	}()
	for _, n := range existing {
		if err = copyFile(filepath.Join(dir, n), filepath.Join(folder, n)); err != nil {
			return
		}
	}
	if abs, aerr := filepath.Abs(dir); aerr == nil {
		dir = abs
	}
	err = writeManifest(folder, manifest{
		Version:   manifestVersion,
		Game:      int(game),
		SourceDir: dir,
		Created:   created.Format(time.RFC3339),
		Files:     existing,
		SaveType:  saveTypeName(saveType),
		Reason:    reason,
	})
	return
}

// makeBackupFolder creates a new timestamped folder in gameDir. Saving twice
// in the same second adds a -2, -3... suffix.
func makeBackupFolder(gameDir string, t time.Time) (string, error) {
	stamp := t.Format(backupTimeLayout)
	for i := 1; i < 100; i++ {
		name := stamp
		if i > 1 {
			name = fmt.Sprintf("%s-%02d", stamp, i)
		}
		p := filepath.Join(gameDir, name)
		err := os.Mkdir(p, 0o755)
		if err == nil {
			return p, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return "", err
		}
	}
	return "", fmt.Errorf("too many backups named %s in %s", stamp, gameDir)
}

// pruneBackups removes the oldest backup folders of gameDir beyond keep.
// Only folders named by makeBackupFolder are considered, and the folders named
// in protect are never removed (they still count towards keep). Errors are
// ignored: a backup that cannot be removed does no harm.
func pruneBackups(gameDir string, keep int, protect ...string) {
	entries, err := os.ReadDir(gameDir)
	if err != nil {
		return
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() && backupDirName.MatchString(e.Name()) {
			dirs = append(dirs, e.Name())
		}
	}
	remove := len(dirs) - keep
	if remove <= 0 {
		return
	}
	// The names sort chronologically.
	sort.Strings(dirs)
	for _, d := range dirs {
		if remove == 0 {
			break
		}
		if slices.Contains(protect, d) {
			continue
		}
		_ = os.RemoveAll(filepath.Join(gameDir, d))
		remove--
	}
}

func copyFile(from, to string) (err error) {
	var in, out *os.File
	if in, err = os.Open(from); err != nil {
		return
	}
	defer in.Close()
	if out, err = os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644); err != nil {
		return
	}
	if _, err = io.Copy(out, in); err == nil {
		err = out.Sync()
	}
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(to)
		return
	}
	if fi, err := in.Stat(); err == nil {
		_ = os.Chtimes(to, fi.ModTime(), fi.ModTime())
	}
	return nil
}

// writeFileAtomic replaces path with data. The data is written and synced to
// a temporary file in the same directory, then renamed over path, so that
// path holds either the old or the new content, never a partial one. The
// temporary name (".<name>.tmp<random>") does not look like a save file and
// is removed on failure. The permission bits of an existing file are kept;
// a read-only file is not replaced.
func writeFileAtomic(path string, data []byte) (err error) {
	// Replace the target of a symlink, not the link itself.
	if p, err := filepath.EvalSymlinks(path); err == nil {
		path = p
	}
	perm := os.FileMode(0o644)
	if fi, err := os.Stat(path); err == nil {
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", path)
		}
		perm = fi.Mode().Perm()
		if perm&0o200 == 0 {
			return fmt.Errorf("%s is read-only", path)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	dir, name := filepath.Split(path)
	if dir == "" {
		dir = "."
	}
	f, err := os.CreateTemp(dir, "."+name+".tmp*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
		}
	}()

	if _, err = f.Write(data); err != nil {
		return
	}
	if err = f.Sync(); err != nil {
		return
	}
	if err = f.Close(); err != nil {
		return
	}
	if err = os.Chmod(tmp, perm); err != nil {
		return
	}
	if err = renameFile(tmp, path); err != nil {
		return
	}
	syncDir(dir)
	return nil
}

// syncDir makes the rename durable on Unix. Windows cannot open a directory
// for syncing; NTFS journals the rename itself.
func syncDir(dir string) {
	if runtime.GOOS == "windows" {
		return
	}
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
}
