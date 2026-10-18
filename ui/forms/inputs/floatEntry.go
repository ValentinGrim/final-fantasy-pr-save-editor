package inputs

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

type (
	// FloatEntry accepts digits with an optional leading '-' and one decimal
	// separator. A typed ',' is shown as '.', a pasted one is read as '.'.
	FloatEntry struct {
		widget.Entry
	}
	FloatEntryBinding struct {
		s binding.String
		i binding.Float
	}
)

func NewFloatEntry() *FloatEntry {
	entry := &FloatEntry{}
	entry.ExtendBaseWidget(entry)
	entry.Entry.Validator = nil
	return entry
}

func NewFloatEntryWithBinding(data FloatEntryBinding) *FloatEntry {
	entry := NewFloatEntry()
	entry.Bind(data.s)
	entry.Entry.Validator = nil
	return entry
}

func NewFloatEntryWithData(f *float64) *FloatEntry {
	entry := NewFloatEntry()
	data := NewFloatEntryBinding(f)
	entry.Bind(data.s)
	entry.Entry.Validator = nil
	return entry
}

func (e *FloatEntry) Float() float64 {
	f, ok := parseFloat(e.Text)
	if !ok {
		f = 0
	}
	return f
}

func (e *FloatEntry) SetFloat(f float64) {
	e.Entry.SetText(fmt.Sprintf("%f", f))
}

func (e *FloatEntry) TypedRune(r rune) {
	if r == ',' {
		r = '.'
	}
	if (r >= '0' && r <= '9') || r == '.' || r == '-' {
		guardInput(&e.Entry, isPartialFloat, func() { e.Entry.TypedRune(r) })
	}
}

func (e *FloatEntry) TypedShortcut(shortcut fyne.Shortcut) {
	paste, ok := shortcut.(*fyne.ShortcutPaste)
	if !ok {
		e.Entry.TypedShortcut(shortcut)
		return
	}
	paste = cleanPaste(paste, func(s string) string { return strings.Replace(s, ",", ".", 1) })
	if paste.Clipboard == nil || !isPartialFloat(paste.Clipboard.Content()) {
		return
	}
	guardInput(&e.Entry, func(s string) bool {
		_, ok := parseFloat(s)
		return ok || s == ""
	}, func() { e.Entry.TypedShortcut(paste) })
}

func (e *FloatEntry) Keyboard() mobile.KeyboardType {
	return mobile.NumberKeyboard
}

func NewFloatEntryBinding(f *float64) FloatEntryBinding {
	s := binding.NewString()
	_ = s.Set(fmt.Sprintf("%f", *f))
	b := FloatEntryBinding{
		s: s,
		i: binding.BindFloat(f),
	}
	b.s.AddListener(b)
	return b
}

// DataChanged copies the text to the float. Text that is not (yet) a
// number, like "-", leaves the float as it is.
func (b FloatEntryBinding) DataChanged() {
	s, err := b.s.Get()
	if err != nil {
		return
	}
	if s == "" {
		_ = b.i.Set(0)
	} else if f, ok := parseFloat(s); ok {
		_ = b.i.Set(f)
	}
}

func (b FloatEntryBinding) Set(f float64) {
	_ = b.s.Set(fmt.Sprintf("%f", f))
}
