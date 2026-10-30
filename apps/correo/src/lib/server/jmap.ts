/**
 * Cliente JMAP (RFC 8620/8621) contra Stalwart, por la red interna (CORREO_JMAP). Cada persona
 * entra con su casilla; las compartidas de sus grupos llegan como otras cuentas de la sesión, y
 * la bandeja las junta: "Entrada" es la de todas, y elegir una casilla filtra a esa cuenta.
 */
import type { Credenciales } from './sesion.ts';

const MAIL = 'urn:ietf:params:jmap:mail';
const USING = ['urn:ietf:params:jmap:core', MAIL];
const SUBMISSION = 'urn:ietf:params:jmap:submission';
const QUOTA = 'urn:ietf:params:jmap:quota';

export const base = () => (process.env.CORREO_JMAP ?? 'http://127.0.0.1:8080').replace(/\/$/, '');

export class ErrorJmap extends Error {
	constructor(
		message: string,
		readonly estado = 502
	) {
		super(message);
	}
}

type Llamada = [string, Record<string, unknown>, string];
type Respuesta = [string, Record<string, unknown>, string];

export type Cuenta = { id: string; email: string; personal: boolean };

type Sesion = {
	accounts: Record<
		string,
		{ name: string; isPersonal: boolean; accountCapabilities: Record<string, unknown> }
	>;
	primaryAccounts: Record<string, string>;
	downloadUrl: string;
	uploadUrl: string;
};

/** Stalwart atiende hasta 4 pedidos a la vez por persona (maxConcurrentRequests): uno queda libre. */
const SIMULTANEOS = 3;
/** Y hasta 16 llamadas por pedido (maxCallsInRequest). */
export const MAX_LLAMADAS = 16;

export class Jmap {
	private sesion?: Promise<Sesion>;
	private enCurso = 0;
	private espera: (() => void)[] = [];
	constructor(readonly cred: Credenciales) {}

	private async turno<T>(f: () => Promise<T>): Promise<T> {
		if (this.enCurso >= SIMULTANEOS) await new Promise<void>((ok) => this.espera.push(ok));
		this.enCurso++;
		try {
			return await f();
		} finally {
			this.enCurso--;
			this.espera.shift()?.();
		}
	}

	private auth() {
		return 'Basic ' + Buffer.from(`${this.cred.email}:${this.cred.clave}`).toString('base64');
	}

	async pedir(path: string, init: RequestInit = {}): Promise<Response> {
		const r = await this.turno(() =>
			fetch(base() + path, {
				...init,
				headers: { ...(init.headers as Record<string, string>), authorization: this.auth() }
			})
		);
		if (r.status === 401) throw new ErrorJmap('La casilla o la clave no son correctas.', 401);
		return r;
	}

	leerSesion(): Promise<Sesion> {
		this.sesion ??= this.pedir('/jmap/session').then(async (r) => {
			if (!r.ok) throw new ErrorJmap(`JMAP sesión ${r.status}: ${(await r.text()).slice(0, 300)}`);
			return (await r.json()) as Sesion;
		});
		return this.sesion;
	}

	async cuentas(): Promise<Cuenta[]> {
		const s = await this.leerSesion();
		const principal = s.primaryAccounts[MAIL];
		return Object.entries(s.accounts)
			.filter(([, a]) => MAIL in a.accountCapabilities)
			.map(([id, a]) => ({ id, email: a.name, personal: a.isPersonal }))
			.sort((a, b) =>
				a.id === principal ? -1 : b.id === principal ? 1 : a.email.localeCompare(b.email)
			);
	}

	/** Varias llamadas en un solo pedido; devuelve las respuestas por su id. */
	async llamar(
		llamadas: Llamada[],
		extra: string[] = []
	): Promise<Map<string, Record<string, unknown>>> {
		const r = await this.pedir('/jmap/', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ using: [...USING, ...extra], methodCalls: llamadas })
		});
		if (!r.ok) throw new ErrorJmap(`JMAP ${r.status}: ${(await r.text()).slice(0, 300)}`);
		const { methodResponses } = (await r.json()) as { methodResponses: Respuesta[] };
		const out = new Map<string, Record<string, unknown>>();
		for (const [nombre, args, id] of methodResponses) {
			if (nombre === 'error')
				throw new ErrorJmap(`JMAP ${String(args.type)}: ${String(args.description ?? '')}`);
			out.set(id, args);
		}
		return out;
	}

	/** El archivo de un adjunto (o de una imagen cid:), tal como lo guarda Stalwart. */
	async descargar(cuenta: string, blob: string, nombre: string, tipo: string): Promise<Response> {
		const s = await this.leerSesion();
		const u = new URL(
			s.downloadUrl
				.replace('{accountId}', encodeURIComponent(cuenta))
				.replace('{blobId}', encodeURIComponent(blob))
				.replace('{name}', encodeURIComponent(nombre))
				.replace('{type}', encodeURIComponent(tipo))
		);
		return this.pedir(u.pathname + u.search);
	}

	/** Sube un archivo (para adjuntarlo a un correo nuevo) y devuelve su blobId. */
	async subir(
		cuenta: string,
		cuerpo: Blob
	): Promise<{ blobId: string; size: number; type: string }> {
		const s = await this.leerSesion();
		const u = new URL(s.uploadUrl.replace('{accountId}', encodeURIComponent(cuenta)));
		const r = await this.pedir(u.pathname, {
			method: 'POST',
			headers: { 'content-type': cuerpo.type || 'application/octet-stream' },
			body: cuerpo
		});
		if (!r.ok) throw new ErrorJmap(`No se pudo subir el archivo (${r.status}).`);
		return (await r.json()) as { blobId: string; size: number; type: string };
	}
}

export { MAIL, SUBMISSION, QUOTA };
