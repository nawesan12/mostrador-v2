<script lang="ts">
	import { invalidateAll, replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import EditorProducto from '#lib/panel/listas/EditorProducto.svelte';
	import MotorLista from '#lib/panel/listas/MotorLista.svelte';
	import { configProductos, type ProductoFila } from '#lib/panel/listas/productos.ts';
	import { SITE } from '#lib/site.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import type { components } from '@mostrador/contrato/esquema';

	type Producto = components['schemas']['Producto'];
	type Importacion = components['schemas']['Importacion'];

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

	// Importar tienda (Shopify, WooCommerce o Empretienda): primero cuántas páginas hay, después
	// una página por pedido con barra de avance. Al terminar, el resumen queda en el modal.
	type Fuente = 'shopify' | 'woocommerce' | 'empretienda';
	const FUENTES = [
		['shopify', 'Shopify'],
		['woocommerce', 'WooCommerce'],
		['empretienda', 'Empretienda']
	] as const;
	const NOMBRE = Object.fromEntries(FUENTES) as Record<Fuente, string>;
	const EJEMPLO: Record<Fuente, string> = {
		shopify: 'mitienda.myshopify.com',
		woocommerce: 'mitienda.com.ar',
		empretienda: 'mitienda.empretienda.com.ar'
	};
	type Lote = {
		leidos: number;
		creados: number;
		repetidos: number;
		recortados: number;
		fallidos: number;
		conVariantes: number;
		stockEstimado: boolean;
		fin: boolean;
	};
	let importando = $state(false);
	let fuente = $state<Fuente>('shopify');
	let tiendaImp = $state('');
	let avance = $state<{ hechas: number; total: number | null } | null>(null);
	let errorImp = $state('');
	let resumen = $state<(Omit<Lote, 'leidos' | 'fin'> & { fuente: Fuente }) | null>(null);
	const post = (url: string, cuerpo: unknown) =>
		fetch(url, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(cuerpo)
		});
	function abrirImportar() {
		resumen = null;
		errorImp = '';
		importando = true;
		void leerCola();
	}

	// Tiendanube va por OAuth y por la cola de importación (productos, clientes y pedidos): el modal
	// muestra cómo va la importación actual. En la demo cada "avanzar" hace un paso; con la API avanza
	// sola y "avanzar" sólo lee, así que si nada cambió se espera un poco antes de volver a preguntar.
	let cola = $state<Importacion | null>(null);
	let tiendanube = $state<boolean | null>(null);
	let avisoTN = $state('');
	const activa = (i: Importacion | null) => i?.estado === 'pendiente' || i?.estado === 'corriendo';
	const AVISOS: Record<string, string> = {
		'no-disponible': 'La conexión con Tiendanube todavía no está disponible.',
		cancelada: 'No se conectó la tienda: Tiendanube no dio el permiso.',
		estado: 'La conexión venció. Probá de nuevo.',
		canje: 'Tiendanube no aceptó la conexión. Probá de nuevo.',
		activa: 'Ya hay una importación en curso: esperá que termine.',
		'sin-cifrado': 'No podemos guardar la clave de tu tienda. Probá más tarde.',
		'sin-tiendanube': 'La conexión con Tiendanube todavía no está disponible.'
	};
	async function leerCola() {
		const r = await fetch('/panel/importaciones').catch(() => null);
		if (!r?.ok) return;
		const d = (await r.json()) as { importacion: Importacion | null; tiendanube: boolean };
		tiendanube = d.tiendanube;
		// Sólo se muestra la de Tiendanube (las de link van lote por lote, sin la cola).
		if (d.importacion?.plataforma === 'tiendanube' && (activa(d.importacion) || cola)) {
			cola = d.importacion;
			if (activa(cola)) void seguir();
		}
	}
	let siguiendo = false;
	async function seguir() {
		if (siguiendo) return;
		siguiendo = true;
		try {
			while (importando && activa(cola)) {
				const antes = JSON.stringify(cola);
				const r = await post('/panel/importaciones', { accion: 'avanzar' }).catch(() => null);
				if (!r?.ok) break;
				cola = (await r.json()) as Importacion | null;
				if (JSON.stringify(cola) === antes) await new Promise((ok) => setTimeout(ok, 1500));
			}
		} finally {
			siguiendo = false;
		}
		if (cola && !activa(cola)) await invalidateAll();
	}
	async function cancelarCola() {
		if (!cola) return;
		const r = await post('/panel/importaciones', { accion: 'cancelar', id: cola.id });
		if (r.ok) cola = (await r.json()) as Importacion;
	}
	function cerrarCola() {
		cola = null;
		importando = false;
	}
	// La vuelta del OAuth trae ?tiendanube=importando (o el motivo del error): se abre el modal.
	$effect(() => {
		const tn = page.url.searchParams.get('tiendanube');
		if (!tn) return;
		replaceState(resolve('/panel/productos'), {});
		resumen = null;
		errorImp = '';
		avisoTN = tn === 'importando' ? '' : (AVISOS[tn] ?? 'No pudimos conectar tu tienda.');
		importando = true;
		if (tn === 'importando') cola = { estado: 'pendiente' } as Importacion;
		void leerCola();
	});
	const ETAPAS_COLA = [
		['productos', 'Productos'],
		['clientes', 'Clientes'],
		['pedidos', 'Pedidos']
	] as const;
	/** Cuánto va: por etapa, porque Tiendanube no dice el total hasta leer cada una. */
	const avanceCola = $derived(
		!cola || !cola.etapa
			? 4
			: cola.estado === 'lista'
				? 100
				: Math.max(
						4,
						(['productos', 'clientes', 'pedidos', 'redirecciones'].indexOf(cola.etapa) + 0.5) * 25
					)
	);
	async function importarTienda() {
		errorImp = '';
		avance = { hechas: 0, total: null };
		const r = await post('/panel/productos/tienda', { fuente, tienda: tiendaImp });
		if (!r.ok) {
			avance = null;
			errorImp =
				((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
				'No pudimos leer la tienda.';
			return;
		}
		const d = (await r.json()) as {
			fuente: Fuente;
			origen: string;
			total: number | null;
			paginas: number | null;
			fichas?: string[];
		};
		// Si el link era de otra plataforma, se usa la que respondió.
		fuente = d.fuente;
		avance = { hechas: 0, total: d.total };
		const suma = {
			creados: 0,
			repetidos: 0,
			recortados: 0,
			fallidos: 0,
			conVariantes: 0,
			stockEstimado: false
		};
		// Sin total (Shopify no siempre lo dice): hasta una página incompleta, con el tope de 100.
		const paginas = d.paginas ?? 100;
		for (let pagina = 1; pagina <= paginas; pagina++) {
			const fichas = d.fichas?.slice((pagina - 1) * 5, pagina * 5);
			const l = await post('/panel/productos/tienda/lote', {
				fuente: d.fuente,
				origen: d.origen,
				pagina,
				...(fichas ? { fichas } : {})
			});
			// Una página que no respondió cuenta como no leída y se sigue con la próxima.
			const x: Lote = l.ok
				? await l.json()
				: {
						creados: 0,
						repetidos: 0,
						recortados: 0,
						conVariantes: 0,
						leidos: 5,
						fallidos: fichas?.length ?? 5,
						stockEstimado: false,
						fin: false
					};
			suma.creados += x.creados;
			suma.repetidos += x.repetidos;
			suma.recortados += x.recortados;
			suma.fallidos += x.fallidos;
			suma.conVariantes += x.conVariantes;
			suma.stockEstimado ||= x.stockEstimado;
			const hechas: number = (avance?.hechas ?? 0) + x.leidos;
			avance = { hechas: d.total ? Math.min(d.total, hechas) : hechas, total: d.total };
			if (x.fin && !d.fichas) break;
		}
		await invalidateAll();
		avance = null;
		resumen = { ...suma, fuente: d.fuente };
	}
	/** Los renglones del resumen: sólo los que tienen algo que decir. */
	const notas = $derived(
		resumen
			? [
					resumen.repetidos &&
						(resumen.repetidos === 1
							? '1 ya estaba y no se\u00a0repitió.'
							: `${resumen.repetidos} ya estaban y no se\u00a0repitieron.`),
					resumen.recortados &&
						`${resumen.recortados} ${resumen.recortados === 1 ? 'se recortó' : 'se recortaron'}: más de 8\u00a0fotos, opciones o\u00a0textos largos.`,
					resumen.conVariantes &&
						`${resumen.conVariantes} con variantes anotadas en la\u00a0descripción.`,
					resumen.fallidos && `${resumen.fallidos} no se pudieron\u00a0leer.`,
					resumen.stockEstimado &&
						(resumen.fuente === 'shopify'
							? 'Shopify no publica el stock: revisalo.'
							: `${NOMBRE[resumen.fuente]} no publicó el stock de algunos: revisalo.`)
				].filter((x): x is string => !!x)
			: []
	);

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

<Modal bind:open={importando} titulo="Importar tienda" ancho={460} padding={24}>
	<span class="m-ico"><Icon icon={I.phStorefrontFill} size={26} style="color:#2E2A6B" /></span>
	{#if cola}
		<div class="m-t" role="status">
			{cola.estado === 'lista'
				? 'Tu tienda de Tiendanube ya está\u00a0acá'
				: cola.estado === 'fallida'
					? 'La importación se\u00a0cortó'
					: cola.estado === 'cancelada'
						? 'Frenaste la\u00a0importación'
						: 'Importando desde\u00a0Tiendanube'}
		</div>
		<div class="m-d">
			{activa(cola)
				? 'Dejá esta ventana abierta mientras traemos\u00a0todo.'
				: cola.estado === 'lista'
					? 'Revisá precios y stock antes de\u00a0vender.'
					: 'Lo que alcanzó a entrar ya está en tu\u00a0tienda.'}
		</div>
		<div
			class="m-barra"
			role="progressbar"
			aria-label="Avance de la importación"
			aria-valuenow={avanceCola}
			aria-valuemax={100}
		>
			<span style="width:{avanceCola}%"></span>
		</div>
		<ul class="m-etapas">
			{#each ETAPAS_COLA as [k, nombre] (k)}
				<li
					class="chip-estado"
					style:--chip-bg={cola.etapa === k && activa(cola) ? 'rgba(210,88,47,.13)' : 'var(--chip)'}
					style:--chip-fg={cola.etapa === k && activa(cola) ? '#C24E25' : 'var(--mu)'}
				>
					{nombre}
					{cola.progreso?.[k] ?? 0}
				</li>
			{/each}
		</ul>
		{#if cola.progreso?.repetidos || cola.progreso?.recortados}
			<div class="m-d">
				{[
					cola.progreso.repetidos && `${cola.progreso.repetidos} ya estaban`,
					cola.progreso.recortados && `${cola.progreso.recortados} recortados`
				]
					.filter(Boolean)
					.join(' · ')}
			</div>
		{/if}
		{#if cola.errores?.length}
			<ul class="m-notas m-err" role="alert">
				{#each cola.errores.slice(-3) as e, i (i)}<li>{e.detalle}</li>{/each}
			</ul>
		{/if}
		<div class="m-acc">
			{#if activa(cola)}
				<button type="button" class="btn btn-secondary" onclick={cancelarCola}>Frenar</button>
			{:else}
				<button type="button" class="btn btn-dark" onclick={cerrarCola}>Listo</button>
			{/if}
		</div>
	{:else if resumen}
		<div class="m-t" role="status">
			{resumen.creados}
			{resumen.creados === 1 ? 'producto importado' : 'productos importados'} de&nbsp;{NOMBRE[
				resumen.fuente
			]}
		</div>
		{#if notas.length}
			<ul class="m-notas">
				{#each notas as n (n)}<li>{n}</li>{/each}
			</ul>
		{:else}
			<div class="m-d">Con fotos, precios, variantes y&nbsp;categorías.</div>
		{/if}
		<div class="m-acc">
			<button type="button" class="b-pri" onclick={() => (importando = false)}>Listo</button>
		</div>
	{:else}
		<div class="m-t">Traé tus productos de otra&nbsp;tienda</div>
		<div class="m-d">
			Pegá el link y traemos cada producto con fotos, precio, variantes, categoría y descripción. Si
			importás dos veces, no se&nbsp;duplican.
		</div>
		{#if avance}
			<div class="m-c">
				{avance.hechas
					? avance.total
						? `Importando ${avance.hechas} de ${avance.total}…`
						: `Importando ${avance.hechas}…`
					: 'Leyendo tu tienda…'}
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
			{#if tiendanube === false}
				<button type="button" class="btn btn-secondary btn-block m-tn" disabled
					>Tiendanube no está disponible</button
				>
			{:else}
				<a
					class="btn btn-secondary btn-block m-tn"
					href={resolve('/panel/importar/tiendanube')}
					data-sveltekit-reload>Conectar Tiendanube</a
				>
			{/if}
			{#if avisoTN}<div class="m-err" role="alert">{avisoTN}</div>{/if}
			<div class="m-d m-tn-d">Con Tiendanube traemos también tus clientes y&nbsp;pedidos.</div>
			<div class="m-c">O pegá el link de otra plataforma</div>
			<div class="m-seg">
				<Segmentado opciones={FUENTES} bind:value={fuente} label="Plataforma de tu tienda" />
			</div>
			<label class="m-c" for="tienda-imp">Link de tu tienda</label>
			<input
				id="tienda-imp"
				class="m-in"
				bind:value={tiendaImp}
				placeholder={EJEMPLO[fuente]}
				autocomplete="url"
				onkeydown={(e) => e.key === 'Enter' && tiendaImp.trim() && importarTienda()}
			/>
			{#if errorImp}<div class="m-err" role="alert">{errorImp}</div>{/if}
			<div class="m-acc">
				<button type="button" class="b-sec" onclick={() => (importando = false)}>Cancelar</button>
				<button type="button" class="b-pri" disabled={!tiendaImp.trim()} onclick={importarTienda}
					>Importar</button
				>
			</div>
		{/if}
	{/if}
</Modal>

<MotorLista {config} lista={data.lista} {abrir}>
	{#snippet extra()}
		<button type="button" class="secundario" onclick={() => archivo?.click()}
			><Icon icon={I.phFileXlsBold} size={15} />Importar Excel</button
		>
		<button type="button" class="secundario" onclick={abrirImportar}
			><Icon icon={I.phLinkBold} size={15} />Importar tienda</button
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
	.m-seg {
		margin-top: 6px;
	}
	.m-tn {
		margin-top: 16px;
	}
	.m-tn-d {
		font-size: 13px;
	}
	.m-etapas {
		margin: 12px 0 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.m-seg :global(.seg) {
		display: flex;
	}
	/* Las tres plataformas entran a 375 px: menos aire a los costados de cada una. */
	.m-seg :global(.seg button) {
		flex: auto;
		padding-inline: 6px;
	}
	.m-notas {
		margin: 10px 0 0;
		padding-left: 18px;
		font-size: 14px;
		line-height: 1.55;
		color: var(--mu);
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
