package mapData

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/models/finder"
	"pixel-remastered-save-editor/save"
	"pixel-remastered-save-editor/ui/forms/inputs"
)

type (
	MapData struct {
		widget.BaseWidget
		player       *fyne.Container
		gps          *fyne.Container
		misc         *fyne.Container
		otherParties *fyne.Container
	}
)

func NewCore(data *core.MapData) *MapData {
	gpsMapID := inputs.NewIdEntryWithDataWithHint(&data.Gps.MapID, finder.Maps)
	miscMapID := inputs.NewIdEntryWithDataWithHint(&data.Map.MapID, finder.Maps)
	e := &MapData{
		player: inputs.NewForm(
			widget.NewLabel("Position X"), inputs.Fixed(inputs.NewFloatEntryWithData(&data.Player.Position.X), inputs.FloatWidth),
			widget.NewLabel("Position Y"), inputs.Fixed(inputs.NewFloatEntryWithData(&data.Player.Position.Y), inputs.FloatWidth),
			widget.NewLabel("Position Z"), inputs.Fixed(inputs.NewFloatEntryWithData(&data.Player.Position.Z), inputs.FloatWidth),
			widget.NewLabel("Facing Direction"), inputs.Fixed(inputs.NewIntEntryWithData(&data.Player.Direction), inputs.EntryWidth),
		),
		gps: inputs.NewForm(
			widget.NewLabel("Map ID"), inputs.IDValue(gpsMapID.ID, gpsMapID.Label),
			widget.NewLabel("Area ID"), inputs.Fixed(inputs.NewIntEntryWithData(&data.Gps.AreaID), inputs.EntryWidth),
			widget.NewLabel("ID"), inputs.Fixed(inputs.NewIntEntryWithData(&data.Gps.GpsID), inputs.EntryWidth),
			widget.NewLabel("Width"), inputs.Fixed(inputs.NewIntEntryWithData(&data.Gps.Width), inputs.EntryWidth),
			widget.NewLabel("Height"), inputs.Fixed(inputs.NewIntEntryWithData(&data.Gps.Height), inputs.EntryWidth),
		),
		misc: inputs.NewForm(
			widget.NewLabel("Map ID"), inputs.IDValue(miscMapID.ID, miscMapID.Label),
			widget.NewLabel("Point In"), inputs.Fixed(inputs.NewIntEntryWithData(&data.Map.PointIn), inputs.EntryWidth),
			widget.NewLabel("Transportation ID"), inputs.Fixed(inputs.NewIntEntryWithData(&data.Map.TransportationID), inputs.EntryWidth),
			widget.NewLabel("Carrying Hover Ship"), widget.NewCheckWithData("", binding.BindBool(&data.Map.CarryingHoverShip)),
		),
		otherParties: container.NewVBox(),
	}
	e.ExtendBaseWidget(e)
	forms := []*fyne.Container{e.player, e.gps, e.misc}
	if data.OtherParties != nil {
		if parties := *data.OtherParties; len(parties) > 0 && parties[0].PlayableCharacterCorpsId > -1 {
			for i, party := range parties {
				if p, loaded := e.loadParty(party); loaded {
					e.otherParties.Add(widget.NewLabel("Party " + strconv.Itoa(i+1)))
					mapID := inputs.NewIdEntryWithDataWithHint(&p.MapID, finder.Maps)
					c := inputs.NewForm(
						widget.NewLabel("Map ID"), inputs.IDValue(mapID.ID, mapID.Label),
						widget.NewLabel("Point In"), inputs.Fixed(inputs.NewIntEntryWithData(&p.PointIn), inputs.EntryWidth))
					if p.PE != nil {
						c.Add(widget.NewLabel("Position X"))
						c.Add(inputs.Fixed(inputs.NewFloatEntryWithData(&p.PE.Position.X), inputs.FloatWidth))
						c.Add(widget.NewLabel("Position Y"))
						c.Add(inputs.Fixed(inputs.NewFloatEntryWithData(&p.PE.Position.Y), inputs.FloatWidth))
						c.Add(widget.NewLabel("Position Z"))
						c.Add(inputs.Fixed(inputs.NewFloatEntryWithData(&p.PE.Position.Z), inputs.FloatWidth))
					}
					e.otherParties.Add(container.NewPadded(c))
					forms = append(forms, c)
				}
			}
		}
	}
	inputs.AlignForms(forms...)
	return e
}

// CreateRenderer stacks the sections, the other parties (FF6) last, in one
// column so the editor fits next to the map search.
func (e *MapData) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(inputs.WithSearch(
		container.NewVScroll(container.NewVBox(
			widget.NewLabel("Player"),
			container.NewPadded(e.player),
			widget.NewLabel("GPS"),
			container.NewPadded(e.gps),
			widget.NewLabel("Misc"),
			container.NewPadded(e.misc),
			e.otherParties)),
		inputs.GetSearches().Maps.Panel()))
}

// loadParty decodes the player entity of party into party.PE, so the
// editor edits the party itself; MapData.ToSave encodes PE again.
func (e *MapData) loadParty(party *save.OtherPartyData) (p *save.OtherPartyData, loaded bool) {
	pe, err := party.PlayerEntity()
	if err == nil {
		party.PE = pe
		p = party
		loaded = true
	}
	return
}
