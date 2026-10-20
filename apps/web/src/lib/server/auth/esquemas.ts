import * as v from 'valibot';

const email = v.pipe(
	v.string(),
	v.trim(),
	v.toLowerCase(),
	v.nonEmpty('Escribí tu email.'),
	v.email('Ese email no parece válido.'),
	v.maxLength(254, 'Ese email es demasiado largo.')
);

export const Ingreso = v.object({
	email,
	clave: v.pipe(v.string(), v.nonEmpty('Escribí tu contraseña.'), v.maxLength(200)),
	recordarme: v.optional(v.picklist(['on']))
});

export const Registro = v.object({
	nombre: v.pipe(v.string(), v.trim(), v.nonEmpty('¿Cómo te llamás?'), v.maxLength(80)),
	tienda: v.pipe(v.string(), v.trim(), v.nonEmpty('Poné el nombre de tu tienda.'), v.maxLength(80)),
	email,
	clave: v.pipe(
		v.string(),
		v.minLength(8, 'La contraseña tiene que tener al menos 8 caracteres.'),
		v.maxLength(200)
	),
	// Un checkbox sin tildar no viaja en el formulario: falta = no aceptó.
	terminos: v.pipe(
		v.optional(v.string(), ''),
		v.check((x) => x === 'on', 'Para crear la tienda tenés que aceptar los términos.')
	)
});

const telefono = v.pipe(
	v.string(),
	v.trim(),
	v.transform((s) => s.replace(/[^\d+]/g, '')),
	v.regex(/^\+?\d{8,15}$/, 'Escribí el número con código de área, por ejemplo 11 2345 6789.')
);

export const Recuperar = v.variant('canal', [
	v.object({ canal: v.literal('email'), email }),
	v.object({ canal: v.literal('whatsapp'), telefono })
]);

/** Entrar con un link al WhatsApp de la cuenta. */
export const Enlace = v.object({ telefono });

/** Proveedores de ingreso con OAuth (los mismos que la API). */
export const PROVEEDORES = ['google', 'mercadopago', 'mercadolibre'] as const;
export type ProveedorOAuth = (typeof PROVEEDORES)[number];

export const Restablecer = v.object({
	token: v.pipe(v.string(), v.nonEmpty('El link no es válido. Pedí uno nuevo.'), v.maxLength(200)),
	clave: v.pipe(
		v.string(),
		v.minLength(8, 'La contraseña tiene que tener al menos 8 caracteres.'),
		v.maxLength(200)
	)
});

/** Errores de valibot como { campo: mensaje } para mostrar debajo de cada input. */
export function errores(issues: v.BaseIssue<unknown>[]) {
	const out: Record<string, string> = {};
	for (const i of issues) {
		const k = v.getDotPath(i) ?? 'form';
		out[k] ??= i.message;
	}
	return out;
}
