/**
 * Pedido de certificado (CSR, PKCS#10) con WebCrypto, con el mismo sujeto que arma Go
 * (`arca.PedidoCertificado`): C=AR, O=razón social, CN=alias, serialNumber="CUIT <cuit>", clave RSA
 * de 2048 bits y firma SHA-256.
 *
 * Con la cuenta real lo genera Go y la clave queda cifrada en Tu Local Web. En la demo lo genera el
 * navegador del visitante (Workers Free tiene 10 ms de CPU por pedido y una clave RSA tarda de 20 a
 * 100 ms): el pedido es real y válido, pero la clave no se guarda en ningún lado.
 */
import { cuitValido } from './cuit.ts';

const largo = (n: number) => {
	if (n < 0x80) return [n];
	const b: number[] = [];
	for (let x = n; x > 0; x >>= 8) b.unshift(x & 0xff);
	return [0x80 | b.length, ...b];
};
const tlv = (tag: number, ...partes: Uint8Array[]) => {
	const cuerpo = concatenar(partes);
	return concatenar([new Uint8Array([tag, ...largo(cuerpo.length)]), cuerpo]);
};
function concatenar(partes: Uint8Array[]) {
	const r = new Uint8Array(partes.reduce((s, p) => s + p.length, 0));
	let i = 0;
	for (const p of partes) {
		r.set(p, i);
		i += p.length;
	}
	return r;
}
const sec = (...p: Uint8Array[]) => tlv(0x30, ...p);
const set = (...p: Uint8Array[]) => tlv(0x31, ...p);

function oid(s: string) {
	const n = s.split('.').map(Number);
	const b = [40 * n[0] + n[1]];
	for (const v of n.slice(2)) {
		const g = [v & 0x7f];
		for (let x = v >> 7; x > 0; x >>= 7) g.unshift(0x80 | (x & 0x7f));
		b.push(...g);
	}
	return tlv(0x06, new Uint8Array(b));
}

// PrintableString si alcanza (como Go), si no UTF8String.
const texto = (s: string) =>
	tlv(/^[A-Za-z0-9 '()+,\-./:=?]*$/.test(s) ? 0x13 : 0x0c, new TextEncoder().encode(s));
const atributo = (o: string, v: string) => set(sec(oid(o), texto(v)));

const pem = (tipo: string, der: Uint8Array) => {
	let b = '';
	for (const x of der) b += String.fromCharCode(x);
	const b64 = btoa(b).replace(/.{1,64}/g, '$&\n');
	return `-----BEGIN ${tipo}-----\n${b64}-----END ${tipo}-----\n`;
};

/** El CSR en PEM. La clave privada se genera acá y se descarta (no es extraíble). */
export async function pedidoCertificado(
	cuit: string,
	razonSocial: string,
	alias = 'mostrador'
): Promise<string> {
	razonSocial = razonSocial.trim();
	if (!cuitValido(cuit)) throw new Error('CUIT inválido');
	if (!razonSocial || razonSocial.length > 64) throw new Error('Razón social inválida');
	const par = await crypto.subtle.generateKey(
		{
			name: 'RSASSA-PKCS1-v1_5',
			modulusLength: 2048,
			publicExponent: new Uint8Array([1, 0, 1]),
			hash: 'SHA-256'
		},
		false,
		['sign']
	);
	const spki = new Uint8Array(await crypto.subtle.exportKey('spki', par.publicKey));
	const sujeto = sec(
		atributo('2.5.4.6', 'AR'),
		atributo('2.5.4.10', razonSocial),
		atributo('2.5.4.3', alias),
		atributo('2.5.4.5', `CUIT ${cuit}`)
	);
	const info = sec(
		tlv(0x02, new Uint8Array([0])),
		sujeto,
		spki,
		new Uint8Array([0xa0, 0x00]) // atributos: ninguno
	);
	const firma = new Uint8Array(await crypto.subtle.sign('RSASSA-PKCS1-v1_5', par.privateKey, info));
	const algoritmo = sec(oid('1.2.840.113549.1.1.11'), new Uint8Array([0x05, 0x00]));
	const der = sec(info, algoritmo, tlv(0x03, new Uint8Array([0]), firma));
	return pem('CERTIFICATE REQUEST', der);
}
