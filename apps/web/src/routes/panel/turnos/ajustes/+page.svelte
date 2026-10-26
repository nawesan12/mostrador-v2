<script lang="ts">
	import { untrack } from 'svelte';
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { DIAS, duracion } from '#lib/turnos/formato.ts';
	import { hhmm, minutos, type Franja } from '#lib/turnos/libres.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import type { components } from '@mostrador/contrato/esquema';

	type Servicio = components['schemas']['Servicio'];
	type Profesional = components['schemas']['Profesional'];

	/**
	 * Ajustes de turnos: activar, seña, anticipación, días de vista, horario por día con franjas,
	 * servicios y equipo. No hay diseño en el handoff: se arma con las piezas de Configuración
	 * (pestañas, tarjetas, opciones con Interruptor, campos y Modal).
	 */
	let { data } = $props();
	const ui = getUi();

	const TABS = [
		['general', I.phGearSixFill, 'General'],
		['horario', I.phClockFill, 'Horario'],
		['servicios', I.phTagFill, 'Servicios'],
		['equipo', I.phUsersThreeFill, 'Quién atiende']
	] as const;
	let tab = $state<(typeof TABS)[number][0]>('general');

	async function pedir(url: string, method: string, cuerpo?: unknown) {
		const r = await fetch(url, {
			method,
			headers: cuerpo === undefined ? {} : { 'content-type': 'application/json' },
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			ui.toast({ k: 'err', title: 'No se pudo', desc: m ?? 'Probá de nuevo.' });
			return null;
		}
		return r;
	}

	// ---- General y horario (un solo guardado) ----
	// Lo que se edita. Vuelve a lo guardado cuando se recargan los datos, salvo que haya cambios sin
	// guardar (otra sección que recarga no los pisa).
	let a = $state(structuredClone(untrack(() => $state.snapshot(data.ajustes))));
	let base = JSON.stringify(untrack(() => data.ajustes));
	$effect(() => {
		const nuevo = JSON.stringify(data.ajustes);
		if (nuevo === base) return;
		const sinCambios = untrack(() => JSON.stringify(a)) === base;
		base = nuevo;
		if (sinCambios) a = structuredClone($state.snapshot(data.ajustes));
	});
	const cambiado = $derived(JSON.stringify(a) !== JSON.stringify(data.ajustes));
	async function guardar(antes = data.ajustes) {
		if (!(await pedir('/panel/turnos/ajustes', 'PUT', a))) return;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Ajustes guardados',
			desc: a.activo
				? 'La tienda ya muestra la reserva con esto'
				: 'La reserva online está apagada',
			action: 'Deshacer',
			onAction: async () => {
				await pedir('/panel/turnos/ajustes', 'PUT', antes);
				await invalidateAll();
			}
		});
	}
	const ANTICIPACION = [
		[0, 'Sin mínimo'],
		[30, '30 minutos'],
		[60, '1 hora'],
		[120, '2 horas'],
		[240, '4 horas'],
		[720, '12 horas'],
		[1440, '1 día']
	] as const;
	const VISTA = [7, 14, 21, 30, 60, 90];
	const HORAS = [...Array.from({ length: 96 }, (_, i) => hhmm(i * 15)), '24:00'];
	const ORDEN_DIAS = [1, 2, 3, 4, 5, 6, 0];
	const franjasDe = (d: number) => a.horario.filter((f) => f.dia === d);
	function abrirDia(d: number, abierto: boolean) {
		a.horario = abierto
			? [...a.horario, { dia: d, desde: '10:00', hasta: '19:00' }]
			: a.horario.filter((f) => f.dia !== d);
	}
	function sumarFranja(d: number) {
		const ult = franjasDe(d).at(-1);
		const desde = ult ? Math.min(minutos(ult.hasta) + 60, 23 * 60) : 10 * 60;
		a.horario = [
			...a.horario,
			{ dia: d, desde: hhmm(desde), hasta: hhmm(Math.min(desde + 240, 1440)) }
		];
	}
	const quitar = (f: Franja) => (a.horario = a.horario.filter((x) => x !== f));
	const malas = $derived(a.horario.filter((f) => minutos(f.desde) >= minutos(f.hasta)));

	// ---- Servicios ----
	type FormS = Omit<Servicio, 'id' | 'orden'> & { id: string | null };
	let srv = $state<FormS | null>(null);
	const DURACIONES = Array.from({ length: 32 }, (_, i) => (i + 1) * 15);
	const nuevoServicio = () =>
		(srv = {
			id: null,
			nombre: '',
			descripcion: '',
			precio: 0,
			duracion: 30,
			foto: null,
			etiqueta: null,
			activo: true
		});
	async function guardarServicio(e: SubmitEvent) {
		e.preventDefault();
		if (!srv) return;
		const { id, ...cuerpo } = srv;
		const r = await pedir(
			id ? `/panel/turnos/servicios/${id}` : '/panel/turnos/servicios',
			id ? 'PUT' : 'POST',
			cuerpo
		);
		if (!r) return;
		srv = null;
		await invalidateAll();
		ui.toast({ k: 'ok', title: id ? 'Servicio guardado' : 'Servicio sumado', desc: cuerpo.nombre });
	}
	async function borrarServicio(s: Servicio) {
		if (!(await pedir(`/panel/turnos/servicios/${s.id}`, 'DELETE'))) return;
		await invalidateAll();
		// eslint-disable-next-line @typescript-eslint/no-unused-vars -- se descarta el id
		const { id, ...cuerpo } = s;
		ui.toast({
			k: 'ok',
			title: 'Servicio borrado',
			desc: `${s.nombre} · los turnos ya tomados no cambian`,
			action: 'Deshacer',
			onAction: async () => {
				await pedir('/panel/turnos/servicios', 'POST', cuerpo);
				await invalidateAll();
			}
		});
	}
	/** Sube o baja uno: intercambia el orden con el de al lado. */
	async function mover<T extends { id: string; orden: number }>(
		lista: T[],
		i: number,
		d: -1 | 1,
		ruta: string
	) {
		const x = lista[i];
		const y = lista[i + d];
		if (!y) return;
		const sin = (u: T) => Object.fromEntries(Object.entries(u).filter(([k]) => k !== 'id'));
		await Promise.all([
			// Se intercambian; si tenían el mismo orden, el que sube queda en ese y el otro uno más
			// (nunca un orden negativo).
			pedir(`${ruta}/${x.id}`, 'PUT', {
				...sin(x),
				orden: y.orden !== x.orden ? y.orden : d < 0 ? y.orden : y.orden + 1
			}),
			pedir(`${ruta}/${y.id}`, 'PUT', {
				...sin(y),
				orden: y.orden !== x.orden ? x.orden : d < 0 ? y.orden + 1 : y.orden
			})
		]);
		await invalidateAll();
	}

	// ---- Equipo ----
	type FormP = Omit<Profesional, 'id' | 'orden'> & { id: string | null };
	let pro = $state<FormP | null>(null);
	const nuevoProfesional = () =>
		(pro = { id: null, nombre: '', rol: '', foto: null, servicios: [], activo: true });
	async function guardarProfesional(e: SubmitEvent) {
		e.preventDefault();
		if (!pro) return;
		const { id, ...cuerpo } = pro;
		const r = await pedir(
			id ? `/panel/turnos/profesionales/${id}` : '/panel/turnos/profesionales',
			id ? 'PUT' : 'POST',
			cuerpo
		);
		if (!r) return;
		pro = null;
		await invalidateAll();
		ui.toast({ k: 'ok', title: id ? 'Guardado' : 'Sumado al equipo', desc: cuerpo.nombre });
	}
	async function borrarProfesional(p: Profesional) {
		if (!(await pedir(`/panel/turnos/profesionales/${p.id}`, 'DELETE'))) return;
		await invalidateAll();
		// eslint-disable-next-line @typescript-eslint/no-unused-vars -- se descarta el id
		const { id, ...cuerpo } = p;
		ui.toast({
			k: 'ok',
			title: 'Quitado del equipo',
			desc: `${p.nombre} · sus turnos ya tomados no cambian`,
			action: 'Deshacer',
			onAction: async () => {
				await pedir('/panel/turnos/profesionales', 'POST', cuerpo);
				await invalidateAll();
			}
		});
	}
	const hace = (p: Profesional) =>
		!p.servicios.length
			? 'Todos los servicios'
			: data.servicios
					.filter((s) => p.servicios.includes(s.id))
					.map((s) => s.nombre)
					.join(', ') || 'Ningún servicio activo';
	const foto = (f: string | null, w: number) =>
		!f ? null : f.startsWith('https://') ? f : unsplash(f, w, `&h=${w}&crop=faces`);
