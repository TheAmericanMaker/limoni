package main

import (
	"strings"
	"testing"
)

func TestRepIsWrittenOut(t *testing.T) {
	cases := []struct{ in, want string }{
		{"─\x1b[4b|", "─────|"},
		{"▀\x1b[b", "▀▀"},
		{"ab\x1b[3b", "abbbb"},
		{"é\x1b[2b", "ééé"},                           // a combining accent repeats with its letter
		{"\x1b[31m─\x1b[0m\x1b[2b", "\x1b[31m─\x1b[0m──"}, // colours in between are kept
		{"x\x1b[?2b", "x\x1b[?2b"},                        // not REP: left alone
		{"x\x1b]0;title b\x07\x1b[1b", "x\x1b]0;title b\x07x"},
		{"\x1b[5b", ""},                                 // nothing to repeat
		{"\x1b]2;Mир\x07x\x1b[2b", "\x1b]2;Mир\x07xxx"}, // и is D0 B8: a string in UTF-8 goes on
		{"\x1b]2;Мир\x07x\x1b[2b", "xxx"},               // М is D0 9C: 9C is not the end, and the title is kept out
	}
	for _, tc := range cases {
		var f repFilter
		if got := string(f.filter([]byte(tc.in))); got != tc.want {
			t.Errorf("%q → %q, want %q", tc.in, got, tc.want)
		}
	}
}

// A sequence or a character cut between two reads is put together first.
func TestRepAcrossReads(t *testing.T) {
	in := "a─\x1b[3b\x1b]2;t\x07z"
	want := "a────\x1b]2;t\x07z"
	for cut := 1; cut < len(in); cut++ {
		var f repFilter
		got := string(f.filter([]byte(in[:cut]))) + string(f.filter([]byte(in[cut:])))
		if got != want {
			t.Fatalf("cut at %d: %q, want %q", cut, got, want)
		}
	}
}

// Through the emulator: a line of box drawing drawn with REP stays box
// drawing. Without the filter it came out as the last ASCII letter.
func TestRepThroughTheEmulator(t *testing.T) {
	s := reflowSession(20, 3, "")
	s.write1([]byte("c╭─\x1b[6b╮"))
	var row strings.Builder
	for x := 0; x < 10; x++ {
		row.WriteString(s.emu.CellAt(x, 0).Content)
	}
	if got := row.String(); got != "c╭───────╮" {
		t.Fatalf("drawn as %q", got)
	}
}

// Claude Code titles its window "✳ <session name>", and ✳ is E2 9C B3. The
// emulator's parser ends a string at 0x9c, the 8-bit ST, even inside a UTF-8
// character, and printed the rest of the title at the cursor: in the
// middle of the prompt being typed. Such a title goes past the emulator to
// the terminal, whole.
func TestTitleWith9CStaysOffTheScreen(t *testing.T) {
	for _, seq := range []string{
		"\x1b]0;✳ Backdrop-shell animasyon evren dairesi\x07",
		"\x1b]2;✻ working\x1b\\",
		"\x1b]2;Мир\x07", // М is D0 9C
	} {
		s := reflowSession(60, 3, "")
		var titles []string
		s.rep.title = func(t string) { titles = append(titles, t) }
		s.write1([]byte("> " + seq + "x"))
		if got := s.all()[0]; got != "> x" {
			t.Errorf("%q drew %q, want %q", seq, got, "> x")
		}
		want := seq[4 : len(seq)-1]
		if strings.HasSuffix(seq, "\x1b\\") {
			want = seq[4 : len(seq)-2]
		}
		if len(titles) != 1 || titles[0] != want {
			t.Errorf("%q set the titles %q, want %q", seq, titles, want)
		}
	}
}
