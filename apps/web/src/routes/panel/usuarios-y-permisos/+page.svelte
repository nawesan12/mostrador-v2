<script lang="ts">
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import {
		AREAS,
		DETALLE_ROL,
		INVITABLES,
		NOMBRE_ROL,
		QUE_HACE,
		ROLES,
		type RolInvitable
	} from '#lib/panel/permisos.ts';
	import { PLANES } from '#lib/panel/planes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Usuarios y permisos (design/PanelCuenta.dc.html, sUsers): el equipo, las invitaciones y qué puede
	 * cada rol. Los roles son fijos (docs/decisiones.md §37): la matriz editable del handoff se cambió
	 * por la explicación de cada uno, porque Go aplica esos cuatro y nada más.
	 */
	let { data } = $props();
	const ui = getUi();
	type M = (typeof data.equipo)[number];

	const lugares = $derived(PLANES.find((p) => p.id === data.plan)?.personas ?? null);
	const libres = $derived(lugares === null ? null : Math.max(0, lugares - data.equipo.length));

	async function pedir(url: string, method: string, cuerpo?: unknown) {
		const r = await fetch(url, {
			method,
			headers: cuerpo === undefined ? {} : { 'content-type': 'application/json' },
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			ui.toast({ k: 'err', title: 'No se pudo', desc: m ?? 'Probá de nuevo.' });
		}
		return r.ok;
	}
	const fecha = (iso: string) =>
		new Date(iso).toLocaleDateString('es-AR', {
			day: 'numeric',
			month: 'short',
			timeZone: 'America/Argentina/Buenos_Aires'
		});
	const vencida = (m: M) => !!m.vence && Date.parse(m.vence) < Date.now();
	const estadoInvitacion = (m: M) =>
		!m.vence ? 'Sin mandar' : vencida(m) ? 'Vencida' : 'Invitación enviada';
	const ultimo = (m: M) => {
		if (m.pendiente)
			return !m.vence
				? 'Nunca se mandó: reenviala o cancelala'
				: vencida(m)
					? `Venció el ${fecha(m.vence)}`
					: `Vence el ${fecha(m.vence)}`;
		if (!m.ultimo) return 'Nunca entró';
		const min = Math.round((Date.now() - Date.parse(m.ultimo)) / 60_000);
		return min < 2
			? 'Activa ahora'
			: min < 60
				? `Hace ${min} min`
				: min < 1440
					? `Hace ${Math.round(min / 60)} h`
					: 'Hace días';
	};
	const punto = (m: M) =>
		m.pendiente
			? '#E3A93B'
			: m.ultimo && Date.now() - Date.parse(m.ultimo) < 3_600_000
				? '#2F8F63'
				: 'var(--ln2)';

	async function rol(m: M, r: RolInvitable) {
		if (m.vos || m.rol === 'duenio' || r === m.rol) return;
		if (!(await pedir(`/panel/equipo/${m.id}`, 'PUT', { rol: r }))) return;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: `${m.nombre.split(' ')[0]} ahora es ${NOMBRE_ROL[r]}`,
			desc: m.pendiente ? 'Entra con ese rol cuando acepte' : 'Aplica al instante'
		});
	}
	async function quitar(m: M) {
		if (!(await pedir(`/panel/equipo/${m.id}`, 'DELETE'))) return;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: m.pendiente
				? `Cancelaste la invitación a ${m.contacto}`
				: `${m.nombre} ya no tiene acceso`,
			desc: m.pendiente ? 'El link del mail ya no sirve' : 'Cerramos sus sesiones en esta tienda'
		});
	}
	let reenviando = $state('');
	async function reenviar(m: M) {
		reenviando = m.id;
		const ok = await pedir(`/panel/equipo/${m.id}/reenviar`, 'POST');
		reenviando = '';
		if (!ok) return;
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Invitación reenviada', desc: `${m.contacto} · vence en 7 días` });
	}

	// ---- Invitar ----
	let invitando = $state(false);
	let mandando = $state(false);
	let email = $state('');
	let nuevoRol = $state<RolInvitable>('ventas');
	const ICONO: Record<RolInvitable, I.IconData> = {
		admin: I.phShieldStar,
		ventas: I.phChatsCircle,
		local: I.phStorefront
	};
	async function invitar(e: SubmitEvent) {
		e.preventDefault();
		mandando = true;
		const ok = await pedir('/panel/equipo', 'POST', { contacto: email, rol: nuevoRol });
		mandando = false;
		if (!ok) return;
		invitando = false;
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Invitación enviada', desc: `${email} · ${NOMBRE_ROL[nuevoRol]}` });
		email = '';
	}
</script>

