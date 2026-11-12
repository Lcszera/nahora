"""Gera os wallpapers do naHora (arte original, 1920x1080).
Uso: python3 gen.py <pasta_saida>   -> cria <id>.svg/.html; render.py transforma em JPG."""
import math, random, sys, os

W, H = 1920, 1080
OUT = sys.argv[1]
os.makedirs(OUT, exist_ok=True)

def svg(body, defs="", bg="#000"):
    return f'''<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}">
<defs>{defs}</defs><rect width="{W}" height="{H}" fill="{bg}"/>{body}</svg>'''

def esc(s): return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")

GLOW = lambda id, sd: f'<filter id="{id}" x="-50%" y="-50%" width="200%" height="200%"><feGaussianBlur stdDeviation="{sd}" result="b"/><feMerge><feMergeNode in="b"/><feMergeNode in="SourceGraphic"/></feMerge></filter>'
BLUR = lambda id, sd: f'<filter id="{id}" x="-50%" y="-50%" width="200%" height="200%"><feGaussianBlur stdDeviation="{sd}"/></filter>'
GRAIN = '<filter id="grain"><feTurbulence type="fractalNoise" baseFrequency="0.9" numOctaves="2" stitchTiles="stitch"/><feColorMatrix values="0 0 0 0 1  0 0 0 0 1  0 0 0 0 1  0 0 0 .07 0"/></filter>'
VIGNETTE = '<radialGradient id="vig" cx="50%" cy="50%" r="75%"><stop offset="55%" stop-color="#000" stop-opacity="0"/><stop offset="100%" stop-color="#000" stop-opacity=".65"/></radialGradient>'
def finish(grain=True, vig=True):
    s = ""
    if grain: s += f'<rect width="{W}" height="{H}" filter="url(#grain)"/>'
    if vig: s += f'<rect width="{W}" height="{H}" fill="url(#vig)"/>'
    return s

walls = {}

# ---------------------------------------------------------------- CÓDIGO 1: editor à meia-noite
def codigo_editor():
    r = random.Random(1)
    defs = GLOW("g", 8) + BLUR("bb", 60) + BLUR("bk", 4) + GRAIN + VIGNETTE + '''
    <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#0a0e27"/><stop offset="1" stop-color="#05060f"/></linearGradient>
    <linearGradient id="win" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#161b33"/><stop offset="1" stop-color="#0f1328"/></linearGradient>
    <filter id="sh" x="-20%" y="-20%" width="140%" height="160%"><feDropShadow dx="0" dy="30" stdDeviation="40" flood-color="#000" flood-opacity=".7"/></filter>'''
    b = f'<rect width="{W}" height="{H}" fill="url(#bg)"/>'
    b += '<circle cx="1500" cy="200" r="380" fill="#3d2bff" opacity=".35" filter="url(#bb)"/>'
    b += '<circle cx="300" cy="900" r="420" fill="#00b3ff" opacity=".18" filter="url(#bb)"/>'
    b += '<circle cx="1100" cy="950" r="300" fill="#ff3d9a" opacity=".12" filter="url(#bb)"/>'
    for _ in range(60):
        x, y, rr = r.uniform(0, W), r.uniform(0, H), r.uniform(1, 7)
        b += f'<circle cx="{x:.0f}" cy="{y:.0f}" r="{rr:.1f}" fill="#9ecbff" opacity="{r.uniform(.05,.35):.2f}" filter="url(#bk)"/>'
    x0, y0, ww, hh = 760, 170, 1020, 700
    b += f'<g filter="url(#sh)"><rect x="{x0}" y="{y0}" width="{ww}" height="{hh}" rx="18" fill="url(#win)" stroke="#2a3366" stroke-width="1.5"/></g>'
    b += f'<rect x="{x0}" y="{y0}" width="{ww}" height="46" rx="18" fill="#1b2142"/><rect x="{x0}" y="{y0+28}" width="{ww}" height="18" fill="#1b2142"/>'
    for i, c in enumerate(["#ff5f57", "#febc2e", "#28c840"]):
        b += f'<circle cx="{x0+28+i*24}" cy="{y0+23}" r="7" fill="{c}"/>'
    b += f'<rect x="{x0+110}" y="{y0+8}" width="170" height="38" rx="8" fill="#0f1328"/><text x="{x0+132}" y="{y0+33}" font-family="DejaVu Sans Mono" font-size="15" fill="#c8d3ff">☕ Tarefa.java</text>'
    K, T, S, F, C, N, P = "#c792ea", "#ffcb6b", "#c3e88d", "#82aaff", "#546e7a", "#f78c6c", "#89ddff"
    code = [
        [(C, "// naHora: cada coisa no seu horário")],
        [(K, "public record "), (T, "Tarefa"), (P, "("), (T, "String "), ("#eeffff", "nome"), (P, ", "), (T, "LocalDateTime "), ("#eeffff", "entrega"), (P, ") {")],
        [],
        [(K, "    public boolean "), (F, "atrasada"), (P, "() {")],
        [(K, "        return "), (T, "LocalDateTime"), (P, "."), (F, "now"), (P, "()."), (F, "isAfter"), (P, "("), ("#eeffff", "entrega"), (P, ");")],
        [(P, "    }")],
        [(P, "}")],
        [],
        [(T, "List"), (P, "<"), (T, "Tarefa"), (P, "> "), ("#eeffff", "hoje"), (P, " = "), ("#eeffff", "tarefas"), (P, "."), (F, "stream"), (P, "()")],
        [(P, "        ."), (F, "filter"), (P, "("), ("#eeffff", "t"), (K, " -> "), (P, "!"), ("#eeffff", "t"), (P, "."), (F, "atrasada"), (P, "())")],
        [(P, "        ."), (F, "sorted"), (P, "("), (T, "Comparator"), (P, "."), (F, "comparing"), (P, "("), (T, "Tarefa"), (P, "::"), (F, "entrega"), (P, "))")],
        [(P, "        ."), (F, "toList"), (P, "();")],
        [],
        [(K, "if "), (P, "("), ("#eeffff", "hoje"), (P, "."), (F, "isEmpty"), (P, "()) "), (T, "System"), (P, "."), ("#eeffff", "out"), (P, "."), (F, "println"), (P, "("), (S, '"Tudo feito. Bora estudar!"'), (P, ");")],
        [],
        [(K, "var "), ("#eeffff", "foco"), (P, " = "), (T, "Duration"), (P, "."), (F, "ofMinutes"), (P, "("), (N, "50"), (P, ");  "), (C, "// sem procrastinar")],
    ]
    ly = y0 + 100
    for i, line in enumerate(code):
        b += f'<text x="{x0+30}" y="{ly}" font-family="DejaVu Sans Mono" font-size="19" fill="#3b4675" text-anchor="start">{i+1:>2}</text>'
        xx = x0 + 80
        t = f'<text x="{xx}" y="{ly}" font-family="DejaVu Sans Mono" font-size="19" xml:space="preserve">'
        for col, txt in line:
            t += f'<tspan fill="{col}">{esc(txt)}</tspan>'
        b += t + '</text>'
        ly += 34
    b += f'<rect x="{x0+80+11.4*37}" y="{ly-34-21}" width="3" height="26" fill="#82aaff" filter="url(#g)"/>'
    b += f'<rect x="{x0}" y="{y0+hh-34}" width="{ww}" height="34" rx="0" fill="#3d8bff" opacity=".9"/><rect x="{x0}" y="{y0+hh-18}" width="{ww}" height="18" rx="18" fill="#3d8bff" opacity=".9"/>'
    b += f'<text x="{x0+20}" y="{y0+hh-12}" font-family="DejaVu Sans Mono" font-size="14" fill="#fff">⎇ main   ✓ 0 erros   Java 21</text>'
    b += f'<text x="140" y="560" font-family="Poppins" font-weight="800" font-style="italic" font-size="120" fill="#fff" opacity=".07">{esc("{ }")}</text>'
    b += finish()
    return svg(b, defs)
