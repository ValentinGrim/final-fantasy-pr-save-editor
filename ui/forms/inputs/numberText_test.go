package inputs

import (
	"math"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestParseInt(t *testing.T) {
	for _, c := range []struct {
		in          string
		want        int
		clamped, ok bool
	}{
		{"0", 0, false, true},
		{"42", 42, false, true},
		{"-1", -1, false, true},
		{"-1000", -1000, false, true},
		{" 7 ", 7, false, true},
		{"2147483647", MaxInt32, false, true},
		{"2147483648", MaxInt32, true, true},
		{"99999999999", MaxInt32, true, true},
		{"-99999999999", MinInt32, true, true},
		{"999999999999999999999999", MaxInt32, true, true},
		{"", 0, false, false},
		{"-", 0, false, false},
		{"1.5", 0, false, false},
		{"1,5", 0, false, false},
		{"--1", 0, false, false},
		{"1-", 0, false, false},
		{"+1", 0, false, false},
		{"abc", 0, false, false},
	} {
		v, clamped, ok := parseInt(c.in, MinInt32, MaxInt32)
		if v != c.want || clamped != c.clamped || ok != c.ok {
			t.Errorf("parseInt(%q) = %d, %v, %v; want %d, %v, %v", c.in, v, clamped, ok, c.want, c.clamped, c.ok)
		}
	}
	if v, clamped, ok := parseInt("639277029519873290", math.MinInt, math.MaxInt); v != 639277029519873290 || clamped || !ok {
		t.Errorf("parseInt of a 64-bit tick count = %d, %v, %v", v, clamped, ok)
	}
}

func TestParseFloat(t *testing.T) {
	for _, c := range []struct {
		in   string
		want float64
		ok   bool
	}{
		{"1.5", 1.5, true},
		{"1,5", 1.5, true},
		{"-1000", -1000, true},
		{"-0,25", -0.25, true},
		{".5", 0.5, true},
		{"5.", 5, true},
		{"1e400", 0, false},
		{"NaN", 0, false},
		{"Inf", 0, false},
		{"1.2.3", 0, false},
		{"1,2.3", 0, false},
		{"-", 0, false},
		{".", 0, false},
		{"", 0, false},
		{"1-", 0, false},
	} {
		v, ok := parseFloat(c.in)
		if v != c.want || ok != c.ok {
			t.Errorf("parseFloat(%q) = %v, %v; want %v, %v", c.in, v, ok, c.want, c.ok)
		}
	}
	if v, ok := parseFloat("1" + strings.Repeat("0", 50)); !ok || v != math.MaxFloat32 {
		t.Errorf("huge float = %v, %v; want clamped to MaxFloat32", v, ok)
	}
}

func paste(e fyne.Shortcutable, s string) {
	cb := test.NewClipboard()
	cb.SetContent(s)
	e.TypedShortcut(&fyne.ShortcutPaste{Clipboard: cb})
}

func TestIntEntryTyping(t *testing.T) {
	test.NewTempApp(t)
	for _, c := range []struct {
		start int
		typed string
		text  string
		want  int
	}{
		{0, "-12", "-12", -12},
		{0, "1-2", "12", 12}, // a second '-' is dropped
		{0, "1.5", "15", 15}, // no decimal separator
		{0, "99999999999", "2147483647", MaxInt32},
		{0, "-99999999999", "-2147483648", MinInt32},
	} {
		v := c.start
		e := NewIntEntryWithData(&v)
		e.SetText("")
		test.Type(e, c.typed)
		if e.Text != c.text || v != c.want {
			t.Errorf("start %d, typed %q: text %q value %d; want %q %d", c.start, c.typed, e.Text, v, c.text, c.want)
		}
	}
}

func TestIntEntryMinusOnlyFirst(t *testing.T) {
	test.NewTempApp(t)
	v := 5
	e := NewIntEntryWithData(&v)
	e.CursorColumn = 1
	test.Type(e, "-")
	if e.Text != "5" || v != 5 {
		t.Errorf("'-' after 5: text %q value %d", e.Text, v)
	}
}

func TestIntEntryNegativeStart(t *testing.T) {
	test.NewTempApp(t)
	v := -1
	e := NewIntEntryWithData(&v)
	if e.Text != "-1" {
		t.Fatalf("text = %q, want -1", e.Text)
	}
	e.CursorColumn = 0
	test.Type(e, "-") // a second '-' in front is rejected
	if e.Text != "-1" || v != -1 {
		t.Errorf("text %q value %d", e.Text, v)
	}
}

func TestIntEntryPaste(t *testing.T) {
	test.NewTempApp(t)
	for _, c := range []struct {
		pasted string
		text   string
		want   int
	}{
		{"1.5", "", 0},
		{"abc", "", 0},
		{" -42\n", "-42", -42},
		{"99999999999", "2147483647", MaxInt32},
	} {
		v := 0
		e := NewIntEntryWithData(&v)
		e.SetText("")
		paste(e, c.pasted)
		if e.Text != c.text || v != c.want {
			t.Errorf("pasted %q: text %q value %d; want %q %d", c.pasted, e.Text, v, c.text, c.want)
		}
	}

	// A paste that would make the text invalid is undone.
	v := 12
	e := NewIntEntryWithData(&v)
	e.CursorColumn = 2
	paste(e, "-3")
	if e.Text != "12" || v != 12 {
		t.Errorf("paste of -3 after 12: text %q value %d", e.Text, v)
	}
}

func TestIntEntryRange(t *testing.T) {
	test.NewTempApp(t)
	v := 0
	e := NewIntEntryWithDataRange(&v, math.MinInt, math.MaxInt)
	e.SetText("")
	test.Type(e, "639277029519873290")
	if v != 639277029519873290 {
		t.Errorf("value %d", v)
	}
}

func TestFloatEntryTyping(t *testing.T) {
	test.NewTempApp(t)
	for _, c := range []struct {
		typed string
		text  string
		want  float64
	}{
		{"1,5", "1.5", 1.5},
		{"-1000", "-1000", -1000},
		{"1.2.3", "1.23", 1.23},
		{"1-2", "12", 12},
		{"-", "-", 7}, // not a number yet: the value is kept
	} {
		v := 7.0
		e := NewFloatEntryWithData(&v)
		e.SetText("")
		v = 7
		test.Type(e, c.typed)
		if e.Text != c.text || v != c.want {
			t.Errorf("typed %q: text %q value %v; want %q %v", c.typed, e.Text, v, c.text, c.want)
		}
	}
}

func TestFloatEntryPaste(t *testing.T) {
	test.NewTempApp(t)
	for _, c := range []struct {
		pasted string
		text   string
		want   float64
	}{
		{"1,5", "1.5", 1.5},
		{"-0.25", "-0.25", -0.25},
		{"Inf", "", 0},
		{"NaN", "", 0},
		{"1e5", "", 0},
	} {
		v := 0.0
		e := NewFloatEntryWithData(&v)
		e.SetText("")
		paste(e, c.pasted)
		if e.Text != c.text || v != c.want {
			t.Errorf("pasted %q: text %q value %v; want %q %v", c.pasted, e.Text, v, c.text, c.want)
		}
	}
}
