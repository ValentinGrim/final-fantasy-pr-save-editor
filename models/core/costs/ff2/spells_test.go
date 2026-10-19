package ff2

import (
	"strconv"
	"strings"
	"testing"
)

// The arithmetic in spells.go must agree with the names of the ability table.
func TestSpellLevelMatchesTable(t *testing.T) {
	for _, a := range Abilities {
		level, ok := SpellLevel(a.Value)
		if !ok {
			t.Fatalf("%s (%d) is not a spell", a.Name, a.Value)
		}
		i := strings.LastIndexByte(a.Name, ' ')
		want, _ := strconv.Atoi(a.Name[i+1:])
		if level != want {
			t.Fatalf("%s (%d): level %d", a.Name, a.Value, level)
		}
		first := SpellAtLevel(a.Value, 1)
		if base := Abilities[first-firstSpellID].Name; base != a.Name[:i]+" 1" {
			t.Fatalf("%s: level 1 is %s", a.Name, base)
		}
	}
	for _, id := range []int{0, 4, 645, 711} {
		if _, ok := SpellLevel(id); ok {
			t.Fatalf("%d is a spell", id)
		}
	}
}

func TestSpellAtLevel(t *testing.T) {
	const fire1, fire8, fire16, thunder1 = 5, 12, 20, 21
	for _, c := range []struct{ id, level, want int }{
		{fire1, 8, fire8},
		{fire8, 1, fire1},
		{fire8, 16, fire16},
		{fire16, 99, fire16},
		{fire8, 0, fire1},
		{thunder1, 1, thunder1},
	} {
		if got := SpellAtLevel(c.id, c.level); got != c.want {
			t.Errorf("SpellAtLevel(%d, %d) = %d; want %d", c.id, c.level, got, c.want)
		}
	}
}

func TestLevelForPoints(t *testing.T) {
	for _, c := range []struct{ points, level int }{
		{-5, 1}, {0, 1}, {19, 1}, {20, 2}, {251, 8}, {456, 11}, {859, 15}, {860, 16}, {9999, 16},
	} {
		if got := LevelForPoints(c.points); got != c.level {
			t.Errorf("LevelForPoints(%d) = %d; want %d", c.points, got, c.level)
		}
	}
	for l := 1; l <= SpellMaxLevel; l++ {
		if got := LevelForPoints(PointsForLevel(l)); got != l {
			t.Errorf("level %d round trips to %d", l, got)
		}
	}
}
