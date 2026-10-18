package core

import (
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
)

type (
	MapData struct {
		Map          *save.MapData
		Player       *save.PlayerEntity
		Gps          *save.GpsData
		OtherParties *[]*save.OtherPartyData
	}
)

func NewMapData(game global.Game, md *save.MapData) (m *MapData, err error) {
	m = &MapData{Map: md}
	if m.Player, err = md.PlayerEntity(); err == nil {
		if m.Gps, err = md.GpsData(); err == nil {
			if game.IsSix() {
				m.OtherParties = md.OtherPartyDataList
			}
		}
	}
	return
}

// ToSave writes the edited map data into md, the map data unpacked again
// from the save. The editor edits d.Map, which was unpacked from the same
// save, so every field of md (including the members MapData does not know)
// is replaced by its edited value.
func (d MapData) ToSave(game global.Game, md *save.MapData) (err error) {
	*md = *d.Map
	if err = md.SetGpsData(d.Gps); err != nil {
		return
	}
	if err = md.SetPlayerEntity(d.Player); err != nil {
		return
	}
	// d.OtherParties is d.Map.OtherPartyDataList: the parties were copied
	// above, only their edited player entities are left to encode.
	if game.IsSix() && d.OtherParties != nil {
		for _, p := range *d.OtherParties {
			if p != nil && p.PE != nil {
				if err = p.SetPlayerEntity(p.PE); err != nil {
					return
				}
			}
		}
	}
	return
}
