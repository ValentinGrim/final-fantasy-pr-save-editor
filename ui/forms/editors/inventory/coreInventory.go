package inventory

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

type (
	Inventory struct {
		widget.BaseWidget
		items     []*inputs.IdCountEntry
		resetSort binding.Bool
		search    *inputs.Search
	}
)

func NewCore(inv *core.Inventory, finder finder.Find, search *inputs.Search) *Inventory {
	e := &Inventory{
		items:     make([]*inputs.IdCountEntry, len(inv.Rows)),
		resetSort: binding.BindBool(&inv.ResetSortOrder),
		search:    search,
	}
	e.ExtendBaseWidget(e)

	for i, item := range inv.Rows {
		e.items[i] = inputs.NewIdCountEntryWithDataWithHint(&item.ContentID, &item.Count, finder)
	}
	return e
}

// CreateRenderer lists one item per row: name, ID and count, the search on the
// right.
func (e *Inventory) CreateRenderer() fyne.WidgetRenderer {
	id, idWidth := inputs.Header("Item ID")
	count, countWidth := inputs.Header("Count")
	reset := widget.NewCheckWithData("Reset Sort", e.resetSort)
	headers := container.NewPadded(inputs.ListRow(reset, inputs.Sized(id, idWidth), inputs.Sized(count, countWidth)))

	rows := container.NewVBox()
	for _, entry := range e.items {
		rows.Add(container.NewPadded(inputs.ListRow(entry.Label, inputs.Sized(entry.ID, idWidth), inputs.Sized(entry.Count, countWidth))))
	}
	return widget.NewSimpleRenderer(inputs.WithSearch(
		container.NewBorder(headers, nil, nil, nil, container.NewVScroll(rows)),
		e.search.Panel()))
}
