<script lang="ts">
	import { plural } from '#lib/panel/formato.ts';
	import { invalidateAll } from '$app/navigation';
	import Hoja from '#lib/components/Hoja.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesosCortos } from '#lib/panel/formato.ts';
	import MotorLista from '#lib/panel/listas/MotorLista.svelte';
	import {
		configInventario,
		COLOR_ST,
		sugerido,
		type ItemInventario
	} from '#lib/panel/listas/inventario.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Inventario: la lista de PanelListas más lo de PanelCatalogo (KPIs, movimientos de hoy, contar
	 * stock y el pedido a proveedores que arma Tincho).
	 */
	let { data } = $props();
	const ui = getUi();
	const r = $derived(data.lista.resumen);

	async function todos(filtro = '') {
		const res = await fetch(`/panel/listas/inventario?cuantos=200${filtro}`);
		return res.ok ? ((await res.json()) as { filas: ItemInventario[] }).filas : [];
	}

	// ---- Pedido a proveedores ----
	let pedido = $state(false);
	let renglones = $state<(ItemInventario & { pedir: number })[]>([]);
	const porProveedor = $derived(
		[...new Set(renglones.map((x) => x.proveedor))].map((p) => ({
			p,
			items: renglones.filter((x) => x.proveedor === p)
		}))
	);
	async function abrirPedido(ids: string[] | null) {
		const items = ids
			? (await todos()).filter((x) => ids.includes(x.id))
			: await todos('&seg=out,crit,low');
		renglones = items.map((x) => ({ ...x, pedir: sugerido(x) }));
		pedido = true;
	}
	function mandar(proveedor: string, items: (ItemInventario & { pedir: number })[]) {
		const texto = [
			`Hola ${proveedor}! Te paso un pedido:`,
			...items
				.filter((x) => x.pedir > 0)
				.map((x) => `• ${x.pedir} u. de ${x.nombre}${x.sku ? ` (${x.sku})` : ''}`),
			'¡Gracias!'
		].join('\n');
		window.open(`https://wa.me/?text=${encodeURIComponent(texto)}`, '_blank', 'noopener');
		ui.toast({
			k: 'wa',
			title: `Pedido a ${proveedor}`,
			desc: 'Elegí el contacto en WhatsApp y mandalo'
		});
	}
	async function ingreso(items: (ItemInventario & { pedir: number })[]) {
		for (const x of items.filter((y) => y.pedir > 0))
			await fetch('/panel/listas/inventario', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ ids: [x.id], accion: 'ingreso', valor: x.pedir })
			});
		await invalidateAll();
		renglones = renglones.filter((x) => !items.includes(x));
		ui.toast({
			k: 'ok',
			title: 'Mercadería ingresada',
			desc: 'Sumamos el stock y quedó en los movimientos de hoy'
		});
		if (!renglones.length) pedido = false;
	}

	// ---- Contar stock ----
	let conteo = $state(false);
	let contados = $state<(ItemInventario & { contado: string })[]>([]);
	let filtroConteo = $state('');
	const visiblesConteo = $derived(
		contados.filter(
			(x) =>
				!filtroConteo.trim() ||
				`${x.nombre} ${x.sku}`.toLowerCase().includes(filtroConteo.toLowerCase())
		)
	);
	const diferencias = $derived(
		contados.filter((x) => x.contado !== '' && Number(x.contado) !== x.stock).length
	);
	async function abrirConteo() {
		contados = (await todos()).map((x) => ({ ...x, contado: '' }));
		filtroConteo = '';
		conteo = true;
	}
	async function guardarConteo() {
		const valor = contados
			.filter((x) => x.contado !== '' && Number(x.contado) !== x.stock)
			.map((x) => ({ id: x.id, cantidad: Math.max(0, Math.round(Number(x.contado))) }));
		const res = await fetch('/panel/listas/inventario', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ ids: [], accion: 'conteo', valor })
		});
		conteo = false;
		await invalidateAll();
		const n = res.ok ? ((await res.json()) as { hechos: number }).hechos : 0;
		ui.toast({
			k: 'ok',
			title: 'Conteo guardado',
			desc: `${n} ${n === 1 ? 'producto corregido' : 'productos corregidos'} · quedó en los movimientos`
		});
	}

	const config = configInventario(abrirPedido);
	const MOV: Record<string, [typeof I.phShoppingBagFill, string, string]> = {
		web: [I.phShoppingBagFill, 'rgba(210,88,47,.12)', '#C24E25'],
		local: [I.phStorefrontFill, 'rgba(210,88,47,.12)', '#C24E25'],
		ml: [I.phHandshakeFill, 'rgba(210,88,47,.12)', '#C24E25'],
		ingreso: [I.phTruckFill, 'rgba(47,143,99,.13)', '#2F8F63'],
		devolucion: [I.phArrowUUpLeftBold, 'rgba(47,143,99,.13)', '#2F8F63'],
		conteo: [I.phBarcodeBold, 'var(--chip)', 'var(--mu)'],
		ajuste: [I.phArrowsLeftRightBold, 'var(--chip)', 'var(--mu)']
	};
	const hora = (iso: string) =>
		new Intl.DateTimeFormat('es-AR', {
			timeZone: 'America/Argentina/Buenos_Aires',
			hour: '2-digit',
			minute: '2-digit',
			hourCycle: 'h23'
		}).format(new Date(iso));
