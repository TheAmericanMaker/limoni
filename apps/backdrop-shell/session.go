package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
	"github.com/thebanri/limoni/backdrop"
	"github.com/thebanri/limoni/core/buffer"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/core/driver"
	"github.com/thebanri/limoni/core/terminal"
)

// minFrame is the shortest time between two frames drawn for the shell's
// output. A command that prints a flood is drawn at most this often; a key
// typed after a pause is drawn at once.
const minFrame = 8 * time.Millisecond

// scrollbackLines is how much history Shift+PageUp can reach.
const scrollbackLines = 10000

// forwardedModes are the modes the shell sets that change what the real
// terminal has to send: cursor key encoding, mouse reports and bracketed
// paste. They are passed on as they come. The rest — the alternate screen,
// the cursor, origin mode — belong to the emulated screen and stay there.
var forwardedModes = map[ansi.DECMode]bool{
	ansi.ModeCursorKeys:       true,
	ansi.ModeMouseX10:         true,
	ansi.ModeMouseNormal:      true,
	ansi.ModeMouseButtonEvent: true,
	ansi.ModeMouseAnyEvent:    true,
	ansi.ModeMouseExtSgr:      true,
	ansi.DECMode(1005):        true, // UTF-8 mouse
	ansi.DECMode(1015):        true, // urxvt mouse
	ansi.ModeBracketedPaste:   true,
}

type options struct {
	scene   string
	opacity float64
	fps     float64
	still   bool
	image   string
	art     string
	argv    []string

	// backdrop is the scene the options above chose, loaded before the
	// terminal is taken over so a bad file is reported as usual.
	backdrop terminal.Backdrop
}

type session struct {
	opts   options
	in     *os.File
	out    *os.File
	caps   terminal.CapabilityProfile
	scene  terminal.Backdrop
	start  time.Time
	ptmx   *os.File
	emu    *vt.Emulator
	w, h   int
	cursor struct {
		hidden bool
		shape  int // DECSCUSR parameter, 0 for the terminal's default
	}
	childModes map[ansi.DECMode]bool
	scroll     int  // lines scrolled back into history, 0 at the live screen
	focused    bool // the terminal window has focus
	sceneFrame int  // the scene frame last rendered, -1 for none

	app, bg, front, back *buffer.Buffer
	appDirty             bool
	pending              []byte // escape sequences for the real terminal, sent with the next frame
	outBuf               []byte
	lastFrame            time.Time
}

type inputMsg struct {
	events []inputEvent
}

// run starts the shell behind the scene and returns its exit code.
func run(opts options) (int, error) {
	s := &session{
		opts:       opts,
		in:         os.Stdin,
		out:        os.Stdout,
		caps:       terminal.DetectCapabilities(),
		childModes: map[ansi.DECMode]bool{},
		focused:    true,
		sceneFrame: -1,
	}
	s.w, s.h = windowSize(int(s.out.Fd()))

	state, err := driver.MakeRaw(int(s.in.Fd()))
	if err != nil {
		return 1, err
	}
	defer driver.Restore(int(s.in.Fd()), state)

	fg, bg, typeahead := terminalColors(s.in, s.out)
	s.scene = backdrop.Fade(opts.backdrop, bg, opts.opacity)

	cmd := exec.Command(opts.argv[0], opts.argv[1:]...)
	cmd.Env = append(os.Environ(), envNested+"=1")
	s.ptmx, err = pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(s.w), Rows: uint16(s.h)})
	if err != nil {
		return 1, err
	}
	defer s.ptmx.Close()

	s.emu = vt.NewEmulator(s.w, s.h)
	s.emu.SetScrollbackSize(scrollbackLines)
	if fg.Type() == cell.ColorRGB {
		r, g, b := fg.RGB()
		s.emu.SetDefaultForegroundColor(rgba(r, g, b))
	}
	if bg.Type() == cell.ColorRGB {
		r, g, b := bg.RGB()
		s.emu.SetDefaultBackgroundColor(rgba(r, g, b))
	}
	s.emu.SetCallbacks(vt.Callbacks{
		Bell:             func() { s.pending = append(s.pending, '\a') },
		Title:            func(t string) { s.pending = append(s.pending, ansi.SetWindowTitle(t)...) },
		CursorVisibility: func(v bool) { s.cursor.hidden = !v },
		CursorStyle:      s.setCursorShape,
		EnableMode:       func(m ansi.Mode) { s.setChildMode(m, true) },
		DisableMode:      func(m ansi.Mode) { s.setChildMode(m, false) },
	})

	s.app = buffer.NewBuffer(cell.NewRect(0, 0, uint16(s.w), uint16(s.h)))
	s.bg = buffer.NewBuffer(s.app.Area)
	s.front = buffer.NewBuffer(s.app.Area)
	s.back = buffer.NewEmptyBuffer() // a size mismatch: the first frame is drawn whole

	// The alternate screen, so the user's scrollback is left as it was; focus
	// reports, so the scene can rest while the window is in the background;
	// and no alternate scroll, or the wheel would send arrow keys to the
	// shell and walk its history.
	s.write("\x1b[?1049h\x1b[?1004h\x1b[?1007l\x1b[H\x1b[2J")
	defer s.write("\x1b[?1004l\x1b[?1049l\x1b[0 q\x1b[?25h" + resetForwardedModes())

	output := make(chan []byte, 64)
	go readPTY(s.ptmx, output)
	go func() { _, _ = io.Copy(s.ptmx, s.emu) }() // the emulator's answers to the shell's queries
	inputs := make(chan inputMsg, 16)
	go s.readInput(inputs)
	if len(typeahead) > 0 {
		_, _ = s.ptmx.Write(typeahead)
	}

	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	defer signal.Stop(winch)

	s.start = time.Now()
	var ticker *time.Ticker
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()
	settle := time.NewTimer(time.Hour)
	settle.Stop()
	var settled <-chan time.Time

	s.draw()
	for {
		// The scene ticks only while it can be seen: the window has focus
		// and something on screen is left clear for it.
		var tick <-chan time.Time
		if iv := s.interval(); iv > 0 && s.focused && s.sceneVisible() {
			if ticker == nil {
				ticker = time.NewTicker(iv)
			}
			tick = ticker.C
		} else if ticker != nil {
			ticker.Stop()
			ticker = nil
		}

		select {
		case data, ok := <-output:
			if !ok {
				return exitCode(cmd), nil
			}
			s.feed(data, output)
			if since := time.Since(s.lastFrame); since < minFrame {
				if settled == nil {
					settle.Reset(minFrame - since)
					settled = settle.C
				}
				continue
			}
			s.draw()
		case <-settled:
			settled = nil
			s.draw()
		case msg := <-inputs:
			s.handleInput(msg)
		case <-tick:
			s.draw()
		case <-winch:
			s.resize()
		}
	}
}

