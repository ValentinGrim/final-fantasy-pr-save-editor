package ff4

import (
	"slices"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

func ids(l []*save.Ability) []int {
	out := make([]int, len(l))
	for i, a := range l {
		out[i] = a.AbilityID
	}
	return out
}

// button returns the first button in o, the delete button of a row.
func button(o fyne.CanvasObject) *widget.Button {
	switch o := o.(type) {
	case *widget.Button:
		return o
	case *fyne.Container:
		for _, c := range o.Objects {
			if b := button(c); b != nil {
				return b
			}
		}
	}
	return nil
}

func deleteRow(t *testing.T, e *Abilities, row int) {
	t.Helper()
	test.Tap(button(e.abilities.Objects[row]))
}

func TestAddRemoveRows(t *testing.T) {
	test.NewTempApp(t)
	finder.Load(global.Four, nil)
	inputs.Load(global.Four)

	c := &core.Character{
		Base:      &save.OwnedCharacter{},
		Abilities: []*save.Ability{{AbilityID: 10}, {AbilityID: 20}, {AbilityID: 30}},
	}
	e := NewAbilities(c)
	w := test.NewWindow(e)
	defer w.Close()

	// Delete the first visual row twice: each tap must remove the ability
	// shown in that row, never a stale index.
	deleteRow(t, e, 0)
	deleteRow(t, e, 0)
	if got := ids(c.Abilities); !slices.Equal(got, []int{30}) {
		t.Fatalf("abilities = %v; want [30]", got)
	}
	if n := len(e.abilities.Objects); n != 1 {
		t.Fatalf("%d rows shown; want 1", n)
	}

	test.Tap(e.add)
	test.Tap(e.add)
	if n := len(e.abilities.Objects); n != 3 || len(c.Abilities) != 3 {
		t.Fatalf("after 2 adds: %d rows, %d abilities; want 3 and 3", n, len(c.Abilities))
	}
	first := c.Abilities[1]
	deleteRow(t, e, 2)
	if len(c.Abilities) != 2 || c.Abilities[1] != first || len(e.abilities.Objects) != 2 {
		t.Fatalf("deleting the 2nd placeholder: abilities %v, rows %d", ids(c.Abilities), len(e.abilities.Objects))
	}

	// Unfilled rows are not saved.
	if got := ids(c.AbilitiesToSave()); !slices.Equal(got, []int{30}) {
		t.Fatalf("AbilitiesToSave = %v; want [30]", got)
	}

	// Removing everything, then once more on an empty list, must not panic.
	deleteRow(t, e, 1)
	deleteRow(t, e, 0)
	if len(c.Abilities) != 0 || len(e.abilities.Objects) != 0 {
		t.Fatalf("abilities %v, rows %d; want none", ids(c.Abilities), len(e.abilities.Objects))
	}
}
