<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import EditorProducto from '#lib/panel/listas/EditorProducto.svelte';
	import MotorLista from '#lib/panel/listas/MotorLista.svelte';
	import { configProductos, type ProductoFila } from '#lib/panel/listas/productos.ts';
	import { SITE } from '#lib/site.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import type { components } from '@mostrador/contrato/esquema';

	type Producto = components['schemas']['Producto'];

	/**
	 * Productos: la lista de PanelListas más lo de PanelCatalogo (cargar con foto, importar Excel y
	 * el editor lateral).
	 */
	let { data } = $props();
	const ui = getUi();
	const tienda = $derived(page.data.tienda as { nombre: string; slug: string });
	const dominio = $derived(`${SITE.replace(/^https?:\/\//, '')}/t/${tienda.slug}`);

	let editor = $state(false);
	let producto = $state<Producto | null>(null);
	let porcentaje = $state<string[]>([]);
	let cambiandoPrecio = $state(false);
	const config = configProductos(
		() => {
			producto = null;
			editor = true;
		},
		(ids) => {
			porcentaje = ids;
			cambiandoPrecio = true;
		}
	);

	async function abrir(p: ProductoFila) {
		const r = await fetch(`/panel/productos/${encodeURIComponent(p.id)}`);
		if (!r.ok)
			return ui.toast({
				k: 'warn',
				title: 'Ese producto no existe',
				desc: 'Actualizamos la lista.'
			});
		producto = await r.json();
		editor = true;
	}

	let pct = $state(8);
	async function aplicarPrecio(e: SubmitEvent) {
		e.preventDefault();
		const ids = porcentaje;
		const r = await fetch('/panel/listas/productos', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ ids, accion: 'precio', valor: pct })
		});
		cambiandoPrecio = false;
		await invalidateAll();
		if (!r.ok)
			return ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo en un rato.' });
		const { hechos } = (await r.json()) as { hechos: number };
		const vuelta = Math.round((100 / (1 + pct / 100) - 100) * 100) / 100;
		ui.toast({
			k: 'ok',
			title: `Precio ${pct > 0 ? '+' : ''}${pct}% · ${hechos} ${hechos === 1 ? 'producto' : 'productos'}`,
			desc: 'Redondeado a $ 10',
			action: 'Deshacer',
			onAction: async () => {
				await fetch('/panel/listas/productos', {
					method: 'POST',
					headers: { 'content-type': 'application/json' },
					body: JSON.stringify({ ids, accion: 'precio', valor: vuelta })
				});
				await invalidateAll();
			}
		});
	}

	// Importar desde Empretienda: primero la lista de fichas, después de a 5 con barra de avance.
	let empre = $state(false);
	let tiendaEt = $state('');
	let avance = $state<{ hechas: number; total: number } | null>(null);
	let errorEt = $state('');
	async function importarEmpretienda() {
		errorEt = '';
		avance = { hechas: 0, total: 0 };
		const r = await fetch('/panel/productos/empretienda', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ tienda: tiendaEt })
		});
		if (!r.ok) {
			avance = null;
			errorEt =
				((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
				'No pudimos leer la tienda.';
			return;
		}
		const { origen, fichas } = (await r.json()) as { origen: string; fichas: string[] };
		avance = { hechas: 0, total: fichas.length };
		const suma = { creados: 0, repetidos: 0, conVariantes: 0, fallidas: 0 };
		for (let i = 0; i < fichas.length; i += 5) {
			const l = await fetch('/panel/productos/empretienda/lote', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ origen, fichas: fichas.slice(i, i + 5) })
			});
			const x = l.ok
				? ((await l.json()) as {
						creados: number;
						repetidos: number;
						conVariantes: number;
						fallidas: string[];
					})
				: { creados: 0, repetidos: 0, conVariantes: 0, fallidas: fichas.slice(i, i + 5) };
			suma.creados += x.creados;
			suma.repetidos += x.repetidos;
			suma.conVariantes += x.conVariantes;
			suma.fallidas += x.fallidas.length;
			avance = { hechas: Math.min(fichas.length, i + 5), total: fichas.length };
		}
		await invalidateAll();
		empre = false;
		avance = null;
		const notas = [
			suma.repetidos && `${suma.repetidos} ya estaban`,
			suma.conVariantes && `${suma.conVariantes} con variantes anotadas en la descripción`,
			suma.fallidas && `${suma.fallidas} no se pudieron leer`
		].filter(Boolean);
		ui.toast({
			k: suma.creados ? 'ok' : 'info',
			title: `${suma.creados} ${suma.creados === 1 ? 'producto importado' : 'productos importados'} de Empretienda`,
			desc: notas.length ? notas.join(' · ') + '.' : 'Con fotos, precios, stock y categorías.'
		});
	}

	let archivo = $state<HTMLInputElement>();
	async function importar(e: Event) {
		// currentTarget deja de existir después del primer await: se guarda antes.
		const input = e.currentTarget as HTMLInputElement;
		const f = input.files?.[0];
		if (!f) return;
		if (!/\.(csv|txt)$/i.test(f.name)) {
			ui.toast({
				k: 'warn',
				title: 'Guardalo como CSV',
				desc: 'En Excel: Archivo › Guardar como › CSV.'
			});
			return;
		}
		const r = await fetch('/panel/productos/importar', { method: 'POST', body: await f.text() });
		input.value = '';
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({
				k: 'err',
				title: 'No pudimos importarlo',
				desc: m ?? 'Revisá el archivo.'
			});
		}
		const { creados, errores } = (await r.json()) as { creados: number; errores: string[] };
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: `${creados} ${creados === 1 ? 'producto importado' : 'productos importados'}`,
			desc: errores.length
				? `${errores.length} filas sin nombre quedaron afuera.`
				: 'Ya están publicados en tu tienda.'
		});
	}