// interval is how often the scene is drawn: its own pace, or slower if the
// user capped the frame rate.
func (s *session) interval() time.Duration {
	iv := s.scene.Interval()
	if iv > 0 && s.opts.fps > 0 {
		iv = max(iv, time.Duration(float64(time.Second)/s.opts.fps))
	}
	return iv
}

// feed hands the shell's output to the emulator, along with whatever else
// is already waiting, so a burst is one frame.
func (s *session) feed(data []byte, output chan []byte) {
	_, _ = s.emu.Write(data)
	for {
		select {
		case more, ok := <-output:
			if !ok {
				s.appDirty = true
				return
			}
			_, _ = s.emu.Write(more)
			continue
		default:
		}
		break
	}
	s.appDirty = true
	s.scroll = min(s.scroll, s.emu.ScrollbackLen())
}

func (s *session) handleInput(msg inputMsg) {
	for _, ev := range msg.events {
		switch ev {
		case inputScrollUp:
			if !s.emu.IsAltScreen() {
				s.scroll = min(s.scroll+s.h/2, s.emu.ScrollbackLen())
				s.appDirty = true
			}
		case inputScrollDown:
			s.scroll = max(0, s.scroll-s.h/2)
			s.appDirty = true
		case inputFocusIn, inputFocusOut:
			s.focused = ev == inputFocusIn
			if s.childModes[ansi.ModeFocusEvent] {
				report := reportFocusIn
				if ev == inputFocusOut {
					report = reportFocusOut
				}
				_, _ = s.ptmx.Write(report)
			}
		case inputOther:
			if s.scroll > 0 {
				s.scroll = 0
				s.appDirty = true
			}
		}
	}
	if s.appDirty {
		s.draw()
	}
}

func (s *session) resize() {
	w, h := windowSize(int(s.out.Fd()))
	if w == s.w && h == s.h {
		return
	}
	s.w, s.h = w, h
	_ = pty.Setsize(s.ptmx, &pty.Winsize{Cols: uint16(w), Rows: uint16(h)})
	s.emu.Resize(w, h)
	area := cell.NewRect(0, 0, uint16(w), uint16(h))
	s.app.Resize(area)
	s.bg.Resize(area)
	s.front.Resize(area)
	s.sceneFrame = -1
	s.appDirty = true
	s.scroll = 0
	s.draw()
}

// draw composes a frame and sends what changed since the last one.
func (s *session) draw() {
	s.lastFrame = time.Now()
	if s.appDirty {
		s.copyScreen()
		s.appDirty = false
	}
	// The scene is drawn again only when it has moved on a frame; a still
	// one only after a resize.
	if iv := s.interval(); iv <= 0 {
		if s.sceneFrame < 0 {
			s.scene.Render(s.bg, 0)
			s.sceneFrame = 0
		}
	} else if f := int(time.Since(s.start) / iv); f != s.sceneFrame {
		s.scene.Render(s.bg, time.Duration(f)*iv)
		s.sceneFrame = f
	}
	copy(s.front.Content, s.app.Content)
	terminal.ComposeBackdrop(s.front.Content, s.bg.Content)
	s.front.Invalidate()

	out := append(s.outBuf[:0], "\x1b[?2026h"...)
	out = append(out, s.pending...)
	s.pending = s.pending[:0]
	bodyStart := len(out)
	out, _ = buffer.DiffWithOptions(s.front, s.back, out, buffer.DiffOptions{
		TrueColor:  s.caps.TrueColor,
		Colors256:  s.caps.Colors256,
		EraseChar:  s.caps.EraseChar,
		RepeatChar: s.caps.RepeatChar,
	})
	cx, cy := s.cursorAt()
	out = append(out, "\x1b["...)
	out = strconv.AppendInt(out, int64(cy+1), 10)
	out = append(out, ';')
	out = strconv.AppendInt(out, int64(cx+1), 10)
	out = append(out, 'H')
	if s.cursor.hidden || s.scroll > 0 {
		out = append(out, "\x1b[?25l"...)
	} else {
		out = append(out, "\x1b[?25h"...)
	}
	out = append(out, "\x1b[?2026l"...)
	s.outBuf = out
	if len(out) > bodyStart {
		s.write(string(out))
	}
}

