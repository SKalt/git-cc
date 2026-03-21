package controls

import "charm.land/bubbles/v2/key"

var Keymap = struct {
	Next, Back, Cancel key.Binding
}{
	Next:   key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("tab/enter", "next")),
	Back:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "back")),
	Cancel: key.NewBinding(key.WithKeys("ctrl+c", "ctrl+d"), key.WithHelp("ctrl+c/d", "cancel")),
}
