/**
 * La reserva de un turno en la tienda (pasos Servicio → Profesional → Horario → Pago), igual para
 * todas las plantillas: cada una sólo la dibuja con sus piezas. Los horarios los pide al servidor
 * (/t/[tienda]/turnos/libres), que los calcula con lib/turnos/libres.ts o con Go.
 */
import { diaAr, diaSemana, hace, senaDe, sumarDias, type Horario } from '#lib/turnos/libres.ts';
import { DIAS_CORTOS, cuandoLargo, duracion, mesTitulo } from '#lib/turnos/formato.ts';
import type { TurnosTienda } from './tipos.ts';

/** Cuántos días se ven a la vez en la tira de días (dos semanas, como el diseño de Tijera). */
const VENTANA = 14;

export type Inicial = {
	servicio?: string | null;
	profesional?: string | null;
	dia?: string | null;
	inicio?: string | null;
};

export class Reserva {
	servicioId = $state<string | null>(null);
	/** null = cualquiera. */
	profesionalId = $state<string | null>(null);
	dia = $state('');
	inicio = $state<string | null>(null);
	horarios = $state<Horario[] | null>(null);
	cargando = $state(false);
	/** Qué ventana de días se ve (de a 14). */
	ventana = $state(0);
	#pedido = 0;
	readonly hoy = diaAr(Date.now());

	constructor(
		private slug: string,
		private t: TurnosTienda,
		inicial: Inicial = {}
	) {
		const s = t.servicios.find((x) => x.id === inicial.servicio);
		this.servicioId = s?.id ?? null;
		const p = t.profesionales.find((x) => x.id === inicial.profesional);
		this.profesionalId = p && (!s || hace(p, s.id)) ? p.id : null;
		const dias = this.todos();
		this.dia =
			inicial.dia && dias.some((d) => d.dia === inicial.dia && d.abierto)
				? inicial.dia
				: (dias.find((d) => d.abierto)?.dia ?? this.hoy);
		this.ventana = Math.floor(
			Math.max(
				0,
				dias.findIndex((d) => d.dia === this.dia)
			) / VENTANA
		);
		this.inicio = inicial.inicio ?? null;
		if (this.servicioId) void this.cargar();
	}

	get servicio() {
		return this.t.servicios.find((x) => x.id === this.servicioId) ?? null;
	}
	get profesional() {
		return this.t.profesionales.find((x) => x.id === this.profesionalId) ?? null;
	}
	/** Profesionales que hacen el servicio elegido (todos si todavía no hay servicio). */
	get profesionales() {
		const s = this.servicioId;
		return this.t.profesionales.filter((p) => !s || hace(p, s));
	}
	/** Quién atiende el horario elegido: el pedido, o "el primero libre" con cualquiera. */
	get conQuien() {
		return this.profesional?.nombre ?? null;
	}

	/** Todos los días de la vista (hoy y los diasVista − 1 siguientes). */
	todos() {
		return Array.from({ length: this.t.diasVista }, (_, i) => {
			const dia = sumarDias(this.hoy, i);
			const w = diaSemana(dia);
			return {
				dia,
				w: DIAS_CORTOS[w],
				n: Number(dia.slice(8)),
				abierto: this.t.horario.some((f) => f.dia === w)
			};
		});
	}
	/** Los días de la ventana actual. */
	get dias() {
		return this.todos().slice(this.ventana * VENTANA, (this.ventana + 1) * VENTANA);
	}
	get hayAntes() {
		return this.ventana > 0;
	}
	get hayDespues() {
		return (this.ventana + 1) * VENTANA < this.t.diasVista;
	}
	get mes() {
		return mesTitulo(this.dias[0]?.dia ?? this.hoy);
	}
	mover(d: number) {
		this.ventana = Math.max(0, this.ventana + d);
	}

	/** Mañana (hasta las 13), Tarde (hasta las 18) y Noche, sin los grupos vacíos. */
	get grupos() {
		const h = this.horarios ?? [];
		return (
			[
				['manana', 'Mañana', (x: string) => x < '13:00'],
				['tarde', 'Tarde', (x: string) => x >= '13:00' && x < '18:00'],
				['noche', 'Noche', (x: string) => x >= '18:00']
			] as const
		)
			.map(([k, t, f]) => ({ k, t, s: h.filter((x) => f(x.hora)) }))
			.filter((g) => g.s.length);
	}
	get libres() {
		return (this.horarios ?? []).filter((h) => h.libre).length;
	}

	/** La seña en pesos (0 = se paga en el local). */
	get sena() {
		return this.servicio ? senaDe(this.servicio.precio, this.t.sena) : 0;
	}
	get cuando() {
		return this.inicio ? cuandoLargo(this.inicio) : null;
	}
	get duracion() {
		return this.servicio ? duracion(this.servicio.duracion) : '';
	}
	/** En qué paso está: 0 servicio, 1 profesional, 2 horario, 3 pago. */
	get paso() {
		return !this.servicioId ? 0 : !this.inicio ? 2 : 3;
	}
	get listo() {
		return !!this.servicioId && !!this.inicio;
	}

	elegirServicio(id: string) {
		if (this.servicioId === id) return;
		this.servicioId = id;
		if (this.profesional && !hace(this.profesional, id)) this.profesionalId = null;
		this.inicio = null;
		void this.cargar();
	}
	elegirProfesional(id: string | null) {
		if (this.profesionalId === id) return;
		this.profesionalId = id;
		this.inicio = null;
		void this.cargar();
	}
	elegirDia(dia: string) {
		if (this.dia === dia) return;
		this.dia = dia;
		this.inicio = null;
		void this.cargar();
	}
	elegirHorario(h: Horario) {
		if (h.libre) this.inicio = h.inicio;
	}

	async cargar() {
		if (!this.servicioId) return;
		const n = ++this.#pedido;
		this.cargando = true;
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, sólo arma el pedido
		const q = new URLSearchParams({ servicio: this.servicioId, dia: this.dia });
		if (this.profesionalId) q.set('profesional', this.profesionalId);
		const r = await fetch(`/t/${this.slug}/turnos/libres?${q}`).catch(() => null);
		const h = r?.ok ? ((await r.json()) as Horario[]) : [];
		if (n !== this.#pedido) return;
		this.horarios = h;
		this.cargando = false;
		// Lo elegido que ya no está libre (otro lo tomó) se suelta.
		if (this.inicio && !h.some((x) => x.inicio === this.inicio && x.libre)) this.inicio = null;
	}
}
