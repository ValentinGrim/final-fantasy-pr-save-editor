package file

import (
	"fmt"
	"path/filepath"

	oj "github.com/virtuald/go-ordered-json"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/save/config"
)

// SaveSave writes the slot file, and the bestiary file when it was changed.
// The files about to be replaced are first copied to a backup folder (see
// backupFiles); if that fails nothing is written. Each file is replaced
// atomically, so an interrupted save leaves the previous content in place.
func SaveSave(game global.Game, data *save.Data, slot int, dir string, file string, saveType global.SaveFileType) (err error) {
	var (
		out      []byte
		bestiary []byte
	)

	if out, err = oj.Marshal(data); err != nil {
		return
	}
	if out, err = encodeFile(game, out, data.Trimmed, saveType, "save"); err != nil {
		return
	}

	// The bestiary file is shared by all slots: only write it when it was
	// read successfully and its content was changed in the editor.
	toBackup := []string{file}
	if b, prefix, changed := data.BestiaryToSave(); changed {
		if bestiary, err = encodeFile(game, b, prefix, saveType, "bestiary"); err != nil {
			return fmt.Errorf("could not encode the bestiary, nothing was written: %w", err)
		}
		toBackup = append(toBackup, bestiaryFile)
	}

	if _, err = backupFiles(game, dir, toBackup...); err != nil {
		return fmt.Errorf("could not back up the existing save, nothing was written: %w", err)
	}

	toFile := filepath.Join(dir, file)
	if err = writeFileAtomic(toFile, out); err != nil {
		return fmt.Errorf("failed to write save file %s: %w", toFile, err)
	}
	if bestiary != nil {
		toFile = filepath.Join(dir, bestiaryFile)
		if err = writeFileAtomic(toFile, bestiary); err != nil {
			return fmt.Errorf("the save was written but the bestiary was not: failed to write %s: %w", toFile, err)
		}
		data.BestiaryWritten()
	}
	return
}

func encodeFile(game global.Game, data []byte, trimmed []byte, saveType global.SaveFileType, name string) ([]byte, error) {
	printFile(filepath.Join(config.Dir(game), "_"+name+"_saved.json"), data)
	if saveType == global.PC {
		return encode(data, trimmed)
	}
	return data, nil
}
