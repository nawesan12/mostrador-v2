import { describe, expect, it } from 'vitest';
import { sanear } from './html.ts';

const s = (h: string, remotas = false) =>
	sanear(h, { cids: new Map([['logo@x', '/adjunto/a/b?n=logo.png']]), remotas });

describe('sanear el HTML de un correo', () => {
	it('saca scripts, manejadores, formularios e iframes', () => {
		const { html } = s(
			'<p onclick="x()">hola</p><script>alert(1)</script><form action="//x"><input></form><iframe src="//x"></iframe><a href="javascript:alert(1)">a</a>'
		);
		expect(html).not.toMatch(/script|onclick|<form|<input|<iframe|javascript:/i);
		expect(html).toContain('<p>hola</p>');
	});

	it('apaga las imágenes y el CSS de afuera hasta que se piden', () => {
		const r = s(
			'<img src="https://rastreo.com/p.gif"><div style="background:url(https://x.com/a.png)">t</div><style>@import url(https://x.com/a.css);</style>'
		);
		expect(r.remotas).toBe(true);
		expect(r.html).not.toMatch(/rastreo\.com|x\.com/);
		const con = s('<img src="https://rastreo.com/p.gif">', true);
		expect(con.html).toContain('https://rastreo.com/p.gif');
	});

	it('las imágenes cid: van por /adjunto y los links abren afuera', () => {
		const r = s('<img src="cid:logo@x"><a href="https://tulocalweb.com.ar">t</a>');
		expect(r.html).toContain('src="/adjunto/a/b?n=logo.png"');
		expect(r.usados.has('logo@x')).toBe(true);
		expect(r.html).toMatch(/target="_blank"/);
		expect(r.html).toMatch(/rel="noopener noreferrer nofollow"/);
	});
});
