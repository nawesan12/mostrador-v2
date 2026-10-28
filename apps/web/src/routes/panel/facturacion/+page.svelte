<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import * as I from '#lib/icons.ts';
	import { cuitValido, limpiarCuit } from '#lib/panel/cuit.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import MotorLista from '#lib/panel/listas/MotorLista.svelte';
	import {
		configFacturacion,
		esNota,
		estadoDe,
		nombreComprobante,
		numeroDe,
		type Comprobante
	} from '#lib/panel/listas/facturacion.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Facturación (PanelListas) con la factura manual y el detalle de un comprobante con "Anular con
	 * nota de crédito" (sin diseño en el handoff: piezas del modal).
	 */
	let { data } = $props();
	const ui = getUi();
	let emitir = $state(false);
	const config = configFacturacion(() => (emitir = true));
	const inscripta = $derived(data.arca.condicion === 'responsable_inscripto');

	let tipo = $state<'B' | 'A'>('B');
	let cliente = $state('Consumidor final');
	let cuit = $state('');
	let concepto = $state('');
	let total = $state('');
	let error = $state<string | null>(null);
	let enviando = $state(false);
	const monto = $derived(Number(total.replace(/\D/g, '')) || 0);
	// La letra la decide la condición de la tienda: monotributo siempre emite C.
	const letra = $derived(inscripta ? tipo : 'C');
	const faltaArca = $derived(!data.arca.conectada || /conect[áa] ARCA/i.test(error ?? ''));

	const mensaje = async (r: Response) =>
		((await r.json().catch(() => null)) as { message?: string } | null)?.message;

	async function enviar(e: SubmitEvent) {
		e.preventDefault();
		error = null;
		if (cliente.trim().length < 2) return (error = 'Poné a quién le facturás.');
		if (letra === 'A' && !cuitValido(limpiarCuit(cuit)))
			return (error = 'La factura A necesita el CUIT del cliente: revisá que esté bien.');
		if (!monto) return (error = 'Poné el monto total.');
		enviando = true;
		let r: Response;
		try {
			r = await fetch('/panel/listas/facturacion', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({
					ids: [],
					accion: 'emitir',
					valor: { cliente, tipo: letra, cuit: limpiarCuit(cuit), concepto, total: monto }
				})
			});
		} catch {
			return (error = 'No pudimos conectarnos. Revisá la conexión y probá de nuevo.');
		} finally {
			enviando = false;
		}
		// La API explica qué pasó (sin ARCA, rechazo de ARCA, datos): se muestra tal cual.
		if (!r.ok) return (error = (await mensaje(r)) ?? 'No pudimos emitirla. Probá de nuevo.');
		const n = ((await r.json()) as { hechos: number }).hechos;
		if (!n) return (error = 'No pudimos emitirla. Revisá los datos.');
		await invalidateAll();
		emitir = false;
		ui.toast({
			k: 'arca',
			title: `Factura ${letra} emitida`,
			desc: `${cliente.trim()} · ${pesos(monto)}`
		});
		cliente = 'Consumidor final';
		cuit = concepto = total = '';
		tipo = 'B';
	}

	// Detalle de un comprobante y anulación con nota de crédito.
	let detalle = $state<Comprobante | null>(null);
	let verDetalle = $state(false);
	let confirmar = $state(false);
	let anulando = $state(false);
	function abrir(c: Comprobante) {
		detalle = c;
		confirmar = false;
		verDetalle = true;
	}
	async function anular() {
		if (!detalle) return;
		const c = detalle;
		anulando = true;
		let r: Response;
		try {
			r = await fetch('/panel/listas/facturacion', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ ids: [c.id], accion: 'anular' })
			});
		} catch {
			return ui.toast({
				k: 'err',
				title: 'No se pudo anular',
				desc: 'Revisá la conexión y probá de nuevo.'
			});
		} finally {
			anulando = false;
		}
		const n = r.ok ? ((await r.json()) as { hechos: number }).hechos : 0;
		verDetalle = false;
		if (!n)
			return ui.toast({
				k: 'err',
				title: 'No se anuló',
				desc: (!r.ok && (await mensaje(r))) || 'Esa factura ya no se puede anular.'
			});
		await invalidateAll();
		ui.toast({
			k: 'arca',
			title: `${esNota(c) ? 'Nota' : 'Factura'} ${c.tipo} anulada`,
			desc: `Nota de crédito por ${pesos(c.total)} para ${c.cliente}`
		});
	}
