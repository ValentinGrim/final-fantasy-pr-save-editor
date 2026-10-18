package inputs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Widths of the entries in the forms and lists, so that they line up whatever
// the label or the name next to them. The editors must fit the default window
// (global.WindowWidth): a grid sizes every column to its widest cell, so a long
// name in one cell would widen all of them, and the window with them.
const (
	// EntryWidth fits an ID, a count or a stat.
	EntryWidth float32 = 90
	// FloatWidth fits a position such as "-1234.000000".
	FloatWidth float32 = 130
	// WideWidth fits a time stamp.
	WideWidth float32 = 200
)

type minWidth float32

func (w minWidth) MinSize(objects []fyne.CanvasObject) (s fyne.Size) {
	for _, o := range objects {
		if o.Visible() {
			s = s.Max(o.MinSize())
		}
	}
	s.Width = fyne.Max(s.Width, float32(w))
	return
}

func (w minWidth) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(size)
	}
}

// Sized gives o a minimum width of width (or o's own MinSize if wider).
func Sized(o fyne.CanvasObject, width float32) *fyne.Container {
	return container.New(minWidth(width), o)
}

// Fixed keeps o at width, aligned left in a wider cell.
func Fixed(o fyne.CanvasObject, width float32) fyne.CanvasObject {
	return container.NewBorder(nil, nil, Sized(o, width), nil)
}

// Truncated makes l end with an ellipsis when it does not fit, instead of
// making its container as wide as its text.
func Truncated(l *widget.Label) *widget.Label {
	l.Truncation = fyne.TextTruncateEllipsis
	return l
}

// IDValue shows an ID entry followed by the name it resolves to, the name
// taking the remaining width.
func IDValue(id fyne.CanvasObject, name *widget.Label) fyne.CanvasObject {
	return container.NewBorder(nil, nil, Sized(id, EntryWidth), nil, Truncated(name))
}

// NewForm lays its objects out in pairs, label then value: the label column is
// as wide as the widest label and the value column takes the rest.
func NewForm(objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(layout.NewFormLayout(), objects...)
}

// AlignForms gives the label columns of forms made with NewForm the same width,
// call it before they are shown.
func AlignForms(forms ...*fyne.Container) {
	var width float32
	for _, f := range forms {
		for i := 0; i < len(f.Objects); i += 2 {
			width = fyne.Max(width, f.Objects[i].MinSize().Width)
		}
	}
	for _, f := range forms {
		if len(f.Objects) > 0 {
			// the form layout sizes the column to its widest label
			f.Objects[0] = Sized(f.Objects[0], width)
		}
	}
}

// ListRow shows a name followed by cells of fixed width (entries made with
// Sized, buttons): the name takes the remaining width, so the cells of the rows
// and of a header made the same way line up.
func ListRow(name fyne.CanvasObject, cells ...fyne.CanvasObject) fyne.CanvasObject {
	if l, ok := name.(*widget.Label); ok {
		Truncated(l)
	}
	return container.NewBorder(nil, nil, nil, container.NewHBox(cells...), name)
}

// Header returns a centred column title and the width of its column, see
// ColumnWidth.
func Header(title string) (*widget.Label, float32) {
	l := widget.NewLabel(title)
	l.Alignment = fyne.TextAlignCenter
	return l, ColumnWidth(title)
}

// ColumnWidth is the width of a column of entries titled title: at least
// EntryWidth, wider if the title needs it.
func ColumnWidth(title string) float32 {
	return fyne.Max(EntryWidth, widget.NewLabel(title).MinSize().Width)
}

// WithSearch puts the search panel to the right of the editor, in a split the
// user can drag. Without panel it returns the editor.
func WithSearch(editor, panel fyne.CanvasObject) fyne.CanvasObject {
	if panel == nil {
		return editor
	}
	return container.NewHSplit(editor, panel)
}
