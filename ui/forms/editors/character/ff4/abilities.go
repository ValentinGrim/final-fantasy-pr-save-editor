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
	var g *fyne.Container
	g = container.NewGridWithColumns(3, r.Label, r.ID, widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		e.c.RemoveAbility(a)
		e.abilities.Remove(g)
	}))
	e.abilities.Add(g)
}

func (e *Abilities) CreateRenderer() fyne.WidgetRenderer {
	search := inputs.GetSearches().Abilities
	left := container.NewBorder(
		container.NewGridWithColumns(3, container.NewStack(), widget.NewLabel("Ability ID"), e.add),
		nil, nil, nil,
		container.NewVScroll(e.abilities))
	right := container.NewGridWithColumns(2, search.Fields(), search.Filter())
	return widget.NewSimpleRenderer(container.NewBorder(nil, nil, left, right))
}