walls["codigo-editor"] = codigo_editor

# ---------------------------------------------------------------- CÓDIGO 2: chuva de código
def codigo_terminal():
    r = random.Random(2)
    defs = GLOW("g", 6) + BLUR("bb", 80) + GRAIN + VIGNETTE + '''
    <linearGradient id="fade" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#000" stop-opacity=".95"/><stop offset=".35" stop-color="#000" stop-opacity="0"/><stop offset=".8" stop-color="#000" stop-opacity="0"/><stop offset="1" stop-color="#000" stop-opacity=".9"/></linearGradient>'''
    b = '<rect width="1920" height="1080" fill="#020806"/>'
    b += '<ellipse cx="960" cy="540" rx="700" ry="380" fill="#00ff9c" opacity=".09" filter="url(#bb)"/>'
    chars = list("01{}();=<>+*/[]&|!?:.") + ["int", "if", "for", "new", "var", "->", "::", "null", "true"]
    cols = 68
    for c in range(cols):
        x = c * (W / cols) + 6
        head = r.uniform(-200, H + 300)
        length = r.randint(10, 30)
        size = r.choice([16, 18, 20, 22])
        step = size * 1.25
        for k in range(length):
            y = head - k * step
            if y < -20 or y > H + 20: continue
            a = max(0.04, 1 - k / length) * r.uniform(.5, 1)
            ch = r.choice(chars[:21]) if r.random() < .85 else r.choice(chars[21:])
            col = "#d6ffe9" if k == 0 else "#22e88a"
            flt = ' filter="url(#g)"' if k == 0 else ""
            b += f'<text x="{x:.0f}" y="{y:.0f}" font-family="DejaVu Sans Mono" font-size="{size}" fill="{col}" opacity="{a:.2f}"{flt}>{esc(ch)}</text>'
    b += '<rect width="1920" height="1080" fill="url(#fade)"/>'
    b += '<text x="960" y="600" text-anchor="middle" font-family="DejaVu Sans Mono" font-weight="bold" font-size="210" fill="#00ff9c" opacity=".16" filter="url(#g)">&lt;/&gt;</text>'
    b += '<text x="70" y="1020" font-family="DejaVu Sans Mono" font-size="22" fill="#5dffb5" opacity=".75">lucas@nahora:~$ <tspan fill="#d6ffe9">mvn spring-boot:run</tspan><tspan fill="#00ff9c">▌</tspan></text>'
    b += finish()
    return svg(b, defs)
walls["codigo-terminal"] = codigo_terminal

# ---------------------------------------------------------------- ELETRÔNICA 1: placa de circuito
def eletronica_pcb():
    r = random.Random(3)
    defs = GLOW("g", 5) + BLUR("bb", 70) + GRAIN + VIGNETTE + '''
    <radialGradient id="bg" cx="50%" cy="50%" r="70%"><stop offset="0" stop-color="#0c3b33"/><stop offset="1" stop-color="#03130f"/></radialGradient>
    <pattern id="dots" width="24" height="24" patternUnits="userSpaceOnUse"><circle cx="12" cy="12" r="1.2" fill="#1f6b5a" opacity=".5"/></pattern>'''
    b = '<rect width="1920" height="1080" fill="url(#bg)"/><rect width="1920" height="1080" fill="url(#dots)"/>'
    b += '<circle cx="960" cy="540" r="420" fill="#2fffc0" opacity=".08" filter="url(#bb)"/>'
    cx, cy, cw, ch = 960, 540, 300, 300
    pins = []
    for side in range(4):
        for i in range(12):
            t = -cw / 2 + 25 + i * (cw - 50) / 11
            if side == 0: pins.append((cx + t, cy - ch / 2, 0, -1))
            if side == 1: pins.append((cx + cw / 2, cy + t, 1, 0))
            if side == 2: pins.append((cx + t, cy + ch / 2, 0, 1))
            if side == 3: pins.append((cx - cw / 2, cy + t, -1, 0))
    traces = ""
    for (px, py, dx, dy) in pins:
        x, y = px + dx * 18, py + dy * 18
        pts = [(px, py), (x, y)]
        L = r.uniform(80, 260)
        x, y = x + dx * L, y + dy * L; pts.append((x, y))
        # dobra 45°
        turn = r.choice([-1, 1])
        d2 = r.uniform(40, 160)
        if dx == 0: x, y = x + turn * d2, y + dy * d2
        else: x, y = x + dx * d2, y + turn * d2
        pts.append((x, y))
        L2 = r.uniform(150, 700)
        x, y = x + dx * L2, y + dy * L2; pts.append((x, y))
        col = r.choice(["#e8b34a", "#e8b34a", "#f6cf73", "#2fffc0"])
        wdt = r.choice([3, 3, 4, 5])
        d = "M" + " L".join(f"{a:.0f},{bb:.0f}" for a, bb in pts)
        glow = ' filter="url(#g)"' if col == "#2fffc0" else ""
        traces += f'<path d="{d}" fill="none" stroke="{col}" stroke-width="{wdt}" stroke-linecap="round" stroke-linejoin="round" opacity=".85"{glow}/>'
        traces += f'<circle cx="{x:.0f}" cy="{y:.0f}" r="{wdt+5}" fill="#03130f" stroke="{col}" stroke-width="3"{glow}/>'
    b += traces
    for _ in range(26):
        x, y = r.uniform(80, W - 80), r.uniform(80, H - 80)
        if abs(x - cx) < 260 and abs(y - cy) < 260: continue
        rot = r.choice([0, 90])
        b += f'<g transform="translate({x:.0f},{y:.0f}) rotate({rot})"><rect x="-22" y="-9" width="44" height="18" rx="3" fill="#1d1d1d"/><rect x="-22" y="-9" width="9" height="18" fill="#c9c9c9"/><rect x="13" y="-9" width="9" height="18" fill="#c9c9c9"/></g>'
    for (px, py, dx, dy) in pins:
        b += f'<rect x="{px - 5 + dx*8:.0f}" y="{py - 5 + dy*8:.0f}" width="10" height="10" fill="#d9d9d9"/>'
    b += f'<rect x="{cx-cw/2}" y="{cy-ch/2}" width="{cw}" height="{ch}" rx="10" fill="#141414" stroke="#2c2c2c" stroke-width="3"/>'
    b += f'<circle cx="{cx-cw/2+28}" cy="{cy-ch/2+28}" r="8" fill="#2a2a2a"/>'
    b += f'<text x="{cx}" y="{cy-6}" text-anchor="middle" font-family="DejaVu Sans Mono" font-weight="bold" font-size="34" fill="#9a9a9a">NH-1100</text>'
    b += f'<text x="{cx}" y="{cy+34}" text-anchor="middle" font-family="DejaVu Sans Mono" font-size="20" fill="#6d6d6d">⚡ naHora MCU</text>'
    b += '<text x="60" y="1040" font-family="DejaVu Sans Mono" font-size="18" fill="#2fffc0" opacity=".5">R1 10kΩ   C3 100nF   U1   VCC 3V3   GND</text>'
    b += finish()
    return svg(b, defs)
