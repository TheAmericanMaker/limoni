package main

import (
	"strings"
	"testing"

	"github.com/thebanri/limoni/uitest"
)

// The same flow an agent performs through limoni-mcp, as a test: no terminal,
// no build tag, no sleeps.
func TestReleaseFlow(t *testing.T) {
	app := newChecklist()
	page := uitest.Run(t, 80, 24, app.draw)
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

	page.GetByRole("input", "Deploy token").Type("tok-4242")
	page.Expect(page.GetByRole("input", "Deploy token")).ToHaveValue("")
	page.GetByRole("button", "Deploy").Click()
	page.Expect(status).ToHaveLabel("Deployed with 3 tasks complete.")

	if screen := page.Screen(); strings.Contains(screen, "tok-4242") {
		t.Errorf("the token was drawn on screen:\n%s", screen)
	}
}
