"""Gera rsrc_windows_amd64.syso (ícone + manifesto + informações de versão)
sem depender de ferramentas externas. Uso: python3 mkrsrc.py <pasta_dos_png> <saida.syso>"""
import struct, sys
from PIL import Image

SRC, OUT = sys.argv[1], sys.argv[2]
VER = tuple(int(x) for x in (sys.argv[3] if len(sys.argv) > 3 else "1.0.0").split(".")) + (0,)

def dib(img):
    w, h = img.size
    px = img.convert("RGBA").load()
    hdr = struct.pack("<IiiHHIIiiII", 40, w, h * 2, 1, 32, 0, 0, 0, 0, 0, 0)
    rows = bytearray()
    for y in range(h - 1, -1, -1):
        for x in range(w):
            r, g, b, a = px[x, y]
            rows += bytes((b, g, r, a))
    stride = ((w + 31) // 32) * 4
    mask = bytearray()
    for y in range(h - 1, -1, -1):
        row = bytearray(stride)
        for x in range(w):
            if px[x, y][3] == 0:
                row[x // 8] |= 0x80 >> (x % 8)
        mask += row
    return hdr + bytes(rows) + bytes(mask)

def png(path):
    return open(path, "rb").read()

icons = []  # (size, data)
for s in (16, 20, 24, 32, 40, 48):
    icons.append((s, dib(Image.open(f"{SRC}/icon_{s}.png"))))
for s in (64, 96, 128, 256):
    icons.append((s, png(f"{SRC}/icon_{s}.png")))

grp = struct.pack("<HHH", 0, 1, len(icons))
for i, (s, data) in enumerate(icons, start=1):
    grp += struct.pack("<BBBBHHIH", 0 if s >= 256 else s, 0 if s >= 256 else s, 0, 0, 1, 32, len(data), i)

manifest = b"""<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
  <assemblyIdentity type="win32" name="naHora" version="1.0.0.0" processorArchitecture="amd64"/>
  <description>naHora</description>
  <dependency><dependentAssembly>
    <assemblyIdentity type="win32" name="Microsoft.Windows.Common-Controls" version="6.0.0.0" processorArchitecture="*" publicKeyToken="6595b64144ccf1df" language="*"/>
  </dependentAssembly></dependency>
  <compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1"><application>
    <supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/>
  </application></compatibility>
  <application xmlns="urn:schemas-microsoft-com:asm.v3"><windowsSettings>
    <dpiAware xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">true/pm</dpiAware>
    <dpiAwareness xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings">PerMonitorV2</dpiAwareness>
  </windowsSettings></application>
  <trustInfo xmlns="urn:schemas-microsoft-com:asm.v3"><security><requestedPrivileges>
    <requestedExecutionLevel level="asInvoker" uiAccess="false"/>
  </requestedPrivileges></security></trustInfo>
</assembly>
"""

# ---- VS_VERSIONINFO ----
def u16z(s): return s.encode("utf-16-le") + b"\0\0"
def pad4(b):
    while len(b) % 4: b += b"\0"
def node(key, value=b"", vtype=1, vlen=0, children=()):
    b = bytearray(6) + u16z(key)
    pad4(b)
    b += value
    for c in children:
        pad4(b)
        b += c
    struct.pack_into("<HHH", b, 0, len(b), vlen, vtype)
    return bytes(b)
def sval(k, v):
    return node(k, u16z(v), 1, len(v) + 1)

ms, ls = (VER[0] << 16) | VER[1], (VER[2] << 16) | VER[3]
fixed = struct.pack("<13I", 0xFEEF04BD, 0x00010000, ms, ls, ms, ls, 0x3F, 0, 0x40004, 1, 0, 0, 0)
vs = ".".join(map(str, VER))
strings = [("CompanyName", "naHora"), ("FileDescription", "naHora"), ("FileVersion", vs),
           ("InternalName", "naHora"), ("LegalCopyright", "naHora"), ("OriginalFilename", "naHora.exe"),
           ("ProductName", "naHora"), ("ProductVersion", vs)]
table = node("041604B0", b"", 1, 0, [sval(k, v) for k, v in strings])
sfi = node("StringFileInfo", b"", 1, 0, [table])
var = node("Translation", struct.pack("<HH", 0x0416, 0x04B0), 0, 4)
vfi = node("VarFileInfo", b"", 1, 0, [var])
version = node("VS_VERSION_INFO", fixed, 0, len(fixed), [sfi, vfi])

# ---- árvore de recursos: tipo -> id -> dados (idioma 0x0416) ----
LANG = 0x0416
tree = {3: {i: d for i, (_, d) in enumerate(icons, start=1)}, 14: {1: grp}, 16: {1: version}, 24: {1: manifest}}
types = sorted(tree)
def dirsize(n): return 16 + 8 * n

off = dirsize(len(types))
name_dir_off = {}
for t in types:
    name_dir_off[t] = off; off += dirsize(len(tree[t]))
lang_dir_off = {}
for t in types:
    for i in sorted(tree[t]):
        lang_dir_off[(t, i)] = off; off += dirsize(1)
entry_off = {}
for t in types:
    for i in sorted(tree[t]):
        entry_off[(t, i)] = off; off += 16
data_off = {}
for t in types:
    for i in sorted(tree[t]):
        off = (off + 7) & ~7
        data_off[(t, i)] = off; off += len(tree[t][i])
sec = bytearray(off)

def put_dir(at, entries):
    struct.pack_into("<IIHHHH", sec, at, 0, 0, 0, 0, 0, len(entries))
    for k, (ident, target) in enumerate(entries):
        struct.pack_into("<II", sec, at + 16 + 8 * k, ident, target)

put_dir(0, [(t, 0x80000000 | name_dir_off[t]) for t in types])
relocs = []
for t in types:
    put_dir(name_dir_off[t], [(i, 0x80000000 | lang_dir_off[(t, i)]) for i in sorted(tree[t])])
    for i in sorted(tree[t]):
        put_dir(lang_dir_off[(t, i)], [(LANG, entry_off[(t, i)])])
        d = tree[t][i]
        struct.pack_into("<IIII", sec, entry_off[(t, i)], data_off[(t, i)], len(d), 0, 0)
        relocs.append(entry_off[(t, i)])
        sec[data_off[(t, i)]:data_off[(t, i)] + len(d)] = d
while len(sec) % 4: sec += b"\0"

# ---- COFF ----
HDR, SHDR = 20, 40
raw_ptr = HDR + SHDR
rel_ptr = raw_ptr + len(sec)
sym_ptr = rel_ptr + 10 * len(relocs)
out = bytearray()
out += struct.pack("<HHIIIHH", 0x8664, 1, 0, sym_ptr, 1, 0, 0)
out += struct.pack("<8sIIIIIIHHI", b".rsrc", 0, 0, len(sec), raw_ptr, rel_ptr, 0, len(relocs), 0, 0x40000040)
out += sec
for r in relocs:
    out += struct.pack("<IIH", r, 0, 3)  # IMAGE_REL_AMD64_ADDR32NB
out += struct.pack("<8sIhHBB", b".rsrc", 0, 1, 0, 3, 0)
out += struct.pack("<I", 4)
open(OUT, "wb").write(out)
print("ok", len(out), "bytes,", len(relocs), "relocs")
