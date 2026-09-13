package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/thebanri/limoni/uitest"
)

// The same flow an agent performs through limoni-mcp, as a test: no terminal,
// no build tag, no sleeps.
func TestReleaseFlow(t *testing.T) {
	app := newChecklist()
	page := uitest.Run(t, 80, 24, app.draw)
	releaseFlow(t, page)

	if screen := page.Screen(); strings.Contains(screen, "tok-4242") {
		t.Errorf("the token was drawn on screen:\n%s", screen)
	}
}

// TestLiveDemo runs the same flow against the application running in another
// terminal, slowly enough to watch:
//
//	# terminal 1
//	go run -tags limoni_debug ./examples/agent_checklist
//	# terminal 2
//	LIMONI_DEMO_SOCKET=$XDG_RUNTIME_DIR/limoni-checklist.sock \
//		go test -v -run TestLiveDemo ./examples/agent_checklist
//
// LIMONI_DEMO_SLOWMO sets the pause after each action (default 600ms). Restart
// the application before each run: it keeps the tasks from the last one.
func TestLiveDemo(t *testing.T) {
	socket := os.Getenv("LIMONI_DEMO_SOCKET")
	if socket == "" {
		t.Skip("set LIMONI_DEMO_SOCKET to drive a running agent_checklist")
	}
	slowMo := 600 * time.Millisecond
	if v := os.Getenv("LIMONI_DEMO_SLOWMO"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			t.Fatalf("LIMONI_DEMO_SLOWMO: %v", err)
		}
		slowMo = d
	}
	page := uitest.Connect(t, socket, uitest.WithSlowMo(slowMo))
	releaseFlow(t, page)
}

// releaseFlow is deliberately not a test helper, so each logged action and
// any failure points at its own line here.
func releaseFlow(t *testing.T, page *uitest.Page) {
	status := page.GetByID("status")
	rows := page.GetByRole("list-item", "").Within(page.GetByRole("list", "Tasks"))

	page.GetByRole("input", "New task").Type("Tag v1.0")
	page.GetByRole("button", "Add task").Click()
	page.Expect(status).ToHaveLabel(`Added "Tag v1.0".`)
	page.Expect(rows).ToHaveCount(3)

	// Deploying with open tasks is refused, and says which.
	page.GetByRole("button", "Deploy").Click()
	page.Expect(status).ToContainLabel("Blocked: not done")

	for i := 0; i < 3; i++ {
		rows.Nth(i).Click()
		page.Press("space")
		page.Expect(rows.Nth(i)).ToContainLabel("[x]")
	}

	page.GetByRole("button", "Deploy").Click()
	page.Expect(status).ToHaveLabel("Blocked: the deploy token is empty.")

	// The token goes in, and never comes back out.
	token := page.GetByRole("input", "Deploy token")
	token.Type("tok-4242")
	page.Expect(token).ToHaveValue("")
	page.GetByRole("button", "Deploy").Click()
	page.Expect(status).ToHaveLabel("Deployed with 3 tasks complete.")
}
