/** Utilidades criptográficas puras (sin bindings): sirven en el Worker y en tests. */
const enc = new TextEncoder();

export function b64url(bytes: ArrayBuffer | Uint8Array): string {
	const b = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes);
	let s = '';
	for (const x of b) s += String.fromCharCode(x);
	return btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

export function fromB64url(s: string): Uint8Array<ArrayBuffer> {
	const b = atob(s.replace(/-/g, '+').replace(/_/g, '/') + '==='.slice((s.length + 3) % 4));
	const out = new Uint8Array(b.length);
	for (let i = 0; i < b.length; i++) out[i] = b.charCodeAt(i);
	return out;
}

export function randomToken(bytes = 32): string {
	return b64url(crypto.getRandomValues(new Uint8Array(bytes)));
}

/** Código numérico de n dígitos (sin sesgo). */
export function randomDigits(n = 6): string {
	let out = '';
	while (out.length < n) {
		const [x] = crypto.getRandomValues(new Uint8Array(1));
		if (x < 250) out += String(x % 10);
	}
	return out;
}

export async function sha256(input: string): Promise<string> {
	return b64url(await crypto.subtle.digest('SHA-256', enc.encode(input)));
}

export async function hmacHex(secret: string, message: string): Promise<string> {
	const key = await crypto.subtle.importKey(
		'raw',
		enc.encode(secret),
		{ name: 'HMAC', hash: 'SHA-256' },
		false,
		['sign']
	);
	const sig = new Uint8Array(await crypto.subtle.sign('HMAC', key, enc.encode(message)));
	return [...sig].map((b) => b.toString(16).padStart(2, '0')).join('');
}

/** Comparación en tiempo constante. */
export function safeEqual(a: string, b: string): boolean {
	// Un secreto vacío o sin configurar nunca valida (si no, un header vacío pasaría).
	if (!a || !b || a.length !== b.length) return false;
	let r = 0;
	for (let i = 0; i < a.length; i++) r |= a.charCodeAt(i) ^ b.charCodeAt(i);
	return r === 0;
}
