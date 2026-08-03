// FIXME: allow for multiple breaking change footers
package breaking_change_input

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/skalt/git-cc/internal/controls"
	"github.com/skalt/git-cc/internal/utils"
	"github.com/skalt/git-cc/pkg/parser"
)

type Model struct {
	input      textinput.Model
	hasBeenSet bool
}

var _ controls.InputComponent = &Model{}

// without the BREAKING CHANGE prefix
// Value implements [controls.InputComponent]
func (m Model) Value() string {
	return m.input.Value()
}

// render implements [controls.InputComponent]
func (m Model) Render(b *strings.Builder) {
	utils.Must(b.WriteString(m.input.View()))
	utils.Must(b.WriteString("\n\n"))
}

func (m Model) Update(msg tea.Msg) (out Model, cmd tea.Cmd) {
	out = m
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg.Key(), controls.Keymap.Next) {
			out.hasBeenSet = true
		}
	}
	out.input, cmd = m.input.Update(msg)
	return
}

func NewModel(cc *parser.CC) (m Model) {
	var value string
	footers := cc.BreakingChanges()
	if len(footers) > 0 {
		value = strings.Split(footers[0], ": ")[1]
	}
	m.input = textinput.New()
	m.input.SetValue(value)
	m.input.Prompt = lipgloss.NewStyle().Faint(true).Render("BREAKING CHANGES: ")
	m.input.Placeholder = "if any."
	m.input.Focus()
	return m
}

func (m Model) Ready() bool { return m.hasBeenSet }
