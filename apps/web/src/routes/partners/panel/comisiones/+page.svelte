<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Hoja from '#lib/components/Hoja.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Vacio from '#lib/components/Vacio.svelte';
	import * as I from '#lib/icons.ts';
	import { EMPRESA } from '#lib/marca/datos.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import {
		cuandoSePaga,
		fechaDePago,
		fechaLarga,
		nombreMes,
		type CobroPartner,
		type PagoPartner
	} from '#lib/partners/programa.ts';
	import { CONDICIONES_IVA, validarCobro } from '#lib/partners/validar.ts';
	import { ESTADO_PAGO, guardar, nombrePlan } from '#lib/partners/vista.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Comisiones y pagos: mes a mes por tienda, estado, factura y datos de cobro. */
	let { data } = $props();
	const ui = getUi();
	const p = $derived(data.panel!);

	// El mes en curso, estimado con las tiendas de hoy.
	const mes = $derived(p.hoy.slice(0, 7));
	const estimado = $derived<PagoPartner>({
		mes,
		estado: 'pendiente',
		monto: p.tiendas.reduce((s, t) => s + t.comisionMes, 0),
		detalle: p.tiendas
			.filter((t) => t.comisionMes > 0)
			.map((t) => ({ tiendaId: t.id, tienda: t.nombre, plan: t.plan, comision: t.comisionMes })),
		facturaUrl: null,
		fechaPago: fechaDePago(mes),
		pagadoEn: null
	});
	const filas = $derived([
		{ g: estimado, estimado: true },
		...p.pagos.filter((g) => g.mes !== mes).map((g) => ({ g, estimado: false }))
	]);

	// ---- Detalle de un mes ----
	let abierto = $state(false);
	let elegido = $state<{ g: PagoPartner; estimado: boolean } | null>(null);
	let factura = $state('');
	let errFactura = $state<string | null>(null);
	function abrir(f: { g: PagoPartner; estimado: boolean }) {
		elegido = f;
		factura = f.g.facturaUrl ?? '';
		errFactura = null;
		abierto = true;
	}
	async function subir(e: SubmitEvent) {
		e.preventDefault();
		if (!elegido) return;
		const err = await guardar({ que: 'factura', mes: elegido.g.mes, url: factura });
		if (err) {
			errFactura = err;
			return;
		}
		await invalidateAll();
		abierto = false;
		ui.toast({
			k: 'ok',
			title: 'Factura cargada',
			desc: `${nombreMes(elegido.g.mes)} · ${cuandoSePaga({ ...elegido.g, facturaUrl: factura }, p.hoy)}`
		});
	}

	// ---- Datos de cobro ----
	const VACIO: CobroPartner = {
		tipo: 'alias',
		valor: '',
		titular: '',
		cuit: '',
		condicionIva: 'monotributo'
	};
	let editando = $state(false);
	let cobro = $state<CobroPartner>({ ...VACIO });
	let errCobro = $state<string | null>(null);
	$effect(() => {
		if (!editando) cobro = { ...(p.partner.cobro ?? VACIO) };
	});
	async function guardarCobro(e: SubmitEvent) {
		e.preventDefault();
		const v = validarCobro(cobro);
		if (!v.ok) {
			errCobro = v.mensaje;
			return;
		}
		const err = await guardar({ que: 'cobro', cobro: v.valor });
		if (err) {
			errCobro = err;
			return;
		}
		errCobro = null;
		editando = false;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Datos de cobro guardados',
			desc: 'Te pagamos ahí desde el próximo mes'
		});
	}
	const condicion = (k: string) => CONDICIONES_IVA.find(([c]) => c === k)?.[1] ?? k;
	const formulario = $derived(editando || !p.partner.cobro);
</script>

