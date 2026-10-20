/**
 * Captura la portada de cada plantilla trabajada del handoff (design/Plantilla X.dc.html) para las
 * tarjetas de la galería: escritorio, página de inicio, sin la barra de prototipo.
 *
 * Uso (macOS):
 *   (cd ../../design && python3 -m http.server 8765) &
 *   bun scripts/portadas.mjs
 * Deja static/plantillas/<id>.jpg (611 px de alto, JPG 80).
 */
import { execFileSync } from 'node:child_process';
import { chromium } from '@playwright/test';
const IDS = {
	ceibo: 'Ceibo',
	mostaza: 'Mostaza',
	yerbal: 'Yerbal',
	tijera: 'Tijera',
	pampa: 'Pampa',
	malbec: 'Malbec',
	miga: 'Miga',
	rosedal: 'Rosedal',
	andes: 'Andes',
	quebracho: 'Quebracho'
};
const b = await chromium.launch();
for (const [id, nombre] of Object.entries(IDS)) {
	const ctx = await b.newContext({ viewport: { width: 1200, height: 1310 } });
	await ctx.addInitScript((id) => {
		localStorage.setItem(`tpl-${id}-page`, 'inicio');
		localStorage.setItem(`tpl-${id}-dev`, 'desk');
		sessionStorage.setItem('tpl-malbec-18', '1');
	}, id);
	const p = await ctx.newPage();
	await p.goto(`http://localhost:8765/Plantilla%20${encodeURIComponent(nombre)}.dc.html`, {
		waitUntil: 'networkidle'
	});
	await p.waitForTimeout(2500);
	// Oculta la barra de prototipo (fija abajo, con los botones Escritorio/Celular).
	await p.evaluate(() => {
		for (const el of document.querySelectorAll('div')) {
			const s = getComputedStyle(el);
			if (s.position === 'fixed' && s.zIndex === '100' && el.textContent.includes('Inicio'))
				el.style.display = 'none';
		}
	});
	await p.waitForTimeout(400);
	await p.screenshot({ path: `/tmp/tpl-${id}.png` });
	await ctx.close();
	console.log(id);
}
await b.close();
for (const id of Object.keys(IDS))
	execFileSync('sips', [
		'-Z',
		'611',
		'-s',
		'format',
		'jpeg',
		'-s',
		'formatOptions',
		'80',
		`/tmp/tpl-${id}.png`,
		'--out',
		`static/plantillas/${id}.jpg`
	]);
