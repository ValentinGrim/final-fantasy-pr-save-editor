package save

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMapDataRoundTripKeepsEveryMember(t *testing.T) {
	for _, in := range []string{
		// Real FF2 map data.
		`{"mapId":228,"pointIn":-1,"transportationId":1,"carryingHoverShip":false,"playerEntity":"{\"position\":{\"x\":32.0,\"y\":29.0,\"z\":349.0},\"direction\":4}","companionEntity":"","gpsData":"{\"mapId\":1,\"areaId\":26,\"gpsId\":1,\"width\":256,\"height\":256}","moveCount":28,"subtractSteps":72,"telepoCacheData":"","playableCharacterCorpsId":0,"timerData":"","encountEnable":true,"minimapEnable":false}`,
		// Members MapData does not know.
		`{"mapId":1,"pointIn":3,"transportationId":1,"carryingHoverShip":false,"playerEntity":"","companionEntity":"","gpsData":"","moveCount":0,"subtractSteps":0,"telepoCacheData":"","playableCharacterCorpsId":0,"timerData":"","newFlag":true,"newList":[1,2.50,{"b":null,"a":"x"}],"big":12345678901234567890}`,
	} {
		md, err := UnmarshalOne[MapData](in)
		if err != nil {
			t.Fatal(err)
		}
		out, err := MarshalOne(md)
		if err != nil {
			t.Fatal(err)
		}
		var a, b map[string]any
		_ = json.Unmarshal([]byte(in), &a)
		if err = json.Unmarshal([]byte(out), &b); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
		if !reflect.DeepEqual(a, b) {
			t.Errorf("round trip changed the map data:\n in: %s\nout: %s", in, out)
		}
	}
}

func TestMapDataMinimapEnableOmittedWhenAbsent(t *testing.T) {
	md, err := UnmarshalOne[MapData](`{"mapId":1}`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := MarshalOne(md)
	if err != nil {
		t.Fatal(err)
	}
	var b map[string]any
	_ = json.Unmarshal([]byte(out), &b)
	if _, ok := b["minimapEnable"]; ok {
		t.Errorf("minimapEnable added: %s", out)
	}
}
