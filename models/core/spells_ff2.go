package core

import (
	"slices"

	"pixel-remastered-save-editor/models/core/costs/ff2"
	"pixel-remastered-save-editor/save"
)

// preCharacterSaveFF2 gives each ability its content ID and keeps the lists
// that repeat a character's abilities in step with Abilities. In the measured
// FF II saves:
//   - a spell's content ID is its ability ID + 207, so a spell whose level
//     (ability ID) changed in the editor needs a new content ID;
//   - additionOrderOwnedAbilityIds holds the content IDs of abilityList, in
//     the same order;
//   - sortOrderOwnedAbilityIds holds a subset of those content IDs;
//   - abilityDictionary key 3 holds one copy of each ability, with the same
//     IDs and a skill level at most the current one.
func (s *Save) preCharacterSaveFF2(c *Character) (err error) {
	renamed := make(map[int]int) // old content ID -> new one
	for _, a := range c.Abilities {
		if _, spell := ff2.SpellLevel(a.AbilityID); spell {
			cid := ff2.SpellContentID(a.AbilityID)
			if a.ContentID != 0 && a.ContentID != cid {
				renamed[a.ContentID] = cid
			}
			a.ContentID = cid
		} else if a.ContentID == 0 && a.AbilityID > 0 {
			a.ContentID = a.AbilityID + 207
		}
	}

	saved := c.AbilitiesToSave()
	owned := make(map[int]*save.Ability, len(saved))
	for _, a := range saved {
		owned[a.ContentID] = a
	}
	current := func(cid int) (int, bool) {
		if n, ok := renamed[cid]; ok {
			cid = n
		}
		_, ok := owned[cid]
		return cid, ok
	}

	order := make([]int, 0, len(saved))
	for _, cid := range c.AdditionAbilityIds {
		if cid, ok := current(cid); ok && !slices.Contains(order, cid) {
			order = append(order, cid)
		}
	}
	for _, a := range saved {
		if !slices.Contains(order, a.ContentID) {
			order = append(order, a.ContentID)
		}
	}
	c.AdditionAbilityIds = order

	if err = syncSortOrderFF2(c.Base, current); err != nil {
		return
	}
	return syncAbilityDictionaryFF2(c.Base, saved, current)
}

// syncSortOrderFF2 renames and drops the content IDs of
// sortOrderOwnedAbilityIds. It leaves the field untouched when nothing
// changes or when it holds something other than numbers.
func syncSortOrderFF2(c *save.OwnedCharacter, current func(int) (int, bool)) error {
	if c.SortOrderOwnedAbilityIdsInternal == "" {
		return nil
	}
	so, err := c.SortOrderOwnedAbilityIds()
	if err != nil {
		return err
	}
	ids := make([]int, len(so.Target))
	for i, v := range so.Target {
		f, ok := v.(float64)
		if !ok || f != float64(int(f)) {
			return nil
		}
		ids[i] = int(f)
	}
	kept := make([]any, 0, len(ids))
	seen := make(map[int]bool, len(ids))
	changed := false
	for _, old := range ids {
		cid, ok := current(old)
		if !ok || seen[cid] {
			changed = true
			continue
		}
		seen[cid] = true
		changed = changed || cid != old
		kept = append(kept, cid)
	}
	if !changed {
		return nil
	}
	so.Target = kept
	return c.SetSortOrderOwnedAbilityIds(so)
}

// abilityDictionaryKey is the only key of abilityDictionary in the measured
// FF II saves.
const abilityDictionaryKey = 3

// syncAbilityDictionaryFF2 makes abilityDictionary hold, under key 3, one
// copy of each saved spell in abilityList order, as in the measured saves; a
// character without spells has an empty dictionary. A renamed spell's copy
// takes the spell's new IDs and skill level, a removed spell's copy goes, an
// added spell gets one. Other copies keep their skill level. A dictionary
// with another key is left untouched.
func syncAbilityDictionaryFF2(c *save.OwnedCharacter, saved []*save.Ability, current func(int) (int, bool)) error {
	if c.AbilityDictionaryInternal == "" {
		return nil
	}
	d, err := c.AbilityDictionary()
	if err != nil {
		return err
	}
	if len(d.Keys) != len(d.Values) || len(d.Keys) > 1 || (len(d.Keys) == 1 && d.Keys[0] != abilityDictionaryKey) {
		return nil
	}
	var old []*save.Ability
	if len(d.Values) == 1 {
		old = d.Values[0]
	}

	copies := make(map[int]*save.Ability, len(old))
	changed := false
	for _, a := range old {
		cid, ok := current(a.ContentID)
		if !ok || copies[cid] != nil {
			changed = true
			continue
		}
		copies[cid] = a
	}
	list := make([]*save.Ability, 0, len(old))
	for _, a := range saved {
		if _, spell := ff2.SpellLevel(a.AbilityID); !spell {
			continue
		}
		cp, ok := copies[a.ContentID]
		if !ok || cp.ContentID != a.ContentID {
			cp = &save.Ability{AbilityID: a.AbilityID, ContentID: a.ContentID, SkillLevel: a.SkillLevel}
		}
		list = append(list, cp)
	}
	if !changed && len(list) == len(old) {
		for i, a := range list {
			if a != old[i] {
				changed = true
				break
			}
		}
	} else {
		changed = true
	}
	if !changed {
		return nil
	}
	if len(list) == 0 {
		d.Keys, d.Values = []int{}, [][]*save.Ability{}
	} else {
		d.Keys, d.Values = []int{abilityDictionaryKey}, [][]*save.Ability{list}
	}
	return c.SetAbilityDictionary(d)
}
