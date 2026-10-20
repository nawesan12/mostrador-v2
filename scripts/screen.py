#!/usr/bin/env python3
"""Extrae del handoff el HTML de una pantalla por su data-screen-label.

Uso:
  python3 scripts/screen.py                       # lista todas las pantallas
  python3 scripts/screen.py "Inicio desktop"      # imprime el markup (indentado)
  python3 scripts/screen.py "Inicio desktop" --data   # además, los arrays de datos del <script>
"""
import glob, os, re, sys
from html.parser import HTMLParser

ROOT = os.path.join(os.path.dirname(__file__), '..', 'design')
VOID = {'img', 'br', 'input', 'meta', 'link', 'hr', 'source'}

def files():
    return sorted(glob.glob(os.path.join(ROOT, '*.dc.html')))

def labels():
    for f in files():
        for m in re.finditer(r'data-screen-label="([^"]+)"', open(f).read()):
            yield os.path.basename(f), m.group(1)

class Extract(HTMLParser):
    def __init__(self, label):
        super().__init__(convert_charrefs=False)
        self.label, self.depth, self.out, self.on, self.lvl = label, 0, [], False, 0
    def handle_starttag(self, tag, attrs):
        a = dict(attrs)
        if not self.on and a.get('data-screen-label') == self.label:
            self.on, self.lvl = True, 0
        if self.on:
            attrs_s = ''.join(f' {k}="{v}"' if v is not None else f' {k}' for k, v in attrs)
            self.out.append('  ' * self.lvl + f'<{tag}{attrs_s}>')
            if tag not in VOID: self.lvl += 1
    def handle_endtag(self, tag):
        if self.on and tag not in VOID:
            self.lvl -= 1
            self.out.append('  ' * self.lvl + f'</{tag}>')
            if self.lvl == 0: self.on = False; raise StopIteration
    def handle_data(self, data):
        if self.on and data.strip(): self.out.append('  ' * self.lvl + data.strip())

def extract(label):
    for f in files():
        t = open(f).read()
        if f'data-screen-label="{label}"' in t:
            p = Extract(label)
            try: p.feed(t)
            except StopIteration: pass
            return os.path.basename(f), '\n'.join(p.out), t
    return None, None, None

if __name__ == '__main__':
    if len(sys.argv) < 2:
        for f, l in labels(): print(f'{f}: {l}')
        sys.exit()
    f, html, src = extract(sys.argv[1])
    if not html: sys.exit(f'No encontré "{sys.argv[1]}"')
    print(f'<!-- {f} -->'); print(html)
    if '--data' in sys.argv:
        m = re.search(r'<script type="text/x-dc"[^>]*>(.*?)</script>', src, re.S)
        if m: print('\n<!-- datos/lógica -->\n' + m.group(1))
