<script lang="ts">
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import Grano from '#lib/components/Grano.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { marca } from '#lib/marcas.ts';

	/**
	 * La página de un link de pago. El handoff no la diseña: se arma con las piezas del modal "Crear
	 * link de pago" (monto grande, cuotas, medios que acepta) y del checkout de las plantillas.
	 */
	let { data, form } = $props();
	const l = $derived(data.link);
	const pagado = $derived(l.pagado || !!form?.pagado);
	// La vuelta de Mercado Pago (?pago=ok|pendiente|error). El "ok" se confirma con el aviso de
	// Mercado Pago: hasta que llega, se muestra como en camino.
	const vuelta = $derived(page.url.searchParams.get('pago'));
	let pagando = $state(false);
	const pesos = (n: number) => '$ ' + n.toLocaleString('es-AR');
	const MEDIOS = ['mp', 'visa', 'master', 'modo', 'uala'];
</script>

<Seo
	title="Pagar a {l.tienda}"
	description="{l.concepto} · {pesos(l.monto)}"
	path="/p/{l.codigo}"
	noindex
/>

<main class="fondo">
	<Grano />
	<div class="tarjeta">
		<div class="tienda">
			<span class="t-ico">{l.tienda.slice(0, 1)}</span>
			<span class="t-txt"
				><span class="t-n">{l.tienda}</span><span class="t-s">Link de pago</span></span
			>
		</div>

		{#if pagado}
			<div class="listo">
				<span class="l-ico"><Icon icon={I.phCheckBold} size={28} /></span>
				<div class="l-t">¡Listo, pagaste!</div>
				<div class="l-d">
					{pesos(l.monto)} a {l.tienda}{l.pedido ? ` · pedido #${l.pedido}` : ''}. Te llega el
					comprobante por mail.
				</div>
			</div>
		{:else}
			<div class="monto">
				<div class="m-t">{l.concepto}</div>
				<div class="m-v">{pesos(l.monto)}</div>
				{#if l.cuotas}<div class="m-c">
						Hasta 6 cuotas sin interés de {pesos(Math.ceil(l.monto / 6))}
					</div>{/if}
			</div>
			<div class="acepta">
				{#each MEDIOS as k (k)}
					{@const m = marca(k)}
					<span class="medio" style:background={m.bg}
						>{#if m.icon}<Icon
								icon={m.icon}
								size={18}
								style="color:{m.col ?? 'inherit'}"
							/>{:else}<span class="sigla" style:color={m.col}>{m.sigla}</span>{/if}</span
					>
				{/each}
			</div>
			{#if vuelta === 'ok' || vuelta === 'pendiente'}
				<div class="aviso" role="status">
					{vuelta === 'ok'
						? 'Estamos confirmando tu pago con Mercado Pago. Actualizá en un ratito.'
						: 'Tu pago quedó pendiente. Te avisamos por mail cuando se acredite.'}
				</div>
			{:else if vuelta === 'error'}
				<div class="error" role="alert">
					Mercado Pago no aprobó el pago. Probá con otro medio o con otra&nbsp;tarjeta.
				</div>
			{/if}
			{#if form?.error}<div class="error" role="alert">{form.error}</div>{/if}
			<form
				method="POST"
				use:enhance={() => {
					pagando = true;
					return async ({ result, update }) => {
						// El checkout de Mercado Pago es otro sitio: se navega de verdad, no con goto.
						if (result.type === 'redirect') return void (location.href = result.location);
						await update({ reset: false });
						pagando = false;
					};
				}}
			>
				<button
					class="pagar"
					disabled={pagando}
					data-mp-checkout-cta="checkout-pro"
					data-mp-marketplace-checkout="checkout-pro"
					><Icon icon={I.phLockSimpleFill} size={16} />Pagar {pesos(l.monto)}</button
				>
			</form>
			<div class="nota">
				{data.demo
					? 'Es una demo: el pago se simula y no se cobra nada.'
					: 'Pagás con Mercado Pago. Tus datos no pasan por la tienda.'}
			</div>
		{/if}
	</div>
	<a class="marca" href={resolve('/')}>Cobrá así con <b>Mostrador</b></a>
</main>

<style>
	.fondo {
		position: relative;
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 18px;
		padding: 32px 16px;
		background: var(--bg);
		color: var(--tx);
	}
	.tarjeta {
		position: relative;
		width: 100%;
		max-width: 440px;
		padding: 24px;
		border-radius: 26px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.tienda {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.t-ico {
		width: 42px;
		height: 42px;
		border-radius: 13px;
		display: grid;
		place-items: center;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-weight: 700;
		font-size: 18px;
	}
	.t-n {
		display: block;
		font-size: 16px;
		font-weight: 600;
	}
	.t-s {
		display: block;
		font-size: 13px;
		color: var(--mu);
	}
	.monto {
		padding: 18px;
		border-radius: 18px;
		background: var(--sf3);
	}
	.m-t {
		font-size: 14px;
		font-weight: 600;
		color: var(--mu);
	}
	.m-v {
		margin-top: 4px;
		font-size: 44px;
		font-weight: 600;
		letter-spacing: -0.045em;
	}
	.m-c {
		margin-top: 2px;
		font-size: 13.5px;
		color: #237550;
		font-weight: 600;
	}
	.acepta {
		display: flex;
		gap: 8px;
		justify-content: center;
	}
	.medio {
		width: 40px;
		height: 30px;
		border-radius: 8px;
		display: grid;
		place-items: center;
		box-shadow: var(--sh);
	}
	.sigla {
		font-size: 8px;
		font-weight: 800;
	}
	.pagar {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 15px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 15.5px 'Instrument Sans',
			sans-serif;
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.3);
	}
	.pagar:disabled {
		opacity: 0.6;
	}
	.nota {
		font-size: 12.5px;
		color: var(--mu);
		text-align: center;
	}
	.error {
		padding: 10px 14px;
		border-radius: 14px;
		background: rgba(184, 58, 46, 0.1);
		color: #b83a2e;
		font-size: 13.5px;
	}
	.listo {
		text-align: center;
		padding: 8px 0 4px;
	}
	.l-ico {
		width: 60px;
		height: 60px;
		margin: 0 auto;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: rgba(47, 143, 99, 0.14);
		color: #237550;
	}
	.l-t {
		margin-top: 14px;
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.l-d {
		margin-top: 6px;
		font-size: 14.5px;
		color: var(--mu);
		line-height: 1.5;
		text-wrap: pretty;
	}
	.marca {
		position: relative;
		font-size: 13px;
		color: var(--mu);
		text-decoration: none;
	}
	.aviso {
		margin-top: 14px;
		padding: 12px 14px;
		border-radius: 14px;
		background: rgba(62, 107, 253, 0.08);
		color: var(--tx);
		font-size: 13.5px;
		line-height: 1.5;
	}
</style>
