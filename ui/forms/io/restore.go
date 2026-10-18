package io

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/browser"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save/config"
	"pixel-remastered-save-editor/save/file"
)

const backupTimeFormat = "Mon 2 Jan 2006, 15:04:05"

type restorePage struct {
	widget.BaseWidget
	game   global.Game
	window fyne.Window
	onBack func()

	backups  []file.Backup
	selected int

	root       *fyne.Container
	list       *widget.List
	empty      *widget.Label
	summary    *widget.Label
	dir        *widget.Entry
	details    *fyne.Container
	restoreBtn *widget.Button

	// Replaced by tests.
	listBackups   func(global.Game) ([]file.Backup, error)
	restoreBackup func(global.Game, file.Backup, string) error
	saveDir       func(global.Game) string
	openFolder    func(string) error
}

// NewRestore returns the page listing the backups of game, newest first, from
// which one can be restored. onBack is called by the Back button.
func NewRestore(game global.Game, window fyne.Window, onBack func()) fyne.CanvasObject {
	w := newRestorePage(game, window, onBack)
	w.reload()
	return w
}

func newRestorePage(game global.Game, window fyne.Window, onBack func()) *restorePage {
	w := &restorePage{
		game:          game,
		window:        window,
		onBack:        onBack,
		selected:      -1,
		listBackups:   file.ListBackups,
		restoreBackup: file.RestoreBackup,
		saveDir: func(game global.Game) string {
			return initialSaveDir(game, config.Dir(game))
		},
		openFolder: browser.OpenFolder,
	}
	w.ExtendBaseWidget(w)

	w.list = widget.NewList(
		func() int { return len(w.backups) },
		func() fyne.CanvasObject {
			title := widget.NewLabel("")
			title.TextStyle.Bold = true
			title.Truncation = fyne.TextTruncateEllipsis
			sub := widget.NewLabel("")
			sub.Truncation = fyne.TextTruncateEllipsis
			return container.New(&tightVBox{}, title, sub)
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			if id >= len(w.backups) {
				return
			}
			b := w.backups[id]
			c := o.(*fyne.Container)
			c.Objects[0].(*widget.Label).SetText(fmt.Sprintf("%s  ·  %s", b.Time.Format(backupTimeFormat), describeBackupFiles(b.Files)))
			c.Objects[1].(*widget.Label).SetText(describeBackupSource(b))
		})
	w.list.OnSelected = w.selectBackup
	w.list.OnUnselected = func(widget.ListItemID) {
		w.selected = -1
		w.details.Hide()
		w.restoreBtn.Disable()
		w.relayout()
	}

	w.empty = widget.NewLabel("")
	w.empty.Alignment = fyne.TextAlignCenter
	w.empty.Wrapping = fyne.TextWrapWord

	w.summary = widget.NewLabel("")
	w.summary.TextStyle.Bold = true
	w.summary.Truncation = fyne.TextTruncateEllipsis
	w.dir = widget.NewEntry()
	w.details = container.NewVBox(
		w.summary,
		container.NewBorder(nil, nil,
			widget.NewLabel("Restore to:"),
			widget.NewButton("Change", func() {
				pickDirectory(w.window, "Select the save directory to restore to", w.dir.Text, func(dir string) {
					w.dir.SetText(dir)
				})
			}),
			w.dir),
	)
	w.details.Hide()

	w.restoreBtn = widget.NewButton("Restore", w.confirmRestore)
	w.restoreBtn.Importance = widget.HighImportance
	w.restoreBtn.Disable()

	title := widget.NewLabel(fmt.Sprintf("FF%d backups, newest first", w.game))
	title.TextStyle.Bold = true
	hint := widget.NewLabel("Select a backup to put it back in the save directory. Close the game before restoring.")
	hint.Truncation = fyne.TextTruncateEllipsis
	where := widget.NewLabel("Stored in " + file.BackupsDir(w.game))
	where.Truncation = fyne.TextTruncateEllipsis
	top := container.NewVBox(title, hint, where, widget.NewSeparator())

	buttons := container.NewGridWithColumns(3,
		w.restoreBtn,
		widget.NewButton("Open backups folder", w.openBackupsFolder),
		widget.NewButton("Back", func() { w.onBack() }))
	bottom := container.NewVBox(widget.NewSeparator(), w.details, buttons)

	w.root = container.NewBorder(top, bottom, nil, nil, container.NewStack(w.list, w.empty))
	return w
}

func (w *restorePage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.root)
}

// relayout makes room for the details after they are shown or hidden.
func (w *restorePage) relayout() {
	w.root.Refresh()
}

