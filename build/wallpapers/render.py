"""Renderiza os SVGs em JPG 1920x1080 + miniaturas 384x216."""
import asyncio, os, sys, glob
from PIL import Image
from playwright.async_api import async_playwright

SRC, OUT = sys.argv[1], sys.argv[2]
os.makedirs(OUT, exist_ok=True)

def save(png, name):
    im = Image.open(png).convert("RGB")
    im.save(os.path.join(OUT, name + ".jpg"), "JPEG", quality=84, optimize=True, progressive=True)
    im.resize((384, 216), Image.LANCZOS).save(os.path.join(OUT, name + ".thumb.jpg"), "JPEG", quality=80, optimize=True)

async def main():
    async with async_playwright() as p:
        b = await p.chromium.launch()
        pg = await b.new_page(viewport={"width": 1920, "height": 1080}, device_scale_factor=1)
        for f in sorted(glob.glob(os.path.join(SRC, "*.svg"))):
            name = os.path.basename(f)[:-4]
            await pg.goto("file://" + f); await pg.wait_for_timeout(300)
            png = os.path.join(SRC, name + ".png")
            await pg.screenshot(path=png)
            save(png, name)
            print("ok", name)
        await b.close()
    px = os.path.join(SRC, "jogos-pixel.png")
    if os.path.exists(px):
        save(px, "jogos-pixel")
asyncio.run(main())
