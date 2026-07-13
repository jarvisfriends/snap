// Command forms is a script-usable task form built on snap/forms' parsers:
// every keystroke re-validates the focused field with field-naming errors
// inline (ParseRequired, ParseDuration, ParseISODate, SplitAndClean), and
// Ctrl+S submits once everything parses — writing one value per line to
// stdout (task, duration, due date, comma-joined tags; the TUI itself
// renders on stderr):
//
//	go run ./examples/forms | { read task; read dur; read due; read tags; }
//
// --no-help hides the status bar. Esc cancels: nothing printed, exit 1.
package main

import (
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/forms"
)

// fieldCount and the indexes below name the form's inputs.
const (
	fieldName = iota
	fieldDuration
	fieldDate
	fieldTags
	fieldCount
)

var (
	labels       = [fieldCount]string{"Task", "Duration", "Due date", "Tags"}
	placeholders = [fieldCount]string{
		"what needs doing (required)",
		"5m, 1h, 7h30m",
		"YYYY-MM-DD",
		"comma, separated , list,, of tags",
	}
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	labelStyle = lipgloss.NewStyle().Bold(true).Width(10)
)

type demoApp struct {
	inputs    [fieldCount]textinput.Model
	focused   int
	submitted bool
	chrome    *exui.Chrome
	height    int
}

func newDemo() *demoApp {
	a := &demoApp{chrome: exui.NewChrome(
		exui.Bind("tab/shift+tab", "field"),
		exui.Bind("ctrl+s", "submit"),
		exui.Bind("esc", "cancel"),
	)}
	for i := range a.inputs {
		in := textinput.New()
		in.Placeholder = placeholders[i]
		in.SetWidth(44)
		a.inputs[i] = in
	}
	a.inputs[fieldName].Focus()
	return a
}

func (a *demoApp) Init() tea.Cmd { return textinput.Blink }

// parsed returns field i's parsed value, or an error string when it doesn't
// parse. The same call drives the inline status line and the final output.
func (a *demoApp) parsed(i int) (string, error) {
	raw := a.inputs[i].Value()
	switch i {
	case fieldName:
		return forms.ParseRequired(raw, "task")
	case fieldDuration:
		d, err := forms.ParseDuration(raw, "duration")
		if err != nil {
			return "", err
		}
		return d.String(), nil
	case fieldDate:
		t, err := forms.ParseISODate(raw, "due date")
		if err != nil {
			return "", err
		}
		return t.Format(time.DateOnly), nil
	default:
		return strings.Join(forms.SplitAndClean(raw, ","), ","), nil
	}
}

// status renders the live validation line for field i.
func (a *demoApp) status(i int) string {
	val, err := a.parsed(i)
	if err != nil {
		return errStyle.Render("✗ " + err.Error())
	}
	if i == fieldTags {
		tags := forms.SplitAndClean(a.inputs[i].Value(), ",")
		if len(tags) == 0 {
			return dimStyle.Render("– no tags")
		}
		return okStyle.Render("✓ " + strconv.Itoa(len(tags)) + " tags: " + val)
	}
	return okStyle.Render("✓ " + val)
}

// complete reports whether every field parses.
func (a *demoApp) complete() bool {
	for i := range a.inputs {
		if _, err := a.parsed(i); err != nil {
			return false
		}
	}
	return true
}

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.height = msg.Height
		a.chrome.SetWidth(msg.Width)
		return a, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return a, tea.Quit
		case "ctrl+s":
			if a.complete() {
				a.submitted = true
				return a, tea.Quit
			}
			return a, nil
		case "tab", "enter", "down":
			a.moveFocus(1)
			return a, nil
		case "shift+tab", "up":
			a.moveFocus(-1)
			return a, nil
		}
	}
	var cmd tea.Cmd
	a.inputs[a.focused], cmd = a.inputs[a.focused].Update(msg)
	return a, cmd
}

func (a *demoApp) moveFocus(d int) {
	a.inputs[a.focused].Blur()
	a.focused = (a.focused + d + fieldCount) % fieldCount
	a.inputs[a.focused].Focus()
}

func (a *demoApp) View() tea.View {
	rows := make([]string, 0, fieldCount*3)
	for i := range a.inputs {
		rows = append(rows,
			lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render(labels[i]), a.inputs[i].View()),
			lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render(""), a.status(i)),
			"",
		)
	}
	v := tea.NewView(a.chrome.Attach(lipgloss.JoinVertical(lipgloss.Left, rows...), a.height))
	v.AltScreen = true
	return v
}

func main() {
	exui.Init()
	final, err := exui.Program(newDemo()).Run()
	if err != nil {
		exui.Fatal(err)
	}
	if a, ok := final.(*demoApp); ok && a.submitted {
		values := make([]string, 0, fieldCount)
		for i := range a.inputs {
			v, _ := a.parsed(i)
			values = append(values, v)
		}
		exui.Finish(true, values...)
	}
	exui.Finish(false)
}
