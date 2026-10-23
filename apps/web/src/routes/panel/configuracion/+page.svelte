<script lang="ts">
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { PLANES } from '#lib/panel/planes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Configuración (design/PanelOperacion.dc.html, sCfg): negocio, equipo, seguridad, avisos y plan. */
	let { data } = $props();
	const ui = getUi();

	const TABS = [
		['negocio', I.phStorefrontFill, 'Negocio'],
		['equipo', I.phUsersThreeFill, 'Equipo y permisos'],
		['seguridad', I.phShieldCheckFill, 'Seguridad'],
		['avisos', I.phBellRingingFill, 'Notificaciones'],
		['plan', I.phCrownSimpleFill, 'Plan']
	] as const;
	let tab = $state<(typeof TABS)[number][0]>('negocio');

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

	// ---- Negocio ----
	const CAMPOS = [
		['nombre', 'Nombre de fantasía'],
		['razonSocial', 'Razón social'],
		['cuit', 'CUIT'],
		['condicion', 'Condición fiscal'],
		['direccion', 'Dirección'],
		['whatsapp', 'WhatsApp'],
		['email', 'Email'],
		['horario', 'Horario'],
		['alias', 'Alias para transferencias'],
		['cbu', 'CBU o CVU'],
		['titular', 'Titular de la cuenta']
	] as const;
	let negocio = $derived({ ...data.negocio });
	const cambiado = $derived(CAMPOS.some(([k]) => negocio[k] !== data.negocio[k]));
	async function guardarNegocio() {
		if (!(await pedir('/panel/configuracion/negocio', 'PUT', negocio))) return;
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Datos guardados', desc: 'Se ven en la tienda y en tus facturas' });
	}

	// ---- Equipo ----
	const ROL = { duenio: 'Dueña', admin: 'Admin', ventas: 'Ventas', deposito: 'Depósito' } as const;
	const AREAS = {
		duenio: ['Todo'],
		admin: ['Todo menos plan'],
		ventas: ['Pedidos', 'Bandeja'],
		deposito: ['Stock', 'Envíos']
	} as const;
	let invitando = $state(false);
	let contacto = $state('');
	let rol = $state<'admin' | 'ventas' | 'deposito'>('ventas');
	async function invitar(e: SubmitEvent) {
		e.preventDefault();
		if (!(await pedir('/panel/equipo', 'POST', { contacto, rol }))) return;
		invitando = false;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Invitación enviada',
			desc: 'Le pedimos que active la doble verificación al entrar'
		});
		contacto = '';
	}

	// ---- Seguridad y avisos ----
	const SEGURIDAD = [
		[
			'dobleFactor',
			I.phDeviceMobileFill,
			'Doble verificación para todo el equipo',
			'Obligatoria · con app o SMS'
		],
		[
			'otroPais',
			I.phMapPinFill,
			'Avisarme si entran desde otro país',
			'Bloquea y te pregunta por WhatsApp'
		],
		['alertas', I.phBellRingingFill, 'Alertas de cambios sensibles', 'Precios, CBU y permisos'],
		[
			'inactivas',
			I.phClockCounterClockwiseFill,
			'Cerrar sesiones inactivas a los 7 días',
			'Recomendado para cajas compartidas'
		]
	] as const;
	const AVISOS = [
		['venta', 'Cada venta', 'Push y sonido de caja'],
		['chats', 'Chats que te esperan', 'Solo si Hornero no puede responder'],
		['stock', 'Stock bajo', 'Cuando quede menos de 1 semana'],
		['diario', 'Resumen diario por WhatsApp', 'Todos los días a las 21 h'],
		['ideas', 'Ideas de Hornero', 'Máximo 3 por día'],
		['novedades', 'Novedades de Mostrador', 'Una vez por mes']
	] as const;
	async function opcion(grupo: 'seguridad' | 'avisos', clave: string, activo: boolean) {
		if (await pedir(`/panel/configuracion/${grupo}/${clave}`, 'PUT', { activo }))
			await invalidateAll();
	}
	const SESION = {
		compu: I.phDesktopFill,
		celu: I.phDeviceMobileFill,
		caja: I.phDesktopTowerFill
	} as const;
	const hace = (iso: string) => {
		const min = Math.round((Date.now() - Date.parse(iso)) / 60_000);
		return min < 1 ? 'ahora' : min < 60 ? `hace ${min} min` : `hace ${Math.round(min / 60)} h`;
	};
	async function cerrar(id: string, nombre: string) {
		if (!(await pedir(`/panel/sesiones/${id}`, 'DELETE'))) return;
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Sesión cerrada', desc: nombre });
	}

	// ---- Plan ----
	async function elegir(id: (typeof PLANES)[number]['id'], nombre: string) {
		if (id === data.plan.actual) return;
		const r0 = await pedir('/panel/plan', 'PUT', { plan: id });
		if (!r0) return;
		// Plan pago con el abono real: se carga la tarjeta en Mercado Pago.
		if (r0.status === 200) {
			location.href = ((await r0.json()) as { url: string }).url;
			return;
		}
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: `Plan ${nombre} elegido`,
			desc: data.plan.enPrueba
				? 'Se cobra cuando termina tu prueba · sin permanencia'
				: 'Cobramos la diferencia prorrateada'
		});
	}