</script>

<Seo title="Inventario · Panel" description="Stock de tu tienda" path="/panel/inventario" noindex />

<MotorLista {config} lista={data.lista}>
	{#snippet extra()}
		<button type="button" class="secundario" onclick={abrirConteo}
			><Icon icon={I.phBarcodeBold} size={15} />Contar stock</button
		>
	{/snippet}
	{#snippet antes()}
		<div class="kpis" use:anim={{ t: 'rise', delay: 30 }}>
			<div class="kpi">
				<span>Valor del inventario</span><b>{pesosCortos(r.valorTotal ?? 0)}</b><em
					>a precio de costo</em
				>
			</div>
			<div class="kpi">
				<span>Sin stock</span><b style:color="#B83A2E">{r.sinStock ?? 0}</b><em
					>{r.sinStockEnMl ? `${r.sinStockEnMl} publicados en ML` : 'ninguno en Mercado Libre'}</em
				>
			</div>
			<div class="kpi">
				<span>Stock bajo</span><b style:color="#8a5f0e">{r.bajo ?? 0}</b><em>menos que el mínimo</em
				>
			</div>
			<div class="kpi">
				<span>Rotación</span><b style:color="#237550">{plural(r.rotacion ?? 0, 'día', 'días')}</b
				><em>promedio</em>
			</div>
		</div>
	{/snippet}
	{#snippet despues()}
		<div class="movimientos" use:anim={{ t: 'rise', delay: 220 }}>
			<div class="m-t">Movimientos de hoy</div>
			<div class="m-lista">
				{#each data.movimientos as m, i (i)}
					{@const [icono, bg, fg] = MOV[m.tipo] ?? MOV.ajuste}
					<div class="mov">
						<span class="m-ico" style:background={bg} style:color={fg}
							><Icon icon={icono} size={15} /></span
						>
						<span class="m-txt">{m.texto}</span>
						<b style:color={m.cantidad < 0 ? '#C24E25' : '#2F8F63'}
							>{m.cantidad > 0 ? '+' : '−'}{Math.abs(m.cantidad)}</b
						>
						<span class="m-h">{hora(m.en)}</span>
					</div>
				{:else}
					<div class="vacio">Todavía no se movió nada hoy.</div>
				{/each}
			</div>
		</div>
	{/snippet}
</MotorLista>

<Hoja bind:open={pedido} titulo="Pedido a proveedores" ancho={560}>
	{#snippet cabecera()}
		<div class="h-cab">
			<span class="h-t">Pedido a proveedores</span>
			<button type="button" class="redondo" aria-label="Cerrar" onclick={() => (pedido = false)}
				><Icon icon={I.phXBold} size={14} /></button
			>
		</div>
	{/snippet}
	<div class="h-cuerpo">
		<div class="nota">
			<Icon icon={I.phSparkleFill} size={16} style="color:#b0441d;margin-top:2px;flex:none" />
			<div>
				<b>Tincho:</b> armé el pedido con lo que vendés por día, para que te alcance dos veces el mínimo.
				Cambiá las cantidades si querés.
			</div>
		</div>
		{#each porProveedor as g (g.p)}
			<div class="tarjeta">
				<div class="g-cab">
					<span class="g-t">{g.p}</span><span class="g-n">{g.items.length} SKU</span>
				</div>
				{#each g.items as x (x.id)}
					<div class="renglon">
						{#if x.foto}<img src={unsplash(x.foto, 80, '&h=80')} alt="" />{/if}
						<span class="r-txt"
							><span class="r-n">{x.nombre}</span><span class="r-s" style:color={COLOR_ST[x.st]}
								>{x.stock} u. · mínimo {x.min}</span
							></span
						>
						<input
							class="r-q"
							type="number"
							min="0"
							bind:value={x.pedir}
							aria-label="Unidades de {x.nombre}"
						/>
					</div>
				{/each}
				<div class="g-acc">
					<button type="button" class="borde" onclick={() => ingreso(g.items)}
						>Ya llegó: sumar al stock</button
					>
					<button type="button" class="wa" onclick={() => mandar(g.p, g.items)}
						><Icon icon={I.phWhatsappLogoFill} size={15} />Enviar por WhatsApp</button
					>
				</div>
			</div>
		{:else}
			<div class="vacio">No hay nada para reponer: todo está por encima del mínimo.</div>
		{/each}
	</div>
</Hoja>

<Hoja bind:open={conteo} titulo="Contar stock" ancho={560}>
	{#snippet cabecera()}
		<div class="h-cab">
			<span class="h-t">Contar stock</span>
			<button type="button" class="redondo" aria-label="Cerrar" onclick={() => (conteo = false)}
				><Icon icon={I.phXBold} size={14} /></button
			>
		</div>
		<label class="buscar">
			<Icon icon={I.phMagnifyingGlass} size={16} style="color:var(--mu)" />
			<input
				bind:value={filtroConteo}
				placeholder="Escaneá o buscá el producto"
				aria-label="Buscar en el conteo"
			/>
		</label>
	{/snippet}
	<div class="h-cuerpo">
		<div class="tarjeta">
			{#each visiblesConteo as x (x.id)}
				<div class="renglon">
					{#if x.foto}<img src={unsplash(x.foto, 80, '&h=80')} alt="" />{/if}
					<span class="r-txt"
						><span class="r-n">{x.nombre}</span><span class="r-s">Sistema: {x.stock} u.</span></span
					>
					<input
						class="r-q"
						type="number"
						min="0"
						placeholder={String(x.stock)}
						bind:value={x.contado}
						class:dif={x.contado !== '' && Number(x.contado) !== x.stock}
						aria-label="Contado de {x.nombre}"
					/>
				</div>
			{:else}
				<div class="vacio">Nada con ese nombre.</div>
			{/each}
		</div>
	</div>
	{#snippet pie()}
		<div class="pie">
			<span class="p-t"
				>{diferencias ? `${diferencias} con diferencia` : 'Contá y escribí lo que hay'}</span
			>
			<button type="button" class="principal" disabled={!diferencias} onclick={guardarConteo}
				>Guardar conteo</button
			>
		</div>
	{/snippet}
</Hoja>

<style>
	.secundario {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 13.5px;
		font-weight: 600;
		box-shadow: var(--sh);
		white-space: nowrap;
	}
	.kpis {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
		gap: 12px;
	}
	.kpi {
		padding: 16px 18px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.kpi span {
		display: block;
		font-size: 13.5px;
		color: var(--mu);
	}
	.kpi b {
		display: block;
		font-size: 28px;
		font-weight: 600;
		letter-spacing: -0.04em;
		margin-top: 4px;
	}
	.kpi em {
		display: block;
		font-style: normal;
		font-size: 12.5px;
		color: var(--mu);
	}
	@media (max-width: 759px) {
		.kpis {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
		.kpi b {
			font-size: 24px;
		}
	}
	.movimientos {
		margin-top: 14px;
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.m-t {
		font-size: 16px;
		font-weight: 600;
	}
	.m-lista {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.mov {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 9px 0;
		border-bottom: 1px solid var(--ln);
		font-size: 13.5px;
	}
	.m-ico {
		width: 30px;
		height: 30px;
		flex: none;
		border-radius: 9px;
		display: grid;
		place-items: center;
	}
	.m-txt {
		flex: 1;
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.m-h {
		font-size: 12px;
		color: var(--fa);
		min-width: 44px;
		text-align: right;
	}
	.vacio {
		padding: 18px 4px;
		font-size: 13.5px;
		color: var(--mu);
	}
	.h-cab {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.h-t {
		flex: 1;
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.redondo {
		width: 36px;
		height: 36px;
		border: 0;
		border-radius: 50%;
		background: var(--chip);
		color: var(--tx);
		display: grid;
		place-items: center;
	}
	.buscar {
		margin-top: 12px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 14px;
		border-radius: 999px;
		background: var(--sf3);
	}
	.buscar input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 14px;
		color: var(--tx);
	}
	.h-cuerpo {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.nota {
		display: flex;
		gap: 10px;
		padding: 12px 14px;
		border-radius: 16px;
		background: rgba(210, 88, 47, 0.1);
		border: 1px solid rgba(210, 88, 47, 0.22);
		font-size: 13.5px;
		line-height: 1.45;
	}
	.tarjeta {
		padding: 14px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.g-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.g-t {
		font-size: 15px;
		font-weight: 600;
	}
	.g-n {
		font-size: 12.5px;
		color: var(--mu);
	}
	.renglon {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.renglon img {
		width: 40px;
		height: 40px;
		flex: none;
		border-radius: 11px;
		object-fit: cover;
	}
	.r-txt {
		flex: 1;
		min-width: 0;
	}
	.r-n {
		display: block;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.r-s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.r-q {
		width: 76px;
		padding: 8px 10px;
		border-radius: 10px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		color: var(--tx);
		font-size: 14px;
		font-weight: 600;
		text-align: right;
		outline: none;
	}
	.r-q:focus,
	.r-q.dif {
		border-color: #d2582f;
	}
	.g-acc {
		margin-top: 12px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.borde {
		padding: 10px 14px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 13px;
		font-weight: 600;
	}
	.wa {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		padding: 10px 14px;
		border-radius: 999px;
		border: 0;
		background: #e9f8ee;
		color: #1f6f4a;
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
	}
	.pie {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.p-t {
		flex: 1;
		font-size: 13.5px;
		color: var(--mu);
	}
	.principal {
		padding: 13px 22px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 14.5px;
		font-weight: 600;
	}
	.principal:disabled {
		background: var(--chip);
		color: var(--mu);
	}
</style>
