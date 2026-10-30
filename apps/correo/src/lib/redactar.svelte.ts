/**
 * Lo que se está escribiendo: las ventanas de Redactar (varias a la vez, minimizables, abajo a la
 * derecha) y la respuesta o el reenvío que se escribe adentro del correo abierto.
 */
export type AdjuntoNuevo = { blobId: string; nombre: string; tipo: string; tamano: number };

export type Borrador = {
	clave: number;
	modo: 'nuevo' | 'responder' | 'reenviar';
	cuenta: string;
	para: string[];
	cc: string[];
	asunto: string;
	texto: string;
	adjuntos: AdjuntoNuevo[];
	/** El correo al que responde (id de la bandeja). */
	responde: string | null;
	/** Lo citado (respuesta) o el encabezado del reenviado: va al final del texto al mandar. */
	cita: string;
	minimizada: boolean;
	enviando: boolean;
	error: string;
};

let siguiente = 1;
export const ventanas = $state<Borrador[]>([]);

export function nuevo(o: Partial<Borrador> & { cuenta: string }): Borrador {
	return {
		clave: siguiente++,
		modo: 'nuevo',
		para: [],
		cc: [],
		asunto: '',
		texto: '',
		adjuntos: [],
		responde: null,
		cita: '',
		minimizada: false,
		enviando: false,
		error: '',
		...o
	};
}

export function abrirVentana(o: Partial<Borrador> & { cuenta: string }) {
	// Las otras se minimizan, como en el diseño (una abierta, las demás como pestañas).
	for (const v of ventanas) v.minimizada = true;
	ventanas.push(nuevo(o));
}

export function cerrarVentana(clave: number) {
	const i = ventanas.findIndex((v) => v.clave === clave);
	if (i >= 0) ventanas.splice(i, 1);
}

/** Menciona un adjunto pero no hay ninguno (el aviso del diseño). */
export const faltaAdjunto = (b: Borrador) => !b.adjuntos.length && /\badjunt[oaé]/i.test(b.texto);

export async function subir(b: Borrador, archivos: FileList | File[]) {
	b.error = '';
	for (const f of archivos) {
		const fd = new FormData();
		fd.set('archivo', f);
		const r = await fetch(`/subir?cuenta=${encodeURIComponent(b.cuenta)}`, {
			method: 'POST',
			body: fd
		});
		if (!r.ok) {
			b.error = (await r.json().catch(() => null))?.message ?? 'No se pudo subir el archivo.';
			return;
		}
		b.adjuntos.push(await r.json());
	}
}

export async function mandar(b: Borrador, programado: string | null = null): Promise<boolean> {
	b.enviando = true;
	b.error = '';
	const r = await fetch('/enviar', {
		method: 'POST',
		headers: { 'content-type': 'application/json' },
		body: JSON.stringify({
			cuenta: b.cuenta,
			para: b.para,
			cc: b.cc,
			asunto: b.asunto,
			texto: b.cita ? `${b.texto}\n\n${b.cita}` : b.texto,
			adjuntos: b.adjuntos,
			responde: b.modo === 'responder' ? b.responde : null,
			programado
		})
	});
	b.enviando = false;
	if (!r.ok) {
		b.error = (await r.json().catch(() => null))?.error ?? 'No se pudo mandar.';
		return false;
	}
	return true;
}
