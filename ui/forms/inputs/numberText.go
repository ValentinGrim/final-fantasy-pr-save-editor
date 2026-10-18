package inputs

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// The game stores most integers in a C# int.
const (
	MinInt32 = math.MinInt32
	MaxInt32 = math.MaxInt32
)

// isPartialInt reports whether s can be the text of an IntEntry while the
// user types it: digits with an optional leading '-' ("" and "-" included).
func isPartialInt(s string) bool {
	for i, r := range s {
		if (r < '0' || r > '9') && (r != '-' || i != 0) {
			return false
		}
	}
	return true
}

// parseInt parses the text of an IntEntry. A number outside [lo, hi] is
// clamped to it and clamped is true. ok is false when s is not a number.
func parseInt(s string, lo, hi int) (v int, clamped bool, ok bool) {
	s = strings.TrimSpace(s)
	if !isPartialInt(s) || s == "" || s == "-" {
		return 0, false, false
	}
	v, err := strconv.Atoi(s)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return 0, false, false
	}
	// On ErrRange, v is the largest or smallest int.
	if v < lo {
		return lo, true, true
	}
	if v > hi {
		return hi, true, true
	}
	return v, err != nil, true
}

// isPartialFloat reports whether s can be the text of a FloatEntry while the
// user types it: digits with an optional leading '-' and at most one decimal
// separator ('.' or ',').
func isPartialFloat(s string) bool {
	sep := false
	for i, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r == '-' && i == 0:
		case (r == '.' || r == ',') && !sep:
			sep = true
		default:
			return false
		}
	}
	return true
}

// parseFloat parses the text of a FloatEntry; ',' is read as the decimal
// separator like '.'. The game stores positions in a C# float, so the value
// is clamped to its range. ok is false when s is not a number.
func parseFloat(s string) (v float64, ok bool) {
	s = strings.TrimSpace(s)
	if !isPartialFloat(s) || strings.IndexFunc(s, func(r rune) bool { return r >= '0' && r <= '9' }) < 0 {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return 0, false
	}
	return math.Max(-math.MaxFloat32, math.Min(v, math.MaxFloat32)), true
}

// pasteClipboard is the clipboard of a paste with its content cleaned up
// for a number entry.
type pasteClipboard struct {
	fyne.Clipboard
	content string
}

func (c pasteClipboard) Content() string {
	return c.content
}

// cleanPaste returns the paste shortcut with the clipboard content passed
// through clean (surrounding spaces are always removed).
func cleanPaste(p *fyne.ShortcutPaste, clean func(string) string) *fyne.ShortcutPaste {
	if p.Clipboard == nil {
		return p
	}
	c := strings.TrimSpace(p.Clipboard.Content())
	if clean != nil {
		c = clean(c)
	}
	return &fyne.ShortcutPaste{Clipboard: pasteClipboard{Clipboard: p.Clipboard, content: c}, Secondary: p.Secondary}
}

// guardInput runs apply (typing or pasting into e) and puts back the text
// and cursor e had before when the result is not valid.
func guardInput(e *widget.Entry, valid func(string) bool, apply func()) {
	text, row, col := e.Text, e.CursorRow, e.CursorColumn
	apply()
	if !valid(e.Text) {
		e.SetText(text)
		e.CursorRow, e.CursorColumn = row, col
		e.Refresh()
	}
}
