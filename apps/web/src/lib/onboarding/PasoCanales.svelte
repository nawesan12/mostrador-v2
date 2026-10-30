<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { CANALES_REALES, CONEXIONES } from './datos.ts';
	import Etiqueta from './Etiqueta.svelte';
	import { getOnboarding } from './estado.svelte.ts';
	import Logo from './Logo.svelte';
	import { page } from '$app/state';

	/**
	 * Paso 3, Canales: conectar lo que ya se usa. En la demo, simulado. Con la API, Mercado Pago y
	 * Mercado Libre se conectan acá (el OAuth vuelve a este mismo paso); WhatsApp e Instagram,
	 * pronto.
	 */
	let { conectadas = { mp: false, ml: false } }: { conectadas?: { mp: boolean; ml: boolean } } =
		$props();
	const o = getOnboarding();
	const RUTA = { mp: '/panel/integraciones/mercadopago', ml: '/panel/integraciones/mercadolibre' };
	let yendo = $state<'mp' | 'ml' | null>(null);
	let aviso = $state('');
	// La vuelta del proveedor: ?mercadopago=conectada|error (o mercadolibre).
	const vuelta = $derived(
		page.url.searchParams.get('mercadopago') === 'error' ||
			page.url.searchParams.get('mercadolibre') === 'error'
			? 'No se pudo conectar. Probá de nuevo.'
			: ''
	);
	async function conectar(k: 'mp' | 'ml') {
		yendo = k;
		aviso = '';
		const r = await fetch(`${RUTA[k]}?volver=${encodeURIComponent('/onboarding?paso=3')}`, {
			method: 'POST'
		}).catch(() => null);
		if (!r?.ok) {
			yendo = null;
			aviso =
				((await r?.json().catch(() => null)) as { message?: string } | null)?.message ??
				'No se pudo conectar. Probá de nuevo.';
			return;
		}
		location.href = ((await r.json()) as { url: string }).url;
	}
</script>

