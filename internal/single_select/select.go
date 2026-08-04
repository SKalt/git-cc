package single_select

import (
	"fmt"
	"io"
	"iter"
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/skalt/git-cc/internal/controls"
	"github.com/skalt/git-cc/internal/utils"
)

// ListItem implements list.DefaultItem with a title and a description (hint).
type ListItem [2]string

func (i ListItem) FilterValue() string { return i[0] }

type Model struct {
	list   list.Model
	value  string
	logger *slog.Logger
}

// FullHelp implements [help.KeyMap].
func (m Model) FullHelp() [][]key.Binding { return [][]key.Binding{m.ShortHelp()} }

// ShortHelp implements [help.KeyMap].
func (m Model) ShortHelp() (keys []key.Binding) {
	return append(keys, controls.Keymap.Up, controls.Keymap.Down)
}

var _ help.KeyMap = Model{}

func (m Model) Init() tea.Cmd {
	return nil
}

func maxHeight(h int) (max int) {
	const otherLines = 3 // help + input + context
	return h - otherLines
}

func toListItems(input []ListItem) []list.Item {
	items := make([]list.Item, 0, len(input))
	for _, i := range input {
		items = append(items, i)
	}
	return items
}

func NewModel(
	title string,
	value string,
	options []ListItem,
	logger *slog.Logger,
) (m Model) {
	m.logger = logger.With("component", "single_select")
	m.value = value
	optWidth := 0
	for _, opt := range options {
		optWidth = max(optWidth, len(opt[0]))
	}
	delegate := newSelectDelegate(optWidth + 1)
	w := 80
	h := 24
	m.list = list.New(toListItems(options), delegate, w, maxHeight(h))
	m.list.SetShowTitle(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowPagination(true)
	m.list.SetShowHelp(false)
	m.list.SetShowFilter(true)
	m.list.SetFilteringEnabled(true)
	m.list.Title = title // this gets hidden when filtering, so we have to re-render it
	m.list.SetFilterText(value)
	m.list.SetFilterState(list.Filtering)
	m.list.FilterInput.Placeholder = "type to select"
	m.list.FilterInput.Prompt = " "            // to align with the list column
	m.list.FilterInput.ShowSuggestions = false // since the filtered list already provides suggestions
	m.logger.Debug("init", "value", m.Value())
	return m
}

func (m *Model) Focus() tea.Cmd { return m.list.FilterInput.Focus() }
func (m Model) Focused() bool   { return m.list.FilterInput.Focused() }
func (m Model) Blur()           { m.list.FilterInput.Blur() }
func (m Model) Options() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, o := range m.list.Items() {
			if t := o.(ListItem)[0]; t != "" && !yield(t) {
				return
			}
		}
	}
}
func (m Model) SetErr(err error) Model {
	m.list.FilterInput.Err = err
	return m
}

// Value returns the selected item's title, or "" if nothing is selected.
// This will never return an invalid non-empty string.
func (m Model) Value() (value string) {
	selected, ok := m.list.SelectedItem().(ListItem)
	m.logger.Debug("value", "selected", selected, "ok", ok)
	if ok && m.value == selected[0] {
		value = selected[0]
	}
	return value
}

func (m Model) CurrentInput() string { return m.list.FilterValue() }

// UpdateItems replaces the list items, updates the match function, and re-applies the filter.
func (m *Model) UpdateItems(options []ListItem) tea.Cmd {
	return m.list.SetItems(toListItems(options))
}

// selectDelegate renders list items with hints.
type selectDelegate struct {
	height   int
	spacing  int
	optWidth int
	styles   selectDelegateStyles
}

type selectDelegateStyles struct {
	selected, normal lipgloss.Style
}

func newSelectDelegate(optWidth int) selectDelegate {
	return selectDelegate{
		height:   1,
		spacing:  0,
		optWidth: optWidth,
		styles: selectDelegateStyles{
			selected: lipgloss.NewStyle().Underline(true).Bold(true),
			normal:   lipgloss.NewStyle().Faint(true),
		},
	}
}

var _ list.ItemDelegate = selectDelegate{}

func (d selectDelegate) Height() int  { return d.height }
func (d selectDelegate) Spacing() int { return d.spacing }

func (d selectDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	switch t := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetWidth(t.Width)
		m.SetHeight(t.Height)
	}
	return nil
}

func (d selectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if m.Width() <= 0 {
		return
	}

	isSelected := index == m.Index()
	i := item.(ListItem)
	title, desc := i[0], i[1]

	leftGutter := 3 // "   " or " > "
	leftColumn := d.optWidth
	rightColumn := m.Width() - leftGutter - leftColumn - 1

	var style, titleStyle lipgloss.Style
	var gutter string
	if isSelected {
		style = d.styles.selected
		titleStyle = style.Bold(true)
		gutter = ">"
	} else {
		style = d.styles.normal
		gutter = " "
	}
	str := lipgloss.JoinHorizontal(lipgloss.Top,
		style.UnsetUnderline().Width(leftGutter).PaddingLeft(1).PaddingRight(1).Render(gutter),
		titleStyle.Width(d.optWidth).Render(title),
		style.UnsetBold().Width(rightColumn).Render(desc),
	)
	utils.Must(fmt.Fprint(w, str))
}

func (m Model) Update(msg tea.Msg) (w Model, cmd tea.Cmd) {
	switch specific := msg.(type) {
	case tea.KeyPressMsg:
		k := specific.Key()
		switch {
		case key.Matches(k, controls.Keymap.Down):
			m.list.CursorDown()
			return m, nil
		case key.Matches(k, controls.Keymap.Up):
			m.list.CursorUp()
			return m, nil
		case key.Matches(k, controls.Keymap.Next):
			m.value = m.list.SelectedItem().(ListItem)[0]
			return m, nil
		}
	case tea.WindowSizeMsg:
		h := maxHeight(specific.Height)
		m.list.SetSize(specific.Width, h)
		m.list.Help.SetWidth(specific.Width)
		return m, cmd
	}
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) Render(s *strings.Builder) {
	utils.Must(s.WriteString(m.list.Title + "\n"))
	utils.Must(s.WriteString(m.list.View()))
}
