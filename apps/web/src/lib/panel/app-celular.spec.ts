import { describe, expect, it } from 'vitest';
import { aparato, claveABytes, estadoPush } from './app-celular.svelte.ts';

describe('push web en el celular', () => {
	it('pasa la clave VAPID de base64url a los 65 bytes del punto P-256', () => {
		const b = claveABytes(
			'BP4z9KsN6nGRTbVYI_c7VJSPQTBtkgcy27mlmlMoZIIgDll6e3vCYLocInmYWAmS6TlzAC8wEqKK6PBru3jl7A8'
		);
		expect(b.length).toBe(65);
		expect(b[0]).toBe(4);
		expect(b[64]).toBe(0x0f);
	});

	it('nombra el aparato sin datos de más', () => {
		expect(
			aparato(
				'Mozilla/5.0 (iPhone; CPU iPhone OS 18_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Mobile/15E148 Safari/604.1'
			)
		).toBe('iPhone · Safari');
		expect(
			aparato(
				'Mozilla/5.0 (Linux; Android 14; SM-A546E) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Mobile Safari/537.36'
			)
		).toBe('Android · Chrome');
		expect(
			aparato(
				'Mozilla/5.0 (Linux; Android 14; SM-A546E) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/26.0 Chrome/122.0.0.0 Mobile Safari/537.36'
			)
		).toBe('Android · Samsung Internet');
	});

	it('sin clave en el servidor, no está disponible', async () => {
		expect(await estadoPush({ clave: '', suscripta: false })).toBe('no-disponible');
	});
});
