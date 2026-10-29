/**
 * La marca es "Tu Local Web" (tulocalweb.com.ar, desde el 28/10/2026). El handoff de design/ sigue
 * diciendo "Mostrador": los scripts que generan código desde él pasan su salida por acá. Los nombres
 * de archivo del handoff (design/Mostrador v4.dc.html, OG Mostrador.dc.html) no se tocan. La
 * mascota pasa a ser Tincho, el carpincho (29/10/2026): el nombre cambia; las imágenes, después.
 */
export function conMarca(s: string): string {
	return s
		.replace(/Abrí tu mostrador/g, 'Abrí Tu Local Web')
		.replace(/armando tu mostrador/g, 'armando Tu Local Web')
		.replace(/alt="mostrador"/g, 'alt="Tu Local Web"')
		.replace(/mostradorapp\.com\.ar/g, 'tulocalweb.com.ar')
		.replace(
			/(?<![A-Za-z_@/.-])Mostrador(?![A-Za-z_0-9])(?!(?: [\w-]+)?\.dc\.html)/g,
			'Tu Local Web'
		)
		.replace(/(?<![A-Za-z_/.-])Hornero(?![A-Za-z_0-9])(?!\.svelte)/g, 'Tincho');
}