</script>

<Seo
	title="Ajustes de turnos · Panel"
	description="Ajustes de turnos"
	path="/panel/turnos/ajustes"
	noindex
/>

<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas del panel -->
<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<a class="volver" href="/panel/turnos"><Icon icon={I.phCaretLeftBold} size={13} />Agenda</a>
		<h1>Ajustes de turnos</h1>
		<div class="c-s">Qué se reserva, cuándo y con quién</div>
	</div>

	<div class="grilla" use:anim={{ t: 'rise', delay: 60 }}>
		<nav class="tabs" aria-label="Secciones" use:deslizable>
			{#each TABS as [id, icono, t] (id)}
				<button
					type="button"
					aria-current={tab === id ? 'page' : undefined}
					onclick={() => (tab = id)}><Icon icon={icono} size={17} />{t}</button
				>
			{/each}
		</nav>

		<div class="contenido">
			{#if tab === 'general'}
				<div class="tarjeta lista">
					<div class="opcion">
						<span class="o-ico"><Icon icon={I.phCalendarCheckFill} size={18} /></span>
						<span class="o-txt"
							><span class="o-t">Tomar turnos online</span><span class="o-d"
								>La tienda muestra "Reservar turno" con tus servicios y horarios</span
							></span
						>
						<Interruptor bind:checked={a.activo} label="Tomar turnos online" />
					</div>
					<label class="opcion">
						<span class="o-txt"
							><span class="o-t">Seña</span><span class="o-d"
								>{data.mercadoPago
									? 'Porcentaje del precio que se paga al reservar · 0 = sin seña'
									: 'Conectá Mercado Pago para cobrarla; sin eso, el turno se confirma sin seña'}</span
							></span
						>
						<span class="sufijo"
							><input
								class="input chico"
								type="number"
								min="0"
								max="100"
								step="5"
								inputmode="numeric"
								bind:value={a.sena}
								aria-label="Seña en porcentaje"
							/>%</span
						>
					</label>
					<label class="opcion">
						<span class="o-txt"
							><span class="o-t">Anticipación mínima</span><span class="o-d"
								>Cuánto antes hay que reservar (y hasta cuándo se puede cancelar online)</span
							></span
						>
						<select
							class="select medio"
							bind:value={a.anticipacionMin}
							aria-label="Anticipación mínima"
							>{#each ANTICIPACION as [v, t] (v)}<option value={v}>{t}</option>{/each}</select
						>
					</label>
					<label class="opcion">
						<span class="o-txt"
							><span class="o-t">Días para adelante</span><span class="o-d"
								>Cuántos días se pueden reservar, contando hoy</span
							></span
						>
						<select class="select medio" bind:value={a.diasVista} aria-label="Días para adelante"
							>{#each VISTA as v (v)}<option value={v}>{v} días</option>{/each}</select
						>
					</label>
				</div>
			{:else if tab === 'horario'}
				<div class="tarjeta lista">
					{#each ORDEN_DIAS as d (d)}
						{@const fr = franjasDe(d)}
						<div class="dia">
							<div class="d-cab">
								<span class="o-t">{DIAS[d]}</span>
								<span class="d-est">{fr.length ? '' : 'Cerrado'}</span>
								<Interruptor
									checked={!!fr.length}
									label="Abierto el {DIAS[d].toLowerCase()}"
									onchange={(v) => abrirDia(d, v)}
								/>
							</div>
							{#each fr as f, i (i)}
								<div class="franja">
									<select class="select" bind:value={f.desde} aria-label="Desde"
										>{#each HORAS.slice(0, -1) as h (h)}<option value={h}>{h}</option
											>{/each}</select
									>
									<span>a</span>
									<select class="select" bind:value={f.hasta} aria-label="Hasta"
										>{#each HORAS.slice(1) as h (h)}<option value={h}>{h}</option>{/each}</select
									>
									<button
										type="button"
										class="icon-btn"
										aria-label="Quitar franja"
										onclick={() => quitar(f)}><Icon icon={I.phTrash} size={15} /></button
									>
								</div>
							{/each}
							{#if fr.length}
								<button type="button" class="sumar" onclick={() => sumarFranja(d)}
									><Icon icon={I.phPlusBold} size={12} />Sumar franja (horario cortado)</button
								>
							{/if}
						</div>
					{/each}
				</div>
			{:else if tab === 'servicios'}
				<div class="tarjeta">
					<div class="t-cab">
						<span>Servicios · {data.servicios.length}</span>
						<button type="button" class="oscuro" onclick={nuevoServicio}>Sumar servicio</button>
					</div>
					<div class="filas">
						{#each data.servicios as s, i (s.id)}
							<div class="fila" class:pausado={!s.activo}>
								{#if foto(s.foto, 88)}<img src={foto(s.foto, 88)} alt="" />{:else}<span class="av"
										><Icon icon={I.phTagFill} size={16} /></span
									>{/if}
								<span class="m-txt"
									><span class="m-n"
										>{s.nombre}{#if s.etiqueta}<span class="tag">{s.etiqueta}</span
											>{/if}{#if !s.activo}<span class="tag gris">Pausado</span>{/if}</span
									><span class="m-e">{duracion(s.duracion)} · {pesos(s.precio)}</span></span
								>
								<span class="acc">
									<button
										type="button"
										class="icon-btn"
										aria-label="Subir {s.nombre}"
										disabled={i === 0}
										onclick={() => mover(data.servicios, i, -1, '/panel/turnos/servicios')}
										><Icon icon={I.phArrowUpBold} size={14} /></button
									>
									<button
										type="button"
										class="icon-btn"
										aria-label="Bajar {s.nombre}"
										disabled={i === data.servicios.length - 1}
										onclick={() => mover(data.servicios, i, 1, '/panel/turnos/servicios')}
										><Icon icon={I.phArrowDownBold} size={14} /></button
									>
									<button
										type="button"
										class="icon-btn"
										aria-label="Editar {s.nombre}"
										onclick={() => (srv = { ...s })}
										><Icon icon={I.phPencilSimple} size={15} /></button
									>
									<button
										type="button"
										class="icon-btn"
										aria-label="Borrar {s.nombre}"
										onclick={() => borrarServicio(s)}><Icon icon={I.phTrash} size={15} /></button
									>
								</span>
							</div>
						{:else}
							<div class="vacio">Sumá tu primer servicio: nombre, duración y precio.</div>
						{/each}
					</div>
				</div>
			{:else}
				<div class="tarjeta">
					<div class="t-cab">
						<span>Quién atiende · {data.profesionales.length}</span>
						<button type="button" class="oscuro" onclick={nuevoProfesional}>Sumar</button>
					</div>
					<div class="filas">
						{#each data.profesionales as p, i (p.id)}
							<div class="fila" class:pausado={!p.activo}>
								{#if foto(p.foto, 76)}<img
										class="redonda"
										src={foto(p.foto, 76)}
										alt=""
									/>{:else}<span class="av redonda">{p.nombre.slice(0, 1)}</span>{/if}
								<span class="m-txt"
									><span class="m-n"
										>{p.nombre}{#if !p.activo}<span class="tag gris">Pausado</span>{/if}</span
									><span class="m-e">{p.rol ? `${p.rol} · ` : ''}{hace(p)}</span></span
								>
								<span class="acc">
									<button
										type="button"
										class="icon-btn"
										aria-label="Subir {p.nombre}"
										disabled={i === 0}
										onclick={() => mover(data.profesionales, i, -1, '/panel/turnos/profesionales')}
										><Icon icon={I.phArrowUpBold} size={14} /></button
									>
									<button
										type="button"
										class="icon-btn"
										aria-label="Bajar {p.nombre}"
										disabled={i === data.profesionales.length - 1}
										onclick={() => mover(data.profesionales, i, 1, '/panel/turnos/profesionales')}
										><Icon icon={I.phArrowDownBold} size={14} /></button
									>
									<button
										type="button"
										class="icon-btn"
										aria-label="Editar {p.nombre}"
										onclick={() => (pro = { ...p, servicios: [...p.servicios] })}
										><Icon icon={I.phPencilSimple} size={15} /></button
									>
									<button
										type="button"
										class="icon-btn"
										aria-label="Quitar a {p.nombre}"
										onclick={() => borrarProfesional(p)}><Icon icon={I.phTrash} size={15} /></button
									>
								</span>
							</div>
						{:else}
							<div class="vacio">
								Sumá a quien atiende: con "cualquiera", se asigna el primero libre en este orden.
							</div>
						{/each}
					</div>
				</div>
			{/if}

			{#if (tab === 'general' || tab === 'horario') && cambiado}
				<div class="guardar">
					{#if malas.length}<span class="error"
							>Cada franja tiene que terminar después de empezar.</span
						>{/if}
					<button
						type="button"
						class="btn btn-secondary"
						onclick={() => (a = structuredClone($state.snapshot(data.ajustes)))}>Descartar</button
					>
					<button
						type="button"
						class="btn btn-primary"
						disabled={!!malas.length}
						onclick={() => guardar()}>Guardar cambios</button
					>
				</div>
			{/if}
		</div>
	</div>
</div>

<Modal
	bind:open={
		() => !!srv,
		(v) => {
			if (!v) srv = null;
		}
	}
	titulo={srv?.id ? 'Editar servicio' : 'Sumar servicio'}
	ancho={480}
>
	{#if srv}
		<form onsubmit={guardarServicio}>
			<div class="m-t">{srv.id ? 'Editar servicio' : 'Sumar servicio'}</div>
			<div class="campos">
				<label class="field ancho"
					>Nombre<input class="input" bind:value={srv.nombre} maxlength="60" required /></label
				>
				<label class="field ancho"
					>Descripción<input
						class="input"
						bind:value={srv.descripcion}
						maxlength="200"
						placeholder="Opcional"
					/></label
				>
				<label class="field"
					>Precio<input
						class="input"
						type="number"
						min="1"
						inputmode="numeric"
						bind:value={srv.precio}
						required
					/></label
				>
				<label class="field"
					>Duración<select class="select" bind:value={srv.duracion}
						>{#each DURACIONES as d (d)}<option value={d}>{duracion(d)}</option>{/each}</select
					></label
				>
				<label class="field"
					>Etiqueta<input
						class="input"
						bind:value={srv.etiqueta}
						maxlength="20"
						placeholder="Promo, Nuevo…"
					/></label
				>
				<label class="field"
					>Foto (link)<input
						class="input"
						type="url"
						bind:value={srv.foto}
						maxlength="2048"
						placeholder="https://…"
					/></label
				>
			</div>
			<div class="opcion sin-borde">
				<span class="o-txt"
					><span class="o-t">Se puede reservar</span><span class="o-d"
						>Pausado, no aparece en la tienda</span
					></span
				>
				<Interruptor bind:checked={srv.activo} label="Se puede reservar" />
			</div>
			<div class="m-acc">
				<button type="button" class="btn btn-secondary" onclick={() => (srv = null)}
					>Cancelar</button
				>
				<button class="btn btn-primary" disabled={srv.nombre.trim().length < 2 || !(srv.precio > 0)}
					>Guardar</button
				>
			</div>
		</form>
	{/if}
</Modal>

<Modal
	bind:open={
		() => !!pro,
		(v) => {
			if (!v) pro = null;
		}
	}
	titulo={pro?.id ? 'Editar' : 'Sumar al equipo'}
	ancho={480}
>
	{#if pro}
		<form onsubmit={guardarProfesional}>
			<div class="m-t">{pro.id ? `Editar a ${pro.nombre}` : 'Sumar al equipo'}</div>
			<div class="campos">
				<label class="field"
					>Nombre<input class="input" bind:value={pro.nombre} maxlength="40" required /></label
				>
				<label class="field"
					>Qué hace<input
						class="input"
						bind:value={pro.rol}
						maxlength="60"
						placeholder="Barbero · fades y barba"
					/></label
				>
				<label class="field ancho"
					>Foto (link)<input
						class="input"
						type="url"
						bind:value={pro.foto}
						maxlength="2048"
						placeholder="https://…"
					/></label
				>
			</div>
			<div class="m-sub">Servicios que hace <span>(ninguno marcado = todos)</span></div>
			<div class="chips" role="group" aria-label="Servicios que hace">
				{#each data.servicios as s (s.id)}
					{@const on = pro.servicios.includes(s.id)}
					<button
						type="button"
						aria-pressed={on}
						onclick={() =>
							(pro!.servicios = on
								? pro!.servicios.filter((x) => x !== s.id)
								: [...pro!.servicios, s.id])}
						>{#if on}<Icon icon={I.phCheckBold} size={12} />{/if}{s.nombre}</button
					>
				{/each}
			</div>
			<div class="opcion sin-borde">
				<span class="o-txt"
					><span class="o-t">Atiende</span><span class="o-d"
						>Pausado, no se le dan turnos nuevos</span
					></span
				>
				<Interruptor bind:checked={pro.activo} label="Atiende" />
			</div>
			<div class="m-acc">
				<button type="button" class="btn btn-secondary" onclick={() => (pro = null)}
					>Cancelar</button
				>
				<button class="btn btn-primary" disabled={pro.nombre.trim().length < 2}>Guardar</button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.volver {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		font-size: 13.5px;
		font-weight: 600;
		color: #c24e25;
		text-decoration: none;
	}
	h1 {
		margin: 4px 0 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.grilla {
		display: grid;
		grid-template-columns: 240px minmax(0, 1fr);
		gap: 16px;
		align-items: start;
	}
	.tabs {
		padding: 8px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 2px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.tabs button {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 14px;
		border: 0;
		background: transparent;
		color: var(--tx);
		font:
			500 14px 'Instrument Sans',
			sans-serif;
		text-align: left;
		white-space: nowrap;
	}
	.tabs button :global(svg) {
		color: var(--mu);
	}
	.tabs button[aria-current='page'] {
		background: var(--chip);
		font-weight: 600;
	}
	.tabs button[aria-current='page'] :global(svg) {
		color: #d2582f;
	}
	.contenido {
		display: flex;
		flex-direction: column;
		gap: 12px;
		min-width: 0;
	}
	.tarjeta {
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.lista {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.opcion {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 0;
		border-bottom: 1px solid var(--ln);
	}
	.opcion.sin-borde {
		margin-top: 8px;
		border-bottom: 0;
	}
	.o-ico {
		width: 36px;
		height: 36px;
		flex: none;
		border-radius: 11px;
		background: rgba(47, 143, 99, 0.12);
		color: #2f8f63;
		display: grid;
		place-items: center;
	}
	.o-txt {
		flex: 1;
		min-width: 0;
	}
	.o-t {
		display: block;
		font-size: 14.5px;
		font-weight: 600;
	}
	.o-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.sufijo {
		display: flex;
		align-items: center;
		gap: 6px;
		font-weight: 600;
	}
	.input.chico {
		width: 78px;
	}
	.select.medio {
		width: auto;
		max-width: 150px;
	}
	.dia {
		padding: 12px 0;
		border-bottom: 1px solid var(--ln);
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.d-cab {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.d-cab .o-t {
		flex: 1;
	}
	.d-est {
		font-size: 13px;
		color: var(--mu);
	}
	.franja {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr) auto;
		align-items: center;
		gap: 8px;
		font-size: 13.5px;
		color: var(--mu);
		max-width: 420px;
	}
	.sumar {
		align-self: flex-start;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 6px 0;
		border: 0;
		background: transparent;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		color: #c24e25;
	}
	.t-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		font-size: 16px;
		font-weight: 600;
	}
	.oscuro {
		padding: 9px 14px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.filas {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid var(--ln);
		flex-wrap: wrap;
	}
	.fila.pausado {
		opacity: 0.65;
	}
	.fila img,
	.av {
		width: 44px;
		height: 44px;
		flex: none;
		border-radius: 12px;
		object-fit: cover;
	}
	.av {
		display: grid;
		place-items: center;
		background: var(--chip);
		color: var(--mu);
		font-weight: 600;
	}
	.redonda {
		border-radius: 50% !important;
	}
	.m-txt {
		flex: 1 1 160px;
		min-width: 0;
	}
	.m-n {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-wrap: wrap;
		font-size: 14px;
		font-weight: 600;
	}
	.m-e {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		overflow-wrap: anywhere;
	}
	.tag {
		font-size: 11px;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 999px;
		background: rgba(210, 88, 47, 0.12);
		color: #c24e25;
	}
	.tag.gris {
		background: var(--chip);
		color: var(--mu);
	}
	.acc {
		display: flex;
		gap: 4px;
		margin-left: auto;
	}
	.acc .icon-btn:disabled {
		opacity: 0.35;
	}
	.vacio {
		padding: 20px 0 8px;
		text-align: center;
		color: var(--mu);
		font-size: 14px;
	}
	.guardar {
		display: flex;
		justify-content: flex-end;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.error {
		flex: 1 1 200px;
		font-size: 13px;
		color: var(--error);
	}
	.m-t {
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.campos {
		margin-top: 16px;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
	}
	.campos .ancho {
		grid-column: 1 / -1;
	}
	.m-sub {
		margin-top: 14px;
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	.m-sub span {
		font-weight: 500;
	}
	.chips {
		margin-top: 8px;
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.chips button {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		padding: 8px 12px;
		border-radius: 999px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			500 13px 'Instrument Sans',
			sans-serif;
	}
	.chips button[aria-pressed='true'] {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.06);
	}
	.m-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
	@media (max-width: 899px) {
		.grilla {
			grid-template-columns: minmax(0, 1fr);
		}
		.tabs {
			flex-direction: row;
		}
	}
	@media (max-width: 759px) {
		.campos {
			grid-template-columns: minmax(0, 1fr);
		}
		.acc {
			width: 100%;
			justify-content: flex-end;
		}
	}
</style>
