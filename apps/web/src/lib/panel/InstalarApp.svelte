<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import {
		cerrarInvitacion,
		esAndroid,
		esIOS,
		instalacion,
		instalar
	} from '#lib/panel/app-celular.svelte.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * «Instalá Tu Local Web en tu celular» (docs/decisiones.md §43), en el Inicio. Sólo en el celular
	 * y sólo si se puede: en el iPhone, los pasos de Safari; en Android, el botón que abre el diálogo
	 * del sistema. Con la app ya instalada (abierta desde la pantalla de inicio) no aparece, y «Ahora
	 * no» la cierra para siempre en ese navegador.
	 */
	const ui = getUi();
	let ios = $state(false);
	let android = $state(false);
	$effect(() => {
		ios = esIOS();
		android = esAndroid();
	});
	const ver = $derived(
		instalacion.listo &&
			!instalacion.instalada &&
			!instalacion.cerrada &&
			(ios || (android && instalacion.evento !== null))
	);

	async function instalarAndroid() {
		if (await instalar())
			ui.toast({
				k: 'ok',
				title: 'Listo, ya la tenés',
				desc: 'Abrila desde el ícono de Tu Local Web'
			});
	}
</script>

{#if ver}
	<section class="instalar" aria-labelledby="instalar-t" use:anim={{ t: 'rise' }}>
		<img class="i-ico" src="/panel/apple-touch-icon.png" alt="" width="46" height="46" />
		<div class="i-txt">
			<h2 id="instalar-t">Instalá Tu&nbsp;Local&nbsp;Web en tu&nbsp;celular</h2>
			{#if ios}
				<p class="i-d">Se abre como una app, sin la barra de Safari, y te avisa cada&nbsp;venta.</p>
				<ol class="pasos">
					<li>
						<span class="n">1</span><span
							>Tocá <b>Compartir</b>
							<span class="ic" aria-hidden="true"><Icon icon={I.phExport} size={17} /></span> en Safari
							(si no lo ves, está en el menú&nbsp;«…»).</span
						>
					</li>
					<li>
						<span class="n">2</span><span
							>Elegí <b>Agregar a inicio</b>
							<span class="ic" aria-hidden="true"><Icon icon={I.phPlusSquare} size={17} /></span> y después&nbsp;«Agregar».</span
						>
					</li>
				</ol>
			{:else}
				<p class="i-d">
					Se abre como una app, desde la pantalla de inicio, y te avisa cada&nbsp;venta.
				</p>
				<button type="button" class="btn btn-primary btn-sm i-btn" onclick={instalarAndroid}
					><Icon icon={I.phDownloadSimpleBold} size={15} />Instalar</button
				>
			{/if}
		</div>
		<button
			type="button"
			class="i-x"
			aria-label="Ahora no: no mostrar más"
			title="Ahora no"
			onclick={cerrarInvitacion}><Icon icon={I.phXBold} size={14} /></button
		>
	</section>
{/if}

<style>
	.instalar {
		position: relative;
		margin-top: 20px;
		display: flex;
		align-items: flex-start;
		gap: 14px;
		padding: 14px 48px 14px 14px;
		border-radius: 18px;
		background: linear-gradient(90deg, rgba(210, 88, 47, 0.12), rgba(210, 88, 47, 0.04));
		border: 1px solid rgba(210, 88, 47, 0.28);
	}
	.i-ico {
		flex: none;
		width: 46px;
		height: 46px;
		border-radius: 11px;
		box-shadow: 0 1px 3px rgba(58, 29, 18, 0.18);
	}
	.i-txt {
		flex: 1;
		min-width: 0;
	}
	h2 {
		margin: 0;
		font-size: 15px;
		font-weight: 600;
		letter-spacing: -0.01em;
		text-wrap: balance;
	}
	.i-d {
		margin: 2px 0 0;
		font-size: 13px;
		line-height: 1.45;
		color: var(--mu);
		text-wrap: pretty;
	}
	.pasos {
		margin: 10px 0 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 13.5px;
		line-height: 1.45;
	}
	.pasos li {
		display: flex;
		gap: 9px;
		align-items: flex-start;
	}
	.n {
		flex: none;
		width: 21px;
		height: 21px;
		margin-top: 0.5px;
		border-radius: 999px;
		background: var(--terracota);
		color: #fff;
		font-size: 11.5px;
		font-weight: 700;
		display: grid;
		place-items: center;
	}
	.pasos b {
		font-weight: 600;
	}
	.ic {
		display: inline-grid;
		place-items: center;
		width: 24px;
		height: 24px;
		margin: -3px 1px;
		border-radius: 7px;
		background: var(--sf);
		color: #0a84ff;
		box-shadow: 0 0 0 1px var(--ln);
		vertical-align: middle;
	}
	.i-btn {
		margin-top: 10px;
	}
	.i-x {
		position: absolute;
		top: 4px;
		right: 4px;
		width: 44px;
		height: 44px;
		border: 0;
		border-radius: 999px;
		background: transparent;
		color: var(--mu);
		display: grid;
		place-items: center;
		cursor: pointer;
	}
	.i-x:hover {
		color: var(--tx);
	}
</style>
