package ff4

import (
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
		add       *widget.Button
	}
)

func NewAbilities(c *core.Character) *Abilities {
	e := &Abilities{
		c:         c,
		abilities: container.NewVBox(),
	}
	e.ExtendBaseWidget(e)
	e.add = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		a := &save.Ability{}
		e.c.AddAbility(a)
		e.addRow(a)
	})
	for _, a := range e.c.Abilities {
		e.addRow(a)
	}
	return e
}

func (e *Abilities) addRow(a *save.Ability) {
	r := inputs.NewIdEntryWithDataWithHint(&a.AbilityID, finder.Abilities)
	var g fyne.CanvasObject
	g = inputs.ListRow(r.Label, inputs.Sized(r.ID, inputs.ColumnWidth(abilityID)), widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		e.c.RemoveAbility(a)
		e.abilities.Remove(g)
	}))
	e.abilities.Add(g)
}

// CreateRenderer lists the abilities, the search on the right.
func (e *Abilities) CreateRenderer() fyne.WidgetRenderer {
	id, width := inputs.Header(abilityID)
	left := container.NewBorder(
		inputs.ListRow(container.NewStack(), inputs.Sized(id, width), e.add),
		nil, nil, nil,
		container.NewVScroll(e.abilities))
	return widget.NewSimpleRenderer(inputs.WithSearch(left, inputs.GetSearches().Abilities.Panel()))
}
