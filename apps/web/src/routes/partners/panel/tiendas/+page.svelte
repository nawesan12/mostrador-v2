<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import Vacio from '#lib/components/Vacio.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { fechaCorta, linkDe, nombreMes } from '#lib/partners/programa.ts';
	import { ESTADO_TIENDA, iniciales, nombrePlan, tono } from '#lib/partners/vista.ts';
	import { SITE } from '#lib/site.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Mis tiendas: las que trajo el partner, sumar una nueva y las que colabora. */
	let { data } = $props();
	const ui = getUi();
	const NB = String.fromCharCode(160);
	const p = $derived(data.panel!);

	type Filtro = 'todas' | 'activa' | 'prueba' | 'perdidas';
	let filtro = $state<Filtro>('todas');
	const en = (f: Filtro, e: string) =>
		f === 'todas' ? true : f === 'perdidas' ? e === 'baja' || e === 'gratis' : e === f;
	const cuenta = (f: Filtro) => p.tiendas.filter((t) => en(f, t.estado)).length;
	const opciones = $derived([
		['todas', `Todas · ${cuenta('todas')}`],
		['activa', `Activas · ${cuenta('activa')}`],
		['prueba', `En prueba · ${cuenta('prueba')}`],
		['perdidas', `Perdidas · ${cuenta('perdidas')}`]
	] as const);
	const lista = $derived(p.tiendas.filter((t) => en(filtro, t.estado)));
	const mesFin = (fecha: string) => nombreMes(fecha.slice(0, 7));
	const ROL = { admin: 'Admin', ventas: 'Ventas', deposito: 'Depósito' } as const;

	// ---- Sumar una tienda ----
	let sumando = $state(false);
	let como = $state<'invitar' | 'crear'>('invitar');
	let email = $state('');
	const link = $derived(linkDe(p.partner.codigo));
	const mailValido = $derived(/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()));
	const mailto = $derived(
		`mailto:${encodeURIComponent(email.trim())}?subject=${encodeURIComponent('Tu tienda en Mostrador')}&body=${encodeURIComponent(
			`¡Hola! Te armo la tienda en Mostrador. Creá tu cuenta desde este link (14 días gratis, sin tarjeta) y sumame a tu equipo para que pueda cargar todo:\n\n${link}\n\nEl cupón ${p.partner.cupon} queda guardado en tu cuenta.`
		)}`
	);
</script>

