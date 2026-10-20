package ff2

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

// rowEntries returns the ID and points entries of each row, in order.
func rowEntries(e *Abilities) (rows [][2]*inputs.IntEntry) {
	var walk func(o fyne.CanvasObject, found *[]*inputs.IntEntry)
	walk = func(o fyne.CanvasObject, found *[]*inputs.IntEntry) {
		switch v := o.(type) {
		case *inputs.IntEntry:
			*found = append(*found, v)
		case *fyne.Container:
			for _, k := range v.Objects {
				walk(k, found)
			}
		}
	}
	for _, r := range e.abilities.Objects {
		var found []*inputs.IntEntry
		walk(r, &found)
		rows = append(rows, [2]*inputs.IntEntry{found[0], found[1]})
	}
	return
}

func TestSpellLevelFollowsPoints(t *testing.T) {
	test.NewTempApp(t)
	finder.Load(global.Two, nil)
	const fire1, fire8, fire3, item = 5, 12, 7, 711
	fire := &save.Ability{AbilityID: fire1, ContentID: fire1 + 207}
	other := &save.Ability{AbilityID: item, ContentID: 937, SkillLevel: 7}
	e := NewAbilities(&core.Character{Abilities: []*save.Ability{fire, other}})
	rows := rowEntries(e)
	if len(rows) != 2 {
		t.Fatalf("%d rows; want 2", len(rows))
	}
	id, pts := rows[0][0], rows[0][1]

	pts.SetText("251")
	if id.Int() != fire8 || fire.AbilityID != fire8 || fire.SkillLevel != 251 {
		t.Fatalf("251 points: entry %d, ability %d with %d points; want Fire 8 (%d)", id.Int(), fire.AbilityID, fire.SkillLevel, fire8)
	}

	// A level change moves the points to the new level's minimum.
	id.SetText("7")
	if pts.Int() != 45 || fire.AbilityID != fire3 || fire.SkillLevel != 45 {
		t.Fatalf("Fire 3: entry %d points, ability %d with %d points; want 45", pts.Int(), fire.AbilityID, fire.SkillLevel)
	}

	// Points within the level leave it alone.
	pts.SetText("69")
	if fire.AbilityID != fire3 {
		t.Fatalf("69 points moved Fire 3 to %d", fire.AbilityID)
	}

	// Not a spell: neither column drives the other.
	rows[1][1].SetText("500")
	if other.AbilityID != item || other.SkillLevel != 500 {
		t.Fatalf("ability %d: id %d, points %d", item, other.AbilityID, other.SkillLevel)
	}
}
