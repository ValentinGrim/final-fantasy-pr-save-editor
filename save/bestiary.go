package save

import (
	"bytes"
	"errors"

	oj "github.com/virtuald/go-ordered-json"
)

type (
	// Bestiary is the content of the bestiary file. scenarioFlags and
	// monsterDefeats are JSON documents stored as strings, like the slot data.
	// The game keeps totalSubjugationCount equal to the sum of the defeat
	// counts.
	Bestiary struct {
		ScenarioFlagsInternal  string `json:"scenarioFlags"`
		MonsterDefeatsInternal string `json:"monsterDefeats"`
		TotalSubjugationCount  int    `json:"totalSubjugationCount"`
	}
	ScenarioFlags struct {
		Target []int `json:"target"`
	}
	// MonsterDefeats maps monster IDs (Keys) to defeat counts (Values). The
	// game lists monsters in the order they were first defeated.
	MonsterDefeats struct {
		Keys   []int `json:"keys"`
		Values []int `json:"values"`
	}
)

func (b *Bestiary) ScenarioFlags() (v *ScenarioFlags, err error) {
	return UnmarshalOne[ScenarioFlags](b.ScenarioFlagsInternal)
}

func (b *Bestiary) SetScenarioFlags(v *ScenarioFlags) (err error) {
	b.ScenarioFlagsInternal, err = MarshalOne[ScenarioFlags](v)
	return
}

func (b *Bestiary) MonsterDefeats() (v *MonsterDefeats, err error) {
	if v, err = UnmarshalOne[MonsterDefeats](b.MonsterDefeatsInternal); err == nil && len(v.Keys) != len(v.Values) {
		err = errors.New("bestiary: monster defeat keys and values differ in length")
	}
	return
}

func (b *Bestiary) SetMonsterDefeats(v *MonsterDefeats) (err error) {
	b.MonsterDefeatsInternal, err = MarshalOne[MonsterDefeats](v)
	return
}

// decodeOrdered decodes a JSON object keeping its member order and the exact
// text of its numbers.
func decodeOrdered(b []byte) (obj oj.OrderedObject, err error) {
	dec := oj.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	dec.UseOrderedObject()
	err = dec.Decode(&obj)
	return
}

// overlayJSON sets the members of patch in the JSON object orig, in place
// when the member exists and appended otherwise. Other members of orig are
// kept as they were.
func overlayJSON(orig []byte, patch oj.OrderedObject) ([]byte, error) {
	obj, err := decodeOrdered(orig)
	if err != nil {
		return nil, err
	}
	for _, m := range patch {
		found := false
		for i := range obj {
			if obj[i].Key == m.Key {
				obj[i].Value = m.Value
				found = true
				break
			}
		}
		if !found {
			obj = append(obj, m)
		}
	}
	return oj.Marshal(obj)
}
