/**
 * Genera src/lib/icons.ts con los íconos que usa el handoff (design/*.dc.html y templates-data.js):
 * Phosphor (ph:*) y logos de marcas (logos:*, simple-icons:*), cada uno como export con nombre
 * para que el bundle sólo lleve los que se importan. Para cada ph:x también va ph:x-fill si existe
 * (el menú lateral muestra el ícono relleno del ítem activo).
 *
 * Uso: bun scripts/icons.ts [ph:otro-icono ...]   (los extra se suman a los del handoff)
 */
import { readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

type Set = {
	prefix: string;
	width?: number;
	height?: number;
	icons: Record<string, { body: string; width?: number; height?: number }>;
	aliases?: Record<string, { parent: string }>;
};

const root = join(import.meta.dirname, '..');
const design = join(root, '../../design');
const load = (p: string): Set =>
	JSON.parse(readFileSync(join(root, '../../node_modules/@iconify-json', p, 'icons.json'), 'utf8'));
const sets: Record<string, Set> = {
	ph: load('ph'),
	logos: load('logos'),
	'simple-icons': load('simple-icons')
};

const files = readdirSync(design).filter(
	(f) => f.endsWith('.dc.html') || f === 'templates-data.js'
);
const names = new Set<string>(process.argv.slice(2));
for (const f of files)
	for (const m of readFileSync(join(design, f), 'utf8').matchAll(
		/\b(ph|logos|simple-icons):([a-z0-9-]+)/g
	))
		names.add(`${m[1]}:${m[2]}`);
for (const n of [...names])
	if (n.startsWith('ph:') && !n.endsWith('-fill'))
		names.add(n.replace(/-(bold|light|thin|duotone)$/, '') + '-fill');

const camel = (s: string) =>
	s.replace(/^simple-icons:/, 'si-').replace(/[:-]+([a-z0-9])/g, (_, c) => c.toUpperCase());

const out: string[] = [];
const missing: string[] = [];
for (const name of [...names].sort()) {
	const [prefix, id] = name.split(':');
	const set = sets[prefix];
	const icon = set.icons[id] ?? set.icons[set.aliases?.[id]?.parent ?? ''];
	if (!icon) {
		missing.push(name);
		continue;
	}
	const w = icon.width ?? set.width ?? 16;
	const h = icon.height ?? set.height ?? 16;
	out.push(`export const ${camel(name)}: IconData = [${w}, ${h}, ${JSON.stringify(icon.body)}];`);
}

writeFileSync(
	join(root, 'src/lib/icons.ts'),
	`// Generado por scripts/icons.ts — no editar a mano. Para sumar un ícono: bun scripts/icons.ts ph:nombre\n` +
		`/** [ancho, alto, cuerpo SVG] */\nexport type IconData = readonly [number, number, string];\n\n` +
		out.join('\n') +
		'\n'
);
console.log(`${out.length} íconos`);
// Los -fill agregados que no existen en Phosphor no son error: el ícono base no tiene variante.
const real = missing.filter((m) => !m.endsWith('-fill'));
if (real.length) console.log('No encontrados:', real.join(', '));
