"""
Convierte el marcado de una plantilla del handoff (design/Plantilla X.dc.html, formato dc) a Svelte:
sc-if → {#if}, sc-for → {#each}, {{ x }} → {x}, onClick → onclick, iconify-icon → <Icon>.
Uso: python3 scripts/plantilla.py "Plantilla Ceibo" > salida.svelte
Imprime también los íconos usados (para sumarlos a lib/icons.ts).
"""
import re, sys

nombre = sys.argv[1]
s = open(f"../../design/{nombre}.dc.html").read()
i = s.find("<x-dc>"); j = s.find('<script type="text/x-dc"')
cuerpo = s[i + 6 : j]
cuerpo = re.sub(r"<helmet>.*?</helmet>", "", cuerpo, flags=re.S)
cuerpo = cuerpo.replace("</x-dc>", "")

iconos = set()

def camel(n):
    n = n.replace("simple-icons:", "si-")
    partes = re.split(r"[:\-]+", n)
    return partes[0] + "".join(p[:1].upper() + p[1:] for p in partes[1:])

def icono(m):
    attrs = m.group(1)
    nm = re.search(r'icon="([^"]+)"', attrs).group(1)
    w = re.search(r'width="(\d+)"', attrs)
    st = re.search(r'style="([^"]*)"', attrs)
    if "{{" in nm:
        expr = "{" + nm.strip("{} ").strip() + "}"
    else:
        iconos.add(nm); expr = "{I." + camel(nm) + "}"
    out = f"<Icon icon={expr}" + (f" size={{{w.group(1)}}}" if w else "")
    if st:
        out += f' style="{st.group(1)}"'
    return out + " />"

cuerpo = re.sub(r"<iconify-icon([^>]*)></iconify-icon>", icono, cuerpo)
cuerpo = re.sub(r'<sc-if value="\{\{\s*([^}]+?)\s*\}\}"[^>]*>', r"{#if \1}", cuerpo)
cuerpo = cuerpo.replace("</sc-if>", "{/if}")
cuerpo = re.sub(r'<sc-for list="\{\{\s*([^}]+?)\s*\}\}" as="(\w+)"[^>]*>', r"{#each \1 as \2}", cuerpo)
cuerpo = cuerpo.replace("</sc-for>", "{/each}")
cuerpo = re.sub(r'onClick="\{\{\s*([^}]+?)\s*\}\}"', r"onclick={\1}", cuerpo)
cuerpo = re.sub(r'(\w[\w-]*)="\{\{\s*([^}]+?)\s*\}\}"', r"\1={\2}", cuerpo)
cuerpo = re.sub(r"\{\{\s*([^}]+?)\s*\}\}", r"{\1}", cuerpo)
cuerpo = cuerpo.replace(' data-scroll=""', "")
print(cuerpo)
print("<!-- ICONOS: " + " ".join(sorted(iconos)) + " -->", file=sys.stderr)
