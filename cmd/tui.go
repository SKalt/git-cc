package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/skalt/git-cc/internal/breaking_change_input"
	"github.com/skalt/git-cc/internal/config"
	"github.com/skalt/git-cc/internal/controls"
	"github.com/skalt/git-cc/internal/description_editor"
	"github.com/skalt/git-cc/internal/scope_selector"
	"github.com/skalt/git-cc/internal/type_selector"
	"github.com/skalt/git-cc/internal/utils"
	"github.com/skalt/git-cc/pkg/parser"
)

type componentIndex int

const ( // the order of the components
	commitTypeIndex componentIndex = iota
	scopeIndex
	shortDescriptionIndex
	breakingChangeIndex
	// body omitted -- performed by GIT_EDITOR
	nIndices // the number of indices
)

var (
	boolFlags = [...]string{
		"all",
		"signoff",
		"no-signoff",
		"no-post-rewrite",
		"no-gpg-sign",
		"no-verify", // https://git-scm.com/docs/git-commit#Documentation/git-commit.txt---no-verify
		"allow-empty",
	}
)

// TODO: light/dark styles
// var styles = struct{dark, light lipgloss.Style}{
// 	dark: lipgloss.NewStyle(),
// 	light: lipgloss.NewStyle(),
// }

// FIXME: this should be referenceable by all the input components :/

type model struct {
	viewing componentIndex

	typeInput           type_selector.Model
	scopeInput          scope_selector.Model
	descriptionInput    description_editor.Model
	breakingChangeInput breaking_change_input.Model
	// any body stashed during the initial parse of command-line --message args
	remainingBody string
	help          help.Model
	height        int
}

var _ tea.Model = model{}

// returns whether the minimum requirements for a conventional commit are met.
func (m model) ready() bool {
	return m.typeInput.Value() != "" && m.descriptionInput.Value() != ""

}

func (m model) renderContextValue(s *strings.Builder) {
	utils.Must(s.WriteString(m.typeInput.Value()))
	scope := m.scopeInput.Value()
	breakingChange := m.breakingChangeInput.Value()
	if scope != "" {
		utils.Must(fmt.Fprintf(s, "(%s)", scope))
	}
	if breakingChange != "" {
		utils.Must(s.WriteRune('!'))
	}
	utils.Must(s.WriteString(": "))
}

// returns the context portion of the CC header, e.g `type(scope): `.
func (m model) contextValue() string { return utils.Render(m.renderContextValue) }
func (m model) descriptionValue() string {
	return m.descriptionInput.Value()
}
func (m model) breakingChangeValue() string {
	return m.breakingChangeInput.Value()
}

func (m model) withoutBreakingChange() (w model) {
	w = m // clone by value
	w.breakingChangeInput = breaking_change_input.Model{}
	return w
}

func (m model) renderValue(s *strings.Builder) {
	for _, p := range [...]string{
		m.contextValue(), m.descriptionValue(),
		"\n",
		m.remainingBody,
		"\n",
	} {
		utils.Must(s.WriteString(p))
	}
	if breakingChange := m.breakingChangeValue(); breakingChange != "" {
		// TODO: handle multiple breaking change footers(?)
		utils.Must(fmt.Fprintf(s, "\nBREAKING CHANGE: %s\n", breakingChange))
	}
}

// Returns a pretty-printed CC string. The model should be `.ready()` before you call `.value()`.
func (m model) value() string { return utils.Render(m.renderValue) }

// hacky globals
var (
	debugLogFile *os.File
	logger       *slog.Logger
)

func initLogger() {
	var l io.Writer
	if logfile := getLogFile(); logfile == "" {
		l = io.Discard
	} else {
		debugLogFile = utils.Must(tea.LogToFile(logfile, "tui"))
		l = debugLogFile
	}
	h := slog.NewTextHandler(l, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger = slog.New(h)
}

func (m model) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m model) currentComponent() controls.InputComponent {
	return [...]controls.InputComponent{
		&m.typeInput,
		&m.scopeInput,
		&m.descriptionInput,
		&m.breakingChangeInput,
	}[m.viewing]
}

