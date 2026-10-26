<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import Alerta from '#lib/components/Alerta.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import HojaTurno from '#lib/panel/turnos/HojaTurno.svelte';
	import { ESTADO_TURNO, HECHO, type EstadoTurno, type Turno } from '#lib/panel/turnos.ts';
	import { DIAS, DIAS_CORTOS, MESES, diaLargo } from '#lib/turnos/formato.ts';
	import { diaAr, diaSemana, hhmm, horaAr, minutos, sumarDias } from '#lib/turnos/libres.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Turnos: la agenda del día (columnas por profesional en escritorio, lista por hora en el
	 * celular) o de la semana. No hay diseño en el handoff: se arma con las piezas de Pedidos y
	 * Configuración (cabecera, Segmentado, tarjetas, chips de estado, Hoja y Modal).
	 */
	let { data } = $props();
	const APAGADOS =
		'La tienda no muestra la reserva online hasta que los actives. Igual podés cargar turnos a\u00a0mano.';
	const ui = getUi();
	let ancho = $state(1200);
	const celular = $derived(ancho < 760);

	// eslint-disable-next-line svelte/no-navigation-without-resolve -- la agenda con su día y vista
	const ir = (dia: string, vista = data.vista) => goto(`/panel/turnos?dia=${dia}&vista=${vista}`);
	const mover = (n: number) => ir(sumarDias(data.dia, n * (data.vista === 'semana' ? 7 : 1)));

	const titulo = $derived.by(() => {
		if (data.vista === 'dia')
			return data.dia === data.hoy
				? `Hoy, ${diaLargo(data.dia).toLowerCase()}`
				: diaLargo(data.dia);
		const [a, b] = [data.desde, data.hasta];
		const mes = (d: string) => MESES[Number(d.slice(5, 7)) - 1];
		return `Semana del ${Number(a.slice(8))}${mes(a) === mes(b) ? '' : ` de ${mes(a)}`} al ${Number(b.slice(8))} de ${mes(b)}`;
	});
	const vivos = $derived(data.turnos.filter((t) => t.estado !== 'cancelado'));
	const resumen = $derived(
		`${vivos.length} ${vivos.length === 1 ? 'turno' : 'turnos'}${data.vista === 'semana' ? ' en la semana' : ''}`
	);

	// ---- Día: columnas por profesional ----
	const columnas = $derived.by(() => {
		const activos = data.profesionales.filter((p) => p.activo);
		const extra = [...new Set(data.turnos.map((t) => t.profesionalId))]
			.filter((id) => !activos.some((p) => p.id === id))
			.map((id) => {
				const t = data.turnos.find((x) => x.profesionalId === id)!;
				return { id, nombre: t.profesional, foto: null as string | null };
			});
		return [...activos.map((p) => ({ id: p.id, nombre: p.nombre, foto: p.foto })), ...extra];
	});
	/** Desde y hasta (minutos) de la grilla: el horario del día y los turnos que haya. */
	const rango = $derived.by(() => {
		const fr = data.ajustes.horario.filter((f) => f.dia === diaSemana(data.dia));
		const ts = data.turnos.map((t) => [
			minutos(horaAr(Date.parse(t.inicio))),
			minutos(horaAr(Date.parse(t.fin))) || 1440
		]);
		const desde = Math.min(...fr.map((f) => minutos(f.desde)), ...ts.map((x) => x[0]), 10 * 60);
		const hasta = Math.max(...fr.map((f) => minutos(f.hasta)), ...ts.map((x) => x[1]), 18 * 60);
		return { desde: Math.floor(desde / 60) * 60, hasta: Math.ceil(hasta / 60) * 60 };
	});
	const PX = 2;
	const horas = $derived(
		Array.from({ length: (rango.hasta - rango.desde) / 60 }, (_, i) => rango.desde + i * 60)
	);
	const pos = (t: Turno) => {
		const a = minutos(horaAr(Date.parse(t.inicio)));
		const dur = (Date.parse(t.fin) - Date.parse(t.inicio)) / 60_000;
		return `top:${(a - rango.desde) * PX}px;height:${Math.max(dur * PX - 3, 26)}px`;
	};
	const rangoHora = (t: Turno) => `${horaAr(Date.parse(t.inicio))}–${horaAr(Date.parse(t.fin))}`;
	/** Día en el celular: por hora. */
	const porHora = $derived.by(() => {
		const m: Record<string, Turno[]> = {};
		for (const t of data.turnos)
			(m[horaAr(Date.parse(t.inicio)).slice(0, 2) + ':00'] ??= []).push(t);
		return Object.entries(m);
	});
	// ---- Semana ----
	const semana = $derived(
		Array.from({ length: 7 }, (_, i) => {
			const dia = sumarDias(data.desde, i);
			return {
				dia,
				w: DIAS[diaSemana(dia)],
				c: DIAS_CORTOS[diaSemana(dia)],
				n: Number(dia.slice(8)),
				turnos: data.turnos.filter((t) => diaAr(Date.parse(t.inicio)) === dia)
			};
		})
	);

	// ---- Detalle y acciones ----
	let abierto = $state<Turno | null>(null);
	let hoja = $state(false);
	function abrir(t: Turno) {
		abierto = t;
		hoja = true;
	}
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
	async function estado(t: Turno, e: EstadoTurno) {
		const r = await pedir(`/panel/turnos/${t.id}`, 'PATCH', { estado: e });
		if (!r) return;
		abierto = (await r.json()) as Turno;
		await invalidateAll();
		const antes = t.estado;
		ui.toast({
			k: 'ok',
			title: HECHO[e],
			desc:
				e === 'cancelado' && t.senaPagada && !t.senaDevuelta
					? 'La seña no se devuelve sola: usá Devolver la seña'
					: `${t.cliente.nombre} · ${t.servicio}`,
			action: 'Deshacer',
			onAction: async () => {
				const v = await pedir(`/panel/turnos/${t.id}`, 'PATCH', { estado: antes });
				if (v && abierto?.id === t.id) abierto = (await v.json()) as Turno;
				await invalidateAll();
			}
		});
	}
	async function devolver(t: Turno) {
		const r = await pedir(`/panel/turnos/${t.id}/devolver-sena`, 'POST', {});
		if (!r) return;
		abierto = (await r.json()) as Turno;
		await invalidateAll();
		ui.toast({
			k: 'mp',
			title: 'Seña devuelta',
			desc: `Mercado Pago se la reintegra a ${t.cliente.nombre}`
		});
	}

	// ---- Turno a mano ----
	let nuevo = $state(false);
	let n = $state({
		servicioId: '',
		profesionalId: '',
		dia: '',
		hora: '10:00',
		nombre: '',
		telefono: '',
		email: '',
		nota: ''
	});
	let superpuesto = $state(false);
	let guardando = $state(false);
	function abrirNuevo() {
		const s = data.servicios.find((x) => x.activo);
		const p = data.profesionales.find((x) => x.activo);
		n = {
			servicioId: s?.id ?? '',
			profesionalId: p?.id ?? '',
			dia: data.dia,
			hora: '10:00',
			nombre: '',
			telefono: '',
			email: '',
			nota: ''
		};
		superpuesto = false;
		nuevo = true;
	}
	async function crear(e: SubmitEvent, forzar = false) {
		e.preventDefault();
		guardando = true;
		let r: Response;
		try {
			r = await fetch('/panel/turnos/nuevo', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({
					servicioId: n.servicioId,
					profesionalId: n.profesionalId,
					inicio: `${n.dia}T${n.hora}:00-03:00`,
					nombre: n.nombre,
					telefono: n.telefono,
					email: n.email,
					nota: n.nota || null,
					forzar
				})
			});
		} catch {
			return ui.toast({
				k: 'err',
				title: 'No se pudo',
				desc: 'Revisá la conexión y probá de nuevo.'
			});
		} finally {
			guardando = false;
		}
		if (r.status === 409) return void (superpuesto = true);
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({ k: 'err', title: 'No se pudo', desc: m ?? 'Probá de nuevo.' });
		}
		const t = (await r.json()) as Turno;
		nuevo = false;
		if (n.dia !== data.dia && data.vista === 'dia') await ir(n.dia);
		else await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Turno cargado',
			desc: `${t.cliente.nombre} · ${diaLargo(n.dia)} a las ${n.hora}`,
			action: 'Deshacer',
			onAction: async () => {
				await pedir(`/panel/turnos/${t.id}`, 'PATCH', { estado: 'cancelado' });
				await invalidateAll();
			}
		});
	}
	const HORAS = Array.from({ length: 96 }, (_, i) => hhmm(i * 15));
	const valido = $derived(
		!!n.servicioId && !!n.profesionalId && !!n.dia && n.nombre.trim().length >= 2
	);
