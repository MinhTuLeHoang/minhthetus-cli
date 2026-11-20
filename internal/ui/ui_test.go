package ui

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestChooseWithDefault_EmptyOptionsReturnsError(t *testing.T) {
	if _, err := ChooseWithDefault("Pick one", []string{}, 0); err == nil {
		t.Fatalf("expected error for empty options in ChooseWithDefault")
	}
	if _, err := Choose("Pick one", nil); err == nil {
		t.Fatalf("expected error for nil options in Choose")
	}
}

func TestMultiChoose_EmptyOptionsReturnsError(t *testing.T) {
	if _, err := MultiChoose("Select", []string{}, nil, 0); err == nil {
		t.Fatalf("expected error for empty options in MultiChoose")
	}
}

func TestMultiChooseModel_ClampsInitialCursorAndToggles(t *testing.T) {
	m1 := newMultiChooseModel("Title", []string{"a", "b"}, []bool{true}, -5)
	if m1.cursor != 0 {
		t.Fatalf("expected negative initialCursor clamped to 0, got %d", m1.cursor)
	}

	m2 := newMultiChooseModel("Title", []string{"a", "b"}, []bool{false, true}, 99)
	if m2.cursor != 0 {
		t.Fatalf("expected out-of-range initialCursor clamped to 0, got %d", m2.cursor)
	}

	// Toggle space at cursor 0
	next, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m3 := next.(multiChooseModel)
	if !m3.selected[0] || !m3.selected[1] {
		t.Fatalf("expected both items selected after space toggle, got %v", m3.selected)
	}

	// Esc cancels cleanly
	escNext, _ := m3.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m4 := escNext.(multiChooseModel)
	if !m4.canceled || !m4.quitting {
		t.Fatalf("expected canceled=true, quitting=true on Esc")
	}
}

func TestInputModel_CancelAndSubmit(t *testing.T) {
	ti := textinput.New()
	ti.SetValue("1.2.3")
	m := inputModel{textInput: ti}

	// Esc marks canceled=true
	escNext, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	fmEsc := escNext.(inputModel)
	if !fmEsc.canceled || !fmEsc.quitting {
		t.Fatalf("expected canceled=true and quitting=true on Esc")
	}

	// Enter captures value with canceled=false
	enterNext, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	fmEnter := enterNext.(inputModel)
	if fmEnter.canceled || !fmEnter.quitting || fmEnter.value != "1.2.3" {
		t.Fatalf("expected value='1.2.3' and canceled=false on Enter, got %+v", fmEnter)
	}
}

func forceNonTTYStdio(t *testing.T) {
	t.Helper()
	origStdin := os.Stdin
	origStdout := os.Stdout

	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe stdin failed: %v", err)
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		_ = inR.Close()
		_ = inW.Close()
		t.Fatalf("os.Pipe stdout failed: %v", err)
	}

	os.Stdin = inR
	os.Stdout = outW

	t.Cleanup(func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
		_ = inR.Close()
		_ = inW.Close()
		_ = outR.Close()
		_ = outW.Close()
	})
}

func TestPromptsReturnErrNonInteractiveTTYWithoutTerminal(t *testing.T) {
	forceNonTTYStdio(t)

	if isInteractiveTTY() {
		t.Fatalf("expected isInteractiveTTY() == false when stdin/stdout are pipes")
	}

	if _, err := ChooseWithDefault("Pick one", []string{"a", "b"}, 0); !errors.Is(err, ErrNonInteractiveTTY) {
		t.Fatalf("ChooseWithDefault err = %v, want ErrNonInteractiveTTY", err)
	}
	if _, err := Choose("Pick one", []string{"a", "b"}); !errors.Is(err, ErrNonInteractiveTTY) {
		t.Fatalf("Choose err = %v, want ErrNonInteractiveTTY", err)
	}
	if _, err := MultiChoose("Pick many", []string{"a", "b"}, nil, 0); !errors.Is(err, ErrNonInteractiveTTY) {
		t.Fatalf("MultiChoose err = %v, want ErrNonInteractiveTTY", err)
	}
	if _, err := Input("Enter value", ""); !errors.Is(err, ErrNonInteractiveTTY) {
		t.Fatalf("Input err = %v, want ErrNonInteractiveTTY", err)
	}
	if confirmed, err := Confirm("Proceed?", 0, false); confirmed || !errors.Is(err, ErrNonInteractiveTTY) {
		t.Fatalf("Confirm(!autoApprove) = (%v, %v), want (false, ErrNonInteractiveTTY)", confirmed, err)
	}
	if confirmed, err := Confirm("Proceed?", 3*time.Second, true); !confirmed || err != nil {
		t.Fatalf("Confirm(autoApprove) = (%v, %v), want (true, nil)", confirmed, err)
	}
}