<div class="tarjeta" use:anim={{ t: 'rise' }}>
	<div class="t-cab"><span>Mes a mes</span></div>
	<div class="t-d">
		Liquidamos a fin de mes y pagamos el 10 del mes siguiente, con tu&nbsp;factura.
	</div>
	<ul class="filas lista" aria-label="Pagos">
		{#each filas as f (f.g.mes)}
			{@const c = ESTADO_PAGO[f.estimado ? 'estimado' : f.g.estado]}
			<li>
				<button type="button" class="f-in" onclick={() => abrir(f)}>
					<span class="ini" style:background={c.bg} style:color={c.fg}
						><Icon icon={f.estimado ? I.phHourglassMediumFill : I.phReceiptFill} size={18} /></span
					>
					<span class="quien">
						<span class="q-t">{nombreMes(f.g.mes)}{f.estimado ? ' (en curso)' : ''}</span>
						<span class="q-s"
							>{f.g.detalle.length}
							{f.g.detalle.length === 1 ? 'tienda' : 'tiendas'} · {cuandoSePaga(
								f.g,
								p.hoy
							)}{!f.estimado && f.g.estado !== 'pagada'
								? f.g.facturaUrl
									? ' · factura cargada'
									: ' · falta tu factura'
								: ''}</span
						>
					</span>
					<span class="chip-f" style:background={c.bg} style:color={c.fg}>{c.t}</span>
					<span class="valor"><span class="v-v">{pesos(f.g.monto)}</span></span>
				</button>
			</li>
		{/each}
	</ul>
</div>

<div class="tarjeta" use:anim={{ t: 'rise', delay: 40 }}>
	<div class="t-cab">
		<span>Datos de cobro</span>
		{#if p.partner.cobro && !editando}
			<button type="button" class="oscuro" onclick={() => (editando = true)}
				><Icon icon={I.phPencilSimpleFill} size={14} />Editar</button
			>
		{/if}
	</div>
	{#if !formulario && p.partner.cobro}
		{@const c = p.partner.cobro}
		<div class="filas">
			<div class="opcion">
				<span class="o-ico"><Icon icon={I.phBankFill} size={18} /></span>
				<span class="o-txt"
					><span class="o-t">{c.tipo === 'cbu' ? 'CBU o CVU' : 'Alias'} · {c.valor}</span><span
						class="o-d">{c.titular}</span
					></span
				>
			</div>
			<div class="opcion">
				<span class="o-ico"><Icon icon={I.phIdentificationCardFill} size={18} /></span>
				<span class="o-txt"
					><span class="o-t">CUIT {c.cuit}</span><span class="o-d">{condicion(c.condicionIva)}</span
					></span
				>
			</div>
		</div>
	{:else}
		<div class="t-d">Necesitamos esto para transferirte. Lo usamos sólo para tus&nbsp;pagos.</div>
		<form class="campos form" onsubmit={guardarCobro}>
			<div class="campo-t ancho">
				¿Dónde te pagamos?
				<div class="chips" role="group" aria-label="Dónde te pagamos">
					<button
						type="button"
						aria-pressed={cobro.tipo === 'alias'}
						onclick={() => (cobro = { ...cobro, tipo: 'alias' })}
						>Alias (Mercado Pago o banco)</button
					>
					<button
						type="button"
						aria-pressed={cobro.tipo === 'cbu'}
						onclick={() => (cobro = { ...cobro, tipo: 'cbu' })}>CBU o CVU</button
					>
				</div>
			</div>
			<label
				>{cobro.tipo === 'cbu' ? 'CBU o CVU (22 números)' : 'Alias'}<input
					bind:value={cobro.valor}
					maxlength={cobro.tipo === 'cbu' ? 26 : 20}
					inputmode={cobro.tipo === 'cbu' ? 'numeric' : 'text'}
					autocomplete="off"
				/></label
			>
			<label>Titular de la cuenta<input bind:value={cobro.titular} maxlength="80" /></label>
			<label
				>CUIT<input
					bind:value={cobro.cuit}
					maxlength="13"
					inputmode="numeric"
					placeholder="20-12345678-6"
				/></label
			>
			<div class="campo-t ancho">
				Condición frente al IVA
				<div class="chips" role="group" aria-label="Condición frente al IVA">
					{#each CONDICIONES_IVA as [k, t] (k)}
						<button
							type="button"
							aria-pressed={cobro.condicionIva === k}
							onclick={() => (cobro = { ...cobro, condicionIva: k })}>{t}</button
						>
					{/each}
				</div>
			</div>
			{#if errCobro}<div class="error ancho" role="alert">{errCobro}</div>{/if}
			<div class="guardar">
				{#if p.partner.cobro}
					<button
						type="button"
						class="btn btn-secondary"
						onclick={() => {
							editando = false;
							errCobro = null;
						}}>Descartar</button
					>
				{/if}
				<button class="btn btn-primary">Guardar datos de cobro</button>
			</div>
		</form>
	{/if}
</div>

<Hoja bind:open={abierto} titulo={elegido ? nombreMes(elegido.g.mes) : 'Pago'}>
	{#if elegido}
		{@const g = elegido.g}
		{@const c = ESTADO_PAGO[elegido.estimado ? 'estimado' : g.estado]}
		<div class="h-in pm">
			<div class="m-t">{nombreMes(g.mes)}</div>
			<div class="h-cab">
				<span class="chip-f" style:background={c.bg} style:color={c.fg}>{c.t}</span>
				<span class="h-monto">{pesos(g.monto)}</span>
			</div>
			<div class="m-d">
				{#if elegido.estimado}
					Con las tiendas activas de hoy. Se liquida a fin de mes y se paga el {fechaLarga(
						g.fechaPago
					)}.
				{:else}
					{cuandoSePaga(g, p.hoy).replace(/^./, (x) => x.toUpperCase())}.
				{/if}
			</div>

			{#if !elegido.estimado}
				{#if g.estado === 'pagada'}
					{#if g.facturaUrl}
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- la factura del partner -->
						<a class="mas" href={g.facturaUrl} target="_blank" rel="noopener"
							><Icon icon={I.phFilePdfFill} size={15} />Ver tu factura</a
						>
					{/if}
				{:else}
					<form class="m-campo" onsubmit={subir}>
						<label class="field"
							>Link de tu factura<input
								class="input"
								bind:value={factura}
								maxlength="500"
								placeholder="https://drive.google.com/…"
								inputmode="url"
							/></label
						>
						<div class="f-ayuda">
							A nombre de {EMPRESA.titular} (CUIT {EMPRESA.cuit}) por {pesos(g.monto)}. Subila a
							Drive o Dropbox y pegá el&nbsp;link.
						</div>
						{#if errFactura}<div class="m-error" role="alert">{errFactura}</div>{/if}
						<div class="m-acc">
							<button class="btn btn-primary"
								>{g.facturaUrl ? 'Cambiar la factura' : 'Cargar factura'}</button
							>
						</div>
					</form>
				{/if}
			{/if}

			<div class="h-t">Por tienda</div>
			{#if g.detalle.length}
				<div class="h-lista">
					{#each g.detalle as l (l.tiendaId)}
						<div class="h-fila">
							<span class="quien"
								><span class="q-t">{l.tienda}</span><span class="q-s"
									>Plan {nombrePlan(l.plan)}</span
								></span
							>
							<span class="v-v">{pesos(l.comision)}</span>
						</div>
					{/each}
				</div>
			{:else}
				<Vacio hornero="horno-vacio" titulo="Sin tiendas con comisión este mes" />
			{/if}
		</div>
	{/if}
</Hoja>

<style>
	.form {
		margin-top: 14px;
	}
	.lista {
		list-style: none;
		margin: 8px 0 0;
		padding: 0;
	}
	.lista .f-in {
		width: calc(100% + 24px);
	}
	.error {
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	.h-in {
		display: flex;
		flex-direction: column;
	}
	.h-cab {
		margin-top: 8px;
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.h-monto {
		font-size: 26px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.f-ayuda {
		margin-top: 6px;
		font-size: 12.5px;
		color: var(--mu);
	}
	.h-t {
		margin-top: 22px;
		font-size: 14px;
		font-weight: 600;
	}
	.h-lista {
		margin-top: 6px;
		display: flex;
		flex-direction: column;
	}
	.h-fila {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.h-fila .quien {
		flex: 1;
		min-width: 0;
	}
	.h-fila .q-t {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.h-fila .q-s {
		display: block;
		font-size: 12px;
		color: var(--fa);
	}
	.h-fila .v-v {
		font-size: 14.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.chip-f {
		font-size: 11.5px;
		font-weight: 700;
		padding: 4px 8px;
		border-radius: 999px;
		white-space: nowrap;
	}
	.mas {
		margin-top: 12px;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
		color: #c24e25;
		text-decoration: none;
	}
	@media (max-width: 759px) {
		.f-in .chip-f {
			display: none;
		}
	}
</style>
