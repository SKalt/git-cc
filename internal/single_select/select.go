package single_select

import (
	"fmt"
	"io"
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
type ListItem struct {
	Title       string
	Description string
}

func (i ListItem) FilterValue() string { return i.Title }

// MakeItems creates list items from parallel slices of options and hints.
func MakeItems(options, hints []string) []list.Item {
	items := make([]list.Item, len(options))
	for i := range options {
		items[i] = ListItem{Title: options[i], Description: hints[i]}
	}
	return items
}

type Model struct {
	list      list.Model
	textInput textinput.Model
	context   string
	options   []string
	matchFunc func(string, string) bool
}

func (m Model) Init() tea.Cmd {
	return nil
}

func NewModel(
	context string,
	value string,
	options []string,
	hints []string,
	match func(string, string) bool,
) Model {
	switch len(options) {
	case 0:
		panic("empty options")
	case len(hints): // ok
	default:
		panic(fmt.Errorf("len(hints) %d != %d len(options)", len(hints), len(options)))
	}

	items := MakeItems(options, hints)
	optWidth := 0
	for _, opt := range options {
		if optWidth < len(opt) {
			optWidth = len(opt)
		}
	}
	delegate := newSelectDelegate(optWidth + 1)
	l := list.New(items, delegate, 80, 24)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)
	l.Filter = makeFilterFunc(match)
	l.InfiniteScrolling = true
	l.SetFilterText(value)

	input := textinput.New()
	input.Placeholder = "type to select"
	input.Prompt = "   "
	input.ShowSuggestions = true
	input.SetSuggestions(options)
	input.SetValue(value)
	input.SetCursor(len(value))
	input.Focus()

	return Model{
		list:      l,
		textInput: input,
		context:   context,
		options:   options,
		matchFunc: match,
	}
}

func (m *Model) Focus() tea.Cmd {
	return m.textInput.Focus()
}

func (m Model) SetErr(err error) Model {
	m.textInput.Err = err
	return m
}

func (m Model) Focused() bool {
	return m.textInput.Focused()
}

func (m Model) Blur() {
	m.textInput.Blur()
}

func (m Model) Options() []string {
	return m.options
}

// Value returns the selected item's title, or "" if nothing is selected.
func (m Model) Value() string {
	item := m.list.SelectedItem()
	if item == nil {
		return ""
	}
	return item.(ListItem).Title
}

func (m Model) CurrentInput() string {
	return m.textInput.Value()
}

// UpdateItems replaces the list items, updates the match function, and re-applies the filter.
func (m *Model) UpdateItems(options, hints []string, match func(string, string) bool) tea.Cmd {
	m.options = options
	m.matchFunc = match
	m.list.Filter = makeFilterFunc(match)
	cmd := m.list.SetItems(MakeItems(options, hints))
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
	title := i.Title
	desc := strings.Repeat(" ", d.optWidth-len(i.Title)) + i.Description

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

func makeFilterFunc(match func(string, string) bool) list.FilterFunc {
	return func(term string, targets []string) []list.Rank {
		if term == "" {
			ranks := make([]list.Rank, len(targets))
			for i := range targets {
				ranks[i] = list.Rank{Index: i}
			}
			return ranks
		}
		var ranks []list.Rank
		for i, target := range targets {
			if match(term, target) {
				ranks = append(ranks, list.Rank{
					Index:          i,
					MatchedIndexes: sequence(0, len(term)),
				})
			}
		}
		return ranks
	}
}

func sequence(start, end int) []int {
	if start >= end {
		return nil
	}
	idxs := make([]int, end-start)
	for i := range idxs {
		idxs[i] = start + i
	}
	return idxs
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
