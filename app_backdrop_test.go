package limoni

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/thebanri/limoni/core/buffer"
	"github.com/thebanri/limoni/core/cell"
	"github.com/thebanri/limoni/core/driver"
	"github.com/thebanri/limoni/core/terminal"
)

type pulse struct{ renders atomic.Int32 }

func (p *pulse) Interval() time.Duration { return 10 * time.Millisecond }

func (p *pulse) Render(dst *buffer.Buffer, t time.Duration) {
	p.renders.Add(1)
	v := uint8(t / time.Millisecond)
	for i := range dst.Content {
		dst.Content[i] = cell.Cell{Content: ' ', Style: cell.Style{Bg: cell.NewColorRGB(0, v, 0)}}
	}
}

// WithBackdrop keeps the scene moving in an App that draws only on events,
// without calling the application function for it.
func TestRunMovesTheBackdropWithoutCallingTheApplication(t *testing.T) {
	t.Setenv("LIMONI_PROBE", "0")
	t.Setenv("COLORTERM", "truecolor")
	t.Setenv("LIMONI_BACKDROP", "")
	in := make(chan []byte, 4)
	b := driver.NewPortableBackend(&feedIO{in: in})
	if err := b.Setup(); err != nil {
		t.Fatal(err)
	}
	term, err := terminal.New(b)
	if err != nil {
		t.Fatal(err)
	}
	scene := &pulse{}
	var frames atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- NewApp(term, WithBackdrop(scene)).Run(ctx, func(f *Frame, ev *Event) bool {
			frames.Add(1)
			return true
		})
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Errorf("Run = %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("Run did not stop")
		}
		_ = term.Close()
	})

	waitFor(t, func() bool { return frames.Load() == 1 })
	time.Sleep(200 * time.Millisecond)
	if n := frames.Load(); n != 1 {
		t.Fatalf("the application drew %d frames; the backdrop should move without it", n)
	}
	if n := scene.renders.Load(); n < 5 {
		t.Fatalf("the backdrop was drawn %d times in 200ms at 100 fps", n)
	}
}
