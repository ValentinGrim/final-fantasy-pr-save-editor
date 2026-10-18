package core

import (
	"slices"

	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
)

type (
	Bestiary struct {
		Game          global.Game
		ScenarioFlags []int
		// Pointers so editors can bind to the entries while rows are added.
		MonsterDefeats []*MonsterDefeatedCount
	}
	MonsterDefeatedCount struct {
		ID    int
		Count int
	}
)

// bestiaryLastID is the highest monster ID the game records, for the games
// where it is known. The ID tables also list boss variants above it (e.g.
// FF1 #129 "Vampire (Cavern of Earth)", FF2 #129-142); the game counts those
// under their regular entry, and a completed FF1 bestiary holds exactly the
// IDs 1-128.
var bestiaryLastID = map[global.Game]int{
	global.One: 128,
	global.Two: 128,
}

func NewBestiary(game global.Game, bestiary *save.Bestiary) (b *Bestiary, err error) {
	b = &Bestiary{Game: game}
	var d *save.MonsterDefeats
	if d, err = bestiary.MonsterDefeats(); err != nil {
		return
	}
	b.MonsterDefeats = make([]*MonsterDefeatedCount, len(d.Keys))
	for i := range d.Keys {
		b.MonsterDefeats[i] = &MonsterDefeatedCount{ID: d.Keys[i], Count: d.Values[i]}
	}

	var f *save.ScenarioFlags
	if f, err = bestiary.ScenarioFlags(); err == nil {
		b.ScenarioFlags = f.Target
	}
	return
}

func (b *Bestiary) Add() *MonsterDefeatedCount {
	d := &MonsterDefeatedCount{}
	b.MonsterDefeats = append(b.MonsterDefeats, d)
	return d
}

// DefeatAll gives every monster of ids that has no defeat yet a count of 1,
// appended in ID order after the existing entries (the game lists monsters
// in the order they were first defeated). Existing counts are kept. IDs the
// game does not record are skipped.
func (b *Bestiary) DefeatAll(ids []int) {
	defeated := make(map[int]bool, len(b.MonsterDefeats))
	for _, d := range b.MonsterDefeats {
		if d.ID > 0 && d.Count > 0 {
			defeated[d.ID] = true
		}
	}
	ids = slices.Clone(ids)
	slices.Sort(ids)
	last := bestiaryLastID[b.Game]
	for _, id := range slices.Compact(ids) {
		if id <= 0 || (last > 0 && id > last) || defeated[id] {
			continue
		}
		found := false
		for _, d := range b.MonsterDefeats {
			if d.ID == id {
				d.Count = 1
				found = true
				break
			}
		}
		if !found {
			b.MonsterDefeats = append(b.MonsterDefeats, &MonsterDefeatedCount{ID: id, Count: 1})
		}
	}
}

// ToSave drops entries without an ID or a positive count and merges
// duplicated IDs (counts are added, the first position is kept), as the game
// stores a dictionary. The model itself is not modified.
func (b *Bestiary) ToSave() (bestiary *save.Bestiary, err error) {
	bestiary = &save.Bestiary{}

	var (
		keys   []int
		values []int
		index  = make(map[int]int)
		total  int
	)
	for _, m := range b.MonsterDefeats {
		if m == nil || m.ID <= 0 || m.Count <= 0 {
			continue
		}
		if i, found := index[m.ID]; found {
			values[i] += m.Count
		} else {
			index[m.ID] = len(keys)
			keys = append(keys, m.ID)
			values = append(values, m.Count)
		}
		total += m.Count
	}
	if keys == nil {
		keys, values = []int{}, []int{}
	}
	if err = bestiary.SetMonsterDefeats(&save.MonsterDefeats{Keys: keys, Values: values}); err != nil {
		bestiary = nil
		return
	}

	if err = bestiary.SetScenarioFlags(&save.ScenarioFlags{Target: b.ScenarioFlags}); err != nil {
		bestiary = nil
		return
	}
	bestiary.TotalSubjugationCount = total
	return
}