<Seo
	title="Usuarios y permisos · Panel"
	description="Tu equipo"
	path="/panel/usuarios-y-permisos"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Usuarios y permisos</h1>
			<div class="c-s">
				{data.equipo.length}
				{data.equipo.length === 1 ? 'persona' : 'personas'}{libres === null
					? ' · sin límite en tu plan'
					: ` · ${libres} ${libres === 1 ? 'lugar libre' : 'lugares libres'} en tu plan`}
			</div>
		</div>
		<button type="button" class="invitar" onclick={() => (invitando = true)}
			><Icon icon={I.phUserPlusBold} size={15} />Invitar persona</button
		>
	</div>

	<div class="lista" use:anim={{ t: 'rise', delay: 50 }}>
		{#each data.equipo as m (m.id)}
			<div class="persona">
				<span class="foto">
					{#if m.foto}<img
							src={unsplash(m.foto, 84, '&h=84&crop=faces')}
							alt=""
							class:pend={m.pendiente}
						/>{:else}<span class="ini" class:pend={m.pendiente}
							>{m.nombre.slice(0, 1).toUpperCase()}</span
						>{/if}
					<span class="punto" style:background={punto(m)}></span>
				</span>
				<span class="p-txt">
					<span class="p-n"
						>{m.nombre}{#if m.vos}<span class="vos">Vos</span>{/if}{#if m.pendiente}<span
								class="inv"
								class:apagada={!m.vence || vencida(m)}>{estadoInvitacion(m)}</span
							>{/if}</span
					>
					<span class="p-e">{m.contacto} · {ultimo(m)}</span>
				</span>
				<span class="areas"
					>{#each AREAS[m.rol] as a (a)}<span>{a}</span>{/each}</span
				>
				{#if m.rol === 'duenio'}
					<span class="duenia">Dueña</span>
				{:else}
					<span class="roles" role="group" aria-label="Rol de {m.nombre}" use:deslizable>
						{#each INVITABLES as r (r)}
							<button type="button" aria-pressed={m.rol === r} onclick={() => rol(m, r)}
								>{NOMBRE_ROL[r]}</button
							>
						{/each}
					</span>
				{/if}
				{#if m.pendiente || (!m.vos && m.rol !== 'duenio')}<span class="acc">
						{#if m.pendiente}
							<button
								type="button"
								class="reenviar"
								disabled={reenviando === m.id}
								onclick={() => reenviar(m)}
								><Icon icon={I.phPaperPlaneTilt} size={15} />{reenviando === m.id
									? 'Mandando…'
									: 'Reenviar'}</button
							>
						{/if}
						{#if !m.vos && m.rol !== 'duenio'}
							<button
								type="button"
								class="quitar"
								aria-label={m.pendiente
									? `Cancelar la invitación a ${m.contacto}`
									: `Quitar a ${m.nombre}`}
								onclick={() => quitar(m)}><Icon icon={I.phTrash} size={17} /></button
							>
						{/if}
					</span>{/if}
			</div>
		{/each}
	</div>

	<div class="tarjeta" use:anim={{ t: 'rise', delay: 100 }}>
		<div class="mz-cab">
			<span>Qué puede cada rol</span><span class="mz-s">Lo que un rol no puede, no lo ve</span>
		</div>
		<div class="roles-fijos">
			{#each ROLES as r (r)}
				<div class="rf">
					<b>{NOMBRE_ROL[r]}</b>
					<span class="rf-d">{QUE_HACE[r]}</span>
					<ul>
						{#each DETALLE_ROL[r].si as x (x)}<li class="si">
								<Icon icon={I.phCheckBold} size={13} />{x}
							</li>{/each}
						{#each DETALLE_ROL[r].no as x (x)}<li class="no">
								<Icon icon={I.phXBold} size={13} />{x}
							</li>{/each}
					</ul>
				</div>
			{/each}
		</div>
	</div>
</div>

<Modal bind:open={invitando} titulo="Invitar a tu equipo" ancho={480} padding={24}>
	<form onsubmit={invitar}>
		<b class="m-t">Invitar a tu equipo</b>
		<div class="m-d">Le llega un link por email. Sirve una sola vez y vence en 7&nbsp;días.</div>
		<input
			class="m-in"
			type="email"
			bind:value={email}
			placeholder="nombre@email.com"
			aria-label="Email"
			autocomplete="off"
			maxlength="120"
			required
		/>
		<div class="m-r">Rol</div>
		<div class="m-roles" role="group" aria-label="Rol">
			{#each INVITABLES as id (id)}
				<button type="button" aria-pressed={nuevoRol === id} onclick={() => (nuevoRol = id)}
					><Icon icon={ICONO[id]} size={20} /><span
						><b>{NOMBRE_ROL[id]}</b><span>{QUE_HACE[id]}</span></span
					></button
				>
			{/each}
		</div>
		<div class="m-acc">
			<button type="button" class="b-sec" onclick={() => (invitando = false)}>Cancelar</button>
			<button class="b-pri" disabled={email.trim().length < 6 || mandando}
				>{mandando ? 'Mandando…' : 'Mandar invitación'}</button
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
	.cab {
		display: flex;
		align-items: end;
		gap: 12px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 220px;
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
	.invitar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 20px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.lista {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.persona {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 14px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		flex-wrap: wrap;
	}
	.foto {
		position: relative;
		flex: none;
	}
	.foto img {
		width: 42px;
		height: 42px;
		border-radius: 50%;
		object-fit: cover;
		display: block;
		background: var(--chip);
	}
	.foto img.pend {
		opacity: 0.6;
	}
	.punto {
		position: absolute;
		right: -1px;
		bottom: -1px;
		width: 12px;
		height: 12px;
		border-radius: 50%;
		box-shadow: 0 0 0 2px var(--sf);
	}
	.p-txt {
		flex: 1 1 180px;
		min-width: 0;
	}
	.p-n {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-wrap: wrap;
		font-size: 14.5px;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.vos,
	.inv {
		font-size: 11px;
		font-weight: 700;
		padding: 2px 7px;
		border-radius: 999px;
		background: var(--chip);
		color: var(--mu);
	}
	.inv {
		background: rgba(227, 169, 59, 0.18);
		color: #8a5f0e;
	}
	.p-e {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		overflow-wrap: anywhere;
	}
	.areas {
		display: flex;
		gap: 4px;
		flex-wrap: wrap;
	}
	.areas span {
		font-size: 11.5px;
		font-weight: 600;
		padding: 3px 8px;
		border-radius: 7px;
		background: var(--chip);
		color: var(--mu);
	}
	.roles {
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
		max-width: 100%;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.roles button {
		border: 0;
		font:
			600 12px 'Instrument Sans',
			sans-serif;
		padding: 6px 10px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
		white-space: nowrap;
	}
	.roles button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.acc {
		display: flex;
		align-items: center;
		gap: 4px;
		margin-left: auto;
	}
	.duenia {
		font-size: 12px;
		font-weight: 600;
		padding: 7px 12px;
		border-radius: 999px;
		background: var(--chip);
	}
	.reenviar {
		display: flex;
		align-items: center;
		gap: 6px;
		min-height: 34px;
		padding: 6px 12px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.reenviar:disabled {
		opacity: 0.6;
	}
	.inv.apagada {
		background: var(--chip);
		color: var(--mu);
	}
	.mz-s {
		font-size: 12.5px;
		color: var(--mu);
		font-weight: 400;
	}
	.roles-fijos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
	}
	.rf {
		padding: 14px 18px 16px;
		border-top: 1px solid var(--ln);
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.rf b {
		font-size: 15px;
	}
	.rf-d {
		font-size: 12.5px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.rf ul {
		margin: 6px 0 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 5px;
		font-size: 13px;
	}
	.rf li {
		display: flex;
		align-items: center;
		gap: 7px;
	}
	.rf li.si :global(svg) {
		color: #2f8f63;
		flex: none;
	}
	.rf li.no {
		color: var(--mu);
	}
	.rf li.no :global(svg) {
		color: #b0441d;
		flex: none;
	}
	.quitar {
		width: 34px;
		height: 34px;
		border-radius: 10px;
		border: 0;
		background: transparent;
		color: var(--mu);
		display: grid;
		place-items: center;
	}
	.tarjeta {
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		overflow: hidden;
	}
	.mz-cab {
		padding: 14px 18px;
		border-bottom: 1px solid var(--ln);
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
		font-size: 16px;
		font-weight: 600;
	}
	.ini {
		display: grid;
		place-items: center;
		width: 42px;
		height: 42px;
		border-radius: 50%;
		background: var(--durazno);
		color: #8a3a1c;
		font-weight: 700;
		font-size: 17px;
	}
	.ini.pend {
		opacity: 0.55;
	}
	/* Celular: el rol de cada persona ocupa su propia fila con los tres a la vista (antes se
	   cortaba con un scroll escondido). */
	@media (max-width: 760px) {
		.roles {
			flex: 1 1 100%;
			order: 5;
			display: grid;
			grid-template-columns: repeat(3, minmax(0, 1fr));
			overflow: visible;
		}
		.roles button {
			min-height: 40px;
			padding: 6px 4px;
		}
	}
	.m-t {
		font-size: 19px;
	}
	.m-d {
		font-size: 13.5px;
		color: var(--mu);
		margin-top: 4px;
	}
	.m-in {
		margin-top: 14px;
		width: 100%;
		padding: 13px 14px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: var(--sf2);
		font:
			500 15px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
	}
	.m-r {
		margin-top: 12px;
		font-size: 13px;
		font-weight: 600;
	}
	.m-roles {
		margin-top: 8px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.m-roles button {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 14px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		text-align: left;
		font: inherit;
	}
	.m-roles button[aria-pressed='true'] {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.06);
	}
	.m-roles b {
		display: block;
		font-size: 14px;
	}
	.m-roles span span {
		font-size: 12.5px;
		color: var(--mu);
	}
	.m-acc {
		margin-top: 16px;
		display: flex;
		gap: 8px;
	}
	.m-acc button {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
	}
	.b-sec {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.b-pri {
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
	}
	.b-pri:disabled {
		opacity: 0.55;
	}
</style>
