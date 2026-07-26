package single_select

import (
	"fmt"
	"io"
	"iter"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/skalt/git-cc/internal/utils"
)

// ListItem implements list.DefaultItem with a title and a description (hint).
type ListItem [2]string

func (i ListItem) FilterValue() string { return i[0] }

type Model struct {
	list      list.Model
	textInput textinput.Model
	context   string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func toListItems(input []ListItem) []list.Item {
	items := make([]list.Item, 0, len(input))
	for _, i := range input {
		items = append(items, i)
	}
	return items
}

func NewModel(
	context string,
	value string,
	options []ListItem,
) Model {
	optWidth := 0
	for _, opt := range options {
		if optWidth < len(opt[0]) {
			optWidth = len(opt[0])
		}
	}
	delegate := newSelectDelegate(optWidth + 1)
	l := list.New(toListItems(options), delegate, 80, 24)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.InfiniteScrolling = true
	l.SetFilterText(value)

	input := textinput.New()
	input.Placeholder = "type to select"
	input.Prompt = "   "
	input.ShowSuggestions = true
	suggestions := make([]string, 0, len(options))
	for _, o := range options {
		suggestions = append(suggestions, o[0])
	}
	input.SetSuggestions(suggestions)
	input.SetValue(value)
	input.SetCursor(len(value))
	input.Focus()

	return Model{
		list:      l,
		textInput: input,
		context:   context,
	}
}

func (m *Model) Focus() tea.Cmd { return m.textInput.Focus() }

func (m Model) SetErr(err error) Model {
	m.textInput.Err = err
	return m
}

func (m Model) Focused() bool { return m.textInput.Focused() }
func (m Model) Blur()         { m.textInput.Blur() }
func (m Model) Options() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, o := range m.list.Items() {
			if t := o.(ListItem)[0]; t != "" && !yield(t) {
				return
			}
		}
	}
}

// Value returns the selected item's title, or "" if nothing is selected.
func (m Model) Value() string {
	item := m.list.SelectedItem()
	if item == nil {
		return ""
	}
	return item.(ListItem)[0]
}

func (m Model) CurrentInput() string {
	return m.textInput.Value()
}

// UpdateItems replaces the list items, updates the match function, and re-applies the filter.
func (m *Model) UpdateItems(options []ListItem) tea.Cmd {
	cmd := m.list.SetItems(toListItems(options))
	m.list.SetFilterText(m.textInput.Value())
	return cmd
}

// selectDelegate renders list items with hints.
type selectDelegate struct {
	height   int
	spacing  int
	optWidth int
	styles   selectDelegateStyles
}

type selectDelegateStyles struct {
	selectedTitle lipgloss.Style
	selectedDesc  lipgloss.Style
	normalTitle   lipgloss.Style
	normalDesc    lipgloss.Style
}

func newSelectDelegate(optWidth int) selectDelegate {
	return selectDelegate{
		height:   2,
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
	i := item.(ListItem)

	if m.Width() <= 0 {
		return
	}

	isSelected := index == m.Index()
	title := i[0]
	desc := strings.Repeat(" ", d.optWidth-len(title)) + i[1]

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

func MatchStart(query, option string) bool {
	return len(query) <= len(option) && option[:len(query)] == query
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up", "ctrl+p":
			m.list.CursorUp()
			return m, nil
		case "down", "ctrl+n":
			m.list.CursorDown()
			return m, nil
		default:
			oldValue := m.textInput.Value()
			m.textInput, cmd = m.textInput.Update(msg)
			if m.textInput.Value() != oldValue {
				m.list.SetFilterText(m.textInput.Value())
			}
			return m, cmd
		}
	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.list.CursorUp()
			return m, nil
		case tea.MouseWheelDown:
			m.list.CursorDown()
			return m, nil
		default:
			oldValue := m.textInput.Value()
			m.textInput, cmd = m.textInput.Update(msg)
			if m.textInput.Value() != oldValue {
				m.list.SetFilterText(m.textInput.Value())
			}
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	default:
		oldValue := m.textInput.Value()
		m.textInput, cmd = m.textInput.Update(msg)
		if m.textInput.Value() != oldValue {
			m.list.SetFilterText(m.textInput.Value())
		}
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
	utils.Must(s.WriteString(m.context))
	utils.Must(s.WriteString("\n"))
	utils.Must(s.WriteString(m.textInput.View()))
	utils.Must(s.WriteString("\n"))
	utils.Must(s.WriteString(m.list.View()))
	utils.Must(s.WriteString("\n"))
	// utils.Must(s.WriteString(fmt.Sprintf("%03d x %03d\n", m.list.Width(), m.list.Height())))
	// utils.Must(s.WriteString("\n"))
}
