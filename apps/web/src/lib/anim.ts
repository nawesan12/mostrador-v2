/**
 * Animaciones de entrada del handoff (`data-anim` en design/Mostrador v4.dc.html), como acción:
 * `<div use:anim={'rise'}>` o `use:anim={{ t: 'reveal', delay: 90 }}`.
 * Respeta `prefers-reduced-motion`: sin movimiento, el elemento aparece directo.
 */
import type { Action } from 'svelte/action';

export const EASE = 'cubic-bezier(.2,.7,.1,1)';

export type Anim =
	| 'reveal'
	| 'rise'
	| 'fade'
	| 'float'
	| 'pulse'
	| 'dots'
	| 'pop'
	| 'drop'
	| 'marquee'
	| 'marquee-r';
/** HTML o SVG: cualquier elemento con estilo propio. */
type Animable = Element & ElementCSSInlineStyle;
type Opciones = { t: Anim; delay?: number; dur?: number };

export const calma = () =>
	typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;

const infinito = { iterations: Infinity } as const;

function correr(el: Animable, { t, delay = 0, dur = 0 }: Opciones) {
	switch (t) {
		case 'reveal':
		case 'rise':
			return el.animate(
				[
					{ opacity: 0, transform: `translateY(${t === 'reveal' ? 30 : 14}px)` },
					{ opacity: 1, transform: 'none' }
				],
				{ duration: t === 'reveal' ? 900 : 700, delay, fill: 'both', easing: EASE }
			);
		case 'fade':
			return el.animate([{ opacity: 0 }, { opacity: 1 }], { duration: 1000, delay, fill: 'both' });
		case 'float':
			return el.animate([{ transform: 'translateY(0)' }, { transform: 'translateY(-10px)' }], {
				duration: dur || 3600,
				delay,
				direction: 'alternate',
				easing: 'ease-in-out',
				...infinito
			});
		case 'pulse':
			return el.animate(
				[
					{ transform: 'scale(1)', opacity: 0.6 },
					{ transform: 'scale(2.6)', opacity: 0 }
				],
				{ duration: dur || 2000, easing: 'ease-out', ...infinito }
			);
		case 'dots':
			return el.animate([{ opacity: 0.25 }, { opacity: 1 }, { opacity: 0.25 }], {
				duration: 1100,
				delay,
				...infinito
			});
		case 'pop':
			return el.animate(
				[
					{ opacity: 0, transform: 'translateY(-8px) scale(.98)' },
					{ opacity: 1, transform: 'none' }
				],
				{ duration: 240, easing: EASE, fill: 'both' }
			);
		case 'drop':
			return el.animate(
				[
					{ opacity: 0, transform: 'translateY(-16px) scaleY(.96)' },
					{ opacity: 1, transform: 'none' }
				],
				{ duration: 380, easing: 'cubic-bezier(.16,1,.3,1)', fill: 'both' }
			);
		case 'marquee':
		case 'marquee-r': {
			const a = { transform: 'translateX(0)' };
			const b = { transform: 'translateX(-50%)' };
			return el.animate(t === 'marquee' ? [a, b] : [b, a], { duration: dur || 50000, ...infinito });
		}
	}
}

export const anim: Action<Animable, Anim | Opciones> = (el, param) => {
	const op: Opciones = typeof param === 'string' ? { t: param } : param;
	// Sin movimiento: lo decorativo infinito no corre y lo de entrada aparece directo.
	if (calma()) return;
	if (op.t !== 'reveal') {
		const a = correr(el, op);
		return { destroy: () => a?.cancel() };
	}
	// reveal: anima recién cuando entra en pantalla.
	el.style.opacity = '0';
	const io = new IntersectionObserver(
		(entradas) => {
			for (const e of entradas) {
				if (!e.isIntersecting) continue;
				correr(el, op);
				io.unobserve(el);
			}
		},
		{ threshold: 0.1 }
	);
	io.observe(el);
	return { destroy: () => io.disconnect() };
};