walls["eletronica-pcb"] = eletronica_pcb

# ---------------------------------------------------------------- ELETRÔNICA 2: osciloscópio
def eletronica_osc():
    defs = GLOW("g", 7) + GLOW("g2", 16) + BLUR("bb", 90) + GRAIN + VIGNETTE + '''
    <radialGradient id="crt" cx="50%" cy="50%" r="65%"><stop offset="0" stop-color="#0b2a2e"/><stop offset="1" stop-color="#020809"/></radialGradient>
    <pattern id="scan" width="4" height="4" patternUnits="userSpaceOnUse"><rect width="4" height="2" fill="#000" opacity=".22"/></pattern>'''
    b = '<rect width="1920" height="1080" fill="url(#crt)"/>'
    gx, gy, gw, gh = 160, 120, 1600, 840
    for i in range(11):
        x = gx + i * gw / 10
        b += f'<line x1="{x}" y1="{gy}" x2="{x}" y2="{gy+gh}" stroke="#2bd1c4" stroke-width="{2 if i==5 else 1}" opacity="{.45 if i==5 else .18}"/>'
    for j in range(9):
        y = gy + j * gh / 8
        b += f'<line x1="{gx}" y1="{y}" x2="{gx+gw}" y2="{y}" stroke="#2bd1c4" stroke-width="{2 if j==4 else 1}" opacity="{.45 if j==4 else .18}"/>'
    for i in range(51):
        x = gx + i * gw / 50
        b += f'<line x1="{x}" y1="{gy+gh/2-6}" x2="{x}" y2="{gy+gh/2+6}" stroke="#2bd1c4" opacity=".35"/>'
    def wave(f, color, filt, w=4):
        pts = []
        for i in range(0, 1601, 4):
            pts.append((gx + i, f(i / 1600)))
        d = "M" + " L".join(f"{x:.1f},{y:.1f}" for x, y in pts)
        return f'<path d="{d}" fill="none" stroke="{color}" stroke-width="{w}" filter="url(#{filt})" stroke-linejoin="round"/>'
    mid = gy + gh / 2
    b += wave(lambda t: mid - 200 * math.sin(t * 2 * math.pi * 3) * (0.6 + 0.4 * math.cos(t * 2 * math.pi * 0.5)), "#36ffd2", "g")
    sq = lambda t: mid + 230 + (-70 if math.sin(t * 2 * math.pi * 6) > 0 else 70) * 0.9
    pts = []
    prev = None
    for i in range(0, 1601, 2):
        y = sq(i / 1600)
        if prev is not None and y != prev: pts.append((gx + i, prev))
        pts.append((gx + i, y)); prev = y
    b += '<path d="M' + " L".join(f"{x},{y:.0f}" for x, y in pts) + '" fill="none" stroke="#ffd34e" stroke-width="3.5" filter="url(#g)" opacity=".9"/>'
    b += wave(lambda t: mid - 280 + 80 * (2 * abs(2 * ((t * 4) % 1) - 1) - 1), "#ff6bd6", "g", 3)
    b += '<rect width="1920" height="1080" fill="url(#scan)"/>'
    for i, (txt, col) in enumerate([("CH1  2.00V/div", "#36ffd2"), ("CH2  5.00V/div", "#ffd34e"), ("CH3  1.00V/div", "#ff6bd6")]):
        b += f'<text x="{gx+10+i*330}" y="{gy+gh+60}" font-family="DejaVu Sans Mono" font-size="26" fill="{col}" filter="url(#g)">{txt}</text>'
    b += f'<text x="{gx+gw}" y="{gy+gh+60}" text-anchor="end" font-family="DejaVu Sans Mono" font-size="26" fill="#9ffff0" opacity=".8">5.00ms/div   f = 60.0Hz</text>'
    b += f'<text x="{gx+10}" y="{gy-30}" font-family="DejaVu Sans Mono" font-size="24" fill="#9ffff0" opacity=".7">RUN ▶  TRIG ↗ 1.2V</text>'
    b += f'<rect x="{gx}" y="{gy}" width="{gw}" height="{gh}" fill="none" stroke="#2bd1c4" stroke-width="2" opacity=".5" rx="6"/>'
    b += finish()
    return svg(b, defs)
walls["eletronica-osc"] = eletronica_osc

