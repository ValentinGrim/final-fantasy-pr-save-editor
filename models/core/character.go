package core

import (
	"slices"

	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
)

type (
	Character struct {
		Base               *save.OwnedCharacter
		Parameters         *save.CharacterParameters
		Commands           []int
		Abilities          []*save.Ability
		AbilitySlotData    []*save.AbilitySlotData
		Jobs               []*save.Job
		Equipment          *save.EquipmentList
		AdditionAbilityIds []int
		// SortOrderOwnedAbilityIdsInternal ???
		// AbilityDictionary *save.AbilityDictionary
		SkillLevelTargets *save.SkillLevelTargets
		// LearningAbilities ???
		EquipmentAbilities []int
		EnableCommandsSave bool
		// LearningAbilities  []int
	}
	Characters struct {
		characters []*Character
	}
)

func NewCharacters(game global.Game, in *save.OwnedCharacterList) (c *Characters, err error) {
	var l []*save.OwnedCharacter
	if l, err = in.OwnedCharacters(); err != nil {
		return
	}
	c = &Characters{characters: make([]*Character, len(l))}
	for i, j := range l {
		if c.characters[i], err = NewCharacter(game, j); err != nil {
			return
		}
	}
	return
}

func NewCharacter(game global.Game, in *save.OwnedCharacter) (c *Character, err error) {
	c = &Character{Base: in}

	if c.Parameters, err = in.CharacterParameters(); err != nil {
		return
	}

	var cl *save.CommandList
	if cl, err = in.CommandList(); err != nil {
		return
	}
	c.Commands = cl.Target
	if c.Abilities, err = in.Abilities(); err != nil {
		return
	}
	if c.AbilitySlotData, err = in.AbilitySlotData(); err != nil {
		return
	}
	if c.AdditionAbilityIds, err = c.Base.AdditionOrderOwnedAbilityIds(); err != nil {
		return
	}
	// if c.AbilityDictionary, err = in.AbilityDictionary(); err != nil {
	// 	return
	// }
	// if c.LearningAbilities, err = in.LearningAbilities(); err != nil {
	// 	return
	// }
	if c.SkillLevelTargets, err = in.SkillLevelTargets(); err != nil {
		return
	}
	if c.Equipment, err = in.Equipment(); err != nil {
		return
	}
	if c.Jobs, err = in.Jobs(); err != nil {
		return
	}
	return
}

func (c *Characters) All() []*Character {
	return c.characters
}

func (c *Characters) Names() []string {
	names := make([]string, 0, len(c.characters))
	for _, character := range c.characters {
		if character != nil {
			names = append(names, character.Base.Name)
		}
	}
	slices.Sort(names)
	return names
}

func (c *Characters) GetByID(id int) (character *Character, found bool) {
	for _, character = range c.characters {
		if character != nil && character.Base.ID == id {
			return character, true
		}
	}
	return nil, false
}

func (c *Characters) GetByName(name string) (character *Character, found bool) {
	for _, character = range c.characters {
		if character != nil && character.Base.Name == name {
			return character, true
		}
	}
	return nil, false
}

// Abilities are looked up by scanning c.Abilities rather than through a cached
// map: the editors bind entries directly to Ability.AbilityID, so an ID can
// change after the ability was added, and the save code appends to
// c.Abilities directly. A scan always reflects the current state.

// AddAbility appends ability unless it, or another ability with the same
// non-zero ID, is already present. Abilities with ID 0 are placeholders for
// rows the user has not filled in yet; several may coexist and none is saved.
func (c *Character) AddAbility(ability *save.Ability) {
	if ability == nil {
		return
	}
	for _, a := range c.Abilities {
		if a == ability || (ability.AbilityID != 0 && a.AbilityID == ability.AbilityID) {
			return
		}
	}
	c.Abilities = append(c.Abilities, ability)
}

// RemoveAbility removes ability itself if present, otherwise the first
// ability sharing its non-zero ID.
func (c *Character) RemoveAbility(ability *save.Ability) {
	if ability == nil {
		return
	}
	if i := slices.Index(c.Abilities, ability); i >= 0 {
		c.RemoveAbilityIndex(i)
		return
	}
	if ability.AbilityID == 0 {
		return
	}
	if i := slices.IndexFunc(c.Abilities, func(a *save.Ability) bool { return a.AbilityID == ability.AbilityID }); i >= 0 {
		c.RemoveAbilityIndex(i)
	}
}

func (c *Character) RemoveAbilityIndex(index int) {
	if index < 0 || index >= len(c.Abilities) {
		return
	}
	c.Abilities = slices.Delete(c.Abilities, index, index+1)
}

// AbilitiesToSave returns the abilities to write: placeholders (ID 0) are
// skipped and only the first ability of each ID is kept.
func (c *Character) AbilitiesToSave() []*save.Ability {
	a := make([]*save.Ability, 0, len(c.Abilities))
	seen := make(map[int]bool, len(c.Abilities))
	for _, i := range c.Abilities {
		if i.AbilityID != 0 && !seen[i.AbilityID] {
			seen[i.AbilityID] = true
			a = append(a, i)
		}
	}
	return a
}

func (c *Character) HasAbility(id int) bool {
	_, ok := c.GetAbility(id)
	return ok
}

func (c *Character) GetAbility(id int) (a *save.Ability, ok bool) {
	for _, a = range c.Abilities {
		if a.AbilityID == id {
			return a, true
		}
	}
	return nil, false
}

func (c *Character) ToSave() (s string, err error) {
	if err = c.Base.SetCharacterParameters(c.Parameters); err != nil {
		return
	}
	if c.EnableCommandsSave {
		if err = c.Base.SetCommandList(c.Commands); err != nil {
			return
		}
	}
	if err = c.Base.SetAbilityList(c.AbilitiesToSave()); err != nil {
		return
	}
	if err = c.Base.SetAbilitySlotData(c.AbilitySlotData); err != nil {
		return
	}
	if err = c.Base.SetAdditionOrderOwnedAbilityIds(c.AdditionAbilityIds); err != nil {
		return
	}
	// if err = c.Base.SetAbilityDictionary(c.AbilityDictionary); err != nil {
	// 	return
	// }
	// if err = c.Base.SetLearningAbilities(c.LearningAbilities); err != nil {
	// 	return
	// }
	if err = c.Base.SetJobs(c.Jobs); err != nil {
		return
	}
	if err = c.Base.SetSkillLevelTargets(c.SkillLevelTargets); err != nil {
		return
	}
	if err = c.Base.SetEquipment(c.Equipment); err != nil {
		return
	}
	return save.MarshalOne(c.Base)
}