</script>

<Seo
	title="Facturación · Panel"
	description="Facturas de tu tienda"
	path="/panel/facturacion"
	noindex
/>

<MotorLista {config} lista={data.lista} {abrir} />

<Modal bind:open={emitir} titulo="Emitir factura" ancho={480}>
	<form onsubmit={enviar}>
		<div class="m-ico"><Icon icon={I.phReceiptFill} size={24} /></div>
		<div class="m-t">Emitir factura</div>
		<div class="m-d">
			Para una venta que no pasó por la tienda: en el local, por teléfono o un servicio.
		</div>
		<div class="m-campos">
			{#if inscripta}
				<Segmentado
					opciones={[
						['B', 'Factura B'],
						['A', 'Factura A']
					]}
					bind:value={tipo}
					label="Tipo de factura"
				/>
			{:else if data.arca.condicion}
				<div class="iva">Sos monotributista: emitís factura C, sin IVA&nbsp;discriminado.</div>
			{/if}
			<label class="field"
				>{letra === 'A' ? 'Razón social' : 'Cliente'}<input
					class="input"
					bind:value={cliente}
					maxlength="80"
				/></label
			>
			{#if letra === 'A'}
				<label class="field"
					>CUIT<input
						class="input"
						bind:value={cuit}
						inputmode="numeric"
						placeholder="30-71234567-1"
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
			{#if monto && letra === 'A'}<div class="iva">
					IVA incluido: {pesos((monto * 0.21) / 1.21)}
				</div>{/if}
		</div>
		{#if error}<div class="m-error" role="alert">{error}</div>{/if}
		{#if faltaArca}
			<a class="m-link" href="{resolve('/panel/integraciones')}?arca"
				><Icon icon={I.phPlugsConnectedBold} size={16} />Ir a Integraciones › ARCA</a
			>
		{/if}
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (emitir = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary" disabled={enviando}
				>{enviando ? 'Emitiendo…' : 'Emitir'}</button
			>
		</div>
	</form>
</Modal>

<Modal
	bind:open={verDetalle}
	titulo={detalle ? nombreComprobante(detalle) : 'Comprobante'}
	ancho={480}
>
	{#if detalle}
		{@const c = detalle}
		{@const e = estadoDe(c)}
		<div class="m-ico"><Icon icon={I.phReceiptFill} size={24} /></div>
		{#if !confirmar}
			<div class="m-t">{nombreComprobante(c)}</div>
			<div class="m-d">
				{c.cliente} · {pesos(c.total)}{c.pedido ? ` · pedido #${c.pedido}` : ''}
			</div>
			<div class="estado" style:color={e.col}><Icon icon={e.icon} size={16} />{e.t}</div>
			<div class="m-acc envolver">
				<button type="button" class="btn btn-secondary" onclick={() => (verDetalle = false)}
					>Cerrar</button
				>
				{#if c.estado === 'ok' && !esNota(c)}
					<button type="button" class="btn btn-danger" onclick={() => (confirmar = true)}
						><Icon icon={I.phProhibitBold} size={16} />Anular con nota de crédito</button
					>
				{/if}
			</div>
		{:else}
			<div class="m-t">
				¿Anular la factura {c.tipo} <span class="nw">{numeroDe(c)}?</span>
			</div>
			<div class="m-d">
				Se emite una nota de crédito {c.tipo} por {pesos(c.total)} que la cancela entera ante ARCA. No
				se puede deshacer: si hace falta, después hacés una factura&nbsp;nueva.
			</div>
			<div class="m-acc envolver">
				<button type="button" class="btn btn-secondary" onclick={() => (confirmar = false)}
					>Volver</button
				>
				<button type="button" class="btn btn-danger" disabled={anulando} onclick={anular}
					>{anulando ? 'Anulando…' : 'Anular'}</button
				>
			</div>
		{/if}
	{/if}
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
	.m-acc.envolver {
		flex-wrap: wrap;
	}
	.m-link {
		margin-top: 10px;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		min-height: 44px;
		font-size: 14px;
		font-weight: 600;
		color: #c24e25;
	}
	.nw {
		white-space: nowrap;
	}
	.estado {
		margin-top: 12px;
		display: flex;
		align-items: flex-start;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
		line-height: 1.45;
	}
	.estado :global(svg) {
		flex: none;
		margin-top: 2px;
	}
</style>
