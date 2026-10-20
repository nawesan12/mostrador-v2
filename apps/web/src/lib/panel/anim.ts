/**
 * Animaciones del panel del handoff (`data-anim`): rise, pulse, dots, ring y draw, con las mismas
 * duraciones y curvas. Con "reducir movimiento" no se anima nada.
 */
const E = 'cubic-bezier(.2,.7,.1,1)';
const quieto = () => matchMedia('(prefers-reduced-motion: reduce)').matches;

export type Anim =
	| { t: 'rise'; delay?: number }
	| { t: 'pulse' }
	| { t: 'dots' }
	| { t: 'ring'; target: number }
	| { t: 'draw' };

export function anim(el: Element, a: Anim) {
	if (quieto()) return;
	const inf = { iterations: Infinity };
	if (a.t === 'rise')
		el.animate(
			[
				{ opacity: 0, transform: 'translateY(14px)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 700, delay: a.delay ?? 0, fill: 'both', easing: E }
		);
	else if (a.t === 'pulse')
		el.animate(
			[
				{ transform: 'scale(1)', opacity: 0.6 },
				{ transform: 'scale(2.6)', opacity: 0 }
			],
			{ duration: 2000, easing: 'ease-out', ...inf }
		);
	else if (a.t === 'dots')
		el.animate([{ opacity: 0.25 }, { opacity: 1 }, { opacity: 0.25 }], { duration: 1100, ...inf });
	else if (a.t === 'ring')
		el.animate([{ strokeDashoffset: 326.7 }, { strokeDashoffset: 326.7 * (1 - a.target) }], {
			duration: 1600,
			easing: 'cubic-bezier(.3,.7,.1,1)',
			fill: 'both'
		});
	else if (a.t === 'draw' && el instanceof SVGPathElement) {
		const l = el.getTotalLength();
		el.animate(
			[
				{ strokeDasharray: `${l} ${l}`, strokeDashoffset: l },
				{ strokeDasharray: `${l} ${l}`, strokeDashoffset: 0 }
			],
			{ duration: 1600, easing: 'cubic-bezier(.3,.7,.1,1)' }
		);
	}
}
