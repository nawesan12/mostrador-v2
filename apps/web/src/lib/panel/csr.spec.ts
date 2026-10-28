import { execFileSync } from 'node:child_process';
import { describe, expect, it } from 'vitest';
import { pedidoCertificado } from './csr.ts';

const hayOpenssl = (() => {
	try {
		execFileSync('openssl', ['version']);
		return true;
	} catch {
		return false;
	}
})();

describe('pedidoCertificado', () => {
	it('arma un CSR en PEM con el sujeto que pide ARCA', async () => {
		const csr = await pedidoCertificado('20448592341', 'Nahuel Santillán');
		expect(csr.startsWith('-----BEGIN CERTIFICATE REQUEST-----\n')).toBe(true);
		expect(csr.trimEnd().endsWith('-----END CERTIFICATE REQUEST-----')).toBe(true);
		if (hayOpenssl) {
			// La firma verifica y el sujeto lleva el CUIT como serialNumber.
			const out = execFileSync(
				'openssl',
				['req', '-verify', '-noout', '-subject', '-nameopt', 'utf8,sep_comma_plus'],
				{ input: csr, stdio: ['pipe', 'pipe', 'pipe'] }
			).toString();
			expect(out).toContain('C=AR');
			expect(out).toContain('O=Nahuel Santillán');
			expect(out).toContain('CN=mostrador');
			expect(out).toContain('serialNumber=CUIT 20448592341');
		}
	});

	it('no arma nada con un CUIT mal', async () => {
		await expect(pedidoCertificado('20448592342', 'Tienda')).rejects.toThrow('CUIT');
	});
});
