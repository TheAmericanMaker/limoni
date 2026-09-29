package main

import "bytes"

// Keys and reports the shell never sees: the wrapper acts on them itself.
var (
	keyScrollUp    = []byte("\x1b[5;2~") // Shift+PageUp
	keyScrollDown  = []byte("\x1b[6;2~") // Shift+PageDown
	reportFocusIn  = []byte("\x1b[I")
	reportFocusOut = []byte("\x1b[O")
)

// inputEvent is something from the keyboard the wrapper handles.
type inputEvent int

const (
	inputScrollUp inputEvent = iota + 1
	inputScrollDown
	inputFocusIn
	inputFocusOut
	inputOther // anything else: typing ends a look at the scrollback
)

// splitInput walks one read from the terminal and separates what the wrapper
// handles from what goes to the shell, which it appends to pass, in order.
//
// The terminal delivers a key's bytes in one read, so a sequence is only
// looked for whole within a read; one split across two reads goes to the
// shell as it came, which is what it would have got without the wrapper.
func splitInput(in []byte, pass []byte, emit func(inputEvent)) []byte {
	for len(in) > 0 {
		i := bytes.IndexByte(in, 0x1b)
		if i < 0 {
			emit(inputOther)
			return append(pass, in...)
		}
		if i > 0 {
			emit(inputOther)
			pass = append(pass, in[:i]...)
			in = in[i:]
		}
		switch {
		case bytes.HasPrefix(in, keyScrollUp):
			emit(inputScrollUp)
			in = in[len(keyScrollUp):]
		case bytes.HasPrefix(in, keyScrollDown):
			emit(inputScrollDown)
			in = in[len(keyScrollDown):]
		case bytes.HasPrefix(in, reportFocusIn):
			emit(inputFocusIn)
			in = in[len(reportFocusIn):]
		case bytes.HasPrefix(in, reportFocusOut):
			emit(inputFocusOut)
			in = in[len(reportFocusOut):]
		default:
			emit(inputOther)
			pass = append(pass, in[0])
			in = in[1:]
		}
	}
	return pass
}
