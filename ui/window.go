package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/browser"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/save/config"
	"pixel-remastered-save-editor/save/file"
	"pixel-remastered-save-editor/ui/bundled"
	"pixel-remastered-save-editor/ui/forms/io"
	"pixel-remastered-save-editor/ui/forms/selections"
)

const (
	appName = "Final Fantasy Pixel Remastered Save Editor"
	// appID must match the .desktop file name so Wayland compositors find the icon
	appID = "io.github.valentingrim.ffpr-save-editor"
)

type (
	Gui interface {
		App() fyne.App
		Window() fyne.Window
		Canvas() *fyne.Container
		Load()
		Save()
		Restore()
		Run()
	}
	gui struct {
		app    fyne.App
		window fyne.Window
		canvas *fyne.Container
		open   *fyne.MenuItem
		save   *fyne.MenuItem
		// restore is available except while a save page is open
		restore *fyne.MenuItem
		prev    fyne.CanvasObject
		data    *core.Save
	}
	MenuItem interface {
		Item() *fyne.MenuItem
		Disable()
		Enable()
		OnSelected(func())
	}
	menuItem struct {
		*fyne.MenuItem
	}
)

var (
	_ MenuItem = menuItem{}
)

func New() Gui {
	a := app.NewWithID(appID)
	if f := userFont(); f != nil {
		a.Settings().SetTheme(fontTheme{Theme: theme.DefaultTheme(), font: f})
	}
	g := &gui{
		app:    a,
		window: a.NewWindow(fmt.Sprintf("%s - v%s", appName, browser.Version)),
		canvas: container.NewStack(),
	}
	g.window.SetIcon(fyne.NewStaticResource("icon", bundled.ResourceIconPng.StaticContent))
	g.window.SetContent(g.canvas)
	g.open = fyne.NewMenuItem("Open", func() {
		g.Load()
	})
	g.open.ChildMenu = fyne.NewMenu("Open",
		fyne.NewMenuItem("I", func() {
			g.gameSelected(global.One)
		}),
		fyne.NewMenuItem("II", func() {
			g.gameSelected(global.Two)
		}),
		fyne.NewMenuItem("III", func() {
			g.gameSelected(global.Three)
		}),
		fyne.NewMenuItem("IV", func() {
			g.gameSelected(global.Four)
		}),
		fyne.NewMenuItem("V", func() {
			g.gameSelected(global.Five)
		}),
		fyne.NewMenuItem("VI", func() {
			g.gameSelected(global.Six)
		}))
	g.save = fyne.NewMenuItem("Save", func() {
		g.Save()
	})
	g.save.Disabled = true
	g.restore = fyne.NewMenuItem("Restore backup…", func() {
		g.Restore()
	})
	g.window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			g.open,
			fyne.NewMenuItemSeparator(),
			g.save,
			fyne.NewMenuItemSeparator(),
			g.restore,
		),
		createSupportMenu(g.window)))
	g.window.Resize(fyne.NewSize(float32(global.WindowWidth), float32(global.WindowHeight)))
	return g
}

func (g *gui) App() fyne.App {
	return g.app
}

func (g *gui) Window() fyne.Window {
	return g.window
}

func (g *gui) Canvas() *fyne.Container {
	return g.canvas
}

func (g *gui) Load() {
	if len(g.canvas.Objects) > 0 {
		g.prev = g.canvas.Objects[0]
	}
	// Restore stays available on the Load pages: a save that no longer opens is
	// when a backup is needed most. Its Back returns to the Load page.
	g.open.Disabled = true
	g.canvas.RemoveAll()
	g.canvas.Add(io.NewGameSelect(g.gameSelected))
}

func (g *gui) gameSelected(game global.Game) {
	// keep the open editor set by Load() as the screen to go back to
	if g.prev == nil {
		g.prev = g.canvas.Objects[0]
	}
	g.canvas.RemoveAll()
	g.canvas.Add(io.NewFileIO(io.Load, game, g.window, config.Dir(game), func(game global.Game, name, dir, f string, slot int, saveType global.SaveFileType) {
		defer func() {
			g.open.Disabled = false
			g.restore.Disabled = false
		}()
		// Load file
		data, err := file.LoadSave(game, dir, f, saveType)
		var s *core.Save
		if err == nil {
			// NewSave returns a partial save along with its error: never let it replace the loaded one.
			s, err = core.NewSave(data)
		}
		if err != nil {
			if g.prev != nil {
				g.canvas.RemoveAll()
				g.canvas.Add(g.prev)
			}
			dialog.NewError(err, g.window).Show()
		} else {
			// Success
			config.SetSaveDir(game, dir)
			g.data = s
			g.canvas.RemoveAll()
			g.prev = nil
			g.save.Disabled = false
			finder.Load(game, g.data.Characters.All())
			g.canvas.Add(selections.NewEditor(game, g.data))
			g.window.SetTitle(fmt.Sprintf("%s - FF%d - slot %d", appName, game, slot))
		}
	}, func() {
		defer func() {
			g.open.Disabled = false
			g.restore.Disabled = false
		}()
		// Cancel
		if g.prev != nil {
			g.canvas.RemoveAll()
			g.canvas.Add(g.prev)
		}
	}))
}

