/**
 * Turnos de la tienda online (/t/[tienda]/turnos y /turno/[codigo]). En la demo los atiende el
 * sandbox del visitante (o la vitrina, sólo para mirar); con DATA_SOURCE=api, Go
 * (/v1/publico/tiendas/{slug}/turnos). El navegador nunca habla con Go.
 */
import { env } from 'cloudflare:workers';
import {
	diaAr,
	horariosDelDia,
	sumarDias,
	type Horario,
	type HorarioLibre
} from '#lib/turnos/libres.ts';
import type { ProximosTurnos, TurnoPublico } from '#lib/tienda/tipos.ts';
import { avisarVenta } from '../telegram.ts';
import { ENTIDAD } from '../sandbox/turnos.ts';
import {
	buscarCambio,
	cuandoTurno,
	DatosSandbox,
	dbSandbox,
	type Sandbox
} from '../sandbox/sandbox.ts';
import type { TiendaResuelta } from './publica.ts';

const api = (ruta: string, init?: RequestInit) => fetch(`${String(env.API_URL)}${ruta}`, init);
const enc = encodeURIComponent;

async function libresApi(slug: string, servicio: string, profesional: string | null, dia: string) {
	const q = new URLSearchParams({ servicio, dia });
	if (profesional) q.set('profesional', profesional);
	const r = await api(`/v1/publico/tiendas/${enc(slug)}/turnos/libres?${q}`);
	if (!r.ok) return [];
	return (await r.json()) as HorarioLibre[];
}

/**
 * La grilla de un día para el servicio, cada horario con si está libre (los ocupados se dibujan
 * tachados). Con la API, la grilla sale del horario público y lo libre, de Go.
 */
export async function horariosDe(
	r: TiendaResuelta,
	servicio: string,
	profesional: string | null,
	dia: string
): Promise<Horario[] | null> {
	const t = r.tienda.turnos;
	if (!t) return null;
	if (r.datos) return r.datos.horariosTurnos(servicio, profesional, dia);
	const s = t.servicios.find((x) => x.id === servicio);
	if (!s) return null;
	const libres = await libresApi(r.tienda.slug, servicio, profesional, dia);
	const grilla = horariosDelDia({
		ajustes: t,
		servicio: s,
		profesionales: t.profesionales,
		turnos: [],
		profesionalId: profesional,
		dia,
		ahora: Date.now()
	});
	const ok = new Set(libres.map((h) => Date.parse(h.inicio)));
	return grilla.map((h) => ({ ...h, libre: ok.has(Date.parse(h.inicio)) }));
}

/** Los libres de hoy (portada) y el próximo libre de cada profesional (El equipo). */
export async function proximosDe(r: TiendaResuelta): Promise<ProximosTurnos | null> {
	const t = r.tienda.turnos;
	if (!t) return null;
	if (r.datos) return r.datos.proximosTurnos();
	// Con la API: hoy y mañana de cada uno, en paralelo (la portada no espera más que eso).
	const hoy = diaAr(Date.now());
	const primero = t.servicios[0];
	if (!primero) return { hoy: [], porProfesional: {} };
	const [deHoy, ...porPro] = await Promise.all([
		libresApi(r.tienda.slug, primero.id, null, hoy),
		...t.profesionales.map(async (p) => {
			const s = t.servicios.find((x) => !p.servicios.length || p.servicios.includes(x.id));
			if (!s) return [p.id, null] as const;
			for (const d of [hoy, sumarDias(hoy, 1)]) {
				const l = await libresApi(r.tienda.slug, s.id, p.id, d);
				if (l.length) return [p.id, l[0].inicio] as const;
			}
			return [p.id, null] as const;
		})
	]);
	return { hoy: deHoy.map((h) => h.inicio), porProfesional: Object.fromEntries(porPro) };
}

export type Reserva = {
	servicioId: string;
	profesionalId: string | null;
	inicio: string;
	nombre: string;
	telefono: string;
	email: string;
};

/** Reserva desde la tienda; avisa a Telegram (la notificación del panel sale de los turnos). */
export async function reservar(
	r: TiendaResuelta,
	b: Reserva,
	esperar: (p: Promise<unknown>) => void
): Promise<{ codigo: string; url?: string } | { error: string; estado: 400 | 409 }> {
	if (!r.datos || !r.sandboxId) {
		const res = await api(`/v1/publico/tiendas/${enc(r.tienda.slug)}/turnos`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(b)
		});
		if (!res.ok) {
			const e = (await res.json().catch(() => null)) as { detail?: string } | null;
			return {
				estado: res.status === 409 ? 409 : 400,
				error: e?.detail ?? 'No pudimos reservar el turno. Probá de nuevo.'
			};
		}
		return (await res.json()) as { codigo: string; url?: string };
	}
	const t = await r.datos.reservarTurno(b);
	if (t === 'ocupado') return { estado: 409, error: 'Ese horario ya no está libre. Elegí otro.' };
	if (t === 'apagado') return { estado: 409, error: 'Esta tienda no toma turnos online.' };
	if (t === 'invalido') return { estado: 400, error: 'Elegí un servicio de la lista.' };
	esperar(
		avisarVenta(
			dbSandbox(env.SANDBOX),
			r.sandboxId,
			`📅 Turno nuevo: ${t.servicio} con ${t.profesional}, ${cuandoTurno(t.inicio)}. Reservó ${t.cliente.nombre}.`
		).catch(() => {})
	);
	return { codigo: t.codigo };
}

