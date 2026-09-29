package main

import (
	"bytes"
	"strconv"

	"github.com/thebanri/limoni/core/cell"
)

// Keys and reports the shell never sees: the wrapper acts on them itself.
var (
	keyScrollUp    = []byte("\x1b[5;2~") // Shift+PageUp
	keyScrollDown  = []byte("\x1b[6;2~") // Shift+PageDown
	reportFocusIn  = []byte("\x1b[I")
	reportFocusOut = []byte("\x1b[O")
)

// inputKind is something from the terminal the wrapper handles.
type inputKind int

const (
	inputScrollUp inputKind = iota + 1
	inputScrollDown
	inputFocusIn
	inputFocusOut
	inputMouse
	inputColor // the terminal's answer to a colour query, late
	inputOther // anything else: typing ends a look at the scrollback
)

type inputEvent struct {
	kind  inputKind
	mouse mouseEvent // for inputMouse
	osc   string     // for inputColor: "10" or "11"
	color cell.Color // for inputColor
}

// mouseEvent is an SGR mouse report (mode 1006): the button byte as the
// terminal sent it, the cell (0-based), and whether it is a release.
type mouseEvent struct {
	button  int
	x, y    int
	release bool
}

func (m mouseEvent) wheel() bool  { return m.button&64 != 0 }
func (m mouseEvent) motion() bool { return m.button&32 != 0 }

// which is the button without its modifier and motion bits: 0 left,
// 1 middle, 2 right; for a wheel event 0 up and 1 down.
func (m mouseEvent) which() int { return m.button & 3 }

// splitInput walks one read from the terminal and separates what the wrapper
// handles from what goes to the shell, which it appends to pass, in order.
// Mouse reports are taken only when mouse is set — when the wrapper, not a
// program in the shell, has the mouse.
//
// Answers from the terminal (OSC replies, device attributes) never go to the
// shell: they answer the wrapper's own queries, which the shell did not ask,
// and a late one read as typing put stray letters on the command line — the
// hex digits of "rgb:dddd/…". One cut off at the end of the read is returned
// as carry, to be put in front of the next read.
//
// A key's bytes arrive in one read, so other sequences are only looked for
// whole within a read; one split across two reads goes to the shell as it
// came, which is what it would have got without the wrapper.
func splitInput(in []byte, pass []byte, mouse bool, emit func(inputEvent)) (out, carry []byte) {
	other := inputEvent{kind: inputOther}
	for len(in) > 0 {
		i := bytes.IndexByte(in, 0x1b)
		if i < 0 {
			emit(other)
			return append(pass, in...), nil
		}
		if i > 0 {
			emit(other)
			pass = append(pass, in[:i]...)
			in = in[i:]
		}
		if n, complete := terminalReply(in); n > 0 {
			if !complete {
				return pass, in
			}
			for _, num := range []string{"10", "11"} {
				if c, rest := takeColor(in[:n], num); len(rest) < n {
					emit(inputEvent{kind: inputColor, osc: num, color: c})
				}
			}
			in = in[n:]
			continue
		}
		if mouse {
			if ev, n := parseSGRMouse(in); n > 0 {
				emit(inputEvent{kind: inputMouse, mouse: ev})
				in = in[n:]
				continue
			}
		}
		switch {
		case bytes.HasPrefix(in, keyScrollUp):
			emit(inputEvent{kind: inputScrollUp})
			in = in[len(keyScrollUp):]
		case bytes.HasPrefix(in, keyScrollDown):
			emit(inputEvent{kind: inputScrollDown})
			in = in[len(keyScrollDown):]
		case bytes.HasPrefix(in, reportFocusIn):
			emit(inputEvent{kind: inputFocusIn})
			in = in[len(reportFocusIn):]
		case bytes.HasPrefix(in, reportFocusOut):
			emit(inputEvent{kind: inputFocusOut})
			in = in[len(reportFocusOut):]
		default:
			emit(other)
			pass = append(pass, in[0])
			in = in[1:]
		}
	}
	return pass, nil
}

// terminalReply measures a reply from the terminal at the start of in: an
// OSC string ("ESC ] digit … BEL or ST") or device attributes ("ESC [ ?
// digits and semicolons c"). It returns 0 if in does not start with one, and
// complete=false if in ends before the reply does. No key sends either: Alt+]
// is ESC ] alone, never followed by a digit.
func terminalReply(in []byte) (n int, complete bool) {
	switch {
	case len(in) >= 3 && in[1] == ']' && in[2] >= '0' && in[2] <= '9':
		for j := 3; j < len(in); j++ {
			if in[j] == 0x07 {
				return j + 1, true
			}
			if in[j] == 0x1b {
				switch {
				case j+1 >= len(in):
					return len(in), false
				case in[j+1] == '\\':
					return j + 2, true // ST
				default:
					return j, true // cut short by the next sequence, which is kept
				}
			}
		}
		return len(in), false
	case bytes.HasPrefix(in, []byte("\x1b[?")):
		for j := 3; j < len(in); j++ {
			c := in[j]
			if c == 'c' {
				return j + 1, true
			}
			if (c < '0' || c > '9') && c != ';' {
				return 0, true
			}
		}
		return len(in), false
	}
	return 0, true
}

// parseSGRMouse reads "ESC [ < b ; x ; y M" (press or motion) or "... m"
// (release) at the start of in, and returns its length, or 0.
func parseSGRMouse(in []byte) (mouseEvent, int) {
	if !bytes.HasPrefix(in, []byte("\x1b[<")) {
		return mouseEvent{}, 0
	}
	var nums [3]int
	k, start := 0, 3
	for j := 3; j < len(in) && j < 32; j++ {
		c := in[j]
		switch {
		case c >= '0' && c <= '9':
			continue
		case c == ';' || c == 'M' || c == 'm':
			if k > 2 || j == start {
				return mouseEvent{}, 0
			}
			v, err := strconv.Atoi(string(in[start:j]))
			if err != nil {
				return mouseEvent{}, 0
			}
			nums[k] = v
			k++
			start = j + 1
			if c != ';' {
				if k != 3 {
					return mouseEvent{}, 0
				}
				return mouseEvent{button: nums[0], x: nums[1] - 1, y: nums[2] - 1, release: c == 'm'}, j + 1
			}
		default:
			return mouseEvent{}, 0
		}
	}
	return mouseEvent{}, 0
}
