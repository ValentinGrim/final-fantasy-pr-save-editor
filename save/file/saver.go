package file

import (
	"fmt"
	"os"
	"path/filepath"

	oj "github.com/virtuald/go-ordered-json"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/save/config"
)

func SaveSave(game global.Game, data *save.Data, slot int, dir string, file string, saveType global.SaveFileType) (err error) {
	var (
		out []byte
	)

	if out, err = oj.Marshal(data); err != nil {
		return
	}

	if err = saveFile(game, out, dir, file, data.Trimmed, saveType, "save"); err != nil {
		return
	}

	// The bestiary file is shared by all slots: only write it when it was
	// read successfully and its content was changed in the editor.
	if b, prefix, changed := data.BestiaryToSave(); changed {
		if err = saveFile(game, b, dir, bestiaryFile, prefix, saveType, "bestiary"); err != nil {
			err = fmt.Errorf("the save was written but the bestiary was not: %w", err)
		}
	}
	return
}

func saveFile(game global.Game, data []byte, dir string, file string, trimmed []byte, saveType global.SaveFileType, name string) (err error) {
	printFile(filepath.Join(config.Dir(game), "_"+name+"_saved.json"), data)
	if saveType == global.PC {
		if data, err = encode(data, trimmed); err != nil {
			return
		}
	}

	// Write to file
	toFile := filepath.Join(dir, file)
	if err = os.WriteFile(toFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write save file %s: %v", toFile, err)
	}
	return
}