// Pass a channel to the model to listen to the result value. This is a
// function that returns the initialize function and is typically how you would
// pass arguments to a tea.Init function.
func initialModel(cc *parser.CC, cfg *config.Cfg) model {
	typeModel := type_selector.NewModel(cc, cfg)
	scopeModel := scope_selector.NewModel(cc, *cfg)
	descModel := description_editor.NewModel(cc, cfg.HeaderMaxLength, cfg.EnforceMaxLength)
	bcModel := breaking_change_input.NewModel(cc)
	breakingChanges := ""
	if cc.BreakingChange {
		for _, footer := range cc.Footers {
			result, err := parser.BreakingChange([]rune(footer))
			if err == nil {
				breakingChanges += string(result.Remaining) + "\n"
			}
		}
	}
	m := model{
		typeInput:           typeModel,
		scopeInput:          scopeModel,
		descriptionInput:    descModel,
		breakingChangeInput: bcModel,
		viewing:             commitTypeIndex,
		help:                help.New(),
	}
	if cc.Body != nil {
		m.remainingBody = *cc.Body
	}
	m, _ = m.advance() // hmm, this doesn't seem right
	m.descriptionInput = m.descriptionInput.SetPrefix(m.contextValue())
	return m
}

// pass the `msg` to the currently-displayed component/view
func (m model) updateCurrentInput(msg tea.Msg) (w model, cmd tea.Cmd) {
	switch m.viewing {
	case commitTypeIndex:
		m.typeInput, cmd = m.typeInput.Update(msg)
	case scopeIndex:
		m.scopeInput, cmd = m.scopeInput.Update(msg)
	case shortDescriptionIndex:
		m.descriptionInput, cmd = m.descriptionInput.Update(msg)
	case breakingChangeIndex:
		m.breakingChangeInput, cmd = m.breakingChangeInput.Update(msg)
	}
	m.descriptionInput = m.descriptionInput.SetPrefix(m.contextValue())
	return m, cmd
}

func (m model) shouldSkip() (shouldSkip bool) {
	return m.currentComponent().Ready()
}

func (m model) advance() (model, tea.Cmd) {
	for m.viewing < breakingChangeIndex && m.shouldSkip() {
		logger.Debug("advance", "index", m.viewing)
		m.viewing++
	}
	if m.viewing == breakingChangeIndex && m.shouldSkip() {
		return m, tea.Quit
	}
	return m, nil
}

func getLogFile() string {
	d := os.Getenv("GIT_CC_DEBUG")
	switch d {
	case "true", "TRUE", "1":
		return "debug.log"
	case "", "false", "FALSE", "0":
		return ""
	default:
		return d
	}
}

func (m model) Update(msg tea.Msg) (w tea.Model, cmd tea.Cmd) {
	logger.Debug("update", "msg", fmt.Sprintf("%#v", msg))
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		k := msg.Key()
		switch {
		case key.Matches(k, controls.Keymap.Cancel):
			return m, tea.Quit
		case key.Matches(k, controls.Keymap.Back):
			if m.viewing > commitTypeIndex {
				m.viewing--
			}
			return m, cmd
		case key.Matches(k, controls.Keymap.Next):
			cmds := make([]tea.Cmd, 2)
			m, cmds[0] = m.updateCurrentInput(msg)

			if m.currentComponent().Ready() {
				m, cmds[1] = m.advance()
			}
			return m, tea.Batch(cmds...)
		}
	case tea.WindowSizeMsg:
		// ensure instances of tea.WindowSizeMsg reach all child-components
		cmds := make([]tea.Cmd, 4)
		m.typeInput, cmds[0] = m.typeInput.Update(msg)
		m.scopeInput, cmds[1] = m.scopeInput.Update(msg)
		m.descriptionInput, cmds[2] = m.descriptionInput.Update(msg)
		m.breakingChangeInput, cmds[3] = m.breakingChangeInput.Update(msg)
		m.help.SetWidth(msg.Width)
		m.height = msg.Height
		return m, tea.Batch(cmds...)
	}
	return m.updateCurrentInput(msg)
}

func (m model) renderCurrentComponent(s *strings.Builder) {
	style := lipgloss.NewStyle().Faint(true).Align(lipgloss.Left).PaddingRight(0)
	switch m.viewing {
	case breakingChangeIndex:
		buf := strings.Builder{}
		utils.Must(fmt.Fprintf(s, "%#v", style))
		m.withoutBreakingChange().renderValue(&buf) // !!
		utils.Must(s.WriteString((buf.String())))
	}
	m.currentComponent().Render(s)
}

func (m model) View() (v tea.View) {
	v.AltScreen = true
	s := strings.Builder{}
	m.renderCurrentComponent(&s)
	utils.Must(s.WriteString("\n"))
	lines := strings.Count(s.String(), "\n")
	padding := max(m.height-lines-1, 0)
	utils.Must(s.WriteString(strings.Repeat("\n", padding)))
	extra := []key.Binding{}
	if k, ok := m.currentComponent().(help.KeyMap); ok {
		extra = k.ShortHelp()
	}
	s.WriteString(controls.View(&m.help, extra...))
	v.Content = s.String()
	return v
}