# ---------------------------------------------------------------- CÁLCULOS 1: quadro-negro
def calculo_quadro():
    r = random.Random(5)
    defs = GRAIN + VIGNETTE + '''
    <filter id="board"><feTurbulence type="fractalNoise" baseFrequency=".012" numOctaves="4" seed="3"/><feColorMatrix values="0 0 0 0 .16  0 0 0 0 .25  0 0 0 0 .21  0 0 0 .55 0"/></filter>
    <filter id="chalk" x="-5%" y="-5%" width="110%" height="110%"><feTurbulence type="fractalNoise" baseFrequency="1.2" numOctaves="2" seed="7" result="n"/>
      <feColorMatrix in="n" type="matrix" values="0 0 0 0 1  0 0 0 0 1  0 0 0 0 1  0 0 0 -1.6 1.25" result="m"/>
      <feComposite in="SourceGraphic" in2="m" operator="in" result="c"/><feDisplacementMap in="c" in2="n" scale="2.2"/></filter>
    <filter id="smudge"><feGaussianBlur stdDeviation="28"/></filter>
    <linearGradient id="wood" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#7a4b2a"/><stop offset="1" stop-color="#4a2c17"/></linearGradient>'''
    b = '<rect width="1920" height="1080" fill="#22312b"/><rect width="1920" height="1080" filter="url(#board)"/>'
    for _ in range(9):
        b += f'<ellipse cx="{r.uniform(100,1800):.0f}" cy="{r.uniform(100,950):.0f}" rx="{r.uniform(120,320):.0f}" ry="{r.uniform(40,110):.0f}" fill="#cfd8d3" opacity=".06" filter="url(#smudge)" transform="rotate({r.uniform(-15,15):.0f} 960 540)"/>'
    g = '<g filter="url(#chalk)" fill="#f4f1e8" font-family="Latin Modern Roman" font-style="italic">'
    def T(x, y, s, txt, col="#f4f1e8", it=True, extra=""):
        st = "" if it else ' font-style="normal"'
        return f'<text x="{x}" y="{y}" font-size="{s}" fill="{col}"{st} {extra}>{txt}</text>'
    g += T(110, 190, 66, '∫<tspan font-size="40" dy="22">0</tspan><tspan font-size="40" dy="-62" dx="-18">∞</tspan><tspan dy="40"> e</tspan><tspan font-size="40" dy="-30">−x²</tspan><tspan dy="30"> dx = </tspan><tspan>√π / 2</tspan>')
    g += T(120, 340, 60, 'd/dx (sen x) = cos x')
    g += T(120, 470, 64, 'e<tspan font-size="40" dy="-30">iπ</tspan><tspan dy="30"> + 1 = 0</tspan>', "#ffe9a8")
    g += T(120, 610, 56, 'lim<tspan font-size="30" dy="22" dx="-90">x→0</tspan><tspan dy="-22" dx="20"> sen x / x = 1</tspan>')
    g += T(120, 760, 52, 'f(x) = Σ f<tspan font-size="30" dy="-26">(n)</tspan><tspan dy="26">(a) (x−a)</tspan><tspan font-size="30" dy="-26">n</tspan><tspan dy="26"> / n!</tspan>', "#bfe6ff")
    g += T(120, 900, 52, 'A·x = b  ⇒  x = A<tspan font-size="32" dy="-26">−1</tspan><tspan dy="26">b</tspan>')
    # gráfico
    ox, oy = 1280, 640
    g += f'<path d="M{ox-330},{oy} L{ox+430},{oy} M{ox},{oy+300} L{ox},{oy-470}" stroke="#f4f1e8" stroke-width="5" fill="none" stroke-linecap="round"/>'
    g += f'<path d="M{ox+430},{oy} l-22,-12 M{ox+430},{oy} l-22,12 M{ox},{oy-470} l-12,22 M{ox},{oy-470} l12,22" stroke="#f4f1e8" stroke-width="5" fill="none" stroke-linecap="round"/>'
    pts = " L".join(f"{ox + x*60:.0f},{oy - (x*x)*30 + 60:.0f}" for x in [i/10 for i in range(-35, 36)])
    g += f'<path d="M{pts}" stroke="#ffe9a8" stroke-width="6" fill="none" stroke-linecap="round"/>'
    pts = " L".join(f"{ox + x*60:.0f},{oy - 120*math.sin(x*1.2):.0f}" for x in [i/10 for i in range(-50, 70)])
    g += f'<path d="M{pts}" stroke="#bfe6ff" stroke-width="5" fill="none" stroke-linecap="round"/>'
    g += T(ox + 230, oy - 420, 46, 'y = x² − 1', "#ffe9a8")
    g += T(ox + 290, oy + 130, 42, 'y = 2 sen(1,2x)', "#bfe6ff")
    g += T(ox + 440, oy + 14, 40, 'x') + T(ox - 40, oy - 470, 40, 'y')
    g += f'<path d="M1060,860 q320,-40 640,0" stroke="#f4f1e8" stroke-width="4" fill="none" opacity=".7"/>'
    g += T(1080, 940, 44, '∴  estudar todo dia ≈ ∞ evolução', "#f4f1e8")
    g += '</g>'
    b += g
    b += '<rect x="0" y="1040" width="1920" height="40" fill="url(#wood)"/><rect x="0" y="1036" width="1920" height="6" fill="#9a6a43"/>'
    b += '<rect x="1500" y="1022" width="70" height="16" rx="7" fill="#f4f1e8"/><rect x="1600" y="1024" width="46" height="14" rx="6" fill="#ffe9a8"/>'
    b += '<rect x="300" y="1014" width="140" height="26" rx="5" fill="#3a3a3a"/><rect x="300" y="1014" width="140" height="10" rx="4" fill="#d9c9a8"/>'
    b += finish(vig=True)
    return svg(b, defs)
walls["calculo-quadro"] = calculo_quadro

