package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/thebanri/limoni"
	"github.com/thebanri/limoni/core/buffer"
	"github.com/thebanri/limoni/core/cell"
)

func TestLeaderboardStaysVisibleAndScrolls(t *testing.T) {
	g := newGame(1)
	var entries []scoreEntry
	for i := 0; i < 25; i++ {
		entries = append(entries, scoreEntry{Name: fmt.Sprintf("Player%02d", i+1), Score: 1000 - i})
	}
	g.takeAnswer(remoteAnswer{board: entries})
	if g.nWorld != 20 {
		t.Fatalf("kept %d scores, want 20", g.nWorld)
	}
	b := buffer.NewBuffer(cell.Rect{Width: 80, Height: 24})
	for _, ph := range []phase{phName, phTitle, phPlay, phOver} {
		g.phase, g.overAt, g.boardScroll = ph, -5, 0
		g.render(b)
		if s := screen(b); !strings.Contains(s, "WORLD'S BEST") || !strings.Contains(s, "Player01") || strings.Contains(s, "Player20") {
			t.Fatalf("phase %d: unexpected first page:\n%s", ph, s)
		}
		if g.boardView.x+g.boardView.w >= g.bx {
			t.Fatal("leaderboard overlaps the playing field")
		}
		g.mouse(limoni.MouseEvent{Button: limoni.MouseScrollDown, X: uint16(g.boardView.x + 2), Y: uint16(g.boardView.y + 5)})
		if g.boardScroll != 1 {
			t.Fatal("wheel did not scroll")
		}
		g.key(limoni.KeyEvent{Type: limoni.KeyPageDown})
		g.render(b)
		if s := screen(b); !strings.Contains(s, "Player20") || strings.Contains(s, "Player01") {
			t.Fatalf("last page:\n%s", s)
		}
		g.scrollBoard(100)
		if g.boardScroll != 10 {
			t.Fatalf("scroll escaped bottom: %d", g.boardScroll)
		}
		g.key(limoni.KeyEvent{Type: limoni.KeyPageUp})
		g.scrollBoard(-100)
		if g.boardScroll != 0 {
			t.Fatal("scroll escaped top")
		}
	}
	// A smaller replacement board must clamp the old scroll position.
	g.boardScroll = 10
	g.takeAnswer(remoteAnswer{board: entries[:1]})
	g.render(b)
	if g.boardScroll != 0 || !strings.Contains(screen(b), "Player01") {
		t.Fatal("replacement board did not reset the viewport")
	}
}

func TestNarrowLeaderboardCanReachLastScore(t *testing.T) {
	g := newGame(1)
	g.worldState, g.nWorld = worldOK, boardSize
	for i := range g.world {
		g.world[i] = scoreEntry{Name: fmt.Sprintf("P%02d", i+1), Score: 100 - i}
	}
	b := buffer.NewBuffer(cell.Rect{Width: 60, Height: 24})
	g.render(b)
	g.scrollBoard(100)
	g.render(b)
	if s := screen(b); !strings.Contains(s, "P20") {
		t.Fatalf("last score inaccessible in narrow window:\n%s", s)
	}
	// Shrinking further invalidates the old mouse target.
	g.render(buffer.NewBuffer(cell.Rect{Width: 40, Height: 16}))
	if g.boardView.w != 0 {
		t.Fatal("stale mouse target after shrinking")
	}
}

func TestPausedLeaderboardReceivesAnswers(t *testing.T) {
	g := playing()
	g.paused = true
	g.remote = &remote{answers: make(chan remoteAnswer, 1)}
	g.remote.answers <- remoteAnswer{board: []scoreEntry{{Name: "Ada", Score: 100}}}
	before := g.now
	g.update(dt)
	if g.worldState != worldOK || g.nWorld != 1 || g.now != before {
		t.Fatal("paused game did not update the board independently of simulation")
	}
}
