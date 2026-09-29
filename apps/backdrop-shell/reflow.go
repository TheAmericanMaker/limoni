package main

import (
	"strconv"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// Keeping the shell's text through a resize.
//
// The emulator cuts every line at the new width when the window narrows and
// drops the bottom rows when it gets shorter, so narrowing a window for a
// moment lost the right half of fastfetch's output for good. Terminals such
// as kitty and Alacritty reflow instead, and so does this: a narrower window
// wraps the long lines onto the next, a wider one joins them again, and a
// shorter one pushes the top rows into the history rather than losing the
// bottom ones, where the prompt is.
//
// The emulator does not record which lines were wrapped by the text running
// on rather than ended by a newline. The lines this code splits it records
// itself, with what they held, so a wider window joins exactly those — and
// not if they have changed since, after a clear, say. A line the program
// itself let wrap stays split when the window grows, as it does in xterm.
//
// The alternate screen is left alone: vim, btop and the like draw it again
// at the new size themselves.

// splitLine records a line reflow split: the width it was cut at and what
// it and the next line held, so a later reflow can tell whether they are
// still those two halves.
type splitLine struct {
	width int
	text  string
}

// reflow resizes the emulated screen to w×h, keeping its text.
func (s *session) reflow(w, h int) {
	if s.emu.IsAltScreen() || w <= 0 || h <= 0 {
		s.emu.Resize(w, h)
		s.splits = nil
		return
	}
	oldW := s.w
	history := len(s.hist)

	// Every line, history first, trimmed of trailing blanks. The history's
	// lines are trimmed already and are used as they are, and so are the
	// rows cut from them: copying ten thousand lines on every size a drag
	// passes through made a resize take 50 ms.
	lines := make([]uv.Line, 0, history+s.h)
	lines = append(lines, s.hist...)
	for y := 0; y < s.h; y++ {
		row := make(uv.Line, oldW)
		for x := 0; x < oldW; x++ {
			if c := s.emu.CellAt(x, y); c != nil {
				row[x] = *c
			} else {
				row[x] = uv.EmptyCell
			}
		}
		lines = append(lines, trimLine(row))
	}
	cur := s.emu.CursorPosition()
	curLine, curX := history+cur.Y, cur.X

	// Join the lines an earlier reflow split and that still hold what they
	// held then; note where the cursor falls in the result.
	type logical struct {
		cells  uv.Line
		cursor int // offset of the cursor in cells, or -1
		first  int // index in lines of its first part
	}
	var text []logical
	for i := 0; i < len(lines); i++ {
		l := logical{cursor: -1, first: i}
		for {
			if i == curLine {
				l.cursor = len(l.cells) + curX
			}
			piece := lines[i]
			base := len(l.cells)
			sp, ok := s.splits[i]
			join := ok && i+1 < len(lines) && sp.text == lineText(piece)+"\n"+lineText(lines[i+1])
			if base == 0 && !join {
				l.cells = piece // a line on its own is used as it is
				break
			}
			l.cells = append(l.cells[:base:base], piece...)
			if !join {
				break
			}
			// Pad this part back to the width it was cut at: its trailing
			// blanks were part of the line.
			for len(l.cells) < base+sp.width {
				l.cells = append(l.cells, uv.EmptyCell)
			}
			i++
		}
		text = append(text, l)
	}

	// Wrap each at the new width, without cutting a wide character in two.
	var out []uv.Line
	splits := map[int]splitLine{}
	cursorRow, cursorCol := 0, 0
	lastText := -1
	// History that was history before any reflow put lines above it stays
	// history: only lines a reflow pushed off the screen come back when
	// there is room for them again, as they do in kitty and Alacritty.
	earned := s.earned
	earnedRow := -1
	for _, l := range text {
		if earnedRow < 0 && l.first+len(lines)-len(lines) >= earned {
			earnedRow = len(out)
		}
		cells := l.cells
		if len(cells) == 0 {
			out = append(out, nil)
		}
		offset := 0
		for len(cells) > 0 {
			n := min(w, len(cells))
			if n < len(cells) && n > 0 && cells[n].Width == 0 && cells[n].Content == "" {
				n-- // the wide character goes whole to the next line
			}
			if n == 0 {
				n = min(w, len(cells))
			}
			if l.cursor >= offset && l.cursor < offset+n {
				cursorRow, cursorCol = len(out), l.cursor-offset
			}
			out = append(out, cells[:n:n])
			cells = cells[n:]
			offset += n
			if len(cells) > 0 {
				// The width actually cut at: one less than w when a
				// wide character went on to the next row.
				splits[len(out)-1] = splitLine{width: n}
			}
		}
		if l.cursor >= offset {
			// Past the end of the text: on its last row, or further down
			// if that is full.
			cursorRow, cursorCol = len(out)-1, l.cursor-(offset-len(out[len(out)-1]))
			for cursorCol >= w {
				cursorRow, cursorCol = cursorRow+1, cursorCol-w
				out = append(out, nil)
			}
		}
		if len(l.cells) > 0 {
			lastText = len(out) - 1
		}
	}

	// What shows: the rows up to the cursor or the last text, whichever is
	// lower, with the rows above the screen going into the history.
	end := max(cursorRow, lastText) + 1
	out = out[:min(end, len(out))]
	if earnedRow < 0 {
		earnedRow = len(out)
	}
	top := max(0, len(out)-h, min(earnedRow, len(out)))
	top = min(top, cursorRow)
	// The history holds only so many lines; the oldest go first.
	if drop := top - scrollbackLines; drop > 0 {
		out, top, cursorRow = out[drop:], top-drop, cursorRow-drop
		moved := map[int]splitLine{}
		for i, sp := range splits {
			if i >= drop {
				moved[i-drop] = sp
			}
		}
		splits = moved
	}

	s.emu.Resize(w, h)
	s.emu.ClearScrollback() // what the resize pushed off; hist has it all
	s.hist = append(s.hist[:0:0], out[:top]...)
	s.earned = min(earnedRow, top)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s.emu.SetCell(x, y, &uv.EmptyCell)
		}
		if top+y >= len(out) {
			continue
		}
		for x, c := range out[top+y] {
			if x >= w || c.Width == 0 && c.Content == "" {
				continue // a wide character's second column is written with it
			}
			s.emu.SetCell(x, y, &c)
		}
	}
	_, _ = s.emu.Write([]byte("\x1b[" + strconv.Itoa(cursorRow-top+1) + ";" + strconv.Itoa(cursorCol+1) + "H"))

	// Remember the splits by their line number in the history as it is now:
	// the scrollback keeps what it was given, and new output only adds lines
	// after these.
	s.splits = map[int]splitLine{}
	for i, sp := range splits {
		if i+1 < len(out) {
			sp.text = lineText(out[i]) + "\n" + lineText(out[i+1])
			s.splits[i] = sp
		}
	}
}

// trimLine is a copy of l without trailing blank cells. A wide character
// at the end keeps the empty cell after it, which the scrollback drops.
func trimLine(l uv.Line) uv.Line {
	n := len(l)
	for n > 0 && (l[n-1].IsZero() || l[n-1].Equal(&uv.EmptyCell)) {
		n--
	}
	out := make(uv.Line, n, n+1)
	copy(out, l[:n])
	if n > 0 {
		for k := 1; k < out[n-1].Width; k++ {
			out = append(out, uv.Cell{})
		}
	}
	return out
}

// lineText is the text of a line, for telling whether it changed. Trailing
// blanks do not count: the rows cut from a line keep theirs, the same line
// read back from the screen does not.
func lineText(l uv.Line) string {
	n := len(l)
	for n > 0 && (l[n-1].IsZero() || l[n-1].Equal(&uv.EmptyCell)) {
		n--
	}
	l = l[:n]
	var b strings.Builder
	for i := range l {
		b.WriteString(l[i].Content)
		b.WriteByte(0)
	}
	return b.String()
}
