package editors

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

type (
	Party struct {
		widget.BaseWidget
		current *fyne.Container
		parties *fyne.Container
	}
)

func NewCoreParty(p *core.Party, parties *core.Parties) *Party {
	e := &Party{
		BaseWidget: widget.BaseWidget{},
		current:    container.NewVBox(),
		parties:    container.NewVBox(),
	}
	e.ExtendBaseWidget(e)
	e.current.Add(widget.NewLabel("Current Party"))
	for _, m := range p.Members {
		func(m *save.CorpsSlotInfo) {
			j := inputs.NewIdEntryWithDataWithHint(&m.CharacterID, finder.Characters)
			e.current.Add(container.NewPadded(inputs.ListRow(j.Label, inputs.Sized(j.ID, inputs.EntryWidth))))
		}(m)
	}
	for i, j := range parties.Parties {
		if i > 0 {
			e.parties.Add(widget.NewSeparator())
		}
		e.parties.Add(widget.NewLabel(fmt.Sprintf("Party %d", i+1)))
		for _, m := range j {
			in := inputs.NewIdEntryWithDataWithHint(&m.CharacterID, finder.Characters)
			e.parties.Add(container.NewPadded(inputs.ListRow(in.Label, inputs.Sized(in.ID, inputs.EntryWidth))))
		}
	}
	return e
}

// CreateRenderer shows the current party above the other parties, the
// character search on the right.
func (e *Party) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(inputs.WithSearch(
		container.NewVScroll(container.NewVBox(e.current, widget.NewSeparator(), e.parties)),
		inputs.GetSearches().Characters.Panel()))
}