// reload reads the backups again and clears the selection.
func (w *restorePage) reload() {
	var err error
	w.backups, err = w.listBackups(w.game)
	w.selected = -1
	w.list.UnselectAll()
	w.details.Hide()
	w.restoreBtn.Disable()
	switch {
	case err != nil:
		w.backups = nil
		w.empty.SetText(fmt.Sprintf("Could not read the backups of FF%d: %v", w.game, err))
	case len(w.backups) == 0:
		w.empty.SetText(fmt.Sprintf("No backups yet for FF%d. A backup is made each time the editor overwrites a save.", w.game))
	}
	if len(w.backups) == 0 {
		w.list.Hide()
		w.empty.Show()
	} else {
		w.empty.Hide()
		w.list.Show()
	}
	w.list.Refresh()
	w.relayout()
}

func (w *restorePage) selectBackup(id widget.ListItemID) {
	if id < 0 || id >= len(w.backups) {
		return
	}
	w.selected = id
	b := w.backups[id]
	w.summary.SetText(fmt.Sprintf("Selected: %s  ·  %s", b.Time.Format(backupTimeFormat), describeBackupFiles(b.Files)))
	w.dir.SetText(w.targetFor(b))
	w.details.Show()
	w.restoreBtn.Enable()
	w.relayout()
}

// targetFor returns the directory the backup came from when it still exists,
// otherwise the save directory of the game.
func (w *restorePage) targetFor(b file.Backup) string {
	if b.SourceDir != "" {
		if fi, err := os.Stat(b.SourceDir); err == nil && fi.IsDir() {
			return b.SourceDir
		}
	}
	return w.saveDir(w.game)
}

func (w *restorePage) confirmRestore() {
	if w.selected < 0 || w.selected >= len(w.backups) {
		return
	}
	b := w.backups[w.selected]
	dir := strings.TrimSpace(w.dir.Text)
	if dir == "" {
		dialog.NewError(errors.New("choose the save directory to restore to"), w.window).Show()
		return
	}
	msg := fmt.Sprintf("Restore %s from %s?\n\nThis replaces %d file(s) in %s. The current files are backed up first, so you can undo this restore.",
		describeBackupFiles(b.Files), b.Time.Format(backupTimeFormat), len(b.Files), dir)
	d := dialog.NewConfirm("Restore backup", msg, func(ok bool) {
		if ok {
			w.doRestore(b, dir)
		}
	}, w.window)
	d.SetConfirmText("Restore")
	d.Resize(fyne.NewSize(min(500, w.window.Canvas().Size().Width-40), 0))
	d.Show()
}

func (w *restorePage) doRestore(b file.Backup, dir string) {
	if err := w.restoreBackup(w.game, b, dir); err != nil {
		dialog.NewError(err, w.window).Show()
		return
	}
	// The list now starts with the backup made before restoring.
	w.reload()
	dialog.NewInformation("Backup restored",
		"Restored. Make sure the game is closed (it may overwrite the files), and reopen the save in the editor to see it.",
		w.window).Show()
}

func (w *restorePage) openBackupsFolder() {
	dir := file.BackupsDir(w.game)
	err := os.MkdirAll(dir, 0o755)
	if err == nil {
		err = w.openFolder(dir)
	}
	if err != nil {
		dialog.NewError(fmt.Errorf("could not open %s: %w", dir, err), w.window).Show()
	}
}

// describeBackupFiles names the files of a backup after their slots:
// "slot 1 (+ Bestiary)", "Auto Save". Unknown names are shown as they are.
func describeBackupFiles(files []string) string {
	var names []string
	for _, s := range saves {
		if slices.Contains(files, s.UUID) || slices.Contains(files, fmt.Sprintf("slot%d.sav", s.Slot)) {
			names = append(names, s.Name)
		}
	}
	bestiary := false
	for _, f := range files {
		if f == file.BestiaryFile {
			bestiary = true
		} else if !isSaveFile(f) {
			names = append(names, f)
		}
	}
	s := strings.Join(names, ", ")
	if bestiary {
		if s == "" {
			return "Bestiary"
		}
		s += " (+ Bestiary)"
	}
	return s
}

func describeBackupSource(b file.Backup) string {
	var s string
	if b.SourceDir != "" {
		s = "From " + b.SourceDir
	} else {
		s = "Source directory not recorded"
	}
	if b.Reason == file.ReasonRestore {
		s = "Made before a restore  ·  " + s
	}
	return s
}

// tightVBox stacks its objects without padding, for two-line list rows.
type tightVBox struct{}

func (tightVBox) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var s fyne.Size
	for _, o := range objects {
		m := o.MinSize()
		s.Width = max(s.Width, m.Width)
		s.Height += m.Height
	}
	return s
}

func (tightVBox) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for _, o := range objects {
		h := o.MinSize().Height
		o.Move(fyne.NewPos(0, y))
		o.Resize(fyne.NewSize(size.Width, h))
		y += h
	}
}
