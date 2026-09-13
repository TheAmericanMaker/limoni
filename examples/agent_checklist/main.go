// agent_checklist is a small release checklist built to be driven by an AI
// agent through limoni-mcp — or by hand.
//
// An agent adds tasks, ticks them off, enters a deploy token and presses
// Deploy, all through the semantic tree. The token field is secret: the agent
// can type into it, but no tool will ever show it what was typed.
//
// Run it with the automation gateway compiled in:
//
//	go run -tags limoni_debug ./examples/agent_checklist
//
// and, from another terminal, point the bridge at its socket —
// $XDG_RUNTIME_DIR/limoni-checklist.sock, or $LIMONI_AUTOMATION_SOCKET if set:
//
//	claude mcp add limoni -- limoni-mcp -socket "$XDG_RUNTIME_DIR/limoni-checklist.sock"
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebanri/limoni"
	"github.com/thebanri/limoni/core/accessibility"
	"github.com/thebanri/limoni/core/buffer"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/core/driver"
	"github.com/thebanri/limoni/widgets"
)

type task struct {
	title string
	done  bool
}

// button is a clickable label with a semantic node. Embedding
// widgets.Accessible is all it takes for an agent to find it by role and label.
type button struct {
	widgets.Accessible
}

func (b button) Draw(ctx cell.Context, buf *buffer.Buffer) {
	buf.SetStringWithin(ctx.Area.X, ctx.Area.Y, "[ "+b.Label+" ]", ctx.Style, ctx.Area.Width)
}

func (b button) SizeHint(cell.Rect) (uint16, uint16) {
	return uint16(len(b.Label) + 4), 1
}

// status is a line of text that is also announced in the tree, so an agent can
// read the outcome of what it did.
type status struct {
	widgets.Accessible
	style cell.Style
}

func (s status) Draw(ctx cell.Context, buf *buffer.Buffer) {
	buf.SetStringWithin(ctx.Area.X, ctx.Area.Y, s.Label, s.style, ctx.Area.Width)
}

func (s status) SizeHint(cell.Rect) (uint16, uint16) { return uint16(len(s.Label)), 1 }

func socketPath() string {
	if path := os.Getenv("LIMONI_AUTOMATION_SOCKET"); path != "" {
		return path
	}
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = filepath.Join(os.TempDir(), fmt.Sprintf("limoni-%d", os.Getuid()))
	}
	return filepath.Join(dir, "limoni-checklist.sock")
}

func main() {
	tasks := []task{{title: "Run the test suite"}, {title: "Update the changelog"}}
	newTask := widgets.NewTextInputState()
	token := widgets.NewTextInputState()
	list := widgets.NewListState()
	list.Selected = 0
	message := "Tick every task and enter the deploy token, then press Deploy."
	deployed := false

	addTask := func() {
		title := strings.TrimSpace(newTask.Value())
		if title == "" {
			message = "Type a task name first."
			return
		}
		tasks = append(tasks, task{title: title})
		newTask.SetValue("")
		message = fmt.Sprintf("Added %q.", title)
	}
	toggle := func() {
		if list.Selected >= 0 && list.Selected < len(tasks) {
			tasks[list.Selected].done = !tasks[list.Selected].done
		}
	}
	deploy := func() {
		var open []string
		for _, t := range tasks {
			if !t.done {
				open = append(open, t.title)
			}
		}
		switch {
		case len(open) > 0:
			message = "Blocked: not done — " + strings.Join(open, ", ") + "."
		case token.Value() == "":
			message = "Blocked: the deploy token is empty."
		default:
			deployed = true
			message = fmt.Sprintf("Deployed with %d tasks complete.", len(tasks))
		}
	}

	socket := socketPath()
	err := limoni.Run(func(f *limoni.Frame, ev *limoni.Event) bool {
		focus := f.FocusManager
		if ev != nil && ev.Type == limoni.EventKey {
			switch {
			case ev.Key.Type == limoni.KeyEsc:
				return false
			case ev.Key.Type == limoni.KeyTab:
				focus.Next()
			default:
				switch focus.Focused() {
				case "new-task":
					if ev.Key.Type == limoni.KeyEnter {
						addTask()
					} else {
						newTask.HandleKey(ev.Key)
					}
				case "token":
					token.HandleKey(ev.Key)
				case "tasks":
					switch ev.Key.Type {
					case limoni.KeyUp:
						list.Previous()
					case limoni.KeyDown:
						list.Next()
					case limoni.KeyEnter, limoni.KeySpace:
						toggle()
					}
				}
			}
		}

		area := f.Area()
		accent := cell.Style{Fg: cell.NewColorRGB(0, 220, 170)}
		f.RenderWidget(widgets.Block{Title: " Release checklist ", Borders: widgets.BorderAll, BorderStyle: accent},
			cell.NewRect(0, 0, area.Width, area.Height))

		f.RenderWidget(&widgets.TextInput{ID: "new-task", State: newTask, Placeholder: "New task", Focused: focus.IsFocused("new-task")},
			cell.NewRect(2, 2, 36, 1))
		addArea := cell.NewRect(40, 2, 14, 1)
		f.RenderWidget(button{widgets.Accessible{ID: "add", Role: accessibility.RoleButton, Label: "Add task"}}, addArea)
		f.RegisterClickHandler(addArea, func(driver.MouseEvent) { addTask() })

		items := make([]string, len(tasks))
		for i, t := range tasks {
			mark := "[ ] "
			if t.done {
				mark = "[x] "
			}
			items[i] = mark + t.title
		}
		f.RenderWidget(&widgets.List{ID: "tasks", Items: items, State: list,
			SelectedStyle: cell.Style{Fg: cell.NewColorRGB(0, 0, 0), Bg: cell.NewColorRGB(0, 220, 170)}},
			cell.NewRect(2, 4, 52, uint16(len(items))))

		row := uint16(5 + len(items))
		f.RenderWidget(&widgets.TextInput{ID: "token", State: token, Placeholder: "Deploy token", Secret: true, Focused: focus.IsFocused("token")},
			cell.NewRect(2, row, 36, 1))
		deployArea := cell.NewRect(40, row, 12, 1)
		f.RenderWidget(button{widgets.Accessible{ID: "deploy", Role: accessibility.RoleButton, Label: "Deploy"}}, deployArea)
		f.RegisterClickHandler(deployArea, func(driver.MouseEvent) { deploy() })

		style := cell.Style{Fg: cell.NewColorRGB(230, 230, 230)}
		if deployed {
			style = accent
		} else if strings.HasPrefix(message, "Blocked") {
			style = cell.Style{Fg: cell.NewColorRGB(255, 120, 90)}
		}
		f.RenderWidget(status{widgets.Accessible{ID: "status", Label: message}, style}, cell.NewRect(2, row+2, area.Width-4, 1))
		f.RenderWidget(status{Accessible: widgets.Accessible{Label: "Tab: next field   Enter/Space: tick task   Esc: quit"}, style: cell.Style{Fg: cell.NewColorRGB(120, 120, 130)}},
			cell.NewRect(2, area.Height-2, area.Width-4, 1))
		return true
	}, limoni.WithAutomation(socket, limoni.AutomationPolicy{
		AllowInput:        true,
		ExposeInputValues: true, // the task name is fine to show; the token stays secret regardless
	}))

	if errors.Is(err, limoni.ErrAutomationNotCompiled) {
		fmt.Fprintln(os.Stderr, "This example is meant to be driven by an agent; build it with the gateway:")
		fmt.Fprintln(os.Stderr, "  go run -tags limoni_debug ./examples/agent_checklist")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
