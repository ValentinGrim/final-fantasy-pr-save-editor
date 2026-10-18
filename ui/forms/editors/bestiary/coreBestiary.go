package bestiary

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

type (
	Bestiary struct {
		widget.BaseWidget
		monstersDefeated []*inputs.IdCountEntry
		// scenarioFlags         []binding.Bool
		search                *inputs.Search
		killAllMonstersButton *widget.Button
		addRowButton          *widget.Button
		rows                  *fyne.Container
		// widths of the ID and count columns, set with the headers
		idWidth, countWidth float32
	}
)

func NewCore(b *core.Bestiary, finder finder.Find, search *inputs.Search) *Bestiary {
	e := &Bestiary{
		monstersDefeated: make([]*inputs.IdCountEntry, len(b.MonsterDefeats)),
		// scenarioFlags:         make([]binding.Bool, len(b.ScenarioFlags)),
		search: search,
		rows:   container.NewVBox(),
	}
	e.ExtendBaseWidget(e)
	e.killAllMonstersButton = widget.NewButton("Kill All Monsters", func() {
		ids := make([]int, 0, len(*search.Entries))
		for _, i := range *search.Entries {
			ids = append(ids, i.Value)
		}
		b.DefeatAll(ids)

		e.monstersDefeated = make([]*inputs.IdCountEntry, len(b.MonsterDefeats))
		for i, d := range b.MonsterDefeats {
			e.monstersDefeated[i] = inputs.NewIdCountEntryWithDataWithHint(&d.ID, &d.Count, finder)
		}
		e.buildRows()
	})
	e.addRowButton = widget.NewButton("+", func() {
		d := b.Add()
		row := inputs.NewIdCountEntryWithDataWithHint(&d.ID, &d.Count, finder)
		e.monstersDefeated = append(e.monstersDefeated, row)
		e.addRow(row)
	})

	for i, d := range b.MonsterDefeats {
		e.monstersDefeated[i] = inputs.NewIdCountEntryWithDataWithHint(&d.ID, &d.Count, finder)
	}
	// for i, f := range b.ScenarioFlags {
	// 	j := binding.NewBool()
	// 	if f == 1 {
	// 		_ = j.Set(true)
	// 	}
	// 	e.scenarioFlags[i] = j
	// }
	return e
}

// CreateRenderer lists one monster per row: name, ID and count, the search on
// the right.
func (b *Bestiary) CreateRenderer() fyne.WidgetRenderer {
	var id, c *widget.Label
	id, b.idWidth = inputs.Header("Monster ID")
	c, b.countWidth = inputs.Header("Count")
	headers := container.NewPadded(inputs.ListRow(container.NewStack(), inputs.Sized(id, b.idWidth), inputs.Sized(c, b.countWidth)))

	b.buildRows()

	return widget.NewSimpleRenderer(inputs.WithSearch(
		container.NewBorder(
			container.NewHBox(b.killAllMonstersButton, b.addRowButton), nil, nil, nil,
			// middle
			container.NewBorder(
				headers, nil, nil, nil,
				container.NewVScroll(b.rows))),
		b.search.Panel()))
}

func (b *Bestiary) buildRows() {
	b.rows.RemoveAll()
	for _, d := range b.monstersDefeated {
		b.addRow(d)
	}
}

func (b *Bestiary) addRow(d *inputs.IdCountEntry) {
	b.rows.Add(container.NewPadded(inputs.ListRow(d.Label, inputs.Sized(d.ID, b.idWidth), inputs.Sized(d.Count, b.countWidth))))
}
