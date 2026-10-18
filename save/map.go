package save

import (
	"bytes"
	"reflect"
	"strings"
	"sync"

	oj "github.com/virtuald/go-ordered-json"
)

type (
	MapData struct {
		MapID                          int                `json:"mapId"`
		PointIn                        int                `json:"pointIn"`
		TransportationID               int                `json:"transportationId"`
		CarryingHoverShip              bool               `json:"carryingHoverShip"`
		PlayerEntityInternal           string             `json:"playerEntity"`
		CompanionEntity                string             `json:"companionEntity"`
		GpsDataInternal                string             `json:"gpsData"`
		MoveCount                      int                `json:"moveCount"`
		SubtractSteps                  int                `json:"subtractSteps"`
		TelepoCacheData                string             `json:"telepoCacheData"`
		PlayableCharacterCorpsID       int                `json:"playableCharacterCorpsId"`
		CurrentSelectedPartyId         *int               `json:"currentSelectedPartyId,omitempty"`
		TimerData                      string             `json:"timerData"`
		EncountEnable                  *bool              `json:"encountEnable,omitempty"`
		ViewType                       *int               `json:"viewType,omitempty"`
		OtherPartyDataList             *[]*OtherPartyData `json:"otherPartyDataList,omitempty"`
		PartyPlayableCharacterCorpsId  *[]int             `json:"partyPlayableCharacterCorpsId,omitempty"`
		FieldDefenseNpcEntityIDList    *[]any             `json:"fieldDefenseNpcEntityIDList,omitempty"`
		BeastFieldEncountExchangeFlags *[]int             `json:"beastFieldEncountExchangeFlags,omitempty"`
		BeastFieldEncountSeekGroupId   *int               `json:"beastFieldEncountSeekGroupId,omitempty"`
		RtsData                        *any               `json:"rtsData,omitempty"`
		MinimapEnable                  *bool              `json:"minimapEnable,omitempty"`
		// unknown holds the members of the saved JSON object that have no
		// field above, so a save keeps what the editor does not know.
		unknown oj.OrderedObject
	}
	GpsData struct {
		TransportationID *int `json:"transportationId,omitempty"`
		MapID            int  `json:"mapId"`
		AreaID           int  `json:"areaId"`
		GpsID            int  `json:"gpsId"`
		Width            int  `json:"width"`
		Height           int  `json:"height"`
	}
	PlayerEntity struct {
		Position  Position `json:"position"`
		Direction int      `json:"direction"`
	}
	OtherPartyData struct {
		MapID                    int           `json:"mapId"`
		PointIn                  int           `json:"pointIn"`
		PlayerEntityInternal     string        `json:"playerEntity"`
		PE                       *PlayerEntity `json:"-"`
		PlayableCharacterCorpsId int           `json:"playableCharacterCorpsId"`
	}
)

func (d *MapData) GpsData() (v *GpsData, err error) {
	return UnmarshalOne[GpsData](d.GpsDataInternal)
}

func (d *MapData) SetGpsData(v *GpsData) (err error) {
	d.GpsDataInternal, err = MarshalOne(v)
	return
}

func (d *MapData) PlayerEntity() (v *PlayerEntity, err error) {
	return UnmarshalOne[PlayerEntity](d.PlayerEntityInternal)
}

func (d *MapData) SetPlayerEntity(v *PlayerEntity) (err error) {
	d.PlayerEntityInternal, err = MarshalOne(v)
	return
}

func (d *OtherPartyData) PlayerEntity() (v *PlayerEntity, err error) {
	return UnmarshalOne[PlayerEntity](d.PlayerEntityInternal)
}

func (d *OtherPartyData) SetPlayerEntity(v *PlayerEntity) (err error) {
	d.PlayerEntityInternal, err = MarshalOne(v)
	return
}

// mapDataJSON has the fields of MapData without its JSON methods.
type mapDataJSON MapData

var mapDataKeys = sync.OnceValue(func() map[string]bool {
	keys := make(map[string]bool)
	t := reflect.TypeOf(MapData{})
	for i := 0; i < t.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("json"); tag != "" && tag != "-" {
			keys[strings.Split(tag, ",")[0]] = true
		}
	}
	return keys
})

func (d *MapData) UnmarshalJSON(b []byte) (err error) {
	if err = oj.Unmarshal(b, (*mapDataJSON)(d)); err != nil {
		return
	}
	dec := oj.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	dec.UseOrderedObject()
	var obj oj.OrderedObject
	if err = dec.Decode(&obj); err != nil {
		return
	}
	d.unknown = nil
	for _, m := range obj {
		if !mapDataKeys()[m.Key] {
			d.unknown = append(d.unknown, m)
		}
	}
	return
}

func (d *MapData) MarshalJSON() ([]byte, error) {
	b, err := oj.Marshal((*mapDataJSON)(d))
	if err != nil || len(d.unknown) == 0 {
		return b, err
	}
	var buf bytes.Buffer
	buf.Write(bytes.TrimSuffix(bytes.TrimSpace(b), []byte("}")))
	for _, m := range d.unknown {
		var k, v []byte
		if k, err = oj.Marshal(m.Key); err != nil {
			return nil, err
		}
		if v, err = oj.Marshal(m.Value); err != nil {
			return nil, err
		}
		if buf.Len() > 1 {
			buf.WriteByte(',')
		}
		buf.Write(k)
		buf.WriteByte(':')
		buf.Write(v)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
