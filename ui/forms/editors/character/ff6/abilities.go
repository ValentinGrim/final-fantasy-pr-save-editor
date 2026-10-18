package ff6

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/models"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/core/costs/ff6"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

type (
	Abilities struct {
		widget.BaseWidget
		c           *core.Character
		magicInputs []*inputs.IdEntry
		abilities   []*check
	}
	check struct {
		*widget.Check
		id int
	}
)

func NewAbilities(c *core.Character) *Abilities {
	e := &Abilities{
		c:           c,
		magicInputs: make([]*inputs.IdEntry, len(ff6.MagicSorted)),
	}
	e.ExtendBaseWidget(e)

	e.createMagics()

	switch c.Base.JobID {
	case 3: // Cyan
		e.createAbilities(ff6.Bushidos)
	case 6: // Sabin
		e.createAbilities(ff6.Blitzes)
	case 8: // Stago
		e.createAbilities(ff6.Lores)
	case 11: // Mog
		e.createAbilities(ff6.Dances)
	case 12: // Gau
		e.createAbilities(ff6.Rages)
	}
	return e
}

// createMagics binds one entry per spell. A spell the character does not have
// yet gets a skill level 0 entry; GetAbility finds it the next time the
// character is selected, so it is created only once.
func (e *Abilities) createMagics() {
	for i, m := range ff6.MagicSorted {
		a, ok := e.c.GetAbility(m.Value)
		if !ok {
			a = &save.Ability{AbilityID: m.Value}
			e.c.AddAbility(a)
		}
		e.magicInputs[i] = inputs.NewLabeledIntEntryWithData(m.Name, &a.SkillLevel)
	}
}

func (e *Abilities) createAbilities(a []models.NameValue) {
	for _, i := range a {
		c := &check{
			Check: widget.NewCheck(i.Name, nil),
			id:    i.Value,
		}
		// Set the initial state before OnChanged so building the editor does
		// not modify the character.
		c.Checked = e.knows(i.Value)
		c.OnChanged = func(checked bool) {
			e.setAbility(c.id, checked)
		}
		e.abilities = append(e.abilities, c)
	}
}

func (e *Abilities) CreateRenderer() fyne.WidgetRenderer {
	ml := container.NewVBox()
	mm := container.NewVBox()
	mr := container.NewVBox()
	for i, j := range e.magicInputs {
		c := container.NewGridWithColumns(2, j.Label, j.ID)
		if k := i % 3; k == 0 {
			ml.Add(c)
		} else if k == 1 {
			mm.Add(c)
		} else {
			mr.Add(c)
		}
	}

	left := container.NewBorder(
		container.NewHBox(
			widget.NewButton("All", func() { e.toggleAllMagic(true) }),
			widget.NewButton("None", func() { e.toggleAllMagic(false) })),
		nil, nil, nil,
		container.NewVScroll(container.NewGridWithColumns(3, ml, mm, mr)))

	var right *fyne.Container
	if len(e.abilities) == 0 {
		right = container.NewStack()
	} else {
		a := container.NewVBox()
		for _, i := range e.abilities {
			a.Add(i.Check)
		}
		right = container.NewBorder(
			container.NewHBox(
				widget.NewButton("All", func() { e.toggleAllAbilities(true) }),
				widget.NewButton("None", func() { e.toggleAllAbilities(false) })),
			nil, nil, nil,
			container.NewVScroll(a))
	}

	return widget.NewSimpleRenderer(container.NewBorder(nil, nil, left, right))
}

func (e *Abilities) toggleAllMagic(max bool) {
	v := "0"
	if max {
		v = "100"
	}
	for _, i := range e.magicInputs {
		i.ID.SetText(v)
	}
}

func (e *Abilities) toggleAllAbilities(checked bool) {
	for _, i := range e.abilities {
		e.setAbility(i.id, checked)
		// Update the box without SetChecked: it would call OnChanged and
		// apply the change a second time.
		i.Checked = checked
		i.Refresh()
	}
}

// setAbility marks the ability as learned (skill level 100) or not (0).
func (e *Abilities) setAbility(id int, checked bool) {
	lvl := 0
	if checked {
		lvl = 100
	}
	if ab, ok := e.c.GetAbility(id); ok {
		ab.SkillLevel = lvl
	} else {
		e.c.AddAbility(&save.Ability{
			AbilityID:  id,
			ContentID:  id + 330,
			SkillLevel: lvl,
		})
	}
}

func (e *Abilities) knows(id int) bool {
	if a, ok := e.c.GetAbility(id); ok {
		return a.SkillLevel >= 100
	}
	return false
}