func (g *gui) Save() {
	if len(g.canvas.Objects) > 0 {
		g.prev = g.canvas.Objects[0]
	}
	g.open.Disabled = true
	g.save.Disabled = true
	g.restore.Disabled = true
	g.canvas.RemoveAll()
	g.canvas.Add(
		io.NewFileIO(io.Save, g.data.Data.Game, g.window, config.Dir(g.data.Data.Game), func(game global.Game, name, dir, f string, slot int, saveType global.SaveFileType) {
			defer func() {
				g.open.Disabled = false
				g.save.Disabled = false
				g.restore.Disabled = false
			}()
			// Save file
			config.SetSaveDir(game, dir)
			d, err := g.data.ToSave(game, slot)
			if err == nil {
				err = file.SaveSave(game, d, slot, dir, f, saveType)
			}
			if err != nil {
				if g.prev != nil {
					g.canvas.RemoveAll()
					g.canvas.Add(g.prev)
				}
				dialog.NewError(err, g.window).Show()
			} else {
				// Success
				if g.prev != nil {
					g.canvas.RemoveAll()
					g.canvas.Add(g.prev)
				}
				g.window.SetTitle(fmt.Sprintf("%s - FF%d - slot %d", appName, game, slot))
			}
		}, func() {
			defer func() {
				g.open.Disabled = false
				g.save.Disabled = false
				g.restore.Disabled = false
			}()
			// Cancel
			if g.prev != nil {
				g.canvas.RemoveAll()
				g.canvas.Add(g.prev)
			}
		}))
}

// Restore asks for a game, then shows the page restoring one of its backups.
// Back returns to the screen shown before, with the menus as they were.
func (g *gui) Restore() {
	var prev fyne.CanvasObject
	if len(g.canvas.Objects) > 0 {
		prev = g.canvas.Objects[0]
	}
	openDisabled, saveDisabled := g.open.Disabled, g.save.Disabled
	g.open.Disabled, g.save.Disabled, g.restore.Disabled = true, true, true
	back := func() {
		g.open.Disabled, g.save.Disabled, g.restore.Disabled = openDisabled, saveDisabled, false
		g.canvas.RemoveAll()
		if prev != nil {
			g.canvas.Add(prev)
		} else {
			g.Load()
		}
	}
	g.canvas.RemoveAll()
	g.canvas.Add(container.NewBorder(nil, widget.NewButton("Cancel", back), nil, nil,
		io.NewGameSelect(func(game global.Game) {
			g.canvas.RemoveAll()
			g.canvas.Add(io.NewRestore(game, g.window, back))
		})))
}

func (g *gui) Run() {
	g.window.ShowAndRun()
}

func (m menuItem) Item() *fyne.MenuItem {
	return m.MenuItem
}

func (m menuItem) Disable() {
	m.MenuItem.Disabled = true
}

func (m menuItem) Enable() {
	m.MenuItem.Disabled = false
}

func (m menuItem) OnSelected(f func()) {
	m.MenuItem.Action = f
}

func createSupportMenu(w fyne.Window) *fyne.Menu {
	txt := fmt.Sprintf(`
## Final Fantasy Pixel Remastered Save Editor %s
---
Written by [Kiame Vivacity](https://github.com/kiamev)
### Support Me:
- [Patreon: https://www.patreon.com/kiamev](https://www.patreon.com/kiamev)
- [Ko-fi: https://ko-fi.com/kiamev](https://ko-fi.com/kiamev)

Linux port and maintenance by [ValentinGrim](https://github.com/ValentinGrim/final-fantasy-pr-save-editor)
---
### Contributors:
- Yosituna for compiling FF I - V game data
- Fallacies for compiling FF VI game data
- Silvris for figuring out how to decrypt saves
- Scyllinice for helping build out the bestiaries
`,
		browser.Version)

	m := fyne.NewMenu("Support Project")
	m.Items = append(m.Items, fyne.NewMenuItem("About", func() {
		dialog.ShowCustom("About", "ok", container.NewBorder(
			widget.NewRichTextFromMarkdown(txt), nil, nil, nil,
		), w)
	}))
	return m
}
