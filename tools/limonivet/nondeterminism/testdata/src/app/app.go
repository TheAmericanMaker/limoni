package app

import (
	"context"
	"math/rand"
	"time"
)

type Msg any
type Cmd func(context.Context) Msg
type UpdateResult struct{ Commands []Cmd }
type Frame struct{}

type model struct{ seen time.Time }

func (m *model) Init() []Cmd {
	_ = time.Now() // want `time.Now in Init cannot be replayed`
	return nil
}

func (m *model) Update(msg Msg) UpdateResult {
	m.seen = time.Now() // want `time.Now in Update cannot be replayed`
	_ = rand.Intn(10)   // want `math/rand.Intn in Update cannot be replayed`

	// Correct: the clock is read inside a command, and its result comes back
	// as a message the recording keeps.
	tick := func(ctx context.Context) Msg { return time.Now() }

	// Not a command, so still inside the model.
	helper := func() time.Time { return time.Now() } // want `time.Now in Update cannot be replayed`
	_ = helper
	return UpdateResult{Commands: []Cmd{tick}}
}

func (m *model) View(f *Frame) {
	_ = time.Since(m.seen) // want `time.Since in View cannot be replayed`
	_ = time.Now()         // limonivet:allow deliberate, and not reported
}

// Not a model method: wrong shape, so not reported.
type other struct{}

func (other) Update() { _ = time.Now() }

func (other) View(s string) { _ = time.Now() }

// Deterministic uses of the time package are fine.
func (m *model) formatted() string { return m.seen.Format(time.RFC3339) }