</script>

<Seo
	title="Productos · Panel"
	description="Productos de tu tienda"
	path="/panel/productos"
	noindex
/>

<input bind:this={archivo} type="file" accept=".csv,text/csv" hidden onchange={importar} />

<Modal bind:open={empre} titulo="Importar desde Empretienda" ancho={460} padding={24}>
	<span class="m-ico"><Icon icon={I.phStorefrontFill} size={26} style="color:#2E2A6B" /></span>
	<div class="m-t">Traé tus productos de&nbsp;Empretienda</div>
	<div class="m-d">
		Pegá el link de tu tienda y traemos cada producto con sus fotos, precio, stock, categoría y
		descripción. Si importás dos veces, no se&nbsp;duplican.
	</div>
	{#if avance}
		<div class="m-c">
			{avance.total ? `Importando ${avance.hechas} de ${avance.total}…` : 'Leyendo tu tienda…'}
		</div>
		<div
			class="m-barra"
			role="progressbar"
			aria-label="Avance de la importación"
			aria-valuenow={avance.hechas}
			aria-valuemax={avance.total || 1}
		>
			<span style="width:{avance.total ? (avance.hechas / avance.total) * 100 : 4}%"></span>
		</div>
	{:else}
		<label class="m-c" for="tienda-et">Link de tu tienda</label>
		<input
			id="tienda-et"
			class="m-in"
			bind:value={tiendaEt}
			placeholder="mitienda.empretienda.com.ar"
			autocomplete="url"
			onkeydown={(e) => e.key === 'Enter' && tiendaEt.trim() && importarEmpretienda()}
		/>
		{#if errorEt}<div class="m-err" role="alert">{errorEt}</div>{/if}
		<div class="m-acc">
			<button type="button" class="b-sec" onclick={() => (empre = false)}>Cancelar</button>
			<button type="button" class="b-pri" disabled={!tiendaEt.trim()} onclick={importarEmpretienda}
				>Importar</button
			>
		</div>
	{/if}
</Modal>

<MotorLista {config} lista={data.lista} {abrir}>
	{#snippet extra()}
		<button type="button" class="secundario" onclick={() => archivo?.click()}
			><Icon icon={I.phFileXlsBold} size={15} />Importar Excel</button
		>
		<button type="button" class="secundario" onclick={() => (empre = true)}
			><Icon icon={I.phLinkBold} size={15} />Desde Empretienda</button
		>
	{/snippet}
	{#snippet antes()}
		<div class="foto">
			<div class="arcos"></div>
			<img src="/hornero/foto.webp" alt="" width="64" height="64" />
			<div class="f-txt">
				<div class="f-t">Cargá productos con una foto</div>
				<div class="f-d">
					Hornero completa nombre, descripción, categoría y precio sugerido. Vos solo confirmás.
				</div>
			</div>
			<button
				type="button"
				onclick={() => {
					ui.toast({
						k: 'info',
						title: 'Subir fotos llega con el almacenamiento',
						desc: 'Mientras, cargá el producto a mano: Hornero te ayuda con la descripción.'
					});
					producto = null;
					editor = true;
				}}><Icon icon={I.phCameraFill} size={16} style="color:#b0441d" />Subir fotos</button
			>
		</div>
	{/snippet}
</MotorLista>

<EditorProducto
	bind:open={editor}
	{producto}
	tienda={tienda.nombre}
	{dominio}
	guardado={async (p, nuevo) => {
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: nuevo ? 'Producto creado' : 'Producto guardado',
			desc: `${p.nombre} · publicado en ${Object.values(p.canales ?? {}).filter(Boolean).length} canales`
		});
	}}
/>

<Modal bind:open={cambiandoPrecio} titulo="Cambiar precio" ancho={420}>
	<form onsubmit={aplicarPrecio}>
		<div class="m-ico"><Icon icon={I.phPercentBold} size={24} /></div>
		<div class="m-t">
			Cambiar precio de {porcentaje.length}
			{porcentaje.length === 1 ? 'producto' : 'productos'}
		</div>
		<div class="m-d">Subí o bajá todo de una. Redondeamos a $ 10.</div>
		<div class="m-rango">
			<input type="range" min="-30" max="50" step="1" bind:value={pct} aria-label="Porcentaje" />
			<b>{pct > 0 ? '+' : ''}{pct}%</b>
		</div>
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (cambiandoPrecio = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary" disabled={!pct}>Aplicar</button>
		</div>
	</form>
</Modal>

<style>
	.foto {
		position: relative;
		overflow: hidden;
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 16px 18px;
		border-radius: 22px;
		background: linear-gradient(120deg, #2a140c, #4a2416);
		color: #fbf7f2;
		flex-wrap: wrap;
	}
	.arcos {
		position: absolute;
		inset: 0;
		background-image: url('/textura/arch-pattern-light.svg');
	}
	.foto img {
		position: relative;
		object-fit: contain;
	}
	.f-txt {
		position: relative;
		flex: 1 1 260px;
	}
	.f-t {
		font-size: 16px;
		font-weight: 600;
	}
	.f-d {
		font-size: 13.5px;
		color: #d9c8ba;
	}
	.foto button {
		position: relative;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 18px;
		border-radius: 999px;
		border: 0;
		background: #fbf7f2;
		color: #2a140c;
		font-size: 13.5px;
		font-weight: 600;
		white-space: nowrap;
	}
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
	.m-ico {
		width: 48px;
		height: 48px;
		display: grid;
		place-items: center;
		border-radius: 15px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
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
	.m-rango {
		margin-top: 18px;
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.m-rango input {
		flex: 1;
		accent-color: #d2582f;
	}
	.m-rango b {
		min-width: 52px;
		text-align: right;
		font-size: 18px;
	}
	.m-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
	.m-ico {
		width: 48px;
		height: 48px;
		border-radius: 16px;
		background: rgba(46, 42, 107, 0.1);
		display: grid;
		place-items: center;
	}
	.m-t {
		margin-top: 12px;
		font-size: 19px;
		font-weight: 600;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14px;
		line-height: 1.55;
		color: var(--mu);
	}
	.m-c {
		display: block;
		margin-top: 14px;
		font-size: 13px;
		font-weight: 600;
	}
	.m-in {
		margin-top: 6px;
		width: 100%;
		padding: 12px 14px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: var(--sf2);
		font:
			500 15px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
	}
	.m-in:focus {
		border-color: var(--terracota);
	}
	.m-err {
		margin-top: 8px;
		font-size: 13px;
		color: #b83a2e;
	}
	.m-barra {
		margin-top: 8px;
		height: 8px;
		border-radius: 999px;
		background: var(--sf3);
		overflow: hidden;
	}
	.m-barra span {
		display: block;
		height: 100%;
		border-radius: inherit;
		background: var(--terracota);
		transition: width 0.4s;
	}
	.m-acc {
		margin-top: 16px;
		display: flex;
		gap: 8px;
	}
	.m-acc button {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.b-sec {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.b-pri {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.b-pri:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
</style>
