<script lang="ts">
	import { untrack } from 'svelte';
	import { invalidateAll } from '$app/navigation';
	import Alerta from '#lib/components/Alerta.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { nivel, type PerfilPartner } from '#lib/partners/programa.ts';
	import { validarPerfil } from '#lib/partners/validar.ts';
	import { guardar, iniciales } from '#lib/partners/vista.ts';
	import { ESPECIALIDADES } from '#lib/sitio/opciones.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Perfil público: lo que se ve en "Buscar un partner" (/partners), con la vista previa al lado. */
	let { data } = $props();
	const ui = getUi();
	const NB = String.fromCharCode(160);
	const p = $derived(data.panel!);

	const copia = (x: PerfilPartner): PerfilPartner => ({
		...x,
		especialidades: [...x.especialidades],
		casos: x.casos.map((c) => ({ ...c }))
	});
	// Copia editable; vuelve a lo guardado cada vez que llegan datos nuevos.
	// eslint-disable-next-line svelte/prefer-writable-derived -- se edita campo por campo (bind): hace falta el proxy profundo de $state
	let perfil = $state(copia(untrack(() => data.panel!.partner.perfil)));
	$effect.pre(() => {
		perfil = copia(p.partner.perfil);
	});
	const cambiado = $derived(JSON.stringify(perfil) !== JSON.stringify(p.partner.perfil));
	let error = $state<string | null>(null);

	const activas = $derived(p.tiendas.filter((t) => t.estado === 'activa').length);
	const n = $derived(nivel(p.partner.nivelEfectivo));
	const nombreEsp = (k: string) => ESPECIALIDADES.find(([e]) => e === k)?.[1] ?? k;
	const linea = $derived(
		[perfil.especialidades.map(nombreEsp).slice(0, 2).join(' y '), perfil.ciudad]
			.filter(Boolean)
			.join(' · ')
	);

	function alternar(k: PerfilPartner['especialidades'][number]) {
		perfil.especialidades = perfil.especialidades.includes(k)
			? perfil.especialidades.filter((e) => e !== k)
			: [...perfil.especialidades, k];
	}
	async function enviar(e: SubmitEvent) {
		e.preventDefault();
		const v = validarPerfil(perfil);
		if (!v.ok) {
			error = v.mensaje;
			return;
		}
		const err = await guardar({ que: 'perfil', perfil: v.valor });
		if (err) {
			error = err;
			return;
		}
		error = null;
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Perfil guardado', desc: 'Así te ven en Buscar un partner' });
	}
</script>

