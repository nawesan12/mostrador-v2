import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { diaAr, diaSemana, sumarDias } from '#lib/turnos/libres.ts';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox } from './sandbox.ts';
import { turnosDemo } from './turnos.ts';

const dir = join(import.meta.dirname, '../../../../migraciones-sandbox');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));

let d1: ReturnType<typeof d1Local> | undefined;
afterEach(() => {
	d1?.cerrar();
	d1 = undefined;
});

async function demo(plantilla = 'tijera', rubro = 'servicios') {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, rubro, Date.now(), plantilla))!;
	return new DatosSandbox(db, sb);
}

/** Un día hábil (lunes a viernes) a partir de pasado mañana, para no depender de la hora. */
function diaHabil() {
	let d = sumarDias(diaAr(Date.now()), 2);
	while (diaSemana(d) === 0 || diaSemana(d) === 6) d = sumarDias(d, 1);
	return d;
}

describe('turnos de la demo', () => {
	it('Tijera arranca con el horario, los servicios y el equipo del diseño', () => {
		const t = turnosDemo('tijera', Date.UTC(2026, 9, 20, 15));
		expect(t.ajustes.activo).toBe(true);
		expect(t.servicios.map((s) => s.nombre)).toEqual([
			'Corte',
			'Corte + barba',
			'Barba',
			'Color',
			'Corte infantil',
			'Afeitado clásico'
		]);
		expect(t.profesionales.map((p) => p.nombre)).toEqual(['Bruno', 'Lara', 'Nico', 'Ema']);
		expect(t.turnos.length).toBeGreaterThan(8);
		// Ningún turno de ejemplo cae en domingo (cerrado).
		expect(t.turnos.every((x) => diaSemana(x.inicio.slice(0, 10)) !== 0)).toBe(true);
	});

	it('los otros rubros arrancan con turnos apagados y la tienda no los muestra', async () => {
		const d = await demo('ceibo', 'moda');
		expect((await d.ajustesTurnos()).activo).toBe(false);
		expect((await d.tiendaPublica()).turnos).toBeNull();
	});

	it('reservar ocupa el horario, avisa en el panel y cancelar por código lo libera', async () => {
		const d = await demo();
		expect((await d.tiendaPublica()).turnos?.servicios.length).toBe(6);
		const dia = diaHabil();
		const antes = (await d.horariosTurnos('corte', 'bruno', dia))!;
		const h = antes.find((x) => x.libre)!;
		const t = await d.reservarTurno({
			servicioId: 'corte',
			profesionalId: 'bruno',
			inicio: h.inicio,
			nombre: 'Lucía Pérez',
			telefono: '11 5555-0101',
			email: 'lucia@ejemplo.com'
		});
		if (typeof t === 'string') throw new Error(t);
		// En la demo no hay Mercado Pago real: confirmado y sin seña; precio del servicio.
		expect(t).toMatchObject({ estado: 'confirmado', sena: 0, precio: 12500, origen: 'tienda' });
		expect(
			(await d.horariosTurnos('corte', 'bruno', dia))!.find((x) => x.inicio === h.inicio)?.libre
		).toBe(false);
		expect(
			await d.reservarTurno({
				servicioId: 'corte',
				profesionalId: 'bruno',
				inicio: h.inicio,
				nombre: 'Otra',
				telefono: '11 5555-0102',
				email: 'otra@ejemplo.com'
			})
		).toBe('ocupado');
		expect((await d.resumen()).notificaciones[0].titulo).toBe('Turno nuevo: Corte con Bruno');
		expect((await d.agendaTurnos(dia, dia)).some((x) => x.id === t.id)).toBe(true);

		const pub = (await d.turnoPublico(t.codigo))!;
		expect(pub).toMatchObject({ nombre: 'Lucía', cancelable: true, estado: 'confirmado' });
		expect(await d.cancelarTurnoPublico(t.codigo)).toMatchObject({ estado: 'cancelado' });
		expect(
			(await d.horariosTurnos('corte', 'bruno', dia))!.find((x) => x.inicio === h.inicio)?.libre
		).toBe(true);
		expect(await d.turnoPublico('no-existe-este-codigo')).toBeNull();
	});

	it('con cualquiera se asigna el primero libre por orden', async () => {
		const d = await demo();
		const dia = diaHabil();
		const h = (await d.horariosTurnos('color', null, dia))!.find((x) => x.libre)!;
		const t = await d.reservarTurno({
			servicioId: 'color',
			profesionalId: null,
			inicio: h.inicio,
			nombre: 'Ana',
			telefono: '11 5555-0103',
			email: 'ana@ejemplo.com'
		});
		if (typeof t === 'string') throw new Error(t);
		// Lara (orden 1) hace color; Bruno (orden 0) no.
		expect(['lara', 'ema']).toContain(t.profesionalId);
	});

	it('el turno a mano no se superpone salvo que se pida; estados y seña', async () => {
		const d = await demo();
		const dia = diaHabil();
		const base = {
			servicioId: 'barba',
			profesionalId: 'nico',
			inicio: `${dia}T19:00:00-03:00`,
			nombre: 'Juan',
			telefono: '',
			email: '',
			forzar: false
		};
		const t = await d.crearTurno(base);
		if (typeof t === 'string') throw new Error(t);
		expect(t).toMatchObject({ origen: 'panel', estado: 'confirmado', sena: 0 });
		expect(await d.crearTurno({ ...base, inicio: `${dia}T19:15:00-03:00` })).toBe('ocupado');
		expect(
			typeof (await d.crearTurno({ ...base, inicio: `${dia}T19:15:00-03:00`, forzar: true }))
		).toBe('object');
		expect(await d.estadoTurno(t.id, 'atendido')).toMatchObject({ estado: 'atendido' });
		expect(await d.estadoTurno('no-existe', 'atendido')).toBe('no-existe');
		expect(await d.devolverSena(t.id)).toBe('sin-sena');
		const conSena = (await d.agendaTurnos('2000-01-01', '2100-01-01')).find((x) => x.senaPagada)!;
		expect(await d.devolverSena(conSena.id)).toMatchObject({ senaDevuelta: true });
		expect(await d.devolverSena(conSena.id)).toBe('sin-sena');
	});

	it('servicios y profesionales: alta, edición, orden y baja', async () => {
		const d = await demo('ceibo', 'moda');
		const s = await d.guardarServicio(null, {
			nombre: 'Arreglo de ruedo',
			descripcion: '',
			precio: 5000,
			duracion: 30,
			foto: null,
			etiqueta: null,
			activo: true
		});
		expect(s?.orden).toBe(0);
		const p = await d.guardarProfesional(null, {
			nombre: 'Marta',
			rol: 'Costurera',
			foto: null,
			servicios: [s!.id],
			activo: true
		});
		expect(await d.guardarServicio(s!.id, { ...s!, precio: 6000 })).toMatchObject({ precio: 6000 });
		expect(await d.guardarServicio('no-existe', { ...s! })).toBeNull();
		await d.guardarAjustesTurnos({ ...(await d.ajustesTurnos()), activo: true });
		expect((await d.tiendaPublica()).turnos?.profesionales.map((x) => x.nombre)).toEqual(['Marta']);
		expect(await d.borrarProfesional(p!.id)).toBe('ok');
		expect(await d.borrarServicio(s!.id)).toBe('ok');
		expect(await d.serviciosTurnos()).toEqual([]);
	});
});