const CODIGO = /^[A-Za-z0-9_-]{16,64}$/;

/**
 * En la demo, el sandbox dueño del turno por su código (así el link de la confirmación anda sin la
 * cookie, por ejemplo desde el calendario), si su tienda es la de ese slug.
 */
async function sandboxDelTurno(slug: string, codigo: string) {
	const db = dbSandbox(env.SANDBOX);
	const c = await buscarCambio<{ id: string }>(db, ENTIDAD, `c:${codigo}`);
	if (!c) return null;
	const datos = new DatosSandbox(db, c.sb as Sandbox);
	return (await datos.tienda()).slug === slug ? datos : null;
}

/** El turno por su código público, como lo ve quien reservó. */
export async function turnoDe(slug: string, codigo: string, sandbox: boolean) {
	if (!CODIGO.test(codigo)) return null;
	if (!sandbox) {
		const r = await api(`/v1/publico/tiendas/${enc(slug)}/turnos/${enc(codigo)}`);
		return r.ok ? ((await r.json()) as TurnoPublico) : null;
	}
	return (await sandboxDelTurno(slug, codigo))?.turnoPublico(codigo) ?? null;
}

/** Cancela desde la confirmación. 'tarde' si falta menos que la anticipación mínima. */
export async function cancelar(
	slug: string,
	codigo: string,
	sandbox: boolean
): Promise<TurnoPublico | 'no-existe' | 'tarde'> {
	if (!CODIGO.test(codigo)) return 'no-existe';
	if (!sandbox) {
		const r = await api(`/v1/publico/tiendas/${enc(slug)}/turnos/${enc(codigo)}`, {
			method: 'DELETE'
		});
		if (r.status === 409) return 'tarde';
		return r.ok ? ((await r.json()) as TurnoPublico) : 'no-existe';
	}
	const d = await sandboxDelTurno(slug, codigo);
	return d ? d.cancelarTurnoPublico(codigo) : 'no-existe';
}

/** El archivo .ics del turno (Agregar al calendario). */
export function ics(t: TurnoPublico, url: string) {
	const utc = (iso: string) =>
		new Date(iso)
			.toISOString()
			.replace(/[-:]/g, '')
			.replace(/\.\d{3}/, '');
	// RFC 5545: comas, punto y coma y barras se escapan; renglones de hasta 75 octetos.
	const esc = (s: string) => s.replace(/[\\;,]/g, (c) => `\\${c}`).replace(/\n/g, '\\n');
	const lineas = [
		'BEGIN:VCALENDAR',
		'VERSION:2.0',
		'PRODID:-//Mostrador//Turnos//ES',
		'CALSCALE:GREGORIAN',
		'METHOD:PUBLISH',
		'BEGIN:VEVENT',
		`UID:${t.codigo}@mostrador`,
		`DTSTAMP:${utc(new Date().toISOString())}`,
		`DTSTART:${utc(t.inicio)}`,
		`DTEND:${utc(t.fin)}`,
		`SUMMARY:${esc(`${t.servicio} con ${t.profesional} · ${t.tienda}`)}`,
		...(t.direccion ? [`LOCATION:${esc(t.direccion)}`] : []),
		`DESCRIPTION:${esc(`Tu turno en ${t.tienda}. Para verlo o cancelarlo: ${url}`)}`,
		`URL:${url}`,
		'STATUS:CONFIRMED',
		'BEGIN:VALARM',
		'TRIGGER:-PT2H',
		'ACTION:DISPLAY',
		`DESCRIPTION:${esc(`Turno en ${t.tienda}`)}`,
		'END:VALARM',
		'END:VEVENT',
		'END:VCALENDAR'
	];
	const bytes = new TextEncoder();
	const doblar = (l: string) => {
		const partes: string[] = [];
		let actual = '';
		for (const ch of l) {
			// La primera parte hasta 75 octetos; las que siguen llevan un espacio adelante.
			if (bytes.encode(actual + ch).length > (partes.length ? 74 : 75)) {
				partes.push(actual);
				actual = '';
			}
			actual += ch;
		}
		return [...partes, actual].join('\r\n ');
	};
	return lineas.map(doblar).join('\r\n') + '\r\n';
}
