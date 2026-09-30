#!/bin/sh
# Installs backdrop-shell and turns it on for every new terminal.
#
#   curl -fsSL https://raw.githubusercontent.com/thebanri/limoni/main/apps/backdrop-shell/install.sh | sh
#   curl -fsSL .../install.sh | sh -s -- -scene city -opacity 0.3   # with settings
#   sh apps/backdrop-shell/install.sh                               # from a checkout
#
# It needs Go 1.25 or newer. Flags are passed to "backdrop-shell enable":
# -scene, -opacity, -fps, -still, -shells. Change them later by running
# "backdrop-shell enable" again with others, or by editing
# ~/.config/limoni/backdrop.conf.
#
# Turn it off:  backdrop-shell disable      (stays installed; enable brings it back)
# Remove it:    backdrop-shell uninstall    (or uninstall.sh next to this file)
set -eu

say() { printf '%s\n' "$*"; }
fail() { say "backdrop-shell: $*" >&2; exit 1; }

command -v go >/dev/null 2>&1 || fail "Go is not installed. Get it from https://go.dev/dl/ (1.25 or newer) and run this again."
version=$(go env GOVERSION | sed 's/^go//')
major=${version%%.*}
rest=${version#*.}
minor=${rest%%[!0-9]*}
if [ "$major" -lt 1 ] || { [ "$major" -eq 1 ] && [ "$minor" -lt 25 ]; }; then
	fail "Go $version is too old; 1.25 or newer is needed."
fi

bindir=$(go env GOBIN)
[ -n "$bindir" ] || bindir="$(go env GOPATH)/bin"

# From a checkout, build what is in it, since the scenes may be newer than
# the last release. A temporary workspace joins the app to the library next
# to it without touching either go.mod.
here=$(cd "$(dirname "$0")" 2>/dev/null && pwd || true)
if [ -n "$here" ] && [ -f "$here/main.go" ] && [ -f "$here/../../go.mod" ]; then
	say "Building from $here"
	work=$(mktemp -d)
	trap 'rm -rf "$work"' EXIT
	root=$(cd "$here/../.." && pwd)
	printf 'go 1.25.0\n\nuse (\n\t%s\n\t%s\n)\n' "$root" "$here" > "$work/go.work"
	(cd "$here" && GOWORK="$work/go.work" go install .)
else
	say "Installing github.com/thebanri/limoni/apps/backdrop-shell@latest"
	go install github.com/thebanri/limoni/apps/backdrop-shell@latest
fi

bin="$bindir/backdrop-shell"
[ -x "$bin" ] || fail "the build finished but $bin is not there"
say "Installed $bin"
"$bin" enable "$@"

# Ubuntu, Kubuntu and others leave Go's bin directory out of PATH. The
# terminals backdrop-shell runs in get it added, but a plain shell (after
# "backdrop-shell disable", or over SSH) would not find the command.
case ":$PATH:" in
*":$bindir:"*) ;;
*)
	say ""
	say "Note: $bindir is not in your PATH. Terminals opened with the background"
	say "have it; to run backdrop-shell anywhere else, add this line to your"
	say "shell's start-up file (~/.bashrc, ~/.zshrc):"
	say "    export PATH=\"\$PATH:$bindir\""
	say "or for fish: fish_add_path $bindir"
	;;
esac
