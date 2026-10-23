/**
 * Exportar piezas de marca: PNG en el navegador (modern-screenshot), .zip con todos los PNG de una
 * familia (fflate) y PDF con la impresión del navegador, a la medida real y con sangrado.
 */
import { zipSync } from 'fflate';
import { domToBlob } from 'modern-screenshot';

/** 300 ppp: lo que pide una imprenta. Las piezas en px (redes, OG) salen a su tamaño exacto. */
export const ESCALA_IMPRENTA = 300 / 96;

/** Lo que es de la vista (sombra, escala) no va en el archivo. */
const SIN_MARCO: Partial<CSSStyleDeclaration> = {
	boxShadow: 'none',
	transform: 'none',
	margin: '0'
};

export async function aPng(nodo: HTMLElement, escala: number): Promise<Blob> {
	await document.fonts.ready;
	const blob = await domToBlob(nodo, {
		width: nodo.offsetWidth,
		height: nodo.offsetHeight,
		scale: escala,
		type: 'image/png',
		style: SIN_MARCO,
		timeout: 15_000
	});
	if (!blob) throw new Error('No se pudo generar la imagen');
	return blob;
}

export function descargar(blob: Blob, nombre: string) {
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = nombre;
	document.body.appendChild(a);
	a.click();
	a.remove();
	setTimeout(() => URL.revokeObjectURL(url), 10_000);
}

/** Los PNG ya vienen comprimidos: se guardan tal cual (nivel 0), que es instantáneo. */
export async function zip(archivos: { nombre: string; blob: Blob }[]): Promise<Blob> {
	const entradas: Record<string, [Uint8Array, { level: 0 }]> = {};
	for (const a of archivos)
		entradas[a.nombre] = [new Uint8Array(await a.blob.arrayBuffer()), { level: 0 }];
	return new Blob([new Uint8Array(zipSync(entradas))], { type: 'application/zip' });
}

/**
 * Imprime sólo esta pieza en una hoja de su medida (más el sangrado), para "Guardar como PDF".
 * `mm` es la medida del diseño; `escala` la lleva a la real (el roll-up está dibujado a 1:5).
 * El sangrado repite el fondo de la pieza hacia afuera de la línea de corte.
 */
export async function imprimir(
	nodo: HTMLElement,
	{
		mm: [w, h],
		sangrado,
		escala,
		titulo
	}: { mm: [number, number]; sangrado: number; escala: number; titulo: string }
) {
	const ancho = w * escala + sangrado * 2;
	const alto = h * escala + sangrado * 2;
	const fondo = getComputedStyle(nodo);

	const hoja = document.createElement('div');
	hoja.id = 'marca-imprimir';
	hoja.style.cssText = `position:relative;overflow:hidden;width:${ancho}mm;height:${alto}mm;background-color:${fondo.backgroundColor};background-image:${fondo.backgroundImage};background-size:${fondo.backgroundSize};background-position:center;color:${fondo.color};font-family:${fondo.fontFamily};font-variant-numeric:${fondo.fontVariantNumeric}`;
	const copia = nodo.cloneNode(true) as HTMLElement;
	copia.style.position = 'absolute';
	copia.style.left = `${sangrado}mm`;
	copia.style.top = `${sangrado}mm`;
	copia.style.margin = '0';
	copia.style.boxShadow = 'none';
	copia.style.transformOrigin = '0 0';
	copia.style.transform = escala === 1 ? 'none' : `scale(${escala})`;
	// Con sangrado, el corte lo hace la imprenta: la hoja va recta.
	if (sangrado) copia.style.borderRadius = '0';
	hoja.appendChild(copia);

	const estilo = document.createElement('style');
	estilo.textContent = `@page{size:${ancho}mm ${alto}mm;margin:0}
@media screen{#marca-imprimir{display:none}}
@media print{html,body{margin:0!important;padding:0!important;background:none!important}
body>:not(#marca-imprimir){display:none!important}
#marca-imprimir,#marca-imprimir *{-webkit-print-color-adjust:exact;print-color-adjust:exact}}`;
	document.head.appendChild(estilo);
	document.body.appendChild(hoja);

	const antes = document.title;
	let listo = false;
	const limpiar = () => {
		if (listo) return;
		listo = true;
		hoja.remove();
		estilo.remove();
		document.title = antes;
	};
	try {
		// Una imagen que no carga (o no avisa) no traba la impresión: a los 5 s se imprime igual.
		await Promise.race([
			Promise.all(
				[...hoja.querySelectorAll('img')].map((i) =>
					i.complete
						? null
						: new Promise((r) => {
								i.addEventListener('load', r, { once: true });
								i.addEventListener('error', r, { once: true });
							})
				)
			),
			new Promise((r) => setTimeout(r, 5000))
		]);
		document.title = titulo;
		window.addEventListener('afterprint', limpiar, { once: true });
		window.print();
	} finally {
		setTimeout(limpiar, 1000);
	}
}
