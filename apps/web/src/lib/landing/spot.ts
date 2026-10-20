import type { Action } from 'svelte/action';

/**
 * Reflejo que sigue al mouse en las tarjetas del handoff marcadas `data-spot`: actualiza --mx y --my,
 * que usa el radial-gradient del fondo.
 */
export const spot: Action<HTMLElement> = (el) => {
	const mover = (e: PointerEvent) => {
		const r = el.getBoundingClientRect();
		el.style.setProperty('--mx', `${e.clientX - r.left}px`);
		el.style.setProperty('--my', `${e.clientY - r.top}px`);
	};
	el.addEventListener('pointermove', mover);
	return { destroy: () => el.removeEventListener('pointermove', mover) };
};
