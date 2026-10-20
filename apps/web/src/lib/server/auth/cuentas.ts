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
	recuperar(email: string) {
		return this.pedir<void>('POST', '/v1/auth/recuperar', { email });
	}
	restablecer(token: string, clave: string) {
		return this.pedir<void>('POST', '/v1/auth/restablecer', { token, clave });
	}
}
