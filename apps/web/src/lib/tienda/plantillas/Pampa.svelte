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
	 * Plantilla Pampa (design/Plantilla Pampa.dc.html), con los datos de la tienda en lugar de los de
	 * ejemplo. Lo que el diseño muestra sin datos atrás (talles, colores, combo "fit completo", cuenta
	 * regresiva del drop, reserva de stock, aviso del próximo drop, cupón, seguidores) no se inventa:
	 * esas partes no se dibujan.
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
	const off = (p: ProductoPublico) =>
		p.anterior ? Math.round((1 - p.precio / p.anterior) * 100) : 0;
	const tarjeta = (p: ProductoPublico) => {
		const out = p.stock === 0;
		const sale = !!p.anterior;
		return {
			id: p.id,
			n: p.nombre,
			price: fmt(p.precio),
			img: img(p.foto, 640, 800),
			tag: out ? 'Agotado' : sale ? `−${off(p)}%` : p.etiqueta,
			out,
			op: out ? 0.4 : 1,
			flt: out ? 'grayscale(1)' : 'none',
			tbg: out ? '#2A2A2A' : sale ? '#FF4D2E' : p.etiqueta === 'Últimos' ? '#F2F2F2' : '#E6FF4D',
			tfg: out ? '#8A8A8A' : '#111',
			pc: sale ? '#FF4D2E' : '#F2F2F2',
			sub: out
				? 'Sin stock por ahora'
				: sale
					? `Antes ${fmt(p.anterior!)} · ${tienda.cobros.cuotas} cuotas`
					: `${tienda.cobros.cuotas} × ${fmt(p.precio / tienda.cobros.cuotas)} sin interés`,
			href: `${base}/producto/${p.id}`
		};
	};
	type Tarjeta = ReturnType<typeof tarjeta>;
	// Los títulos gigantes del diseño son cortos ("Drop 07"): con textos largos, el cuerpo se achica
	// para que la palabra más larga entre en su columna (ancho en cqw) sin cortarse por la mitad.
	const cabe = (texto: string, cqw: number) =>
		`${(cqw / (Math.max(1, ...texto.split(/\s+/).map((w) => w.length)) * 0.74)).toFixed(2)}cqw`;
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// Título de la portada: el segundo tramo va en amarillo, como en el diseño ("Drop / 07").
	const titulo = $derived(partirTitulo(tienda.portada.titulo));

	// Cinta que corre arriba de todo, con lo que la tienda ofrece de verdad.
	const cinta = $derived([
		...(tienda.envios.gratisDesde ? [`Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`] : []),
		...(tienda.envios.moto ? ['Envío en moto en el día'] : []),
		`${tienda.cobros.cuotas} cuotas sin interés`,
		`${tienda.cobros.transferencia}% off por transferencia`,
		...(tienda.envios.retiro ? ['Retiro en el local'] : [])
	]);
	const perks = $derived([
		tienda.envios.moto
			? { i: I.phMopedBold, t: 'Moto en el día', d: 'En CABA' }
			: { i: I.phTruckBold, t: 'Envíos', d: 'A todo el país' },
		{ i: I.phCreditCardBold, t: `${tienda.cobros.cuotas} cuotas`, d: 'Sin interés' },
		{ i: I.phBankBold, t: `${tienda.cobros.transferencia}% off`, d: 'Por transferencia' },
		tienda.envios.gratisDesde
			? { i: I.phTruckBold, t: 'Envío gratis', d: `Desde ${fmt(tienda.envios.gratisDesde)}` }
			: { i: I.phArrowsClockwiseBold, t: 'Arrepentimiento', d: '10 días para devolverlo' }
	]);
	const best = $derived(
		tienda.masVendidos.slice(0, 8).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
	);
	const destacado = $derived(
		tienda.masVendidos.map((id) => porId.get(id)).find((p) => p && p.stock > 0) ?? null
	);
	const maxOff = $derived(Math.max(0, ...tienda.productos.map(off)));

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
	const tabs = $derived([
		{ c: null, t: 'Todo', n: tienda.productos.length },
		...tienda.categorias.map((c) => ({ c: c.nombre, t: c.nombre, n: c.cantidad }))
	]);
	const tituloCat = $derived(categoria ? partirTitulo(categoria) : ['Todo el', 'catálogo']);

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
	let cp = $state('');
	let cotizado = $state<OpcionEnvio[] | null>(null);
	function calcular() {
		if (cpValido(cp)) cotizado = cotizar(cp, producto?.precio ?? subtotal, tienda);
	}
	function agregar() {
		if (!producto || producto.stock === 0 || tienda.vitrina) return;
		carrito.sumar(producto.id, 1, producto.stock);
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
			.slice(0, 4)
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
			d: `${tienda.cobros.transferencia}% off · alias al confirmar`,
			i: I.phBankBold
		},
		...(tienda.cobros.mercadoPago
			? [
					{
						id: 'mercadopago' as const,
						t: 'Mercado Pago',
						d: `Tarjetas en hasta ${tienda.cobros.cuotas} cuotas sin interés, dinero en cuenta`,
						i: I.siMercadopago
					}
				]
			: [])
	]);
	const pasos = $derived([
		{ t: 'Datos', on: true },
		{ t: 'Entrega', on: !!envio },
		{ t: 'Pago', on: !!envio }
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
				{ t: 'Cambios', href: `/t/${tienda.slug}/arrepentimiento` },
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
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link
		href="https://fonts.googleapis.com/css2?family=Archivo+Black&family=Archivo:wght@400;500;600;700;800&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet card(p: Tarjeta, cuadrada: boolean, sub: boolean)}
	<a
		href={p.href}
		class="celda"
		style="position:relative;display:block;border-right:2px solid #2A2A2A;border-bottom:2px solid #2A2A2A;padding:0;text-align:left;color:#F2F2F2;text-decoration:none;min-width:0"
	>
		<span
			class="zoom"
			style="position:relative;display:block;aspect-ratio:{cuadrada
				? '1'
				: '4/5'};overflow:hidden;background:#1E1E1E"
			>{#if p.img}<img
					src={p.img}
					alt={p.n}
					style="width:100%;height:100%;object-fit:cover;display:block;opacity:{p.op};filter:{p.flt}"
				/>{/if}{#if p.tag && !cuadrada}<span
					style="position:absolute;top:10px;left:10px;padding:5px 8px;background:{p.tbg};color:{p.tfg};font:800 11px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
					>{p.tag}</span
				>{/if}{#if p.out && !cuadrada}<span
					style="position:absolute;inset:0;display:grid;place-items:center"
					><span
						style="padding:10px 16px;border:2px solid #F2F2F2;background:rgba(17,17,17,.7);font:800 13px 'Archivo';letter-spacing:.14em;text-transform:uppercase;transform:rotate(-6deg)"
						>Agotado</span
					></span
				>{/if}</span
		>
		<span style="display:block;padding:12px 14px {sub ? '14px' : '12px'}"
			><span style="display:flex;justify-content:space-between;gap:10px;align-items:baseline"
				><span
					style="font-size:{sub
						? '14px'
						: '13.5px'};font-weight:700;text-transform:uppercase;letter-spacing:.02em;min-width:0"
					>{p.n}</span
				><b style="font-size:14.5px;white-space:nowrap;color:{p.pc}">{p.price}</b></span
			>{#if sub}<span style="display:block;margin-top:8px;font-size:12px;color:#8A8A8A"
					>{p.sub}</span
				>{/if}</span
		>
	</a>
{/snippet}

<div
	class="pampa"
	style="min-height:100vh;font-family:'Archivo',system-ui,sans-serif;color:#F2F2F2;background:#0B0B0B"
>
	<div
		style="container-type:inline-size;width:100%;overflow-x:clip;background:#111;position:relative"
	>
		<!-- cinta -->
		<div style="overflow:hidden;background:#E6FF4D;color:#111;border-bottom:2px solid #111">
			<div class="mq" style="display:flex;width:max-content">
				{#each [0, 1] as k (k)}<span style="display:flex;flex:none" aria-hidden={k === 1}
						>{#each cinta as t (t)}<span
								style="padding:9px 22px;font:800 12.5px 'Archivo';letter-spacing:.14em;text-transform:uppercase;white-space:nowrap"
								>{t}&nbsp;&nbsp;✦</span
							>{/each}</span
					>{/each}
			</div>
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:stretch;background:#111;border-bottom:2px solid #2A2A2A;height:58px"
		>
			<a
				href={base}
				style="display:flex;align-items:center;min-width:0;border-right:2px solid #2A2A2A;padding:0 clamp(14px,2.4cqw,28px);color:#F2F2F2;text-decoration:none;font-family:'Archivo Black',sans-serif;font-size:{nar
					? 19
					: 24}px;letter-spacing:-.02em;text-transform:uppercase"
				><span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
					>{tienda.nombre}</span
				></a
			>
			{#if wide}<nav style="display:flex">
					{#each nav as n (n.t)}<a
							href={n.href}
							style="display:flex;align-items:center;border-right:2px solid #2A2A2A;padding:0 20px;background:{n.on
								? '#F2F2F2'
								: 'transparent'};color:{n.on
								? '#111'
								: '#F2F2F2'};text-decoration:none;font:800 12.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase;white-space:nowrap"
							>{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			<a
				href="{base}/carrito"
				style="display:flex;align-items:center;gap:10px;flex:none;border-left:2px solid #2A2A2A;background:#E6FF4D;color:#111;padding:0 clamp(14px,2cqw,22px);text-decoration:none;font:800 13px 'Archivo';letter-spacing:.1em;text-transform:uppercase;white-space:nowrap"
				><Icon
					icon={I.phBagSimpleFill}
					size={18}
				/>Bolsa{#if carrito.unidades}&nbsp;({carrito.unidades}){/if}</a
			>
			{#if nar}<a
					href="{base}/catalogo"
					aria-label="Catálogo"
					style="display:grid;place-items:center;flex:none;width:58px;border-left:2px solid #2A2A2A;color:#F2F2F2"
					><Icon icon={I.phListBold} size={22} /></a
				>{/if}
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:10px 16px;background:#1A1A1A;border-bottom:2px solid #2A2A2A;color:#E6FF4D;font:800 12px 'Archivo';letter-spacing:.1em;text-transform:uppercase;text-align:center"
			>
				Es una tienda de ejemplo de la plantilla Pampa: podés recorrerla, pero no se compra
				de&nbsp;verdad.
			</div>
		{/if}

		<!-- ============ INICIO ============ -->
		{#if pagina === 'inicio'}
			{#if ver('hero')}
				<section
					style="position:relative;border-bottom:2px solid #2A2A2A;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,420px),1fr))"
				>
					<div
						style="padding:clamp(22px,4cqw,56px);display:flex;flex-direction:column;justify-content:space-between;gap:28px;min-width:0;border-right:2px solid #2A2A2A"
					>
						{#if tienda.portada.lema}<div
								style="display:flex;align-items:center;gap:10px;font-size:12.5px;font-weight:800;letter-spacing:.14em;text-transform:uppercase"
							>
								<span style="width:10px;height:10px;background:#FF4D2E;border-radius:50%;flex:none"
								></span>{tienda.portada.lema}
							</div>{/if}
						<h1
							style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(68px,13cqw,190px),{cabe(
								tienda.portada.titulo,
								wide ? 42 : 88
							)});line-height:.82;letter-spacing:-.045em;text-transform:uppercase;overflow-wrap:anywhere"
						>
							{titulo[0]}{#if titulo[1]}<br /><span style="color:#E6FF4D">{titulo[1]}</span>{/if}
						</h1>
						<div>
							<p style="margin:0;font-size:17px;line-height:1.5;color:#BDBDBD;max-width:420px">
								{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. {tienda.cobros
									.cuotas} cuotas sin interés o {tienda.cobros.transferencia}% off pagando
								por&nbsp;transferencia.
							</p>
							<div style="margin-top:22px;display:flex;gap:0;flex-wrap:wrap">
								<a
									href="{base}/catalogo"
									style="padding:18px 28px;border:2px solid #E6FF4D;background:#E6FF4D;color:#111;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none"
									>Ver todo →</a
								><a
									href="{base}/nosotros"
									style="padding:18px 24px;border:2px solid #F2F2F2;border-left-width:0;background:transparent;color:#F2F2F2;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none"
									>Nosotros</a
								>
							</div>
						</div>
					</div>
					<div style="position:relative;min-height:min(640px,120cqw);background:#1E1E1E">
						{#if tienda.portada.foto}<img
								src={img(tienda.portada.foto, 1200)}
								alt=""
								style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover;filter:grayscale(.25) contrast(1.05)"
							/>{/if}
						{#if maxOff}<span
								style="position:absolute;top:16px;right:16px;padding:8px 12px;background:#FF4D2E;color:#111;font:800 12px 'Archivo';letter-spacing:.12em;text-transform:uppercase;transform:rotate(3deg)"
								>Hasta −{maxOff}%</span
							>{/if}
						{#if destacado}<div
								style="position:absolute;left:0;right:0;bottom:0;display:flex;align-items:center;gap:12px;padding:14px 16px;background:#111;border-top:2px solid #2A2A2A"
							>
								{#if destacado.foto}<img
										src={img(destacado.foto, 120, 120)}
										alt=""
										style="width:46px;height:46px;object-fit:cover;flex:none"
									/>{/if}<span style="flex:1;min-width:0"
									><span style="display:block;font-size:14px;font-weight:700"
										>{destacado.nombre}</span
									><span style="display:block;font-size:12.5px;color:#8A8A8A"
										>{destacado.stock <= 3
											? destacado.stock === 1
												? 'Queda 1'
												: `Quedan ${destacado.stock}`
											: 'El más vendido'}</span
									></span
								><a
									href="{base}/producto/{destacado.id}"
									style="padding:10px 14px;border:0;background:#F2F2F2;color:#111;font:800 12px 'Archivo';letter-spacing:.1em;text-transform:uppercase;text-decoration:none;white-space:nowrap"
									>{fmt(destacado.precio)}</a
								>
							</div>{/if}
					</div>
				</section>
			{/if}

			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));border-bottom:2px solid #2A2A2A"
			>
				{#each perks as p (p.t)}<div
						style="display:flex;align-items:center;gap:12px;padding:16px clamp(14px,2.4cqw,28px);border-right:2px solid #2A2A2A"
					>
						<Icon icon={p.i} size={22} style="color:#E6FF4D;flex:none" />
						<div>
							<div
								style="font-size:12.5px;font-weight:800;letter-spacing:.1em;text-transform:uppercase"
							>
								{p.t}
							</div>
							<div style="font-size:12.5px;color:#8A8A8A;margin-top:2px">{p.d}</div>
						</div>
					</div>{/each}
			</div>

			{#if ver('cats') && tienda.categorias.length}
				<div
					style="display:flex;overflow-x:auto;scrollbar-width:none;border-bottom:2px solid #2A2A2A"
				>
					{#each tabs as t (t.c ?? '')}<a
							href={enlaceCat(t.c)}
							style="flex:none;padding:14px 18px;border-right:2px solid #2A2A2A;color:#F2F2F2;text-decoration:none;font:800 12.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase;white-space:nowrap"
							>{t.t} <span style="opacity:.55">{t.n}</span></a
						>{/each}
				</div>
			{/if}

			{#if ver('prods') && best.length}
				<section style="padding:clamp(40px,6cqw,80px) 0 0">
					<div
						style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap;padding:0 clamp(14px,2.4cqw,28px) 18px"
					>
						<h2
							style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(40px,6.4cqw,88px);line-height:.85;letter-spacing:-.04em;text-transform:uppercase"
						>
							Lo más pedido
						</h2>
						<a
							href="{base}/catalogo"
							style="font-size:12.5px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A;text-decoration:none"
							>{tienda.productos.length}
							{tienda.productos.length === 1 ? 'producto' : 'productos'} →</a
						>
					</div>
					<div
						style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(50%,260px),1fr));border-top:2px solid #2A2A2A"
					>
						{#each best as p (p.id)}{@render card(p, false, true)}{/each}
					</div>
				</section>
			{/if}

			{#if ver('revs') && tienda.resenas.length}
				<section style="padding:clamp(40px,6cqw,80px) clamp(14px,2.4cqw,28px) 0">
					<div
						style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
					>
						<h2
							style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(40px,6.4cqw,88px);line-height:.85;letter-spacing:-.04em;text-transform:uppercase"
						>
							Lo que dicen
						</h2>
						{#if tienda.promedio}<span
								style="display:flex;align-items:center;gap:8px;font-size:14px;font-weight:700"
								><span style="display:flex;color:#E6FF4D"
									>{#each estrellas(tienda.promedio) as e, i (i)}<Icon
											icon={e}
											size={15}
										/>{/each}</span
								>{tienda.promedio.toLocaleString('es-AR')} · {tienda.resenas.length}
								{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}</span
							>{/if}
					</div>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,300px),1fr));border:2px solid #2A2A2A"
					>
						{#each tienda.resenas.slice(0, 3) as r, i (i)}
							{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
							<div style="display:flex;gap:14px;padding:16px;border-right:2px solid #2A2A2A">
								{#if pr?.foto}<img
										src={img(pr.foto, 200, 250)}
										alt=""
										style="width:92px;height:116px;object-fit:cover;flex:none;background:#1E1E1E"
									/>{/if}
								<div style="min-width:0">
									<div
										style="font-size:12px;font-weight:800;letter-spacing:.1em;text-transform:uppercase;color:#E6FF4D"
									>
										{r.producto}
									</div>
									<p style="margin:8px 0 0;font-size:15px;line-height:1.4;font-weight:500">
										“{r.texto}”
									</p>
									<div style="margin-top:8px;font-size:12.5px;color:#8A8A8A">{r.autor}</div>
								</div>
							</div>
						{/each}
					</div>
				</section>
			{/if}

			{#if ver('ig') && tienda.fotos.length >= 3}
				<section style="padding:clamp(40px,6cqw,80px) clamp(14px,2.4cqw,28px)">
					<h2
						style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(40px,6.4cqw,88px);line-height:.85;letter-spacing:-.04em;text-transform:uppercase"
					>
						En la calle
					</h2>
					<div
						style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,170px),1fr));gap:2px;background:#2A2A2A;border:2px solid #2A2A2A"
					>
						{#each tienda.fotos as g (g)}<div
								style="aspect-ratio:1;overflow:hidden;background:#1E1E1E"
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
				<div style="height:clamp(40px,6cqw,80px)"></div>
			{/if}
		{/if}

		<!-- ============ CATÁLOGO ============ -->
		{#if pagina === 'catalogo'}
			<section
				style="padding:clamp(26px,4cqw,48px) clamp(14px,2.4cqw,28px) 0;border-bottom:2px solid #2A2A2A"
			>
				<h1
					style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(56px,10cqw,150px),{cabe(
						tituloCat.join(' '),
						92
					)});line-height:.82;letter-spacing:-.045em;text-transform:uppercase;overflow-wrap:anywhere"
				>
					{tituloCat[0]}{#if tituloCat[1]}<br /><span style="color:#E6FF4D">{tituloCat[1]}</span
						>{/if}
				</h1>
				<div
					style="margin-top:22px;display:flex;overflow-x:auto;scrollbar-width:none;border-top:2px solid #2A2A2A;margin-left:calc(-1 * clamp(14px,2.4cqw,28px));margin-right:calc(-1 * clamp(14px,2.4cqw,28px))"
				>
					{#each tabs as t (t.c ?? '')}{@const on = t.c === categoria}<a
							href={enlaceCat(t.c)}
							style="flex:none;padding:14px 18px;border-right:2px solid #2A2A2A;background:{on
								? '#F2F2F2'
								: 'transparent'};color:{on
								? '#111'
								: '#F2F2F2'};text-decoration:none;font:800 12.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase;white-space:nowrap"
							>{t.t} <span style="opacity:.55">{t.n}</span></a
						>{/each}
				</div>
			</section>
			<div style="display:flex;flex-wrap:wrap;align-items:stretch;border-bottom:2px solid #2A2A2A">
				<span
					style="display:flex;align-items:center;gap:8px;padding:12px 16px;border-right:2px solid #2A2A2A;font:800 12px 'Archivo';letter-spacing:.12em;text-transform:uppercase"
					>{listado.length} {listado.length === 1 ? 'producto' : 'productos'}</span
				>
				<span style="flex:1"></span>
				<label
					style="position:relative;display:flex;align-items:center;gap:6px;padding:12px 16px;border-left:2px solid #2A2A2A;font-size:13px;font-weight:700"
					>Ordenar:<select
						value={orden}
						onchange={(e) => goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
						style="appearance:none;border:0;background:transparent;padding:0 18px 0 0;font:700 13px 'Archivo';color:#F2F2F2;cursor:pointer"
						>{#each ORDENES as [k, t] (k)}<option value={k} style="background:#111">{t}</option
							>{/each}</select
					><Icon
						icon={I.phCaretDownBold}
						size={12}
						style="position:absolute;right:16px;pointer-events:none"
					/></label
				>
			</div>
			{#if catalog.length}
				<div
					style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(50%,240px),1fr))"
				>
					{#each catalog as p (p.id)}{@render card(p, false, false)}{/each}
				</div>
				{#if listado.length > mostrar}
					<div style="display:flex;justify-content:center;padding:28px">
						<button
							type="button"
							onclick={() => (mostrar += 24)}
							style="padding:16px 26px;border:2px solid #F2F2F2;background:transparent;color:#F2F2F2;font:800 13px 'Archivo';letter-spacing:.12em;text-transform:uppercase;cursor:pointer"
							>Cargar 24 más</button
						>
					</div>
				{:else}
					<div style="height:28px"></div>
				{/if}
			{:else}
				<p style="margin:0;padding:40px clamp(14px,2.4cqw,28px);font-size:15px;color:#8A8A8A">
					Todavía no hay productos en esta categoría.
				</p>
			{/if}
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto}
			<section
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,420px),1fr));border-bottom:2px solid #2A2A2A"
			>
				<div style="background:#2A2A2A;border-right:2px solid #2A2A2A">
					<div style="aspect-ratio:4/5;background:#1E1E1E;overflow:hidden">
						{#if producto.foto}<img
								src={img(producto.foto, 1100)}
								alt={producto.nombre}
								style="width:100%;height:100%;object-fit:cover"
							/>{/if}
					</div>
				</div>
				<div
					style="padding:clamp(20px,3.4cqw,44px);position:sticky;top:58px;align-self:start;min-width:0"
				>
					<div style="display:flex;gap:6px;flex-wrap:wrap">
						{#if producto.categoria}<a
								href={enlaceCat(producto.categoria)}
								style="padding:5px 8px;background:#E6FF4D;color:#111;font:800 11px 'Archivo';letter-spacing:.1em;text-transform:uppercase;text-decoration:none"
								>{producto.categoria}</a
							>{/if}{#if producto.anterior}<span
								style="padding:5px 8px;background:#FF4D2E;color:#111;font:800 11px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
								>−{off(producto)}%</span
							>{:else if producto.etiqueta}<span
								style="padding:5px 8px;background:#F2F2F2;color:#111;font:800 11px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
								>{producto.etiqueta}</span
							>{/if}{#if producto.stock > 0 && producto.stock <= 3}<span
								style="padding:5px 8px;border:1.5px solid #FF4D2E;color:#FF4D2E;font:800 11px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
								>{producto.stock === 1 ? 'Queda 1' : `Quedan ${producto.stock}`}</span
							>{/if}
					</div>
					<h1
						style="margin:14px 0 0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(40px,5.4cqw,72px),{cabe(
							producto.nombre,
							wide ? 40 : 86
						)});line-height:.86;letter-spacing:-.04em;text-transform:uppercase;overflow-wrap:anywhere"
					>
						{producto.nombre}
					</h1>
					<div style="margin-top:16px;display:flex;align-items:baseline;gap:12px;flex-wrap:wrap">
						<b style="font-family:'Archivo Black',sans-serif;font-size:32px;font-weight:400"
							>{fmt(producto.precio)}</b
						>{#if producto.anterior}<s style="font-size:17px;color:#8A8A8A"
								>{fmt(producto.anterior)}</s
							>{/if}
					</div>
					<div style="margin-top:6px;font-size:14px;color:#BDBDBD">
						{tienda.cobros.cuotas} × {fmt(producto.precio / tienda.cobros.cuotas)} sin interés ·
						<b style="color:#B9F27C"
							>{fmt(producto.precio * (1 - tienda.cobros.transferencia / 100))} por transferencia</b
						>
					</div>
					<button
						type="button"
						onclick={agregar}
						disabled={producto.stock === 0 || tienda.vitrina}
						style="margin-top:20px;width:100%;padding:19px;border:0;background:#E6FF4D;color:#111;font:800 14.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase;cursor:pointer"
						>{producto.stock === 0 ? 'Agotado' : 'Sumar a la bolsa'}</button
					>
					{#if whatsapp}<a
							href={wa(`¡Hola! Quiero comprar ${producto.nombre} (${fmt(producto.precio)}).`)}
							target="_blank"
							rel="noopener"
							style="margin-top:8px;display:flex;justify-content:center;align-items:center;gap:8px;padding:16px;border:2px solid #F2F2F2;color:#F2F2F2;text-decoration:none;font:800 13px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
							><Icon icon={I.phWhatsappLogoBold} size={18} />Comprar por WhatsApp</a
						>{/if}
					<div style="margin-top:18px;border:2px solid #2A2A2A">
						<div style="display:flex">
							<label for="cp-ficha" class="oculto">Código postal</label><input
								id="cp-ficha"
								bind:value={cp}
								inputmode="numeric"
								maxlength="4"
								placeholder="Código postal"
								onkeydown={(e) => e.key === 'Enter' && calcular()}
								style="flex:1;min-width:0;padding:13px 14px;border:0;background:transparent;font:400 14px 'Archivo';color:#F2F2F2;outline:none"
							/><button
								type="button"
								onclick={calcular}
								style="padding:13px 16px;border:0;background:#F2F2F2;color:#111;font:800 12px 'Archivo';letter-spacing:.1em;text-transform:uppercase;cursor:pointer"
								>Calcular</button
							>
						</div>
						{#if cotizado}<div
								style="border-top:2px solid #2A2A2A;padding:12px 14px;display:flex;flex-direction:column;gap:8px;font-size:13.5px"
							>
								{#each cotizado as o (o.id)}<span
										style="display:flex;justify-content:space-between;gap:10px"
										><span>{o.nombre} · {o.detalle}</span><b
											style="white-space:nowrap;color:{o.precio ? '#F2F2F2' : '#B9F27C'}"
											>{o.precio ? fmt(o.precio) : 'Gratis'}</b
										></span
									>{/each}
							</div>{/if}
					</div>
					<div style="margin-top:18px;border:2px solid #2A2A2A">
						{#each acc as a, i (a.t)}<div
								style="padding:13px 14px;{i ? 'border-top:2px solid #2A2A2A' : ''}"
							>
								<button
									type="button"
									onclick={() => (abierto = a.open ? '' : a.t)}
									aria-expanded={a.open}
									style="width:100%;display:flex;justify-content:space-between;align-items:center;gap:10px;padding:0;border:0;background:transparent;font:800 12.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase;color:#F2F2F2;cursor:pointer;text-align:left"
									>{a.t}<Icon icon={a.i} size={13} /></button
								>{#if a.open}<p
										style="margin:10px 0 0;font-size:14px;line-height:1.6;color:#BDBDBD"
									>
										{a.d}
									</p>{/if}
							</div>{/each}
					</div>
				</div>
			</section>
			{#if related.length}
				<section style="padding:clamp(36px,5cqw,64px) 0 0">
					<h2
						style="margin:0 clamp(14px,2.4cqw,28px) 18px;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(34px,5cqw,64px);line-height:.85;letter-spacing:-.04em;text-transform:uppercase"
					>
						Va con
					</h2>
					<div
						style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(50%,240px),1fr));border-top:2px solid #2A2A2A"
					>
						{#each related as p (p.id)}{@render card(p, true, false)}{/each}
					</div>
				</section>
			{/if}
			<div style="height:clamp(36px,5cqw,64px)"></div>
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));border-bottom:2px solid #2A2A2A"
			>
				<div style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A;min-width:0">
					<h1
						style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(48px,7cqw,100px);line-height:.85;letter-spacing:-.045em;text-transform:uppercase"
					>
						Bolsa ({carrito.unidades})
					</h1>
					{#if !lineas.length}
						<p style="margin:18px 0 0;font-size:15px;color:#8A8A8A">Todavía no sumaste nada.</p>
						<a
							href="{base}/catalogo"
							style="display:inline-block;margin-top:18px;padding:18px 28px;background:#E6FF4D;color:#111;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none"
							>Ver todo →</a
						>
					{:else}
						{#if tienda.envios.gratisDesde}
							<div
								style="margin-top:16px;padding:12px 14px;border:2px solid #2A2A2A;font-size:13.5px"
							>
								<div style="display:flex;justify-content:space-between;gap:10px">
									<span
										>{#if falta}Te faltan <b>{fmt(falta)}</b> para envío gratis{:else}<b
												>Tenés envío gratis</b
											>{/if}</span
									><span style="color:#8A8A8A;white-space:nowrap"
										>{fmt(tienda.envios.gratisDesde)}</span
									>
								</div>
								<div style="margin-top:8px;height:6px;background:#2A2A2A">
									<div
										style="width:{Math.min(
											100,
											(subtotal / tienda.envios.gratisDesde) * 100
										)}%;height:100%;background:#E6FF4D"
									></div>
								</div>
							</div>
						{/if}
						<div style="margin-top:14px;border-top:2px solid #2A2A2A">
							{#each lineas as c (c.id)}<div
									style="display:flex;gap:14px;padding:14px 0;border-bottom:2px solid #2A2A2A"
								>
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 200, 240)}
											alt=""
											style="width:92px;height:110px;object-fit:cover;flex:none;background:#1E1E1E"
										/>{/if}
									<div style="flex:1;min-width:0;display:flex;flex-direction:column">
										<div style="display:flex;justify-content:space-between;gap:10px">
											<a
												href="{base}/producto/{c.id}"
												style="font-size:14px;font-weight:800;text-transform:uppercase;color:#F2F2F2;text-decoration:none"
												>{c.producto.nombre}</a
											><b style="white-space:nowrap">{fmt(c.subtotal)}</b>
										</div>
										<span style="font-size:13px;color:#8A8A8A;margin-top:4px"
											>{fmt(c.producto.precio)} c/u</span
										>
										<span style="flex:1;min-height:10px"></span>
										<div style="display:flex;justify-content:space-between;align-items:center">
											<span style="display:flex;border:2px solid #2A2A2A"
												><button
													type="button"
													aria-label="Uno menos"
													onclick={() => carrito.fijar(c.id, c.cantidad - 1)}
													style="width:34px;height:32px;display:grid;place-items:center;border:0;background:transparent;color:#F2F2F2;cursor:pointer"
													><Icon icon={I.phMinusBold} size={12} /></button
												><b style="width:26px;display:grid;place-items:center;font-size:14px"
													>{c.cantidad}</b
												><button
													type="button"
													aria-label="Uno más"
													disabled={c.cantidad >= c.producto.stock}
													onclick={() => carrito.fijar(c.id, c.cantidad + 1)}
													style="width:34px;height:32px;display:grid;place-items:center;border:0;background:transparent;color:#F2F2F2;cursor:pointer"
													><Icon icon={I.phPlusBold} size={12} /></button
												></span
											><button
												type="button"
												onclick={() => carrito.fijar(c.id, 0)}
												style="padding:0;border:0;background:transparent;font:800 12px 'Archivo';letter-spacing:.1em;text-transform:uppercase;color:#8A8A8A;cursor:pointer"
												>Quitar</button
											>
										</div>
									</div>
								</div>{/each}
						</div>
					{/if}
				</div>
				{#if lineas.length}
					<div
						style="padding:clamp(22px,3.4cqw,44px);background:#161616;display:flex;flex-direction:column;gap:12px;font-size:14px;min-width:0"
					>
						<div
							style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
						>
							Resumen
						</div>
						<span style="display:flex;justify-content:space-between"
							><span>Subtotal</span><span>{fmt(subtotal)}</span></span
						>
						<span style="display:flex;justify-content:space-between;gap:10px;color:#8A8A8A"
							><span>Envío</span><span>Se calcula en el próximo paso</span></span
						>
						<div style="height:2px;background:#2A2A2A"></div>
						<span style="display:flex;justify-content:space-between;align-items:baseline"
							><b style="text-transform:uppercase;letter-spacing:.08em">Total</b><b
								style="font-family:'Archivo Black',sans-serif;font-size:30px;font-weight:400"
								>{fmt(subtotal)}</b
							></span
						>
						<span style="color:#BDBDBD"
							>{tienda.cobros.cuotas} × {fmt(subtotal / tienda.cobros.cuotas)} sin interés</span
						>
						<span style="color:#B9F27C;font-weight:700"
							>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} por transferencia</span
						>
						<a
							href="{base}/checkout"
							style="margin-top:6px;padding:19px;background:#E6FF4D;color:#111;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-align:center;text-decoration:none"
							>Pagar →</a
						>
						{#if upsell.length}
							<div
								style="margin-top:12px;font-size:12.5px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A"
							>
								Sumale
							</div>
							<div
								style="display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:2px;background:#2A2A2A;border:2px solid #2A2A2A"
							>
								{#each upsell as p (p.id)}<button
										type="button"
										aria-label="Sumar {p.nombre} a la bolsa"
										onclick={() => carrito.sumar(p.id, 1, p.stock)}
										style="display:block;border:0;background:#111;padding:8px;text-align:left;color:#F2F2F2;cursor:pointer;font-family:inherit;min-width:0"
										><span
											style="position:relative;display:block;aspect-ratio:1;overflow:hidden;background:#1E1E1E"
											>{#if p.foto}<img
													src={img(p.foto, 300, 300)}
													alt=""
													style="width:100%;height:100%;object-fit:cover"
												/>{/if}<span
												style="position:absolute;right:6px;bottom:6px;width:28px;height:28px;display:grid;place-items:center;background:#E6FF4D;color:#111"
												><Icon icon={I.phPlusBold} size={13} /></span
											></span
										><span
											style="margin-top:8px;display:flex;justify-content:space-between;gap:6px;font-size:12.5px;font-weight:700;text-transform:uppercase"
											><span style="min-width:0">{p.nombre}</span><span style="white-space:nowrap"
												>{fmt(p.precio)}</span
											></span
										></button
									>{/each}
							</div>
						{/if}
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ CHECKOUT ============ -->
		{#if pagina === 'checkout'}
			{#if !lineas.length}
				<section style="padding:clamp(22px,3.4cqw,44px);border-bottom:2px solid #2A2A2A">
					<p style="margin:0;font-size:15px;color:#8A8A8A">Tu bolsa está vacía.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-block;margin-top:16px;color:#E6FF4D;font-weight:700">Ver todo</a
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
					style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));border-bottom:2px solid #2A2A2A"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<div
						style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A;display:flex;flex-direction:column;gap:24px;min-width:0"
					>
						<div style="display:flex;border:2px solid #2A2A2A">
							{#each pasos as s, i (s.t)}<span
									style="flex:1;padding:11px 8px;text-align:center;{i < pasos.length - 1
										? 'border-right:2px solid #2A2A2A;'
										: ''}background:{s.on ? '#F2F2F2' : 'transparent'};color:{s.on
										? '#111'
										: '#8A8A8A'};font:800 11.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase"
									>{s.t}</span
								>{/each}
						</div>
						<div>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
							>
								Tus datos
							</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:2px;background:#2A2A2A;border:2px solid #2A2A2A"
							>
								{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'Teléfono', 'tel', 'tel'], ['dni', 'DNI', 'text', 'off']] as [n, l, tipo, ac] (n)}
									<label style="display:block;padding:11px 13px;background:#111"
										><span
											style="display:block;font-size:11px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A"
											>{l}</span
										><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required={n !== 'dni'}
											maxlength="120"
											style="width:100%;margin-top:3px;padding:0;border:0;background:transparent;font:400 15px 'Archivo';color:#F2F2F2;outline:none"
										/></label
									>
								{/each}
							</div>
						</div>
						<div>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
							>
								Entrega
							</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:minmax(0,130px) minmax(0,1fr);gap:2px;background:#2A2A2A;border:2px solid #2A2A2A"
							>
								<label style="display:block;padding:11px 13px;background:#111"
									><span
										style="display:block;font-size:11px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A"
										>Código postal</span
									><input
										name="cp"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										style="width:100%;margin-top:3px;padding:0;border:0;background:transparent;font:400 15px 'Archivo';color:#F2F2F2;outline:none"
									/></label
								>
								<label style="display:block;padding:11px 13px;background:#111"
									><span
										style="display:block;font-size:11px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required={envio !== 'retiro'}
										style="width:100%;margin-top:3px;padding:0;border:0;background:transparent;font:400 15px 'Archivo';color:#F2F2F2;outline:none"
									/></label
								>
							</div>
							<div
								style="margin-top:8px;display:flex;flex-direction:column;border:2px solid #2A2A2A;border-bottom:0"
								role="radiogroup"
								aria-label="Forma de entrega"
							>
								{#each opciones as s (s.id)}{@const on = envio === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (envio = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border:0;border-bottom:2px solid #2A2A2A;background:{on
											? '#E6FF4D'
											: '#111'};color:{on
											? '#111'
											: '#F2F2F2'};text-align:left;cursor:pointer;font-family:inherit"
										><span
											style="width:18px;height:18px;border:2px solid currentColor;display:grid;place-items:center;flex:none"
											>{#if on}<span style="width:8px;height:8px;background:currentColor"
												></span>{/if}</span
										><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700">{s.nombre}</span
											><span style="display:block;font-size:12.5px;opacity:.7">{s.detalle}</span
											></span
										><b style="white-space:nowrap">{s.precio ? fmt(s.precio) : 'Gratis'}</b></button
									>{:else}<span
										style="padding:15px;border-bottom:2px solid #2A2A2A;font-size:13.5px;color:#8A8A8A"
										>Poné tu código postal para ver cómo te llega.</span
									>{/each}
							</div>
						</div>
						<div>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
							>
								Pago
							</div>
							<div
								style="margin-top:12px;display:flex;flex-direction:column;border:2px solid #2A2A2A;border-bottom:0"
								role="radiogroup"
								aria-label="Forma de pago"
							>
								{#each pagos as s (s.id)}{@const on = pago === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (pago = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border:0;border-bottom:2px solid #2A2A2A;background:{on
											? '#E6FF4D'
											: '#111'};color:{on
											? '#111'
											: '#F2F2F2'};text-align:left;cursor:pointer;font-family:inherit"
										><span
											style="width:18px;height:18px;border:2px solid currentColor;display:grid;place-items:center;flex:none"
											>{#if on}<span style="width:8px;height:8px;background:currentColor"
												></span>{/if}</span
										><Icon icon={s.i} size={22} style="flex:none" /><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700">{s.t}</span
											><span style="display:block;font-size:12.5px;opacity:.7">{s.d}</span></span
										></button
									>{/each}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;border:2px solid #FF4D2E;color:#FF4D2E;font-size:14px;font-weight:600"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || enviando || tienda.vitrina}
							style="padding:20px;border:0;background:#E6FF4D;color:#111;font:800 15px 'Archivo';letter-spacing:.12em;text-transform:uppercase;cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
							><Icon icon={I.phLockSimpleFill} size={16} />{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: envio
									? `Pagar ${fmt(total)}`
									: 'Elegí cómo te llega'}</button
						>
					</div>
					<div style="padding:clamp(22px,3.4cqw,44px);background:#161616;min-width:0">
						<div
							style="font-family:'Archivo Black',sans-serif;font-size:26px;text-transform:uppercase"
						>
							Tu pedido
						</div>
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px">
							{#each lineas as c (c.id)}<div style="display:flex;gap:12px;align-items:center">
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 120, 140)}
											alt=""
											style="width:56px;height:66px;object-fit:cover;flex:none;background:#1E1E1E"
										/>{/if}<span
										style="flex:1;min-width:0;font-size:13.5px;font-weight:700;text-transform:uppercase"
										>{c.producto.nombre}<br /><span
											style="font-weight:500;text-transform:none;color:#8A8A8A"
											>{c.cantidad} × {fmt(c.producto.precio)}</span
										></span
									><b style="font-size:14px;white-space:nowrap">{fmt(c.subtotal)}</b>
								</div>{/each}
						</div>
						<div
							style="margin-top:16px;padding-top:14px;border-top:2px solid #2A2A2A;display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							><span style="display:flex;justify-content:space-between"
								><span>{elegida ? elegida.nombre : 'Envío'}</span><span
									>{elegida ? (elegida.precio ? fmt(elegida.precio) : 'Gratis') : '—'}</span
								></span
							>{#if descuento}<span
									style="display:flex;justify-content:space-between;color:#B9F27C;font-weight:700"
									><span>Transferencia −{tienda.cobros.transferencia}%</span><span
										>− {fmt(descuento)}</span
									></span
								>{/if}<span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b style="text-transform:uppercase">Total</b><b
									style="font-family:'Archivo Black',sans-serif;font-size:28px;font-weight:400"
									>{fmt(total)}</b
								></span
							>
						</div>
						<div
							style="margin-top:16px;display:flex;align-items:center;gap:8px;font-size:12.5px;color:#8A8A8A"
						>
							<Icon icon={I.phShieldCheckBold} size={16} />Compra protegida · 10 días para
							arrepentirte
						</div>
					</div>
				</form>
			{/if}
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas del checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));border-bottom:2px solid #2A2A2A"
			>
				<div
					style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A;display:flex;flex-direction:column;gap:20px;min-width:0"
				>
					<div
						style="display:flex;align-items:center;gap:10px;font-size:12.5px;font-weight:800;letter-spacing:.14em;text-transform:uppercase;color:#E6FF4D"
					>
						<Icon icon={I.phCheckBold} size={16} />Pedido #{pedido.numero}
					</div>
					<h1
						style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(48px,7cqw,100px),{cabe(
							`Gracias, ${pedido.nombre}!`,
							wide ? 42 : 86
						)});line-height:.85;letter-spacing:-.045em;text-transform:uppercase;overflow-wrap:anywhere"
					>
						¡Gracias,<br /><span style="color:#E6FF4D">{pedido.nombre}!</span>
					</h1>
					<p style="margin:0;font-size:17px;line-height:1.5;color:#BDBDBD;max-width:460px">
						Recibimos tu pedido por <b style="color:#F2F2F2">{fmt(pedido.total)}</b>. Te escribimos
						por mail con cada&nbsp;novedad.
					</p>
					<div>
						<a
							href={base}
							style="display:inline-block;padding:18px 28px;background:#E6FF4D;color:#111;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none"
							>Volver a la tienda</a
						>
					</div>
				</div>
				{#if pedido.pago === 'transferencia'}
					<div
						style="padding:clamp(22px,3.4cqw,44px);background:#161616;font-size:15px;line-height:1.6;min-width:0"
					>
						<div
							style="font-family:'Archivo Black',sans-serif;font-size:26px;text-transform:uppercase"
						>
							Transferí {fmt(pedido.total)}
						</div>
						<p style="margin:12px 0 0;color:#BDBDBD">
							{#if pedido.alias}Al alias <b style="color:#E6FF4D">{pedido.alias}</b>{:else}Con los
								datos que te mandamos por mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
									href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
									target="_blank"
									rel="noopener"
									style="color:#E6FF4D">por WhatsApp</a
								>{/if}. Lo preparamos apenas se acredite.
						</p>
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ NOSOTROS ============ -->
		{#if pagina === 'nosotros'}
			<section
				style="position:relative;min-height:clamp(420px,56cqw,720px);border-bottom:2px solid #2A2A2A;overflow:hidden;background:#1E1E1E"
			>
				{#if tienda.portada.foto}<img
						src={img(tienda.portada.foto, 1800)}
						alt=""
						style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover;filter:grayscale(1) contrast(1.1);opacity:.6"
					/>{/if}
				<div
					style="position:absolute;left:clamp(14px,2.4cqw,28px);right:clamp(14px,2.4cqw,28px);bottom:clamp(20px,3cqw,40px)"
				>
					{#if tienda.ciudad}<div
							style="font-size:12.5px;font-weight:800;letter-spacing:.14em;text-transform:uppercase;color:#E6FF4D"
						>
							{tienda.ciudad}
						</div>{/if}
					<h1
						style="margin:10px 0 0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(50px,9cqw,136px),{cabe(
							tienda.portada.titulo,
							92
						)});line-height:.84;letter-spacing:-.045em;text-transform:uppercase;overflow-wrap:anywhere"
					>
						{titulo[0]}{#if titulo[1]}<br />{titulo[1]}{/if}
					</h1>
				</div>
			</section>
			<section
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));border-bottom:2px solid #2A2A2A"
			>
				<div style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A">
					<p style="margin:0;font-size:clamp(19px,2.2cqw,26px);line-height:1.4;font-weight:500">
						{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. Comprás online y te lo
						mandamos a todo el país{tienda.envios.retiro ? ', o lo retirás en el local' : ''}.
					</p>
				</div>
				<div style="display:grid;grid-template-columns:repeat(3,minmax(0,1fr))">
					{#each [{ v: String(tienda.productos.length), t: 'Productos' }, { v: String(tienda.cobros.cuotas), t: 'Cuotas sin interés' }, ...(tienda.promedio ? [{ v: tienda.promedio.toLocaleString('es-AR'), t: 'En reseñas' }] : [{ v: `${tienda.cobros.transferencia}%`, t: 'Off por transferencia' }])] as s (s.t)}<div
							style="padding:clamp(18px,2.6cqw,32px);border-right:2px solid #2A2A2A;display:flex;flex-direction:column;justify-content:flex-end;min-width:0"
						>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:clamp(34px,4.4cqw,58px);line-height:.9;color:#E6FF4D"
							>
								{s.v}
							</div>
							<div
								style="margin-top:6px;font-size:12px;font-weight:800;letter-spacing:.1em;text-transform:uppercase;color:#8A8A8A"
							>
								{s.t}
							</div>
						</div>{/each}
				</div>
			</section>
			{#if tienda.direccion || whatsapp}
				<section
					style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));border-bottom:2px solid #2A2A2A"
				>
					{#if tienda.direccion}
						<div
							style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A;display:flex;flex-direction:column;gap:14px"
						>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:32px;text-transform:uppercase"
							>
								{tienda.envios.retiro ? 'Showroom' : 'Dónde estamos'}
							</div>
							<div style="display:flex;flex-direction:column;gap:12px;font-size:15px">
								<span style="display:flex;gap:12px"
									><Icon
										icon={I.phMapPinBold}
										size={19}
										style="color:#E6FF4D;flex:none"
									/>{tienda.direccion}</span
								>{#if tienda.envios.retiro}<span style="display:flex;gap:12px"
										><Icon
											icon={I.phStorefrontBold}
											size={19}
											style="color:#E6FF4D;flex:none"
										/>Retiro sin costo</span
									>{/if}
							</div>
						</div>
					{/if}
					{#if whatsapp}
						<div style="padding:clamp(22px,3.4cqw,44px);display:flex;flex-direction:column;gap:2px">
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:32px;text-transform:uppercase;margin-bottom:12px"
							>
								Hablemos
							</div>
							<a
								href={wa('¡Hola!')}
								target="_blank"
								rel="noopener"
								style="display:flex;align-items:center;gap:12px;padding:16px;background:#25D366;color:#111;text-decoration:none;font:800 13px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
								><Icon icon={I.phWhatsappLogoFill} size={22} />{tienda.whatsapp}</a
							>
						</div>
					{/if}
				</section>
			{/if}
		{/if}

		<!-- footer -->
		<footer style="padding:clamp(28px,4cqw,48px) clamp(14px,2.4cqw,28px) 22px">
			<div
				aria-hidden="true"
				style="font-family:'Archivo Black',sans-serif;font-size:clamp(80px,22cqw,320px);line-height:.78;letter-spacing:-.06em;color:#1C1C1C;user-select:none;text-transform:uppercase;overflow-wrap:anywhere"
			>
				{tienda.nombre}
			</div>
			<div
				style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,180px),1fr));gap:24px"
			>
				{#each foot as f (f.t)}<div
						style="display:flex;flex-direction:column;gap:8px;font-size:14px"
					>
						<b
							style="font-size:12px;font-weight:800;letter-spacing:.14em;text-transform:uppercase;color:#E6FF4D;margin-bottom:4px"
							>{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:#BDBDBD;text-decoration:none"
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
				style="margin-top:30px;padding-top:14px;border-top:2px solid #2A2A2A;display:flex;justify-content:space-between;gap:14px;flex-wrap:wrap;font-size:12px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#6E6E6E"
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
				style="position:fixed;right:18px;bottom:18px;z-index:30;width:56px;height:56px;background:#25D366;color:#111;display:grid;place-items:center;border:2px solid #111;box-shadow:4px 4px 0 #E6FF4D"
				><Icon icon={I.phWhatsappLogoFill} size={28} /></a
			>{/if}
	</div>
</div>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.mq {
		animation: mq 28s linear infinite;
	}
	@keyframes mq {
		to {
			transform: translateX(-50%);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.mq {
			animation: none;
		}
	}
	.celda {
		background: #111;
	}
	.celda:hover {
		background: #161616;
	}
	.zoom :global(img) {
		transition: transform 0.7s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.celda:hover .zoom :global(img) {
		transform: scale(1.05);
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
