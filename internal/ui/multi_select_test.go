package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMultiChoose_EmptyOptionsReturnsEmptyWithoutPanic(t *testing.T) {
	selected, err := MultiChoose("Select", []string{}, nil, 0)
	if err != nil {
		t.Fatalf("expected nil error for empty options, got %v", err)
	}
	if len(selected) != 0 {
		t.Fatalf("expected empty selection slice, got %#v", selected)
	}

	// Out-of-bounds initialCursor with empty options must also return safely.
	selected, err = MultiChoose("Select", nil, []bool{true}, 5)
	if err != nil {
		t.Fatalf("expected nil error for nil options, got %v", err)
	}
	if len(selected) != 0 {
		t.Fatalf("expected empty selection slice, got %#v", selected)
	}
}

func TestMultiChooseModel_UpdateGuardsEmptyOptionsAndOutOfRangeCursor(t *testing.T) {
	// Empty options with Space/Up/Down/Esc must not panic and should quit cleanly.
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeySpace, Runes: []rune{' '}},
		{Type: tea.KeyUp},
		{Type: tea.KeyDown},
		{Type: tea.KeyEsc},
	} {
		m := multiChooseModel{
			title:   "Empty",
			options: []string{},
			cursor:  0,
		}
		updated, cmd := m.Update(key)
		if cmd == nil {
			t.Fatalf("expected quit cmd for key %v on empty options", key)
		}
		fm := updated.(multiChooseModel)
		if !fm.quitting {
			t.Fatalf("expected model to quit on empty options")
		}
		if key.Type == tea.KeyEsc && !fm.canceled {
			t.Fatalf("expected canceled=true on esc")
		}
	}

	// Out-of-bounds cursor with non-empty options must clamp safely without panic.
	m := multiChooseModel{
		title:    "NonEmpty",
		options:  []string{"one", "two"},
		selected: []bool{false, false},
		cursor:   99,
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	fm := updated.(multiChooseModel)
	if fm.cursor != 1 {
		t.Fatalf("expected cursor clamped to 1, got %d", fm.cursor)
	}
	if !fm.selected[1] {
		t.Fatalf("expected option 1 to be toggled")
	}

	mNeg := multiChooseModel{
		title:    "NegCursor",
		options:  []string{"one", "two"},
		selected: []bool{false, false},
		cursor:   -5,
	}
	updatedNeg, _ := mNeg.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	fmNeg := updatedNeg.(multiChooseModel)
	if fmNeg.cursor != 0 {
		t.Fatalf("expected negative cursor clamped to 0, got %d", fmNeg.cursor)
	}
	if !fmNeg.selected[0] {
		t.Fatalf("expected option 0 to be toggled")
	}
}

func TestMultiChooseModel_ViewGuardsShorterSelectedSlice(t *testing.T) {
	m := multiChooseModel{
		title:    "ShortSelected",
		options:  []string{"one", "two", "three"},
		selected: []bool{true},
		cursor:   1,
	}
	out := m.View()
	if out == "" {
		t.Fatalf("expected non-empty View output")
	}

	mNil := multiChooseModel{
		title:    "NilSelected",
		options:  []string{"one", "two"},
		selected: nil,
		cursor:   0,
	}
	if outNil := mNil.View(); outNil == "" {
		t.Fatalf("expected non-empty View output for nil selected slice")
	}
}
