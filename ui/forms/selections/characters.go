package selections

import (
	"cmp"
	"fmt"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/models/core"
	cff6 "pixel-remastered-save-editor/models/core/costs/ff6"
	"pixel-remastered-save-editor/ui/forms/editors/character"
	"pixel-remastered-save-editor/ui/forms/editors/character/ff1"
	"pixel-remastered-save-editor/ui/forms/editors/character/ff2"
	"pixel-remastered-save-editor/ui/forms/editors/character/ff4"
	"pixel-remastered-save-editor/ui/forms/editors/character/ff5"
	"pixel-remastered-save-editor/ui/forms/editors/character/ff6"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

type (
	Characters struct {
		widget.BaseWidget
		top    *fyne.Container
		middle *fyne.Container
	}
)

func NewCharacters(save *core.Save) *Characters {
	s := &Characters{
		top:    container.NewHBox(),
		middle: container.NewStack(),
	}
	s.ExtendBaseWidget(s)
	choices := characterChoices(save.Game(), save.Characters.All())
	labels := make([]string, len(choices))
	ids := make(map[string]int, len(choices))
	for i, ch := range choices {
		labels[i] = ch.label
		ids[ch.label] = ch.id
	}
	s.top.Add(inputs.NewLabeledEntry("Character:", widget.NewSelect(labels, func(label string) {
		s.middle.RemoveAll()
		id, ok := ids[label]
		c, found := save.Characters.GetByID(id)
		if !ok || !found || c == nil || c.Base == nil {
			s.middle.Add(widget.NewLabel(fmt.Sprintf("Character %q was not found in this save.", label)))
			return
		}
		tabs := container.NewAppTabs(
			container.NewTabItem("Stats", character.NewCoreStats(c)),
			container.NewTabItem("Abilities", s.abilities(save.Game(), c)),
			container.NewTabItem("Equipment", character.NewCoreEquipment(save.Game(), c)),
			container.NewTabItem("Commands", s.commands(save.Game(), c)),
		)
		if save.Game().IsFive() {
			tabs.Append(container.NewTabItem("Jobs", NewFF5Jobs(c)))
		}
		s.middle.Add(tabs)
	}), 2))
	return s
}

func (s *Characters) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewBorder(s.top, nil, nil, nil, s.middle))
}

func (s *Characters) abilities(game global.Game, c *core.Character) (a fyne.CanvasObject) {
	if game == global.One || game == global.Three {
		a = ff1.NewAbilities(c)
	} else if game == global.Two {
		a = ff2.NewAbilities(c)
	} else if game == global.Four {
		a = ff4.NewAbilities(c)
	} else if game == global.Five {
		a = ff5.NewAbilities(c)
	} else { // ff6
		a = ff6.NewAbilities(c)
	}
	return
}

func (s *Characters) commands(game global.Game, c *core.Character) (a fyne.CanvasObject) {
	if game == global.Five {
		a = ff5.NewCommands(c)
	} else {
		a = character.NewCoreCommands(c)
	}
	return
}

type characterChoice struct {
	label string
	id    int
}

// characterChoices lists the characters of the save, identified by ID, with a
// unique label each. Labels use the name stored in the save, which the player
// may have changed (and which is localized). For FF6 the characters follow
// the order of cff6.Characters and the English name is used when the save has
// none, or shown next to the save's name when it differs.
func characterChoices(game global.Game, all []*core.Character) []characterChoice {
	var (
		choices = make([]characterChoice, 0, len(all))
		byID    = make(map[int]*core.Character, len(all))
		listed  = make(map[int]bool)
	)
	for _, c := range all {
		if c != nil && c.Base != nil {
			if _, dup := byID[c.Base.ID]; !dup {
				byID[c.Base.ID] = c
			}
		}
	}
	label := func(c *core.Character, def string) string {
		switch name := c.Base.Name; {
		case name == "" && def == "":
			return fmt.Sprintf("Character %d", c.Base.ID)
		case name == "":
			return def
		case def == "" || name == def:
			return name
		default:
			return fmt.Sprintf("%s (%s)", name, def)
		}
	}
	if game.IsSix() {
		for _, nv := range cff6.Characters {
			if c, ok := byID[nv.Value]; ok && !listed[nv.Value] {
				listed[nv.Value] = true
				choices = append(choices, characterChoice{label: label(c, nv.Name), id: nv.Value})
			}
		}
	}
	var rest []characterChoice
	for id, c := range byID {
		if !listed[id] {
			rest = append(rest, characterChoice{label: label(c, ""), id: id})
		}
	}
	slices.SortFunc(rest, func(a, b characterChoice) int {
		return cmp.Or(cmp.Compare(a.label, b.label), cmp.Compare(a.id, b.id))
	})
	choices = append(choices, rest...)

	// The player can give several characters the same name, and FF6 has two
	// Mogs: add the ID to every label that is not unique.
	count := make(map[string]int, len(choices))
	for _, ch := range choices {
		count[ch.label]++
	}
	for i, ch := range choices {
		if count[ch.label] > 1 {
			choices[i].label = fmt.Sprintf("%s [#%d]", ch.label, ch.id)
		}
	}
	return choices
}
