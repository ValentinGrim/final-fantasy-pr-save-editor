package character

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/core/costs/ff6"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

type (
	Equipment struct {
		widget.BaseWidget
		game   global.Game
		inputs []*eqRow
	}
	eqRow struct {
		*inputs.IdEntry
		name string
	}
)

func NewCoreEquipment(game global.Game, c *core.Character) *Equipment {
	e := &Equipment{
		game:   game,
		inputs: make([]*eqRow, 0, len(c.Equipment.Values)+1),
	}
	e.ExtendBaseWidget(e)
	for i, j := range c.Equipment.Values {
		e.inputs = append(e.inputs, &eqRow{
			IdEntry: inputs.NewIdEntryWithDataWithHint(&j.ContentID, finder.Items),
		})
		switch i {
		case 0:
			e.inputs[i].name = "R. Hand:"
		case 1:
			e.inputs[i].name = "L. Hand:"
		case 2:
			e.inputs[i].name = "Body:"
		case 3:
			e.inputs[i].name = "Head:"
		default:
			e.inputs[i].name = "Arms:"
		}
	}
	if game.IsSix() {
		e.inputs = append(e.inputs, &eqRow{
			IdEntry: inputs.NewIdEntryWithDataWithHint(&c.Base.MagicStoneID, finder.New(ff6.Espers)),
			name:    "Esper:",
		})
	}
	return e
}

// CreateRenderer shows a form of the slots, the searches on the right.
func (e *Equipment) CreateRenderer() fyne.WidgetRenderer {
	rows := inputs.NewForm()
	for _, i := range e.inputs {
		rows.Add(widget.NewLabel(i.name))
		rows.Add(inputs.IDValue(i.ID, i.Label))
	}
	searches := []inputs.NamedSearch{{Name: "Equipment", Search: inputs.GetSearches().Equipment}}
	if e.game.IsSix() {
		searches = append(searches, inputs.NamedSearch{Name: "Espers", Search: inputs.NewSearch(ff6.Espers)})
	}
	searches = append(searches, inputs.NamedSearch{Name: "Items", Search: inputs.GetSearches().Items})
	return widget.NewSimpleRenderer(inputs.WithSearch(container.NewVScroll(rows), inputs.SearchPanels(searches...)))
}
