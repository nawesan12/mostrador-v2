/**
 * Formularios públicos del sitio contra la API Go (DATA_SOURCE=api): botón de arrepentimiento,
 * consulta de ventas, postulación de partners y novedades. Sin sesión, pero por el BFF para que Go
 * vea la IP real del visitante (los límites son por IP y por email).
 */
import { fail } from '@sveltejs/kit';
import * as v from 'valibot';
import type { components } from '@mostrador/contrato/esquema';
import { errores } from './auth/esquemas.ts';
import { ClienteBff, ErrorApi } from './bff.ts';
import { formulariosDe, usaSandbox } from './datos/index.ts';

type S = components['schemas'];

export class FormulariosApi extends ClienteBff {
	arrepentimiento(p: S['PedidoArrepentimiento']) {
		return this.pedir<{ codigo: string }>('POST', '/v1/arrepentimientos', p);
	}
	consultar(c: S['Consulta']) {
		return this.pedir<void>('POST', '/v1/contacto', c);
	}
	postular(p: S['Postulacion']) {
		return this.pedir<void>('POST', '/v1/partners/postulaciones', p);
	}
	suscribir(email: string) {
		return this.pedir<void>('POST', '/v1/novedades/suscripciones', { email });
	}
	baja(token: string) {
		return this.pedir<void>('POST', '/v1/novedades/baja', { token });
	}
}

/**
 * Procesa un formulario público en una action: valida con el mismo esquema que Go y lo manda a la
 * API. En la demo (sandbox) no hay a quién mandárselo: se valida igual y se responde como si
 * hubiera salido, sin guardar nada (docs/decisiones.md §14).
 */
export async function procesar<E extends v.GenericSchema>(
	event: Parameters<typeof formulariosDe>[0],
	esquema: E,
	entrada: Record<string, unknown>,
	enviar: (api: FormulariosApi, datos: v.InferOutput<E>) => Promise<unknown>
) {
	const r = v.safeParse(esquema, entrada);
	if (!r.success) return fail(400, { ...entrada, errores: errores(r.issues) });
	if (usaSandbox()) return { enviado: true as const };
	try {
		await enviar(formulariosDe(event), r.output);
		return { enviado: true as const };
	} catch (e) {
		if (!(e instanceof ErrorApi)) throw e;
		const errs = Object.keys(e.campos).length ? e.campos : { form: e.message };
		return fail(e.estado >= 500 ? 503 : e.estado, { ...entrada, errores: errs });
	}
}
