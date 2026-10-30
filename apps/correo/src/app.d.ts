declare global {
	namespace App {
		interface Locals {
			/** La casilla con la que se entró (de la cookie cifrada); null sin sesión. */
			cred: import('#lib/server/sesion.ts').Credenciales | null;
		}
	}
}

export {};