# ---------------------------------------------------------------- CÁLCULOS 2: blueprint
def calculo_blueprint():
    defs = GRAIN + VIGNETTE + '''
    <pattern id="g1" width="24" height="24" patternUnits="userSpaceOnUse"><path d="M24 0H0V24" fill="none" stroke="#9fd0ff" stroke-width=".6" opacity=".25"/></pattern>
    <pattern id="g2" width="120" height="120" patternUnits="userSpaceOnUse"><path d="M120 0H0V120" fill="none" stroke="#bfe0ff" stroke-width="1.4" opacity=".35"/></pattern>
    <radialGradient id="bg" cx="40%" cy="40%" r="80%"><stop offset="0" stop-color="#1b5fb8"/><stop offset="1" stop-color="#0b2f66"/></radialGradient>'''
    b = '<rect width="1920" height="1080" fill="url(#bg)"/><rect width="1920" height="1080" fill="url(#g1)"/><rect width="1920" height="1080" fill="url(#g2)"/>'
    L = 'stroke="#eaf5ff" fill="none" stroke-linecap="round"'
    F = 'font-family="DejaVu Sans Mono" fill="#eaf5ff"'
    M = 'font-family="Latin Modern Roman" font-style="italic"'
    # círculo trigonométrico
    cx, cy, R = 480, 540, 300
    b += f'<circle cx="{cx}" cy="{cy}" r="{R}" {L} stroke-width="3"/>'
    b += f'<path d="M{cx-380},{cy} H{cx+380} M{cx},{cy+380} V{cy-380}" {L} stroke-width="2" opacity=".8"/>'
    a = math.radians(40)
    px, py = cx + R*math.cos(a), cy - R*math.sin(a)
    b += f'<path d="M{cx},{cy} L{px:.0f},{py:.0f}" {L} stroke-width="3"/>'
    b += f'<path d="M{px:.0f},{py:.0f} V{cy}" stroke="#ffd36b" fill="none" stroke-width="2.5" stroke-dasharray="10 8"/>'
    b += f'<path d="M{cx},{cy} H{px:.0f}" stroke="#7dffcf" stroke-width="5" fill="none"/>'
    b += f'<path d="M{px:.0f},{py:.0f} V{cy}" stroke="#ffd36b" stroke-width="5" fill="none"/>'
    b += f'<path d="M{cx+70},{cy} A70,70 0 0 0 {cx+70*math.cos(a):.0f},{cy-70*math.sin(a):.0f}" {L} stroke-width="2.5"/>'
    b += f'<circle cx="{px:.0f}" cy="{py:.0f}" r="9" fill="#eaf5ff"/>'
    b += f'<text x="{cx+86}" y="{cy-22}" font-size="34" {M} fill="#eaf5ff">θ</text>'
    b += f'<text x="{cx+60}" y="{cy+48}" font-size="30" {M} fill="#7dffcf">cos θ</text>'
    b += f'<text x="{px+16:.0f}" y="{cy-90}" font-size="30" {M} fill="#ffd36b">sen θ</text>'
    b += f'<text x="{cx-180}" y="{cy+430}" font-size="36" {M} fill="#eaf5ff">sen²θ + cos²θ = 1</text>'
    # gráfico de funções
    ox, oy = 1180, 560
    b += f'<path d="M{ox-60},{oy} H{ox+660} M{ox},{oy+300} V{oy-330}" {L} stroke-width="2.5"/>'
    for i in range(1, 11):
        b += f'<path d="M{ox+i*60},{oy-8} V{oy+8}" {L} stroke-width="2"/>'
    pts = " L".join(f"{ox + x:.0f},{oy - 200*math.sin(x/60):.0f}" for x in range(0, 640, 6))
    b += f'<path d="M{pts}" stroke="#7dffcf" stroke-width="4" fill="none"/>'
    pts = " L".join(f"{ox + x:.0f},{oy - 260*(1-math.exp(-x/160)):.0f}" for x in range(0, 640, 6))
    b += f'<path d="M{pts}" stroke="#ffd36b" stroke-width="4" fill="none"/>'
    b += f'<text x="{ox+330}" y="{oy-330}" font-size="32" {M} fill="#ffd36b">1 − e^(−t/τ)</text>'
    b += f'<text x="{ox+500}" y="{oy+190}" font-size="32" {M} fill="#7dffcf">sen(t)</text>'
    # triângulo
    tx, ty = 1240, 960
    b += f'<path d="M{tx},{ty} h360 v-200 z" {L} stroke-width="3"/><path d="M{tx+330},{ty} v-30 h30" {L} stroke-width="2"/>'
    b += f'<text x="{tx+150}" y="{ty+42}" font-size="30" {M} fill="#eaf5ff">a</text><text x="{tx+375}" y="{ty-90}" font-size="30" {M} fill="#eaf5ff">b</text><text x="{tx+140}" y="{ty-120}" font-size="30" {M} fill="#eaf5ff">c</text>'
    b += f'<text x="{tx+430}" y="{ty-10}" font-size="34" {M} fill="#eaf5ff">c² = a² + b²</text>'
    # carimbo
    b += f'<rect x="1500" y="40" width="380" height="120" {L} stroke-width="2"/><path d="M1500,80 H1880 M1700,80 V160" {L} stroke-width="1.5"/>'
    b += f'<text x="1516" y="70" font-size="20" {F}>PROJETO: naHora</text><text x="1516" y="120" font-size="18" {F}>FOLHA 01/01</text><text x="1716" y="120" font-size="18" {F}>ESC 1:1</text><text x="1516" y="148" font-size="16" {F} opacity=".8">CÁLCULO · REV A</text>'
    b += finish()
    return svg(b, defs)
walls["calculo-blueprint"] = calculo_blueprint

