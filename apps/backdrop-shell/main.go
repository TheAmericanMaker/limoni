// backdrop-shell runs your shell in front of an animated scene, in any
// terminal with 256 colours or more.
//
//	backdrop-shell                        # $SHELL over the aurora
//	backdrop-shell -scene city -opacity 0.5
//	backdrop-shell -scene synthwave -still  # a picture, not an animation
//	backdrop-shell -- htop                # any command instead of the shell
//
// A terminal has no layers: one grid of cells, and whatever writes last is
// what shows. So the scene cannot run beside the shell; it has to know where
// the shell's text is. backdrop-shell runs the shell on a pseudo-terminal,
// keeps an emulated screen of what it writes, and draws that over the scene:
// every cell the shell leaves without a background colour shows the scene
// behind it, faded into the terminal's own background by -opacity.
//
// What reaches the terminal is Limoni's cell diff, so the scene costs only
// the cells it changes. It stops while the window is out of focus, and while
// a full-screen program covers it all; a still scene costs nothing at all.
//
// Shift+PageUp and Shift+PageDown scroll back through the history, since the
// terminal's own scrollback is set aside while the wrapper runs.
//
// To have it in every new terminal, "backdrop-shell enable" adds a few marked
// lines to the start-up files of fish, bash and zsh; "backdrop-shell disable"
// takes exactly those lines out again, and "backdrop-shell uninstall" removes
// the settings and the binary too. install.sh builds it and enables it.
//
// It starts the plain shell instead when it is already running inside
// itself, when it is not on a terminal at all, or on a 16-colour one.
package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/thebanri/limoni/backdrop"
	"github.com/thebanri/limoni/core/terminal"
	"golang.org/x/sys/unix"
)

// envNested is set for the shell, so a backdrop-shell started inside it
// runs the plain shell instead of a scene inside a scene.
const envNested = "LIMONI_BACKDROP_SHELL"

func main() {
	if handled, err := command(os.Args[1:]); handled {
		if err != nil {
			fmt.Fprintln(os.Stderr, "backdrop-shell:", err)
			os.Exit(1)
		}
		return
	}

	// The settings file gives the defaults; flags override them.
	s, err := loadSettings(settingsPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "backdrop-shell:", err)
	}
	opts := options{}
	flag.StringVar(&opts.scene, "scene", s.Scene, "scene: "+strings.Join(backdrop.Names(), ", "))
	flag.Float64Var(&opts.opacity, "opacity", s.Opacity, "how strongly the scene shows, from 0 to 1")
	flag.Float64Var(&opts.fps, "fps", s.FPS, "cap the scene's frame rate (0: the scene's own)")
	flag.BoolVar(&opts.still, "still", s.Still, "a still picture of the scene instead of an animation")
	flag.Usage = usage
	flag.Parse()
	opts.argv = flag.Args()
	if len(opts.argv) == 0 {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		opts.argv = []string{shell}
	}
	if backdrop.New(opts.scene) == nil {
		fmt.Fprintf(os.Stderr, "backdrop-shell: no scene %q; there are %s\n", opts.scene, strings.Join(backdrop.Names(), ", "))
		opts.scene = "aurora"
	}

	// The plain shell wherever a scene cannot or should not be drawn: inside
	// backdrop-shell already, turned off, not on a terminal, or on one with
	// 16 colours, where a gradient is blocks of eight.
	caps := terminal.DetectCapabilities()
	if os.Getenv(envNested) != "" || os.Getenv("LIMONI_BACKDROP") == "off" ||
		!isTerminal(os.Stdin) || !isTerminal(os.Stdout) ||
		(!caps.TrueColor && !caps.Colors256) {
		execPlain(opts.argv)
	}
	code, err := run(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "backdrop-shell:", err)
		execPlain(opts.argv)
	}
	os.Exit(code)
}

func usage() {
	fmt.Fprint(os.Stderr, `Usage:
  backdrop-shell [flags] [-- command args]   run a shell (or command) in front of a scene
  backdrop-shell enable [flags]              start it in every new terminal
  backdrop-shell disable                     stop starting it; stays installed
  backdrop-shell status                      what is on, and the settings
  backdrop-shell uninstall                   disable, and remove the settings and this binary

Flags (enable saves them as the defaults):
`)
	flag.PrintDefaults()
}

// execPlain replaces the process with the command, as if the wrapper had
// never been there.
func execPlain(argv []string) {
	path, err := exec.LookPath(argv[0])
	if err == nil {
		err = syscall.Exec(path, argv, os.Environ())
	}
	fmt.Fprintln(os.Stderr, "backdrop-shell:", err)
	os.Exit(127)
}

func isTerminal(f *os.File) bool {
	_, err := unix.IoctlGetWinsize(int(f.Fd()), unix.TIOCGWINSZ)
	return err == nil
}

func rgba(r, g, b uint8) color.Color { return color.RGBA{r, g, b, 0xff} }
