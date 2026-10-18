package selections

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/ui/forms/editors"
	"pixel-remastered-save-editor/ui/forms/editors/mapData"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

type (
	Editor struct {
		widget.BaseWidget
		game     global.Game
		save     *core.Save
		renderer fyne.WidgetRenderer
	}
)

func NewEditor(game global.Game, save *core.Save) *Editor {
	s := &Editor{
		game: game,
		save: save,
	}
	s.ExtendBaseWidget(s)
	inputs.Load(game)

	tabs := container.NewAppTabs(
		tab("Characters", NewCharacters(s.save)),
		tab("Inventory", NewInventory(s.save)),
		tab("Party", editors.NewCoreParty(s.save.Party, s.save.Parties)),
		tab("Map Data", mapData.NewCore(s.save.Map)),
		tab("Transportation", editors.NewCoreTransportation(s.game, s.save.Transportations)),
		tab("Misc", NewMisc(s.save.Misc)),
	)
	if s.save.Bestiary != nil {
		tabs.Append(tab("Bestiary", NewBestiary(s.save.Bestiary)))
	}
	if s.game.IsSix() {
		tabs.Append(tab("Espers", editors.NewEspers(s.save)))
	}
	s.renderer = widget.NewSimpleRenderer(tabs)
	return s
}

// tab lets the content scroll when it does not fit: the window grows to the
// MinSize of its content, it must not get wider than the screen.
func tab(text string, content fyne.CanvasObject) *container.TabItem {
	return container.NewTabItem(text, container.NewScroll(content))
}

func (s *Editor) CreateRenderer() fyne.WidgetRenderer {
	return s.renderer
}
