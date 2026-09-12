package limoni

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/thebanri/limoni/automation"
	"github.com/thebanri/limoni/core/accessibility"
	"github.com/thebanri/limoni/core/buffer"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/core/driver"
	"github.com/thebanri/limoni/core/terminal"
	"github.com/thebanri/limoni/widgets"
)

// This drives the real Run loop over a real socket. Building the automation
// package proves nothing about whether an application can actually be steered
// through it: the wiring between the loop, the frame and the server is where
// this would break.
func TestAutomationDrivesARunningApplication(t *testing.T) {
	socket := shortSocketPath(t)

	var mu sync.Mutex
	var stop atomic.Bool
	typed := ""
	items := []string{"alpha", "beta", "gamma"}
	listState := widgets.NewListState()
	listState.Selected = 0

	backend := driver.NewPortableBackend(driver.NewMemoryTerminalIO(nil, 60, 20))
	if err := backend.Setup(); err != nil {
		t.Fatalf("backend setup: %v", err)
	}
	term, err := terminal.New(backend)
	if err != nil {
		t.Fatalf("terminal: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- runLoop(term, func(f *Frame, ev *Event) bool {
			if stop.Load() {
				return false
			}
			mu.Lock()
			defer mu.Unlock()

			if ev != nil && ev.Type == EventKey && ev.Key.Type == KeyRune {
				typed += string(ev.Key.Ch)
			}

			f.RenderWidget(&widgets.List{
				ID:    "menu",
				Items: items,
				State: listState,
			}, NewRect(0, 0, 30, 5))

			f.RenderWidget(submitButton{Accessible: widgets.Accessible{
				ID:    "submit",
				Role:  accessibility.RoleButton,
				Label: "Submit",
			}}, NewRect(40, 10, 10, 3))

			return true
		}, appConfig{automationPath: socket, catchCtrlC: true})
	}()
	t.Cleanup(func() {
		// runLoop exits when the application returns false, so ask it to.
		stop.Store(true)
		Wakeup()
		<-done
		term.Close()
	})

	client, err := dialWithRetry(t, socket)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	version, w, h, err := client.Hello()
	if err != nil {
		t.Fatalf("hello: %v", err)
	}
	if version != automation.Version {
		t.Errorf("version = %d, want %d", version, automation.Version)
	}
	if w != 60 || h != 20 {
		t.Errorf("viewport = %dx%d, want 60x20", w, h)
	}

	// The button is addressable by what it means, never by where it is.
	node, err := client.WaitFor(automation.Selector{Role: "button", Label: "Submit"}, 3*time.Second)
	if err != nil {
		t.Fatalf("waiting for the button: %v", err)
	}
	if node.ID != "submit" {
		t.Errorf("resolved %q, want submit", node.ID)
	}
	if node.Bounds.X != 40 || node.Bounds.Y != 10 {
		t.Errorf("bounds = %v, want the rendered position", node.Bounds)
	}

	// The list reports its selection through the same tree.
	list, err := client.WaitFor(automation.Selector{Role: "list"}, 3*time.Second)
	if err != nil {
		t.Fatalf("waiting for the list: %v", err)
	}
	if list.Value != "alpha" || list.Position != 1 || list.SetSize != 3 {
		t.Errorf("list = %q %d/%d, want alpha 1/3", list.Value, list.Position, list.SetSize)
	}

	// Typing reaches the application's event handler.
	if err := client.Type("hey"); err != nil {
		t.Fatalf("type: %v", err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		mu.Lock()
		got := typed
		mu.Unlock()
		if got == "hey" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("application received %q, want \"hey\"", got)
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Selecting a different item changes what the tree reports, without the
	// test ever naming a coordinate.
	mu.Lock()
	listState.Selected = 2
	mu.Unlock()
	if err := client.Key("down"); err != nil {
		t.Fatalf("key: %v", err)
	}
	moved, err := client.WaitFor(automation.Selector{Role: "list", Value: "gamma"}, 3*time.Second)
	if err != nil {
		t.Fatalf("waiting for the new selection: %v", err)
	}
	if moved.Position != 3 {
		t.Errorf("position = %d, want 3", moved.Position)
	}

	// The rendered grid is still available for assertions the tree cannot make.
	screen, err := client.Screen()
	if err != nil {
		t.Fatalf("screen: %v", err)
	}
	if screen == "" {
		t.Error("screen snapshot is empty")
	}

}

// submitButton is how an application gives a custom widget a semantic node:
// embed widgets.Accessible and the node comes for free.
type submitButton struct {
	widgets.Accessible
}

func (b submitButton) Draw(ctx cell.Context, buf *buffer.Buffer) {
	buf.SetStringWithin(ctx.Area.X, ctx.Area.Y, "[ Submit ]", ctx.Style, ctx.Area.Width)
}

func (b submitButton) SizeHint(maxArea cell.Rect) (uint16, uint16) {
	return 10, 1
}

// shortSocketPath keeps the address under the sockaddr_un limit. t.TempDir()
// embeds the test name and a long random suffix, which on macOS pushes the
// path past 104 bytes and makes bind fail with a bare EINVAL.
func shortSocketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "lmn")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "a.sock")
}

// dialWithRetry waits for the loop's goroutine to have created the socket.
func dialWithRetry(t *testing.T, socket string) (*automation.Client, error) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		client, err := automation.Dial(socket)
		if err == nil {
			return client, nil
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(5 * time.Millisecond)
	}
}
