package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	oj "github.com/virtuald/go-ordered-json"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/save/config"
)

// bestiaryFile holds the bestiary shared by every slot of a save directory.
const bestiaryFile = "dp3fS2vqP7GDj8eF72YKqbT7FIAF=e7Shy2CsTITm2E="

func LoadSave(game global.Game, dir string, fileName string, saveType global.SaveFileType) (data *save.Data, err error) {
	var b []byte
	data = save.New(game)
	if b, data.Trimmed, err = loadFile(game, filepath.Join(dir, fileName), saveType, "save"); err != nil {
		return
	}
	if err = oj.Unmarshal(b, data); err != nil {
		return
	}
	loadBestiary(game, dir, saveType, data)
	return
}

// loadBestiary is best effort: if the bestiary file is missing or unreadable
// the slot still loads, the bestiary is not editable and SaveSave leaves the
// file alone.
func loadBestiary(game global.Game, dir string, saveType global.SaveFileType, data *save.Data) {
	b, prefix, err := loadFile(game, filepath.Join(dir, bestiaryFile), saveType, "bestiary")
	if err == nil && !json.Valid(b) {
		err = errors.New("not valid JSON")
	}
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("bestiary not loaded: %v", err)
		}
		return
	}
	data.SetLoadedBestiary(b, prefix)
}

func loadFile(game global.Game, fromFile string, saveType global.SaveFileType, name string) (out []byte, trimmed []byte, err error) {
	var b []byte
	if b, err = os.ReadFile(fromFile); err != nil {
		return
	}
	if saveType == global.PS {
		return b, nil, nil
	}
	if out, trimmed, err = decode(b); err != nil {
		return nil, nil, fmt.Errorf("unable to load %s: %w", filepath.Base(fromFile), err)
	}
	printFile(filepath.Join(config.Dir(game), "_"+name+"_loaded.json"), out)
	return
}
