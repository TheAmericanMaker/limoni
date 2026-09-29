// Backdrop: animated scenes behind an ordinary application.
//
//	go run ./examples/backdrop            # starts with the aurora
//	go run ./examples/backdrop city       # or city, starfield, synthwave
//
// The panel paints a background colour of its own and hides the scene; the
// text under it paints none and floats over it. The application draws only
// when a key arrives — the scene moves by itself, and what reaches the
// terminal is the cells it changes.
//
// Keys: ← → change scene · d dim · h hide the panel · q / Esc quit
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/thebanri/limoni"
	"github.com/thebanri/limoni/backdrop"
	"github.com/thebanri/limoni/core/cell"
)

var (
	panelBg = limoni.RGB(14, 16, 28)
	accent  = limoni.RGB(120, 230, 210)
	muted   = limoni.RGB(130, 140, 165)
	text    = limoni.RGB(215, 220, 235)
)

type demo struct {
	term   *limoni.Terminal
	names  []string
	scene  int
	dim    bool
	hidden bool
}

func (d *demo) apply() {
	s := backdrop.New(d.names[d.scene])
	if d.dim {
		s = backdrop.Dim(s, 0.55)
	}
	d.term.SetBackdrop(s)
}

func (d *demo) frame(f *limoni.Frame, ev *limoni.Event) bool {
	if ev != nil && ev.Type == limoni.EventKey && !ev.Key.Release {
		switch k := ev.Key; {
		case k.Type == limoni.KeyEsc, k.Type == limoni.KeyRune && k.Ch == 'q':
			return false
		case k.Type == limoni.KeyRight, k.Type == limoni.KeyRune && k.Ch == 'n':
			d.scene = (d.scene + 1) % len(d.names)
			d.apply()
		case k.Type == limoni.KeyLeft, k.Type == limoni.KeyRune && k.Ch == 'p':
			d.scene = (d.scene + len(d.names) - 1) % len(d.names)
			d.apply()
		case k.Type == limoni.KeyRune && k.Ch == 'd':
			d.dim = !d.dim
			d.apply()
		case k.Type == limoni.KeyRune && k.Ch == 'h':
			d.hidden = !d.hidden
		}
	}

	area := f.Area()
	hint := " ← →  scene   d  dim   h  panel   q  quit "
	f.Buffer.SetString(uint16(max(0, int(area.Width)-len([]rune(hint))-1)), area.Height-1, hint,
		limoni.Style{Fg: muted})
	if d.hidden {
		return true
	}

	const w, h = 46, 12
	if area.Width < w+2 || area.Height < h+4 {
		f.Buffer.SetString(1, 1, "make the window larger", limoni.Style{Fg: text})
		return true
	}
	x, y := (area.Width-w)/2, (area.Height-h)/2-1
	panel := cell.NewRect(x, y, w, h)
	st := limoni.Style{Fg: text, Bg: panelBg}
	f.RenderWidget(limoni.NewBlock().Rounded().
		WithTitle(" limoni · backdrop ").
		WithTitleStyle(limoni.Style{Fg: accent, Bg: panelBg, Modifier: limoni.ModifierBold}).
		WithBorderStyle(limoni.Style{Fg: accent, Bg: panelBg}).
		WithStyle(st), panel)

	row := y + 2
	for i, n := range d.names {
		mark, style := "   ", st
		if i == d.scene {
			mark, style = " ▸ ", limoni.Style{Fg: accent, Bg: panelBg, Modifier: limoni.ModifierBold}
		}
		f.Buffer.SetString(x+3, row, mark+n, style)
		row++
	}
	row++
	dim := "off"
	if d.dim {
		dim = "on"
	}
	pace := "still"
	if iv := d.term.BackdropInterval(); iv > 0 {
		pace = fmt.Sprintf("%d fps, on its own", time.Second/iv)
	}
	f.Buffer.SetString(x+3, row, fmt.Sprintf("dim %-3s   scene: %s", dim, pace),
		limoni.Style{Fg: muted, Bg: panelBg})
	row++
	f.Buffer.SetString(x+3, row, "LIMONI_BACKDROP=off turns it off anywhere", limoni.Style{Fg: muted, Bg: panelBg})

	// No background colour: this line floats over the scene.
	line := "This text paints no background, so the scene shows behind it."
	f.Buffer.SetString(uint16(max(0, (int(area.Width)-len(line))/2)), y+h+1, line, limoni.Style{Fg: text})
	return true
}

func main() {
	d := &demo{names: backdrop.Names()}
	if len(os.Args) > 1 {
		for i, n := range d.names {
			if n == os.Args[1] {
				d.scene = i
			}
		}
	}
	term, err := limoni.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer term.Close()
	d.term = term
	app := limoni.NewApp(term,
		limoni.WithTitle("Limoni · Backdrop"),
		limoni.WithBackdrop(backdrop.New(d.names[d.scene])))
	if err := app.Run(context.Background(), d.frame); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
