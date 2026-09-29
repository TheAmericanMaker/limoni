package main

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/thebanri/limoni/core/cell"
)

func TestSplitInputKeepsBackWhatTheWrapperHandles(t *testing.T) {
	in := []byte("ls\x1b[5;2~\x1b[A\x1b[I\x1b[6;2~x\x1b[O")
	var events []inputEvent
	pass := splitInput(in, nil, func(ev inputEvent) { events = append(events, ev) })
	if want := []byte("ls\x1b[Ax"); !bytes.Equal(pass, want) {
		t.Fatalf("passed %q, want %q", pass, want)
	}
	want := []inputEvent{inputOther, inputScrollUp, inputOther, inputOther, inputFocusIn, inputScrollDown, inputOther, inputFocusOut}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events %v, want %v", events, want)
	}
}

func TestTerminalColorAnswersAreParsedAndRemoved(t *testing.T) {
	in := []byte("a\x1b]10;rgb:ffff/8080/0000\x1b\\b\x1b]11;rgb:12/34/56\x07\x1b[?62;22cc")
	fg, rest := takeColor(in, "10")
	bg, rest := takeColor(rest, "11")
	if fg != cell.NewColorRGB(255, 128, 0) {
		t.Errorf("fg = %06x", uint32(fg))
	}
	if bg != cell.NewColorRGB(0x12, 0x34, 0x56) {
		t.Errorf("bg = %06x", uint32(bg))
	}
	end := da1End(rest)
	if end < 0 || rest[end] != 'c' || !bytes.HasPrefix(rest, []byte("ab\x1b[?62;22")) {
		t.Fatalf("DA1 not found in %q (end %d)", rest, end)
	}
	if da1End([]byte("\x1b[?62;2")) >= 0 {
		t.Fatal("a cut-off DA1 answer was taken as whole")
	}
}

// The shell's colours are kept as it asked for them, and a cell it left
// without a background stays without one, which is what lets the scene show.
func TestConvertCell(t *testing.T) {
	c := convertCell(&uv.Cell{Content: "x", Width: 1, Style: uv.Style{
		Fg: ansi.BasicColor(2), Attrs: uv.AttrBold, Underline: uv.UnderlineCurly,
	}})
	want := cell.Cell{Content: 'x', Style: cell.Style{
		Fg: cell.NewColorANSI(2), Modifier: cell.ModifierBold | cell.ModifierUndercurl,
	}}
	if c != want {
		t.Fatalf("got %+v, want %+v", c, want)
	}
	if c := convertCell(&uv.Cell{Content: " ", Width: 1, Style: uv.Style{Bg: ansi.IndexedColor(236)}}); c.Style.Bg != cell.NewColorANSI(236) {
		t.Fatalf("background lost: %+v", c)
	}
	if c := convertCell(&uv.Cell{Width: 0}); c.Content != cell.RuneContinuation {
		t.Fatalf("second column of a wide character: %+v", c)
	}
	if c := convertCell(nil); c != (cell.Cell{Content: ' '}) {
		t.Fatalf("missing cell: %+v", c)
	}
}

// The example art that ships next to the program, and that the guide
// points to, must load and choose as a background.
func TestExampleArtLoads(t *testing.T) {
	files, _ := filepath.Glob("art/*.txt")
	if len(files) == 0 {
		t.Fatal("no example art")
	}
	for _, f := range files {
		if _, err := chooseScene("", "", f, false); err != nil {
			t.Errorf("%s: %v", f, err)
		}
	}
	if _, err := chooseScene("", "", "art/missing.txt", false); err == nil {
		t.Error("a missing file was accepted")
	}
	if bd, _ := chooseScene("aurora", "", "art/cat.txt", true); bd.Interval() != 0 {
		t.Error("-still did not freeze animated art")
	}
}
