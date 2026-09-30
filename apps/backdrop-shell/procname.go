package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

// Terminals that name a tab after the program running in it — Konsole's
// default title is "%d : %n", the foreground process's directory and name —
// find the wrapper there, not the shell, and the tab reads "backdrop-shell"
// whatever runs inside. So the wrapper takes on the name of the program in
// the foreground of its own terminal, as the kernel reports it in
// /proc/<pid>/stat: bash at the prompt, htop while htop runs.
//
// Writing /proc/self/comm renames the main thread, which is the name the
// process has in stat and ps; prctl(PR_SET_NAME) would rename only whichever
// thread the goroutine happens to be on. Where there is no /proc (macOS)
// nothing changes.

// procName follows the foreground process group of the shell's terminal.
// It is asked on the shell's output, and shortly after Enter, for a command
// that starts without writing anything (sleep); never on a clock.
type procName struct {
	mu   sync.Mutex
	pgrp int    // the group last seen, 0 for none yet
	name string // the name last given, "" for none
}

// update renames the wrapper after the foreground program of the terminal
// behind ptmx, if that program changed since the last call: one ioctl when
// it did not.
func (p *procName) update(ptmx *os.File) {
	p.mu.Lock()
	defer p.mu.Unlock()
	pgrp, err := unix.IoctlGetInt(int(ptmx.Fd()), unix.TIOCGPGRP)
	if err != nil || pgrp <= 0 || pgrp == p.pgrp {
		return
	}
	p.pgrp = pgrp
	comm, err := os.ReadFile("/proc/" + strconv.Itoa(pgrp) + "/comm")
	if err != nil {
		return
	}
	p.setLocked(strings.TrimSpace(string(comm)))
}

// set renames the wrapper, if the name is new.
func (p *procName) set(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.setLocked(name)
}

func (p *procName) setLocked(name string) {
	if name == "" || name == p.name {
		return
	}
	if os.WriteFile("/proc/self/comm", []byte(name), 0) == nil {
		p.name = name
	}
}

// withOwnDir returns env with the directory of this binary at the end of
// PATH, if it is not in PATH already. install.sh puts the binary in
// $(go env GOPATH)/bin, which many systems — Ubuntu and Kubuntu among them —
// leave out of PATH: the start-up hook names the binary by its full path, so
// the scene appears, but "backdrop-shell opacity 0.3" typed at the prompt
// would be "command not found".
func withOwnDir(env []string, exe string) []string {
	dir := filepath.Dir(exe)
	if exe == "" || !filepath.IsAbs(dir) {
		return env
	}
	for i, kv := range env {
		path, ok := strings.CutPrefix(kv, "PATH=")
		if !ok {
			continue
		}
		for _, d := range filepath.SplitList(path) {
			if filepath.Clean(d) == dir {
				return env
			}
		}
		out := append([]string(nil), env...)
		if path == "" {
			out[i] = "PATH=" + dir
		} else {
			out[i] = kv + string(filepath.ListSeparator) + dir
		}
		return out
	}
	return append(env, "PATH="+dir)
}
