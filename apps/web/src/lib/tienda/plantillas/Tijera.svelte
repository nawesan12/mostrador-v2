<script lang="ts">
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { carritoDe } from '../carrito.svelte.ts';
	import { cotizar, cpValido } from '../envio.ts';
	import { partirTitulo } from '../titulo.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../tipos.ts';

	/**
	 * Plantilla Tijera (design/Plantilla Tijera.dc.html), con los datos de la tienda en lugar de los
	 * de ejemplo. Lo que el diseño muestra sin datos atrás (turnos, equipo, pase mensual, seña,
	 * horarios, cupón, cuenta regresiva, seguidores) no se inventa: esas partes no se dibujan.
	 */
	type Pagina = 'inicio' | 'catalogo' | 'producto' | 'carrito' | 'checkout' | 'nosotros' | 'listo';
	let {
		tienda,
		pagina,
		producto = null,
		categoria = null,
		orden = 'vendidos',
		pedido = null,
		error = null
	}: {
		tienda: TiendaPublica;
		pagina: Pagina;
		producto?: ProductoPublico | null;
		categoria?: string | null;
		orden?: string;
		pedido?: {
			numero: number;
			total: number;
			pago: string;
			nombre: string;
			alias: string | null;
		} | null;
		error?: string | null;
	} = $props();

	const base = $derived(`/t/${tienda.slug}`);
	const carrito = $derived(carritoDe(tienda.slug));
	let ancho = $state(1200);
	const nar = $derived(ancho < 820);
	const wide = $derived(!nar);

	const fmt = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
	const img = (f: string | null, w: number, h?: number) =>
		!f
			? null
			: f.startsWith('https://')
				? f
				: `https://images.unsplash.com/photo-${f}?auto=format&fit=crop&w=${w}${h ? '&h=' + h : ''}&q=72`;
	const porId = $derived(new Map(tienda.productos.map((p) => [p.id, p])));
	const tarjeta = (p: ProductoPublico) => ({
		id: p.id,
		b: p.categoria,
		n: p.nombre,
		price: fmt(p.precio),
		old: p.anterior ? fmt(p.anterior) : null,
		img: img(p.foto, 600, 600),
		tag: p.stock === 0 ? 'Sin stock' : p.etiqueta,
		transf: fmt(p.precio * (1 - tienda.cobros.transferencia / 100)),
		stock: p.stock,
		href: `${base}/producto/${p.id}`
	});
	type Tarjeta = ReturnType<typeof tarjeta>;
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// Título de la portada: el segundo tramo va en violeta, como en el diseño ("Tu turno, / en 30 segundos.").
	const titulo = $derived(partirTitulo(tienda.portada.titulo));

	const cats = $derived(tienda.categorias.slice(0, 6));
	const best = $derived(
		tienda.masVendidos.slice(0, 4).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
	);

	// Catálogo: categoría, orden y de a 24.
	const ORDENES = [
		['vendidos', 'Más vendidos'],
		['menor', 'Menor precio'],
		['mayor', 'Mayor precio'],
		['az', 'De la A a la Z']
	] as const;
	const listado = $derived.by(() => {
		const l = tienda.productos.filter((p) => !categoria || p.categoria === categoria);
		const rank = new Map(tienda.masVendidos.map((id, i) => [id, i]));
		const f =
			{
				vendidos: (a: ProductoPublico, b: ProductoPublico) =>
					(rank.get(a.id) ?? 999) - (rank.get(b.id) ?? 999),
				menor: (a: ProductoPublico, b: ProductoPublico) => a.precio - b.precio,
				mayor: (a: ProductoPublico, b: ProductoPublico) => b.precio - a.precio,
				az: (a: ProductoPublico, b: ProductoPublico) => a.nombre.localeCompare(b.nombre, 'es')
			}[orden as 'vendidos'] ?? (() => 0);
		return [...l].sort(f);
	});
	let mostrar = $state(24);
	const catalog = $derived(listado.slice(0, mostrar).map(tarjeta));
	const enlaceCat = (c: string | null, o = orden) => {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, sólo arma el link
		const q = new URLSearchParams();
		if (c) q.set('cat', c);
		if (o !== 'vendidos') q.set('orden', o);
		return `${base}/catalogo${q.size ? '?' + q : ''}`;
	};
	const chipsCat = $derived([
		{ c: null, t: 'Todo', n: tienda.productos.length },
		...tienda.categorias.map((c) => ({ c: c.nombre, t: c.nombre, n: c.cantidad }))
	]);

	// Ficha.
	let abierto = $state<string>('Descripción');
	const acc = $derived(
		producto
			? [
					['Descripción', producto.descripcion || `${producto.nombre}, de ${tienda.nombre}.`],
					[
						'Envíos',
						`Enviamos a todo el país${tienda.envios.retiro ? ' y podés retirarlo en el local' : ''}. Calculá el costo con tu código postal.`
					],
					[
						'Cambios y devoluciones',
						'Tenés 10 días desde que lo recibís para arrepentirte de la compra, sin dar explicaciones (botón de arrepentimiento, abajo de todo).'
					]
				].map(([t, d]) => ({
					t,
					d,
					open: abierto === t,
					i: abierto === t ? I.phMinusBold : I.phPlusBold
				}))
			: []
	);
	const related = $derived(
		producto
			? tienda.productos
					.filter(
						(p) => p.id !== producto!.id && p.categoria === producto!.categoria && p.stock > 0
					)
					.slice(0, 4)
					.map(tarjeta)
			: []
	);
	const resenasProd = $derived(
		producto ? tienda.resenas.filter((r) => r.producto === producto!.nombre) : []
	);
	const promedioProd = $derived(
		resenasProd.length
			? Math.round((resenasProd.reduce((s, r) => s + r.estrellas, 0) / resenasProd.length) * 10) /
					10
			: 0
	);
	let cantidad = $state(1);
	$effect(() => {
		if (producto) cantidad = 1;
	});
	let cp = $state('');
	let cotizado = $state<OpcionEnvio[] | null>(null);
	function calcular() {
		if (cpValido(cp)) cotizado = cotizar(cp, producto?.precio ?? subtotal, tienda);
	}
	function agregar() {
		if (!producto || producto.stock === 0 || tienda.vitrina) return;
		carrito.sumar(producto.id, cantidad, producto.stock);
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(`${base}/carrito`);
	}

	// Carrito.
	const lineas = $derived(carrito.detalle(tienda.productos));
	const subtotal = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));
	const falta = $derived(
		tienda.envios.gratisDesde ? Math.max(0, tienda.envios.gratisDesde - subtotal) : 0
	);
	const upsell = $derived(
		tienda.masVendidos
			.filter((id) => !carrito.lineas.some((l) => l.id === id) && (porId.get(id)?.stock ?? 0) > 0)
			.slice(0, 3)
			.map((id) => porId.get(id)!)
	);

	// Checkout.
	let envio = $state<OpcionEnvio['id'] | null>(null);
	let pago = $state<'transferencia' | 'mercadopago'>('transferencia');
	let enviando = $state(false);
	const opciones = $derived(
		cpValido(cp)
			? cotizar(cp, subtotal, tienda)
			: tienda.envios.retiro
				? cotizar('0000', subtotal, tienda)
				: []
	);
	const elegida = $derived(opciones.find((o) => o.id === envio) ?? null);
	const costoEnvio = $derived(elegida?.precio ?? 0);
	const descuento = $derived(
		pago === 'transferencia' ? Math.round((subtotal * tienda.cobros.transferencia) / 100) : 0
	);
	const total = $derived(subtotal + costoEnvio - descuento);
	$effect(() => {
		if (envio && !opciones.some((o) => o.id === envio)) envio = null;
	});
	const pagos = $derived([
		{
			id: 'transferencia' as const,
			t: 'Transferencia',
			d: `${tienda.cobros.transferencia}% off · te mandamos el alias al confirmar`,
			i: I.phBankFill,
			c: '#2F7D4F'
		},
		...(tienda.cobros.mercadoPago
			? [
					{
						id: 'mercadopago' as const,
						t: 'Mercado Pago',
						d: `Tarjetas en hasta ${tienda.cobros.cuotas} cuotas sin interés, dinero en cuenta`,
						i: I.siMercadopago,
						c: '#00B1EA'
					}
				]
			: [])
	]);

	const foot = $derived([
		{
			t: 'Tienda',
			l: [
				{ t: 'Todo', href: `${base}/catalogo` },
				...tienda.categorias.slice(0, 4).map((c) => ({ t: c.nombre, href: enlaceCat(c.nombre) }))
			]
		},
		{
			t: 'Ayuda',
			l: [
				{ t: 'Envíos', href: `${base}/nosotros` },
				{ t: 'Cambios y devoluciones', href: `/t/${tienda.slug}/arrepentimiento` },
				{ t: 'Medios de pago', href: `${base}/nosotros` }
			]
		},
		{
			t: tienda.nombre,
			l: [
				{ t: 'Nosotros', href: `${base}/nosotros` },
				...(whatsapp ? [{ t: 'WhatsApp', href: wa('¡Hola!') }] : [])
			]
		}
	]);
	const nav = $derived([
		{ t: 'Todo', href: `${base}/catalogo`, on: pagina === 'catalogo' && !categoria },
		...tienda.categorias.slice(0, 3).map((c) => ({
			t: c.nombre,
			href: enlaceCat(c.nombre),
			on: pagina === 'catalogo' && categoria === c.nombre
		})),
		{ t: 'Nosotros', href: `${base}/nosotros`, on: pagina === 'nosotros' }
	]);
	const estrellas = (n: number) =>
		Array.from({ length: 5 }, (_, i) =>
			n >= i + 1 ? I.phStarFill : n >= i + 0.5 ? I.phStarHalfFill : I.phStar
		);
	const sumar = (p: { id: string; stock: number }) => {
		if (p.stock > 0 && !tienda.vitrina) carrito.sumar(p.id, 1, p.stock);
	};
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link
		href="https://fonts.googleapis.com/css2?family=Syne:wght@600;700;800&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet card(p: Tarjeta)}
	<div
		style="padding:10px;border-radius:24px;background:#fff;text-align:left;color:#1E1B2E;box-shadow:0 1px 2px rgba(30,27,46,.06);min-width:0"
	>
		<a
			href={p.href}
			style="position:relative;display:block;aspect-ratio:1;border-radius:18px;overflow:hidden;background:#EAE6EF"
			>{#if p.img}<img
					src={p.img}
					alt={p.n}
					style="width:100%;height:100%;object-fit:cover;display:block"
				/>{/if}{#if p.tag}<span
					style="position:absolute;top:8px;left:8px;padding:4px 9px;border-radius:999px;background:#5B3FD6;color:#fff;font-size:11px;font-weight:700"
					>{p.tag}</span
				>{/if}</a
		>
		<span style="display:block;padding:12px 6px 4px"
			>{#if p.b}<span style="display:block;font-size:12px;color:#8B879A;font-weight:600">{p.b}</span
				>{/if}<a
				href={p.href}
				style="display:block;font-size:15px;font-weight:600;margin-top:2px;color:#1E1B2E;text-decoration:none"
				>{p.n}</a
			><span
				style="display:flex;justify-content:space-between;align-items:center;gap:8px;margin-top:8px"
				><span style="min-width:0"
					><b style="font-size:16px">{p.price}</b>{#if p.old}<s
							style="margin-left:6px;font-size:13px;color:#8B879A">{p.old}</s
						>{/if}<span style="display:block;font-size:12px;color:#2F7D4F;font-weight:600"
						>{p.transf} transf.</span
					></span
				><button
					type="button"
					aria-label="Sumar {p.n} al carrito"
					disabled={p.stock === 0 || tienda.vitrina}
					onclick={() => sumar(p)}
					style="flex:none;width:36px;height:36px;border:0;border-radius:50%;background:#1E1B2E;color:#fff;display:grid;place-items:center;cursor:pointer"
					><Icon icon={I.phPlusBold} size={15} /></button
				></span
			></span
		>
	</div>
{/snippet}

<div
	class="tijera"
	style="min-height:100vh;font-family:'Instrument Sans',system-ui,sans-serif;color:#1E1B2E"
>
	<div
		style="container-type:inline-size;width:100%;background:#F6F4F1;position:relative;overflow-x:clip"
	>
		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:12px;flex-wrap:wrap;padding:9px 16px;background:#1E1B2E;color:#F6F4F1;font-size:13px;text-align:center"
		>
			<span
				>{tienda.envios.gratisDesde
					? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
					: `${tienda.cobros.cuotas} cuotas sin interés`}</span
			>{#if wide}<span style="display:flex;gap:4px"
					><span style="padding:2px 7px;border-radius:6px;background:rgba(246,244,241,.14)"
						>{tienda.cobros.cuotas} cuotas sin interés</span
					><span style="padding:2px 7px;border-radius:6px;background:rgba(246,244,241,.14)"
						>{tienda.cobros.transferencia}% off por transferencia</span
					></span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:center;gap:16px;padding:14px clamp(14px,3.5cqw,44px);background:rgba(246,244,241,.92);backdrop-filter:blur(10px);border-bottom:1px solid rgba(30,27,46,.08)"
		>
			<a
				href={base}
				style="display:flex;align-items:center;gap:10px;min-width:0;text-decoration:none;color:#1E1B2E"
				><span
					style="width:36px;height:36px;border-radius:11px;background:#5B3FD6;color:#fff;display:grid;place-items:center;flex:none"
					><Icon icon={I.phScissorsBold} size={19} /></span
				><span
					style="font-family:'Syne',sans-serif;font-weight:800;font-size:22px;letter-spacing:-.02em;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
					>{tienda.nombre}</span
				></a
			>
			{#if wide}<nav style="display:flex;gap:4px;margin-left:12px">
					{#each nav as n (n.t)}<a
							href={n.href}
							style="text-decoration:none;font:600 14px 'Instrument Sans';padding:9px 14px;border-radius:999px;background:{n.on
								? '#EAE6F2'
								: 'transparent'};color:#1E1B2E;white-space:nowrap">{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			<a
				href="{base}/carrito"
				aria-label="Carrito"
				style="position:relative;color:#1E1B2E;display:grid;flex:none"
				><Icon icon={I.phShoppingBag} size={22} />{#if carrito.unidades}<span
						style="position:absolute;top:-6px;right:-8px;min-width:17px;height:17px;border-radius:999px;background:#5B3FD6;color:#fff;font-size:10.5px;font-weight:700;display:grid;place-items:center"
						>{carrito.unidades}</span
					>{/if}</a
			>
			{#if nar}<a
					href="{base}/catalogo"
					aria-label="Productos"
					style="display:grid;place-items:center;flex:none;width:40px;height:40px;border-radius:999px;background:#1E1B2E;color:#F6F4F1"
					><Icon icon={I.phList} size={16} /></a
				>{/if}
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:10px 16px;background:#EEE9FF;color:#5B3FD6;font-size:13px;font-weight:600;text-align:center"
			>
				Es una tienda de ejemplo de la plantilla Tijera: podés recorrerla, pero no se compra
				de&nbsp;verdad.
			</div>
		{/if}

		<!-- ============ INICIO ============ -->
		{#if pagina === 'inicio'}
			{#if ver('hero')}
				<section
					style="padding:clamp(18px,3cqw,40px) clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,400px),1fr));gap:clamp(14px,2cqw,24px)"
				>
					<div
						style="display:flex;flex-direction:column;justify-content:center;gap:22px;padding:clamp(8px,2cqw,24px) 0;min-width:0"
					>
						{#if tienda.promedio && tienda.resenas.length}<span
								style="align-self:flex-start;display:flex;align-items:center;gap:8px;padding:7px 12px 7px 10px;border-radius:999px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.08);font-size:13px;font-weight:600"
								><Icon
									icon={I.phStarFill}
									size={15}
									style="color:#E3A93B"
								/>{tienda.promedio.toLocaleString('es-AR')} · {tienda.resenas.length}
								{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}</span
							>{/if}
						<h1
							style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(42px,5.6cqw,92px);line-height:.92;letter-spacing:-.04em;overflow-wrap:anywhere"
						>
							{titulo[0]}{#if titulo[1]}<br /><span style="color:#5B3FD6">{titulo[1]}</span>{/if}
						</h1>
						{#if tienda.portada.lema}<p
								style="margin:0;font-size:17px;line-height:1.5;color:#5A566B;max-width:460px"
							>
								{tienda.portada.lema}
							</p>{/if}
						<div style="display:flex;gap:10px;flex-wrap:wrap">
							<a
								href="{base}/catalogo"
								style="display:flex;align-items:center;gap:8px;padding:16px 26px;border:0;border-radius:999px;background:#5B3FD6;color:#fff;font:700 15px 'Instrument Sans';text-decoration:none;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
								><Icon icon={I.phShoppingBagFill} size={17} />Ver productos</a
							>{#if whatsapp}<a
									href={wa('¡Hola!')}
									target="_blank"
									rel="noopener"
									style="padding:16px 22px;border:1.5px solid rgba(30,27,46,.18);border-radius:999px;background:transparent;color:#1E1B2E;font:700 15px 'Instrument Sans';text-decoration:none"
									>Escribinos</a
								>{/if}
						</div>
					</div>
					<div
						style="position:relative;min-height:min(560px,120cqw);border-radius:32px;overflow:hidden;background:#CFC8DA"
					>
						{#if tienda.portada.foto}<img
								src={img(tienda.portada.foto, 1100)}
								alt=""
								style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
							/>{/if}
					</div>
				</section>
			{/if}

			{#if ver('cats') && cats.length}
				<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0">
					<div
						style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
					>
						<h2
							style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
						>
							Categorías
						</h2>
						<span style="font-size:14px;color:#5A566B"
							>{tienda.cobros.cuotas} cuotas sin interés · {tienda.cobros.transferencia}% off por
							transferencia</span
						>
					</div>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(100%,300px),1fr));gap:12px"
					>
						{#each cats as c (c.nombre)}<a
								href={enlaceCat(c.nombre)}
								class="sube"
								style="display:flex;gap:14px;align-items:center;padding:14px;border-radius:22px;background:#fff;text-align:left;text-decoration:none;color:#1E1B2E;box-shadow:0 1px 2px rgba(30,27,46,.06);transition:transform .2s"
								>{#if c.foto}<img
										src={img(c.foto, 200, 200)}
										alt=""
										style="width:76px;height:76px;border-radius:16px;object-fit:cover;flex:none;background:#E4E1EA"
									/>{:else}<span
										style="width:76px;height:76px;border-radius:16px;flex:none;background:#E4E1EA"
									></span>{/if}<span style="flex:1;min-width:0"
									><span style="display:block;font-size:16px;font-weight:700">{c.nombre}</span><span
										style="display:block;font-size:13px;color:#5A566B;margin-top:3px"
										>{c.cantidad} {c.cantidad === 1 ? 'producto' : 'productos'}</span
									></span
								><Icon icon={I.phPlusCircleFill} size={28} style="color:#5B3FD6;flex:none" /></a
							>{/each}
					</div>
				</section>
			{/if}

			{#if ver('prods') && best.length}
				<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0">
					<div
						style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
					>
						<h2
							style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
						>
							Lo más elegido
						</h2>
						<a
							href="{base}/catalogo"
							style="display:flex;align-items:center;gap:8px;padding:11px 16px;border-radius:999px;background:#fff;font-size:13.5px;font-weight:700;color:#1E1B2E;text-decoration:none;box-shadow:0 1px 2px rgba(30,27,46,.08)"
							>Ver todo</a
						>
					</div>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:12px"
					>
						{#each best as p (p.id)}{@render card(p)}{/each}
					</div>
				</section>
			{/if}

			{#if ver('revs') && tienda.resenas.length}
				<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0">
					<div
						style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
					>
						<h2
							style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
						>
							Lo que dicen
						</h2>
						{#if tienda.promedio}<span
								style="display:flex;align-items:center;gap:6px;font-size:14px"
								><Icon icon={I.phStarFill} size={16} style="color:#E3A93B" /><b
									>{tienda.promedio.toLocaleString('es-AR')}</b
								><span style="color:#5A566B"
									>· {tienda.resenas.length}
									{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}</span
								></span
							>{/if}
					</div>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,290px),1fr));gap:12px"
					>
						{#each tienda.resenas.slice(0, 3) as r, i (i)}
							{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
							<div
								style="padding:12px;border-radius:24px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
							>
								{#if pr?.foto}<div
										style="aspect-ratio:4/3;border-radius:16px;overflow:hidden;background:#E4E1EA"
									>
										<img
											src={img(pr.foto, 600, 450)}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>
									</div>{/if}
								<div style="padding:12px 4px 4px">
									<div style="display:flex;gap:1px;color:#E3A93B">
										{#each estrellas(r.estrellas) as e, j (j)}<Icon icon={e} size={14} />{/each}
									</div>
									<p style="margin:8px 0 0;font-size:15px;line-height:1.45">“{r.texto}”</p>
									<div style="margin-top:8px;font-size:12.5px;color:#5A566B">
										{r.autor} · {r.producto}
									</div>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			{#if ver('ig') && tienda.fotos.length >= 3}
				<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px)">
					<h2
						style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
					>
						Así trabajamos
					</h2>
					<div
						style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,160px),1fr));gap:8px"
					>
						{#each tienda.fotos as g (g)}<div
								style="aspect-ratio:1;border-radius:16px;overflow:hidden;background:#E4E1EA"
							>
								<img
									src={img(g, 400, 400)}
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>{/each}
					</div>
				</section>
			{:else}
				<div style="height:clamp(48px,7cqw,96px)"></div>
			{/if}
		{/if}

		<!-- ============ CATÁLOGO ============ -->
		{#if pagina === 'catalogo'}
			<section style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
				<h1
					style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,5cqw,68px);line-height:.95;letter-spacing:-.04em;overflow-wrap:anywhere"
				>
					{categoria ?? 'Todos los productos'}
				</h1>
				<p style="margin:10px 0 0;font-size:15.5px;color:#5A566B;max-width:560px">
					{listado.length}
					{listado.length === 1 ? 'producto' : 'productos'} · {tienda.cobros.cuotas} cuotas sin interés
					o {tienda.cobros.transferencia}% off por&nbsp;transferencia.
				</p>
				<div
					style="margin-top:20px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none;padding:2px"
				>
					{#each chipsCat as c (c.c ?? '')}{@const on = c.c === categoria}<a
							href={enlaceCat(c.c)}
							style="flex:none;text-decoration:none;padding:9px 15px;border-radius:999px;background:{on
								? '#1E1B2E'
								: '#fff'};color:{on
								? '#fff'
								: '#1E1B2E'};font-size:13.5px;font-weight:600;white-space:nowrap;box-shadow:0 1px 2px rgba(30,27,46,.06)"
							>{c.t}<span style="opacity:.55;margin-left:5px">{c.n}</span></a
						>{/each}
				</div>
				<div
					style="margin-top:14px;display:flex;justify-content:flex-end;align-items:center;gap:10px;flex-wrap:wrap;font-size:13.5px"
				>
					<label style="position:relative;display:flex;align-items:center;gap:6px;font-weight:700"
						><span class="oculto">Ordenar</span><select
							value={orden}
							onchange={(e) => goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
							style="appearance:none;border:0;background:transparent;padding:0 18px 0 0;font:700 13.5px 'Instrument Sans';color:#1E1B2E;cursor:pointer"
							>{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
						><Icon
							icon={I.phCaretDownBold}
							size={12}
							style="position:absolute;right:0;pointer-events:none"
						/></label
					>
				</div>
				{#if catalog.length}
					<div
						style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:12px"
					>
						{#each catalog as p (p.id)}{@render card(p)}{/each}
					</div>
					{#if listado.length > mostrar}
						<div style="margin-top:32px;display:flex;justify-content:center">
							<button
								type="button"
								onclick={() => (mostrar += 24)}
								style="padding:15px 26px;border:1.5px solid rgba(30,27,46,.18);border-radius:999px;background:transparent;color:#1E1B2E;font:700 15px 'Instrument Sans';cursor:pointer"
								>Ver más</button
							>
						</div>
					{/if}
				{:else}
					<p style="margin:40px 0;font-size:15.5px;color:#5A566B">
						Todavía no hay productos en esta categoría.
					</p>
				{/if}
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto}
			<section
				style="padding:24px clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));gap:clamp(18px,4cqw,56px);align-items:start"
			>
				<div
					style="position:relative;aspect-ratio:1;border-radius:28px;overflow:hidden;background:#EAE6EF"
				>
					{#if producto.foto}<img
							src={img(producto.foto, 1000)}
							alt={producto.nombre}
							style="width:100%;height:100%;object-fit:cover"
						/>{/if}{#if producto.anterior || producto.etiqueta}<span
							style="position:absolute;top:14px;left:14px;padding:4px 9px;border-radius:999px;background:#5B3FD6;color:#fff;font-size:11px;font-weight:700"
							>{producto.anterior
								? `−${Math.round((1 - producto.precio / producto.anterior) * 100)}%`
								: producto.etiqueta}</span
						>{/if}
				</div>
				<div style="position:sticky;top:90px;min-width:0">
					{#if producto.categoria}<a
							href={enlaceCat(producto.categoria)}
							style="font-size:13px;font-weight:700;color:#5B3FD6;text-decoration:none"
							>{producto.categoria}</a
						>{/if}
					<h1
						style="margin:6px 0 0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,4.4cqw,54px);line-height:.98;letter-spacing:-.035em;overflow-wrap:anywhere"
					>
						{producto.nombre}
					</h1>
					{#if resenasProd.length}<div
							style="margin-top:10px;display:flex;align-items:center;gap:8px;font-size:13.5px"
						>
							<span style="display:flex;color:#E3A93B"
								>{#each estrellas(promedioProd) as e, i (i)}<Icon icon={e} size={14} />{/each}</span
							><b>{promedioProd.toLocaleString('es-AR')}</b><span style="color:#5A566B"
								>{resenasProd.length} {resenasProd.length === 1 ? 'reseña' : 'reseñas'}</span
							>
						</div>{/if}
					<div style="margin-top:18px;display:flex;align-items:baseline;gap:10px">
						<b style="font-size:30px;letter-spacing:-.02em">{fmt(producto.precio)}</b
						>{#if producto.anterior}<s style="font-size:16px;color:#8B879A"
								>{fmt(producto.anterior)}</s
							>{/if}
					</div>
					<div style="margin-top:6px;display:flex;flex-direction:column;gap:3px;font-size:14px">
						<span
							>{tienda.cobros.cuotas} cuotas sin interés de
							<b>{fmt(producto.precio / tienda.cobros.cuotas)}</b></span
						><span style="color:#2F7D4F;font-weight:700"
							>{fmt(producto.precio * (1 - tienda.cobros.transferencia / 100))} por transferencia</span
						>
					</div>
					{#if producto.stock > 0 && producto.stock <= 3}<div
							style="margin-top:14px;font-size:13px;font-weight:700;color:#5B3FD6"
						>
							{producto.stock === 1 ? 'Queda 1' : `Quedan ${producto.stock}`}
						</div>{/if}
					<div style="margin-top:20px;display:flex;gap:8px;flex-wrap:wrap">
						{#if producto.stock > 0}<span
								style="display:flex;align-items:center;border-radius:999px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.08)"
								><button
									type="button"
									aria-label="Uno menos"
									disabled={cantidad <= 1}
									onclick={() => (cantidad = Math.max(1, cantidad - 1))}
									style="width:44px;height:48px;display:grid;place-items:center;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
									><Icon icon={I.phMinusBold} size={14} /></button
								><b style="width:22px;text-align:center">{cantidad}</b><button
									type="button"
									aria-label="Uno más"
									disabled={cantidad >= producto.stock}
									onclick={() => (cantidad = Math.min(producto!.stock, cantidad + 1))}
									style="width:44px;height:48px;display:grid;place-items:center;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
									><Icon icon={I.phPlusBold} size={14} /></button
								></span
							>{/if}<button
							type="button"
							onclick={agregar}
							disabled={producto.stock === 0 || tienda.vitrina}
							style="flex:1;min-width:180px;padding:15px;border:0;border-radius:999px;background:#1E1B2E;color:#fff;font:700 15px 'Instrument Sans';cursor:pointer"
							>{producto.stock === 0 ? 'Sin stock' : 'Agregar al carrito'}</button
						>
					</div>
					{#if whatsapp}<a
							href={wa(`¡Hola! Quiero comprar ${producto.nombre} (${fmt(producto.precio)}).`)}
							target="_blank"
							rel="noopener"
							style="margin-top:8px;display:flex;justify-content:center;align-items:center;gap:8px;padding:14px;border-radius:999px;border:1.5px solid rgba(30,27,46,.14);font-size:14.5px;font-weight:700;color:#1E1B2E;text-decoration:none"
							><Icon icon={I.logosWhatsappIcon} size={17} />Consultar por WhatsApp</a
						>{/if}
					{#if tienda.envios.retiro}<div
							style="margin-top:16px;padding:14px;border-radius:20px;background:#EEE9FF;display:flex;gap:12px;align-items:center"
						>
							<Icon icon={I.phStorefrontFill} size={26} style="color:#5B3FD6;flex:none" />
							<div style="flex:1;font-size:13.5px;line-height:1.4">
								<b>Retiralo en el local</b>
								{tienda.direccion ? `en ${tienda.direccion}` : ''} y no pagás&nbsp;envío.
							</div>
						</div>{/if}
					<div
						style="margin-top:12px;padding:14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
					>
						<label for="cp-ficha" style="display:block;font-size:13.5px;font-weight:700"
							>Envío a domicilio</label
						>
						<div style="margin-top:8px;display:flex;gap:6px">
							<input
								id="cp-ficha"
								bind:value={cp}
								inputmode="numeric"
								maxlength="4"
								placeholder="Código postal"
								onkeydown={(e) => e.key === 'Enter' && calcular()}
								style="flex:1;min-width:0;padding:10px 12px;border:0;border-radius:12px;background:#F6F4F1;font:400 14px 'Instrument Sans';color:#1E1B2E"
							/><button
								type="button"
								onclick={calcular}
								style="padding:10px 14px;border:0;border-radius:12px;background:#1E1B2E;color:#fff;font:700 13.5px 'Instrument Sans';cursor:pointer"
								>Calcular</button
							>
						</div>
						{#if cotizado}<div
								style="margin-top:10px;display:flex;flex-direction:column;gap:6px;font-size:13.5px"
							>
								{#each cotizado as o (o.id)}<span
										style="display:flex;justify-content:space-between;gap:10px"
										><span>{o.nombre} · {o.detalle}</span><b
											style="white-space:nowrap;color:{o.precio ? '#1E1B2E' : '#2F7D4F'}"
											>{o.precio ? fmt(o.precio) : 'Gratis'}</b
										></span
									>{/each}
							</div>{/if}
					</div>
					<div
						style="margin-top:12px;padding:4px 14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
					>
						{#each acc as a, i (a.t)}<div
								style="padding:12px 0;{i ? 'border-top:1px solid rgba(30,27,46,.06)' : ''}"
							>
								<button
									type="button"
									onclick={() => (abierto = a.open ? '' : a.t)}
									aria-expanded={a.open}
									style="width:100%;display:flex;justify-content:space-between;align-items:center;gap:10px;padding:0;border:0;background:transparent;font:700 13.5px 'Instrument Sans';color:#1E1B2E;cursor:pointer;text-align:left"
									>{a.t}<Icon icon={a.i} size={13} /></button
								>{#if a.open}<p style="margin:8px 0 0;font-size:14px;line-height:1.6;color:#4A4659">
										{a.d}
									</p>{/if}
							</div>{/each}
					</div>
				</div>
			</section>
			{#if related.length}
				<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,44px)">
					<h2
						style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
					>
						Combinalo con
					</h2>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:12px"
					>
						{#each related as p (p.id)}{@render card(p)}{/each}
					</div>
				</section>
			{:else}
				<div style="height:clamp(48px,7cqw,90px)"></div>
			{/if}
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
				<h1
					style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(40px,6cqw,72px);line-height:.95;letter-spacing:-.04em"
				>
					Tu carrito
				</h1>
				{#if !lineas.length}
					<p style="margin:16px 0 0;font-size:15.5px;color:#5A566B">Todavía no sumaste nada.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-flex;align-items:center;gap:8px;margin-top:18px;padding:16px 26px;border-radius:999px;background:#5B3FD6;color:#fff;font:700 15px 'Instrument Sans';text-decoration:none;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
						><Icon icon={I.phShoppingBagFill} size={17} />Ver productos</a
					>
				{:else}
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(16px,3cqw,36px);align-items:start"
					>
						<div style="display:flex;flex-direction:column;gap:10px;min-width:0">
							{#each lineas as c (c.id)}<div
									style="display:flex;gap:14px;padding:12px;border-radius:22px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
								>
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 200, 200)}
											alt=""
											style="width:88px;height:88px;border-radius:16px;object-fit:cover;flex:none;background:#EAE6EF"
										/>{/if}
									<div style="flex:1;min-width:0;display:flex;flex-direction:column">
										<div style="display:flex;justify-content:space-between;gap:10px">
											<a
												href="{base}/producto/{c.id}"
												style="font-size:15px;font-weight:600;color:#1E1B2E;text-decoration:none"
												>{c.producto.nombre}</a
											><b style="white-space:nowrap">{fmt(c.subtotal)}</b>
										</div>
										<span style="font-size:13px;color:#5A566B">{fmt(c.producto.precio)} c/u</span>
										<span style="flex:1;min-height:10px"></span>
										<div style="display:flex;justify-content:space-between;align-items:center">
											<span
												style="display:flex;align-items:center;gap:4px;padding:0 4px;border-radius:999px;background:#F6F4F1"
												><button
													type="button"
													aria-label="Uno menos"
													onclick={() => carrito.fijar(c.id, c.cantidad - 1)}
													style="width:30px;height:30px;display:grid;place-items:center;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
													><Icon icon={I.phMinusBold} size={12} /></button
												><b style="font-size:14px;min-width:16px;text-align:center">{c.cantidad}</b
												><button
													type="button"
													aria-label="Uno más"
													disabled={c.cantidad >= c.producto.stock}
													onclick={() => carrito.fijar(c.id, c.cantidad + 1)}
													style="width:30px;height:30px;display:grid;place-items:center;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
													><Icon icon={I.phPlusBold} size={12} /></button
												></span
											><button
												type="button"
												aria-label="Quitar {c.producto.nombre}"
												onclick={() => carrito.fijar(c.id, 0)}
												style="display:grid;padding:4px;border:0;background:transparent;color:#8B879A;cursor:pointer"
												><Icon icon={I.phTrash} size={18} /></button
											>
										</div>
									</div>
								</div>{/each}
							{#if tienda.envios.gratisDesde}
								<div
									style="padding:16px;border-radius:22px;background:#5B3FD6;color:#fff;display:flex;gap:14px;align-items:center;flex-wrap:wrap"
								>
									<Icon icon={I.phTruck} size={30} />
									<div style="flex:1;min-width:200px">
										<div style="font-size:15px;font-weight:700">
											{#if falta}Te faltan {fmt(falta)} para el envío gratis{:else}Tenés envío
												gratis{/if}
										</div>
										<div
											style="margin-top:8px;height:6px;border-radius:999px;background:rgba(255,255,255,.25);overflow:hidden"
										>
											<div
												style="width:{Math.min(
													100,
													(subtotal / tienda.envios.gratisDesde) * 100
												)}%;height:100%;background:#fff"
											></div>
										</div>
									</div>
								</div>
							{/if}
							{#if upsell.length}
								<div style="margin-top:10px;font-size:15px;font-weight:700">Combinalo con</div>
								<div style="display:flex;gap:10px;overflow-x:auto;scrollbar-width:none">
									{#each upsell as p (p.id)}<div
											style="flex:none;width:160px;padding:8px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
										>
											<a
												href="{base}/producto/{p.id}"
												style="display:block;aspect-ratio:1;border-radius:14px;overflow:hidden;background:#EAE6EF"
												>{#if p.foto}<img
														src={img(p.foto, 300, 300)}
														alt=""
														style="width:100%;height:100%;object-fit:cover"
													/>{/if}</a
											>
											<div style="padding:8px 4px 2px;font-size:13.5px;font-weight:600">
												{p.nombre}
											</div>
											<div
												style="padding:0 4px;display:flex;justify-content:space-between;align-items:center"
											>
												<b style="font-size:14px">{fmt(p.precio)}</b><button
													type="button"
													aria-label="Sumar {p.nombre}"
													onclick={() => carrito.sumar(p.id, 1, p.stock)}
													style="display:grid;padding:0;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
													><Icon icon={I.phPlusCircleFill} size={26} /></button
												>
											</div>
										</div>{/each}
								</div>
							{/if}
						</div>
						<div
							style="position:sticky;top:90px;padding:22px;border-radius:28px;background:#fff;box-shadow:0 20px 40px -24px rgba(30,27,46,.25);display:flex;flex-direction:column;gap:11px;font-size:14px"
						>
							<div style="font-family:'Syne',sans-serif;font-weight:700;font-size:24px">
								Resumen
							</div>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							>
							<span style="display:flex;justify-content:space-between;gap:10px;color:#5A566B"
								><span>Envío</span><span>Se calcula en el próximo paso</span></span
							>
							<div style="height:1px;background:rgba(30,27,46,.08)"></div>
							<span style="display:flex;justify-content:space-between;align-items:baseline"
								><b>Total</b><b style="font-size:24px">{fmt(subtotal)}</b></span
							>
							<span style="color:#5A566B"
								>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
									subtotal / tienda.cobros.cuotas
								)}</span
							>
							<span style="color:#2F7D4F;font-weight:700"
								>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
							>
							<a
								href="{base}/checkout"
								style="margin-top:4px;padding:16px;border-radius:999px;background:#1E1B2E;color:#fff;font:700 15px 'Instrument Sans';text-align:center;text-decoration:none"
								>Ir a pagar</a
							>
						</div>
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ CHECKOUT ============ -->
		{#if pagina === 'checkout'}
			{#if !lineas.length}
				<section style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
					<p style="margin:0;font-size:15.5px;color:#5A566B">Tu carrito está vacío.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-block;margin-top:16px;color:#5B3FD6;font-weight:700"
						>Ver productos</a
					>
				</section>
			{:else}
				<form
					method="POST"
					use:enhance={() => {
						enviando = true;
						return async ({ result, update }) => {
							enviando = false;
							// Mercado Pago es otro sitio: se navega de verdad.
							if (result.type === 'redirect' && result.location.startsWith('https://'))
								return void (location.href = result.location);
							if (result.type === 'redirect') carrito.vaciar();
							await update();
						};
					}}
					style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(16px,3cqw,40px);align-items:start"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<div style="display:flex;flex-direction:column;gap:22px;min-width:0">
						<h1
							style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,5cqw,60px);line-height:.95;letter-spacing:-.04em"
						>
							Confirmá y pagá
						</h1>
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">Tus datos</div>
							<div
								style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:8px"
							>
								{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'Teléfono', 'tel', 'tel'], ['dni', 'DNI', 'text', 'off']] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(30,27,46,.1)"
										><span style="display:block;font-size:11.5px;color:#8B879A">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required={n !== 'dni'}
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1E1B2E;outline:none"
										/></label
									>
								{/each}
							</div>
						</div>
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">Entrega</div>
							<div style="display:grid;grid-template-columns:minmax(0,130px) minmax(0,1fr);gap:8px">
								<label
									style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(30,27,46,.1)"
									><span style="display:block;font-size:11.5px;color:#8B879A">Código postal</span
									><input
										name="cp"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1E1B2E;outline:none"
									/></label
								>
								<label
									style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(30,27,46,.1)"
									><span style="display:block;font-size:11.5px;color:#8B879A"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required={envio !== 'retiro'}
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1E1B2E;outline:none"
									/></label
								>
							</div>
							<div
								style="margin-top:8px;display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Forma de entrega"
							>
								{#each opciones as s (s.id)}{@const on = envio === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (envio = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border-radius:18px;border:1.5px solid {on
											? '#5B3FD6'
											: 'rgba(30,27,46,.1)'};background:{on
											? '#F3F0FF'
											: '#fff'};text-align:left;cursor:pointer;font-family:inherit;color:#1E1B2E"
										><span
											style="width:20px;height:20px;border-radius:50%;border:2px solid {on
												? '#5B3FD6'
												: 'rgba(30,27,46,.25)'};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:10px;height:10px;border-radius:50%;background:#5B3FD6"
												></span>{/if}</span
										><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:#5A566B">{s.detalle}</span
											></span
										><b
											style="font-size:14px;white-space:nowrap;color:{s.precio
												? '#1E1B2E'
												: '#2F7D4F'}">{s.precio ? fmt(s.precio) : 'Gratis'}</b
										></button
									>{:else}<span style="font-size:13.5px;color:#5A566B"
										>Poné tu código postal para ver cómo te llega.</span
									>{/each}
							</div>
						</div>
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">Pago</div>
							<div
								style="display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Forma de pago"
							>
								{#each pagos as s (s.id)}{@const on = pago === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (pago = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border-radius:18px;border:1.5px solid {on
											? '#5B3FD6'
											: 'rgba(30,27,46,.1)'};background:{on
											? '#F3F0FF'
											: '#fff'};text-align:left;cursor:pointer;font-family:inherit;color:#1E1B2E"
										><span
											style="width:20px;height:20px;border-radius:50%;border:2px solid {on
												? '#5B3FD6'
												: 'rgba(30,27,46,.25)'};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:10px;height:10px;border-radius:50%;background:#5B3FD6"
												></span>{/if}</span
										><Icon icon={s.i} size={22} style="color:{s.c};flex:none" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.t}</span
											><span style="display:block;font-size:12.5px;color:#5A566B">{s.d}</span></span
										></button
									>{/each}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;border-radius:16px;background:#FDECEC;color:#B3261E;font-size:14px"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || enviando || tienda.vitrina}
							style="padding:17px;border:0;border-radius:999px;background:#5B3FD6;color:#fff;font:700 15.5px 'Instrument Sans';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
							><Icon icon={I.phLockSimpleFill} size={16} />{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: envio
									? `Pagar ${fmt(total)}`
									: 'Elegí cómo te llega'}</button
						>
					</div>
					<div style="position:sticky;top:90px;display:flex;flex-direction:column;gap:10px">
						<div
							style="padding:18px;border-radius:24px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
						>
							<div
								style="display:flex;align-items:center;gap:8px;font-size:12.5px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#5A566B"
							>
								<Icon icon={I.phShoppingBagFill} size={15} />Tu pedido
							</div>
							<div style="margin-top:10px;display:flex;flex-direction:column;gap:10px">
								{#each lineas as c (c.id)}<div style="display:flex;gap:10px;align-items:center">
										{#if c.producto.foto}<img
												src={img(c.producto.foto, 120, 120)}
												alt=""
												style="width:46px;height:46px;border-radius:12px;object-fit:cover;flex:none;background:#EAE6EF"
											/>{/if}<span style="flex:1;min-width:0;font-size:14px"
											>{c.producto.nombre}{#if c.cantidad > 1}<span style="color:#8B879A"
													>&nbsp;× {c.cantidad}</span
												>{/if}</span
										><b style="font-size:14px;white-space:nowrap">{fmt(c.subtotal)}</b>
									</div>{/each}
							</div>
							<div
								style="margin-top:14px;padding-top:12px;border-top:1px solid rgba(30,27,46,.08);display:flex;flex-direction:column;gap:8px;font-size:14px"
							>
								<span style="display:flex;justify-content:space-between"
									><span>Subtotal</span><span>{fmt(subtotal)}</span></span
								><span style="display:flex;justify-content:space-between"
									><span>Envío</span><span
										>{elegida ? (elegida.precio ? fmt(elegida.precio) : 'Gratis') : '—'}</span
									></span
								>{#if descuento}<span
										style="display:flex;justify-content:space-between;color:#2F7D4F;font-weight:700"
										><span>Transferencia −{tienda.cobros.transferencia}%</span><span
											>− {fmt(descuento)}</span
										></span
									>{/if}
							</div>
						</div>
						<div
							style="display:flex;justify-content:space-between;align-items:baseline;padding:0 6px"
						>
							<b>Pagás hoy</b><b style="font-size:24px">{fmt(total)}</b>
						</div>
						<div
							style="padding:12px;border-radius:16px;background:#1E1B2E;color:#F6F4F1;font-size:13px;line-height:1.45;display:flex;gap:10px;align-items:center"
						>
							<Icon icon={I.phShieldCheckFill} size={20} style="color:#A996FF;flex:none" />Compra
							protegida · 10 días para arrepentirte
						</div>
					</div>
				</form>
			{/if}
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas de "Tu reserva" y el checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(14px,3.5cqw,44px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:#5B3FD6;color:#fff;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
					><Icon icon={I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,5cqw,60px);line-height:.95;letter-spacing:-.04em;overflow-wrap:anywhere"
				>
					¡Gracias, {pedido.nombre}!
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:#5A566B">
					Recibimos tu pedido <b style="color:#1E1B2E">#{pedido.numero}</b> por
					<b style="color:#1E1B2E">{fmt(pedido.total)}</b>. Te escribimos por mail con
					cada&nbsp;novedad.
				</p>
				{#if pedido.pago === 'transferencia'}
					<div
						style="margin-top:24px;padding:22px;border-radius:28px;background:#1E1B2E;color:#F6F4F1;text-align:left;font-size:14.5px;line-height:1.6"
					>
						<b>Transferí {fmt(pedido.total)}</b>
						{#if pedido.alias}al alias <b style="color:#A996FF">{pedido.alias}</b>{:else}con los
							datos que te mandamos por mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
								target="_blank"
								rel="noopener"
								style="color:#A996FF">por WhatsApp</a
							>{/if}. Lo preparamos apenas se acredite.
					</div>
				{/if}
				<a
					href={base}
					style="display:inline-block;margin-top:28px;padding:16px 26px;border-radius:999px;background:#1E1B2E;color:#fff;font:700 15px 'Instrument Sans';text-decoration:none"
					>Volver a la tienda</a
				>
			</section>
		{/if}

		<!-- ============ NOSOTROS ============ -->
		{#if pagina === 'nosotros'}
			<section
				style="padding:clamp(18px,3cqw,40px) clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(14px,2cqw,24px);align-items:stretch"
			>
				<div
					style="position:relative;min-height:420px;border-radius:32px;overflow:hidden;background:#CFC8DA"
				>
					{#if tienda.portada.foto}<img
							src={img(tienda.portada.foto, 1100)}
							alt=""
							style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
						/>{/if}
				</div>
				<div
					style="display:flex;flex-direction:column;justify-content:center;gap:18px;padding:clamp(8px,2cqw,24px) 0;min-width:0"
				>
					<h1
						style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,4.6cqw,66px);line-height:.94;letter-spacing:-.04em;overflow-wrap:anywhere"
					>
						{titulo[0]}{#if titulo[1]}<span style="color:#5B3FD6">{` ${titulo[1]}`}</span>{/if}
					</h1>
					<p style="margin:0;font-size:16px;line-height:1.6;color:#5A566B;max-width:500px">
						{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. Comprás online y te lo
						mandamos a todo el país{tienda.envios.retiro ? ', o lo retirás en el local' : ''}.
					</p>
					<div style="display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:10px">
						<div
							style="padding:14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
						>
							<div
								style="font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(26px,3cqw,36px);letter-spacing:-.03em"
							>
								{tienda.productos.length}
							</div>
							<div style="font-size:12.5px;color:#5A566B;margin-top:2px">productos</div>
						</div>
						<div
							style="padding:14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
						>
							<div
								style="font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(26px,3cqw,36px);letter-spacing:-.03em"
							>
								{tienda.cobros.cuotas}
							</div>
							<div style="font-size:12.5px;color:#5A566B;margin-top:2px">cuotas sin interés</div>
						</div>
						{#if tienda.promedio}<div
								style="padding:14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
							>
								<div
									style="font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(26px,3cqw,36px);letter-spacing:-.03em"
								>
									{tienda.promedio.toLocaleString('es-AR')}
								</div>
								<div style="font-size:12.5px;color:#5A566B;margin-top:2px">
									de promedio en reseñas
								</div>
							</div>{/if}
					</div>
				</div>
			</section>
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(14px,3.5cqw,44px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:12px"
			>
				{#if tienda.direccion || tienda.ciudad}
					<div
						style="padding:22px;border-radius:28px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06);display:flex;flex-direction:column;gap:14px"
					>
						<div style="font-family:'Syne',sans-serif;font-weight:700;font-size:24px">
							Dónde estamos
						</div>
						<div style="display:flex;flex-direction:column;gap:10px;font-size:14.5px">
							<span style="display:flex;gap:10px"
								><Icon
									icon={I.phMapPinFill}
									size={18}
									style="color:#5B3FD6;flex:none"
								/>{tienda.direccion || tienda.ciudad}</span
							>{#if tienda.envios.retiro}<span style="display:flex;gap:10px"
									><Icon
										icon={I.phStorefrontFill}
										size={18}
										style="color:#5B3FD6;flex:none"
									/>Retiro en el local sin costo</span
								>{/if}
						</div>
					</div>
				{/if}
				<div
					style="padding:22px;border-radius:28px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
				>
					<div style="font-family:'Syne',sans-serif;font-weight:700;font-size:24px">
						Envíos y pagos
					</div>
					<div style="margin-top:12px;display:flex;flex-direction:column">
						{#each [...(tienda.envios.gratisDesde ? [['Envío gratis', `Desde ${fmt(tienda.envios.gratisDesde)}`]] : []), ['Envíos', 'A todo el país'], ['Tarjetas', `${tienda.cobros.cuotas} cuotas sin interés`], ['Transferencia', `${tienda.cobros.transferencia}% off`]] as [d, h] (d)}<div
								style="display:flex;justify-content:space-between;gap:10px;padding:11px 0;border-bottom:1px solid rgba(30,27,46,.06);font-size:14.5px;font-weight:500"
							>
								<span>{d}</span><span style="text-align:right">{h}</span>
							</div>{/each}
					</div>
				</div>
				{#if whatsapp}
					<div
						style="padding:22px;border-radius:28px;background:#1E1B2E;color:#F6F4F1;display:flex;flex-direction:column;gap:10px"
					>
						<div style="font-family:'Syne',sans-serif;font-weight:700;font-size:24px">
							Escribinos
						</div>
						<a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="display:flex;align-items:center;gap:10px;padding:14px;border-radius:16px;background:#25D366;color:#fff;font-weight:700;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={22} />{tienda.whatsapp}</a
						>
					</div>
				{/if}
			</section>
		{/if}

		<!-- footer -->
		<footer
			style="padding:clamp(40px,6cqw,72px) clamp(14px,3.5cqw,44px) 26px;background:#1E1B2E;color:#F6F4F1"
		>
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:28px"
			>
				<div style="min-width:0">
					<div
						style="font-family:'Syne',sans-serif;font-weight:800;font-size:34px;letter-spacing:-.02em;overflow-wrap:anywhere"
					>
						{tienda.nombre}
					</div>
					<p style="margin:10px 0 0;font-size:14px;color:#B3AEC6;line-height:1.55;max-width:300px">
						{tienda.cobros.cuotas} cuotas sin interés y {tienda.cobros.transferencia}% off por
						transferencia{tienda.direccion || tienda.ciudad
							? `. ${tienda.direccion || tienda.ciudad}`
							: ''}.
					</p>
				</div>
				{#each foot as f (f.t)}<div
						style="display:flex;flex-direction:column;gap:9px;font-size:14px"
					>
						<b
							style="font-size:12px;letter-spacing:.1em;text-transform:uppercase;color:#A996FF;margin-bottom:4px"
							>{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:#DAD6E6;text-decoration:none"
								>{l.t}</a
							>{/each}
					</div>{/each}
			</div>
			<div
				style="margin-top:32px;padding-top:24px;border-top:1px solid color-mix(in srgb,currentColor 16%,transparent);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,250px),1fr));gap:22px"
			>
				<div style="min-width:0">
					<div
						style="font-size:11.5px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;opacity:.7;margin-bottom:10px"
					>
						Medios de pago
					</div>
					<div class="sellos">
						<span><Icon icon={I.logosVisa} size={34} /></span><span
							><Icon icon={I.logosMastercard} size={22} /></span
						><span><Icon icon={I.logosAmex} size={30} /></span>{#if tienda.cobros.mercadoPago}<span
								><Icon icon={I.siMercadopago} size={17} style="color:#00B1EA" /><span
									style="color:#0A2D6E">Mercado Pago</span
								></span
							>{/if}<span
							><Icon icon={I.phBankFill} size={14} style="color:#2F7A43" />Transferencia</span
						>
					</div>
				</div>
				<div style="min-width:0">
					<div
						style="font-size:11.5px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;opacity:.7;margin-bottom:10px"
					>
						Medios de envío
					</div>
					<div class="sellos">
						<span
							><span style="color:#D71920;font-weight:800;letter-spacing:-.02em">andreani</span
							></span
						>{#if tienda.envios.correo}<span
								><span
									style="width:14px;height:14px;border-radius:3px;background:#FFD100;display:inline-block"
								></span><span style="color:#00205B">Correo Argentino</span></span
							>{/if}{#if tienda.envios.moto}<span
								><Icon icon={I.phMopedFill} size={15} style="color:#1A1A1A" />Moto en el día</span
							>{/if}{#if tienda.envios.retiro}<span
								><Icon icon={I.phStorefrontFill} size={14} style="color:#1A1A1A" />Retiro en el
								local</span
							>{/if}
					</div>
				</div>
				<div style="min-width:0">
					<div
						style="font-size:11.5px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;opacity:.7;margin-bottom:10px"
					>
						Compra segura
					</div>
					<div class="sellos">
						<span
							><Icon icon={I.phLockSimpleFill} size={14} style="color:#2F7A43" />Sitio seguro SSL</span
						><span
							><Icon icon={I.phShieldCheckFill} size={14} style="color:#1A1A1A" />Compra protegida</span
						>
					</div>
				</div>
			</div>
			<div
				style="margin-top:36px;padding-top:16px;border-top:1px solid rgba(246,244,241,.1);display:flex;justify-content:space-between;gap:14px;flex-wrap:wrap;font-size:12.5px;color:#8B879A"
			>
				<span
					>© {new Date().getFullYear()}
					{tienda.nombre}{tienda.ciudad ? ` · ${tienda.ciudad}` : ''}</span
				>
			</div>
			<div
				style="margin-top:16px;padding-top:16px;border-top:1px solid color-mix(in srgb,currentColor 16%,transparent);display:flex;justify-content:space-between;align-items:center;gap:12px 20px;flex-wrap:wrap;font-size:12.5px"
			>
				<span style="display:flex;gap:6px 16px;flex-wrap:wrap"
					><a
						href="/t/{tienda.slug}/arrepentimiento"
						style="color:inherit;text-decoration:underline;text-underline-offset:3px;font-weight:700;display:inline-flex;align-items:center;gap:5px"
						><Icon icon={I.phArrowCounterClockwiseBold} size={13} />Botón de arrepentimiento</a
					><a
						href="/legales/consumidor"
						style="color:inherit;opacity:0.75;text-decoration:underline;text-underline-offset:3px;font-weight:500"
						>Defensa de las y los consumidores</a
					><a
						href="/legales/terminos"
						style="color:inherit;opacity:0.75;text-decoration:underline;text-underline-offset:3px;font-weight:500"
						>Términos y condiciones</a
					><a
						href="/legales/privacidad"
						style="color:inherit;opacity:0.75;text-decoration:underline;text-underline-offset:3px;font-weight:500"
						>Políticas de privacidad</a
					></span
				>
				<a
					href="/"
					style="display:inline-flex;align-items:center;gap:8px;padding:5px 13px 5px 5px;border-radius:999px;background:#fff;color:#2A140C;text-decoration:none;font:500 12.5px 'Instrument Sans',sans-serif;box-shadow:0 1px 3px rgba(0,0,0,.12);white-space:nowrap"
					><span
						style="width:24px;height:24px;border-radius:50%;background:#FCEDE5;display:grid;place-items:center"
						><img src="/mark.svg" alt="" style="width:13px;height:auto;display:block" /></span
					>Tienda creada con <b style="font-weight:800;letter-spacing:-.02em">mostrador</b></a
				>
			</div>
		</footer>

		{#if whatsapp}<a
				href={wa('¡Hola!')}
				target="_blank"
				rel="noopener"
				aria-label="Escribinos por WhatsApp"
				style="position:fixed;right:18px;bottom:18px;z-index:30;width:54px;height:54px;border-radius:50%;background:#25D366;color:#fff;display:grid;place-items:center;box-shadow:0 12px 26px -8px rgba(0,0,0,.4)"
				><Icon icon={I.phWhatsappLogoFill} size={28} /></a
			>{/if}
	</div>
</div>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.sube:hover {
		transform: translateY(-2px);
	}
	.oculto {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
		white-space: nowrap;
	}
	.sellos {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.sellos > span {
		height: 28px;
		display: inline-flex;
		align-items: center;
		gap: 5px;
		padding: 0 9px;
		border-radius: 6px;
		background: #fff;
		color: #1a1a1a;
		box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.06);
		font:
			700 11.5px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	button:disabled {
		opacity: 0.45;
		cursor: not-allowed !important;
	}
</style>
