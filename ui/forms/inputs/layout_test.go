package inputs

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/models"
)

// A long name must not widen the editors: the window grows to their MinSize.
func TestLongNamesDoNotWiden(t *testing.T) {
	test.NewTempApp(t)
	long := strings.Repeat("Castle Fynn - Secret Passage ", 10)
	const max = 300

	id := NewIntEntry()
	if w := IDValue(id, widget.NewLabel(long)).MinSize().Width; w > max {
		t.Errorf("IDValue MinSize width = %v", w)
	}
	if w := ListRow(widget.NewLabel(long), Sized(NewIntEntry(), EntryWidth)).MinSize().Width; w > max {
		t.Errorf("ListRow MinSize width = %v", w)
	}
	if w := NewForm(widget.NewLabel("Map ID"), IDValue(NewIntEntry(), widget.NewLabel(long))).MinSize().Width; w > max {
		t.Errorf("form MinSize width = %v", w)
	}

	s := NewSearch([]models.NameValue{models.NewValueName(1, long)})
	s.Input.SetText("castle") // fills the results
	if w := s.Panel().MinSize().Width; w > max {
		t.Errorf("search panel MinSize width = %v", w)
	}
	if w := SearchPanels(NamedSearch{"A", s}, NamedSearch{"B", s}).MinSize().Width; w > max {
		t.Errorf("search tabs MinSize width = %v", w)
	}
	if p := SearchPanels(NamedSearch{"A", nil}); p != nil {
		t.Errorf("SearchPanels without search = %v; want nil", p)
	}
	if w := WithSearch(container.NewStack(), s.Panel()).MinSize().Width; w > max {
		t.Errorf("WithSearch MinSize width = %v", w)
	}
}

func TestSizedKeepsWidth(t *testing.T) {
	test.NewTempApp(t)
	if w := Sized(NewIntEntry(), EntryWidth).MinSize().Width; w != EntryWidth {
		t.Errorf("Sized MinSize width = %v; want %v", w, EntryWidth)
	}
	wide := widget.NewLabel(strings.Repeat("x", 50))
	if w, want := Sized(wide, EntryWidth).MinSize().Width, wide.MinSize().Width; w != want {
		t.Errorf("Sized MinSize width = %v; want the label's %v", w, want)
	}
}