// cursorAt is where the real cursor goes: the shell's cursor, or the
// bottom-right corner while history is shown.
func (s *session) cursorAt() (x, y int) {
	if s.scroll > 0 {
		return s.w - 1, s.h - 1
	}
	p := s.emu.CursorPosition()
	return min(max(p.X, 0), s.w-1), min(max(p.Y, 0), s.h-1)
}

// copyScreen converts the emulated screen — or, scrolled back, the stretch
// of history and screen being looked at — into Limoni cells.
func (s *session) copyScreen() {
	w, h := s.w, s.h
	history := s.emu.ScrollbackLen()
	for y := 0; y < h; y++ {
		line := history - s.scroll + y
		row := s.app.Content[y*w : (y+1)*w]
		for x := 0; x < w; x++ {
			var c *uv.Cell
			if line < history {
				c = s.emu.ScrollbackCellAt(x, line)
			} else {
				c = s.emu.CellAt(x, line-history)
			}
			row[x] = convertCell(c)
		}
	}
}

// sceneVisible reports whether any cell leaves the scene showing. A
// full-screen program with its own background covers it all, and then the
// scene need not move.
func (s *session) sceneVisible() bool {
	for _, c := range s.app.Content {
		if c.Style.Bg.Type() == cell.ColorDefault && c.Style.Modifier&cell.ModifierReverse == 0 {
			return true
		}
	}
	return false
}

func (s *session) setCursorShape(style vt.CursorStyle, blink bool) {
	// DECSCUSR: 1-2 block, 3-4 underline, 5-6 bar; odd blinks.
	shape := 2*int(style) + 1
	if !blink {
		shape++
	}
	if shape == s.cursor.shape {
		return
	}
	s.cursor.shape = shape
	s.pending = append(s.pending, "\x1b["...)
	s.pending = strconv.AppendInt(s.pending, int64(shape), 10)
	s.pending = append(s.pending, " q"...)
}

func (s *session) setChildMode(m ansi.Mode, on bool) {
	dm, ok := m.(ansi.DECMode)
	if !ok {
		return
	}
	s.childModes[dm] = on
	switch {
	case dm == ansi.ModeNumericKeypad:
		// DECKPAM / DECKPNM
		if on {
			s.pending = append(s.pending, "\x1b="...)
		} else {
			s.pending = append(s.pending, "\x1b>"...)
		}
	case forwardedModes[dm]:
		s.pending = append(s.pending, "\x1b[?"...)
		s.pending = strconv.AppendInt(s.pending, int64(dm), 10)
		if on {
			s.pending = append(s.pending, 'h')
		} else {
			s.pending = append(s.pending, 'l')
		}
	}
	if len(s.pending) > 0 && s.emu != nil {
		// A mode change must reach the terminal before the next key it
		// affects, not with the next frame.
		s.write(string(s.pending))
		s.pending = s.pending[:0]
	}
}

// readInput passes the keyboard to the shell, keeping back what the wrapper
// handles itself.
func (s *session) readInput(inputs chan<- inputMsg) {
	buf := make([]byte, 4096)
	var pass []byte
	for {
		n, err := s.in.Read(buf)
		if n > 0 {
			var events []inputEvent
			pass = splitInput(buf[:n], pass[:0], func(ev inputEvent) {
				if ev == inputOther && len(events) > 0 && events[len(events)-1] == inputOther {
					return
				}
				events = append(events, ev)
			})
			if len(pass) > 0 {
				_, _ = s.ptmx.Write(pass)
			}
			if len(events) > 0 {
				inputs <- inputMsg{events: events}
			}
		}
		if err != nil {
			return
		}
	}
}

func readPTY(ptmx *os.File, output chan<- []byte) {
	defer close(output)
	for {
		buf := make([]byte, 32*1024)
		n, err := ptmx.Read(buf)
		if n > 0 {
			output <- buf[:n]
		}
		if err != nil {
			return
		}
	}
}

func (s *session) write(str string) {
	_, _ = s.out.WriteString(str)
}

func resetForwardedModes() string {
	out := "\x1b>"
	for m := range forwardedModes {
		out += "\x1b[?" + strconv.Itoa(int(m)) + "l"
	}
	return out
}

func exitCode(cmd *exec.Cmd) int {
	err := cmd.Wait()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	if err != nil {
		return 1
	}
	return 0
}
