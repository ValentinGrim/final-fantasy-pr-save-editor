package core

import (
	"testing"

	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
)

func miscLoad(game global.Game, ts string) (*save.Data, *Misc) {
	d := save.New(game)
	d.TimeStamp = ts
	return d, NewMisc(d, &save.UserData{}, &save.DataStorage{})
}

func TestMiscTimeStampUntouched(t *testing.T) {
	for _, ts := range []string{
		"1/5/2026 4:09:06 PM",   // en-US, single digit month and day
		"10/16/2026 4:09:06 PM", // en-US, real FF1 save
		"18/10/2026 14:46:28",   // fr-FR, real FF2 save
		"05/10/2026 09:03:01",   // fr-FR, leading zeros
		"2026/10/5 9:03:01",     // ja-JP
		"16.10.2026 15:42:07",   // de-DE
		"2026-10-16T15:42:07",   // ISO 8601
	} {
		d, m := miscLoad(global.One, ts)
		if d.TimeStamp != ts {
			t.Errorf("NewMisc changed %q to %q", ts, d.TimeStamp)
		}
		if m.TimeStamp == nil || *m.TimeStamp != ts {
			t.Errorf("%q: not editable", ts)
			continue
		}
		m.ToSaveData(d)
		if d.TimeStamp != ts {
			t.Errorf("ToSaveData changed %q to %q", ts, d.TimeStamp)
		}
	}
}

func TestMiscTimeStampUnknownLayout(t *testing.T) {
	const ts = "2026年10月16日 15:42:07"
	d, m := miscLoad(global.One, ts)
	if m.TimeStamp != nil {
		t.Errorf("unknown layout %q is editable", ts)
	}
	m.ToSaveData(d)
	if d.TimeStamp != ts {
		t.Errorf("time stamp changed to %q", d.TimeStamp)
	}
}

func TestMiscTimeStampEdit(t *testing.T) {
	for _, c := range []struct{ saved, edit, want string }{
		{"10/16/2026 4:09:06 PM", "1/2/2027 11:00:00 AM", "1/2/2027 11:00:00 AM"},
		{"18/10/2026 14:46:28", "19/10/2026 08:00:00", "19/10/2026 08:00:00"},
		{"18/10/2026 14:46:28", " 19/10/2026 08:00:00 ", "19/10/2026 08:00:00"},
		// Not in the layout of the saved time stamp: ignored.
		{"18/10/2026 14:46:28", "10/19/2026 8:00:00 AM", "18/10/2026 14:46:28"},
		{"10/16/2026 4:09:06 PM", "garbage", "10/16/2026 4:09:06 PM"},
	} {
		d, m := miscLoad(global.Two, c.saved)
		*m.TimeStamp = c.edit
		m.ToSaveData(d)
		if d.TimeStamp != c.want {
			t.Errorf("%q edited to %q: saved %q, want %q", c.saved, c.edit, d.TimeStamp, c.want)
		}
	}
}
