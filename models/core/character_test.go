package core

import (
	"testing"

	"pixel-remastered-save-editor/save"
)

func abilityIDs(l []*save.Ability) []int {
	ids := make([]int, len(l))
	for i, a := range l {
		ids[i] = a.AbilityID
	}
	return ids
}

func TestAddAbilityIsVisibleAndNotDuplicated(t *testing.T) {
	c := &Character{Abilities: []*save.Ability{{AbilityID: 1}}}

	added := &save.Ability{AbilityID: 2, SkillLevel: 100}
	c.AddAbility(added)
	if a, ok := c.GetAbility(2); !ok || a != added {
		t.Fatalf("GetAbility(2) = %v, %v; want the just-added ability", a, ok)
	}
	if !c.HasAbility(2) {
		t.Fatal("HasAbility(2) = false after AddAbility")
	}

	// Same pointer, same ID with another pointer, and an ID loaded from the save.
	c.AddAbility(added)
	c.AddAbility(&save.Ability{AbilityID: 2})
	c.AddAbility(&save.Ability{AbilityID: 1})
	c.AddAbility(nil)
	if got := abilityIDs(c.Abilities); len(got) != 2 {
		t.Fatalf("abilities = %v; want [1 2]", got)
	}
}

func TestAddAbilityPlaceholders(t *testing.T) {
	c := &Character{}
	p1, p2 := &save.Ability{}, &save.Ability{}
	c.AddAbility(p1)
	c.AddAbility(p2)
	if len(c.Abilities) != 2 {
		t.Fatalf("got %d abilities; want 2 distinct placeholders", len(c.Abilities))
	}

	// The editors write the ID in place once the user fills the row.
	p1.AbilityID = 7
	if a, ok := c.GetAbility(7); !ok || a != p1 {
		t.Fatal("ability not found by the ID set after it was added")
	}
	if got := abilityIDs(c.AbilitiesToSave()); len(got) != 1 || got[0] != 7 {
		t.Fatalf("AbilitiesToSave = %v; want [7] (placeholders skipped)", got)
	}

	// Removing the second placeholder must not remove the first row's ability.
	c.RemoveAbility(p2)
	if len(c.Abilities) != 1 || c.Abilities[0] != p1 {
		t.Fatalf("abilities = %v; want only p1", abilityIDs(c.Abilities))
	}
	c.RemoveAbility(p2) // already gone: no-op
	if len(c.Abilities) != 1 {
		t.Fatal("removing an absent placeholder removed something")
	}
}

func TestRemoveAbilityKeepsLookupConsistent(t *testing.T) {
	c := &Character{}
	for _, id := range []int{1, 2, 3} {
		c.AddAbility(&save.Ability{AbilityID: id})
	}

	// Removal by ID with a different pointer (as the FF5 magic checkboxes do).
	c.RemoveAbility(&save.Ability{AbilityID: 2})
	if c.HasAbility(2) {
		t.Fatal("HasAbility(2) = true after RemoveAbility")
	}
	// Re-adding after a removal must work and not duplicate.
	c.AddAbility(&save.Ability{AbilityID: 2})
	c.AddAbility(&save.Ability{AbilityID: 2})
	if got := abilityIDs(c.Abilities); len(got) != 3 || got[2] != 2 {
		t.Fatalf("abilities = %v; want [1 3 2]", got)
	}

	c.RemoveAbilityIndex(0)
	if c.HasAbility(1) {
		t.Fatal("HasAbility(1) = true after RemoveAbilityIndex")
	}
	c.RemoveAbilityIndex(-1)
	c.RemoveAbilityIndex(10)
	if got := abilityIDs(c.Abilities); len(got) != 2 || got[0] != 3 || got[1] != 2 {
		t.Fatalf("abilities = %v; want [3 2]", got)
	}
}

func TestAbilitiesToSaveDropsDuplicateIDs(t *testing.T) {
	first := &save.Ability{AbilityID: 4, SkillLevel: 50}
	c := &Character{Abilities: []*save.Ability{first, {}, {AbilityID: 5}}}
	// A placeholder filled in with an ID that is already present.
	c.Abilities[1].AbilityID = 4
	got := c.AbilitiesToSave()
	if len(got) != 2 || got[0] != first || got[1].AbilityID != 5 {
		t.Fatalf("AbilitiesToSave = %v; want [4 5] keeping the first 4", abilityIDs(got))
	}
}
