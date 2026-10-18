package ff6

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"pixel-remastered-save-editor/models/core"
	cff6 "pixel-remastered-save-editor/models/core/costs/ff6"
	"pixel-remastered-save-editor/save"
)

func count(c *core.Character, id int) (n, lvl int) {
	for _, a := range c.Abilities {
		if a.AbilityID == id {
			n++
			lvl = a.SkillLevel
		}
	}
	return
}

func TestReselectDoesNotDuplicateMagic(t *testing.T) {
	test.NewTempApp(t)
	c := &core.Character{Base: &save.OwnedCharacter{JobID: 1}}
	NewAbilities(c)
	n := len(c.Abilities)
	if n != len(cff6.MagicSorted) {
		t.Fatalf("%d abilities; want %d", n, len(cff6.MagicSorted))
	}
	NewAbilities(c)
	NewAbilities(c)
	if len(c.Abilities) != n {
		t.Fatalf("reselecting grew abilities from %d to %d", n, len(c.Abilities))
	}
}

func TestToggleAbilities(t *testing.T) {
	test.NewTempApp(t)
	known := cff6.Blitzes[0].Value
	c := &core.Character{
		Base:      &save.OwnedCharacter{JobID: 6}, // Sabin
		Abilities: []*save.Ability{{AbilityID: known, SkillLevel: 100}},
	}
	e := NewAbilities(c)
	before := len(c.Abilities)
	w := test.NewWindow(e)
	defer w.Close()

	if !e.abilities[0].Checked || e.abilities[1].Checked {
		t.Fatal("initial check state does not match the save")
	}
	if len(c.Abilities) != before {
		t.Fatal("building the editor changed the abilities")
	}

	// Tap a Blitz on then off: one entry, level 0.
	b := e.abilities[1]
	test.Tap(b.Check)
	test.Tap(b.Check)
	if n, lvl := count(c, b.id); n != 1 || lvl != 0 {
		t.Fatalf("after on/off: %d entries, level %d; want 1, 0", n, lvl)
	}

	e.toggleAllAbilities(true)
	for _, i := range e.abilities {
		if n, lvl := count(c, i.id); n != 1 || lvl != 100 || !i.Checked {
			t.Fatalf("All: %s has %d entries, level %d, checked %v", i.Text, n, lvl, i.Checked)
		}
	}
	e.toggleAllAbilities(false)
	for _, i := range e.abilities {
		if n, lvl := count(c, i.id); n != 1 || lvl != 0 || i.Checked {
			t.Fatalf("None: %s has %d entries, level %d, checked %v", i.Text, n, lvl, i.Checked)
		}
	}
	if want := before + len(cff6.Blitzes) - 1; len(c.Abilities) != want {
		t.Fatalf("%d abilities; want %d", len(c.Abilities), want)
	}
}
