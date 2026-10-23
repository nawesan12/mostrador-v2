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
	 * Plantilla Mostaza (design/Plantilla Mostaza.dc.html), con los datos de la tienda en lugar de los
	 * de ejemplo. El catálogo es el "Menú y pedido" del diseño (se suma desde la carta y el pedido se
	 * arma al costado). Lo que el diseño muestra sin datos atrás (cuenta regresiva, horarios, turnos de
	 * retiro, propina, mesa con QR, molienda y tamaños, suscripción, club, seguidores) no se dibuja.
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
	// Etiquetas con los colores del diseño ("Más pedido", "Vegano", "Nuevo").
	const tagColor = (p: ProductoPublico) =>
		p.stock === 0
			? ['#3B2416', '#FBF3E4']
			: p.anterior
				? ['#F5C26B', '#2A170C']
				: ['#E8F2E2', '#2F6B3A'];
	const tarjeta = (p: ProductoPublico) => ({
		id: p.id,
		n: p.nombre,
		d: p.descripcion,
		price: fmt(p.precio),
		old: p.anterior ? fmt(p.anterior) : null,
		img: img(p.foto, 500, 500),
		tag: p.stock === 0 ? 'Sin stock' : p.etiqueta,
		tagBg: tagColor(p)[0],
		tagFg: tagColor(p)[1],
		stock: p.stock,
		href: `${base}/producto/${p.id}`
	});
	type Tarjeta = ReturnType<typeof tarjeta>;
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;
	const enCarrito = (id: string) => carrito.lineas.find((l) => l.id === id)?.cantidad ?? 0;
	const sumar = (id: string, stock: number) => {
		if (stock === 0 || tienda.vitrina) return;
		carrito.sumar(id, 1, stock);
	};

	// Título de la portada: el último tramo va en mostaza, como en el diseño ("Café rico, / sin vueltas.").
	const titulo = $derived(partirTitulo(tienda.portada.titulo));

	// Las tres formas de recibirlo del diseño, con los envíos que la tienda tiene de verdad.
	const modes = $derived([
		...(tienda.envios.retiro
			? [
					{
						t: 'Retirá en el local',
						d: tienda.direccion || tienda.ciudad || 'Sin costo',
						i: I.phStorefrontFill,
						bg: '#F1E2C8'
					}
				]
			: []),
		...(tienda.envios.moto
			? [
					{
						t: 'Envío en moto',
						d: 'En el día en CABA, si pedís antes de las 16 h',
						i: I.phMopedFill,
						bg: '#F6E7CC'
					}
				]
			: []),
		{
			t: 'Envío a todo el país',
			d: tienda.envios.gratisDesde
				? `Gratis desde ${fmt(tienda.envios.gratisDesde)}`
				: tienda.envios.correo
					? 'Por Andreani o Correo Argentino'
					: 'Por Andreani',
			i: I.phPackageFill,
			bg: '#EFDDBF'
		}
	]);
	const best = $derived(
		tienda.masVendidos.slice(0, 4).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
	);
	// El bloque "Combo" del diseño, con las ofertas reales (precio tachado), la de mayor descuento primero.
	const ofertas = $derived(
		tienda.productos
			.filter((p) => p.anterior && p.stock > 0)
			.sort((a, b) => a.precio / a.anterior! - b.precio / b.anterior!)
			.slice(0, 3)
	);
	const oferta = $derived(ofertas[0] ?? null);
	const fotoGrande = $derived(img(tienda.portada.foto, 900) ?? best[0]?.img ?? null);
	const fotoChica = $derived(
		tienda.fotos[0] ? img(tienda.fotos[0], 600, 600) : (best[1]?.img ?? null)
	);

	// Catálogo (el menú del diseño): categoría, orden y de a 24, agrupado por categoría.
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
	const grupos = $derived.by(() => {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, sólo agrupa
		const g = new Map<string, Tarjeta[]>();
		for (const p of listado.slice(0, mostrar)) {
			const k = categoria ? '' : (p.categoria ?? 'Otros');
			g.set(k, [...(g.get(k) ?? []), tarjeta(p)]);
		}
		return [...g].map(([t, items]) => ({ t, items }));
	});
	const enlaceCat = (c: string | null, o = orden) => {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, sólo arma el link
		const q = new URLSearchParams();
		if (c) q.set('cat', c);
		if (o !== 'vendidos') q.set('orden', o);
		return `${base}/catalogo${q.size ? '?' + q : ''}`;
	};

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
	const resenasDe = $derived(
		producto ? tienda.resenas.filter((r) => r.producto === producto!.nombre).length : 0
	);
	let cantidad = $state(1);
	let cp = $state('');
	let cotizado = $state<OpcionEnvio[] | null>(null);
	function calcular() {
		if (cpValido(cp)) cotizado = cotizar(cp, producto?.precio ?? subtotal, tienda);
	}
	function agregar() {
		if (!producto || producto.stock === 0 || tienda.vitrina) return;
		carrito.sumar(producto.id, Math.min(cantidad, producto.stock), producto.stock);
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(`${base}/carrito`);
	}
	function sumarOferta() {
		if (!oferta || tienda.vitrina) return;
		carrito.sumar(oferta.id, 1, oferta.stock);
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
	// El pedido armado, para mandarlo por WhatsApp desde el menú.
	const textoPedido = $derived(
		`¡Hola! Quiero pedir:\n${lineas.map((l) => `${l.cantidad} × ${l.producto.nombre} (${fmt(l.subtotal)})`).join('\n')}\nTotal: ${fmt(subtotal)}`
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
			c: '#2F6B3A'
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
				{ t: 'Cambios y devoluciones', href: '/legales/arrepentimiento' },
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
		{ t: 'Menú', href: `${base}/catalogo`, on: pagina === 'catalogo' && !categoria },
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
	const B = "font-family:'Bricolage Grotesque',sans-serif";
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link
		href="https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,500;12..96,700;12..96,800&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<div
	class="mostaza"
	style="min-height:100vh;font-family:'Instrument Sans',system-ui,sans-serif;color:#3B2416;background:#FBF3E4"
>
	<div
		style="container-type:inline-size;width:100%;background:#FBF3E4;position:relative;overflow-x:clip"
	>
		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:12px;flex-wrap:wrap;padding:9px 16px;background:#C9821A;color:#2A170C;font-size:13px;font-weight:600;text-align:center"
		>
			<span
				>{tienda.envios.gratisDesde
					? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
					: `${tienda.cobros.cuotas} cuotas sin interés`}</span
			>{#if wide}<span style="padding:2px 7px;border-radius:6px;background:#2A170C;color:#FBF3E4"
					>{tienda.cobros.transferencia}% off por transferencia</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:center;gap:16px;padding:12px clamp(14px,3.5cqw,40px);background:rgba(251,243,228,.92);backdrop-filter:blur(10px)"
		>
			<a
				href={base}
				style="display:flex;align-items:center;gap:9px;min-width:0;text-decoration:none;color:#3B2416"
				><span
					style="width:38px;height:38px;border-radius:50%;background:#3B2416;color:#F5C26B;display:grid;place-items:center;flex:none"
					><Icon icon={I.phCoffeeBeanFill} size={20} /></span
				><span
					style="{B};font-weight:800;font-size:23px;letter-spacing:-.03em;white-space:nowrap;overflow:hidden;text-overflow:ellipsis"
					>{tienda.nombre}</span
				></a
			>
			{#if wide}<nav
					style="display:flex;gap:4px;margin-left:14px;min-width:0;overflow-x:auto;scrollbar-width:none"
				>
					{#each nav as n (n.t)}<a
							href={n.href}
							style="text-decoration:none;font:600 14px 'Instrument Sans';padding:9px 14px;border-radius:999px;background:{n.on
								? '#F1E2C8'
								: 'transparent'};color:#3B2416;white-space:nowrap">{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			<a
				href="{base}/carrito"
				aria-label="Carrito"
				style="display:flex;align-items:center;gap:8px;flex:none;text-decoration:none;padding:10px 16px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 14px 'Instrument Sans';white-space:nowrap"
				><Icon icon={I.phBagSimpleFill} size={17} />{carrito.unidades}{#if wide && subtotal}<span
						style="opacity:.7;font-weight:500">· {fmt(subtotal)}</span
					>{/if}</a
			>
			{#if nar}<a
					href="{base}/catalogo"
					aria-label="Menú"
					style="display:grid;flex:none;color:#3B2416"><Icon icon={I.phListBold} size={24} /></a
				>{/if}
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:10px 16px;background:#F1E2C8;font-size:13px;font-weight:600;text-align:center"
			>
				Es una tienda de ejemplo de la plantilla Mostaza: podés recorrerla, pero no se compra
				de&nbsp;verdad.
			</div>
		{/if}

		<!-- ============ INICIO ============ -->
		{#if pagina === 'inicio'}
			{#if ver('hero')}
				<section
					style="padding:clamp(16px,3cqw,40px) clamp(14px,3.5cqw,40px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));gap:clamp(12px,2cqw,24px)"
				>
					<div
						style="padding:clamp(26px,4.5cqw,60px);border-radius:28px;background:#3B2416;color:#FBF3E4;display:flex;flex-direction:column;justify-content:space-between;gap:30px;min-height:min(560px,120cqw);min-width:0"
					>
						{#if tienda.ciudad}<span
								style="align-self:flex-start;display:flex;align-items:center;gap:8px;padding:7px 12px;border-radius:999px;background:rgba(245,194,107,.16);color:#F5C26B;font-size:13px;font-weight:700"
								><Icon icon={I.phMapPinFill} size={15} />{tienda.ciudad}</span
							>{/if}
						<div>
							<h1
								style="margin:0;{B};font-weight:800;font-size:clamp(52px,8.4cqw,112px);line-height:.88;letter-spacing:-.05em;overflow-wrap:break-word"
							>
								{titulo[0]}{#if titulo[1]}<br /><span style="color:#F5C26B">{titulo[1]}</span>{/if}
							</h1>
							{#if tienda.portada.lema}<p
									style="margin:18px 0 0;font-size:17px;line-height:1.5;color:#E6D3BC;max-width:440px"
								>
									{tienda.portada.lema}
								</p>{/if}
						</div>
						<div style="display:flex;gap:10px;flex-wrap:wrap">
							<a
								href="{base}/catalogo"
								style="display:flex;align-items:center;gap:8px;padding:15px 24px;border-radius:999px;background:#F5C26B;color:#2A170C;font:700 15px 'Instrument Sans';text-decoration:none"
								><Icon icon={I.phHandTapFill} size={17} />Pedir ahora</a
							>{#if tienda.categorias[0]}<a
									href={enlaceCat(tienda.categorias[0].nombre)}
									style="padding:15px 22px;border:1.5px solid rgba(251,243,228,.35);border-radius:999px;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none"
									>{tienda.categorias[0].nombre}</a
								>{/if}
						</div>
					</div>
					<div
						style="display:grid;grid-template-columns:1fr 1fr;grid-template-rows:1fr 1fr;gap:clamp(10px,1.4cqw,16px);min-height:min(560px,110cqw)"
					>
						<div
							style="grid-row:span 2;border-radius:28px;overflow:hidden;background:#D9B98C;position:relative"
						>
							{#if fotoGrande}<img
									src={fotoGrande}
									alt=""
									style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
								/>{/if}{#if !tienda.portada.foto && best[0]}<a
									href={best[0].href}
									style="position:absolute;left:14px;bottom:14px;max-width:calc(100% - 28px);padding:8px 12px;border-radius:14px;background:#FBF3E4;font-size:13px;font-weight:700;color:#3B2416;text-decoration:none"
									>{best[0].n} · {best[0].price}</a
								>{/if}
						</div>
						<div
							style="border-radius:28px;overflow:hidden;background:#D9B98C;position:relative;grid-row:{tienda.promedio
								? 'auto'
								: 'span 2'}"
						>
							{#if fotoChica}<img
									src={fotoChica}
									alt=""
									style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
								/>{/if}
						</div>
						{#if tienda.promedio}
							<div
								style="border-radius:28px;background:#C9821A;color:#2A170C;padding:18px;display:flex;flex-direction:column;justify-content:space-between"
							>
								<Icon icon={I.phStarFill} size={24} />
								<div>
									<div
										style="{B};font-weight:800;font-size:clamp(34px,4cqw,48px);line-height:1;letter-spacing:-.04em"
									>
										{tienda.promedio.toLocaleString('es-AR')}
									</div>
									<div style="font-size:13px;font-weight:600;margin-top:4px">
										de promedio en reseñas
									</div>
								</div>
							</div>
						{/if}
					</div>
				</section>
			{/if}

			<section
				style="padding:clamp(12px,2cqw,24px) clamp(14px,3.5cqw,40px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,220px),1fr));gap:clamp(10px,1.4cqw,16px)"
			>
				{#each modes as m (m.t)}<a
						href="{base}/catalogo"
						style="display:flex;align-items:center;gap:14px;padding:18px;border-radius:22px;background:{m.bg};color:#3B2416;text-align:left;text-decoration:none;font-family:inherit"
						><span
							style="width:48px;height:48px;border-radius:16px;background:#FBF3E4;display:grid;place-items:center;flex:none"
							><Icon icon={m.i} size={24} style="color:#C9821A" /></span
						><span style="flex:1;min-width:0"
							><span style="display:block;{B};font-size:19px;font-weight:700;letter-spacing:-.02em"
								>{m.t}</span
							><span style="display:block;font-size:13px;color:#7A5A44;margin-top:2px">{m.d}</span
							></span
						><Icon icon={I.phArrowRightBold} size={17} /></a
					>{/each}
			</section>

			{#if ver('prods') && best.length}
				<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px) 0">
					<div
						style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
					>
						<h2
							style="margin:0;{B};font-weight:800;font-size:clamp(36px,5cqw,64px);line-height:.95;letter-spacing:-.045em"
						>
							Lo que más sale
						</h2>
						<a
							href="{base}/catalogo"
							style="text-decoration:none;padding:11px 18px;border-radius:999px;background:#F1E2C8;color:#3B2416;font:700 14px 'Instrument Sans'"
							>Ver menú completo</a
						>
					</div>
					<div
						style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,230px),1fr));gap:clamp(10px,1.4cqw,16px)"
					>
						{#each best as p (p.id)}<div
								style="padding:10px;border-radius:24px;background:#fff;box-shadow:0 1px 2px rgba(59,36,22,.06);min-width:0"
							>
								<a
									href={p.href}
									class="zoom"
									style="position:relative;display:block;aspect-ratio:1;border-radius:18px;overflow:hidden;background:#E9D6B8"
									>{#if p.img}<img
											src={p.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}{#if p.tag}<span
											style="position:absolute;top:8px;left:8px;padding:4px 9px;border-radius:999px;background:#FBF3E4;color:#3B2416;font-size:11.5px;font-weight:700"
											>{p.tag}</span
										>{/if}</a
								>
								<div style="padding:12px 6px 4px">
									<a
										href={p.href}
										style="display:block;{B};font-size:18px;font-weight:700;letter-spacing:-.02em;color:#3B2416;text-decoration:none"
										>{p.n}</a
									>
									{#if p.d}<div
											class="dos"
											style="font-size:12.5px;color:#7A5A44;margin-top:2px;line-height:1.4"
										>
											{p.d}
										</div>{/if}
									<div
										style="margin-top:10px;display:flex;justify-content:space-between;align-items:center;gap:8px"
									>
										<span
											><b style="font-size:16px">{p.price}</b>{#if p.old}
												<s style="font-size:13px;color:#9C8270">{p.old}</s>{/if}</span
										><button
											type="button"
											onclick={() => sumar(p.id, p.stock)}
											disabled={p.stock === 0 || tienda.vitrina}
											aria-label="Sumar {p.n}"
											style="width:36px;height:36px;flex:none;border:0;border-radius:50%;background:#3B2416;color:#FBF3E4;display:grid;place-items:center;cursor:pointer"
											><Icon icon={I.phPlusBold} size={15} /></button
										>
									</div>
								</div>
							</div>{/each}
					</div>
				</section>

				{#if oferta}
					<section
						style="margin:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px) 0;padding:clamp(24px,4cqw,48px);border-radius:28px;background:#F5C26B;color:#2A170C;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:28px;align-items:center"
					>
						<div>
							<div
								style="font-size:13px;font-weight:700;letter-spacing:.08em;text-transform:uppercase"
							>
								Oferta · −{Math.round((1 - oferta.precio / oferta.anterior!) * 100)}%
							</div>
							<h2
								style="margin:10px 0 0;{B};font-weight:800;font-size:clamp(38px,5.4cqw,70px);line-height:.92;letter-spacing:-.045em;overflow-wrap:break-word"
							>
								{oferta.nombre}
							</h2>
							<div
								style="margin-top:16px;display:flex;align-items:baseline;gap:12px;flex-wrap:wrap"
							>
								<b style="{B};font-size:40px;font-weight:800;letter-spacing:-.03em"
									>{fmt(oferta.precio)}</b
								><s style="font-size:17px;opacity:.6">{fmt(oferta.anterior!)}</s>
							</div>
							<button
								type="button"
								onclick={sumarOferta}
								disabled={tienda.vitrina}
								style="margin-top:18px;padding:15px 24px;border:0;border-radius:999px;background:#2A170C;color:#FBF3E4;font:700 15px 'Instrument Sans';cursor:pointer"
								>Sumar al pedido</button
							>
						</div>
						<div style="display:grid;grid-template-columns:repeat({ofertas.length},1fr);gap:10px">
							{#each ofertas as o (o.id)}<a
									href="{base}/producto/{o.id}"
									class="zoom"
									aria-label={o.nombre}
									style="display:block;aspect-ratio:3/4;border-radius:20px;overflow:hidden;background:#E3A94F"
									>{#if o.foto}<img
											src={img(o.foto, 300, 400)}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}</a
								>{/each}
						</div>
					</section>
				{/if}
			{/if}

			{#if ver('revs') && tienda.resenas.length}
				<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px) 0">
					<h2
						style="margin:0;{B};font-weight:800;font-size:clamp(36px,5cqw,64px);line-height:.95;letter-spacing:-.045em"
					>
						Dicen en el barrio
					</h2>
					<div
						style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,280px),1fr));gap:14px"
					>
						{#each tienda.resenas.slice(0, 3) as r, i (i)}
							{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
							<div
								style="padding:14px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:12px"
							>
								{#if pr?.foto}<div
										style="aspect-ratio:16/10;border-radius:16px;overflow:hidden;background:#E9D6B8"
									>
										<img
											src={img(pr.foto, 600, 375)}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>
									</div>{/if}
								<div style="display:flex;gap:2px;color:#C9821A">
									{#each estrellas(r.estrellas) as e, j (j)}<Icon icon={e} size={15} />{/each}
								</div>
								<p style="margin:0;font-size:15px;line-height:1.5">“{r.texto}”</p>
								<span style="font-size:13px;color:#7A5A44;font-weight:600"
									>{r.autor} · compró {r.producto}</span
								>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			{#if ver('ig') && tienda.fotos.length >= 3}
				<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px)">
					<h2
						style="margin:0;{B};font-weight:800;font-size:clamp(36px,5cqw,64px);line-height:.95;letter-spacing:-.045em"
					>
						De cerca
					</h2>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,170px),1fr));gap:8px"
					>
						{#each tienda.fotos as g (g)}<div
								style="aspect-ratio:1;border-radius:18px;overflow:hidden;background:#E9D6B8"
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
				<div style="height:clamp(48px,7cqw,90px)"></div>
			{/if}
		{/if}

		<!-- ============ CATÁLOGO (el "Menú y pedido" del diseño) ============ -->
		{#if pagina === 'catalogo'}
			<section style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px)">
				<div style="font-size:13px;color:#7A5A44">
					<a href={base} style="color:inherit;text-decoration:none">Inicio</a> /
					<span style="color:#3B2416">{categoria ?? 'Menú'}</span>
				</div>
				<div
					style="margin-top:10px;display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
				>
					<h1
						style="margin:0;{B};font-weight:800;font-size:clamp(46px,7cqw,92px);line-height:.9;letter-spacing:-.05em;overflow-wrap:break-word;min-width:0"
					>
						{categoria ?? 'Menú'}
					</h1>
					<label
						style="display:flex;align-items:center;gap:6px;padding:10px 15px;border-radius:999px;border:1.5px solid rgba(59,36,22,.12);background:#fff;color:#3B2416;font-size:13.5px;font-weight:700;white-space:nowrap"
						>Ordenar<select
							value={orden}
							onchange={(e) => goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
							style="border:0;background:transparent;font:700 13.5px 'Instrument Sans';color:#3B2416;cursor:pointer;padding:0"
							>{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
						></label
					>
				</div>
				<div
					style="margin-top:12px;display:flex;align-items:center;gap:8px;font-size:13.5px;color:#7A5A44;flex-wrap:wrap"
				>
					<Icon icon={I.phStorefrontFill} size={15} /><b style="color:#3B2416"
						>{listado.length} {listado.length === 1 ? 'producto' : 'productos'}</b
					>{#if tienda.envios.retiro && tienda.direccion}<span>· Retirás en {tienda.direccion}</span
						>{/if}{#if tienda.envios.gratisDesde}<span
							>· Envío gratis desde {fmt(tienda.envios.gratisDesde)}</span
						>{/if}
				</div>
				<div
					style="position:sticky;top:62px;z-index:10;margin:18px calc(-1 * clamp(14px,3.5cqw,40px)) 0;padding:10px clamp(14px,3.5cqw,40px);background:rgba(251,243,228,.95);backdrop-filter:blur(8px);display:flex;gap:6px;overflow-x:auto;scrollbar-width:none"
				>
					{#each [null, ...tienda.categorias.map((c) => c.nombre)] as c (c ?? '')}{@const on =
							c === categoria}<a
							href={enlaceCat(c)}
							style="flex:none;text-decoration:none;padding:9px 15px;border-radius:999px;background:{on
								? '#3B2416'
								: '#F1E2C8'};color:{on
								? '#FBF3E4'
								: '#3B2416'};font-size:14px;font-weight:700;white-space:nowrap">{c ?? 'Todo'}</a
						>{/each}
				</div>
				<div style="margin-top:14px;display:flex;gap:clamp(20px,3cqw,36px);align-items:flex-start">
					<div style="flex:1;min-width:0;display:flex;flex-direction:column;gap:30px">
						{#each grupos as s (s.t)}
							<div>
								{#if s.t}<h2
										style="margin:0 0 12px;{B};font-weight:800;font-size:30px;letter-spacing:-.03em"
									>
										{s.t}
									</h2>{/if}
								<div
									style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(100%,320px),1fr));gap:10px"
								>
									{#each s.items as p (p.id)}{@const q = enCarrito(p.id)}
										<div
											style="display:flex;gap:12px;padding:10px;border-radius:22px;background:#fff;border:1.5px solid {q
												? '#C9821A'
												: 'transparent'};min-width:0"
										>
											<div
												style="flex:1;min-width:0;padding:4px 2px 4px 6px;display:flex;flex-direction:column"
											>
												<div style="display:flex;align-items:center;gap:6px;flex-wrap:wrap">
													<a
														href={p.href}
														style="font-size:16px;font-weight:700;color:#3B2416;text-decoration:none"
														>{p.n}</a
													>{#if p.tag}<span
															style="padding:2px 8px;border-radius:999px;background:{p.tagBg};color:{p.tagFg};font-size:11px;font-weight:700"
															>{p.tag}</span
														>{/if}
												</div>
												{#if p.d}<span
														class="dos"
														style="font-size:13px;color:#7A5A44;line-height:1.4;margin-top:3px"
														>{p.d}</span
													>{/if}
												<span style="flex:1"></span>
												<span style="margin-top:8px"
													><b style="font-size:15.5px">{p.price}</b>{#if p.old}
														<s style="font-size:13px;color:#9C8270">{p.old}</s>{/if}</span
												>
											</div>
											<div
												style="position:relative;width:110px;height:110px;flex:none;border-radius:16px;overflow:hidden;background:#E9D6B8"
											>
												<a
													href={p.href}
													aria-label={p.n}
													style="display:block;width:100%;height:100%"
													>{#if p.img}<img
															src={img(porId.get(p.id)?.foto ?? null, 240, 240)}
															alt=""
															style="width:100%;height:100%;object-fit:cover"
														/>{/if}</a
												>
												<span
													style="position:absolute;right:6px;bottom:6px;display:flex;align-items:center;gap:2px;padding:3px;border-radius:999px;background:#fff;box-shadow:0 4px 10px rgba(59,36,22,.2)"
													>{#if q}<button
															type="button"
															aria-label="Uno menos"
															onclick={() => carrito.fijar(p.id, q - 1)}
															style="width:26px;height:26px;display:grid;place-items:center;border:0;background:transparent;padding:0;cursor:pointer;color:#3B2416"
															><Icon icon={I.phMinusBold} size={12} /></button
														><b style="min-width:16px;text-align:center;font-size:13px">{q}</b
														>{/if}<button
														type="button"
														aria-label="Sumar {p.n}"
														onclick={() => sumar(p.id, p.stock)}
														disabled={p.stock === 0 || tienda.vitrina || q >= p.stock}
														style="width:26px;height:26px;border:0;padding:0;border-radius:50%;background:#3B2416;color:#FBF3E4;display:grid;place-items:center;cursor:pointer"
														><Icon icon={I.phPlusBold} size={12} /></button
													></span
												>
											</div>
										</div>{/each}
								</div>
							</div>
						{:else}
							<p style="margin:20px 0;font-size:15px;color:#7A5A44">
								Todavía no hay productos en esta&nbsp;categoría.
							</p>
						{/each}
						{#if listado.length > mostrar}
							<div style="display:flex;justify-content:center">
								<button
									type="button"
									onclick={() => (mostrar += 24)}
									style="border:0;cursor:pointer;padding:11px 22px;border-radius:999px;background:#F1E2C8;color:#3B2416;font:700 14px 'Instrument Sans'"
									>Ver más</button
								>
							</div>
						{/if}
					</div>
					{#if wide}
						<aside
							style="width:330px;flex:none;position:sticky;top:130px;padding:20px;border-radius:26px;background:#fff;box-shadow:0 24px 50px -30px rgba(59,36,22,.4);display:flex;flex-direction:column;gap:12px"
						>
							<div style="{B};font-weight:800;font-size:24px;letter-spacing:-.03em">Tu pedido</div>
							{#each lineas as o (o.id)}<div
									style="display:flex;align-items:center;gap:10px;font-size:14px"
								>
									<b
										style="width:26px;height:26px;border-radius:8px;background:#F1E2C8;display:grid;place-items:center;font-size:12.5px;flex:none"
										>{o.cantidad}</b
									><span style="flex:1;min-width:0">{o.producto.nombre}</span><b
										>{fmt(o.subtotal)}</b
									>
								</div>{:else}<span style="font-size:13.5px;color:#7A5A44"
									>Sumá productos del menú con el&nbsp;+.</span
								>{/each}
							{#if lineas.length}
								<div style="height:1px;background:rgba(59,36,22,.1)"></div>
								<div style="display:flex;justify-content:space-between;align-items:baseline">
									<b>Total</b><b style="{B};font-size:28px;font-weight:800">{fmt(subtotal)}</b>
								</div>
								<span style="font-size:13px;color:#2F6B3A;font-weight:600"
									>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
								>
								<a
									href="{base}/checkout"
									style="padding:16px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-align:center;text-decoration:none"
									>Finalizar pedido</a
								>{#if whatsapp}<a
										href={wa(textoPedido)}
										target="_blank"
										rel="noopener"
										style="display:flex;justify-content:center;align-items:center;gap:8px;padding:14px;border-radius:999px;background:#E8F7EC;color:#1F6F3A;font-size:14.5px;font-weight:700;text-decoration:none"
										><Icon icon={I.phWhatsappLogoFill} size={18} />Mandar por WhatsApp</a
									>{/if}
							{/if}
						</aside>
					{/if}
				</div>
				{#if nar && carrito.unidades}<a
						href="{base}/carrito"
						style="position:sticky;bottom:14px;margin-top:20px;width:100%;box-sizing:border-box;display:flex;align-items:center;gap:10px;padding:14px 18px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none;box-shadow:0 16px 30px -12px rgba(59,36,22,.6)"
						><span
							style="width:26px;height:26px;border-radius:50%;background:#F5C26B;color:#2A170C;display:grid;place-items:center;font-size:13px"
							>{carrito.unidades}</span
						><span style="flex:1;text-align:left">Ver pedido</span>{fmt(subtotal)}</a
					>{/if}
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto}
			<section
				style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));gap:clamp(20px,4cqw,56px);align-items:start"
			>
				<div
					style="position:relative;aspect-ratio:1;border-radius:30px;overflow:hidden;background:#E9C98F"
				>
					{#if producto.foto}<img
							src={img(producto.foto, 1000, 1000)}
							alt={producto.nombre}
							style="width:100%;height:100%;object-fit:cover"
						/>{/if}{#if producto.anterior || producto.etiqueta}<span
							style="position:absolute;top:14px;left:14px;padding:4px 10px;border-radius:999px;background:#FBF3E4;font-size:11.5px;font-weight:700"
							>{producto.anterior
								? `−${Math.round((1 - producto.precio / producto.anterior) * 100)}%`
								: producto.etiqueta}</span
						>{/if}
				</div>
				<div style="position:sticky;top:80px;min-width:0">
					<div style="font-size:13px;color:#7A5A44">
						<a href={base} style="color:inherit;text-decoration:none">Inicio</a> /
						<a href="{base}/catalogo" style="color:inherit;text-decoration:none">Menú</a>
					</div>
					{#if producto.categoria}<a
							href={enlaceCat(producto.categoria)}
							style="display:inline-flex;margin-top:12px;padding:5px 11px;border-radius:999px;background:#F1E2C8;font-size:12.5px;font-weight:700;color:#3B2416;text-decoration:none"
							>{producto.categoria}</a
						>{/if}
					<h1
						style="margin:12px 0 0;{B};font-weight:800;font-size:clamp(42px,5cqw,68px);line-height:.92;letter-spacing:-.045em;overflow-wrap:break-word"
					>
						{producto.nombre}
					</h1>
					{#if resenasDe}<div
							style="margin-top:10px;display:flex;align-items:center;gap:8px;font-size:13px;font-weight:600"
						>
							<span style="display:flex;color:#C9821A"
								>{#each estrellas(5) as e, i (i)}<Icon icon={e} size={13} />{/each}</span
							><span>{resenasDe} {resenasDe === 1 ? 'reseña' : 'reseñas'}</span>
						</div>{/if}
					{#if producto.descripcion}<p
							style="margin:12px 0 0;font-size:16px;line-height:1.55;color:#5E4332"
						>
							{producto.descripcion}
						</p>{/if}
					<div style="margin-top:18px;display:flex;align-items:baseline;gap:10px">
						<b style="{B};font-size:36px;font-weight:800;letter-spacing:-.03em"
							>{fmt(producto.precio)}</b
						>{#if producto.anterior}<s style="color:#9C8270">{fmt(producto.anterior)}</s>{/if}
					</div>
					<div style="font-size:13.5px;color:#5E4332">
						{tienda.cobros.cuotas} cuotas sin interés de {fmt(
							producto.precio / tienda.cobros.cuotas
						)} ·
						<b style="color:#2F6B3A"
							>{fmt(producto.precio * (1 - tienda.cobros.transferencia / 100))} por transferencia</b
						>
					</div>
					{#if producto.stock > 0 && producto.stock <= 3}<div
							style="margin-top:12px;font-size:13px;font-weight:700;color:#C9821A"
						>
							{producto.stock === 1 ? 'Queda 1' : `Quedan ${producto.stock}`}
						</div>{/if}
					<div style="margin-top:14px;display:flex;gap:8px">
						{#if producto.stock > 0}<span
								style="display:flex;align-items:center;gap:4px;padding:4px;border-radius:999px;background:#fff;flex:none"
								><button
									type="button"
									aria-label="Uno menos"
									disabled={cantidad <= 1}
									onclick={() => (cantidad = Math.max(1, cantidad - 1))}
									style="width:40px;height:40px;display:grid;place-items:center;border:0;background:transparent;padding:0;cursor:pointer;color:#3B2416"
									><Icon icon={I.phMinusBold} size={14} /></button
								><b style="width:24px;text-align:center">{cantidad}</b><button
									type="button"
									aria-label="Uno más"
									disabled={cantidad >= producto.stock}
									onclick={() => (cantidad = Math.min(producto!.stock, cantidad + 1))}
									style="width:40px;height:40px;display:grid;place-items:center;border:0;background:transparent;padding:0;cursor:pointer;color:#3B2416"
									><Icon icon={I.phPlusBold} size={14} /></button
								></span
							>{/if}<button
							type="button"
							onclick={agregar}
							disabled={producto.stock === 0 || tienda.vitrina}
							style="flex:1;padding:15px;border:0;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';cursor:pointer"
							>{producto.stock === 0
								? 'Sin stock'
								: `Agregar · ${fmt(producto.precio * cantidad)}`}</button
						>
					</div>
					{#if whatsapp}<a
							href={wa(`¡Hola! Quiero consultar por ${producto.nombre} (${fmt(producto.precio)}).`)}
							target="_blank"
							rel="noopener"
							style="margin-top:8px;display:flex;justify-content:center;align-items:center;gap:8px;padding:14px;border-radius:999px;background:#E8F7EC;color:#1F6F3A;font-size:14.5px;font-weight:700;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={18} />Consultar por WhatsApp</a
						>{/if}
					<div
						style="margin-top:14px;padding:14px;border-radius:18px;background:#fff;display:flex;flex-direction:column;gap:8px;font-size:13.5px"
					>
						<label for="cp-ficha" style="font-weight:700">Calculá el envío</label>
						<div style="display:flex;gap:6px">
							<input
								id="cp-ficha"
								bind:value={cp}
								inputmode="numeric"
								maxlength="4"
								placeholder="Código postal"
								onkeydown={(e) => e.key === 'Enter' && calcular()}
								style="flex:1;min-width:0;padding:10px 12px;border-radius:12px;border:0;background:#FBF3E4;font:400 13.5px 'Instrument Sans';color:#3B2416"
							/><button
								type="button"
								onclick={calcular}
								style="padding:10px 14px;border:0;border-radius:12px;background:#3B2416;color:#FBF3E4;font:700 13.5px 'Instrument Sans';cursor:pointer"
								>Calcular</button
							>
						</div>
						{#if cotizado}{#each cotizado as o (o.id)}<span
									style="display:flex;justify-content:space-between;gap:10px"
									><span>{o.nombre} · {o.detalle}</span><b
										style="white-space:nowrap;color:{o.precio ? '#3B2416' : '#2F6B3A'}"
										>{o.precio ? fmt(o.precio) : 'Gratis'}</b
									></span
								>{/each}{/if}
					</div>
					<div style="margin-top:14px;display:flex;flex-direction:column;gap:6px">
						{#each acc as a (a.t)}<div
								style="padding:12px 16px;border-radius:14px;background:{a.open
									? '#F1E2C8'
									: '#fff'};font-size:14.5px"
							>
								<button
									type="button"
									onclick={() => (abierto = a.open ? '' : a.t)}
									aria-expanded={a.open}
									style="width:100%;display:flex;justify-content:space-between;align-items:center;padding:0;border:0;background:transparent;font:600 14.5px 'Instrument Sans';color:#3B2416;cursor:pointer;text-align:left"
									>{a.t}<Icon icon={a.i} size={14} /></button
								>{#if a.open}<p
										style="margin:8px 0 0;font-size:14px;line-height:1.55;color:#5E4332"
									>
										{a.d}
									</p>{/if}
							</div>{/each}
					</div>
				</div>
			</section>
			{#if related.length}
				<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px)">
					<h2
						style="margin:0;{B};font-weight:800;font-size:clamp(32px,4.4cqw,52px);letter-spacing:-.04em"
					>
						Va bien con
					</h2>
					<div
						style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:12px"
					>
						{#each related as p (p.id)}<a
								href={p.href}
								class="zoom"
								style="display:block;padding:10px;border-radius:22px;background:#fff;color:#3B2416;text-decoration:none;min-width:0"
								><span
									style="display:block;aspect-ratio:1;border-radius:16px;overflow:hidden;background:#E9D6B8"
									>{#if p.img}<img
											src={p.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}</span
								><span
									style="padding:10px 4px 2px;display:flex;justify-content:space-between;gap:8px;flex-wrap:wrap"
									><b style="font-size:14.5px">{p.n}</b><span style="font-size:14px;font-weight:700"
										>{p.price}</span
									></span
								></a
							>{/each}
					</div>
				</section>
			{:else}
				<div style="height:clamp(48px,7cqw,90px)"></div>
			{/if}
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section
				style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(20px,4cqw,48px);align-items:start"
			>
				<div style="min-width:0">
					<h1
						style="margin:0;{B};font-weight:800;font-size:clamp(44px,6cqw,80px);line-height:.9;letter-spacing:-.05em"
					>
						Tu pedido
					</h1>
					{#if !lineas.length}
						<p style="margin:18px 0 0;font-size:15px;color:#7A5A44">Todavía no sumaste nada.</p>
						<a
							href="{base}/catalogo"
							style="display:inline-block;margin-top:16px;padding:15px 24px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none"
							>Ver el menú</a
						>
					{:else}
						{#if tienda.envios.gratisDesde}
							<div
								style="margin-top:14px;padding:12px 14px;border-radius:16px;background:#F1E2C8;font-size:13.5px;display:flex;flex-direction:column;gap:10px"
							>
								<span style="display:flex;align-items:center;gap:10px"
									><Icon icon={I.phPackageFill} size={20} style="color:#C9821A" /><span
										style="flex:1"
										>{#if falta}Sumá <b>{fmt(falta)}</b> y el envío es gratis{:else}<b
												>Tenés envío gratis</b
											>{/if}</span
									></span
								><span
									style="height:5px;border-radius:3px;background:#EADBC2;overflow:hidden;display:block"
									><span
										style="display:block;height:100%;width:{Math.min(
											100,
											(subtotal / tienda.envios.gratisDesde) * 100
										)}%;background:#C9821A"
									></span></span
								>
							</div>
						{/if}
						<div style="margin-top:16px;display:flex;flex-direction:column;gap:10px">
							{#each lineas as o (o.id)}<div
									style="display:flex;gap:12px;align-items:center;padding:10px;border-radius:20px;background:#fff"
								>
									{#if o.producto.foto}<img
											src={img(o.producto.foto, 160, 160)}
											alt=""
											style="width:72px;height:72px;border-radius:14px;object-fit:cover;flex:none;background:#E9D6B8"
										/>{/if}
									<div style="flex:1;min-width:0">
										<a
											href="{base}/producto/{o.id}"
											style="font-size:15px;font-weight:700;color:#3B2416;text-decoration:none"
											>{o.producto.nombre}</a
										>
										<div style="font-size:12.5px;color:#7A5A44">{fmt(o.producto.precio)} c/u</div>
										<div style="margin-top:4px;display:flex;align-items:center;gap:10px">
											<b style="font-size:14.5px">{fmt(o.subtotal)}</b><button
												type="button"
												onclick={() => carrito.fijar(o.id, 0)}
												style="border:0;background:transparent;padding:0;font:600 12.5px 'Instrument Sans';text-decoration:underline;color:#7A5A44;cursor:pointer"
												>Quitar</button
											>
										</div>
									</div>
									<span
										style="display:flex;align-items:center;gap:2px;padding:3px;border-radius:999px;background:#FBF3E4;flex:none"
										><button
											type="button"
											aria-label="Uno menos"
											onclick={() => carrito.fijar(o.id, o.cantidad - 1)}
											style="width:30px;height:30px;display:grid;place-items:center;border:0;background:transparent;padding:0;cursor:pointer;color:#3B2416"
											><Icon icon={I.phMinusBold} size={12} /></button
										><b style="min-width:18px;text-align:center">{o.cantidad}</b><button
											type="button"
											aria-label="Uno más"
											disabled={o.cantidad >= o.producto.stock}
											onclick={() => carrito.fijar(o.id, o.cantidad + 1)}
											style="width:30px;height:30px;border:0;padding:0;border-radius:50%;background:#3B2416;color:#FBF3E4;display:grid;place-items:center;cursor:pointer"
											><Icon icon={I.phPlusBold} size={12} /></button
										></span
									>
								</div>{/each}
						</div>
						{#if upsell.length}
							<div style="margin-top:20px">
								<b style="font-size:15px">¿Le sumás algo?</b>
								<div
									style="margin-top:10px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none"
								>
									{#each upsell as u (u.id)}<div
											style="flex:none;display:flex;align-items:center;gap:10px;padding:8px 12px 8px 8px;border-radius:18px;background:#fff"
										>
											{#if u.foto}<img
													src={img(u.foto, 120, 120)}
													alt=""
													style="width:48px;height:48px;border-radius:12px;object-fit:cover"
												/>{/if}
											<div>
												<a
													href="{base}/producto/{u.id}"
													style="display:block;font-size:13.5px;font-weight:700;white-space:nowrap;color:#3B2416;text-decoration:none"
													>{u.nombre}</a
												>
												<div style="font-size:13px">{fmt(u.precio)}</div>
											</div>
											<button
												type="button"
												aria-label="Sumar {u.nombre}"
												disabled={tienda.vitrina}
												onclick={() => sumar(u.id, u.stock)}
												style="width:30px;height:30px;border:0;padding:0;border-radius:50%;background:#F5C26B;color:#2A170C;display:grid;place-items:center;cursor:pointer"
												><Icon icon={I.phPlusBold} size={13} /></button
											>
										</div>{/each}
								</div>
							</div>
						{/if}
					{/if}
				</div>
				{#if lineas.length}
					<div
						style="position:sticky;top:80px;padding:22px;border-radius:26px;background:#3B2416;color:#FBF3E4;display:flex;flex-direction:column;gap:12px;font-size:14.5px"
					>
						<span style="display:flex;justify-content:space-between;gap:10px"
							><span style="color:#E6D3BC">Subtotal</span><span>{fmt(subtotal)}</span></span
						>
						<span style="display:flex;justify-content:space-between;gap:10px"
							><span style="color:#E6D3BC">Envío</span><span style="text-align:right"
								>Se calcula en el próximo paso</span
							></span
						>
						<div style="height:1px;background:rgba(251,243,228,.14)"></div>
						<span style="display:flex;justify-content:space-between;align-items:baseline"
							><b>Total</b><b style="{B};font-size:32px;font-weight:800">{fmt(subtotal)}</b></span
						>
						<span style="font-size:13px;color:#E6D3BC"
							>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
								subtotal / tienda.cobros.cuotas
							)}</span
						>
						<span style="font-size:13px;color:#F5C26B;font-weight:600"
							>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
						>
						<a
							href="{base}/checkout"
							style="padding:16px;border-radius:999px;background:#F5C26B;color:#2A170C;font:700 15px 'Instrument Sans';text-align:center;text-decoration:none"
							>Continuar</a
						>
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ CHECKOUT ============ -->
		{#if pagina === 'checkout'}
			{#if !lineas.length}
				<section
					style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);max-width:760px;margin:0 auto"
				>
					<p style="font-size:15px;color:#7A5A44">Tu pedido está vacío.</p>
					<a href="{base}/catalogo" style="color:#3B2416;font-weight:700">Ver el menú</a>
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
					style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);max-width:760px;margin:0 auto;display:flex;flex-direction:column;gap:16px"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<h1
						style="margin:0;{B};font-weight:800;font-size:clamp(40px,5.4cqw,68px);line-height:.92;letter-spacing:-.05em"
					>
						Casi listo
					</h1>
					<div
						style="padding:18px;border-radius:24px;background:#fff;display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:10px"
					>
						<b style="grid-column:1/-1;font-size:16px">Tus datos</b>
						{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['dni', 'DNI (opcional)', 'text', 'off']] as [n, l, tipo, ac] (n)}
							<label style="display:block;padding:10px 12px;border-radius:14px;background:#FBF3E4"
								><span style="display:block;font-size:11.5px;color:#7A5A44">{l}</span><input
									name={n}
									type={tipo}
									autocomplete={ac as 'email'}
									required={n !== 'dni'}
									maxlength="120"
									style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 15px 'Instrument Sans';color:#3B2416;outline:none"
								/></label
							>
						{/each}
					</div>
					<div
						style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:8px"
					>
						<b style="font-size:16px">¿Cómo te llega?</b>
						<div style="display:grid;grid-template-columns:minmax(0,130px) minmax(0,1fr);gap:10px">
							<label style="display:block;padding:10px 12px;border-radius:14px;background:#FBF3E4"
								><span style="display:block;font-size:11.5px;color:#7A5A44">Código postal</span
								><input
									name="cp"
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 15px 'Instrument Sans';color:#3B2416;outline:none"
								/></label
							>
							<label style="display:block;padding:10px 12px;border-radius:14px;background:#FBF3E4"
								><span style="display:block;font-size:11.5px;color:#7A5A44"
									>Dirección (calle, número, piso)</span
								><input
									name="direccion"
									maxlength="160"
									autocomplete="street-address"
									required={envio !== 'retiro'}
									style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 15px 'Instrument Sans';color:#3B2416;outline:none"
								/></label
							>
						</div>
						<div
							style="display:flex;flex-direction:column;gap:8px"
							role="radiogroup"
							aria-label="Forma de entrega"
						>
							{#each opciones as s (s.id)}{@const on = envio === s.id}<button
									type="button"
									role="radio"
									aria-checked={on}
									onclick={() => (envio = s.id)}
									style="display:flex;align-items:center;gap:12px;padding:14px;border-radius:16px;border:1.5px solid {on
										? '#3B2416'
										: 'rgba(59,36,22,.12)'};background:{on
										? '#FBF3E4'
										: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#3B2416"
									><Icon
										icon={s.id === 'retiro'
											? I.phStorefrontFill
											: s.id === 'moto'
												? I.phMopedFill
												: I.phPackageFill}
										size={22}
										style="color:#C9821A"
									/><span style="flex:1;min-width:0"
										><b style="display:block;font-size:14.5px">{s.nombre}</b><span
											style="display:block;font-size:12.5px;color:#7A5A44">{s.detalle}</span
										></span
									><b
										style="font-size:14px;white-space:nowrap;color:{s.precio
											? '#3B2416'
											: '#2F6B3A'}">{s.precio ? fmt(s.precio) : 'Gratis'}</b
									><span
										style="width:20px;height:20px;border-radius:50%;border:2px solid #3B2416;display:grid;place-items:center;flex:none"
										>{#if on}<span
												style="width:10px;height:10px;border-radius:50%;background:#3B2416"
											></span>{/if}</span
									></button
								>{:else}<span style="font-size:13.5px;color:#7A5A44"
									>Poné tu código postal para ver cómo te&nbsp;llega.</span
								>{/each}
						</div>
					</div>
					<div
						style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:8px"
						role="radiogroup"
						aria-label="Forma de pago"
					>
						<b style="font-size:16px">Pago</b>
						{#each pagos as p (p.id)}{@const on = pago === p.id}<button
								type="button"
								role="radio"
								aria-checked={on}
								onclick={() => (pago = p.id)}
								style="display:flex;align-items:center;gap:12px;padding:14px;border-radius:16px;border:1.5px solid {on
									? '#3B2416'
									: 'rgba(59,36,22,.12)'};background:{on
									? '#FBF3E4'
									: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#3B2416"
								><Icon icon={p.i} size={22} style="color:{p.c}" /><span style="flex:1;min-width:0"
									><b style="display:block;font-size:14.5px">{p.t}</b><span
										style="display:block;font-size:12.5px;color:#7A5A44">{p.d}</span
									></span
								><span
									style="width:20px;height:20px;border-radius:50%;border:2px solid #3B2416;display:grid;place-items:center;flex:none"
									>{#if on}<span style="width:10px;height:10px;border-radius:50%;background:#3B2416"
										></span>{/if}</span
								></button
							>{/each}
					</div>
					<div
						style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:12px"
					>
						<b style="font-size:16px">Tu pedido</b>
						{#each lineas as o (o.id)}<div
								style="display:flex;align-items:center;gap:10px;font-size:14px"
							>
								<b
									style="width:26px;height:26px;border-radius:8px;background:#F1E2C8;display:grid;place-items:center;font-size:12.5px;flex:none"
									>{o.cantidad}</b
								><span style="flex:1;min-width:0">{o.producto.nombre}</span><b>{fmt(o.subtotal)}</b>
							</div>{/each}
						<div style="height:1px;background:rgba(59,36,22,.1)"></div>
						<span style="display:flex;justify-content:space-between;font-size:14px"
							><span>Subtotal</span><span>{fmt(subtotal)}</span></span
						><span style="display:flex;justify-content:space-between;font-size:14px"
							><span>Envío</span><span
								>{elegida ? (elegida.precio ? fmt(elegida.precio) : 'Gratis') : '—'}</span
							></span
						>{#if descuento}<span
								style="display:flex;justify-content:space-between;font-size:14px;color:#2F6B3A;font-weight:600"
								><span>Transferencia −{tienda.cobros.transferencia}%</span><span
									>− {fmt(descuento)}</span
								></span
							>{/if}
						<div style="display:flex;justify-content:space-between;align-items:baseline">
							<b>Total</b><b style="{B};font-size:28px;font-weight:800">{fmt(total)}</b>
						</div>
					</div>
					{#if error}<div
							role="alert"
							style="padding:12px 14px;border-radius:16px;background:#F6E2D2;color:#9C3B1F;font-size:14px;font-weight:600"
						>
							{error}
						</div>{/if}
					<button
						disabled={!envio || enviando || tienda.vitrina}
						style="padding:18px;border:0;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 16px 'Instrument Sans';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:10px"
						>{tienda.vitrina
							? 'Es una tienda de ejemplo'
							: envio
								? `Confirmar pedido · ${fmt(total)}`
								: 'Elegí cómo te llega'}<Icon icon={I.phArrowRightBold} size={16} /></button
					>
					<span
						style="display:flex;justify-content:center;align-items:center;gap:8px;text-align:center;font-size:13px;color:#7A5A44"
						><Icon icon={I.phShieldCheckFill} size={16} />Compra protegida · 10 días
						para&nbsp;arrepentirte</span
					>
				</form>
			{/if}
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas del checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="padding:clamp(32px,5cqw,64px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);max-width:760px;margin:0 auto;display:flex;flex-direction:column;gap:16px;text-align:center;align-items:center"
			>
				<span
					style="width:64px;height:64px;border-radius:50%;background:#3B2416;color:#F5C26B;display:grid;place-items:center"
					><Icon icon={I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:0;{B};font-weight:800;font-size:clamp(40px,5.4cqw,68px);line-height:.92;letter-spacing:-.05em;overflow-wrap:break-word;max-width:100%"
				>
					¡Gracias, {pedido.nombre}!
				</h1>
				<p style="margin:0;font-size:16px;line-height:1.55;color:#5E4332">
					Recibimos tu pedido <b>#{pedido.numero}</b> por <b>{fmt(pedido.total)}</b>. Te escribimos
					por mail con cada&nbsp;novedad.
				</p>
				{#if pedido.pago === 'transferencia'}
					<div
						style="padding:18px;border-radius:24px;background:#fff;text-align:left;font-size:14.5px;line-height:1.6;align-self:stretch"
					>
						<b>Transferí {fmt(pedido.total)}</b>
						{#if pedido.alias}al alias <b>{pedido.alias}</b>{:else}con los datos que te mandamos por
							mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
								target="_blank"
								rel="noopener"
								style="color:#3B2416;font-weight:700">por WhatsApp</a
							>{/if}. Lo preparamos apenas se&nbsp;acredite.
					</div>
				{/if}
				<a
					href={base}
					style="margin-top:6px;padding:15px 24px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none"
					>Volver a la tienda</a
				>
			</section>
		{/if}

		<!-- ============ NOSOTROS ============ -->
		{#if pagina === 'nosotros'}
			<section
				style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(14px,2cqw,24px)"
			>
				<div
					style="padding:clamp(26px,4.5cqw,56px);border-radius:28px;background:#C9821A;color:#2A170C;display:flex;flex-direction:column;justify-content:flex-end;min-height:min(520px,110cqw);min-width:0"
				>
					{#if tienda.ciudad}<div
							style="font-size:13px;font-weight:700;letter-spacing:.08em;text-transform:uppercase"
						>
							{tienda.ciudad}
						</div>{/if}
					<h1
						style="margin:12px 0 0;{B};font-weight:800;font-size:clamp(46px,6.6cqw,92px);line-height:.88;letter-spacing:-.05em;overflow-wrap:break-word"
					>
						{tienda.portada.titulo}
					</h1>
				</div>
				{#if tienda.portada.foto || tienda.fotos[0]}
					<div
						style="border-radius:28px;overflow:hidden;background:#D9B98C;min-height:min(520px,110cqw);position:relative"
					>
						<img
							src={img(tienda.portada.foto ?? tienda.fotos[0], 1000)}
							alt=""
							style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
						/>
					</div>
				{/if}
			</section>
			<section
				style="padding:clamp(14px,2cqw,24px) clamp(14px,3.5cqw,40px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:clamp(10px,1.4cqw,16px)"
			>
				<div style="padding:22px;border-radius:24px;background:#fff">
					<div style="{B};font-weight:800;font-size:48px;line-height:1;letter-spacing:-.04em">
						{tienda.productos.length}
					</div>
					<div style="font-size:14px;color:#7A5A44;margin-top:6px">productos en el menú</div>
				</div>
				<div style="padding:22px;border-radius:24px;background:#fff">
					<div style="{B};font-weight:800;font-size:48px;line-height:1;letter-spacing:-.04em">
						{tienda.cobros.cuotas}
					</div>
					<div style="font-size:14px;color:#7A5A44;margin-top:6px">cuotas sin interés</div>
				</div>
				{#if tienda.promedio}<div style="padding:22px;border-radius:24px;background:#fff">
						<div style="{B};font-weight:800;font-size:48px;line-height:1;letter-spacing:-.04em">
							{tienda.promedio.toLocaleString('es-AR')}
						</div>
						<div style="font-size:14px;color:#7A5A44;margin-top:6px">de promedio en reseñas</div>
					</div>{/if}
			</section>
			<section
				style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(20px,3cqw,40px)"
			>
				<div style="min-width:0">
					<h2
						style="margin:0;{B};font-weight:800;font-size:clamp(36px,5cqw,60px);letter-spacing:-.045em"
					>
						{tienda.envios.retiro ? 'Vení a vernos' : 'Hablemos'}
					</h2>
					<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:#5E4332">
						{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. Pedís online y te lo
						mandamos{tienda.envios.retiro ? ', o lo retirás en el local' : ' a todo el país'}.
					</p>
					<div style="margin-top:16px;display:flex;flex-direction:column;gap:10px;font-size:15px">
						{#if tienda.direccion}<span style="display:flex;gap:10px"
								><Icon
									icon={I.phMapPinFill}
									size={19}
									style="color:#C9821A"
								/>{tienda.direccion}</span
							>{/if}{#if tienda.whatsapp}<span style="display:flex;gap:10px"
								><Icon
									icon={I.phWhatsappLogoFill}
									size={19}
									style="color:#25A85A"
								/>{tienda.whatsapp}</span
							>{/if}
					</div>
					{#if whatsapp}<a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="margin-top:18px;display:inline-flex;align-items:center;gap:8px;padding:14px 22px;border-radius:999px;background:#E8F7EC;color:#1F6F3A;font-size:14.5px;font-weight:700;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={18} />Escribinos por WhatsApp</a
						>{/if}
				</div>
				{#if tienda.direccion}
					<a
						href="https://www.google.com/maps/search/?api=1&query={encodeURIComponent(
							[tienda.direccion, tienda.ciudad].filter(Boolean).join(', ')
						)}"
						target="_blank"
						rel="noopener"
						style="position:relative;display:block;min-height:340px;border-radius:28px;overflow:hidden;background:#EADBC2;text-decoration:none"
						><span
							style="position:absolute;inset:0;background-image:linear-gradient(rgba(59,36,22,.07) 1px,transparent 1px),linear-gradient(90deg,rgba(59,36,22,.07) 1px,transparent 1px);background-size:32px 32px"
						></span><span
							style="position:absolute;left:-10%;right:-10%;top:52%;height:16px;background:#fff;transform:rotate(-12deg)"
						></span><span
							style="position:absolute;top:-10%;bottom:-10%;left:40%;width:14px;background:#fff;transform:rotate(8deg)"
						></span><span
							style="position:absolute;left:44%;top:40%;max-width:52%;display:flex;align-items:center;gap:8px;padding:8px 12px 8px 8px;border-radius:999px;background:#3B2416;color:#FBF3E4;font-size:13px;font-weight:700;box-shadow:0 10px 20px -6px rgba(59,36,22,.6)"
							><span
								style="width:26px;height:26px;border-radius:50%;background:#F5C26B;color:#2A170C;display:grid;place-items:center;flex:none"
								><Icon icon={I.phCoffeeBeanFill} size={15} /></span
							><span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
								>{tienda.nombre}</span
							></span
						><span
							style="position:absolute;left:16px;bottom:14px;font-size:12px;font-weight:600;color:#7A5A44"
							>Ver en Google Maps</span
						></a
					>
				{/if}
			</section>
		{/if}

		<!-- footer -->
		<footer
			style="margin:0 clamp(14px,3.5cqw,40px) clamp(14px,3.5cqw,40px);padding:clamp(26px,4cqw,48px);border-radius:28px;background:#3B2416;color:#FBF3E4"
		>
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,190px),1fr));gap:28px"
			>
				<div style="min-width:0">
					<div style="{B};font-weight:800;font-size:32px;letter-spacing:-.04em">
						{tienda.nombre}
					</div>
					<p style="margin:10px 0 0;font-size:14px;color:#E6D3BC;line-height:1.5">
						{tienda.cobros.cuotas} cuotas sin interés y {tienda.cobros.transferencia}% off por
						transferencia, en toda la&nbsp;tienda.
					</p>
				</div>
				{#each foot as f (f.t)}<div
						style="display:flex;flex-direction:column;gap:8px;font-size:14px"
					>
						<b style="color:#F5C26B;margin-bottom:4px">{f.t}</b>{#each f.l as l (l.t)}<a
								href={l.href}
								style="color:#E6D3BC;text-decoration:none">{l.t}</a
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
				style="margin-top:28px;padding-top:16px;border-top:1px solid rgba(251,243,228,.12);display:flex;justify-content:space-between;gap:12px;flex-wrap:wrap;font-size:12.5px;color:#C9B29A"
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
						href="/legales/arrepentimiento"
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
	.zoom :global(img) {
		transition: transform 0.8s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.zoom:hover :global(img) {
		transform: scale(1.04);
	}
	.dos {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
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
