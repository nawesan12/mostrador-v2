declare global {
	namespace App {
		interface Error {
			message: string;
			code?: string;
		}
	}

	namespace Cloudflare {
		interface Env {
			PUBLIC_SITE_URL: string;
			/** sandbox: demo con datos por visitante (D1). api: la API Go del VPS. */
			DATA_SOURCE: 'sandbox' | 'api';
			/** Clave AES-256 en base64url (secreto). Cifra credenciales de integraciones. */
			ENCRYPTION_KEY: string;
			/** Secreto compartido con mostrador-jobs para /api/internal/*. */
			INTERNAL_SECRET: string;
		}
	}
}

export {};
