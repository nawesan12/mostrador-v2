/**
 * La sesión: la casilla y su clave en una cookie httpOnly cifrada (AES-256-GCM con CORREO_CLAVE).
 * La clave hace falta en cada pedido a JMAP; así el servidor no guarda nada y un reinicio no
 * cierra la sesión de nadie. La cookie vence a los 30 días.
 */
import { createCipheriv, createDecipheriv, randomBytes } from 'node:crypto';
import type { Cookies } from '@sveltejs/kit';

export const COOKIE = 'correo_sesion';
const DIAS = 30;

export type Credenciales = { email: string; clave: string };

function llave(): Buffer {
	const k = Buffer.from(process.env.CORREO_CLAVE ?? '', 'base64');
	if (k.length !== 32) throw new Error('CORREO_CLAVE: 32 bytes en base64');
	return k;
}

export function sellar(c: Credenciales): string {
	const iv = randomBytes(12);
	const ci = createCipheriv('aes-256-gcm', llave(), iv);
	const datos = Buffer.concat([ci.update(JSON.stringify(c), 'utf8'), ci.final()]);
	return Buffer.concat([iv, ci.getAuthTag(), datos]).toString('base64url');
}

export function abrir(sello: string | undefined): Credenciales | null {
	if (!sello) return null;
	try {
		const b = Buffer.from(sello, 'base64url');
		const de = createDecipheriv('aes-256-gcm', llave(), b.subarray(0, 12));
		de.setAuthTag(b.subarray(12, 28));
		const c = JSON.parse(Buffer.concat([de.update(b.subarray(28)), de.final()]).toString('utf8'));
		return typeof c?.email === 'string' && typeof c?.clave === 'string' ? c : null;
	} catch {
		return null;
	}
}

export function guardar(cookies: Cookies, c: Credenciales, seguro: boolean) {
	cookies.set(COOKIE, sellar(c), {
		path: '/',
		httpOnly: true,
		secure: seguro,
		sameSite: 'strict',
		maxAge: DIAS * 86_400
	});
}

export function cerrar(cookies: Cookies) {
	cookies.delete(COOKIE, { path: '/' });
}
