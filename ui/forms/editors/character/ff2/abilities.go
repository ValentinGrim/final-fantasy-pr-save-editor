package ff2

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/models/core"
	cff2 "pixel-remastered-save-editor/models/core/costs/ff2"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

const (
	abilityID   = "Ability ID"
	skillPoints = "Skill Points"
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

// addRow edits one ability. For a spell the ability ID holds the level and
// the second column the total skill points, so each follows the other: a new
// level moves the points to that level's minimum, new points move the spell
// to the level they reach.
func (e *Abilities) addRow(a *save.Ability) {
	pts := inputs.NewIntEntryWithData(&a.SkillLevel)
	id := inputs.NewIdEntryWithDataWithHintWithChange(&a.AbilityID, finder.Abilities, func(s string) {
		if j, err := strconv.Atoi(s); err == nil {
			if lvl, ok := cff2.SpellLevel(j); ok && cff2.LevelForPoints(pts.Int()) != lvl {
				pts.SetInt(cff2.PointsForLevel(lvl))
			}
		}
	})
	idEntry := id.ID.(*inputs.IntEntry)
	pts.OnChanged = func(string) {
		j := idEntry.Int()
		if _, ok := cff2.SpellLevel(j); ok {
			if k := cff2.SpellAtLevel(j, cff2.LevelForPoints(pts.Int())); k != j {
				idEntry.SetInt(k)
			}
		}
	}
	var g fyne.CanvasObject
	g = inputs.ListRow(id.Label,
		inputs.Sized(id.ID, inputs.ColumnWidth(abilityID)),
		inputs.Sized(pts, inputs.ColumnWidth(skillPoints)),
		widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			e.c.RemoveAbility(a)
			e.abilities.Remove(g)
		}))
	e.abilities.Add(g)
}

// CreateRenderer lists the abilities, the search on the right.
func (e *Abilities) CreateRenderer() fyne.WidgetRenderer {
	id, idWidth := inputs.Header(abilityID)
	pts, ptsWidth := inputs.Header(skillPoints)
	left := container.NewBorder(
		inputs.ListRow(container.NewStack(), inputs.Sized(id, idWidth), inputs.Sized(pts, ptsWidth), e.add),
		nil, nil, nil,
		container.NewVScroll(e.abilities))
	return widget.NewSimpleRenderer(inputs.WithSearch(left, inputs.GetSearches().Abilities.Panel()))
}