</script>

<Seo
	title="Configuración · Panel"
	description="Configuración de la tienda"
	path="/panel/configuracion"
	noindex
/>

<div class="pagina">
	<div use:anim={{ t: 'rise' }}>
		<h1>Configuración</h1>
		<div class="c-s">Tu negocio, tu equipo y tu seguridad</div>
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
			{#if tab === 'negocio'}
				<form
					class="tarjeta campos"
					onsubmit={(e) => {
						e.preventDefault();
						void guardarNegocio();
					}}
				>
					{#each CAMPOS as [k, t] (k)}
						<label
							>{t}<input
								value={negocio[k]}
								oninput={(e) => (negocio = { ...negocio, [k]: e.currentTarget.value })}
								maxlength={k === 'direccion' ? 160 : 120}
							/></label
						>
					{/each}
					{#if cambiado}
						<div class="guardar">
							<button
								type="button"
								class="btn btn-secondary"
								onclick={() => (negocio = { ...data.negocio })}>Descartar</button
							>
							<button class="btn btn-primary">Guardar cambios</button>
						</div>
					{/if}
				</form>
			{:else if tab === 'equipo'}
				<div class="tarjeta">
					<div class="t-cab">
						<span
							>Equipo · {data.equipo.length}
							{data.equipo.length === 1 ? 'persona' : 'personas'}</span
						>
						<button type="button" class="oscuro" onclick={() => (invitando = true)}>Invitar</button>
					</div>
					<div class="miembros">
						{#each data.equipo as m (m.id)}
							<div class="miembro" class:pendiente={m.pendiente}>
								{#if m.foto}<img
										src={unsplash(m.foto, 76, '&h=76&crop=faces')}
										alt=""
									/>{:else}<span class="av">{m.nombre[0].toUpperCase()}</span>{/if}
								<span class="m-txt"
									><span class="m-n">{m.nombre}{m.vos ? ' (vos)' : ''}</span><span class="m-e"
										>{m.pendiente ? 'Invitación enviada' : m.contacto}</span
									></span
								>
								<span class="permisos"
									>{#each AREAS[m.rol] as p (p)}<span>{p}</span>{/each}</span
								>
								<span class="rol" class:duenio={m.rol === 'duenio'}>{ROL[m.rol]}</span>
								{#if m.dobleFactor}<span title="Doble verificación activa"
										><Icon icon={I.phShieldCheckFill} size={17} style="color:#2F8F63" /></span
									>{/if}
							</div>
						{/each}
					</div>
					<a class="mas" href={resolve('/panel/[seccion]', { seccion: 'usuarios-y-permisos' })}
						>Roles y permisos de cada uno <Icon icon={I.phArrowRightBold} size={13} /></a
					>
				</div>
			{:else if tab === 'seguridad'}
				<div class="tarjeta lista">
					{#each SEGURIDAD as [k, icono, t, d] (k)}
						<div class="opcion">
							<span class="o-ico"><Icon icon={icono} size={18} /></span>
							<span class="o-txt"><span class="o-t">{t}</span><span class="o-d">{d}</span></span>
							<Interruptor
								checked={data.seguridad[k]}
								label={t}
								onchange={(v) => opcion('seguridad', k, v)}
							/>
						</div>
					{/each}
					<div class="sesiones-t">Sesiones activas</div>
					{#each data.sesiones as s (s.id)}
						<div class="sesion">
							<Icon icon={SESION[s.tipo]} size={18} style="color:var(--mu)" />
							<span class="s-t">{s.nombre} <span>· {s.donde} · {hace(s.cuando)}</span></span>
							{#if s.actual}<span class="esta">Esta</span>{:else}<button
									type="button"
									onclick={() => cerrar(s.id, s.nombre)}>Cerrar</button
								>{/if}
						</div>
					{/each}
				</div>
			{:else if tab === 'avisos'}
				<div class="tarjeta lista">
					{#each AVISOS as [k, t, d] (k)}
						<div class="opcion">
							<span class="o-txt"><span class="o-t">{t}</span><span class="o-d">{d}</span></span>
							<Interruptor
								checked={data.avisos[k]}
								label={t}
								onchange={(v) => opcion('avisos', k, v)}
							/>
						</div>
					{/each}
				</div>
			{:else}
				<div class="planes">
					{#each PLANES as p (p.id)}
						{@const actual = data.plan.actual === p.id}
						<div class="plan" class:actual>
							<div class="p-cab">
								<span>{p.nombre}</span>{#if actual}<span class="tuyo"
										>{data.plan.enPrueba ? 'Tu prueba' : 'Tu plan'}</span
									>{/if}
							</div>
							<div class="precio">
								{pesos(p.precio)}<span> /mes</span>
							</div>
							<div class="incluye">
								{#each p.incluye as f (f)}<span
										><Icon icon={I.phCheckBold} size={13} style="color:#2F8F63" />{f}</span
									>{/each}
							</div>
							<button type="button" onclick={() => elegir(p.id, p.nombre)}
								>{actual ? `Quedarme en ${p.nombre}` : `Elegir ${p.nombre}`}</button
							>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>

<Modal bind:open={invitando} titulo="Invitar al equipo" ancho={440}>
	<form onsubmit={invitar}>
		<div class="m-t">Invitar al equipo</div>
		<div class="m-d">Le llega la invitación y entra con su propio usuario.</div>
		<label class="field m-campo"
			>Email o WhatsApp<input
				class="input"
				bind:value={contacto}
				maxlength="120"
				placeholder="nombre@mail.com o 11 5555 5555"
			/></label
		>
		<div class="roles" role="group" aria-label="Rol">
			{#each [['admin', 'Admin', 'Todo menos plan y usuarios'], ['ventas', 'Ventas', 'Pedidos, clientes y WhatsApp'], ['deposito', 'Depósito', 'Stock, preparación y envíos']] as const as [id, t, d] (id)}
				<button type="button" aria-pressed={rol === id} onclick={() => (rol = id)}
					><b>{t}</b><span>{d}</span></button
				>
			{/each}
		</div>
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (invitando = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary" disabled={contacto.trim().length < 6}
				>Mandar invitación</button
			>
		</div>
	</form>
</Modal>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
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
	.campos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 12px;
	}
	.campos label {
		display: flex;
		flex-direction: column;
		gap: 5px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.campos input {
		padding: 11px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font:
			500 14.5px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
	}
	.guardar {
		grid-column: 1 / -1;
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}
	.t-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
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
	}
	.miembros {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.miembro {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid var(--ln);
		flex-wrap: wrap;
	}
	.miembro.pendiente {
		opacity: 0.6;
	}
	.miembro img,
	.av {
		width: 38px;
		height: 38px;
		border-radius: 50%;
		object-fit: cover;
	}
	.av {
		display: grid;
		place-items: center;
		background: var(--chip);
		font-weight: 600;
	}
	.m-txt {
		flex: 1 1 160px;
		min-width: 0;
	}
	.m-n {
		display: block;
		font-size: 14px;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.m-e {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		overflow-wrap: anywhere;
	}
	.permisos {
		display: flex;
		gap: 4px;
		flex-wrap: wrap;
	}
	.permisos span {
		font-size: 11.5px;
		font-weight: 600;
		padding: 3px 8px;
		border-radius: 999px;
		background: var(--chip);
	}
	.rol {
		font-size: 12.5px;
		font-weight: 600;
		padding: 5px 11px;
		border-radius: 999px;
		background: var(--chip);
	}
	.rol.duenio {
		background: rgba(210, 88, 47, 0.12);
		color: #c24e25;
	}
	.mas {
		margin-top: 12px;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
		color: #c24e25;
		text-decoration: none;
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
	.sesiones-t {
		margin-top: 10px;
		font-size: 14px;
		font-weight: 600;
	}
	.sesion {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 9px 0;
		font-size: 13.5px;
	}
	.s-t {
		flex: 1;
	}
	.s-t span {
		color: var(--mu);
	}
	.esta {
		font-size: 12px;
		font-weight: 600;
		color: #2f8f63;
	}
	.sesion button {
		border: 0;
		background: transparent;
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		color: #b83a2e;
	}
	.planes {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 12px;
	}
	.plan {
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--shd);
		border: 2px solid transparent;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.plan.actual {
		background: #2a140c;
		color: #fbf7f2;
		border-color: #d2582f;
	}
	.p-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 17px;
		font-weight: 600;
	}
	.tuyo {
		font-size: 11.5px;
		font-weight: 700;
		padding: 3px 9px;
		border-radius: 999px;
		background: #d2582f;
		color: #fff;
	}
	.precio {
		font-size: 30px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.precio span {
		font-size: 13px;
		font-weight: 500;
		opacity: 0.7;
		letter-spacing: 0;
	}
	.incluye {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 13.5px;
	}
	.incluye span {
		display: flex;
		gap: 7px;
		align-items: center;
	}
	.plan button {
		margin-top: auto;
		padding: 11px;
		border-radius: 999px;
		border: 0;
		background: var(--chip);
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.plan.actual button {
		background: #d2582f;
		color: #fff;
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
	.m-campo {
		margin-top: 16px;
	}
	.roles {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.roles button {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 2px;
		padding: 10px 12px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		text-align: left;
		font:
			500 13px 'Instrument Sans',
			sans-serif;
	}
	.roles button span {
		color: var(--mu);
	}
	.roles button[aria-pressed='true'] {
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
</style>
