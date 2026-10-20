/** Cifrado AES-256-GCM con ENCRYPTION_KEY (credenciales de integraciones). Como en Caja. */
import { env } from 'cloudflare:workers';
import { b64url, fromB64url } from './crypto.ts';

const enc = new TextEncoder();

let aesKey: Promise<CryptoKey> | undefined;
function key() {
	aesKey ??= crypto.subtle.importKey('raw', fromB64url(env.ENCRYPTION_KEY), 'AES-GCM', false, [
		'encrypt',
		'decrypt'
	]);
	return aesKey;
}

/** Cifra con AES-256-GCM. Formato: iv.ciphertext (base64url). */
export async function encrypt(plain: string): Promise<string> {
	const iv = crypto.getRandomValues(new Uint8Array(12));
	const ct = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, await key(), enc.encode(plain));
	return `${b64url(iv)}.${b64url(ct)}`;
}

export async function decrypt(payload: string): Promise<string> {
	const [iv, ct] = payload.split('.');
	const pt = await crypto.subtle.decrypt(
		{ name: 'AES-GCM', iv: fromB64url(iv) },
		await key(),
		fromB64url(ct)
	);
	return new TextDecoder().decode(pt);
}
