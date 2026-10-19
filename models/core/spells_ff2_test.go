package core

import (
	"slices"
	"testing"

	"pixel-remastered-save-editor/save"
)

// ff2Character mirrors a measured FF II character: ability 711 first, then
// spells; abilityDictionary key 3 holds the spells with older skill levels.
func ff2Character(t *testing.T) *Character {
	t.Helper()
	const (
		fire8, cure11, item = 12, 335, 711
	)
	base := &save.OwnedCharacter{
		IsEnableCorps:                    true,
		SortOrderOwnedAbilityIdsInternal: `{"target":[219,542]}`,
	}
	d := &save.AbilityDictionary{
		Keys: []int{3},
		Values: [][]*save.Ability{{
			{AbilityID: cure11, ContentID: 542, SkillLevel: 402},
			{AbilityID: fire8, ContentID: 219, SkillLevel: 227},
		}},
	}
	if err := base.SetAbilityDictionary(d); err != nil {
		t.Fatal(err)
	}
	return &Character{
		Base: base,
		Abilities: []*save.Ability{
			{AbilityID: item, ContentID: 937},
			{AbilityID: cure11, ContentID: 542, SkillLevel: 456},
			{AbilityID: fire8, ContentID: 219, SkillLevel: 251},
		},
		AdditionAbilityIds: []int{937, 542, 219},
	}
}

func dictionary(t *testing.T, c *Character) (keys []int, spells []save.Ability) {
	t.Helper()
	d, err := c.Base.AbilityDictionary()
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range d.Values {
		for _, a := range v {
			spells = append(spells, *a)
		}
	}
	return d.Keys, spells
}

func TestPreSaveFF2UnchangedLeavesListsAlone(t *testing.T) {
	c := ff2Character(t)
	dict, sort := c.Base.AbilityDictionaryInternal, c.Base.SortOrderOwnedAbilityIdsInternal
	if err := (&Save{}).preCharacterSaveFF2(c); err != nil {
		t.Fatal(err)
	}
	if c.Base.AbilityDictionaryInternal != dict || c.Base.SortOrderOwnedAbilityIdsInternal != sort {
		t.Fatal("an unchanged character had its dictionary or sort order rewritten")
	}
	if !slices.Equal(c.AdditionAbilityIds, []int{937, 542, 219}) {
		t.Fatalf("addition order %v", c.AdditionAbilityIds)
	}
}

func TestPreSaveFF2LevelChange(t *testing.T) {
	c := ff2Character(t)
	// Fire 8 -> Fire 9 in the editor; a new Thunder 1; Cure 11 removed.
	c.Abilities[2].AbilityID, c.Abilities[2].SkillLevel = 13, 270
	c.Abilities = append(c.Abilities, &save.Ability{AbilityID: 21})
	c.RemoveAbilityIndex(1)
	if err := (&Save{}).preCharacterSaveFF2(c); err != nil {
		t.Fatal(err)
	}

	if got := c.Abilities[1].ContentID; got != 220 {
		t.Fatalf("Fire 9 content ID %d; want 220", got)
	}
	if got := c.Abilities[2].ContentID; got != 228 {
		t.Fatalf("Thunder 1 content ID %d; want 228", got)
	}
	if want := []int{937, 220, 228}; !slices.Equal(c.AdditionAbilityIds, want) {
		t.Fatalf("addition order %v; want %v", c.AdditionAbilityIds, want)
	}
	if want := `{"target":[220]}`; c.Base.SortOrderOwnedAbilityIdsInternal != want {
		t.Fatalf("sort order %s; want %s", c.Base.SortOrderOwnedAbilityIdsInternal, want)
	}
	keys, spells := dictionary(t, c)
	want := []save.Ability{{AbilityID: 13, ContentID: 220, SkillLevel: 270}, {AbilityID: 21, ContentID: 228}}
	if !slices.Equal(keys, []int{3}) || !slices.Equal(spells, want) {
		t.Fatalf("dictionary %v %v; want [3] %v", keys, spells, want)
	}
}

func TestPreSaveFF2NoSpellsLeft(t *testing.T) {
	c := ff2Character(t)
	c.Abilities = c.Abilities[:1]
	if err := (&Save{}).preCharacterSaveFF2(c); err != nil {
		t.Fatal(err)
	}
	if want := `{"keys":[],"values":[]}`; c.Base.AbilityDictionaryInternal != want {
		t.Fatalf("dictionary %s; want %s", c.Base.AbilityDictionaryInternal, want)
	}
	if !slices.Equal(c.AdditionAbilityIds, []int{937}) {
		t.Fatalf("addition order %v", c.AdditionAbilityIds)
	}
}

func TestPreSaveFF2FirstSpell(t *testing.T) {
	c := &Character{
		Base: &save.OwnedCharacter{
			AbilityDictionaryInternal:        `{"keys":[],"values":[]}`,
			SortOrderOwnedAbilityIdsInternal: `{"target":[]}`,
		},
		Abilities:          []*save.Ability{{AbilityID: 711, ContentID: 937}, {AbilityID: 5, SkillLevel: 0}},
		AdditionAbilityIds: []int{937},
	}
	if err := (&Save{}).preCharacterSaveFF2(c); err != nil {
		t.Fatal(err)
	}
	keys, spells := dictionary(t, c)
	if !slices.Equal(keys, []int{3}) || !slices.Equal(spells, []save.Ability{{AbilityID: 5, ContentID: 212}}) {
		t.Fatalf("dictionary %v %v", keys, spells)
	}
	if c.Base.SortOrderOwnedAbilityIdsInternal != `{"target":[]}` {
		t.Fatalf("sort order rewritten: %s", c.Base.SortOrderOwnedAbilityIdsInternal)
	}
}
