declare global {
	namespace App {
		interface Locals {
			/** Sandbox de la demo del visitante (sólo en /demo y /panel, para no leer D1 en la landing). */
			sandbox: import('#lib/server/sandbox/sandbox.ts').Sandbox | null;
			/** Token de la sesión de la API Go (DATA_SOURCE=api), leído de la cookie httpOnly. */
			token: string | null;
			/** Token de la sesión del equipo (/staff, DATA_SOURCE=api), leído de su cookie. */
			equipo: string | null;
			/** El panel abierto por alguien del equipo en sólo lectura ("ver la tienda"). */
			mirada: boolean;
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
			/** Workers AI (Tincho). */
			AI?: Ai;
			/** workers-ai en producción; simulado en desarrollo y pruebas (no llama al modelo). */
			HORNERO?: 'workers-ai' | 'simulado';
			/** App de socio de Tiendanube (importar con OAuth en la demo). Secretos de wrangler. */
			TIENDANUBE_CLIENT_ID?: string;
			TIENDANUBE_SECRETO?: string;
			/** Sólo en el preview de las pruebas: un Tiendanube falso (http://127.0.0.1:puerto). */
			TIENDANUBE_LOCAL?: string;
			/** Bot de Telegram de la plataforma (@mostradorhornero_bot). Secreto. */
			TELEGRAM_BOT_TOKEN?: string;
			/** Lo manda Telegram en X-Telegram-Bot-Api-Secret-Token. Secreto (lo generamos nosotros). */
			TELEGRAM_WEBHOOK_SECRET?: string;
		}
	}
}

export {};
