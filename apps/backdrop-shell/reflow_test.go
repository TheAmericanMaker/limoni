package main

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
)

// reflowSession is an emulated w×h screen with some output on it.
func reflowSession(w, h int, output string) *session {
	s := &session{w: w, h: h, emu: vt.NewEmulator(w, h), childModes: map[ansi.DECMode]bool{}}
	_, _ = s.emu.Write([]byte(output))
	s.drainHistory()
	return s
}

func (s *session) resizeTo(w, h int) {
	s.reflow(w, h)
	s.w, s.h = w, h
}

// all is the history and the screen as text, one line each, trimmed.
func (s *session) all() []string {
	var out []string
	for _, l := range s.hist {
		out = append(out, lineString(l))
	}
	for y := 0; y < s.h; y++ {
		var b strings.Builder
		for x := 0; x < s.w; x++ {
			if c := s.emu.CellAt(x, y); c != nil {
				b.WriteString(c.Content)
			}
		}
		out = append(out, strings.TrimRight(b.String(), " "))
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}

func lineString(l uv.Line) string {
	var b strings.Builder
	for i := range l {
		b.WriteString(l[i].Content)
	}
	return strings.TrimRight(b.String(), " ")
}

// Narrowing the window for a moment must not lose what was on the right,
// and widening it again must give the lines back as they were: that is how
// fastfetch's output was lost.
func TestNarrowingAndWideningKeepsTheText(t *testing.T) {
	logo := "   /\\     OS  CachyOS x86_64 with a long description of it\r\n" +
		"  /  \\    Kernel  Linux 7.2.8-1-cachyos and some more text\r\n$ "
	s := reflowSession(80, 10, logo)
	before := s.all()

	s.resizeTo(30, 10)
	narrow := strings.Join(s.all(), "")
	for _, word := range []string{"CachyOS", "description", "Kernel", "cachyos"} {
		if !strings.Contains(narrow, word) {
			t.Fatalf("narrowing lost %q:\n%s", word, strings.Join(s.all(), "\n"))
		}
	}

	s.resizeTo(80, 10)
	if got := s.all(); strings.Join(got, "\n") != strings.Join(before, "\n") {
		t.Fatalf("after widening again:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(before, "\n"))
	}
	if p := s.emu.CursorPosition(); p.Y != 2 || p.X != 2 {
		t.Fatalf("cursor at %+v, want after the prompt", p)
	}
}

// A shorter window keeps the bottom — the prompt — and moves the top rows
// into the history, where the emulator dropped the bottom rows.
func TestShorterWindowKeepsThePrompt(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 8; i++ {
		b.WriteString("line " + strings.Repeat("x", i) + "\r\n")
	}
	b.WriteString("$ ")
	s := reflowSession(40, 10, b.String())
	s.resizeTo(40, 4)
	// Eight lines and the prompt are nine rows: five go to the history.
	if len(s.hist) != 5 {
		t.Fatalf("history has %d lines, want 5", len(s.hist))
	}
	var screen []string
	for y := 0; y < 4; y++ {
		var row strings.Builder
		for x := 0; x < 40; x++ {
			row.WriteString(s.emu.CellAt(x, y).Content)
		}
		screen = append(screen, strings.TrimRight(row.String(), " "))
	}
	if screen[3] != "$" || screen[2] != "line xxxxxxxx" {
		t.Fatalf("screen %q", screen)
	}
	if p := s.emu.CursorPosition(); p.Y != 3 || p.X != 2 {
		t.Fatalf("cursor at %+v", p)
	}
}

// Lines split by a narrowing are joined again only while they still hold
// what they held: after a clear, the rows are new text.
func TestChangedLinesAreNotJoined(t *testing.T) {
	s := reflowSession(40, 6, strings.Repeat("a", 30)+"\r\n$ ")
	s.resizeTo(20, 6) // splits the a's in two
	_, _ = s.emu.Write([]byte("\x1b[H\x1b[2Jfirst\r\nsecond\r\n$ "))
	s.resizeTo(40, 6)
	got := s.all()
	if len(got) < 2 || got[len(got)-3] != "first" || got[len(got)-2] != "second" {
		t.Fatalf("new lines were joined: %q", got)
	}
}

func TestAWideCharacterIsNotCutInTwo(t *testing.T) {
	s := reflowSession(20, 4, "abcd🍋ef\r\n$ ")
	s.resizeTo(5, 6)
	// "abcd" fills four columns; the lemon is two wide and goes whole to
	// the next row.
	if c := s.emu.CellAt(0, 1); c == nil || c.Content != "🍋" || c.Width != 2 {
		t.Fatalf("row 2 starts with %+v", c)
	}
	s.resizeTo(20, 4)
	if c := s.emu.CellAt(4, 0); c == nil || c.Content != "🍋" {
		t.Fatalf("widening did not join the lemon back: %+v", c)
	}
}

// fastfetch's "OS  CachyOS" line, cut at 45 columns, ends its first half on
// a blank. The half kept in the history keeps that blank, the same line
// read from the screen does not, and the check that the halves are still
// the same took that for a change: the line stayed split.
func TestASplitThatEndsOnABlankJoinsBack(t *testing.T) {
	line := strings.Repeat(" ", 18) + "`ooo/" + strings.Repeat(" ", 19) + "OS  CachyOS x86_64"
	var b strings.Builder
	for i := 0; i < 12; i++ { // enough to push the split into the history
		b.WriteString(line + "\r\n\r\n")
	}
	b.WriteString("$ ")
	s := reflowSession(120, 10, b.String())
	before := strings.Join(s.all(), "\n")
	s.resizeTo(45, 10)
	if len(s.hist) == 0 {
		t.Fatal("test needs split lines in the history")
	}
	s.resizeTo(120, 10)
	if after := strings.Join(s.all(), "\n"); after != before {
		t.Fatalf("after narrowing and widening:\n%s\nwant:\n%s", after, before)
	}
}

// clear erases the saved lines with CSI 3 J; the history kept outside the
// emulator must go with them, or a resize brings old output back.
func TestClearErasesTheHistory(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 20; i++ {
		b.WriteString("old output\r\n")
	}
	s := reflowSession(40, 5, b.String())
	if len(s.hist) == 0 {
		t.Fatal("test needs history")
	}
	s.write1([]byte("\x1b[H\x1b[2J\x1b[3J$ "))
	s.resizeTo(30, 8)
	if got := strings.Join(s.all(), "|"); got != "$" {
		t.Fatalf("after clear and a resize: %q", got)
	}
}

// History that was history before the resize stays out of view; lines the
// resize itself pushed off come back when there is room — even after the
// shell has scrolled a line redrawing its prompt, as fish does on a resize.
func TestOnlyLinesAResizePushedOffComeBack(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 15; i++ { // five of them scroll into the history
		b.WriteString("old\r\n")
	}
	b.WriteString("\x1b[H\x1b[2J") // clear without erasing the saved lines
	for i := 0; i < 4; i++ {
		b.WriteString(strings.Repeat("n", 30) + "\r\n")
	}
	b.WriteString("$ ")
	s := reflowSession(40, 10, b.String())
	visible := func() string {
		var rows []string
		for y := 0; y < s.h; y++ {
			var r strings.Builder
			for x := 0; x < s.w; x++ {
				r.WriteString(s.emu.CellAt(x, y).Content)
			}
			rows = append(rows, strings.TrimRight(r.String(), " "))
		}
		return strings.TrimRight(strings.Join(rows, "|"), "|")
	}
	before := visible()
	if len(s.hist) == 0 {
		t.Fatal("test needs history from before the resize")
	}

	s.resizeTo(10, 10) // four 30-wide lines become twelve rows: some go up
	s.write1([]byte("\r\n$ "))
	s.resizeTo(40, 10)
	got := visible()
	if strings.Contains(got, "old") {
		t.Fatalf("history from before the resize came back into view: %q", got)
	}
	if strings.Count(got, strings.Repeat("n", 30)) != 4 {
		t.Fatalf("the lines the resize pushed off did not all come back: %q\nbefore: %q", got, before)
	}
}