<div class="tarjeta" use:anim={{ t: 'rise' }}>
	<div class="t-cab">
		<span>Tiendas que trajiste</span>
		<button type="button" class="oscuro" onclick={() => (sumando = true)}
			><Icon icon={I.phPlusBold} size={14} />Sumar una tienda</button
		>
	</div>
	<div class="t-d">
		Comisión por 12 meses desde el alta de cada una, mientras tenga un plan&nbsp;pago.
	</div>
	<div class="filtro">
		<Segmentado {opciones} bind:value={filtro} label="Filtrar tiendas" />
	</div>
	{#if lista.length}
		<div class="filas" role="list" aria-label="Tiendas referidas">
			{#each lista as t (t.id)}
				{@const c = ESTADO_TIENDA[t.estado]}
				{@const [bg, fg] = tono(t.nombre)}
				<div class="f-in" role="listitem">
					<span class="ini" style:background={bg} style:color={fg}>{iniciales(t.nombre)}</span>
					<span class="quien">
						<span class="q-t">{t.nombre}</span>
						<span class="q-s"
							>{nombrePlan(t.plan)}{t.ciclo === 'anio' ? ' anual' : ''} · desde{NB}{fechaCorta(
								t.alta
							)}</span
						>
					</span>
					<span class="chip-f" style:background={c.bg} style:color={c.fg}>{c.t}</span>
					<span class="valor">
						<span class="v-v">{t.comisionMes ? pesos(t.comisionMes) : '—'}</span>
						<span class="v-s hasta"
							>{t.comisionMes ? `por mes hasta ${mesFin(t.finComision)}` : 'sin comisión'}</span
						>
					</span>
				</div>
			{/each}
		</div>
	{:else}
		<div class="vacio-in">
			<Vacio
				hornero="lupa"
				titulo={p.tiendas.length ? 'Ninguna en este grupo' : 'Todavía no trajiste tiendas'}
				texto={p.tiendas.length
					? 'Probá con otro filtro.'
					: 'Compartí tu link: cada tienda que se crea con él queda a tu nombre.'}
			/>
		</div>
	{/if}
</div>

<div class="tarjeta" use:anim={{ t: 'rise', delay: 40 }}>
	<div class="t-cab"><span>Tiendas donde colaborás</span></div>
	<div class="t-d">Las dueñas te sumaron a su equipo. Entrás con tu misma&nbsp;cuenta.</div>
	{#if p.colaboraciones.length}
		<div class="filas">
			{#each p.colaboraciones as c (c.tiendaId)}
				{@const [bg, fg] = tono(c.nombre)}
				<!-- eslint-disable svelte/no-navigation-without-resolve -- la tienda pública de otra dueña -->
				<a class="f-in" href="{SITE}/t/{c.slug}" target="_blank" rel="noopener">
					<span class="ini" style:background={bg} style:color={fg}>{iniciales(c.nombre)}</span>
					<span class="quien">
						<span class="q-t">{c.nombre}</span>
						<span class="q-s">/t/{c.slug}</span>
					</span>
					<span class="chip-f" style:background="var(--chip)">{ROL[c.rol]}</span>
					<Icon icon={I.phArrowUpRightBold} size={15} style="color:var(--mu)" />
				</a>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
			{/each}
		</div>
	{:else}
		<div class="t-d vacio-t">
			Cuando una dueña te sume desde Usuarios y permisos, la vas a ver&nbsp;acá.
		</div>
	{/if}
</div>

<Modal bind:open={sumando} titulo="Sumar una tienda" ancho={460}>
	<div class="pm">
		<div class="m-t">Sumar una tienda</div>
		<div class="m-d">Queda a tu nombre si se crea con tu link o con tu&nbsp;cupón.</div>
		<div class="roles m-campo" role="group" aria-label="Cómo">
			<button type="button" aria-pressed={como === 'invitar'} onclick={() => (como = 'invitar')}
				><b>Invitar a la dueña por mail</b><span>Le llega tu link y crea su cuenta.</span></button
			>
			<button type="button" aria-pressed={como === 'crear'} onclick={() => (como = 'crear')}
				><b>La creo yo y después se la paso</b><span
					>Armala con tu link y sumá a la dueña desde Usuarios y&nbsp;permisos.</span
				></button
			>
		</div>
		{#if como === 'invitar'}
			<label class="field m-campo"
				>Email de la dueña<input
					class="input"
					type="email"
					bind:value={email}
					maxlength="254"
					placeholder="nombre@mail.com"
				/></label
			>
		{/if}
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (sumando = false)}
				>Cancelar</button
			>
			<!-- eslint-disable svelte/no-navigation-without-resolve -- mailto y el link de referido -->
			{#if como === 'invitar'}
				<a
					class="btn btn-primary"
					class:apagado={!mailValido}
					aria-disabled={!mailValido}
					href={mailValido ? mailto : undefined}
					onclick={() => {
						if (!mailValido) return;
						sumando = false;
						ui.toast({ k: 'ok', title: 'Invitación lista', desc: 'Se abrió tu mail con el link' });
					}}>Armar el mail</a
				>
			{:else}
				<a class="btn btn-primary" href={link} target="_blank" rel="noopener">Crear con mi link</a>
			{/if}
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
		</div>
	</div>
</Modal>

<style>
	.filtro {
		margin-top: 14px;
		max-width: 100%;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.vacio-in {
		margin-top: 12px;
	}
	.vacio-t {
		margin-top: 10px;
	}
	.apagado {
		opacity: 0.5;
		pointer-events: none;
	}
	@media (max-width: 759px) {
		.f-in .chip-f,
		.valor .hasta {
			display: none;
		}
	}
</style>
