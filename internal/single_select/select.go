package single_select

import (
	"fmt"
	"io"
	"iter"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/skalt/git-cc/internal/controls"
	"github.com/skalt/git-cc/internal/utils"
)

// ListItem implements list.DefaultItem with a title and a description (hint).
type ListItem [2]string

func (i ListItem) FilterValue() string { return i[0] }

type Model struct {
	list  list.Model
	value string
}

// FullHelp implements [help.KeyMap].
func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

// ShortHelp implements [help.KeyMap].
func (m Model) ShortHelp() (keys []key.Binding) {
	keys = append(keys, controls.Keymap.Up, controls.Keymap.Down)
	return keys
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
) (m Model) {
	optWidth := 0
	for _, opt := range options {
		if optWidth < len(opt[0]) {
			optWidth = len(opt[0])
		}
	}
	delegate := newSelectDelegate(optWidth + 1)
	w := 80
	h := 24
	l := list.New(toListItems(options), delegate, w, maxHeight(h))
	l.Help = help.New()
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)
	l.SetShowHelp(false)
	l.SetShowFilter(true)
	l.SetFilteringEnabled(true)
	l.Title = title // this gets hidden when filtering, so we have to re-render it
	l.InfiniteScrolling = true
	l.SetFilterText(value)
	l.SetFilterState(list.Filtering)
	input := &l.FilterInput
	input.Placeholder = "type to select"
	input.Prompt = " "            // to align with the list column
	input.ShowSuggestions = false // since the filtered list already provides suggestions
	return Model{list: l, value: value}
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
	selected := m.list.SelectedItem().(ListItem)[0]
	if m.value == selected {
		value = selected
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
	selectedTitle, selectedDesc, normalTitle, normalDesc lipgloss.Style
}

func newSelectDelegate(optWidth int) selectDelegate {
	return selectDelegate{
		height:   1,
		spacing:  0,
		optWidth: optWidth,
		styles: selectDelegateStyles{
			selectedTitle: lipgloss.NewStyle().Bold(true).Underline(true),
			selectedDesc:  lipgloss.NewStyle().Underline(true),
			normalTitle:   lipgloss.NewStyle().Faint(true),
			normalDesc:    lipgloss.NewStyle().Faint(true),
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
	desc = strings.Repeat(" ", d.optWidth-len(title)) + desc

	leftGutter := 3 // "   " or " > "
	leftColumn := leftGutter + 1
	rightColumn := m.Width() - leftColumn

	if isSelected {
		gutter := " > "
		styledTitle := d.styles.selectedTitle.Render(title)
		utils.Must(fmt.Fprint(w, gutter+styledTitle))
		wrappedDesc := wrapLine(uint(leftColumn), desc, rightColumn, d.styles.selectedDesc)
		utils.Must(fmt.Fprint(w, wrappedDesc))
	} else {
		gutter := "   "
		styledTitle := d.styles.normalTitle.Render(title)
		utils.Must(fmt.Fprint(w, gutter+styledTitle))
		wrappedDesc := wrapLine(uint(leftColumn), desc, rightColumn, d.styles.normalDesc)
		utils.Must(fmt.Fprint(w, wrappedDesc))
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		k := msg.Key()
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
		default:
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	case tea.MouseWheelMsg:
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		h := maxHeight(msg.Height)
		m.list.SetSize(msg.Width, h)
		m.list.Help.SetWidth(msg.Width)
		return m, cmd
	default:
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
}

func wrapLine(left uint, text string, right int, style lipgloss.Style) string {
	lines := strings.SplitN(wordwrap.String(text, right), "\n", 2)
	result := style.Render(lines[0])
	if len(lines) > 1 {
		result += "\n" + indent.String(style.Render(lines[1]), left)
	}
	return result
}

func (m Model) Render(s io.StringWriter) {
	utils.Must(s.WriteString(m.list.Title + "\n"))
	utils.Must(s.WriteString(m.list.View()))
}
