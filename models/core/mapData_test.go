package core

import (
	"encoding/json"
	"testing"

	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
)

// mapDataJSON is the map data of a real FF2 save, plus a member the editor
// does not know ("futureKey").
const mapDataJSON = `{"mapId":228,"pointIn":-1,"transportationId":1,"carryingHoverShip":false,"playerEntity":"{\"position\":{\"x\":32.0,\"y\":29.0,\"z\":349.0},\"direction\":4}","companionEntity":"","gpsData":"{\"mapId\":1,\"areaId\":26,\"gpsId\":1,\"width\":256,\"height\":256}","moveCount":28,"subtractSteps":72,"telepoCacheData":"","playableCharacterCorpsId":0,"timerData":"","encountEnable":true,"minimapEnable":false,"futureKey":{"a":[1,2.50]}}`

const ff6MapDataJSON = `{"mapId":3,"pointIn":0,"transportationId":0,"carryingHoverShip":false,"playerEntity":"{\"position\":{\"x\":1.0,\"y\":2.0,\"z\":3.0},\"direction\":2}","companionEntity":"","gpsData":"{\"mapId\":1,\"areaId\":1,\"gpsId\":1,\"width\":256,\"height\":256}","moveCount":0,"subtractSteps":0,"telepoCacheData":"","playableCharacterCorpsId":0,"timerData":"","otherPartyDataList":[{"mapId":10,"pointIn":1,"playerEntity":"{\"position\":{\"x\":5.0,\"y\":6.0,\"z\":7.0},\"direction\":1}","playableCharacterCorpsId":1},{"mapId":11,"pointIn":2,"playerEntity":"{\"position\":{\"x\":8.0,\"y\":9.0,\"z\":10.0},\"direction\":1}","playableCharacterCorpsId":2}]}`

func mapDataLoad(t *testing.T, game global.Game, js string) (*MapData, *save.MapData) {
	t.Helper()
	md, err := save.UnmarshalOne[save.MapData](js)
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewMapData(game, md)
	if err != nil {
		t.Fatal(err)
	}
	// Save.ToSave unpacks the map data again before calling MapData.ToSave.
	fresh, err := save.UnmarshalOne[save.MapData](js)
	if err != nil {
		t.Fatal(err)
	}
	return m, fresh
}

func mapDataSaved(t *testing.T, md *save.MapData) map[string]any {
	t.Helper()
	s, err := save.MarshalOne(md)
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err = json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("%v: %s", err, s)
	}
	return v
}

func TestMapDataToSaveKeepsMiscEdits(t *testing.T) {
	m, md := mapDataLoad(t, global.Two, mapDataJSON)
	m.Map.MapID = 999
	m.Map.PointIn = 4
	m.Map.TransportationID = 3
	m.Map.CarryingHoverShip = true
	m.Player.Position.X = 40
	m.Gps.AreaID = 27
	if err := m.ToSave(global.Two, md); err != nil {
		t.Fatal(err)
	}
	v := mapDataSaved(t, md)
	for k, want := range map[string]any{"mapId": 999.0, "pointIn": 4.0, "transportationId": 3.0, "carryingHoverShip": true, "minimapEnable": false, "encountEnable": true, "moveCount": 28.0} {
		if v[k] != want {
			t.Errorf("%s = %v, want %v", k, v[k], want)
		}
	}
	if fk, ok := v["futureKey"].(map[string]any); !ok || len(fk["a"].([]any)) != 2 {
		t.Errorf("futureKey = %v, want it kept", v["futureKey"])
	}
	pe, _ := save.UnmarshalOne[save.PlayerEntity](v["playerEntity"].(string))
	if pe.Position.X != 40 {
		t.Errorf("player x = %v, want 40", pe.Position.X)
	}
	gps, _ := save.UnmarshalOne[save.GpsData](v["gpsData"].(string))
	if gps.AreaID != 27 {
		t.Errorf("gps area = %v, want 27", gps.AreaID)
	}
}

func TestMapDataToSaveFF6OtherParties(t *testing.T) {
	m, md := mapDataLoad(t, global.Six, ff6MapDataJSON)
	// What the map editor does: decode each party's player entity into PE
	// and edit the party in place.
	for _, p := range *m.OtherParties {
		var err error
		if p.PE, err = p.PlayerEntity(); err != nil {
			t.Fatal(err)
		}
	}
	p := (*m.OtherParties)[1]
	p.MapID = 50
	p.PointIn = 7
	p.PE.Position.Y = 99
	if err := m.ToSave(global.Six, md); err != nil {
		t.Fatal(err)
	}
	saved := *md.OtherPartyDataList
	if len(saved) != 2 || saved[1].MapID != 50 || saved[1].PointIn != 7 || saved[0].MapID != 10 {
		t.Fatalf("other parties not saved: %+v %+v", saved[0], saved[1])
	}
	pe, err := saved[1].PlayerEntity()
	if err != nil {
		t.Fatal(err)
	}
	if pe.Position.Y != 99 || pe.Position.X != 8 {
		t.Errorf("party 2 position = %+v, want y 99", pe.Position)
	}
	v := mapDataSaved(t, md)
	if l, ok := v["otherPartyDataList"].([]any); !ok || len(l) != 2 {
		t.Errorf("otherPartyDataList = %v", v["otherPartyDataList"])
	}
}