{#if o.demo}
	<div class="cab">
		<div class="cab-t">
			<Etiqueta rotulo="Canales" titulo="Conectá lo que ya usás. Traemos todo de una." />
		</div>
		<button type="button" class="todas" onclick={() => o.conectarRecomendadas()}
			><Icon icon={I.phLightningFill} size={15} style="color:#F5B796" />Conectar las recomendadas</button
		>
	</div>
	<div class="grilla">
		{#each CONEXIONES as [k, n, d, trae, rec] (k)}
			{@const c = o.e.conexiones[k] ?? 'off'}
			<div class="con" class:on={c === 'on'}>
				<Logo {k} />
				<div class="c-txt">
					<div class="c-n">
						<span>{n}</span>{#if rec}<span class="reco">Recomendado</span>{/if}
					</div>
					<div class="c-d" class:trae={c === 'on'}>{c === 'on' ? trae : d}</div>
				</div>
				{#if c === 'off'}
					<button type="button" class="conectar" onclick={() => o.conectar(k)}>Conectar</button>
				{:else if c === 'wait'}
					<span class="espera"
						><Icon icon={I.phCircleNotchBold} size={15} class="gira" />Conectando</span
					>
				{:else}
					<span class="listo" aria-label="Conectado"><Icon icon={I.phCheckBold} size={14} /></span>
				{/if}
			</div>
		{/each}
	</div>
	<div class="resumen">
		{o.conectadas} de {CONEXIONES.length} conectadas. Podés sumar más después desde Integraciones.
	</div>
{:else}
	<Etiqueta rotulo="Canales" titulo="Conectá lo que ya usás." />
	<div class="reales">
		{#each CANALES_REALES as c (c.k)}
			{@const real = c.k === 'mp' || c.k === 'ml'}
			{@const on = c.k === 'mp' ? conectadas.mp : c.k === 'ml' ? conectadas.ml : false}
			<div class="real" class:on>
				<Logo k={c.k} tam={36} icono={18} radio={11} />
				<span class="r-txt"><span class="r-t">{c.t}</span><span class="r-d">{c.d}</span></span>
				{#if on}
					<span class="r-ok"><Icon icon={I.phCheckBold} size={14} />Conectado</span>
				{:else if real}
					<button
						type="button"
						class="r-con"
						disabled={yendo !== null}
						onclick={() => conectar(c.k as 'mp' | 'ml')}
						>{yendo === c.k ? 'Abriendo…' : 'Conectar'}</button
					>
				{:else}
					<span class="r-pronto">Pronto</span>
				{/if}
			</div>
		{/each}
	</div>
	{#if aviso || vuelta}<p class="r-aviso" role="alert">{aviso || vuelta}</p>{/if}
	<p class="r-nota">Lo que no conectes ahora lo conectás después desde Integraciones.</p>
{/if}

<style>
	.cab {
		display: flex;
		align-items: end;
		gap: 20px;
		flex-wrap: wrap;
	}
	.cab-t {
		flex: 1 1 400px;
	}
	.todas {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 18px;
		border-radius: 999px;
		border: 0;
		background: #2a140c;
		color: #fff;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
	}
	.grilla {
		margin-top: 24px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr));
		gap: 12px;
	}
	.con {
		display: flex;
		gap: 14px;
		padding: 16px;
		border-radius: 20px;
		background: #fff;
		border: 1.5px solid rgba(58, 29, 18, 0.08);
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.05);
		transition: border-color 0.3s;
	}
	.con.on {
		border-color: rgba(47, 143, 99, 0.45);
	}
	.c-txt {
		flex: 1;
		min-width: 0;
	}
	.c-n {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 15px;
		font-weight: 600;
		flex-wrap: wrap;
	}
	.reco {
		font-size: 11px;
		font-weight: 700;
		color: #b0441d;
		background: rgba(210, 88, 47, 0.12);
		padding: 2px 7px;
		border-radius: 999px;
	}
	.c-d {
		font-size: 13px;
		color: #7a6252;
		margin-top: 3px;
		line-height: 1.4;
	}
	.c-d.trae {
		color: #237550;
		font-weight: 500;
	}
	.conectar {
		align-self: center;
		padding: 8px 14px;
		border-radius: 999px;
		border: 1px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		color: #2a140c;
		font-size: 13px;
		font-weight: 600;
	}
	.espera {
		align-self: center;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		color: #b0441d;
		white-space: nowrap;
	}
	.espera :global(.gira) {
		animation: gira 0.9s linear infinite;
	}
	@keyframes gira {
		to {
			transform: rotate(360deg);
		}
	}
	.listo {
		align-self: center;
		width: 30px;
		height: 30px;
		flex: none;
		border-radius: 50%;
		background: #2f8f63;
		color: #fff;
		display: grid;
		place-items: center;
	}
	.resumen {
		margin-top: 14px;
		font-size: 13px;
		color: #7a6252;
	}
	.reales {
		margin-top: 22px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr));
		gap: 10px;
	}
	.real {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 16px;
		border-radius: 18px;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.1);
		color: #2a140c;
	}
	.real.on {
		border-color: rgba(47, 143, 99, 0.45);
	}
	.r-txt {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.r-t {
		font-size: 15px;
		font-weight: 600;
	}
	.r-d {
		font-size: 13.5px;
		color: #6b5244;
		line-height: 1.35;
	}
	.r-con {
		flex: none;
		padding: 10px 16px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 14px;
		font-weight: 600;
	}
	.r-con:disabled {
		opacity: 0.6;
	}
	.r-ok {
		flex: none;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 14px;
		font-weight: 600;
		color: #2f8f63;
	}
	.r-pronto {
		flex: none;
		padding: 6px 12px;
		border-radius: 999px;
		background: rgba(58, 29, 18, 0.06);
		color: #6b5244;
		font-size: 13px;
		font-weight: 600;
	}
	.r-aviso {
		margin: 12px 0 0;
		font-size: 14px;
		color: #b83a2e;
	}
	.r-nota {
		margin: 14px 0 0;
		font-size: 14px;
		color: #6b5244;
	}
</style>
