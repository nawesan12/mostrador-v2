<script lang="ts">
	import { enhance } from '$app/forms';
	import ChipEstado from '#lib/components/ChipEstado.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { cara, unsplash } from '#lib/landing/formato.ts';
	import MetaMes from '#lib/panel/MetaMes.svelte';
	import { anim } from '#lib/panel/anim.ts';
	import { compartirTienda } from '#lib/panel/compartir.ts';
	import {
		curva,
		diaCorto,
		diferencia,
		etiquetaEje,
		fechaLarga,
		haceCorto,
		horaMin,
		mes,
		pesos,
		pesosCortos,
		porcentaje,
		puntos,
		saludo,
		variacion
	} from '#lib/panel/formato.ts';
	import { rutaDe, visible } from '#lib/panel/navegacion.ts';
	import { getShell } from '#lib/panel/shell.svelte.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Resumen del panel (design/Mostrador v4.dc.html, `isDash`), con los datos de la tienda. */
	let { data } = $props();
	const shell = getShell();
	const ui = getUi();
	const t = $derived(data.tablero);
	/**
	 * "Ahora en tu tienda" y las acreditaciones sólo existen en la demo: la API real no las mide, así
	 * que ahí no se muestra un cero que parezca un dato. Las visitas sí: las cuenta Tu Local Web sin
	 * cookies (decisiones §36), y la conversión es real desde el primer día con visitas.
	 */
	const mide = $derived(data.demo);
	const visitas = $derived(data.demo ? null : (t.visitas ?? null));
	const midoVisitas = $derived(data.demo || !!visitas?.desde);
	/**
	 * Una tienda real que todavía no tuvo ningún pedido: los números del período, el gráfico, la meta
	 * y los rankings serían todos cero, así que el Inicio muestra sólo lo que sirve para arrancar.
	 */
	const sinPedidos = $derived(!data.demo && !t.ultimosPedidos.length);
	const pasos = $derived(data.pasos);
	const puedeTienda = $derived(shell.permisos.includes('tienda'));
	const compartir = () => compartirTienda(ui, data.tienda);
	const ahora = new Date();

	const PERIODOS = [
		['hoy', 'Hoy'],
		['7d', '7 días'],
		['30d', '30 días']
	] as const;
	const ETIQUETA = { hoy: 'hoy', '7d': 'últimos 7 días', '30d': 'últimos 30 días' };
	const CONTRA = { hoy: 'vs. ayer', '7d': 'vs. semana anterior', '30d': 'vs. mes anterior' };
	const CANALES: Record<string, string> = {
		web: 'Tienda online',
		wa: 'WhatsApp',
		ml: 'Mercado Libre',
		ig: 'Instagram',
		pos: 'Local',
		tg: 'Telegram'
	};
	const MEDIOS: Record<string, [string, string]> = {
		efectivo: ['efectivo', 'Efectivo'],
		tarjeta: ['visa', 'Tarjetas'],
		mp: ['mp', 'QR Mercado Pago'],
		modo: ['modo', 'MODO'],
		transf: ['transf', 'Transferencia']
	};
	const LIQUIDA: Record<string, string> = {
		mp: 'Mercado Pago',
		payway: 'Payway',
		modo: 'MODO',
		uala: 'Ualá'
	};
	const RETIRO: Record<string, string> = {
		andreani: 'Retiro de Andreani',
		moto: 'Moto propia',
		correo: 'Correo Argentino',
		oca: 'OCA en sucursal'
	};

	const plural = (n: number, uno: string, varios: string) =>
		`${n.toLocaleString('es-AR')} ${n === 1 ? uno : varios}`;

	const c = $derived(curva(t.serie.actual, t.serie.anterior));
	const eje = $derived(
		Array.from({ length: 5 }, (_, i) => {
			const d0 = Date.parse(t.desde);
			const d1 = Date.parse(t.hasta);
			return etiquetaEje(new Date(d0 + ((d1 - d0) * i) / 4), t.periodo);
		})
	);
	const delta = $derived(variacion(t.ventas.actual, t.ventas.anterior));
	const kpis = $derived([
		{
			label: 'Pedidos',
			icon: I.phPackage,
			valor: t.kpis.pedidos.actual.toLocaleString('es-AR'),
			delta: diferencia(t.kpis.pedidos.actual, t.kpis.pedidos.anterior),
			nota: CONTRA[t.periodo]
		},
		{
			label: 'Ticket promedio',
			icon: I.phReceipt,
			valor: pesos(t.kpis.ticket.actual),
			delta: variacion(t.kpis.ticket.actual, t.kpis.ticket.anterior),
			nota: CONTRA[t.periodo]
		},
		{
			label: 'Conversión',
			icon: I.phFunnelSimple,
			valor: midoVisitas ? porcentaje(t.kpis.conversion.actual) : '—',
			// Sin visitas en el período anterior (se empezó a medir hace poco) no hay con qué comparar.
			delta:
				data.demo || visitas?.visitantes.anterior
					? puntos(t.kpis.conversion.actual, t.kpis.conversion.anterior)
					: null,
			nota: !midoVisitas
				? 'Sin visitas todavía'
				: data.demo || visitas?.visitantes.anterior
					? 'puntos'
					: 'pedidos web por cada 100 visitas'
		},
		{
			label: 'Clientes nuevos',
			icon: I.phUserPlus,
			valor: t.kpis.clientesNuevos.actual.toLocaleString('es-AR'),
			delta: diferencia(t.kpis.clientesNuevos.actual, t.kpis.clientesNuevos.anterior),
			nota: CONTRA[t.periodo]
		}
	]);
	const PAGINAS: Record<string, string> = {
		inicio: 'Inicio',
		catalogo: 'Catálogo',
		producto: 'Productos',
		carrito: 'Carrito',
		checkout: 'Checkout'
	};
	/** "30 de octubre": la fecha (YYYY-MM-DD) desde la que se miden las visitas. */
	const fechaCorta = (d: string) =>
		new Date(`${d}T12:00:00Z`).toLocaleDateString('es-AR', { day: 'numeric', month: 'long' });
	const visitasPeriodos = $derived(
		visitas
			? [
					{
						label: 'Hoy',
						actual: visitas.hoy.actual,
						delta: null,
						sin: `ayer ${visitas.hoy.anterior.toLocaleString('es-AR')}`
					},
					{
						label: '7 días',
						actual: visitas.dias7.actual,
						delta: variacion(visitas.dias7.actual, visitas.dias7.anterior),
						sin: 'sin comparar'
					},
					{
						label: '30 días',
						actual: visitas.dias30.actual,
						delta: variacion(visitas.dias30.actual, visitas.dias30.anterior),
						sin: 'sin comparar'
					}
				]
			: []
	);
	const maxVistos = $derived(Math.max(1, ...(visitas?.masVistos ?? []).map((p) => p.vistas)));
	const metaPct = $derived(
		t.meta.objetivo ? Math.round((t.meta.actual / t.meta.objetivo) * 100) : 0
	);

	const accesos = $derived(
		[
			{
				icon: I.phCameraFill,
				sec: 'Productos',
				t: 'Cargar con foto',
				k: 'Subí las fotos del celular',
				bg: 'rgba(210,88,47,.12)',
				fg: '#b0441d',
				// Sin productos, directo al editor (el primero se carga desde acá).
				href: pasos?.productos === 0 ? `${rutaDe('Productos')}?nuevo=1` : rutaDe('Productos')
			},
			{
				icon: I.phLinkBold,
				sec: 'Pedidos',
				t: 'Link de pago',
				k: 'Mandalo por donde quieras',
				bg: 'rgba(47,143,99,.13)',
				fg: '#237550',
				href: rutaDe('Pedidos'),
				abrir: () => shell.abrirLink()
			},
			{
				icon: I.phCashRegisterFill,
				sec: 'Ventas en local',
				t: 'Cobrar en el local',
				k: 'Punto de venta',
				bg: 'rgba(227,169,59,.2)',
				fg: '#8a5f0e',
				href: rutaDe('Ventas en local')
			},
			{
				icon: I.phMegaphoneFill,
				sec: mide ? 'Marketing' : 'Cupones',
				t: mide ? 'Enviar campaña' : 'Crear un cupón',
				k: mide ? 'WhatsApp con costo exacto' : 'Para compartir con tus clientes',
				bg: 'rgba(106,79,160,.13)',
				fg: '#6A4FA0',
				// Las campañas salen por WhatsApp Business, que todavía no se conecta: el acceso lleva a lo
				// que sí funciona para traer ventas.
				href: rutaDe(mide ? 'Marketing' : 'Cupones')
			},
			// Sin envíos para hoy, imprimir etiquetas no lleva a nada: en una tienda real ese lugar es para
			// compartir el link, que es lo que trae los pedidos.
			data.demo || t.envios.paquetes
				? {
						icon: I.phPrinterFill,
						sec: 'Envíos',
						t: 'Imprimir etiquetas',
						k: t.envios.paquetes
							? plural(t.envios.paquetes, 'envío de hoy', 'envíos de hoy')
							: 'Sin envíos hoy',
						bg: 'var(--chip)',
						fg: 'var(--mu)',
						href: rutaDe('Envíos')
					}
				: {
						icon: I.phStorefrontFill,
						sec: 'Resumen',
						t: 'Compartir tu tienda',
						k: 'Por WhatsApp o redes',
						bg: 'rgba(47,143,99,.13)',
						fg: '#237550',
						href: rutaDe('Mi tienda'),
						abrir: compartir
					}
		].filter((q) => visible(q.sec, shell.turnos, shell.permisos))
	);

	type Tarea = {
		icon: typeof I.phPackageFill;
		title: string;
		sub: string;
		bg: string;
		fg: string;
		href: string;
		hacer?: () => unknown;
	};
	/** Lo que falta para empezar a vender (sólo en una tienda real, y sólo lo que se sabe seguro). */
	const primeros = $derived(
		!pasos
			? []
			: (
					[
						pasos.productos === 0 && {
							icon: I.phCameraFill,
							title: 'Cargá tu primer producto',
							sub: 'Con foto, precio y stock: un minuto desde el celular',
							bg: 'rgba(210,88,47,.13)',
							fg: '#b0441d',
							href: `${rutaDe('Productos')}?nuevo=1`
						},
						// Cobros e integraciones son de quien maneja la tienda (dueña o administración).
						puedeTienda &&
							pasos.transferencia === false && {
								icon: I.phBankFill,
								title: 'Cargá tu alias o CBU',
								sub: 'Para que te paguen por transferencia',
								bg: 'rgba(227,169,59,.18)',
								fg: '#8a5f0e',
								href: `${rutaDe('Configuración')}?tab=negocio`
							},
						puedeTienda &&
							pasos.mercadoPago === false && {
								icon: I.phHandCoinsFill,
								title: 'Conectá Mercado Pago',
								sub: 'Para cobrar con tarjeta y QR',
								bg: 'rgba(0,158,227,.12)',
								fg: '#00699a',
								href: rutaDe('Integraciones')
							},
						sinPedidos &&
							!!pasos.productos && {
								icon: I.phStorefrontFill,
								title: 'Compartí tu tienda',
								sub: 'Mandá el link por WhatsApp o subilo a tus redes',
								bg: 'rgba(47,143,99,.13)',
								fg: '#237550',
								href: rutaDe('Mi tienda'),
								hacer: compartir
							}
					] as (Tarea | false)[]
				).filter((x): x is Tarea => !!x)
	);
	const tareas = $derived([
		...primeros,
		...[
			t.paraHacer.preparar && {
				icon: I.phPackageFill,
				title: plural(t.paraHacer.preparar, 'pedido para preparar', 'pedidos para preparar'),
				sub: 'Pagados y esperando su paquete',
				bg: 'rgba(210,88,47,.13)',
				fg: '#b0441d',
				href: `${rutaDe('Pedidos')}?estado=para-preparar`
			},
			t.paraHacer.chats && {
				icon: I.phChatCircleDotsFill,
				title: plural(t.paraHacer.chats, 'chat sin responder', 'chats sin responder'),
				sub: `El más viejo, hace ${haceCorto(new Date(Date.now() - t.paraHacer.chatMasViejoMin * 60_000).toISOString())}`,
				bg: 'rgba(47,143,99,.13)',
				fg: '#237550',
				href: rutaDe('WhatsApp')
			},
			t.paraHacer.transferencias && {
				icon: I.phBankFill,
				title: plural(
					t.paraHacer.transferencias,
					'transferencia para confirmar',
					'transferencias para confirmar'
				),
				sub: `${pesos(t.paraHacer.transferenciasMonto)} · marcalas cuando te llegue la plata`,
				bg: 'rgba(227,169,59,.18)',
				fg: '#8a5f0e',
				href: `${rutaDe('Pedidos')}?estado=pago-pendiente`
			},
			t.paraHacer.sinStock && {
				icon: I.phWarningCircleFill,
				title: plural(t.paraHacer.sinStock, 'producto sin stock', 'productos sin stock'),
				sub: t.paraHacer.sinStockEnMl
					? `${t.paraHacer.sinStockEnMl} ${t.paraHacer.sinStockEnMl === 1 ? 'publicado' : 'publicados'} en Mercado Libre`
					: 'Cargá stock o pausalos',
				bg: 'var(--chip)',
				fg: 'var(--mu)',
				href: rutaDe('Inventario')
			}
		].filter((x): x is Tarea => !!x)
	]);

	const integraciones = $derived(
		// Con la API, lo conectado de verdad (el tablero de Go no trae integraciones).
		data.conectadas?.length && !t.integraciones.length
			? data.conectadas.map((c) => ({ id: c.marca, n: c.nombre, st: 'Conectada', dot: '#2F8F63' }))
			: t.integraciones.map(({ id, valor }) => {
					const [n, st, aviso] =
						id === 'mp'
							? ['Mercado Pago', `Cobros al día · ${pesosCortos(valor)} esta semana`, false]
							: id === 'ml'
								? [
										'Mercado Libre',
										`${plural(valor, 'publicación sincronizada', 'publicaciones sincronizadas')}`,
										false
									]
								: id === 'wa'
									? [
											'WhatsApp',
											valor ? plural(valor, 'chat esperando', 'chats esperando') : 'Al día',
											valor > 0
										]
									: id === 'arca'
										? ['ARCA', `Certificado vence en ${plural(valor, 'día', 'días')}`, valor <= 15]
										: id === 'andreani'
											? ['Andreani', plural(valor, 'envío en camino', 'envíos en camino'), false]
											: [id, '', false];
					return { id, n, st, dot: aviso ? '#E3A93B' : '#2F8F63' };
				})
	);

	// "Ahora en tu tienda": en la demo se mueve como en el handoff.
	let tick = $state(0);
	$effect(() => {
		const id = setInterval(() => tick++, 3200);
		return () => clearInterval(id);
	});
	const enVivo = $derived(
		data.demo ? t.enVivo.personas + ((tick * 7) % 13) - 6 : t.enVivo.personas
	);
	const barras = $derived(
		Array.from({ length: 30 }, (_, i) => ({
			h: 25 + ((i * 37 + tick * 11) % 70),
			c: i > 26 ? '#D2582F' : 'var(--sf3)'
		}))
	);
	const maxVendidos = $derived(Math.max(1, ...t.masVendidos.map((p) => p.unidades)));

	let ocultando = $state(false);
	let recordando = $state(false);
