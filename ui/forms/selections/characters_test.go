package selections

import (
	"slices"
	"testing"

	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/models/core"
	"pixel-remastered-save-editor/save"
)

func chars(l ...save.OwnedCharacter) []*core.Character {
	out := make([]*core.Character, len(l))
	for i := range l {
		out[i] = &core.Character{Base: &l[i]}
	}
	return out
}

func labelsAndIDs(choices []characterChoice) (labels []string, ids []int) {
	for _, c := range choices {
		labels = append(labels, c.label)
		ids = append(ids, c.id)
	}
	return
}

func TestCharacterChoicesFF6(t *testing.T) {
	all := chars(
		save.OwnedCharacter{ID: 1, Name: "Tina"}, // renamed / localized
		save.OwnedCharacter{ID: 5, Name: "Locke"},
		save.OwnedCharacter{ID: 21, Name: "Mog"},
		save.OwnedCharacter{ID: 16, Name: "Mog"},
		save.OwnedCharacter{ID: 22, Name: ""}, // no name in the save
		save.OwnedCharacter{ID: 99, Name: "Unknown"},
	)
	all = append(all, nil, &core.Character{}) // must be ignored
	labels, ids := labelsAndIDs(characterChoices(global.Six, all))
	wantLabels := []string{"Cyan", "Locke", "Mog [#21]", "Tina (Terra)", "Mog [#16]", "Unknown"}
	wantIDs := []int{22, 5, 21, 1, 16, 99}
	if !slices.Equal(labels, wantLabels) || !slices.Equal(ids, wantIDs) {
		t.Fatalf("got %q %v; want %q %v", labels, ids, wantLabels, wantIDs)
	}
}

func TestCharacterChoicesDuplicateAndEmptyNames(t *testing.T) {
	all := chars(
		save.OwnedCharacter{ID: 3, Name: "A"},
		save.OwnedCharacter{ID: 1, Name: "A"},
		save.OwnedCharacter{ID: 2, Name: ""},
		save.OwnedCharacter{ID: 4, Name: "B"},
	)
	labels, ids := labelsAndIDs(characterChoices(global.One, all))
	wantLabels := []string{"A [#1]", "A [#3]", "B", "Character 2"}
	wantIDs := []int{1, 3, 4, 2}
	if !slices.Equal(labels, wantLabels) || !slices.Equal(ids, wantIDs) {
		t.Fatalf("got %q %v; want %q %v", labels, ids, wantLabels, wantIDs)
	}
}
