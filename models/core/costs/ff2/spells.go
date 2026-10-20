package ff2

// In FF II each spell level is its own ability: Fire 1 is ability 5, Fire 2
// is 6, up to Fire 16 at 20, then Thunder 1 at 21. The 40 spells cover
// abilities 5 to 644. The ability ID holds the level; skillLevel holds the
// total skill points, and the game does not derive one from the other when it
// loads a save.

const (
	firstSpellID  = 5
	lastSpellID   = 644
	SpellMaxLevel = 16
	// contentOffset maps a spell's ability ID to its content ID (game table
	// "content": type 4, type_value = ability ID).
	contentOffset = 207
)

// spellPoints is the total skill points each level needs, level 1 first
// (game table "exp_table", type 1).
var spellPoints = [SpellMaxLevel]int{0, 20, 45, 70, 100, 135, 175, 220, 270, 330, 400, 480, 570, 660, 760, 860}

// SpellLevel returns the level of spell ability id, or false when id is not
// a spell.
func SpellLevel(id int) (level int, ok bool) {
	if id < firstSpellID || id > lastSpellID {
		return 0, false
	}
	return (id-firstSpellID)%SpellMaxLevel + 1, true
}

// SpellAtLevel returns the ability ID of spell id at level, clamped to
// [1, SpellMaxLevel]. id must be a spell.
func SpellAtLevel(id, level int) int {
	level = min(max(level, 1), SpellMaxLevel)
	first := id - (id-firstSpellID)%SpellMaxLevel
	return first + level - 1
}

// LevelForPoints returns the level reached with points.
func LevelForPoints(points int) int {
	level := 1
	for l, p := range spellPoints {
		if points >= p {
			level = l + 1
		}
	}
	return level
}

// PointsForLevel returns the points level needs, clamped to
// [1, SpellMaxLevel].
func PointsForLevel(level int) int {
	return spellPoints[min(max(level, 1), SpellMaxLevel)-1]
}

// SpellContentID returns the content ID of spell ability id. id must be a
// spell.
func SpellContentID(id int) int {
	return id + contentOffset
}
