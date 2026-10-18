package core

import (
	"encoding/json"
	"reflect"
	"testing"

	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
)

const bestiaryJSON = `{"scenarioFlags":"{\"target\":[1,0,1]}","monsterDefeats":"{\"keys\":[3,1,96],\"values\":[198,122,1],\"extra\":7}","totalSubjugationCount":321,"unknown":[1,2.50]}`

func loadBestiary(t *testing.T, game global.Game) (*save.Data, *Bestiary) {
	t.Helper()
	d := save.New(game)
	d.SetLoadedBestiary([]byte(bestiaryJSON), nil)
	sb, err := d.Bestiary()
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewBestiary(game, sb)
	if err != nil {
		t.Fatal(err)
	}
	return d, b
}

func pack(t *testing.T, d *save.Data, b *Bestiary) []byte {
	t.Helper()
	sb, err := b.ToSave()
	if err != nil {
		t.Fatal(err)
	}
	if err = d.SetBestiary(sb); err != nil {
		t.Fatal(err)
	}
	out, _, _ := d.BestiaryToSave()
	return out
}

func TestBestiaryRoundTripUnchanged(t *testing.T) {
	d, b := loadBestiary(t, global.Two)
	if got := pack(t, d, b); string(got) != bestiaryJSON {
		t.Fatalf("unchanged bestiary re-encoded differently:\n%s\n%s", got, bestiaryJSON)
	}
	if _, _, changed := d.BestiaryToSave(); changed {
		t.Fatal("unchanged bestiary reported as changed")
	}
}

func TestBestiaryEdit(t *testing.T) {
	d, b := loadBestiary(t, global.Two)
	b.MonsterDefeats[0].Count = 200                  // 3: 198 -> 200
	b.MonsterDefeats[2].Count = 0                    // 96 dropped
	b.Add().ID = 50                                  // no count: dropped
	*b.Add() = MonsterDefeatedCount{ID: 1, Count: 3} // duplicate: merged into 1
	b.MonsterDefeats = append(b.MonsterDefeats, nil, &MonsterDefeatedCount{ID: 10, Count: 2})

	out := pack(t, d, b)
	want := `{"scenarioFlags":"{\"target\":[1,0,1]}","monsterDefeats":"{\"keys\":[3,1,10],\"values\":[200,125,2],\"extra\":7}","totalSubjugationCount":327,"unknown":[1,2.50]}`
	if string(out) != want {
		t.Fatalf("got\n%s\nwant\n%s", out, want)
	}
	if _, _, changed := d.BestiaryToSave(); !changed {
		t.Fatal("edit not reported as changed")
	}
	if len(b.MonsterDefeats) != 7 {
		t.Fatal("ToSave must not modify the model")
	}
}

func TestBestiaryDefeatAll(t *testing.T) {
	d, b := loadBestiary(t, global.Two)
	b.MonsterDefeats = append(b.MonsterDefeats, &MonsterDefeatedCount{ID: 5}) // listed, never defeated
	ids := []int{130, 4, 1, 2, 3, 5, 96, 128, 129, 2, 0}
	b.DefeatAll(ids)

	sb, err := b.ToSave()
	if err != nil {
		t.Fatal(err)
	}
	md, err := sb.MonsterDefeats()
	if err != nil {
		t.Fatal(err)
	}
	// Existing counts kept, new monsters appended in ID order with a count
	// of 1, IDs above 128 (FF2 boss variants) skipped.
	wantKeys := []int{3, 1, 96, 5, 2, 4, 128}
	wantValues := []int{198, 122, 1, 1, 1, 1, 1}
	if !reflect.DeepEqual(md.Keys, wantKeys) || !reflect.DeepEqual(md.Values, wantValues) {
		t.Fatalf("got %v %v", md.Keys, md.Values)
	}
	if sb.TotalSubjugationCount != 198+122+1+4 {
		t.Fatalf("total %d", sb.TotalSubjugationCount)
	}
	if err = d.SetBestiary(sb); err != nil {
		t.Fatal(err)
	}
	if !json.Valid(d.BestiaryDataInternal) {
		t.Fatal("invalid JSON")
	}

	// Games without a known limit take every listed ID.
	b = &Bestiary{Game: global.Six}
	b.DefeatAll([]int{300, 1})
	if len(b.MonsterDefeats) != 2 || b.MonsterDefeats[1].ID != 300 {
		t.Fatalf("got %+v", b.MonsterDefeats)
	}
}

func TestBestiaryMismatchedKeysValues(t *testing.T) {
	sb := &save.Bestiary{MonsterDefeatsInternal: `{"keys":[1,2],"values":[1]}`}
	if _, err := NewBestiary(global.One, sb); err == nil {
		t.Fatal("expected an error")
	}
}