# ---------------------------------------------------------------- LO-FI 1: janela à noite
def lofi_janela():
    r = random.Random(7)
    defs = GLOW("g", 6) + BLUR("bb", 50) + BLUR("b8", 8) + GRAIN + VIGNETTE + '''
    <linearGradient id="sky" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#1a1240"/><stop offset=".55" stop-color="#5b2a7a"/><stop offset="1" stop-color="#ff7aa8"/></linearGradient>
    <linearGradient id="wall" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="#1b1430"/><stop offset="1" stop-color="#241a3d"/></linearGradient>
    <clipPath id="win"><rect x="460" y="120" width="1000" height="700" rx="8"/></clipPath>'''
    b = '<rect width="1920" height="1080" fill="url(#wall)"/>'
    w = '<g clip-path="url(#win)"><rect x="460" y="120" width="1000" height="700" fill="url(#sky)"/>'
    for _ in range(90):
        w += f'<circle cx="{r.uniform(460,1460):.0f}" cy="{r.uniform(120,520):.0f}" r="{r.uniform(.8,2.4):.1f}" fill="#fff" opacity="{r.uniform(.3,.9):.2f}"/>'
    w += '<circle cx="1250" cy="260" r="70" fill="#fff4d6" filter="url(#g)"/><circle cx="1282" cy="240" r="62" fill="#3a1f60" opacity=".95"/>'
    for layer, (base, col, lit) in enumerate([(560, "#3b2563", .25), (620, "#2a1a4a", .45), (690, "#170f2c", .7)]):
        x = 440
        while x < 1480:
            bw = r.randint(50, 130); bh = r.randint(60, 300 - layer * 60)
            w += f'<rect x="{x}" y="{base+180-bh}" width="{bw}" height="{bh+400}" fill="{col}"/>'
            for wy in range(base + 200 - bh, 840, 22):
                for wx in range(x + 8, x + bw - 10, 16):
                    if r.random() < lit * .45:
                        c = r.choice(["#ffd27a", "#ffe9b0", "#ff9bd2", "#9be7ff"])
                        w += f'<rect x="{wx}" y="{wy}" width="7" height="10" fill="{c}" opacity="{r.uniform(.5,1):.2f}"/>'
            x += bw + r.randint(0, 12)
    for _ in range(120):
        x, y = r.uniform(460, 1460), r.uniform(120, 820)
        w += f'<path d="M{x:.0f},{y:.0f} l-8,30" stroke="#cfe3ff" stroke-width="1.3" opacity="{r.uniform(.12,.35):.2f}"/>'
    w += '</g>'
    b += w
    b += '<rect x="460" y="120" width="1000" height="700" rx="8" fill="none" stroke="#120c22" stroke-width="28"/>'
    b += '<path d="M960,120 V820 M460,470 H1460" stroke="#120c22" stroke-width="20"/>'
    b += '<rect x="400" y="820" width="1120" height="34" rx="6" fill="#120c22"/><rect x="400" y="852" width="1120" height="12" fill="#0b0716"/>'
    # cortina
    b += '<path d="M330,90 C380,300 300,600 360,900 L460,900 C420,620 480,320 430,90 Z" fill="#3a2161" opacity=".95"/>'
    b += '<path d="M1590,90 C1540,300 1620,600 1560,900 L1460,900 C1500,620 1440,320 1490,90 Z" fill="#3a2161" opacity=".95"/>'
    b += '<rect x="300" y="78" width="1320" height="16" rx="8" fill="#0f0a1d"/>'
    # planta
    b += '<path d="M560,820 l-20,-90 h120 l-20,90z" fill="#1a1029"/>'
    for i in range(9):
        ang = -150 + i * 15
        x2 = 600 + 140 * math.cos(math.radians(ang)); y2 = 730 + 140 * math.sin(math.radians(ang))
        b += f'<path d="M600,735 Q{(600+x2)/2-10:.0f},{(735+y2)/2-40:.0f} {x2:.0f},{y2:.0f}" stroke="#1d1530" stroke-width="22" fill="none" stroke-linecap="round"/>'
    # caneca + vapor
    b += '<rect x="1240" y="740" width="90" height="80" rx="12" fill="#1a1029"/><path d="M1330,760 q40,0 40,28 q0,28 -40,28" stroke="#1a1029" stroke-width="12" fill="none"/>'
    for i in range(3):
        b += f'<path d="M{1262+i*22},725 q-14,-30 0,-55 q14,-25 0,-55" stroke="#fff" stroke-width="5" fill="none" opacity=".28" filter="url(#b8)" stroke-linecap="round"/>'
    b += '<ellipse cx="960" cy="1060" rx="900" ry="120" fill="#ff7aa8" opacity=".09" filter="url(#bb)"/>'
    b += finish()
    return svg(b, defs)
walls["lofi-janela"] = lofi_janela

# ---------------------------------------------------------------- LO-FI 2: fim de tarde
def lofi_fimdetarde():
    r = random.Random(8)
    defs = BLUR("bb", 40) + GRAIN + VIGNETTE + '''
    <linearGradient id="sky" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#2c2a5a"/><stop offset=".45" stop-color="#b56a9c"/><stop offset=".75" stop-color="#f6a37d"/><stop offset="1" stop-color="#ffd6a0"/></linearGradient>
    <radialGradient id="sun" cx="50%" cy="50%" r="50%"><stop offset="0" stop-color="#fff2c9"/><stop offset="1" stop-color="#ffb37a"/></radialGradient>'''
    b = '<rect width="1920" height="1080" fill="url(#sky)"/>'
    b += '<circle cx="1180" cy="400" r="320" fill="#ffd9a0" opacity=".45" filter="url(#bb)"/><circle cx="1180" cy="420" r="150" fill="url(#sun)"/>'
    for _ in range(7):
        y = r.uniform(120, 380); x = r.uniform(-200, 1700); wdt = r.uniform(250, 600)
        b += f'<rect x="{x:.0f}" y="{y:.0f}" width="{wdt:.0f}" height="{r.uniform(10,22):.0f}" rx="12" fill="#ffe3e0" opacity="{r.uniform(.12,.3):.2f}"/>'
    cols = ["#9a5a8c", "#7a4479", "#5a3266", "#3d2452", "#26183a"]
    for i, c in enumerate(cols):
        base = 600 + i * 95
        pts = [(0, base)]
        x = 0
        rr = random.Random(20 + i)
        while x < W + 100:
            x += rr.uniform(80, 220)
            pts.append((x, base - rr.uniform(20, 200 - i * 25)))
        d = f"M0,{H} L" + " L".join(f"{a:.0f},{bb:.0f}" for a, bb in pts) + f" L{W},{H} Z"
        b += f'<path d="{d}" fill="{c}" stroke="{c}" stroke-width="20" stroke-linejoin="round"/>'
        if i == 3:
            for k in range(26):
                tx = rr.uniform(0, W); ty = base - rr.uniform(-40, 40)
                hh = rr.uniform(50, 110)
                b += f'<path d="M{tx:.0f},{ty-hh:.0f} l{hh*.28:.0f},{hh:.0f} h-{hh*.56:.0f} z" fill="#2e1d45"/>'
    for i in range(5):
        x, y = 600 + i * 55 + r.uniform(-10, 10), 330 + r.uniform(-30, 30)
        s = r.uniform(.7, 1.2)
        b += f'<path d="M{x},{y} q{10*s},{-10*s} {20*s},0 q{10*s},{-10*s} {20*s},0" stroke="#2a1838" stroke-width="3" fill="none"/>'
    b += finish()
    return svg(b, defs)
walls["lofi-fimdetarde"] = lofi_fimdetarde

