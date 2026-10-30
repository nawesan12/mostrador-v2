<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { nombreProvincia } from '#lib/tienda/provincias.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import type { components } from '@mostrador/contrato/esquema';

	type Envio = components['schemas']['EnvioCorreo'];
	type Entrega = NonNullable<components['schemas']['PedidoDetalle']['entrega']>;

	/**
	 * Correo Argentino en la hoja del pedido: cómo va (domicilio o sucursal), generar el envío en
	 * MiCorreo y el número de seguimiento. MiCorreo no da el número por API: llega al pagar el rótulo
	 * allá, y la tienda lo pega acá. El handoff no lo diseña: se arma con las piezas de la hoja.
	 */
	let { numero, entrega }: { numero: number; entrega: Entrega | null } = $props();
	const ui = getUi();

	let envio = $state<Envio | null>(null);
	let error = $state<string | null>(null);
	let ocupado = $state(false);
	let numeroSeg = $state('');

	const ruta = $derived(`/panel/pedidos/${numero}/correo`);

	async function pedir(metodo: 'GET' | 'POST' | 'PUT', cuerpo?: unknown) {
		const r = await fetch(ruta, {
			method: metodo,
			headers: { 'content-type': 'application/json' },
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!r.ok) {
			const msg = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			throw new Error(msg ?? 'No pudimos hablar con el servidor. Probá de nuevo.');
		}
		return (await r.json()) as Envio;
	}

	$effect(() => {
		const n = numero;
		envio = null;
		error = null;
		numeroSeg = '';
		pedir('GET')
			.then((e) => {
				if (n === numero) envio = e;
			})
			.catch((e: Error) => {
				if (n === numero) error = e.message;
			});
	});

	async function generar() {
		ocupado = true;
		error = null;
		try {
			envio = await pedir('POST');
			ui.toast({
				k: 'ok',
				title: 'Envío cargado en MiCorreo',
				desc: 'Pagalo e imprimí el rótulo allá'
			});
		} catch (e) {
			error = (e as Error).message;
		} finally {
			ocupado = false;
		}
	}

	async function guardarSeguimiento(e: SubmitEvent) {
		e.preventDefault();
		ocupado = true;
		error = null;
		try {
			envio = await pedir('PUT', { seguimiento: numeroSeg });
			numeroSeg = '';
			ui.toast({ k: 'ok', title: 'Seguimiento guardado', desc: 'Quien compra ya lo ve' });
		} catch (e) {
			error = (e as Error).message;
		} finally {
			ocupado = false;
		}
	}

	const fecha = (iso: string) =>
		new Date(iso).toLocaleDateString('es-AR', {
			day: 'numeric',
			month: 'short',
			timeZone: 'America/Argentina/Buenos_Aires'
		});
</script>

<div class="correo">
	{#if entrega?.tipo === 'S'}
		<div class="dato">
			A sucursal: <b>{entrega.sucursalNombre || entrega.sucursal}</b>{#if entrega.provincia}
				· {nombreProvincia(entrega.provincia)}{/if}
		</div>
	{:else if entrega?.tipo === 'D'}
		<div class="dato">
			A domicilio: <b>{entrega.direccion}</b>, {entrega.localidad}{#if entrega.provincia}
				· {nombreProvincia(entrega.provincia)}{/if} (CP&nbsp;{entrega.cp})
		</div>
	{/if}

	{#if !envio && !error}
		<div class="dato">Buscando el envío…</div>
	{:else if envio && !envio.conectada && !envio.generado}
		<div class="dato">
			Para cargarlo en MiCorreo, conectá Correo Argentino en
			<a href={resolve('/panel/integraciones')}>Integraciones</a>.
		</div>
	{:else if envio && !envio.generado}
		<button type="button" class="principal" disabled={ocupado} onclick={generar}
			>{ocupado ? 'Cargando en MiCorreo…' : 'Generar envío en MiCorreo'}</button
		>
	{:else if envio}
		<div class="dato">
			<Icon icon={I.phCheckCircleFill} size={14} /> Cargado en MiCorreo{#if envio.importado}
				el {fecha(envio.importado)}{/if}.
			{#if !envio.seguimiento}Pagalo e imprimí el rótulo en
				<a href="https://micorreo.correoargentino.com.ar" target="_blank" rel="noopener">MiCorreo</a
				>.{/if}
		</div>
		{#if envio.seguimiento}
			<div class="dato">
				Seguimiento <b>{envio.seguimiento}</b> ·
				<a
					href="https://www.correoargentino.com.ar/formularios/e-commerce"
					target="_blank"
					rel="noopener">Ver en Correo Argentino</a
				>
			</div>
		{/if}
		<form class="seg" onsubmit={guardarSeguimiento}>
			<input
				class="input"
				bind:value={numeroSeg}
				maxlength="40"
				autocomplete="off"
				placeholder={envio.seguimiento ? 'Cambiar el número' : 'Número de seguimiento'}
				aria-label="Número de seguimiento"
			/>
			<button class="borde" disabled={ocupado || numeroSeg.trim().length < 6}>Guardar</button>
		</form>
	{/if}
	{#if error}<div class="error" role="alert">{error}</div>{/if}
</div>

<style>
	.correo {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.dato {
		font-size: 13px;
		line-height: 1.45;
		color: var(--mu);
		text-wrap: pretty;
	}
	.dato b {
		color: var(--tx);
	}
	.dato :global(svg) {
		color: #237550;
		vertical-align: -2px;
	}
	.dato a {
		color: var(--tx);
		text-decoration: underline;
	}
	.principal {
		min-height: 44px;
		border: 0;
		border-radius: 999px;
		background: var(--tx);
		color: var(--sf);
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.principal:disabled,
	.borde:disabled {
		opacity: 0.6;
		cursor: default;
	}
	.seg {
		display: flex;
		gap: 8px;
	}
	.seg .input {
		flex: 1;
		min-width: 0;
		min-height: 44px;
	}
	.borde {
		min-height: 44px;
		border: 1px solid var(--ln2);
		background: transparent;
		font-size: 13px;
		font-weight: 600;
		padding: 8px 14px;
		border-radius: 999px;
		color: var(--tx);
		white-space: nowrap;
		cursor: pointer;
	}
	.error {
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
		text-wrap: pretty;
	}
</style>
