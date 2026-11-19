import { describe, expect, it } from 'vitest';
import { UiState } from './ui.svelte.ts';

describe('el aviso de un tope de plan', () => {
	const desc = 'Cupones, combos y carritos abandonados: desde el plan Emprende. Tu plan es Gratis.';

	it('reemplaza al error genérico de la pantalla si éste llegó primero', () => {
		const ui = new UiState();
		ui.toast({ k: 'err', title: 'No se pudo', desc });
		ui.toast({
			k: 'info',
			tope: true,
			title: 'Es del plan Emprende',
			desc,
			action: 'Pasar a Emprende'
		});
		expect(ui.toasts.map((t) => t.title)).toEqual(['Es del plan Emprende']);
	});

	it('no deja pasar el error genérico si el tope ya se avisó', () => {
		const ui = new UiState();
		ui.toast({ k: 'info', tope: true, title: 'Es del plan Emprende', desc });
		ui.toast({ k: 'err', title: 'No se pudo', desc });
		ui.toast({ k: 'info', tope: true, title: 'Es del plan Emprende', desc });
		expect(ui.toasts).toHaveLength(1);
	});

	it('un error distinto sí se muestra', () => {
		const ui = new UiState();
		ui.toast({ k: 'info', tope: true, title: 'Es del plan Emprende', desc });
		ui.toast({ k: 'err', title: 'No se pudo', desc: 'Otra cosa.' });
		expect(ui.toasts).toHaveLength(2);
	});
});
