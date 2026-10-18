package core

import (
	"slices"
	"testing"

	"pixel-remastered-save-editor/global"
	"pixel-remastered-save-editor/save"
)

func TestInventoryResetSortOrder(t *testing.T) {
	items := []*save.OwnedItems{{ContentID: 47, Count: 3}, {ContentID: 17, Count: 1}, {ContentID: 22, Count: 1}}
	saved := []int{17, 22}
	inv := NewInventory(global.Two, items, saved)
	// Edits in the editor: a cleared row, an item added to a padding row and
	// a duplicate.
	inv.Rows[2].ContentID = 0
	inv.Rows[len(inv.Rows)-1].ContentID = 30
	inv.Rows[len(inv.Rows)-2].ContentID = 17

	_, order, err := inv.ToSave()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(order, saved) {
		t.Errorf("order without reset = %v, want %v", order, saved)
	}

	inv.ResetSortOrder = true
	_, order, err = inv.ToSave()
	if err != nil {
		t.Fatal(err)
	}
	if want := []int{47, 17, 30}; !slices.Equal(order, want) {
		t.Errorf("reset order = %v, want %v", order, want)
	}
}

func TestInventoryResetSortOrderDebugSort(t *testing.T) {
	t.Setenv("PR_SORT_INV", "true")
	items := []*save.OwnedItems{{ContentID: 47, Count: 3}, {ContentID: 17, Count: 1}}
	inv := NewInventory(global.Two, items, []int{17})
	inv.ResetSortOrder = true
	rows, order, err := inv.ToSave()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].ContentID != 17 {
		t.Fatalf("rows not sorted: %v", rows)
	}
	if want := []int{17, 47}; !slices.Equal(order, want) {
		t.Errorf("reset order = %v, want %v", order, want)
	}
}
