#!/usr/bin/env python3
"""Fireflies drifting in the dark: an example scene for backdrop-shell.

    backdrop-shell -scene-cmd scenes/fireflies.py

A scene program draws on its standard output as on a terminal. backdrop-shell
runs it on a terminal the size of the screen (also given in LIMONI_COLS and
LIMONI_ROWS), starts it again when the window changes size, and pauses it
while the scene cannot be seen. Each frame begins with ESC [ 2 J ESC [ H:
clear, then home. backdrop-shell shows a frame once the next one begins, so
a half-drawn frame never shows.

Cells left without a background colour show the terminal's own, so the
fireflies float over whatever the terminal looks like. Copy this file, change
what frame() draws, and point -scene-cmd at the copy.
"""

import math
import os
import random
import sys
import time

cols = int(os.environ.get("LIMONI_COLS", "80"))
rows = int(os.environ.get("LIMONI_ROWS", "24"))
FPS = 10

rng = random.Random(1)
flies = [
    dict(x=rng.uniform(0, cols), y=rng.uniform(0, rows),
         phase=rng.uniform(0, 2 * math.pi), speed=rng.uniform(0.6, 1.4),
         pulse=rng.uniform(0.15, 0.4))
    for _ in range(max(6, cols * rows // 90))
]


def frame(t):
    out = ["\x1b[2J\x1b[H"]
    for f in flies:
        # Wander along a slow curve, wrapping around the edges.
        x = (f["x"] + 3 * math.sin(t * 0.3 * f["speed"] + f["phase"]) + t * 0.4 * f["speed"]) % cols
        y = (f["y"] + 2 * math.cos(t * 0.25 * f["speed"] + f["phase"])) % rows
        glow = 0.5 + 0.5 * math.sin(t * f["pulse"] * 6 + f["phase"])
        if glow < 0.15:
            continue  # between flashes
        # Few colour steps: a cell is sent again only when its colour changes.
        k = round(glow * 4) / 4
        r, g, b = int(120 + 135 * k), int(140 + 115 * k), int(40 * k)
        dot = "•" if k > 0.5 else "·"
        out.append(f"\x1b[{int(y) + 1};{int(x) + 1}H\x1b[38;2;{r};{g};{b}m{dot}")
    out.append("\x1b[0m")
    return "".join(out)


start = time.monotonic()
while True:
    sys.stdout.write(frame(time.monotonic() - start))
    sys.stdout.flush()
    time.sleep(1 / FPS)
