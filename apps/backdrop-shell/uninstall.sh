#!/bin/sh
# Removes backdrop-shell: the lines it added to your shells, its settings and
# the binary. To only turn it off and keep it, run "backdrop-shell disable".
#
#   curl -fsSL https://raw.githubusercontent.com/thebanri/limoni/main/apps/backdrop-shell/uninstall.sh | sh
set -eu

bin=""
if command -v go >/dev/null 2>&1; then
	dir=$(go env GOBIN)
	[ -n "$dir" ] || dir="$(go env GOPATH)/bin"
	[ -x "$dir/backdrop-shell" ] && bin="$dir/backdrop-shell"
fi
[ -n "$bin" ] || bin=$(command -v backdrop-shell 2>/dev/null || true)
[ -n "$bin" ] || { echo "backdrop-shell is not installed." >&2; exit 1; }

"$bin" uninstall
