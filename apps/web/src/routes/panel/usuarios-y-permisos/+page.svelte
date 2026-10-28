<script lang="ts">
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { AREAS, fijo, NOMBRE_ROL, ROLES, type Rol } from '#lib/panel/permisos.ts';
	import { PLANES } from '#lib/panel/planes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Usuarios y permisos (design/PanelCuenta.dc.html, sUsers): equipo, roles y la matriz de permisos. */
	let { data } = $props();
	const ui = getUi();
	type M = (typeof data.equipo)[number];

	const lugares = $derived(PLANES.find((p) => p.id === data.plan)?.personas ?? null);
	const libres = $derived(lugares === null ? null : Math.max(0, lugares - data.equipo.length));
	const sin2fa = $derived(data.equipo.filter((m) => !m.pendiente && !m.dobleFactor));
	const activos = $derived(data.equipo.filter((m) => !m.pendiente));

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
	const ultimo = (m: M) => {
		if (m.pendiente) return 'Invitación enviada';
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

	async function rol(m: M, r: Rol) {
		if (m.vos || r === 'duenio' || r === m.rol) return;
		if (!(await pedir(`/panel/equipo/${m.id}`, 'PUT', { rol: r }))) return;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: `${m.nombre.split(' ')[0]} ahora es ${NOMBRE_ROL[r]}`,
			desc: 'Le avisamos por email'
		});
	}
	async function quitar(m: M) {
		if (!(await pedir(`/panel/equipo/${m.id}`, 'DELETE'))) return;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: `${m.nombre} ya no tiene acceso`,
			desc: m.pendiente ? 'Retiramos la invitación' : 'Cerramos todas sus sesiones'
		});
	}
	async function permiso(id: string, titulo: string, r: Rol, on: boolean) {
		if (fijo(id, r)) return;
		if (!(await pedir(`/panel/permisos/${id}/${r}`, 'PUT', { activo: !on }))) return;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: `${NOMBRE_ROL[r]} ${on ? 'ya no puede' : 'ahora puede'}: ${titulo.toLowerCase()}`,
			desc: 'Aplica al instante'
		});
	}

	// ---- Invitar ----
	let invitando = $state(false);
	let contacto = $state('');
	let nuevoRol = $state<'admin' | 'ventas' | 'deposito'>('ventas');
	async function invitar(e: SubmitEvent) {
		e.preventDefault();
		if (!(await pedir('/panel/equipo', 'POST', { contacto, rol: nuevoRol }))) return;
		invitando = false;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Invitación enviada',
			desc: `${contacto} · ${NOMBRE_ROL[nuevoRol]}`
		});
		contacto = '';
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
						/>{:else}<img src="/hornero/avatar.webp" alt="" class:pend={m.pendiente} />{/if}
					<span class="punto" style:background={punto(m)}></span>
				</span>
				<span class="p-txt">
					<span class="p-n"
						>{m.nombre}{#if m.vos}<span class="vos">Vos</span>{/if}{#if m.pendiente}<span
								class="inv">Invitación enviada</span
							>{/if}</span
					>
					<span class="p-e">{m.contacto} · {ultimo(m)}</span>
				</span>
				<span class="areas"
					>{#each AREAS[m.rol] as a (a)}<span>{a}</span>{/each}</span
				>
				<span class="roles" role="group" aria-label="Rol de {m.nombre}" use:deslizable>
					{#each ROLES as r (r)}
						<button
							type="button"
							aria-pressed={m.rol === r}
							class:apagado={m.vos && r !== 'duenio'}
							onclick={() => rol(m, r)}>{NOMBRE_ROL[r]}</button
						>
					{/each}
				</span>
				{#if !m.vos}
					<button
						type="button"
						class="quitar"
						aria-label="Quitar a {m.nombre}"
						onclick={() => quitar(m)}><Icon icon={I.phTrash} size={17} /></button
					>
				{/if}
			</div>
		{/each}
	</div>

	<div class="tarjeta matriz" use:anim={{ t: 'rise', delay: 100 }}>
		<div class="mz-cab">
			<span>Qué puede hacer cada rol</span><span class="mz-s">Tocá un permiso para cambiarlo</span>
		</div>
		<div class="desliza" use:deslizable>
			<div class="tabla">
				<div class="fila cab-t">
					<span>Permiso</span>{#each ROLES as r (r)}<span>{NOMBRE_ROL[r]}</span>{/each}
				</div>
				{#each data.permisos as p (p.id)}
					<div class="fila">
						<span><b>{p.titulo}</b><span class="d">{p.detalle}</span></span>
						{#each ROLES as r (r)}
							{@const on = p.roles[r]}
							<span class="celda">
								<button
									type="button"
									class:on
									class:fijo={fijo(p.id, r)}
									aria-pressed={on}
									aria-label="{NOMBRE_ROL[r]}: {p.titulo}"
									onclick={() => permiso(p.id, p.titulo, r, on)}
									>{#if on}<Icon icon={I.phCheckBold} size={14} />{/if}</button
								>
							</span>
						{/each}
					</div>
				{/each}
			</div>
		</div>
	</div>

	<div class="doble" use:anim={{ t: 'rise', delay: 150 }}>
		<Icon icon={I.phShieldCheckFill} size={22} style="color:#237550" />
		<span
			><b>Verificación en dos pasos obligatoria</b> para Dueños y Administradores. {activos.length -
				sin2fa.length} de {activos.length}
			{activos.length === 1 ? 'persona ya la activó' : 'personas ya la activaron'}.</span
		>
		{#if sin2fa[0]}
			<button
				type="button"
				onclick={() =>
					ui.toast({
						k: 'wa',
						title: `Le mandamos un WhatsApp a ${sin2fa[0].nombre.split(' ')[0]}`,
						desc: 'Con el paso a paso para activar la verificación'
					})}>Recordarle a {sin2fa[0].nombre.split(' ')[0]}</button
			>
		{/if}
	</div>
</div>

<Modal bind:open={invitando} titulo="Invitar a tu equipo" ancho={480} padding={24}>
	<form onsubmit={invitar}>
		<b class="m-t">Invitar a tu equipo</b>
		<div class="m-d">Le llega un link por email o WhatsApp. Vence en 7 días.</div>
		<input
			class="m-in"
			bind:value={contacto}
			placeholder="email o WhatsApp"
			aria-label="Email o WhatsApp"
			maxlength="120"
		/>
		<div class="m-r">Rol</div>
		<div class="m-roles" role="group" aria-label="Rol">
			{#each [['admin', I.phShieldStar, 'Todo menos plan y usuarios'], ['ventas', I.phChatsCircle, 'Pedidos, clientes y WhatsApp'], ['deposito', I.phPackage, 'Stock, preparación y envíos']] as const as [id, icono, d] (id)}
				<button type="button" aria-pressed={nuevoRol === id} onclick={() => (nuevoRol = id)}
					><Icon icon={icono} size={20} /><span><b>{NOMBRE_ROL[id]}</b><span>{d}</span></span
					></button
				>
			{/each}
		</div>
		<div class="m-acc">
			<button type="button" class="b-sec" onclick={() => (invitando = false)}>Cancelar</button>
			<button class="b-pri" disabled={contacto.trim().length < 6}>Mandar invitación</button>
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
	.roles button.apagado {
		opacity: 0.4;
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
	.mz-s {
		font-size: 12.5px;
		color: var(--mu);
		font-weight: 400;
	}
	.desliza {
		overflow-x: auto;
	}
	.tabla {
		min-width: 560px;
	}
	.fila {
		display: grid;
		grid-template-columns: minmax(180px, 1.6fr) repeat(4, minmax(80px, 1fr));
		padding: 10px 18px;
		border-top: 1px solid var(--ln);
		align-items: center;
		font-size: 14px;
	}
	.fila b {
		font-weight: 600;
	}
	.d {
		display: block;
		font-size: 12px;
		color: var(--mu);
	}
	.cab-t {
		border-top: 0;
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--mu);
		background: var(--sf2);
	}
	.cab-t span:not(:first-child) {
		text-align: center;
	}
	.celda {
		display: grid;
		place-items: center;
	}
	.celda button {
		width: 28px;
		height: 28px;
		border-radius: 9px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: #fff;
		display: grid;
		place-items: center;
	}
	.celda button.on {
		border-color: #d2582f;
		background: #d2582f;
	}
	.celda button.on.fijo {
		background: #e8a589;
		border-color: #e8a589;
	}
	.celda button.fijo {
		cursor: not-allowed;
	}
	.doble {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
		padding: 14px 18px;
		border-radius: 20px;
		background: var(--sf2);
		border: 1px solid var(--ln);
	}
	.doble > span {
		flex: 1;
		min-width: 220px;
		font-size: 14px;
	}
	.doble button {
		padding: 9px 14px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font:
			600 13px 'Instrument Sans',
			sans-serif;
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
