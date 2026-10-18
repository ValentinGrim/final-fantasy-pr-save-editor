package ui

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"github.com/go-text/typesetting/font"
	"pixel-remastered-save-editor/save/config"
)

// fontTheme replaces the UI text font. Glyphs it lacks still fall back to the
// bundled font, then to the system fonts.
type fontTheme struct {
	fyne.Theme
	font fyne.Resource
}

func (t fontTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace || s.Symbol {
		return t.Theme.Font(s)
	}
	return t.font
}

// userFont returns the first loadable .ttf/.otf/.ttc found next to the
// executable, in the config dir or in the working dir. FYNE_FONT wins if set.
func userFont() fyne.Resource {
	if os.Getenv("FYNE_FONT") != "" {
		return nil
	}
	for _, d := range fontDirs() {
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range entries {
			switch strings.ToLower(filepath.Ext(e.Name())) {
			case ".ttf", ".otf", ".ttc":
			default:
				continue
			}
			if e.IsDir() {
				continue
			}
			p := filepath.Join(d, e.Name())
			r, err := readFont(p)
			if err == nil {
				return r
			}
			fyne.LogError("Skipping font "+p, err)
		}
	}
	return nil
}

func fontDirs() (dirs []string) {
	var candidates []string
	if exe, err := os.Executable(); err == nil {
		if r, err := filepath.EvalSymlinks(exe); err == nil {
			exe = r
		}
		candidates = append(candidates, filepath.Dir(exe))
	}
	candidates = append(candidates, filepath.Dir(config.Path()))
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd)
	}
	seen := make(map[string]bool)
	for _, d := range candidates {
		if a, err := filepath.Abs(d); err == nil && !seen[a] {
			seen[a] = true
			dirs = append(dirs, a)
		}
	}
	return
}

func readFont(p string) (fyne.Resource, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	// Fyne only loads single fonts, use the first one of a collection
	if bytes.HasPrefix(b, []byte("ttcf")) {
		if b, err = firstFont(b); err != nil {
			return nil, err
		}
	}
	// Fyne panics on fonts it cannot parse, so check with the same parser
	if _, err = font.ParseTTF(bytes.NewReader(b)); err != nil {
		return nil, err
	}
	return fyne.NewStaticResource(filepath.Base(p), b), nil
}

var errFontCollection = errors.New("invalid font collection")

// firstFont rebuilds the first font of a TrueType/OpenType collection as a
// standalone font file.
func firstFont(b []byte) ([]byte, error) {
	be := binary.BigEndian
	size := uint64(len(b))
	if size < 16 {
		return nil, errFontCollection
	}
	off := uint64(be.Uint32(b[12:]))
	if off+12 > size {
		return nil, errFontCollection
	}
	n := uint64(be.Uint16(b[off+4:]))
	head := 12 + 16*n
	if off+head > size {
		return nil, errFontCollection
	}
	out := make([]byte, head, size)
	copy(out, b[off:off+head])
	for i := uint64(0); i < n; i++ {
		rec := out[12+16*i:]
		start, length := uint64(be.Uint32(rec[8:])), uint64(be.Uint32(rec[12:]))
		if start+length > size {
			return nil, errFontCollection
		}
		be.PutUint32(rec[8:], uint32(len(out)))
		out = append(out, b[start:start+length]...)
		for len(out)%4 != 0 {
			out = append(out, 0)
		}
	}
	return out, nil
}
