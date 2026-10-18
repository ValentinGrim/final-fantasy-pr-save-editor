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

func (b *Bestiary) CreateRenderer() fyne.WidgetRenderer {
	const columns = 6
	id := widget.NewLabel("Monster ID")
	id.Alignment = fyne.TextAlignCenter
	c := widget.NewLabel("Count")
	c.Alignment = fyne.TextAlignCenter

	headers := container.NewPadded(container.NewGridWithColumns(columns, container.NewStack(), id, c, container.NewStack()))

	b.buildRows()

	top := container.NewGridWithColumns(6)
	top.Add(b.killAllMonstersButton)
	top.Add(container.NewStack())
	top.Add(b.addRowButton)

	return widget.NewSimpleRenderer(
		container.NewBorder(
			top, nil, nil, nil,
			// middle
			container.NewBorder(
				headers, nil, nil, nil,
				container.NewVScroll(b.rows))))
}

func (b *Bestiary) buildRows() {
	b.rows.RemoveAll()
	for _, d := range b.monstersDefeated {
		b.addRow(d)
	}
}

func (b *Bestiary) addRow(d *inputs.IdCountEntry) {
	const columns = 6
	entry := d
	g := container.NewGridWithColumns(columns, entry.Label, entry.ID, entry.Count)
	b.rows.Add(container.NewPadded(g))
}
