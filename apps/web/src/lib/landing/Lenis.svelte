<script lang="ts">
	import Lenis from 'lenis';
	import { onMount } from 'svelte';
	import { calma } from '#lib/anim.ts';
	import { getLanding } from './estado.svelte.ts';

	/**
	 * Smooth scroll de la landing (handoff: Lenis, duration 1.15, easing exponencial, anclas con
	 * offset −80) y parallax de los elementos [data-px] / [data-pxi]. Se frena con el menú o el modal
	 * de plantillas abiertos. Con prefers-reduced-motion no se activa nada.
	 */
	const l = getLanding();
	let lenis: Lenis | null = null;

	function parallax() {
		const vh = window.innerHeight;
		// En el celular las piezas flotantes no se desplazan: con poco ancho pisan lo que tienen al lado.
		const flota = window.innerWidth > 640;
		document.querySelectorAll<HTMLElement>('[data-px]').forEach((el) => {
			if (!flota) return void (el.style.translate = '');
			const r = (el.parentElement ?? el).getBoundingClientRect();
			if (r.bottom < -200 || r.top > vh + 200) return;
			const c = r.top + r.height / 2 - vh / 2;
			el.style.translate = `0 ${(c * Number(el.dataset.px)).toFixed(1)}px`;
		});
		document.querySelectorAll<HTMLElement>('[data-pxi]').forEach((el) => {
			const r = (el.parentElement ?? el).getBoundingClientRect();
			if (r.bottom < 0 || r.top > vh) return;
			const p = (r.top + r.height / 2 - vh / 2) / vh;
			el.style.scale = '1.16';
			el.style.translate = `0 ${(p * -9).toFixed(2)}%`;
		});
	}

	onMount(() => {
		if (calma()) return;
		lenis = new Lenis({
			duration: 1.15,
			easing: (t) => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
			smoothWheel: true
		});
		lenis.on('scroll', parallax);
		let raf = 0;
		const loop = (t: number) => {
			lenis?.raf(t);
			raf = requestAnimationFrame(loop);
		};
		raf = requestAnimationFrame(loop);
		const ancla = (e: MouseEvent) => {
			const a = (e.target as Element | null)?.closest?.('a[href^="#"]');
			const id = a?.getAttribute('href')?.slice(1);
			const el = id ? document.getElementById(id) : null;
			if (el && lenis) {
				e.preventDefault();
				lenis.start();
				lenis.scrollTo(el, { offset: -80, force: true });
				history.replaceState(null, '', `#${id}`);
			}
		};
		document.addEventListener('click', ancla);
		parallax();
		return () => {
			cancelAnimationFrame(raf);
			document.removeEventListener('click', ancla);
			lenis?.destroy();
			lenis = null;
		};
	});

	$effect(() => {
		const quieto = l.menu || l.tplI !== null;
		if (!lenis) return;
		if (quieto) lenis.stop();
		else lenis.start();
	});
</script>