</script>

<Seo
	title="Panel · {data.tienda.nombre}"
	description="Panel de Tu Local Web"
	path="/panel"
	noindex
/>

<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas del panel armadas con rutaDe -->
<div class="cab" use:anim={{ t: 'rise' }}>
	<button
		type="button"
		class="hornero"
		aria-label="Preguntale a Tincho"
		onclick={() => shell.hornero()}><span><img src="/hornero/avatar.webp" alt="" /></span></button
	>
	<div class="hola">
		<h1>{saludo(ahora)}{t.hola ? `, ${t.hola}` : ''}</h1>
		<div class="sub">
			{fechaLarga(ahora)}
			{#if t.meta.objetivo}· vas por el <b>{metaPct}%</b> de tu meta del mes{/if}
		</div>
	</div>
	{#if !sinPedidos}<div class="periodos" role="group" aria-label="Período">
			{#each PERIODOS as [id, label] (id)}
				<a
					href="?periodo={id}"
					data-sveltekit-noscroll
					data-sveltekit-replacestate
					aria-current={t.periodo === id ? 'true' : undefined}>{label}</a
				>
			{/each}
		</div>{/if}
	<button type="button" class="nuevo" onclick={() => (shell.cmdk = true)}
		><Icon icon={I.phPlusBold} size={14} />Nuevo</button
	>
</div>

{#if t.avisos.certificadoDias !== null}
	<div class="alerta" use:anim={{ t: 'rise' }}>
		<span class="a-ico"><Icon icon={I.phCertificateFill} size={20} /></span>
		<div class="a-txt">
			<div class="a-t">
				Tu certificado de ARCA vence en {plural(t.avisos.certificadoDias, 'día', 'días')}
			</div>
			<div class="a-d">
				Renovalo para seguir facturando sin cortes. Tarda 2 minutos y te guiamos paso a paso.
			</div>
		</div>
		<a class="a-btn" href={rutaDe('Facturación')}>Renovar ahora</a>
		<form
			method="POST"
			action="?/ocultarCertificado"
			use:enhance={() => {
				ocultando = true;
				return async ({ update }) => {
					await update();
					ocultando = false;
				};
			}}
		>
			<button class="a-x" title="Descartar" aria-label="Descartar el aviso" disabled={ocultando}
				><Icon icon={I.phXBold} size={14} /></button
			>
		</form>
	</div>
{/if}

<div class="accesos">
	{#each accesos as q (q.t)}
		<a
			class="acceso"
			href={q.href}
			onclick={(e) => {
				if (!('abrir' in q) || !q.abrir) return;
				e.preventDefault();
				q.abrir();
			}}
		>
			<span class="q-ico" style:background={q.bg} style:color={q.fg}
				><Icon icon={q.icon} size={20} /></span
			>
			<span class="q-txt"><span class="q-t">{q.t}</span><span class="q-k">{q.k}</span></span>
		</a>
	{/each}
</div>

{#if !sinPedidos}
	<div class="kpis">
		{#each kpis as k (k.label)}
			<div class="kpi" use:anim={{ t: 'rise' }}>
				<div class="k-cab">
					<span>{k.label}</span><span class="k-ico"><Icon icon={k.icon} size={17} /></span>
				</div>
				<div class="k-v">{k.valor}</div>
				<div class="k-n">
					{#if k.delta}<span class="k-d" class:baja={k.delta.startsWith('−')}>{k.delta}</span>{/if}
					{k.nota}
				</div>
			</div>
		{/each}
	</div>
{/if}

{#snippet paraHacer(primero: boolean)}
	<div class="tarjeta col" class:primero class:urgente={!primero && tareas.length > 0}>
		<div class="fila-cab">
			<div class="titulo">{primero ? 'Para empezar a vender' : 'Para hacer hoy'}</div>
			{#if tareas.length}<span class="cuenta">{tareas.length}</span>{/if}
		</div>
		<div class="tareas">
			{#each tareas as x (x.title)}
				<a
					class="tarea"
					href={x.href}
					onclick={(e) => {
						if (!x.hacer) return;
						e.preventDefault();
						x.hacer();
					}}
				>
					<span class="t-ico" style:background={x.bg} style:color={x.fg}
						><Icon icon={x.icon} size={19} /></span
					>
					<span class="t-txt"
						><span class="t-t">{x.title}</span><span class="t-s">{x.sub}</span></span
					>
					<Icon icon={I.phCaretRight} size={14} style="color:var(--fa)" />
				</a>
			{:else}
				<div class="vacio">
					{primero
						? 'Tu tienda ya está lista para vender. Cuando entre el primer pedido, lo ves acá.'
						: 'Nada pendiente por hoy. Buen momento para sumar productos o un cupón.'}
				</div>
			{/each}
		</div>
	</div>
{/snippet}

<div class="grilla">
	{#if sinPedidos}
		{@render paraHacer(true)}
	{:else}
		<div class="tarjeta ventas" use:anim={{ t: 'rise', delay: 60 }}>
			<div class="v-cab">
				<div>
					<div class="v-l">Ventas · {ETIQUETA[t.periodo]}</div>
					<div class="v-total">
						<span>{pesos(t.ventas.actual)}</span>
						{#if delta}
							<span class="v-delta" class:baja={delta.startsWith('−')}
								><Icon
									icon={delta.startsWith('−') ? I.phTrendDownBold : I.phTrendUpBold}
									size={13}
								/>{delta}</span
							>
						{/if}
					</div>
				</div>
				<div class="leyenda">
					<span><span class="l-act"></span>Este período</span>
					<span><span class="l-ant"></span>Anterior</span>
				</div>
			</div>
			<div class="grafico">
				<svg viewBox="0 0 600 200" preserveAspectRatio="none" aria-hidden="true">
					<defs
						><linearGradient id="ventas-area" x1="0" x2="0" y1="0" y2="1"
							><stop offset="0" stop-color="#E8875F" stop-opacity=".35"></stop><stop
								offset="1"
								stop-color="#E8875F"
								stop-opacity="0"
							></stop></linearGradient
						></defs
					>
					<line x1="0" x2="600" y1="66" y2="66" class="guia"></line>
					<line x1="0" x2="600" y1="133" y2="133" class="guia"></line>
					<path d={c.anterior} class="anterior"></path>
					<path d={c.area} fill="url(#ventas-area)"></path>
					{#key c.linea}<path use:anim={{ t: 'draw' }} d={c.linea} class="linea"></path>{/key}
				</svg>
				<div class="punto" style:top="{c.punto}%">
					<span use:anim={{ t: 'pulse' }}></span><span class="p-c"></span>
				</div>
			</div>
			<div class="eje">
				{#each eje as x, i (i)}<span>{x}</span>{/each}
			</div>
			<div class="canales">
				{#each t.canales.slice(0, 4) as ch (ch.canal)}
					<div class="canal">
						<MarcaIcono k={ch.canal} size={30} radius={9} />
						<span class="c-txt"
							><span class="c-n">{CANALES[ch.canal] ?? ch.canal}</span><span class="c-p"
								>{Math.round(ch.parte * 100)}%</span
							></span
						>
					</div>
				{/each}
			</div>
		</div>

		<MetaMes
			mes={mes(ahora)}
			objetivo={t.meta.objetivo}
			actual={t.meta.actual}
			proyeccion={t.meta.proyeccion}
			carritos={t.hornero.recordados ? 0 : t.hornero.carritos}
		/>

		<div class="tarjeta pedidos" use:anim={{ t: 'rise' }}>
			<div class="p-cab">
				<div class="titulo">Últimos pedidos</div>
				<a class="ver" href={rutaDe('Pedidos')}
					>Ver todos <Icon icon={I.phArrowRight} size={13} /></a
				>
			</div>
			{#each t.ultimosPedidos as o (o.numero)}
				<a class="pedido" href="{rutaDe('Pedidos')}?pedido={o.numero}">
					<span class="o-quien">
						{#if o.foto}<img src={cara(o.foto)} alt="" width="38" height="38" />{:else}<span
								class="o-ini"
								>{o.cliente
									.split(' ')
									.map((w) => w[0])
									.slice(0, 2)
									.join('')}</span
							>{/if}
						<span class="o-txt"
							><span class="o-c">{o.cliente}</span><span class="o-n"
								>#{o.numero} · hace {haceCorto(o.creado)}</span
							></span
						>
					</span>
					<span class="o-canal"
						><MarcaIcono k={o.canal} size={26} radius={8} />{CANALES[o.canal] ?? o.canal}</span
					>
					<span><ChipEstado estado={o.estado} /></span>
					<span class="o-total">{pesos(o.total)}</span>
				</a>
			{:else}
				<div class="vacio">Todavía no entró ningún pedido. Compartí tu tienda y aparecen acá.</div>
			{/each}
			<div class="p-pie"></div>
		</div>

		{@render paraHacer(false)}
	{/if}

	{#if !mide}
		<div class="tarjeta col">
			<div class="fila-cab">
				<div class="titulo">Visitas de tu tienda</div>
				{#if visitas?.desde}<span class="nota">sin cookies</span>{/if}
			</div>
			{#if visitas?.desde}
				<div class="visitas">
					{#each visitasPeriodos as v (v.label)}
						<div class="vi">
							<span class="vi-l">{v.label}</span>
							<span class="vi-n">{v.actual.toLocaleString('es-AR')}</span>
							<span class="vi-d" class:sube={!!v.delta} class:baja={v.delta?.startsWith('−')}
								>{v.delta ?? v.sin}</span
							>
						</div>
					{/each}
				</div>
				<div class="paginas">
					{#each visitas.paginas as p (p.pagina)}
						<div>
							<span class="pg-n">{PAGINAS[p.pagina] ?? p.pagina}</span><span class="pg-v"
								>{plural(p.vistas, 'vista', 'vistas')}</span
							>
						</div>
					{:else}
						<div><span class="pg-n">Sin visitas en este período</span></div>
					{/each}
				</div>
				<p class="vi-pie">
					Una visita es una persona distinta en el día. Medimos desde el {fechaCorta(
						visitas.desde
					)}.
				</p>
			{:else}
				<div class="sin-dato">
					<img src="/hornero/grafico.webp" alt="" width="72" height="72" />
					<p>
						Todavía no entró nadie a tu tienda. Contamos cada visita sin cookies: compartí el link y
						las vas a ver acá.
					</p>
				</div>
				<button type="button" class="b-borde abajo" onclick={compartir}>Compartir mi tienda</button>
			{/if}
		</div>
	{:else}
		<div class="tarjeta col">
			<div class="fila-cab">
				<div class="titulo">Ahora en tu tienda</div>
				<span class="vivo"
					><span class="v-punto"><span use:anim={{ t: 'pulse' }}></span><span></span></span>En vivo</span
				>
			</div>
			<div class="grande">
				<span class="g-n">{Math.max(0, enVivo)}</span><span class="g-l">personas mirando</span>
			</div>
			<div class="barras" aria-hidden="true">
				{#each barras as b, i (i)}<span style:height="{b.h}%" style:background={b.c}></span>{/each}
			</div>
			<div class="paginas">
				{#each t.enVivo.paginas as p (p.nombre)}
					<div>
						<span class="pg-n">{p.nombre}</span><span class="pg-v"
							>{p.personas + (data.demo ? tick % 3 : 0)}</span
						>
					</div>
				{/each}
			</div>
			{#if t.enVivo.carritos}
				<div class="pie">
					<div class="armando">
						<b>{plural(t.enVivo.carritos, 'carrito', 'carritos')}</b>
						{t.enVivo.carritos === 1 ? 'armándose' : 'armándose'} ahora por
						<b>{pesos(t.enVivo.carritosMonto)}</b>
					</div>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Con la API, sin cobros hoy la caja es un cero: «Cobrar en el local» ya está en los accesos. -->
	{#if mide || t.caja.total || t.caja.medios.length}
		<div class="tarjeta col">
			<div class="fila-cab">
				<div class="titulo">Caja del local</div>
				{#if t.caja.abierta}
					<span class="chip verde">Abierta {horaMin(new Date(t.caja.abierta))}</span>
				{:else if mide}
					<span class="chip gris">Cerrada</span>
				{/if}
			</div>
			<div class="grande">
				<span class="g-m">{pesos(t.caja.total)}</span><span class="g-l">cobrado hoy</span>
			</div>
			<div class="lista">
				{#each t.caja.medios as m (m.medio)}
					{@const [k, n] = MEDIOS[m.medio] ?? [m.medio, m.medio]}
					<div class="item">
						<MarcaIcono {k} size={28} radius={8} />
						<span class="i-n">{n}</span><span class="i-v">{pesos(m.monto)}</span>
					</div>
				{:else}
					<div class="vacio chico">Todavía no cobraste nada en el local hoy.</div>
				{/each}
			</div>
			<!-- Con la API no hay apertura ni cierre de caja: sólo se cobra. -->
			<div class="pie botones">
				<a class="b-osc" href={rutaDe('Ventas en local')}>Cobrar</a>
				{#if mide}
					<a class="b-borde" href={rutaDe('Ventas en local')}
						>{t.caja.abierta ? 'Cerrar caja' : 'Abrir caja'}</a
					>
				{/if}
			</div>
		</div>
	{/if}

	<!-- Las acreditaciones todavía no se traen de Mercado Pago: con la API la tarjeta no diría nada. -->
	{#if mide}
		<div class="tarjeta col">
			<div class="fila-cab">
				<div class="titulo">Próximas liquidaciones</div>
				<a class="enlace" href={rutaDe('Finanzas')}>Finanzas</a>
			</div>
			<div class="grande">
				<span class="g-m">{pesos(t.liquidaciones.reduce((s, l) => s + l.monto, 0))}</span><span
					class="g-l">en 7 días</span
				>
			</div>
			<div class="lista">
				{#each t.liquidaciones as l (l.medio)}
					{@const dia = diaCorto(l.fecha, ahora)}
					<div class="item">
						<MarcaIcono k={l.medio} size={28} radius={8} />
						<span class="i-n2"
							><span>{LIQUIDA[l.medio] ?? l.medio}</span><span class="i-d">{dia}</span></span
						>
						<span class="i-der"
							><span class="i-v">{pesos(l.monto)}</span>{#if dia === 'Hoy'}<span class="disp"
									>Disponible</span
								>{:else}<span class="prog">Programada</span>{/if}</span
						>
					</div>
				{:else}
					<div class="vacio chico">
						{mide
							? 'No hay acreditaciones en los próximos 7 días.'
							: 'Todavía no traemos las acreditaciones: las ves en tu cuenta de Mercado Pago.'}
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<div class="hornero-card" use:anim={{ t: 'rise' }}>
		<div class="h-cab">
			<div class="h-av"><img src="/hornero/avatar.webp" alt="Tincho" /></div>
			<div>
				<div class="h-n">Tincho</div>
				<div class="h-s">
					{#if t.hornero.chatsHoy || t.hornero.ideas}
						Atendió {plural(t.hornero.chatsHoy, 'chat', 'chats')} hoy · {plural(
							t.hornero.ideas,
							'idea nueva',
							'ideas nuevas'
						)}
					{:else}
						Tu asistente de la tienda
					{/if}
				</div>
			</div>
		</div>
		<div class="burbuja">
			{#if t.hornero.carritos && !t.hornero.recordados}
				Hay <b>{plural(t.hornero.carritos, 'carrito abandonado', 'carritos abandonados')}</b> por
				<b class="naranja">{pesos(t.hornero.carritosMonto)}</b>.
				{#if mide}
					¿Les mando un recordatorio por WhatsApp? Cuesta {pesos(
						t.hornero.carritos * t.hornero.costoMensaje
					)}.
				{:else}
					¿Les mando un mail con el link a su carrito? No tiene costo.
				{/if}
			{:else if t.hornero.recordados}
				Listo: les mandé el recordatorio a {t.hornero.carritos === 1
					? 'el carrito'
					: `los ${t.hornero.carritos} carritos`}
				{mide ? 'por WhatsApp' : 'por mail'}. Si alguien vuelve y compra, lo ves en Carritos.
			{:else if pasos?.productos === 0}
				Cuando cargues tus productos y entren las primeras visitas, te aviso acá cómo vender más.
			{:else}
				Todo en orden por ahora. Cuando vea algo para vender más, te aviso acá.
			{/if}
		</div>
		<div class="h-acc">
			{#if t.hornero.carritos && !t.hornero.recordados}
				<form
					method="POST"
					action="?/recordarCarritos"
					use:enhance={() => {
						recordando = true;
						return async ({ result, update }) => {
							recordando = false;
							await update();
							if (result.type === 'success') {
								const n = Number(result.data?.enviados ?? 0);
								ui.toast({
									k: 'ai',
									title: 'Recordatorios enviados',
									desc: mide
										? `${plural(n, 'mensaje', 'mensajes')} en cola por WhatsApp · ${pesos(n * t.hornero.costoMensaje)}`
										: `${plural(n, 'mail', 'mails')} en camino`
								});
							}
						};
					}}
				>
					<button class="h-si" disabled={recordando}>Sí, enviar</button>
				</form>
			{:else if mide || t.hornero.carritos}
				<a class="h-si" href={rutaDe('Carritos abandonados')}>Ver carritos</a>
			{/if}
			<button type="button" class="h-ideas" onclick={() => shell.hornero()}
				>{t.hornero.ideas
					? `Ver ${t.hornero.ideas === 1 ? 'la idea' : `las ${t.hornero.ideas} ideas`}`
					: 'Hablar con Tincho'}</button
			>
		</div>
	</div>

	{#if mide || t.bandeja.length}
		<div class="tarjeta col">
			<div class="fila-cab">
				<div class="titulo con-ico"><Icon icon={I.logosWhatsappIcon} size={18} />Bandeja</div>
				{#if t.bandeja.length}<span class="chip ambar">{t.bandeja.length} sin responder</span>{/if}
			</div>
			<div class="lista">
				{#each t.bandeja as v (v.nombre)}
					<div class="conv">
						{#if v.foto}<img src={cara(v.foto)} alt="" width="38" height="38" />{:else}<span
								class="o-ini"
								>{v.nombre
									.split(' ')
									.map((w) => w[0])
									.slice(0, 2)
									.join('')}</span
							>{/if}
						<span class="cv-txt">
							<span class="cv-cab"
								><span class="cv-n">{v.nombre}</span><span class="cv-h">{haceCorto(v.creado)}</span
								></span
							>
							<span class="cv-m">{v.mensaje}</span>
							{#if v.atiende === 'hornero'}
								<span class="cv-bot"
									><Icon icon={I.phSparkleFill} size={11} />Tincho respondiendo<span
										class="dots"
										use:anim={{ t: 'dots' }}
									></span></span
								>
							{:else}
								<span class="cv-vos"><Icon icon={I.phUserFill} size={11} />Te espera a vos</span>
							{/if}
						</span>
						{#if v.sinLeer}<span class="cv-sin">{v.sinLeer}</span>{/if}
					</div>
				{:else}
					<div class="vacio chico">No hay conversaciones esperando.</div>
				{/each}
			</div>
			<a class="b-borde abajo" href={rutaDe('WhatsApp')}>Abrir bandeja</a>
		</div>
	{/if}

	{#if mide || integraciones.length}
		<div class="tarjeta">
			<div class="fila-cab">
				<div class="titulo">Integraciones</div>
				<a class="enlace" href={rutaDe('Integraciones')}>Explorar</a>
			</div>
			<div class="lista integ">
				{#each integraciones as l (l.id)}
					<div class="item">
						<MarcaIcono k={l.id} size={32} radius={10} />
						<span class="i-n2"><span class="in-n">{l.n}</span><span class="in-s">{l.st}</span></span
						>
						<span class="in-dot" style:background={l.dot}></span>
					</div>
				{:else}
					<div class="vacio chico">Conectá tus canales y medios de pago desde Integraciones.</div>
				{/each}
			</div>
		</div>
	{/if}

	{#if mide || t.masVendidos.length}
		<div class="tarjeta">
			<div class="fila-cab">
				<div class="titulo">Más vendidos</div>
				<span class="nota">esta semana</span>
			</div>
			<div class="tops">
				{#each t.masVendidos as p (p.productoId)}
					<div class="top">
						{#if p.foto}<img
								src={unsplash(p.foto, 96, '&h=96')}
								alt=""
								width="40"
								height="40"
							/>{/if}
						<div class="tp-txt">
							<div class="tp-cab">
								<span class="tp-n">{p.nombre}</span><span class="tp-u">{p.unidades}&nbsp;u</span>
							</div>
							<div class="tp-barra">
								<div style:width="{(p.unidades / maxVendidos) * 100}%"></div>
							</div>
						</div>
					</div>
				{:else}
					<div class="vacio chico">Cuando vendas, acá ves lo que más sale.</div>
				{/each}
			</div>
		</div>
	{/if}

	{#if visitas?.desde}
		<div class="tarjeta">
			<div class="fila-cab">
				<div class="titulo">Más vistos</div>
				<span class="nota">{ETIQUETA[t.periodo]}</span>
			</div>
			<div class="tops">
				{#each visitas.masVistos as p (p.productoId)}
					<div class="top">
						{#if p.foto}<img
								src={unsplash(p.foto, 96, '&h=96')}
								alt=""
								width="40"
								height="40"
							/>{/if}
						<div class="tp-txt">
							<div class="tp-cab">
								<span class="tp-n">{p.nombre}</span><span class="tp-u"
									>{plural(p.vistas, 'vista', 'vistas')}</span
								>
							</div>
							<div class="tp-barra">
								<div style:width="{(p.vistas / maxVistos) * 100}%"></div>
							</div>
						</div>
					</div>
				{:else}
					<div class="vacio chico">Cuando miren tus productos, acá ves cuáles interesan más.</div>
				{/each}
			</div>
		</div>
	{/if}

	{#if mide || t.resenas.nuevas.length}
		<div class="tarjeta col">
			<div class="fila-cab">
				<div class="titulo">Reseñas nuevas</div>
				{#if t.resenas.promedio}<span class="estrellas-prom"
						>★ {t.resenas.promedio.toLocaleString('es-AR', { minimumFractionDigits: 1 })}</span
					>{/if}
			</div>
			<div class="lista">
				{#each t.resenas.nuevas as r, i (i)}
					<div class="resena">
						<div class="r-cab">
							<span class="r-est" aria-label="{r.estrellas} estrellas"
								>{'★'.repeat(r.estrellas)}<span class="r-off">{'★'.repeat(5 - r.estrellas)}</span
								></span
							>
							<span class="r-quien">{r.autor} · {r.producto}</span>
							<MarcaIcono k={r.canal} size={22} radius={7} />
						</div>
						<div class="r-q">{r.texto}</div>
					</div>
				{:else}
					<div class="vacio chico">Todavía no hay reseñas.</div>
				{/each}
			</div>
			{#if t.resenas.nuevas.length}
				<button
					type="button"
					class="b-ia abajo"
					onclick={() => shell.hornero('Escribí respuestas cortas para las reseñas nuevas.')}
					><Icon icon={I.phSparkleFill} size={14} />Responder con Tincho</button
				>
			{/if}
		</div>
	{/if}

	{#if mide || t.envios.paquetes || t.envios.retiros.length}
		<div class="tarjeta col ultima">
			<div class="fila-cab">
				<div class="titulo">Envíos de hoy</div>
				<span class="nota">{plural(t.envios.paquetes, 'paquete', 'paquetes')}</span>
			</div>
			<div class="retiros">
				{#each t.envios.retiros as e (e.transporte)}
					<div class="retiro">
						<span class="e-h">{e.hora}</span>
						<MarcaIcono k={e.transporte} size={32} radius={10} />
						<span class="e-txt">
							<span class="e-t"
								>{RETIRO[e.transporte]}{#if e.proximo}<span class="proximo">Próximo</span
									>{/if}</span
							>
							<span class="e-d">
								{#if e.transporte === 'andreani'}{plural(e.paquetes, 'paquete', 'paquetes')} · etiquetas
									listas
								{:else if e.transporte === 'moto'}{plural(
										e.paquetes,
										'entrega',
										'entregas'
									)}{e.ciudad ? ` en ${e.ciudad}` : ''}
								{:else}{plural(e.paquetes, 'paquete', 'paquetes')}{e.ciudad
										? ` para ${e.ciudad}`
										: ''}{/if}
							</span>
						</span>
					</div>
				{:else}
					<div class="vacio chico">Nada para despachar hoy.</div>
				{/each}
			</div>
			{#if t.envios.paquetes}
				<a class="b-osc abajo con-ico" href={rutaDe('Envíos')}
					><Icon icon={I.phPrinterBold} size={14} />Imprimir {plural(
						t.envios.paquetes,
						'etiqueta',
						'etiquetas'
					)}</a
				>
			{/if}
		</div>
	{/if}
</div>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	/* Encabezado */
	.cab {
		display: flex;
		align-items: center;
		gap: 16px;
		flex-wrap: wrap;
	}
	.hornero {
		width: 60px;
		height: 60px;
		flex: none;
		border-radius: 50%;
		padding: 3px;
		border: 0;
		background: linear-gradient(135deg, #e8875f, #c24e25);
		box-shadow: 0 10px 24px -10px rgba(210, 88, 47, 0.7);
	}
	.hornero span {
		display: block;
		width: 54px;
		height: 54px;
		border-radius: 50%;
		overflow: hidden;
		background: #fbf7f2;
	}
	.hornero img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.hola {
		flex: 1;
		min-width: 240px;
	}
	h1 {
		margin: 0;
		font-size: 32px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.sub {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.sub b {
		color: #b0441d;
	}
	.periodos {
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
	}
	.periodos a {
		font-size: 13px;
		font-weight: 600;
		padding: 8px 15px;
		border-radius: 999px;
		color: var(--mu);
	}
	.periodos a[aria-current='true'] {
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--sh);
	}
	.nuevo {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 20px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			0 10px 24px -10px rgba(210, 88, 47, 0.8);
	}

	/* Aviso de ARCA */
	.alerta {
		margin-top: 20px;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 12px 12px 12px 14px;
		border-radius: 18px;
		background: linear-gradient(90deg, rgba(227, 169, 59, 0.18), rgba(227, 169, 59, 0.06));
		border: 1px solid rgba(227, 169, 59, 0.4);
	}
	.a-ico {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 12px;
		background: #e3a93b;
		color: #fff;
		display: grid;
		place-items: center;
	}
	.a-txt {
		flex: 1;
		min-width: 0;
	}
	.a-t {
		font-size: 14.5px;
		font-weight: 600;
	}
	.a-d {
		font-size: 13px;
		color: var(--mu);
		margin-top: 1px;
	}
	.a-btn {
		font-size: 13px;
		font-weight: 600;
		padding: 9px 16px;
		border-radius: 999px;
		background: var(--tx);
		color: var(--sf);
		white-space: nowrap;
	}
	.a-x {
		width: 32px;
		height: 32px;
		border: 0;
		border-radius: 50%;
		background: transparent;
		color: var(--mu);
		display: grid;
		place-items: center;
	}

	/* Accesos rápidos */
	.accesos {
		margin-top: 18px;
		display: flex;
		flex-wrap: wrap;
		gap: 12px;
	}
	.acceso {
		/* Los cinco entran en una fila a 1280 px y más (sin uno solo abajo). */
		flex: 1 1 170px;
		min-width: 0;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 13px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--shd);
		color: var(--tx);
		transition: transform 0.2s;
	}
	.acceso:hover {
		transform: translateY(-2px);
	}
	.q-ico {
		width: 40px;
		height: 40px;
		flex: none;
		border-radius: 12px;
		display: grid;
		place-items: center;
	}
	.q-txt {
		min-width: 0;
	}
	.q-t,
	.q-k {
		display: block;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.q-t {
		font-size: 14px;
		font-weight: 600;
	}
	.q-k {
		font-size: 12px;
		color: var(--mu);
	}

	/* KPIs */
	.kpis {
		margin-top: 18px;
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 18px;
	}
	.kpi {
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		padding: 18px 20px;
	}
	.k-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		font-size: 14px;
		color: var(--mu);
	}
	.k-ico {
		width: 32px;
		height: 32px;
		border-radius: 10px;
		background: var(--sf3);
		display: grid;
		place-items: center;
		color: #b0441d;
	}
	.k-v {
		margin-top: 10px;
		font-size: 28px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.k-n {
		margin-top: 4px;
		font-size: 13px;
		color: var(--mu);
	}
	.k-d {
		font-weight: 600;
		color: #237550;
	}
	.k-d.baja {
		color: #8a5f0e;
	}

	/* Grilla de tarjetas */
	.grilla {
		margin-top: 18px;
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		grid-auto-flow: row dense;
		gap: 18px;
	}
	.tarjeta {
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		padding: 20px;
	}
	.col {
		display: flex;
		flex-direction: column;
	}
	.titulo {
		font-size: 17px;
		font-weight: 600;
	}
	.con-ico {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.fila-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
	}
	.enlace {
		font-size: 13.5px;
		font-weight: 600;
		color: var(--tx);
	}
	.nota {
		font-size: 13px;
		color: var(--mu);
	}
	.cuenta {
		font-size: 12.5px;
		font-weight: 600;
		color: #b0441d;
		background: rgba(210, 88, 47, 0.12);
		padding: 3px 9px;
		border-radius: 999px;
	}
	.chip {
		white-space: nowrap;
		font-size: 12.5px;
		font-weight: 600;
		padding: 3px 9px;
		border-radius: 999px;
	}
	.chip.verde {
		color: #237550;
		background: rgba(47, 143, 99, 0.12);
	}
	.chip.gris {
		color: var(--mu);
		background: var(--chip);
	}
	.chip.ambar {
		color: #8a5f0e;
		background: rgba(227, 169, 59, 0.18);
	}
	.vacio {
		padding: 18px 4px;
		font-size: 13.5px;
		color: var(--mu);
		line-height: 1.45;
	}
	.vacio.chico {
		padding: 12px 0;
	}
	.pie {
		margin-top: auto;
		padding-top: 14px;
	}
	.abajo {
		margin-top: auto;
	}
	.lista + .abajo,
	.retiros + .abajo {
		margin-top: 14px;
	}
	.col > .lista:has(+ .abajo),
	.col > .retiros:has(+ .abajo) {
		margin-bottom: auto;
	}
	.b-osc,
	.b-borde,
	.b-ia {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		padding: 10px;
		border-radius: 999px;
		font-size: 13.5px;
		font-weight: 600;
		text-align: center;
	}
	.b-osc {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.b-borde {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.b-ia {
		border: 0;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}

	/* Ventas */
	.ventas {
		grid-column: span 2;
		padding: 24px 26px;
		display: flex;
		flex-direction: column;
		min-height: 420px;
	}
	.v-cab {
		display: flex;
		justify-content: space-between;
		align-items: start;
		gap: 16px;
	}
	.v-l {
		font-size: 14px;
		font-weight: 500;
		color: var(--mu);
	}
	.v-total {
		display: flex;
		align-items: baseline;
		gap: 12px;
		margin-top: 6px;
		flex-wrap: wrap;
	}
	.v-total > span:first-child {
		font-size: 46px;
		font-weight: 600;
		letter-spacing: -0.045em;
	}
	.v-delta {
		display: flex;
		align-items: center;
		gap: 3px;
		font-size: 13.5px;
		font-weight: 600;
		color: #237550;
		background: rgba(47, 143, 99, 0.12);
		padding: 4px 10px;
		border-radius: 999px;
	}
	.v-delta.baja {
		color: #8a5f0e;
		background: rgba(227, 169, 59, 0.18);
	}
	.leyenda {
		display: flex;
		gap: 16px;
		font-size: 12.5px;
		color: var(--mu);
		white-space: nowrap;
	}
	.leyenda > span {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.l-act {
		width: 12px;
		height: 3px;
		border-radius: 2px;
		background: #d2582f;
	}
	.l-ant {
		width: 12px;
		height: 0;
		border-top: 2px dashed var(--fa);
	}
	.grafico {
		position: relative;
		flex: 1;
		margin-top: 16px;
		min-height: 170px;
	}
	.grafico svg {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		overflow: visible;
	}
	.guia {
		stroke: var(--ln);
		stroke-dasharray: 3 5;
		vector-effect: non-scaling-stroke;
	}
	.anterior {
		fill: none;
		stroke: var(--fa);
		stroke-width: 1.5;
		stroke-dasharray: 4 5;
		vector-effect: non-scaling-stroke;
	}
	.linea {
		fill: none;
		stroke: #d2582f;
		stroke-width: 2.5;
		stroke-linecap: round;
		vector-effect: non-scaling-stroke;
	}
	.punto {
		position: absolute;
		right: 0;
		transform: translate(50%, -50%);
		width: 12px;
		height: 12px;
	}
	.punto span {
		position: absolute;
		inset: 0;
		border-radius: 50%;
		background: #d2582f;
	}
	.punto .p-c {
		box-shadow: 0 0 0 3px var(--sf);
	}
	.eje {
		display: flex;
		justify-content: space-between;
		font-size: 12px;
		color: var(--fa);
		margin-top: 10px;
	}
	.canales {
		margin-top: 20px;
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 10px;
	}
	.canal {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 14px;
		background: var(--sf3);
		min-width: 0;
	}
	.c-txt {
		min-width: 0;
	}
	.c-n {
		display: block;
		font-size: 12px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.c-p {
		display: block;
		font-size: 15px;
		font-weight: 600;
	}

	/* Últimos pedidos */
	.pedidos {
		grid-column: span 2;
		padding: 0;
		overflow: hidden;
	}
	.primero {
		grid-column: span 2;
	}
	.p-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 20px 24px 12px;
	}
	.ver {
		font-size: 13.5px;
		font-weight: 600;
		color: #b0441d;
		display: flex;
		align-items: center;
		gap: 4px;
	}
	.pedido {
		display: grid;
		grid-template-columns: minmax(0, 1.6fr) minmax(0, 1.1fr) 150px 100px;
		align-items: center;
		gap: 12px;
		margin: 0 12px;
		padding: 11px 12px;
		border-radius: 14px;
		font-size: 14px;
		color: var(--tx);
	}
	.pedido:hover {
		background: var(--chip);
	}
	.o-quien {
		display: flex;
		align-items: center;
		gap: 12px;
		min-width: 0;
	}
	.o-quien img,
	.conv img {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 50%;
		object-fit: cover;
	}
	.o-ini {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 50%;
		background: var(--sf3);
		display: grid;
		place-items: center;
		font-size: 12px;
		font-weight: 600;
		color: var(--mu);
	}
	.o-txt {
		min-width: 0;
	}
	.o-c,
	.o-n {
		display: block;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.o-c {
		font-weight: 600;
	}
	.o-n {
		font-size: 12.5px;
		color: var(--fa);
	}
	.o-canal {
		display: flex;
		align-items: center;
		gap: 8px;
		color: var(--mu);
		font-size: 13.5px;
		white-space: nowrap;
		min-width: 0;
		overflow: hidden;
	}
	.o-total {
		text-align: right;
		font-size: 14.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.pedidos .vacio {
		padding: 8px 24px 20px;
	}
	.p-pie {
		height: 12px;
	}

	/* Para hacer hoy */
	.tareas {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.tarea {
		display: flex;
		gap: 12px;
		align-items: center;
		padding: 12px;
		border-radius: 16px;
		border: 1px solid var(--ln);
		background: var(--sf2);
		color: var(--tx);
		transition: transform 0.2s;
	}
	.tarea:hover {
		transform: translateX(2px);
	}
	.t-ico {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 12px;
		display: grid;
		place-items: center;
	}
	.t-txt {
		flex: 1;
		min-width: 0;
	}
	.t-t {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.t-s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		margin-top: 2px;
	}

	.sin-dato {
		flex: 1;
		display: grid;
		justify-items: center;
		align-content: center;
		gap: 8px;
		padding: 8px 4px 12px;
		text-align: center;
	}
	.sin-dato p {
		margin: 0;
		max-width: 260px;
		font-size: 13px;
		line-height: 1.45;
		color: var(--fa);
		text-wrap: pretty;
	}
	/* Visitas (medición propia) */
	.visitas {
		margin-top: 12px;
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 10px;
	}
	.vi {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}
	.vi-l {
		font-size: 12.5px;
		color: var(--mu);
	}
	.vi-n {
		font-size: 28px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.vi-d {
		font-size: 12.5px;
		color: var(--fa);
		white-space: nowrap;
	}
	.vi-d.sube {
		font-weight: 600;
		color: #237550;
	}
	.vi-d.baja {
		color: #8a5f0e;
	}
	.vi-pie {
		margin: auto 0 0;
		padding-top: 14px;
		font-size: 12.5px;
		line-height: 1.45;
		color: var(--fa);
		text-wrap: pretty;
	}
	/* En vivo */
	.vivo {
		white-space: nowrap;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12.5px;
		font-weight: 600;
		color: #237550;
	}
	.v-punto {
		position: relative;
		width: 7px;
		height: 7px;
	}
	.v-punto span {
		position: absolute;
		inset: 0;
		border-radius: 50%;
		background: #2f8f63;
	}
	.grande {
		margin-top: 10px;
		display: flex;
		align-items: baseline;
		gap: 8px;
		flex-wrap: wrap;
	}
	.g-n {
		font-size: 44px;
		font-weight: 600;
		letter-spacing: -0.05em;
	}
	.g-m {
		font-size: 34px;
		font-weight: 600;
		letter-spacing: -0.045em;
	}
	.g-l {
		font-size: 14px;
		color: var(--mu);
	}
	.barras {
		margin-top: 8px;
		display: flex;
		gap: 2px;
		height: 34px;
		align-items: end;
	}
	.barras span {
		flex: 1;
		border-radius: 2px;
		transition: height 0.6s;
	}
	.paginas {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.paginas div {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		font-size: 13.5px;
	}
	.pg-n {
		color: var(--mu);
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.pg-v {
		font-weight: 600;
	}
	.armando {
		padding: 10px 12px;
		border-radius: 12px;
		background: rgba(210, 88, 47, 0.1);
		font-size: 13px;
		line-height: 1.4;
	}

	/* Listas con ícono (caja, liquidaciones, integraciones) */
	.lista {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.integ {
		margin-top: 8px;
	}
	.item {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 0;
		border-bottom: 1px solid var(--ln);
	}
	.integ .item {
		gap: 12px;
		padding: 9px 0;
	}
	.item :global(.marca) {
		box-shadow: var(--sh);
	}
	.i-n {
		flex: 1;
		font-size: 13.5px;
	}
	.i-n2 {
		flex: 1;
		min-width: 0;
	}
	.i-n2 > span {
		display: block;
	}
	.i-n2 > span:first-child {
		font-size: 13.5px;
	}
	.i-d {
		font-size: 12px;
		color: var(--mu);
	}
	.i-v {
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
	}
	.i-der {
		text-align: right;
	}
	.i-der > span {
		display: block;
	}
	.disp {
		white-space: nowrap;
		font-size: 11.5px;
		font-weight: 600;
		color: #237550;
	}
	.prog {
		font-size: 11.5px;
		color: var(--fa);
	}
	.in-n {
		font-size: 14px !important;
		font-weight: 600;
	}
	.in-s {
		font-size: 12.5px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.in-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex: none;
	}
	.botones {
		display: flex;
		gap: 8px;
	}
	.botones a {
		flex: 1;
	}

	/* Tincho */
	.hornero-card {
		position: relative;
		overflow: hidden;
		border-radius: 24px;
		background: linear-gradient(150deg, #d2582f, #b9481f);
		color: #fff;
		padding: 22px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		box-shadow: 0 20px 40px -20px rgba(210, 88, 47, 0.6);
	}
	.h-cab {
		position: relative;
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.h-av {
		width: 52px;
		height: 52px;
		flex: none;
		border-radius: 50%;
		overflow: hidden;
		background: #fbf7f2;
		box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.3);
	}
	.h-av img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.h-n {
		font-size: 16px;
		font-weight: 600;
	}
	.h-s {
		font-size: 13px;
		color: #fce3d6;
	}
	.burbuja {
		position: relative;
		padding: 14px 16px;
		border-radius: 4px 18px 18px 18px;
		background: #fff;
		color: #2a140c;
		font-size: 14.5px;
		line-height: 1.5;
	}
	.naranja {
		color: #b0441d;
	}
	.h-acc {
		position: relative;
		display: flex;
		gap: 8px;
		margin-top: auto;
	}
	.h-acc form {
		flex: 1;
		display: flex;
	}
	.h-si {
		flex: 1;
		padding: 11px;
		border-radius: 999px;
		border: 0;
		background: #2a140c;
		color: #fff;
		font-size: 13.5px;
		font-weight: 600;
		text-align: center;
	}
	.h-ideas {
		white-space: nowrap;
		padding: 11px 16px;
		border-radius: 999px;
		border: 1px solid rgba(255, 255, 255, 0.45);
		background: transparent;
		color: #fff;
		font-size: 13.5px;
		font-weight: 600;
	}

	/* Bandeja */
	.conv {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.cv-txt {
		flex: 1;
		min-width: 0;
	}
	.cv-cab {
		display: flex;
		justify-content: space-between;
		gap: 8px;
	}
	.cv-n {
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.cv-h {
		font-size: 12px;
		color: var(--fa);
		flex: none;
	}
	.cv-m {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.cv-bot,
	.cv-vos {
		white-space: nowrap;
		margin-top: 4px;
		display: inline-flex;
		align-items: center;
		gap: 5px;
		font-size: 11.5px;
		font-weight: 600;
	}
	.cv-bot {
		color: #b0441d;
	}
	.cv-vos {
		color: #8a5f0e;
	}
	.dots {
		width: 4px;
		height: 4px;
		border-radius: 50%;
		background: #c24e25;
	}
	.cv-sin {
		white-space: nowrap;
		min-width: 20px;
		height: 20px;
		padding: 0 6px;
		border-radius: 999px;
		background: #2f8f63;
		color: #fff;
		font-size: 11.5px;
		font-weight: 700;
		display: grid;
		place-items: center;
	}

	/* Más vendidos */
	.tops {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.top {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.top img {
		width: 40px;
		height: 40px;
		flex: none;
		border-radius: 12px;
		object-fit: cover;
	}
	.tp-txt {
		flex: 1;
		min-width: 0;
	}
	.tp-cab {
		display: flex;
		justify-content: space-between;
		gap: 10px;
		font-size: 14px;
	}
	.tp-n {
		font-weight: 500;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.tp-u {
		font-weight: 600;
		white-space: nowrap;
	}
	.tp-barra {
		margin-top: 7px;
		height: 6px;
		border-radius: 4px;
		background: var(--sf3);
		overflow: hidden;
	}
	.tp-barra div {
		height: 100%;
		border-radius: 4px;
		background: linear-gradient(90deg, #d2582f, #e8875f);
	}

	/* Reseñas */
	.estrellas-prom {
		font-size: 13px;
		font-weight: 600;
		color: #c98a12;
	}
	.resena {
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.r-cab {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.r-est {
		color: #e3a93b;
		font-size: 12px;
		letter-spacing: 1px;
		white-space: nowrap;
	}
	.r-off {
		color: var(--ln2);
	}
	.r-quien {
		flex: 1;
		min-width: 0;
		font-size: 12.5px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.r-q {
		margin-top: 4px;
		font-size: 13.5px;
		line-height: 1.45;
	}

	/* Envíos */
	.retiros {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
	}
	.retiro {
		display: flex;
		gap: 12px;
		padding-bottom: 14px;
	}
	.e-h {
		width: 44px;
		flex: none;
		font-size: 13px;
		font-weight: 600;
		padding-top: 6px;
	}
	.e-txt {
		flex: 1;
		min-width: 0;
	}
	.e-t {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 14px;
		font-weight: 600;
	}
	.proximo {
		white-space: nowrap;
		font-size: 11px;
		font-weight: 700;
		color: #fff;
		background: #d2582f;
		padding: 2px 7px;
		border-radius: 999px;
	}
	.e-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}

	/* 760–1099: dos columnas */
	@media (max-width: 1099px) {
		.kpis {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
		.grilla {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
		.canales {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
		.ultima {
			grid-column: 1 / -1;
		}
	}
	/* Celular */
	@media (max-width: 759px) {
		/* El aviso: texto a lo ancho y el botón abajo (en el handoff el texto queda en una columna angosta). */
		.alerta {
			display: grid;
			grid-template-columns: 38px minmax(0, 1fr) 32px;
			align-items: start;
		}
		.a-btn {
			grid-column: 2;
			justify-self: start;
		}
		.alerta form {
			grid-column: 3;
			grid-row: 1;
		}
		/* La leyenda del gráfico baja debajo del total. */
		.v-cab {
			flex-wrap: wrap;
		}
		.acceso {
			flex-basis: 150px;
		}
		/* A lo angosto el título no se corta con «…»: baja de renglón, parejo. */
		.q-t {
			white-space: normal;
			text-wrap: balance;
			line-height: 1.25;
		}
		.q-k {
			display: none;
		}
		.kpis,
		.grilla {
			gap: 12px;
		}
		.grilla {
			grid-template-columns: minmax(0, 1fr);
		}
		.ventas,
		.pedidos,
		.primero,
		.ultima {
			grid-column: span 1;
		}
		/* En el celular, lo que hay que hacer va antes que los números. */
		.urgente {
			order: -1;
		}
		.pedido {
			grid-template-columns: minmax(0, 1fr) auto auto;
		}
		.o-canal {
			display: none;
		}
	}

	/* Área de toque de 44 px en pantallas táctiles, sin cambiar el dibujo del handoff. */
	@media (pointer: coarse) {
		.ver {
			position: relative;
		}
		.ver::after {
			content: '';
			position: absolute;
			inset: -14px -8px;
		}
	}
</style>
