package ff1

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

const abilityID = "Ability ID"

type (
	Abilities struct {
		widget.BaseWidget
		c         *core.Character
		abilities *fyne.Container
		body      *fyne.Container
		add       *widget.Button
	}
)

func NewAbilities(c *core.Character) *Abilities {
	e := &Abilities{
		c:         c,
		body:      container.NewVBox(),
		abilities: container.NewVBox(),
	}
	e.ExtendBaseWidget(e)

	e.add = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		a := &save.Ability{}
		e.c.AddAbility(a)
		e.addRow(a)
	})

	for i, asd := range e.c.AbilitySlotData {
		e.body.Add(widget.NewLabel(fmt.Sprintf("%d", i+1)))
		for _, a := range asd.SlotInfo.Values {
			r := inputs.NewIdEntryWithDataWithHint(&a.AbilityID, finder.Abilities)
			e.body.Add(container.NewPadded(inputs.ListRow(r.Label, inputs.Sized(r.ID, inputs.EntryWidth))))
		}
	}
	for _, a := range e.c.Abilities {
		e.addRow(a)
	}
	return e
}

func (e *Abilities) addRow(a *save.Ability) {
	var g fyne.CanvasObject
	r := inputs.NewIdEntryWithDataWithHint(&a.AbilityID, finder.Abilities)
	g = inputs.ListRow(r.Label, inputs.Sized(r.ID, inputs.ColumnWidth(abilityID)), widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		e.c.RemoveAbility(a)
		e.abilities.Remove(g)
	}))
	e.abilities.Add(g)
}

// CreateRenderer shows the abilities above the ability slots, the search on
// the right.
func (e *Abilities) CreateRenderer() fyne.WidgetRenderer {
	id, width := inputs.Header(abilityID)
	var editor fyne.CanvasObject = container.NewBorder(
		inputs.ListRow(container.NewStack(), inputs.Sized(id, width), e.add),
		nil, nil, nil,
		container.NewVScroll(e.abilities))
	if len(e.body.Objects) > 0 {
		editor = container.NewVSplit(editor, container.NewVScroll(e.body))
	}
	return widget.NewSimpleRenderer(inputs.WithSearch(editor, inputs.GetSearches().Abilities.Panel()))
}
