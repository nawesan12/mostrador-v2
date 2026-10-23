import { describe, expect, it } from 'vitest';
import { origenDe } from './origen.ts';

describe('origenDe', () => {
	const ahora = new Date('2026-10-23T15:00:00Z');
	it('toma utm, ref y cupón válidos', () => {
		expect(
			origenDe(new URL('https://x/empezar?utm_source=flyer&ref=BRUMA-REF&cupon=arranco50'), ahora)
		).toEqual({
			fuente: 'flyer',
			ref: 'BRUMA-REF',
			cupon: 'ARRANCO50',
			llego: ahora.toISOString()
		});
	});
	it('descarta lo que no es un código', () => {
		expect(origenDe(new URL('https://x/empezar?ref=<script>&utm_source=a b'), ahora)).toBeNull();
	});
	it('sin nada no anota', () => {
		expect(origenDe(new URL('https://x/empezar'), ahora)).toBeNull();
	});
});