{#if n.id === 'partner'}
	<Alerta
		tipo="info"
		titulo="Todavía no aparecés en el directorio"
		texto="Desde Partner Plata (5 tiendas activas; hoy tenés {activas}) tu perfil se ve en Buscar un partner. Dejalo listo para ese{NB}día."
	/>
{/if}

<div class="dos" use:anim={{ t: 'rise' }}>
	<form class="tarjeta campos" onsubmit={enviar}>
		<div class="t-cab ancho"><span>Tu perfil</span></div>
		<label>Nombre o agencia<input bind:value={perfil.nombre} maxlength="80" /></label>
		<label>Ciudad<input bind:value={perfil.ciudad} maxlength="60" placeholder="Rosario" /></label>
		<label class="ancho"
			>Qué hacés<textarea
				bind:value={perfil.descripcion}
				maxlength="280"
				rows="3"
				placeholder="Armamos tiendas para marcas de moda: fotos, catálogo y Mercado Libre."
			></textarea></label
		>
		<div class="campo-t ancho">
			Especialidades
			<div class="chips" role="group" aria-label="Especialidades">
				{#each ESPECIALIDADES as [k, t] (k)}
					<button
						type="button"
						aria-pressed={perfil.especialidades.includes(k)}
						onclick={() => alternar(k)}>{t}</button
					>
				{/each}
			</div>
		</div>
		<div class="campo-t ancho">
			Casos (tiendas que armaste)
			{#each perfil.casos as c, i (i)}
				<div class="caso">
					<input
						bind:value={c.nombre}
						maxlength="60"
						placeholder="Café Bruma"
						aria-label="Tienda {i + 1}"
					/>
					<input
						bind:value={c.url}
						maxlength="200"
						placeholder="https://…"
						inputmode="url"
						aria-label="Link de la tienda {i + 1}"
					/>
					<button
						type="button"
						class="quitar"
						aria-label="Quitar el caso {i + 1}"
						onclick={() => (perfil.casos = perfil.casos.filter((_, j) => j !== i))}
						><Icon icon={I.phTrashFill} size={15} /></button
					>
				</div>
			{/each}
			{#if perfil.casos.length < 6}
				<button
					type="button"
					class="mas"
					onclick={() => (perfil.casos = [...perfil.casos, { nombre: '', url: '' }])}
					><Icon icon={I.phPlusBold} size={13} />Sumar un caso</button
				>
			{/if}
		</div>
		<label>Email<input bind:value={perfil.email} type="email" maxlength="254" /></label>
		<label
			>WhatsApp<input
				bind:value={perfil.whatsapp}
				inputmode="tel"
				maxlength="20"
				placeholder="341 555 0123"
			/></label
		>
		<label
			>Web<input
				bind:value={perfil.web}
				inputmode="url"
				maxlength="200"
				placeholder="https://…"
			/></label
		>
		<label
			>Logo (link a la imagen)<input
				bind:value={perfil.logo}
				inputmode="url"
				maxlength="300"
				placeholder="https://…/logo.png"
			/></label
		>
		{#if error}<div class="error ancho" role="alert">{error}</div>{/if}
		{#if cambiado}
			<div class="guardar">
				<button
					type="button"
					class="btn btn-secondary"
					onclick={() => (perfil = copia(p.partner.perfil))}>Descartar</button
				>
				<button class="btn btn-primary">Guardar perfil</button>
			</div>
		{/if}
	</form>

	<div class="previa">
		<div class="p-tit">Así te ven en Buscar un partner</div>
		<div class="partner" data-testid="vista-previa">
			<div class="foto">
				{#if perfil.logo.startsWith('https://')}
					<img src={perfil.logo} alt="" />
				{:else}
					<span class="sin-logo">{iniciales(perfil.nombre || 'P')}</span>
				{/if}
			</div>
			<div class="p-in">
				<div class="p-cab">
					<b>{perfil.nombre || 'Tu nombre'}</b><span class="nivel-chip" class:oro={n.id === 'oro'}
						>{n.nombre}</span
					>
				</div>
				{#if linea}<div class="p-d">{linea}</div>{/if}
				{#if perfil.descripcion}<div class="p-txt">{perfil.descripcion}</div>{/if}
				<div class="p-val">
					<Icon icon={I.phStorefrontFill} size={14} style="color:#D2582F" /><b>{activas}</b><span
						>tiendas activas con Mostrador</span
					>
				</div>
				{#if perfil.casos.some((c) => c.nombre)}
					<div class="p-casos">
						{#each perfil.casos.filter((c) => c.nombre) as c, i (i)}
							<span>{c.nombre}</span>
						{/each}
					</div>
				{/if}
				<div class="p-contacto">
					{#if perfil.whatsapp}<span><Icon icon={I.phWhatsappLogoFill} size={14} />WhatsApp</span
						>{/if}
					{#if perfil.email}<span><Icon icon={I.phEnvelopeSimpleFill} size={14} />Email</span>{/if}
					{#if perfil.web}<span><Icon icon={I.phGlobeFill} size={14} />Web</span>{/if}
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.dos {
		display: grid;
		grid-template-columns: minmax(0, 1.3fr) minmax(0, 1fr);
		gap: 12px;
		align-items: start;
	}
	.caso {
		display: grid;
		grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.2fr) 40px;
		gap: 6px;
		align-items: center;
	}
	.caso input {
		padding: 11px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font:
			500 14.5px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
		min-width: 0;
	}
	.quitar {
		width: 40px;
		height: 40px;
		border-radius: 12px;
		border: 0;
		background: var(--chip);
		color: #b83a2e;
		display: grid;
		place-items: center;
	}
	.mas {
		margin-top: 2px;
		align-self: flex-start;
	}
	.error {
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	.previa {
		position: sticky;
		top: 80px;
	}
	.p-tit {
		margin: 4px 0 10px;
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	/* La tarjeta del directorio de /partners, tal cual. */
	.partner {
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		overflow: hidden;
		color: #2a140c;
	}
	.foto {
		aspect-ratio: 16 / 9;
		background: #e9ddcf;
		display: grid;
		place-items: center;
	}
	.foto img {
		width: 100%;
		height: 100%;
		object-fit: contain;
		display: block;
	}
	.sin-logo {
		font-size: 44px;
		font-weight: 600;
		letter-spacing: -0.03em;
		color: #c24e25;
	}
	.p-in {
		padding: 18px;
	}
	.p-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 10px;
	}
	.p-cab b {
		font-size: 17px;
		min-width: 0;
		overflow-wrap: anywhere;
	}
	.nivel-chip {
		padding: 4px 10px;
		border-radius: 999px;
		background: #efeae4;
		color: #5c4536;
		font-size: 12px;
		font-weight: 700;
		white-space: nowrap;
	}
	.nivel-chip.oro {
		background: #fcede5;
		color: #c24e25;
	}
	.p-d {
		font-size: 14px;
		color: #7a6252;
		margin-top: 4px;
	}
	.p-txt {
		margin-top: 10px;
		font-size: 14px;
		line-height: 1.45;
		text-wrap: pretty;
	}
	.p-val {
		margin-top: 12px;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13.5px;
	}
	.p-val span {
		color: #7a6252;
	}
	.p-casos {
		margin-top: 12px;
		display: flex;
		gap: 4px;
		flex-wrap: wrap;
	}
	.p-casos span {
		font-size: 11.5px;
		font-weight: 600;
		padding: 3px 8px;
		border-radius: 999px;
		background: rgba(58, 29, 18, 0.055);
	}
	.p-contacto {
		margin-top: 14px;
		display: flex;
		gap: 14px;
		flex-wrap: wrap;
		font-size: 13px;
		font-weight: 600;
		color: #c24e25;
	}
	.p-contacto span {
		display: inline-flex;
		align-items: center;
		gap: 5px;
	}
	@media (max-width: 1099px) {
		.dos {
			grid-template-columns: minmax(0, 1fr);
		}
		.previa {
			position: static;
		}
	}
</style>
