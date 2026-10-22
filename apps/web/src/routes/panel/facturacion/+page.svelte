<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import * as I from '#lib/icons.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import MotorLista from '#lib/panel/listas/MotorLista.svelte';
	import { configFacturacion } from '#lib/panel/listas/facturacion.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Facturación (PanelListas) con la factura manual (sin diseño en el handoff: piezas del modal). */
	let { data } = $props();
	const ui = getUi();
	let emitir = $state(false);
	const config = configFacturacion(() => (emitir = true));

	let tipo = $state<'B' | 'A'>('B');
	let cliente = $state('Consumidor final');
	let cuit = $state('');
	let concepto = $state('');
	let total = $state('');
	let error = $state<string | null>(null);
	const monto = $derived(Number(total.replace(/\D/g, '')) || 0);

	async function enviar(e: SubmitEvent) {
		e.preventDefault();
		error = null;
		if (cliente.trim().length < 2) return (error = 'Poné a quién le facturás.');
		if (tipo === 'A' && cuit.replace(/\D/g, '').length !== 11)
			return (error = 'La factura A necesita el CUIT (11 números).');
		if (!monto) return (error = 'Poné el monto total.');
		const r = await fetch('/panel/listas/facturacion', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				ids: [],
				accion: 'emitir',
				valor: { cliente, tipo, cuit, concepto, total: monto }
			})
		});
		const n = r.ok ? ((await r.json()) as { hechos: number }).hechos : 0;
		if (!n) return (error = 'No pudimos emitirla. Revisá los datos.');
		await invalidateAll();
		emitir = false;
		ui.toast({
			k: 'arca',
			title: `Factura ${tipo} emitida`,
			desc: `${cliente.trim()} · ${pesos(monto)}`
		});
		cliente = 'Consumidor final';
		cuit = concepto = total = '';
		tipo = 'B';
	}
</script>

<Seo
	title="Facturación · Panel"
	description="Facturas de tu tienda"
	path="/panel/facturacion"
	noindex
/>

<MotorLista {config} lista={data.lista} />

<Modal bind:open={emitir} titulo="Emitir factura" ancho={480}>
	<form onsubmit={enviar}>
		<div class="m-ico"><Icon icon={I.phReceiptFill} size={24} /></div>
		<div class="m-t">Emitir factura</div>
		<div class="m-d">
			Para una venta que no pasó por la tienda: en el local, por teléfono o un servicio.
		</div>
		<div class="m-campos">
			<Segmentado
				opciones={[
					['B', 'Factura B'],
					['A', 'Factura A']
				]}
				bind:value={tipo}
				label="Tipo de factura"
			/>
			<label class="field"
				>{tipo === 'A' ? 'Razón social' : 'Cliente'}<input
					class="input"
					bind:value={cliente}
					maxlength="80"
				/></label
			>
			{#if tipo === 'A'}
				<label class="field"
					>CUIT<input
						class="input"
						bind:value={cuit}
						inputmode="numeric"
						placeholder="30-71234567-8"
						maxlength="13"
					/></label
				>
			{/if}
			<label class="field"
				>Concepto<input
					class="input"
					bind:value={concepto}
					maxlength="200"
					placeholder="Por ejemplo: 3 mates de calabaza"
				/></label
			>
			<label class="field"
				>Total<input
					class="input"
					bind:value={total}
					inputmode="numeric"
					placeholder="$ 42.900"
				/></label
			>
			{#if monto && tipo === 'A'}<div class="iva">
					IVA incluido: {pesos((monto * 0.21) / 1.21)}
				</div>{/if}
		</div>
		{#if error}<div class="m-error" role="alert">{error}</div>{/if}
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (emitir = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary">Emitir</button>
		</div>
	</form>
</Modal>

<style>
	.m-ico {
		width: 48px;
		height: 48px;
		display: grid;
		place-items: center;
		border-radius: 15px;
		background: rgba(31, 58, 110, 0.12);
		color: #1f3a6e;
	}
	.m-t {
		margin-top: 16px;
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14.5px;
		line-height: 1.5;
		color: var(--mu);
	}
	.m-campos {
		margin-top: 16px;
		display: grid;
		gap: 12px;
	}
	.iva {
		font-size: 13px;
		color: var(--mu);
	}
	.m-error {
		margin-top: 10px;
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	.m-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
</style>
