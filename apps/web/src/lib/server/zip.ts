/**
 * ZIP mínimo sin compresión (método "store"): alcanza para juntar los CSV de "Exportar todo" sin
 * sumar una dependencia. Los nombres van en UTF-8 (bit 11).
 */
const TABLA = (() => {
	const t = new Uint32Array(256);
	for (let n = 0; n < 256; n++) {
		let c = n;
		for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
		t[n] = c >>> 0;
	}
	return t;
})();

export function crc32(b: Uint8Array) {
	let c = 0xffffffff;
	for (let i = 0; i < b.length; i++) c = TABLA[(c ^ b[i]) & 0xff] ^ (c >>> 8);
	return (c ^ 0xffffffff) >>> 0;
}

export function zip(
	archivos: { nombre: string; datos: Uint8Array }[],
	fecha = new Date()
): Uint8Array {
	const enc = new TextEncoder();
	const hora =
		(fecha.getHours() << 11) | (fecha.getMinutes() << 5) | Math.floor(fecha.getSeconds() / 2);
	const dia = ((fecha.getFullYear() - 1980) << 9) | ((fecha.getMonth() + 1) << 5) | fecha.getDate();
	const locales: Uint8Array[] = [];
	const centrales: Uint8Array[] = [];
	let offset = 0;
	for (const a of archivos) {
		const nombre = enc.encode(a.nombre);
		const crc = crc32(a.datos);
		const l = new DataView(new ArrayBuffer(30));
		l.setUint32(0, 0x04034b50, true);
		l.setUint16(4, 20, true);
		l.setUint16(6, 0x0800, true);
		l.setUint16(8, 0, true);
		l.setUint16(10, hora, true);
		l.setUint16(12, dia, true);
		l.setUint32(14, crc, true);
		l.setUint32(18, a.datos.length, true);
		l.setUint32(22, a.datos.length, true);
		l.setUint16(26, nombre.length, true);
		l.setUint16(28, 0, true);
		const c = new DataView(new ArrayBuffer(46));
		c.setUint32(0, 0x02014b50, true);
		c.setUint16(4, 20, true);
		c.setUint16(6, 20, true);
		c.setUint16(8, 0x0800, true);
		c.setUint16(10, 0, true);
		c.setUint16(12, hora, true);
		c.setUint16(14, dia, true);
		c.setUint32(16, crc, true);
		c.setUint32(20, a.datos.length, true);
		c.setUint32(24, a.datos.length, true);
		c.setUint16(28, nombre.length, true);
		c.setUint32(42, offset, true);
		locales.push(new Uint8Array(l.buffer), nombre, a.datos);
		centrales.push(new Uint8Array(c.buffer), nombre);
		offset += 30 + nombre.length + a.datos.length;
	}
	const tamCentral = centrales.reduce((s, x) => s + x.length, 0);
	const fin = new DataView(new ArrayBuffer(22));
	fin.setUint32(0, 0x06054b50, true);
	fin.setUint16(8, archivos.length, true);
	fin.setUint16(10, archivos.length, true);
	fin.setUint32(12, tamCentral, true);
	fin.setUint32(16, offset, true);
	const partes = [...locales, ...centrales, new Uint8Array(fin.buffer)];
	const out = new Uint8Array(new ArrayBuffer(partes.reduce((s, x) => s + x.length, 0)));
	let i = 0;
	for (const p of partes) {
		out.set(p, i);
		i += p.length;
	}
	return out;
}

/** CSV: con coma (formato csv) o punto y coma y BOM para que Excel en castellano lo abra bien. */
export function csv(filas: (string | number | null)[][], formato: 'csv' | 'excel') {
	const sep = formato === 'excel' ? ';' : ',';
	const celda = (v: string | number | null) => {
		const s = v === null ? '' : String(v);
		return /[",;\n]/.test(s) ? `"${s.replaceAll('"', '""')}"` : s;
	};
	const texto = filas.map((f) => f.map(celda).join(sep)).join('\r\n');
	return (formato === 'excel' ? String.fromCharCode(0xfeff) : '') + texto;
}
