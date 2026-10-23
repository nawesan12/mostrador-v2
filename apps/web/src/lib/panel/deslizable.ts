/**
 * Filas que se deslizan de costado (vistas, pestañas, chips): un degradé en el borde avisa que hay
 * más, y desaparece al llegar al final. Marca data-mas="der", "izq" o "ambos"; el CSS está en
 * app.css ([data-deslizable]).
 */
export function deslizable(el: HTMLElement) {
	el.dataset.deslizable = '';
	const medir = () => {
		const der = el.scrollLeft + el.clientWidth < el.scrollWidth - 2;
		const izq = el.scrollLeft > 2;
		el.dataset.mas = der && izq ? 'ambos' : der ? 'der' : izq ? 'izq' : '';
	};
	medir();
	el.addEventListener('scroll', medir, { passive: true });
	const ro = new ResizeObserver(medir);
	ro.observe(el);
	return {
		destroy() {
			el.removeEventListener('scroll', medir);
			ro.disconnect();
		}
	};
}
