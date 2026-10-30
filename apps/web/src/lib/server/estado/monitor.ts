/**
 * Monitor de /estado. Lo corre /api/internal/cron, que el VPS llama cada hora: un chequeo por componente, todos
 * en paralelo y con 10 s de tope (unos 170 pedidos por día en total). Guarda un resumen por día,
 * el último resultado de cada uno y abre o cierra incidentes solo.
 *
 * Qué se mira de cada componente:
 * - tiendas y panel: el sitio (por service binding: un worker no puede pedirle por URL pública a
 *   otro de la misma zona). Tienen que responder 200.
 * - pagos, whatsapp, arca y mercadolibre: que el proveedor responda (cualquier cosa menor a 500;
 *   sin credenciales, un 401 o 400 también dice que está arriba).
 * - api: /salud/listo de la API Go, sólo si hay API_URL (en la demo no hay API: queda sin datos).
 */
import { COMPONENTES, diaAR, nombreDe, type Componente } from './componentes.ts';

export type Resultado = {
	componente: Componente;
	ok: boolean;
	latencia: number;
	detalle: string;
};

type Chequeo = { componente: Componente; url: string; sitio?: boolean; exige200?: boolean };

export function chequeos(sitio: string, apiUrl: string): Chequeo[] {
	const lista: Chequeo[] = [
		{ componente: 'tiendas', url: `${sitio}/t/tienda-lumbre`, sitio: true, exige200: true },
		{ componente: 'panel', url: `${sitio}/ingresar`, sitio: true, exige200: true },
		{ componente: 'pagos', url: 'https://api.mercadopago.com/v1/payment_methods' },
		{ componente: 'whatsapp', url: 'https://graph.facebook.com/v21.0/' },
		{
			componente: 'arca',
			url: 'https://servicios1.afip.gov.ar/wsfev1/service.asmx',
			exige200: true
		},
		// Mercado Libre responde 403 a los pedidos sin token desde datacenters: igual dice que está arriba.
		{ componente: 'mercadolibre', url: 'https://api.mercadolibre.com/sites/MLA' }
	];
	if (apiUrl) lista.push({ componente: 'api', url: `${apiUrl}/salud/listo`, exige200: true });
	return lista;
}

type Pedir = (url: string, init: RequestInit) => Promise<Response>;

/** Corre los chequeos. `sitio` pide al propio sitio (service binding); `externo`, al resto. */
export async function chequear(
	lista: Chequeo[],
	pedir: { sitio: Pedir; externo: Pedir },
	reloj: () => number = Date.now
): Promise<Resultado[]> {
	return Promise.all(
		lista.map(async (c): Promise<Resultado> => {
			const t0 = reloj();
			try {
				const r = await (c.sitio ? pedir.sitio : pedir.externo)(c.url, {
					method: 'GET',
					redirect: 'manual',
					headers: { 'user-agent': 'mostrador-estado/1 (+https://tulocalweb.com.ar/estado)' },
					signal: AbortSignal.timeout(10_000)
				});
				await r.body?.cancel();
				const ok = c.exige200 ? r.status === 200 : r.status < 500;
				return {
					componente: c.componente,
					ok,
					latencia: reloj() - t0,
					detalle: `HTTP ${r.status}`
				};
			} catch (e) {
				const tiempo = e instanceof Error && e.name === 'TimeoutError';
				return {
					componente: c.componente,
					ok: false,
					latencia: reloj() - t0,
					detalle: tiempo ? 'No respondió en 10 segundos' : 'No se pudo conectar'
				};
			}
		})
	);
}

export type Cambio = {
	tipo: 'abre' | 'cierra';
	componente: Componente;
	titulo: string;
	detalle: string;
};

const hora = (ms: number) =>
	new Intl.DateTimeFormat('es-AR', {
		hour: '2-digit',
		minute: '2-digit',
		timeZone: 'America/Argentina/Buenos_Aires'
	}).format(ms);

/**
 * Guarda los resultados y abre o cierra incidentes. Devuelve los cambios (para avisar a los
 * suscriptos). Un incidente se abre con dos fallas seguidas, para no alarmar por un pico suelto.
 */
export async function registrar(db: D1Database, resultados: Resultado[], ahora: number) {
	const dia = diaAR(ahora);
	const previos = new Map(
		(
			await db
				.prepare('SELECT componente, fallas_seguidas FROM ultimo')
				.all<{ componente: string; fallas_seguidas: number }>()
		).results.map((f) => [f.componente, f.fallas_seguidas])
	);
	const abiertos = new Map(
		(
			await db
				.prepare('SELECT id, componente, inicio, detalle FROM incidentes WHERE fin IS NULL')
				.all<{ id: number; componente: string; inicio: number; detalle: string }>()
		).results.map((f) => [f.componente, f])
	);
	const sentencias: D1PreparedStatement[] = [];
	const cambios: Cambio[] = [];
	for (const r of resultados) {
		const seguidas = r.ok ? 0 : (previos.get(r.componente) ?? 0) + 1;
		sentencias.push(
			db
				.prepare(
					`INSERT INTO dias (componente, dia, chequeos, fallas) VALUES (?1, ?2, 1, ?3)
					 ON CONFLICT (componente, dia) DO UPDATE SET chequeos = chequeos + 1, fallas = fallas + ?3`
				)
				.bind(r.componente, dia, r.ok ? 0 : 1),
			db
				.prepare(
					`INSERT INTO ultimo (componente, ok, fallas_seguidas, latencia_ms, detalle, chequeado_en)
					 VALUES (?1, ?2, ?3, ?4, ?5, ?6)
					 ON CONFLICT (componente) DO UPDATE SET ok = ?2, fallas_seguidas = ?3, latencia_ms = ?4,
					 detalle = ?5, chequeado_en = ?6`
				)
				.bind(r.componente, r.ok ? 1 : 0, seguidas, r.latencia, r.detalle, ahora)
		);
		const abierto = abiertos.get(r.componente);
		if (!r.ok && seguidas >= 2 && !abierto) {
			const titulo = `Problemas en ${nombreDe(r.componente)}`;
			const detalle = `Falló en dos chequeos seguidos (${r.detalle}). Lo estamos mirando.`;
			sentencias.push(
				db
					.prepare(
						'INSERT INTO incidentes (componente, titulo, detalle, inicio) VALUES (?1, ?2, ?3, ?4)'
					)
					.bind(r.componente, titulo, detalle, ahora)
			);
			cambios.push({ tipo: 'abre', componente: r.componente, titulo, detalle });
		}
		if (r.ok && abierto) {
			const minutos = Math.max(1, Math.round((ahora - abierto.inicio) / 60_000));
			const detalle = `${abierto.detalle.replace(/ Lo estamos mirando\.$/, '')} Volvió a funcionar a las ${hora(ahora)}, después de ${minutos} minutos.`;
			sentencias.push(
				db
					.prepare('UPDATE incidentes SET fin = ?1, detalle = ?2 WHERE id = ?3')
					.bind(ahora, detalle, abierto.id)
			);
			cambios.push({
				tipo: 'cierra',
				componente: r.componente,
				titulo: `${nombreDe(r.componente)} funciona de nuevo`,
				detalle
			});
		}
	}
	await db.batch(sentencias);
	return cambios;
}

export { COMPONENTES };
