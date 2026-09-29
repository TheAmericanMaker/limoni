# backdrop-shell

Your shell, in front of an animated scene, a picture, or your own ASCII art —
in any terminal with 256 colours or more.

```bash
backdrop-shell                          # $SHELL over the aurora
backdrop-shell -scene city              # aurora, city, starfield, synthwave
backdrop-shell -opacity 0.3             # quieter (0 to 1, default 0.45)
backdrop-shell -fps 10                  # cap the scene's frame rate
backdrop-shell -still                   # a still picture instead of an animation
backdrop-shell -image ~/Pictures/wall.jpg   # a picture as wallpaper
backdrop-shell -art art/cat.txt         # your own ASCII art, still or animated
backdrop-shell -- btop                  # any command instead of the shell
```

The files in [`art/`](art) are examples: a lemon, a cat that blinks, a bird
that flies, clouds and rain. **[docs/backdrop-art.md](../../docs/backdrop-art.md)**
(Türkçe: [docs/tr/backdrop-art.md](../../docs/tr/backdrop-art.md)) shows how
to draw your own, colour it, and animate it frame by frame or across the
screen — and how to turn the output of chafa, jp2a or lolcat into one.

## Install

It needs Go 1.25 or newer.

```bash
curl -fsSL https://raw.githubusercontent.com/thebanri/limoni/main/apps/backdrop-shell/install.sh | sh
# with settings:
curl -fsSL https://raw.githubusercontent.com/thebanri/limoni/main/apps/backdrop-shell/install.sh | sh -s -- -scene city -opacity 0.3
# from a checkout of this repository:
sh apps/backdrop-shell/install.sh
```

The installer builds the binary into `$(go env GOPATH)/bin` and runs
`backdrop-shell enable`, which makes every new terminal start with the scene —
Alacritty, kitty, Konsole, an editor's terminal, all of them — by adding a few
marked lines to each installed shell's start-up file:

| Shell | Where |
| :--- | :--- |
| fish | `~/.config/fish/conf.d/limoni-backdrop.fish`, a file of its own |
| bash | a block at the top of `~/.bashrc` |
| zsh  | a block at the top of `~/.zshrc` (above a prompt theme's instant prompt) |

The lines hand an interactive shell over to backdrop-shell, which starts the
same shell again inside it. Scripts, SSH sessions and shells already inside
are left alone, and if the binary is gone the shell starts as usual.

## Turn it off, turn it on, remove it

```bash
backdrop-shell disable      # new terminals open without it; stays installed
backdrop-shell enable       # back on (flags change the settings: -scene city ...)
backdrop-shell status       # what is on, and the settings
backdrop-shell uninstall    # remove the lines, the settings and the binary
```

`uninstall.sh` next to `install.sh` does the last one, also over `curl | sh`.
`disable` takes out exactly the lines `enable` put in; the rest of each file is
left byte for byte as it was. For one terminal only, `LIMONI_BACKDROP=off`
starts the plain shell.

Settings live in `~/.config/limoni/backdrop.conf`:

```ini
scene = aurora     # aurora, city, starfield, synthwave
opacity = 0.45     # 0 to 1
fps = 0            # cap the frame rate; 0 is the scene's own
still = false      # a still picture instead of an animation
image =            # a picture file instead of a scene
art =              # an ASCII art file instead of a scene; beats image
```

`backdrop-shell enable -art ~/my-art.txt` (or `-image`, or `-scene`) switches
between them.

## Why a wrapper

A terminal has no layers: it is one grid of cells, and whatever writes last is
what shows. A scene drawn by another program beside the shell would paint over
your text, and your text would wipe it. So backdrop-shell runs the shell on a
pseudo-terminal, keeps an emulated screen of what it writes, and draws that
over the scene. Every cell the shell leaves without a background colour shows
the scene, faded into the terminal's own background (asked for with OSC 11);
a program that paints its own background — btop, a vim colour scheme — covers
it.

## What it costs

What reaches the terminal is Limoni's cell diff, so a background costs only
the cells it changes. Measured behind fish in kitty at 120×40, twenty seconds
each, CPU time of all threads, as a share of one core:

| Background | backdrop-shell | kitty |
| :--- | ---: | ---: |
| a picture (`-image`) | 0.00% | 0.01% |
| aurora, `-still` | 0.01% | 1.82% |
| aurora | 0.54% | 1.26% |
| synthwave | 0.77% | 1.45% |
| `art/rain.txt` | 0.22% | 2.30% |

kitty costs about the same with the aurora still and moving: most of its
share is its own. The scene stops, and the wrapper uses nothing, while the
window is out of focus, and while a full-screen program covers the whole
scene. A still background is drawn once and costs nothing while the shell is
idle. The full table is in the guide.

## Keys

- **Shift+PageUp / Shift+PageDown** scroll back through history. The
  terminal's own scrollback is set aside while the wrapper runs (it draws on
  the alternate screen); typing returns to the live screen.

Everything else goes to the shell as the terminal sent it.

## Off switches

`LIMONI_BACKDROP=off` starts the plain shell. So does running backdrop-shell
inside itself, or anywhere that is not a terminal (scripts, `ssh host cmd`).