</script>

<svelte:window bind:innerWidth={ancho} />

<Seo title="Turnos · Panel" description="La agenda de turnos" path="/panel/turnos" noindex />

<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas del panel con su día -->
{#snippet tarjeta(t: Turno, compacta = false)}
	{@const e = ESTADO_TURNO[t.estado]}
	<button
		type="button"
		class="turno"
		class:compacta
		class:apagado={t.estado === 'cancelado' || t.vencido}
		style:--borde={e.c[1]}
		onclick={() => abrir(t)}
	>
		<span class="t-h">{rangoHora(t)}</span>
		<span class="t-n">{t.cliente.nombre}</span>
		<span class="t-s">{t.servicio}{compacta ? '' : ` · ${t.profesional}`}</span>
		{#if !compacta || t.estado !== 'confirmado'}<span
				class="chip-estado"
				style:--chip-bg={e.c[0]}
				style:--chip-fg={e.c[1]}>{t.vencido ? 'Venció la seña' : e.t}</span
			>{/if}
	</button>
{/snippet}

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Turnos</h1>
			<div class="c-s">{resumen}</div>
		</div>
		<a class="btn btn-secondary" href="/panel/turnos/ajustes"
			><Icon icon={I.phGearSix} size={16} />Ajustes</a
		>
		<button type="button" class="btn btn-primary" onclick={abrirNuevo}
			><Icon icon={I.phPlusBold} size={15} />Nuevo turno</button
		>
	</div>

	{#if !data.ajustes.activo}
		<Alerta
			tipo="warn"
			titulo="Los turnos están apagados"
			texto={APAGADOS}
			accion="Activar"
			onaccion={() => goto('/panel/turnos/ajustes')}
		/>
	{/if}

	<div class="barra" use:anim={{ t: 'rise', delay: 40 }}>
		<Segmentado
			label="Vista"
			opciones={[
				['dia', 'Día'],
				['semana', 'Semana']
			]}
			value={data.vista as 'dia' | 'semana'}
			onchange={(v) => ir(data.dia, v)}
		/>
		<div class="nav">
			<button type="button" class="icon-btn" aria-label="Anterior" onclick={() => mover(-1)}
				><Icon icon={I.phCaretLeftBold} size={14} /></button
			>
			<button
				type="button"
				class="btn btn-secondary btn-sm"
				disabled={data.dia === data.hoy}
				onclick={() => ir(data.hoy)}>Hoy</button
			>
			<button type="button" class="icon-btn" aria-label="Siguiente" onclick={() => mover(1)}
				><Icon icon={I.phCaretRightBold} size={14} /></button
			>
		</div>
		<span class="fecha">{titulo}</span>
	</div>

	{#if data.vista === 'dia'}
		{#if !data.turnos.length}
			<div class="vacio card">
				No hay turnos este día. Los que reserven desde la tienda aparecen&nbsp;acá.
			</div>
		{:else if celular}
			<div class="lista" use:anim={{ t: 'rise', delay: 60 }}>
				{#each porHora as [h, ts] (h)}
					<div class="hora-t">{h}</div>
					{#each ts as t (t.id)}{@render tarjeta(t)}{/each}
				{/each}
			</div>
		{:else}
			<div class="grilla card" use:anim={{ t: 'rise', delay: 60 }}>
				<div class="g-cab" style:--cols={columnas.length}>
					<span></span>
					{#each columnas as c (c.id)}
						<span class="g-pro"
							>{#if c.foto}<img src={unsplash(c.foto, 64, '&h=64&crop=faces')} alt="" />{:else}<span
									class="av">{c.nombre.slice(0, 1)}</span
								>{/if}{c.nombre}</span
						>
					{/each}
				</div>
				<div
					class="g-cuerpo"
					style:--cols={columnas.length}
					style:height="{(rango.hasta - rango.desde) * PX}px"
				>
					<div class="g-horas">
						{#each horas as h (h)}<span style:top="{(h - rango.desde) * PX}px">{hhmm(h)}</span
							>{/each}
					</div>
					{#each columnas as c (c.id)}
						<div class="g-col">
							{#each horas as h (h)}<span class="g-linea" style:top="{(h - rango.desde) * PX}px"
								></span>{/each}
							{#each data.turnos.filter((t) => t.profesionalId === c.id) as t (t.id)}
								<div class="g-turno" style={pos(t)}>{@render tarjeta(t, true)}</div>
							{/each}
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{:else}
		<div class="semana" use:anim={{ t: 'rise', delay: 60 }}>
			{#each semana as d (d.dia)}
				<div class="s-dia card" class:hoy={d.dia === data.hoy}>
					<a class="s-cab" href="/panel/turnos?dia={d.dia}&vista=dia"
						><span>{celular ? d.w : d.c}</span><b>{d.n}</b>{#if d.turnos.length}<span class="s-n"
								>{d.turnos.filter((t) => t.estado !== 'cancelado').length}</span
							>{/if}</a
					>
					{#each d.turnos as t (t.id)}{@render tarjeta(t, !celular)}{:else}<span class="s-vacio"
							>Sin turnos</span
						>{/each}
				</div>
			{/each}
		</div>
	{/if}
</div>

<HojaTurno bind:open={hoja} turno={abierto} cambiar={estado} {devolver} />

<Modal bind:open={nuevo} titulo="Nuevo turno" ancho={480}>
	<form onsubmit={(e) => crear(e)}>
		<div class="m-t">Nuevo turno</div>
		<div class="m-d">
			Se carga confirmado y sin seña. Si el cliente reserva desde la tienda, entra&nbsp;solo.
		</div>
		<div class="campos">
			<label class="field"
				>Servicio<select class="select" bind:value={n.servicioId}
					>{#each data.servicios as s (s.id)}<option value={s.id}>{s.nombre}</option>{/each}</select
				></label
			>
			<label class="field"
				>Quién atiende<select class="select" bind:value={n.profesionalId}
					>{#each data.profesionales as p (p.id)}<option value={p.id}>{p.nombre}</option
						>{/each}</select
				></label
			>
			<label class="field">Día<input class="input" type="date" bind:value={n.dia} required /></label
			>
			<label class="field"
				>Hora<select class="select" bind:value={n.hora}
					>{#each HORAS as h (h)}<option value={h}>{h}</option>{/each}</select
				></label
			>
			<label class="field ancho"
				>Cliente<input
					class="input"
					bind:value={n.nombre}
					maxlength="80"
					required
					autocomplete="off"
				/></label
			>
			<label class="field"
				>WhatsApp<input class="input" type="tel" bind:value={n.telefono} maxlength="20" /></label
			>
			<label class="field"
				>Email<input class="input" type="email" bind:value={n.email} maxlength="120" /></label
			>
			<label class="field ancho"
				>Nota<input
					class="input"
					bind:value={n.nota}
					maxlength="300"
					placeholder="Opcional"
				/></label
			>
		</div>
		{#if !data.servicios.length || !data.profesionales.length}
			<p class="aviso">
				Primero cargá servicios y quién atiende en <a href="/panel/turnos/ajustes">Ajustes</a>.
			</p>
		{/if}
		{#if superpuesto}
			<p class="aviso" role="alert">
				Ese horario se superpone con otro turno de la misma persona. ¿Lo cargás&nbsp;igual?
			</p>
		{/if}
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (nuevo = false)}
				>Cancelar</button
			>
			{#if superpuesto}
				<button
					type="button"
					class="btn btn-primary"
					disabled={guardando}
					onclick={(e) => crear(e as unknown as SubmitEvent, true)}>Cargar igual</button
				>
			{:else}
				<button class="btn btn-primary" disabled={!valido || guardando}>Cargar turno</button>
			{/if}
		</div>
	</form>
</Modal>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.cab {
		display: flex;
		align-items: end;
		gap: 10px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 200px;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.barra {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
	}
	.nav {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.fecha {
		font-size: 15px;
		font-weight: 600;
	}
	.vacio {
		padding: 28px;
		text-align: center;
		color: var(--mu);
	}
	.turno {
		width: 100%;
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 2px;
		padding: 12px 14px;
		border: 0;
		border-left: 4px solid var(--borde);
		border-radius: 14px;
		background: var(--sf);
		box-shadow: var(--shd);
		color: var(--tx);
		text-align: left;
		font-family: inherit;
		cursor: pointer;
		min-width: 0;
	}
	.turno .chip-estado {
		margin-top: 6px;
	}
	.turno.compacta {
		height: 100%;
		padding: 6px 9px;
		border-radius: 11px;
		overflow: hidden;
		gap: 0;
	}
	.turno.compacta .chip-estado {
		margin-top: 4px;
		padding: 2px 8px;
		font-size: 11.5px;
	}
	.turno.apagado {
		opacity: 0.55;
	}
	.turno.apagado .t-n {
		text-decoration: line-through;
	}
	.t-h {
		font-size: 12px;
		font-weight: 600;
		color: var(--mu);
	}
	.t-n {
		font-size: 14.5px;
		font-weight: 600;
		max-width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.compacta .t-n {
		font-size: 13px;
	}
	.t-s {
		font-size: 12.5px;
		color: var(--mu);
		max-width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.lista {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.hora-t {
		margin-top: 6px;
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	.grilla {
		padding: 14px 14px 18px;
		overflow-x: auto;
	}
	.g-cab,
	.g-cuerpo {
		display: grid;
		grid-template-columns: 52px repeat(var(--cols), minmax(150px, 1fr));
		gap: 8px;
		min-width: calc(52px + var(--cols) * 158px);
	}
	.g-cab {
		padding-bottom: 12px;
		margin-bottom: 8px;
		border-bottom: 1px solid var(--ln);
	}
	.g-pro {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		min-width: 0;
	}
	.g-pro img,
	.av {
		width: 30px;
		height: 30px;
		flex: none;
		border-radius: 50%;
		object-fit: cover;
	}
	.av {
		display: grid;
		place-items: center;
		background: var(--chip);
		font-size: 13px;
	}
	.g-cuerpo {
		position: relative;
	}
	.g-horas,
	.g-col {
		position: relative;
	}
	.g-horas span {
		position: absolute;
		left: 0;
		transform: translateY(-50%);
		font-size: 12px;
		color: var(--mu);
		font-variant-numeric: tabular-nums;
	}
	.g-horas span:first-child {
		transform: none;
	}
	.g-linea {
		position: absolute;
		left: 0;
		right: 0;
		border-top: 1px dashed var(--ln);
	}
	.g-turno {
		position: absolute;
		left: 0;
		right: 0;
	}
	.semana {
		display: grid;
		grid-template-columns: repeat(7, minmax(0, 1fr));
		gap: 8px;
		align-items: start;
	}
	.s-dia {
		padding: 10px;
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
	}
	.s-dia.hoy {
		outline: 2px solid rgba(210, 88, 47, 0.45);
	}
	.s-cab {
		display: flex;
		align-items: baseline;
		gap: 6px;
		padding: 2px 4px 6px;
		color: var(--tx);
		text-decoration: none;
		font-size: 13px;
		color: var(--mu);
	}
	.s-cab b {
		font-size: 18px;
		color: var(--tx);
	}
	.s-n {
		margin-left: auto;
		font-size: 11.5px;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 999px;
		background: var(--chip);
		color: var(--tx);
	}
	.s-dia .turno.compacta {
		height: auto;
	}
	.s-vacio {
		padding: 4px;
		font-size: 12.5px;
		color: var(--fa);
	}
	.m-t {
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14.5px;
		color: var(--mu);
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
	.aviso {
		margin: 12px 0 0;
		padding: 10px 12px;
		border-radius: 14px;
		background: rgba(227, 169, 59, 0.14);
		font-size: 13.5px;
	}
	.m-acc {
		margin-top: 18px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
	@media (max-width: 1099px) {
		.semana {
			grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
		}
	}
	@media (max-width: 759px) {
		.semana {
			grid-template-columns: minmax(0, 1fr);
		}
		.cab .btn {
			flex: 1 1 auto;
		}
		.campos {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
