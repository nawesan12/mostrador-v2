/**
 * Lo compartido por las piezas de reseñas y preguntas de la tienda (FormResena, PreguntasProducto):
 * cada plantilla las viste con sus colores, su radio y su tipografía.
 */
export type TemaOpiniones = {
	/** Texto principal. */
	texto: string;
	/** Texto secundario (ayudas, fechas). */
	suave: string;
	/** Fondo de los campos. */
	campo: string;
	/** Fondo de la tarjeta que los contiene ('transparent' si va suelto). */
	tarjeta: string;
	/** Bordes y separadores. */
	borde: string;
	/** Botón principal y estrellas elegidas. */
	acento: string;
	/** Texto sobre el acento. */
	sobreAcento: string;
	/** Color de las estrellas (si no, el acento). */
	estrella?: string;
	/** Radio de campos y botones (p. ej. '2px' o '999px' para botones de píldora). */
	radio: string;
	/** Radio de los botones, si es distinto del de los campos. */
	radioBoton?: string;
	/** Tipografía del cuerpo. */
	fuente: string;
	/** Tipografía y peso de los títulos. */
	titulos: string;
	pesoTitulos?: number;
	/** Títulos en mayúsculas espaciadas (plantillas tipo afiche). */
	mayusculas?: boolean;
	/** Color de los errores (en fondos oscuros, uno claro). */
	error?: string;
};

/** Las variables CSS que leen las piezas. */
export const varsTema = (t: TemaOpiniones) =>
	[
		`--op-tx:${t.texto}`,
		`--op-mu:${t.suave}`,
		`--op-campo:${t.campo}`,
		`--op-sf:${t.tarjeta}`,
		`--op-ln:${t.borde}`,
		`--op-ac:${t.acento}`,
		`--op-on:${t.sobreAcento}`,
		`--op-st:${t.estrella ?? t.acento}`,
		`--op-r:${t.radio}`,
		`--op-rb:${t.radioBoton ?? t.radio}`,
		`--op-f:${t.fuente}`,
		`--op-tf:${t.titulos}`,
		`--op-tw:${t.pesoTitulos ?? 600}`,
		`--op-tt:${t.mayusculas ? 'uppercase' : 'none'}`,
		`--op-ls:${t.mayusculas ? '.04em' : '-.01em'}`,
		`--op-err:${t.error ?? '#b3261e'}`
	].join(';');

/** Manda el formulario a la ruta de la tienda; devuelve el error para mostrar, o null. */
export async function enviarOpinion(
	slug: string,
	que: 'resena' | 'pregunta',
	cuerpo: Record<string, unknown>
): Promise<string | null> {
	try {
		const r = await fetch(`/t/${slug}/${que}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(cuerpo)
		});
		if (r.ok) return null;
		const e = (await r.json().catch(() => null)) as { error?: string } | null;
		return e?.error ?? 'No se pudo mandar. Probá de nuevo.';
	} catch {
		return 'No hay conexión. Probá de nuevo en un rato.';
	}
}
