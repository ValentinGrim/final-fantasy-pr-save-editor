package ff2

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

const (
	abilityID  = "Ability ID"
	skillLevel = "Skill Level"
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
	lvl := inputs.NewIntEntryWithData(&a.SkillLevel)
	id := inputs.NewIdEntryWithDataWithHintWithChange(&a.AbilityID, finder.Abilities, func(s string) {
		if j, err := strconv.Atoi(s); err == nil && j > 0 {
			if namelvl, ok := finder.Abilities(j); ok {
				l, _ := strconv.Atoi(strings.Split(namelvl, " ")[1])
				var k int
				if l <= 5 {
					k = l * 25
				} else if l <= 10 {
					k = (l-5)*50 + 100
				} else {
					k = (l-10)*75 + 300
				}
				if k > l {
					lvl.Entry.SetText(strconv.Itoa(k))
				}
			}
		}
	})
	var g fyne.CanvasObject
	g = inputs.ListRow(id.Label,
		inputs.Sized(id.ID, inputs.ColumnWidth(abilityID)),
		inputs.Sized(lvl, inputs.ColumnWidth(skillLevel)),
		widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			e.c.RemoveAbility(a)
			e.abilities.Remove(g)
		}))
	e.abilities.Add(g)
}

// CreateRenderer lists the abilities, the search on the right.
func (e *Abilities) CreateRenderer() fyne.WidgetRenderer {
	id, idWidth := inputs.Header(abilityID)
	lvl, lvlWidth := inputs.Header(skillLevel)
	left := container.NewBorder(
		inputs.ListRow(container.NewStack(), inputs.Sized(id, idWidth), inputs.Sized(lvl, lvlWidth), e.add),
		nil, nil, nil,
		container.NewVScroll(e.abilities))
	return widget.NewSimpleRenderer(inputs.WithSearch(left, inputs.GetSearches().Abilities.Panel()))
}
