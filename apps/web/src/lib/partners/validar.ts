/**
 * Validación del perfil público y de los datos de cobro: la misma que Go
 * (apps/api/internal/dominio/partners/partners.go). La usa la demo y el formulario antes de mandar.
 */
import { cbuValido } from '#lib/panel/cbu.ts';
import { cuitValido, type CobroPartner, type PerfilPartner } from './programa.ts';

export type Invalido = { campo: string; mensaje: string };
type Resultado<T> = { ok: true; valor: T } | ({ ok: false } & Invalido);

export const ESPECIALIDADES = [
	'diseno',
	'desarrollo',
	'marketing',
	'contabilidad',
	'fotografia'
] as const;
export const CONDICIONES_IVA = [
	['monotributo', 'Monotributo'],
	['responsable_inscripto', 'Responsable inscripto'],
	['exento', 'Exento']
] as const;

const limpiar = (v: unknown) =>
	String(v ?? '')
		.trim()
		.replace(/\s+/g, ' ');
const largo = (v: string) => [...v].length;
const mal = (campo: string, mensaje: string) => ({ ok: false as const, campo, mensaje });

function https(v: unknown, max: number) {
	const s = String(v ?? '').trim();
	if (!s) return '';
	try {
		const u = new URL(s);
		if (u.protocol !== 'https:' || !u.host || s.length > max) return null;
	} catch {
		return null;
	}
	return s;
}

export function validarPerfil(p: Partial<PerfilPartner>): Resultado<PerfilPartner> {
	const nombre = limpiar(p.nombre);
	if (!nombre || largo(nombre) > 80)
		return mal('nombre', 'Poné el nombre con el que te conocen (hasta 80 letras).');
	const ciudad = limpiar(p.ciudad);
	if (largo(ciudad) > 60) return mal('ciudad', 'La ciudad es muy larga.');
	const descripcion = limpiar(p.descripcion);
	if (largo(descripcion) > 280) return mal('descripcion', 'Contalo en 280 letras o menos.');
	const especialidades: PerfilPartner['especialidades'] = [];
	for (const e of p.especialidades ?? []) {
		if (!(ESPECIALIDADES as readonly string[]).includes(e))
			return mal('especialidades', 'Elegí especialidades de la lista.');
		if (!especialidades.includes(e)) especialidades.push(e);
	}
	if ((p.casos ?? []).length > 6) return mal('casos', 'Hasta 6 casos.');
	const casos: PerfilPartner['casos'] = [];
	for (const c of p.casos ?? []) {
		const n = limpiar(c.nombre);
		if (!n && !String(c.url ?? '').trim()) continue;
		if (!n || largo(n) > 60)
			return mal('casos', 'Cada caso lleva el nombre de la tienda (hasta 60 letras).');
		const url = https(c.url, 200);
		if (!url) return mal('casos', 'Cada caso lleva el link de la tienda, con https://.');
		casos.push({ nombre: n, url });
	}
	const email = String(p.email ?? '').trim();
	if (email && (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email) || email.length > 254))
		return mal('email', 'Revisá el email.');
	const whatsapp = String(p.whatsapp ?? '').replace(/\D/g, '');
	if (whatsapp && (whatsapp.length < 8 || whatsapp.length > 15))
		return mal('whatsapp', 'El WhatsApp va con la característica: 11 5555 1234.');
	const web = https(p.web, 200);
	if (web === null) return mal('web', 'Poné un link que empiece con https://.');
	const logo = https(p.logo, 300);
	if (logo === null) return mal('logo', 'Poné un link que empiece con https://.');
	return {
		ok: true,
		valor: { nombre, ciudad, descripcion, especialidades, casos, email, whatsapp, web, logo }
	};
}

export function validarCobro(c: Partial<CobroPartner>): Resultado<CobroPartner> {
	let valor = String(c.valor ?? '');
	if (c.tipo === 'cbu') {
		valor = valor.replace(/\D/g, '');
		if (!cbuValido(valor)) return mal('valor', 'Revisá el CBU o CVU: son 22 números.');
	} else if (c.tipo === 'alias') {
		valor = valor.trim();
		if (!/^[a-zA-Z0-9.-]{6,20}$/.test(valor))
			return mal('valor', 'El alias tiene entre 6 y 20 letras, números, puntos o guiones.');
	} else return mal('tipo', 'Elegí CBU/CVU o alias.');
	const titular = limpiar(c.titular);
	if (!titular || largo(titular) > 80) return mal('titular', 'Poné el titular de la cuenta.');
	const cuit = String(c.cuit ?? '').replace(/\D/g, '');
	if (!cuitValido(cuit)) return mal('cuit', 'Revisá el CUIT: son 11 números.');
	if (!CONDICIONES_IVA.some(([k]) => k === c.condicionIva))
		return mal('condicionIva', 'Elegí tu condición frente al IVA.');
	return {
		ok: true,
		valor: { tipo: c.tipo, valor, titular, cuit, condicionIva: c.condicionIva! }
	};
}

/** El link de una factura: https, hasta 500. */
export function validarFactura(url: unknown): Resultado<string> {
	const u = https(url, 500);
	if (!u) return mal('url', 'Pegá el link de la factura (PDF en Drive, Dropbox o ARCA).');
	return { ok: true, valor: u };
}
