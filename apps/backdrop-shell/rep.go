package main

import (
	"bytes"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Repeating the previous character.
//
// REP (CSI n b) prints the previous graphic character n more times. kitty's,
// Alacritty's and xterm's terminfo all offer it, so ncurses programs use it
// for runs of one character. The emulator repeats the last character only if
// that was a single byte: after "─" or "▀" it repeated whatever ASCII letter
// came before, and a line of box drawing turned into a line of "c". So REP
// never reaches the emulator: repFilter writes the character out n times in
// its place, on the way in.

// repFilter rewrites REP in a stream of terminal output. It keeps its state
// across writes, since a sequence can be cut between two reads.
type repFilter struct {
	state   int    // where in a sequence the stream is
	seq     []byte // the sequence so far
	last    []byte // the last graphic character, with any combining marks
	partial []byte // a UTF-8 character cut at the end of the last write
}

const (
	repGround = iota
	repEsc    // after ESC
	repCSI    // in a control sequence
	repString // in an OSC, DCS, APC, PM or SOS string
)

// maxRepeat bounds a REP count, as terminals do.
const maxRepeat = 65535

// maxString bounds how much of an unterminated string is held back.
const maxString = 8 << 20

// filter returns in with every REP replaced by the character it repeats.
func (f *repFilter) filter(in []byte) []byte {
	if len(f.partial) > 0 {
		in = append(f.partial, in...)
		f.partial = nil
	}
	out := make([]byte, 0, len(in))
	for i := 0; i < len(in); {
		c := in[i]
		switch f.state {
		case repGround:
			switch {
			case c == 0x1b:
				f.state, f.seq = repEsc, append(f.seq[:0], c)
				i++
			case c < 0x20 || c == 0x7f:
				out = append(out, c)
				i++
			default:
				if !utf8.FullRune(in[i:]) {
					f.partial = append([]byte(nil), in[i:]...)
					return out
				}
				r, n := utf8.DecodeRune(in[i:])
				if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || r == 0x200d || unicode.Is(unicode.Variation_Selector, r) {
					f.last = append(f.last, in[i:i+n]...) // part of the character before
				} else {
					f.last = append(f.last[:0], in[i:i+n]...)
				}
				out = append(out, in[i:i+n]...)
				i += n
			}
		case repEsc:
			f.seq = append(f.seq, c)
			i++
			switch c {
			case '[':
				f.state = repCSI
			case ']', 'P', '_', '^', 'X':
				f.state = repString
			default:
				out = append(out, f.seq...)
				f.state = repGround
			}
		case repCSI:
			f.seq = append(f.seq, c)
			i++
			if c >= 0x40 && c <= 0x7e {
				if c == 'b' {
					if n, ok := repCount(f.seq[2 : len(f.seq)-1]); ok {
						for k := 0; k < n && len(f.last) > 0; k++ {
							out = append(out, f.last...)
						}
						f.state = repGround
						continue
					}
				}
				out = append(out, f.seq...)
				f.state = repGround
			}
		case repString:
			// Ended by BEL or ESC \ only: the 8-bit ST, 0x9c, is also a
			// UTF-8 continuation byte, and a title in Cyrillic has it.
			f.seq = append(f.seq, c)
			i++
			end := len(f.seq)
			if c == 0x07 || end >= 2 && f.seq[end-2] == 0x1b && c == '\\' || end > maxString {
				// Terminated — or too long to be anything but noise, which
				// goes on as it is rather than being held forever.
				out = append(out, f.seq...)
				f.state = repGround
			}
		}
	}
	// A sequence cut at the end waits for the rest; the emulator would only
	// have waited for it too.
	return out
}

// repCount reads REP's parameter: digits only, 1 when empty or zero.
func repCount(p []byte) (int, bool) {
	if len(p) == 0 {
		return 1, true
	}
	if bytes.ContainsFunc(p, func(r rune) bool { return r < '0' || r > '9' }) {
		return 0, false
	}
	n, err := strconv.Atoi(string(p))
	if err != nil || n > maxRepeat {
		n = maxRepeat
	}
	return max(n, 1), true
}
