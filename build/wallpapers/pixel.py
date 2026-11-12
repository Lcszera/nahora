"""Jogos 1: paisagem em pixel art (240x135 ampliada 8x, sem suavização)."""
import random, sys, math
from PIL import Image, ImageDraw

w, h, S = 240, 135, 8
r = random.Random(11)
img = Image.new("RGB", (w, h))
d = ImageDraw.Draw(img)

sky = [(20, 18, 54), (32, 26, 78), (52, 36, 104), (84, 48, 128), (128, 64, 140), (176, 86, 138), (222, 120, 128), (246, 164, 120)]
band = 90 // len(sky)
for i, c in enumerate(sky):
    d.rectangle([0, i * band, w, (i + 1) * band + 1], fill=c)
for _ in range(70):
    x, y = r.randrange(w), r.randrange(45)
    img.putpixel((x, y), (255, 255, 230) if r.random() < .7 else (180, 220, 255))
for (x, y) in [(40, 10), (180, 18), (120, 6)]:
    for dx, dy in [(0, -1), (0, 1), (-1, 0), (1, 0), (0, 0)]:
        img.putpixel((x + dx, y + dy), (255, 255, 255))
# lua
for yy in range(-7, 8):
    for xx in range(-7, 8):
        if xx * xx + yy * yy <= 49 and not ((xx - 3) ** 2 + (yy + 2) ** 2 <= 30):
            img.putpixel((196 + xx, 22 + yy), (255, 240, 200))

def ridge(base, amp, col, seed, step):
    rr = random.Random(seed)
    y = base
    for x in range(w):
        if x % step == 0:
            y = max(base - amp, min(base + 6, y + rr.choice([-2, -1, -1, 0, 1, 1, 2])))
        d.line([(x, y), (x, h)], fill=col)

ridge(70, 22, (70, 40, 100), 1, 3)
ridge(82, 16, (50, 30, 80), 2, 2)
# castelo no morro
cx, cy = 60, 70
d.rectangle([cx - 14, cy - 10, cx + 14, cy + 20], fill=(34, 22, 58))
for t in (-14, 8):
    d.rectangle([cx + t, cy - 20, cx + t + 6, cy], fill=(34, 22, 58))
    for k in range(0, 7, 2):
        d.point((cx + t + k, cy - 21), fill=(34, 22, 58))
d.rectangle([cx - 3, cy - 26, cx + 3, cy - 10], fill=(34, 22, 58))
d.polygon([(cx - 4, cy - 26), (cx, cy - 33), (cx + 4, cy - 26)], fill=(34, 22, 58))
d.point((cx, cy - 37), fill=(255, 80, 120)); d.line([(cx, cy - 33), (cx, cy - 37)], fill=(200, 200, 200))
for (x, y) in [(cx - 9, cy - 4), (cx + 9, cy - 4), (cx, cy - 18), (cx - 2, cy + 6)]:
    d.rectangle([x, y, x + 1, y + 2], fill=(255, 210, 110))
ridge(96, 10, (30, 22, 50), 3, 2)
# rio
for y in range(104, 135):
    t = (y - 104) / 31
    x0 = int(110 - 30 * t - 6 * math.sin(y / 4)); x1 = int(140 + 40 * t + 6 * math.sin(y / 5))
    d.line([(x0, y), (x1, y)], fill=(40, 70, 130))
    if y % 3 == 0:
        for x in range(x0 + 2, x1 - 2, 7):
            d.point((x + (y % 5), y), fill=(150, 190, 255))
# chão
for y in range(100, 135):
    for x in range(w):
        p = img.getpixel((x, y))
        if p != (40, 70, 130) and p != (150, 190, 255) and y > 103:
            img.putpixel((x, y), (24, 44, 40) if (x + y) % 9 else (30, 56, 48))
# pinheiros
def pine(x, y, s, col):
    for i in range(s):
        d.line([(x - i // 2, y + i), (x + i // 2, y + i)], fill=col)
    for i in range(s // 2):
        d.line([(x - i // 2 - 1, y + s // 2 + i), (x + i // 2 + 1, y + s // 2 + i)], fill=col)
    d.line([(x, y + s), (x, y + s + 3)], fill=(40, 26, 20))
for _ in range(28):
    x = r.randrange(w)
    if 80 < x < 170: continue
    pine(x, r.randint(88, 112), r.randint(10, 18), r.choice([(18, 40, 34), (22, 50, 40), (14, 32, 28)]))
# herói + baú
hx, hy = 92, 112
hero = ["..rr..", ".rrrr.", "..ss..", ".bbbb.", "s.bb.s", "..bb..", ".b..b."]
pal = {"r": (220, 50, 60), "s": (250, 200, 160), "b": (60, 110, 230)}
for j, row in enumerate(hero):
    for i, c in enumerate(row):
        if c in pal:
            img.putpixel((hx + i, hy + j), pal[c])
d.rectangle([170, 116, 177, 121], fill=(150, 90, 40)); d.line([(170, 118), (177, 118)], fill=(240, 200, 60))
# fogueira
for (x, y, c) in [(80, 118, (255, 200, 60)), (81, 117, (255, 120, 40)), (79, 117, (255, 120, 40)), (80, 116, (255, 80, 40)), (80, 115, (255, 220, 120))]:
    img.putpixel((x, y), c)
d.line([(77, 119), (83, 119)], fill=(90, 60, 30))

big = img.resize((w * S, h * S), Image.NEAREST)
dd = ImageDraw.Draw(big)
# HUD em pixel
font = {
    "L": ["1..", "1..", "1..", "1..", "111"], "V": ["1.1", "1.1", "1.1", "1.1", ".1."], "9": ["111", "1.1", "111", "..1", "111"],
    "X": ["1.1", "1.1", ".1.", "1.1", "1.1"], "P": ["111", "1.1", "111", "1..", "1.."], " ": ["...", "...", "...", "...", "..."],
    "E": ["111", "1..", "111", "1..", "111"], "0": ["111", "1.1", "1.1", "1.1", "111"], "1": [".1.", "11.", ".1.", ".1.", "111"],
    "2": ["111", "..1", "111", "1..", "111"], "4": ["1.1", "1.1", "111", "..1", "..1"], "7": ["111", "..1", "..1", "..1", "..1"],
}
def ptext(x, y, s, col, px=6):
    for ch in s:
        g = font.get(ch, font[" "])
        for j, row in enumerate(g):
            for i, c in enumerate(row):
                if c == "1":
                    dd.rectangle([x + i * px, y + j * px, x + i * px + px - 1, y + j * px + px - 1], fill=col)
        x += 4 * px
ptext(48, 40, "LV 99", (255, 255, 255))
ptext(48, 84, "XP 1024", (255, 210, 110))
for i in range(5):
    x = 1700 + i * 36
    col = (230, 50, 70) if i < 4 else (90, 40, 60)
    for (a, b2) in [(1, 0), (2, 0), (4, 0), (5, 0), (0, 1), (1, 1), (2, 1), (3, 1), (4, 1), (5, 1), (6, 1), (1, 2), (2, 2), (3, 2), (4, 2), (5, 2), (2, 3), (3, 3), (4, 3), (3, 4)]:
        dd.rectangle([x + a * 4, 50 + b2 * 4, x + a * 4 + 3, 50 + b2 * 4 + 3], fill=col)
big.save(sys.argv[1])
print("pixel ok")
