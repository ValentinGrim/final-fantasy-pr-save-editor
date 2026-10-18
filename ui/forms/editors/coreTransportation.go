package editors

import (
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

type (
	Transportation struct {
		widget.BaseWidget
		tabs []*container.TabItem
	}
)

func NewCoreTransportation(game global.Game, transportations *core.Transportations) *Transportation {
	e := &Transportation{}
	e.ExtendBaseWidget(e)

	for _, t := range transportations.Transportations {
		if i, ok := e.createTabItem(game, t); ok {
			e.tabs = append(e.tabs, i)
		}
	}
	return e
}

func (e *Transportation) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewAppTabs(e.tabs...))
}

func (e *Transportation) createTabItem(game global.Game, transport *save.OwnedTransportation) (*container.TabItem, bool) {
	p := transport.Position
	if p == nil {
		return nil, false
	}
	name := strconv.Itoa(transport.ID)
	if game.IsSix() {
		if transport.ID == 4 {
			name = "Blackjack"
		}
		if transport.ID == 5 {
			name = "Falcon"
		}
	}
	in := inputs.NewIdEntryWithDataWithHint(&transport.MapID, finder.Maps)
	// The time stamp ticks are a C# long, not an int.
	c := inputs.NewForm(
		widget.NewLabel("Map ID"), inputs.IDValue(in.ID, in.Label),
		widget.NewLabel("Position X"), inputs.Fixed(inputs.NewFloatEntryWithData(&p.X), inputs.FloatWidth),
		widget.NewLabel("Position Y"), inputs.Fixed(inputs.NewFloatEntryWithData(&p.Y), inputs.FloatWidth),
		widget.NewLabel("Position Z"), inputs.Fixed(inputs.NewFloatEntryWithData(&p.Z), inputs.FloatWidth),
		widget.NewLabel("Facing"), inputs.Fixed(inputs.NewIntEntryWithData(&transport.Direction), inputs.EntryWidth),
		widget.NewLabel("Time Stamp Ticks"), inputs.Fixed(inputs.NewIntEntryWithDataRange(&transport.TimeStampTicks, math.MinInt, math.MaxInt), inputs.WideWidth))
	return container.NewTabItem(name, container.NewBorder(c, nil, nil, nil)), true
}
