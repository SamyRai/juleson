package conflict

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSpinnerModel_Init(t *testing.T) {
	m := InitialSpinnerModel("Loading...")
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a non-nil command")
	}
}

func TestSpinnerModel_Update_QuitMsg(t *testing.T) {
	m := InitialSpinnerModel("Loading...")
	newModel, cmd := m.Update(QuitMsg{})

	sm, ok := newModel.(SpinnerModel)
	if !ok {
		t.Fatalf("expected model to be SpinnerModel")
	}
	if !sm.quitting {
		t.Errorf("expected quitting to be true")
	}
	if cmd == nil {
		t.Errorf("expected tea.Quit command, got nil")
	}
}

func TestSpinnerModel_Update_KeyMsg(t *testing.T) {
	m := InitialSpinnerModel("Loading...")

	// Test 'q'
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	sm, ok := newModel.(SpinnerModel)
	if !ok || !sm.quitting || cmd == nil {
		t.Errorf("expected 'q' to quit, got model: %v, cmd: %v", sm, cmd)
	}
}

func TestSpinnerModel_View(t *testing.T) {
	m := InitialSpinnerModel("Loading...")
	view := m.View()
	if !strings.Contains(view, "Loading...") {
		t.Errorf("expected view to contain message, got: %s", view)
	}

	m.quitting = true
	if m.View() != "" {
		t.Errorf("expected empty view when quitting, got: %s", m.View())
	}
}
