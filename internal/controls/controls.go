package controls

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

var Keymap = struct {
	Next, Back, Cancel, Up, Down key.Binding
}{
	Next:   key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("tab/enter", "next")),
	Back:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "back")),
	Cancel: key.NewBinding(key.WithKeys("ctrl+c", "ctrl+d"), key.WithHelp("ctrl+c/d", "cancel")),
	Up:     key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("up/ctrl+p", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("down/ctrl+n", "down")),
}

func View(h *help.Model, extras ...key.Binding) string {
	return h.ShortHelpView(append([]key.Binding{
		Keymap.Next,
		Keymap.Back,
		Keymap.Cancel,
	}, extras...))
}

// use this to communicate Values up the chain
type ValueMsg string

var (
	_ tea.Msg = ValueMsg("")
)

// the commonalities of the sub-components
type InputComponent interface {
	Render(*strings.Builder)
	Value() string
	Ready() bool
}

// Update(tea.Msg) (Model, tea.Cmd)