# ---------------------------------------------------------------- ESTUDOS 1: mesa de estudos
def estudos_mesa():
    r = random.Random(9)
    defs = GLOW("g", 10) + BLUR("bb", 70) + BLUR("b8", 7) + GRAIN + VIGNETTE + '''
    <radialGradient id="cone" cx="38%" cy="22%" r="70%"><stop offset="0" stop-color="#ffd58a" stop-opacity=".55"/><stop offset=".5" stop-color="#ff9f43" stop-opacity=".12"/><stop offset="1" stop-color="#000" stop-opacity="0"/></radialGradient>
    <linearGradient id="desk" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#5a3a24"/><stop offset="1" stop-color="#2b1b10"/></linearGradient>
    <linearGradient id="wall" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#16182c"/><stop offset="1" stop-color="#2a2236"/></linearGradient>'''
    b = '<rect width="1920" height="1080" fill="url(#wall)"/>'
    # janela pequena
    b += '<rect x="1260" y="120" width="460" height="360" rx="6" fill="#0d1b3a"/>'
    for _ in range(40):
        b += f'<circle cx="{r.uniform(1270,1710):.0f}" cy="{r.uniform(130,470):.0f}" r="{r.uniform(.8,2):.1f}" fill="#fff" opacity="{r.uniform(.3,.9):.2f}"/>'
    b += '<circle cx="1610" cy="210" r="34" fill="#fff1c9" filter="url(#g)"/>'
    b += '<rect x="1260" y="120" width="460" height="360" rx="6" fill="none" stroke="#0e0d18" stroke-width="18"/><path d="M1490,120 V480 M1260,300 H1720" stroke="#0e0d18" stroke-width="12"/>'
    # quadro de cortiça com post-its
    b += '<rect x="180" y="150" width="520" height="330" rx="10" fill="#7a5532"/><rect x="196" y="166" width="488" height="298" rx="6" fill="#a3784a"/>'
    for (x, y, c, t, rot) in [(230, 200, "#ffe066", "JPA", -6), (360, 190, "#8ce99a", "REST", 4), (490, 210, "#74c0fc", "SQL", -3), (260, 330, "#ffa8a8", "Prova sex.", 3), (430, 330, "#ffe066", "Git push!", -5)]:
        b += f'<g transform="rotate({rot} {x+55} {y+50})"><rect x="{x}" y="{y}" width="110" height="100" fill="{c}"/><text x="{x+55}" y="{y+58}" text-anchor="middle" font-family="Poppins" font-weight="700" font-size="20" fill="#3b2a1a">{t}</text><circle cx="{x+55}" cy="{y+10}" r="6" fill="#d9480f"/></g>'
    b += '<rect width="1920" height="1080" fill="url(#cone)"/>'
    # mesa
    b += '<rect x="0" y="760" width="1920" height="320" fill="url(#desk)"/><rect x="0" y="752" width="1920" height="14" fill="#6e4a30"/>'
    # luminária
    b += '<ellipse cx="560" cy="770" rx="120" ry="18" fill="#111"/><path d="M560,770 L520,560 L720,400" stroke="#1b1b22" stroke-width="18" fill="none" stroke-linecap="round" stroke-linejoin="round"/>'
    b += '<path d="M690,360 L830,420 L770,520 Z" fill="#23232c"/><ellipse cx="800" cy="470" rx="26" ry="18" fill="#fff3c4" filter="url(#g)"/>'
    b += '<ellipse cx="960" cy="800" rx="520" ry="70" fill="#ffcf7a" opacity=".18" filter="url(#bb)"/>'
    # livros empilhados
    y = 760
    for (wdt, hh, c) in [(360, 46, "#3d8bff"), (330, 40, "#e8590c"), (350, 52, "#2f9e44"), (300, 38, "#ae3ec9")]:
        x = 1140 + r.uniform(-20, 20)
        b += f'<rect x="{x:.0f}" y="{y-hh}" width="{wdt}" height="{hh}" rx="5" fill="{c}"/><rect x="{x+wdt-16:.0f}" y="{y-hh+5}" width="10" height="{hh-10}" fill="#f1e3c8"/>'
        y -= hh
    b += '<text x="1300" y="733" text-anchor="middle" font-family="Poppins" font-weight="800" font-size="20" fill="#fff" opacity=".85">JAVA</text>'
    # caderno aberto
    b += '<g transform="translate(760,830) rotate(-4)"><path d="M0,0 L300,-10 L300,150 L0,160 Z" fill="#f4efe3"/><path d="M300,-10 L600,0 L600,160 L300,150 Z" fill="#ece5d6"/>'
    for i in range(7):
        b += f'<path d="M20,{20+i*19} L280,{12+i*19} M320,{12+i*19} L580,{20+i*19}" stroke="#9fb3d1" stroke-width="1.5" opacity=".7"/>'
    b += '<text x="40" y="44" font-family="Latin Modern Roman" font-style="italic" font-size="22" fill="#2b3a67">∫ f(x) dx</text><text x="340" y="44" font-family="DejaVu Sans Mono" font-size="16" fill="#2b3a67">@GetMapping</text></g>'
    b += '<rect x="1400" y="880" width="260" height="10" rx="5" fill="#ffd43b" transform="rotate(-12 1400 880)"/><path d="M1655,826 l18,-6 l-10,14z" fill="#333"/>'
    # caneca
    b += '<rect x="380" y="700" width="96" height="100" rx="14" fill="#e9ecef"/><path d="M476,722 q44,4 40,36 q-4,30 -40,32" stroke="#e9ecef" stroke-width="13" fill="none"/>'
    b += '<text x="428" y="765" text-anchor="middle" font-family="Poppins" font-weight="800" font-size="30" fill="#3d8bff">⚡</text>'
    for i in range(3):
        b += f'<path d="M{404+i*24},690 q-14,-30 0,-55 q14,-25 0,-55" stroke="#fff" stroke-width="5" fill="none" opacity=".3" filter="url(#b8)" stroke-linecap="round"/>'
    b += finish()
    return svg(b, defs)
walls["estudos-mesa"] = estudos_mesa

