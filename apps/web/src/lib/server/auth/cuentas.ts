/**
 * Cuentas contra la API Go (DATA_SOURCE=api). El navegador nunca ve el token de Go: lo guarda este
 * servidor en una cookie httpOnly.
 */
import type { components } from '@mostrador/contrato/esquema';
import { ClienteBff } from '../bff.ts';

export { ErrorApi as ErrorCuenta } from '../bff.ts';
export type Sesion = components['schemas']['Sesion'];
export type DatosSesion = components['schemas']['DatosSesion'];

export class CuentasApi extends ClienteBff {
	registrar(a: components['schemas']['Alta']) {
		return this.pedir<Sesion>('POST', '/v1/auth/registro', a);
	}
	ingresar(email: string, clave: string, recordarme: boolean) {
		return this.pedir<Sesion>('POST', '/v1/auth/ingresar', { email, clave, recordarme });
	}
	sesion(token: string) {
		return this.pedir<DatosSesion>('GET', '/v1/auth/sesion', undefined, token);
	}
	salir(token: string) {
		return this.pedir<void>('POST', '/v1/auth/salir', undefined, token);
	}
	/** El link de contraseña nueva, por email o al WhatsApp de la cuenta. */
	recuperar(a: { email: string } | { telefono: string }) {
		return this.pedir<void>('POST', '/v1/auth/recuperar', a);
	}
	/** Qué ingresos están configurados en la API. */
	proveedores() {
		return this.pedir<{ oauth: string[]; clerk: boolean; whatsapp: boolean }>(
			'GET',
			'/v1/auth/proveedores'
		);
	}
	/** URL del proveedor a la que se manda a la persona (con state y PKCE guardados en Go). */
	iniciarOAuth(proveedor: string) {
		return this.pedir<{ url: string }>(
			'POST',
			`/v1/auth/oauth/${encodeURIComponent(proveedor)}/inicio`,
			{}
		);
	}
	volverOAuth(proveedor: string, codigo: string, estado: string) {
		return this.pedir<Sesion & { nueva: boolean }>(
			'POST',
			`/v1/auth/oauth/${encodeURIComponent(proveedor)}/vuelta`,
			{ codigo, estado }
		);
	}
	/** Entra con el token de sesión de Clerk (Google, Apple): Go lo verifica con la instancia. */
	ingresarConClerk(token: string) {
		return this.pedir<Sesion & { nueva: boolean }>('POST', '/v1/auth/clerk', { token });
	}
	/** Link para entrar sin contraseña, al WhatsApp de la cuenta. */
	pedirEnlace(telefono: string) {
		return this.pedir<void>('POST', '/v1/auth/enlace', { telefono });
	}
	usarEnlace(token: string) {
		return this.pedir<Sesion>('POST', '/v1/auth/enlace/usar', { token });
	}
	/** Si el link de contraseña nueva sirve todavía (sin usarlo). */
	verificarRestablecer(token: string) {
		return this.pedir<void>('POST', '/v1/auth/restablecer/verificar', { token });
	}
	restablecer(token: string, clave: string) {
		return this.pedir<void>('POST', '/v1/auth/restablecer', { token, clave });
	}
}
