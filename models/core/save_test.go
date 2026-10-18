package core

import (
	"slices"
	"testing"

	"pixel-remastered-save-editor/save"
)

func ff1Slot(ids ...int) *save.AbilitySlotData {
	s := &save.AbilitySlot{}
	for _, id := range ids {
		s.Values = append(s.Values, &save.Ability{AbilityID: id})
	}
	return &save.AbilitySlotData{SlotInfo: s}
}

func contentIDs(l []*save.Ability) (ids []int) {
	for _, a := range l {
		if a.AbilityID > 0 {
			ids = append(ids, a.ContentID)
		}
	}
	return
}

func TestPreCharacterSaveFF1NoDuplicateIDs(t *testing.T) {
	c := &Character{
		// Spell 1 (content 209) is owned; 5 (213) and 7 (215) are not.
		Abilities:          []*save.Ability{{AbilityID: 1, ContentID: 209}, {}},
		AdditionAbilityIds: []int{209},
		AbilitySlotData: []*save.AbilitySlotData{
			ff1Slot(1, 5, 0),
			// The same new spell set in two slots.
			ff1Slot(7, 7),
			// Malformed slot data is skipped.
			{SlotInfo: &save.AbilitySlot{Values: []*save.Ability{nil}}},
			{},
		},
	}
	if err := (&Save{}).preCharacterSaveFF1(c); err != nil {
		t.Fatal(err)
	}
	if want := []int{209, 213, 215}; !slices.Equal(c.AdditionAbilityIds, want) {
		t.Fatalf("AdditionAbilityIds = %v; want %v", c.AdditionAbilityIds, want)
	}
	if got, want := contentIDs(c.Abilities), []int{209, 213, 215}; !slices.Equal(got, want) {
		t.Fatalf("abilities content IDs = %v; want %v", got, want)
	}

	// Running it again (a second save) must not add anything.
	if err := (&Save{}).preCharacterSaveFF1(c); err != nil {
		t.Fatal(err)
	}
	if len(c.AdditionAbilityIds) != 3 || len(c.Abilities) != 4 {
		t.Fatalf("second pass changed the lists: %v, %d abilities", c.AdditionAbilityIds, len(c.Abilities))
	}
}

func TestPreCharacterSaveFF1EditorAddedAbilities(t *testing.T) {
	// Abilities added with the editor's "+" button have no content ID.
	added := &save.Ability{AbilityID: 10}
	inSlot := &save.Ability{AbilityID: 7}
	c := &Character{
		Abilities:       []*save.Ability{added, inSlot, {}},
		AbilitySlotData: []*save.AbilitySlotData{ff1Slot(7)},
	}
	if err := (&Save{}).preCharacterSaveFF1(c); err != nil {
		t.Fatal(err)
	}
	if added.ContentID != 218 || inSlot.ContentID != 215 {
		t.Fatalf("content IDs = %d, %d; want 218, 215", added.ContentID, inSlot.ContentID)
	}
	if want := []int{215, 218}; !slices.Equal(c.AdditionAbilityIds, want) {
		t.Fatalf("AdditionAbilityIds = %v; want %v", c.AdditionAbilityIds, want)
	}
	// The slot spell is already owned: no second ability 7.
	if got, want := contentIDs(c.Abilities), []int{218, 215}; !slices.Equal(got, want) {
		t.Fatalf("abilities content IDs = %v; want %v", got, want)
	}
}
