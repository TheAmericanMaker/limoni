package terminal

import (
	"os"
	"runtime"
	"strings"

	"github.com/thebanri/limoni/graphics"
)

// CapabilityProfile defines the capability flags supported by the active terminal.
type CapabilityProfile struct {
	TrueColor      bool
	Colors256      bool
	MouseSupport   bool
	BracketedPaste bool
	SyncOutput     bool
	GraphicsProto  graphics.Protocol

	// EraseChar enables ECH (CSI n X) for runs of blanks. It is ECMA-48 and
	// implemented essentially everywhere, so it is on by default.
	EraseChar bool
	// RepeatChar enables REP (CSI n b) for runs of one glyph. Also ECMA-48, but
	// unevenly implemented — a terminal without it would print the escape and
	// corrupt the frame — so it stays off unless the terminal is recognised.
	// This becomes a runtime query once the capability handshake lands.
	RepeatChar bool
}

// DetectCapabilities automatically detects the active terminal's capability profile using environment variables.
func DetectCapabilities() CapabilityProfile {
	profile := CapabilityProfile{
		TrueColor:      false,
		Colors256:      false,
		MouseSupport:   true, // Most modern terminals support mouse reporting
		BracketedPaste: true, // Most modern terminals support bracketed paste
		SyncOutput:     true, // Synchronized Output (?2026) enables atomic tear-free frames (safely ignored if unsupported)
		GraphicsProto:  graphics.DetectProtocol(),
		EraseChar:      true,
	}

	// Under js/wasm there is no process environment to inspect, but the host is
	// a browser terminal emulator (xterm.js and friends), all of which speak
	// 24-bit color. Without this the browser playground would be downsampled to
	// 16 colors purely because COLORTERM is absent.
	if runtime.GOOS == "js" {
		profile.TrueColor = true
		profile.Colors256 = true
		// xterm.js implements REP.
		profile.RepeatChar = true
		return profile
	}

	term := os.Getenv("TERM")
	if term == "dumb" || os.Getenv("LIMONI_NO_SYNC") == "1" {
		profile.SyncOutput = false
	}
	if term == "dumb" {
		profile.EraseChar = false
	}

	// 1. Detect TrueColor support
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		profile.TrueColor = true
		profile.Colors256 = true
	}

	if strings.Contains(term, "direct") {
		profile.TrueColor = true
		profile.Colors256 = true
	} else if strings.Contains(term, "256color") {
		profile.Colors256 = true
	}

	// Some known modern terminals support TrueColor
	termProg := os.Getenv("TERM_PROGRAM")
	if termProg == "kitty" || termProg == "WezTerm" || termProg == "Ghostty" || termProg == "iTerm.app" || termProg == "Apple_Terminal" {
		profile.TrueColor = true
		profile.Colors256 = true
	}

	// REP is only enabled where it is known to work. xterm defined it; VTE,
	// kitty, foot, WezTerm and Ghostty implement it. Anything unrecognised
	// keeps it off rather than risking a literal escape on screen.
	switch {
	case termProg == "kitty", termProg == "WezTerm", termProg == "Ghostty",
		termProg == "foot", termProg == "iTerm.app":
		profile.RepeatChar = true
	case strings.HasPrefix(term, "xterm"), strings.HasPrefix(term, "vte"),
		strings.HasPrefix(term, "kitty"), strings.HasPrefix(term, "foot"),
		strings.HasPrefix(term, "alacritty"), strings.HasPrefix(term, "wezterm"):
		profile.RepeatChar = true
	}
	// Escape hatches in both directions, until the handshake can ask.
	switch os.Getenv("LIMONI_REP") {
	case "1":
		profile.RepeatChar = true
	case "0":
		profile.RepeatChar = false
	}

	return profile
}
