declare global {
	namespace App {
		interface Locals {
			/** Sandbox de la demo del visitante (sólo en /demo y /panel, para no leer D1 en la landing). */
			sandbox: import('#lib/server/sandbox/sandbox.ts').Sandbox | null;
			/** Token de la sesión de la API Go (DATA_SOURCE=api), leído de la cookie httpOnly. */
			token: string | null;
		}
		interface Error {
			message: string;
			code?: string;
		}
		interface Platform {
			/** El contexto del Worker (adapter-cloudflare): waitUntil para terminar después de responder. */
			ctx?: ExecutionContext;
		}
	}

	namespace Cloudflare {
		interface Env {
			PUBLIC_SITE_URL: string;
			/** sandbox: demo con datos por visitante (D1). api: la API Go del VPS. */
			DATA_SOURCE: 'sandbox' | 'api';
			/** URL de la API Go (DATA_SOURCE=api). */
			API_URL: string;
			/** Clave AES-256 en base64url (secreto). Cifra credenciales de integraciones. */
			ENCRYPTION_KEY: string;
			/** Clave que comparte el BFF con la API Go (X-Mostrador-BFF). Secreto. */
			CLAVE_BFF?: string;
			/** Secreto compartido con mostrador-jobs para /api/internal/*. */
			INTERNAL_SECRET: string;
			/** Workers AI (Hornero). */
			AI?: Ai;
			/** workers-ai en producción; simulado en desarrollo y pruebas (no llama al modelo). */
			HORNERO?: 'workers-ai' | 'simulado';
			/** Bot de Telegram de la plataforma (@mostradorhornero_bot). Secreto. */
			TELEGRAM_BOT_TOKEN?: string;
			/** Lo manda Telegram en X-Telegram-Bot-Api-Secret-Token. Secreto (lo generamos nosotros). */
			TELEGRAM_WEBHOOK_SECRET?: string;
		}
	}
}

export {};
