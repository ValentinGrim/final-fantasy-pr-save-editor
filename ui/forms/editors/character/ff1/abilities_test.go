package ff1

import (
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

func TestAddedAbilityIsKept(t *testing.T) {
	test.NewTempApp(t)
	finder.Load(global.One, nil)
	inputs.Load(global.One)

	c := &core.Character{
		Base:      &save.OwnedCharacter{},
		Abilities: []*save.Ability{{AbilityID: 5, ContentID: 213}},
	}
	e := NewAbilities(c)
	w := test.NewWindow(e)
	defer w.Close()

	test.Tap(e.add)
	if len(c.Abilities) != 2 || len(e.abilities.Objects) != 2 {
		t.Fatalf("after add: %d abilities, %d rows; want 2, 2", len(c.Abilities), len(e.abilities.Objects))
	}
	added := c.Abilities[1]
	added.AbilityID = 7
	if got := c.AbilitiesToSave(); len(got) != 2 || got[1] != added {
		t.Fatal("added ability is not saved")
	}

	// Delete the original row: the added ability must stay.
	test.Tap(button(e.abilities.Objects[0]))
	if len(c.Abilities) != 1 || c.Abilities[0] != added || len(e.abilities.Objects) != 1 {
		t.Fatalf("after delete: %d abilities, %d rows", len(c.Abilities), len(e.abilities.Objects))
	}
}
