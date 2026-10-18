package core

import (
	"strings"
	"time"

	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
)

// timestampLayouts are the layouts of the save time stamps the game writes:
// it formats the time with the culture of the system, so a save made on an
// en-US system has "10/16/2026 4:09:06 PM" and one made on a fr-FR system
// "18/10/2026 14:46:28". The layouts accept numbers with or without a
// leading zero. A time stamp that matches none of them is not editable.
var timestampLayouts = []string{
	"1/2/2006 3:04:05 PM", // en-US: M/d/yyyy h:mm:ss tt
	"2/1/2006 15:04:05",   // fr-FR, en-GB, es-ES, it-IT, pt-BR: dd/MM/yyyy HH:mm:ss
	"1/2/2006 15:04:05",   // M/d/yyyy HH:mm:ss
	"2.1.2006 15:04:05",   // de-DE, ru-RU: dd.MM.yyyy HH:mm:ss
	"2006/1/2 15:04:05",   // ja-JP, zh-CN: yyyy/MM/dd H:mm:ss
	"2006-1-2 15:04:05",   // sv-SE, ISO 8601 with a space
	"2006-01-02T15:04:05", // ISO 8601
	"2006-01-02T15:04:05Z07:00",
}

// timestampLayout returns the first layout of timestampLayouts that parses s.
func timestampLayout(s string) (layout string, ok bool) {
	for _, l := range timestampLayouts {
		if _, err := time.Parse(l, s); err == nil {
			return l, true
		}
	}
	return "", false
}

type (
	Misc struct {
		OwnedGil               int
		TotalGil               *int
		Steps                  int
		NumberOfSaves          int
		CursedShieldFightCount int
		EscapeCount            int
		BattleCount            int
		OpenedChestCount       int
		IsCompleteFlag         bool
		PlayTime               float64
		OwnedCrystals          []bool
		MonstersKilledCount    int
		CorpsSlotIndex         int
		// TimeStamp is nil when the saved time stamp has no known layout.
		TimeStamp *string
		// timeStamp is the saved time stamp and timeStampLayout its layout.
		timeStamp       string
		timeStampLayout string
	}
)

func NewMisc(data *save.Data, ud *save.UserData, ds *save.DataStorage) (m *Misc) {
	m = &Misc{
		OwnedGil:               ud.OwnedGil,
		TotalGil:               ud.TotalGil,
		Steps:                  ud.Steps,
		CorpsSlotIndex:         ud.CorpsSlotIndex,
		NumberOfSaves:          ud.SaveCompleteCount,
		CursedShieldFightCount: 0,
		EscapeCount:            ud.EscapeCount,
		BattleCount:            ud.BattleCount,
		OpenedChestCount:       ud.OpenChestCount,
		PlayTime:               ud.PlayTime,
		MonstersKilledCount:    ud.MonstersKilledCount,
		IsCompleteFlag:         data.IsCompleteFlag == 1,
	}
	m.OwnedCrystals, _ = ud.OwnedCrystalFlags()
	if layout, ok := timestampLayout(data.TimeStamp); ok {
		ts := data.TimeStamp
		m.TimeStamp = &ts
		m.timeStamp = data.TimeStamp
		m.timeStampLayout = layout
	}
	if data.Game.IsSix() {
		if len(ds.Global) > 9 {
			m.CursedShieldFightCount = ds.Global[9]
		}
	}
	return
}

func (m *Misc) ToSave(game global.Game, ud *save.UserData, ds *save.DataStorage) {
	ud.OwnedGil = m.OwnedGil
	ud.TotalGil = m.TotalGil
	ud.Steps = m.Steps
	ud.CorpsSlotIndex = m.CorpsSlotIndex
	ud.SaveCompleteCount = m.NumberOfSaves
	ud.EscapeCount = m.EscapeCount
	ud.BattleCount = m.BattleCount
	ud.OpenChestCount = m.OpenedChestCount
	ud.PlayTime = m.PlayTime
	ud.MonstersKilledCount = m.MonstersKilledCount
	_ = ud.SetOwnedCrystalFlags(m.OwnedCrystals)
	if game.IsSix() {
		if len(ds.Global) > 9 {
			ds.Global[9] = m.CursedShieldFightCount
		}
	}
}

func (m *Misc) ToSaveData(data *save.Data) {
	if m.IsCompleteFlag {
		data.IsCompleteFlag = 1
	} else {
		data.IsCompleteFlag = 0
	}
	// The time stamp is written only when it was edited, and only in the
	// layout of the saved one so the game reads it with the same culture.
	if m.TimeStamp != nil && *m.TimeStamp != m.timeStamp {
		t := strings.TrimSpace(*m.TimeStamp)
		if _, err := time.Parse(m.timeStampLayout, t); err == nil {
			data.TimeStamp = t
		}
	}
}