# ---------------------------------------------------------------- ESTUDOS 2: biblioteca
def estudos_biblioteca():
    r = random.Random(10)
    defs = BLUR("bb", 90) + GRAIN + VIGNETTE + '''
    <linearGradient id="wood" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#4a2e1a"/><stop offset="1" stop-color="#2e1c10"/></linearGradient>
    <radialGradient id="light" cx="50%" cy="30%" r="60%"><stop offset="0" stop-color="#ffcf7a" stop-opacity=".35"/><stop offset="1" stop-color="#000" stop-opacity="0"/></radialGradient>'''
    b = '<rect width="1920" height="1080" fill="#1e130b"/>'
    palette = ["#7b2d26", "#2d4a7b", "#2f6b4f", "#8a6a2b", "#5b3a73", "#a3462f", "#2b5e6b", "#6b2b4a", "#c9a227", "#3d3d5c", "#8f3b1b", "#1f4e3d"]
    shelf_h = 200
    for row in range(6):
        top = 30 + row * shelf_h
        b += f'<rect x="0" y="{top+shelf_h-26}" width="1920" height="26" fill="url(#wood)"/><rect x="0" y="{top+shelf_h-26}" width="1920" height="5" fill="#6a4529"/>'
        x = 40
        while x < 1880:
            if r.random() < .06:
                x += r.randint(30, 80); continue
            w = r.randint(22, 46); h = r.randint(120, 168)
            c = r.choice(palette)
            lean = r.random() < .05
            y = top + shelf_h - 26 - h
            tr = f' transform="rotate(-10 {x} {top+shelf_h-26})"' if lean else ""
            b += f'<g{tr}><rect x="{x}" y="{y}" width="{w}" height="{h}" rx="3" fill="{c}"/>'
            b += f'<rect x="{x}" y="{y+18}" width="{w}" height="5" fill="#e6c97a" opacity=".55"/><rect x="{x}" y="{y+h-26}" width="{w}" height="5" fill="#e6c97a" opacity=".55"/>'
            if w > 32 and r.random() < .5:
                b += f'<rect x="{x+w/2-5}" y="{y+40}" width="10" height="{h-90}" rx="3" fill="#000" opacity=".18"/>'
            b += '</g>'
            x += w + (30 if lean else 2)
    for x in (0, 1890):
        b += f'<rect x="{x}" y="0" width="30" height="1080" fill="url(#wood)"/>'
    b += '<rect width="1920" height="1080" fill="url(#light)"/>'
    b += '<rect width="1920" height="1080" fill="#000" opacity=".25"/>'
    b += finish()
    return svg(b, defs)
walls["estudos-biblioteca"] = estudos_biblioteca

# ---------------------------------------------------------------- JOGOS 2: arcade neon
def jogos_neon():
    defs = GLOW("g", 6) + GLOW("g2", 14) + BLUR("bb", 60) + GRAIN + VIGNETTE + '''
    <linearGradient id="sky" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#090418"/><stop offset=".6" stop-color="#2a0a4a"/><stop offset="1" stop-color="#5c0f6b"/></linearGradient>
    <linearGradient id="sun" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#ffe259"/><stop offset="1" stop-color="#ff2e88"/></linearGradient>
    <linearGradient id="floor" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#1a0630"/><stop offset="1" stop-color="#05010d"/></linearGradient>
    <clipPath id="sunclip"><rect x="0" y="0" width="1920" height="600"/></clipPath>'''
    b = '<rect width="1920" height="1080" fill="url(#sky)"/>'
    r = random.Random(12)
    for _ in range(120):
        b += f'<circle cx="{r.uniform(0,W):.0f}" cy="{r.uniform(0,520):.0f}" r="{r.uniform(.6,2):.1f}" fill="#fff" opacity="{r.uniform(.2,.9):.2f}"/>'
    sun = '<g clip-path="url(#sunclip)"><circle cx="960" cy="470" r="250" fill="url(#sun)" filter="url(#g)"/>'
    for i in range(8):
        y = 500 + i * 16; h = 4 + i * 1.6
        sun += f'<rect x="700" y="{y:.0f}" width="520" height="{h:.0f}" fill="#2a0a4a"/>'
    b += sun + '</g>'
    b += '<path d="M0,600 L180,470 L330,560 L520,420 L700,600 Z" fill="#14052a" stroke="#ff2e88" stroke-width="2" opacity=".9"/>'
    b += '<path d="M1220,600 L1400,440 L1560,540 L1740,410 L1920,600 Z" fill="#14052a" stroke="#22d3ff" stroke-width="2" opacity=".9"/>'
    b += '<rect x="0" y="600" width="1920" height="480" fill="url(#floor)"/>'
    g = '<g stroke="#ff2e88" stroke-width="2.5" filter="url(#g)" opacity=".85">'
    for i in range(-24, 25):
        g += f'<line x1="{960 + i*40}" y1="600" x2="{960 + i*260}" y2="1080"/>'
    y = 600; dy = 6
    while y < 1080:
        g += f'<line x1="0" y1="{y:.0f}" x2="1920" y2="{y:.0f}"/>'
        y += dy; dy *= 1.28
    b += g + '</g>'
    b += '<line x1="0" y1="600" x2="1920" y2="600" stroke="#22d3ff" stroke-width="4" filter="url(#g2)"/>'
    # controle genérico em neon
    cx, cy = 960, 860
    pad = f'<g transform="translate({cx},{cy})" fill="none" stroke="#22d3ff" stroke-width="6" stroke-linejoin="round" filter="url(#g)">'
    pad += '<path d="M-230,-70 C-300,-70 -330,40 -300,110 C-280,160 -220,160 -190,110 L-150,60 L150,60 L190,110 C220,160 280,160 300,110 C330,40 300,-70 230,-70 Z"/>'
    pad += '<path d="M-200,-10 h30 v-30 h30 v30 h30 v30 h-30 v30 h-30 v-30 h-30 z" stroke="#ff2e88"/>'
    for (x, y, c) in [(180, -30, "#ffe259"), (220, 5, "#ff2e88"), (180, 40, "#7CFF6B"), (140, 5, "#22d3ff")]:
        pad += f'<circle cx="{x}" cy="{y}" r="15" stroke="{c}"/>'
    pad += '<rect x="-40" y="-10" width="30" height="12" rx="6" stroke="#fff"/><rect x="10" y="-10" width="30" height="12" rx="6" stroke="#fff"/></g>'
    b += pad
    b += '<text x="960" y="180" text-anchor="middle" font-family="DejaVu Sans Mono" font-weight="bold" font-size="54" fill="#fff" letter-spacing="14" filter="url(#g2)">PRESS START</text>'
    b += '<text x="960" y="236" text-anchor="middle" font-family="DejaVu Sans Mono" font-size="24" fill="#22d3ff" letter-spacing="8" opacity=".85">PLAYER 1 · LEVEL 99</text>'
    b += finish()
    return svg(b, defs)
walls["jogos-neon"] = jogos_neon

for k, fn in walls.items():
    open(os.path.join(OUT, k + ".svg"), "w").write(fn())
print("svgs:", len(walls))
