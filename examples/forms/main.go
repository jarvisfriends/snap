// Command forms is a script-usable task form proving snap/forms extends huh
// rather than replacing it: a plain huh.Form whose fields validate through
// forms.HuhValidate(ParseRequired/ParseDuration/ParseISODate), with
// SplitAndClean cleaning the tags on submit. Completing the form writes one
// value per line to stdout (task, duration, due date, comma-joined tags; the
// TUI itself renders on stderr):
//
//	go run ./examples/forms | { read task; read dur; read due; read tags; }
//
// --no-help hides the status bar (huh's own help line is off — the bar shows
// the keys instead). Ctrl+C cancels: nothing printed, exit 1.
package main

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/forms"
	"github.com/jarvisfriends/snap/styles"
)

// newTaskForm builds the huh form: standard huh fields, snap/forms parsers as
// their validators, themed by the shared example palette.
func newTaskForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("task").
				Title("Task").
				Placeholder("what needs doing (required)").
				Validate(forms.HuhValidate(forms.ParseRequired, "task")),
			huh.NewInput().
				Key("duration").
				Title("Duration").
				Placeholder("5m, 1h, 7h30m").
				Validate(forms.HuhValidate(forms.ParseDuration, "duration")),
			huh.NewInput().
				Key("due").
				Title("Due date").
				Placeholder("YYYY-MM-DD").
				Validate(forms.HuhValidate(forms.ParseISODate, "due date")),
			huh.NewInput().
				Key("tags").
				Title("Tags").
				Placeholder("comma, separated , list,, of tags"),
		),
	).
		WithTheme(styles.HuhThemeFunc()).
		WithShowHelp(false) // the status bar below carries the keys
}

type demoApp struct {
	form   *huh.Form
	chrome *exui.Chrome
	height int
}

func newDemo() *demoApp {
	return &demoApp{
		form: newTaskForm(),
		chrome: exui.NewChrome(
			exui.Bind("tab/shift+tab", "field"),
			exui.Bind("enter", "next/submit"),
			exui.Bind("ctrl+c", "cancel"),
		),
	}
}

func (a *demoApp) Init() tea.Cmd { return a.form.Init() }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		a.height = msg.Height
		a.chrome.SetWidth(msg.Width)
	}
	model, cmd := a.form.Update(msg)
	if f, ok := model.(*huh.Form); ok {
		a.form = f
	}
	if a.form.State != huh.StateNormal {
		return a, tea.Quit
	}
	return a, cmd
}

func (a *demoApp) View() tea.View {
	v := tea.NewView(a.form.View())
	a.chrome.Apply(&v, a.height)
	v.AltScreen = true
	return v
}

func main() {
	exui.Init()
	final, err := exui.Program(newDemo()).Run()
	if err != nil {
		exui.Fatal(err)
	}
	a, ok := final.(*demoApp)
	if !ok || a.form.State != huh.StateCompleted {
		exui.Finish(false)
	}
	// The same parsers that validated the fields now produce the typed
	// values, so output can never disagree with what validation accepted.
	d, _ := forms.ParseDuration(a.form.GetString("duration"), "duration")
	due, _ := forms.ParseISODate(a.form.GetString("due"), "due date")
	exui.Finish(true,
		strings.TrimSpace(a.form.GetString("task")),
		d.String(),
		due.Format(time.DateOnly),
		strings.Join(forms.SplitAndClean(a.form.GetString("tags"), ","), ","),
	)
}
