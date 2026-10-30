<script lang="ts">
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { onMount } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import {
		alcanceTexto,
		cuando,
		encabezado,
		costoPorUso,
		grande,
		vigenciaTexto,
		type CuponNuevo
	} from '#lib/panel/cupones.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { SITE } from '#lib/site.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Cupones y promociones (design/PanelCupones.dc.html): cupones, automáticas y relámpago. */
	let { data } = $props();
	const ui = getUi();
	type C = (typeof data.cupones)[number];

	let ahora = $state(Date.now());
	onMount(() => {
		// Desde Marketing, "+ Nuevo cupón" llega con ?nuevo y abre el armador.
		if (page.url.searchParams.has('nuevo')) abrir();
		const t = setInterval(() => (ahora = Date.now()), 1000);
		return () => clearInterval(t);
	});

	const veces = (n: number) => '× ' + n.toFixed(1).replace('.', ',');
	const k = $derived(data.kpis);
	const kpis = $derived([
		{
			i: I.phTicketFill,
			c: '#D2582F',
			t: 'Cupones activos',
			v: String(k.activos),
			d: `${k.usosMes} usos este mes`,
			dc: 'var(--mu)'
		},
		{
			i: I.phTrendUpBold,
			c: '#2F8F63',
			t: 'Ventas con cupón',
			v: pesos(k.vendido),
			d:
				k.vsMesAnterior === null
					? 'este mes'
					: `${k.vsMesAnterior > 0 ? '+' : ''}${k.vsMesAnterior} % vs. el mes pasado`,
			dc: '#2F8F63'
		},
		{
			i: I.phHandCoinsBold,
			c: '#9A6A10',
			t: 'Descontaste',
			v: pesos(k.descontado),
			d: `${k.vendido ? Math.round((k.descontado / k.vendido) * 100) : 0} % de lo vendido`,
			dc: 'var(--mu)'
		},
		{
			i: I.phScalesBold,
			c: '#3E6BFD',
			t: 'Rendimiento',
			v: k.descontado ? veces(k.vendido / k.descontado) : '—',
			d: 'Ventas por cada $ 1 de descuento',
			dc: 'var(--mu)'
		}
	]);

	async function pedir(url: string, method: string, cuerpo?: unknown) {
		const r = await fetch(url, {
			method,
			headers: cuerpo === undefined ? {} : { 'content-type': 'application/json' },
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			ui.toast({ k: 'err', title: 'No se pudo', desc: m ?? 'Probá de nuevo.' });
			return null;
		}
		return r;
	}
	const accion = (id: string, a: string, minimo?: number) =>
		pedir(`/panel/cupones/${id}`, 'POST', { accion: a, minimo });

	// ---- Consejo de Tincho ----
	async function aplicarConsejo() {
		const c = data.consejo!;
		if (!(await accion(c.cuponId, 'minimo', c.minimo))) return;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: `${c.codigo} actualizado`,
			desc: `Ahora pide un mínimo de ${pesos(c.minimo)}`,
			action: 'Deshacer',
			onAction: async () => {
				await accion(c.cuponId, 'deshacer-minimo');
				await invalidateAll();
			}
		});
	}
	async function ahoraNo() {
		await pedir('/panel/cupones/consejo', 'DELETE');
		await invalidateAll();
	}

	// ---- Pestañas ----
	let tab = $state<'cup' | 'auto' | 'flash'>('cup');
	const TABS = $derived([
		['cup', 'Cupones', I.phTicketBold, data.cupones.length],
		['auto', 'Automáticas', I.phMagicWandBold, data.automaticas.filter((a) => a.activa).length],
		[
			'flash',
			'Relámpago',
			I.phLightningBold,
			data.relampago.vivo && !data.relampago.vivo.terminada ? 1 : 0
		]
	] as const);

	// ---- Lista de cupones ----
	const ESTADOS = {
		activo: ['Activo', 'rgba(47,143,99,.13)', '#2F8F63', '#2A140C'],
		programado: ['Programado', 'rgba(62,107,253,.12)', '#3E6BFD', '#3E3A63'],
		pausado: ['Pausado', 'rgba(227,169,59,.18)', '#9A6A10', '#6B5244'],
		terminado: ['Terminado', 'var(--chip)', 'var(--mu)', '#7A6252']
	} as const;
	const FILTROS = [
		['todos', 'Todos', 'var(--mu)'],
		['activo', 'Activos', '#2F8F63'],
		['programado', 'Programados', '#3E6BFD'],
		['pausado', 'Pausados', '#E3A93B'],
		['terminado', 'Terminados', '#9C8676']
	] as const;
	let filtro = $state<(typeof FILTROS)[number][0]>('todos');
	let q = $state('');
	const ORDENES = [
		['ventas', 'Más ventas'],
		['rinde', 'Mejor rendimiento'],
		['usos', 'Más usados'],
		['nuevos', 'Más nuevos']
	] as const;
	let orden = $state(0);
	const rinde = (c: C) => (c.desconto ? c.vendio / c.desconto : 0);
	const lista = $derived(
		data.cupones
			.filter(
				(c) =>
					(filtro === 'todos' || c.estado === filtro) &&
					(!q.trim() ||
						`${c.codigo} ${c.descripcion}`.toLowerCase().includes(q.trim().toLowerCase()))
			)
			.sort((a, b) => {
				const o = ORDENES[orden][0];
				if (o === 'ventas') return b.vendio - a.vendio;
				if (o === 'rinde') return rinde(b) - rinde(a);
				if (o === 'usos') return b.usos - a.usos;
				return Date.parse(b.creado) - Date.parse(a.creado);
			})
	);
	const CANAL = {
		web: [I.phGlobeSimpleFill, '#D2582F', 'Tienda'],
		wa: [I.logosWhatsappIcon, null, 'WhatsApp'],
		ig: [I.logosInstagramIcon, null, 'Instagram'],
		pos: [I.phStorefrontFill, '#8A5A3E', 'Local']
	} as const;

	/** El link que deja el cupón listo para usar en la tienda (/t/[tienda]/c/[codigo]). */
	const linkDe = (codigo: string) => `${SITE}/t/${data.slug}/c/${codigo.toLowerCase()}`;

	async function copiar(c: C) {
		try {
			await navigator.clipboard.writeText(c.codigo);
		} catch {
			// Sin permiso: igual queda el aviso con el código.
		}
		ui.toast({ k: 'ok', title: `${c.codigo} copiado`, desc: 'Pegalo donde quieras' });
	}
	function mandar(c: C) {
		const texto = `¡Hola! Te paso un código de ${data.tienda}: ${c.codigo} · ${c.descripcion}. Se aplica solo desde acá: ${linkDe(c.codigo)}`;
		window.open(`https://wa.me/?text=${encodeURIComponent(texto)}`, '_blank', 'noopener');
	}
	async function historia(c: C) {
		const texto = `${encabezado(c.tipo, c.valor)} en ${data.tienda} con el código ${c.codigo} · ${linkDe(c.codigo)}`;
		try {
			await navigator.clipboard.writeText(texto);
		} catch {
			// Sin permiso de portapapeles.
		}
		ui.toast({ k: 'ok', title: 'Texto para la historia copiado', desc: texto });
	}
	async function duplicar(c: C) {
		const r = await accion(c.id, 'duplicar');
		if (!r) return;
		const copia = (await r.json()) as C;
		orden = 3;
		filtro = 'todos';
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Duplicado',
			desc: `${copia.codigo} quedó pausado para que lo edites`
		});
	}
	async function alternar(c: C) {
		if (c.estado === 'terminado')
			return ui.toast({
				k: 'info',
				title: `${c.codigo} ya terminó`,
				desc: 'Duplicalo para usarlo de nuevo'
			});
		const activo = c.estado === 'activo';
		if (!(await accion(c.id, activo ? 'pausar' : 'activar'))) return;
		await invalidateAll();
		ui.toast({
			k: activo ? 'warn' : 'ok',
			title: `${c.codigo} ${activo ? 'pausado' : 'activo'}`,
			desc: activo ? 'Nadie lo puede usar hasta que lo actives' : 'Ya se puede usar',
			action: 'Deshacer',
			onAction: async () => {
				await accion(c.id, activo ? 'activar' : 'pausar');
				await invalidateAll();
			}
		});
	}

	// ---- Armador ----
	let armando = $state(false);
	const OBJETIVOS = [
		[
			'first',
			'Primera compra',
			'Que un visitante nuevo se anime',
			I.phHandWavingBold,
			{ tipo: 'fijo', valor: 3000, quien: 'nuevos', alcance: 'todo', codigo: 'HOLA3000' }
		],
		[
			'back',
			'Que vuelvan',
			'Clientes que hace 90 días no compran',
			I.phArrowUUpLeftBold,
			{ tipo: 'pct', valor: 15, quien: 'riesgo', alcance: 'todo', codigo: 'VOLVE15' }
		],
		[
			'stock',
			'Liquidar stock',
			'Productos que no rotan',
			I.phPackageBold,
			{ tipo: 'pct', valor: 20, quien: 'todos', alcance: 'productos', codigo: 'LIQUIDA20' }
		],
		[
			'infl',
			'Influencer',
			'Código propio para medir cada cuenta',
			I.phStarBold,
			{ tipo: 'pct', valor: 12, quien: 'todos', alcance: 'todo', codigo: 'INFLU-' }
		]
	] as const;
	const TIPOS = [
		['pct', 'Porcentaje', I.phPercentBold],
		['fijo', 'Monto fijo', I.phCurrencyDollarBold],
		['envio', 'Envío gratis', I.phTruckBold],
		['regalo', '2x1 / regalo', I.phGiftBold]
	] as const;
	const VIGENCIAS = [
		['7', '7 días'],
		['finde', 'Este finde'],
		['30', '30 días'],
		['hotsale', 'Hot Sale'],
		['sin', 'Sin vencimiento']
	] as const;
	const SUGERIDOS = ['OTONO15', 'VUELVE10', 'YAPA10', 'FINDELARGO', 'VOLVISTE', 'GRACIAS15'];
	let sugerido = 0;
	let objetivo = $state<string>('first');
	let n = $state<CuponNuevo>(nuevoCupon());
	const elegidos = new SvelteSet<string>();
	function nuevoCupon(): CuponNuevo {
		return {
			codigo: 'HOLA3000',
			tipo: 'fijo',
			valor: 3000,
			alcance: 'todo',
			productos: [],
			quien: 'nuevos',
			minimo: 15_000,
			limite: 200,
			porCliente: '1',
			vigencia: '30',
			canales: ['web', 'wa'],
			opciones: { hornero: true, combinable: false, unicos: false, link: false },
			programar: false
		};
	}
	function abrir() {
		n = nuevoCupon();
		objetivo = 'first';
		elegidos.clear();
		if (data.productos[0]) elegidos.add(data.productos[0].id);
		armando = true;
		tab = 'cup';
	}
	function elegirTipo(t: CuponNuevo['tipo']) {
		n.tipo = t;
		n.valor = t === 'pct' ? 15 : t === 'fijo' ? 3000 : 0;
	}
	const codigo = $derived((n.codigo || 'CODIGO').toUpperCase());
	const nombresElegidos = $derived(
		data.productos.filter((p) => elegidos.has(p.id)).map((p) => p.nombre)
	);
	const alcance = $derived(alcanceTexto(n.alcance, data.categoria, nombresElegidos));
	const sub = $derived(
		alcance[0].toUpperCase() +
			alcance.slice(1) +
			(n.minimo ? ` · compra mínima ${pesos(n.minimo)}` : '') +
			' · ' +
			vigenciaTexto(n.vigencia, ahora)
	);
	const est = $derived(costoPorUso(n, data.ticket));
	const OPCIONES = $derived([
		[
			'hornero',
			I.phSparkleBold,
			'Tincho lo puede ofrecer',
			'Lo usa en el chat para cerrar ventas o rescatar carritos'
		],
		[
			'combinable',
			I.phIntersectBold,
			'Combinable con otras promos',
			'Si está apagado, se aplica el mejor descuento y listo'
		],
		[
			'unicos',
			I.phFingerprintBold,
			'Códigos únicos en lote',
			'Genera 200 códigos de un solo uso para mandar por WhatsApp'
		],
		[
			'link',
			I.phLinkBold,
			'Link que lo aplica solo',
			`${linkDe(codigo).replace(/^https?:\/\//, '')} · ideal para historias`
		]
	] as const);
	let creando = $state(false);
	async function crear(programar: boolean) {
		if (creando) return;
		creando = true;
		const r = await pedir('/panel/cupones', 'POST', {
			...n,
			codigo,
			productos: n.alcance === 'productos' ? [...elegidos] : [],
			programar
		});
		creando = false;
		if (!r) return;
		armando = false;
		filtro = 'todos';
		orden = 3;
		await invalidateAll();
		if (programar)
			ui.toast({
				k: 'info',
				title: `${codigo} programado`,
				desc: 'Arranca el lunes a las 9 h. Te avisamos.'
			});
		else
			ui.toast({
				k: 'ok',
				title: `${codigo} está activo`,
				desc: `${encabezado(n.tipo, n.valor)} · ${vigenciaTexto(n.vigencia, ahora).toLowerCase()}`
			});
	}

	// ---- Automáticas ----
	const AUTO: Record<
		(typeof data.automaticas)[number]['id'],
		readonly [I.IconData, string, string]
	> = {
		transf: [I.phBankBold, '#2F8F63', 'rgba(47,143,99,.12)'],
		envio: [I.phTruckBold, '#3E6BFD', 'rgba(62,107,253,.12)'],
		cantidad: [I.phStackBold, '#D2582F', 'rgba(210,88,47,.12)'],
		'3x2': [I.phGiftBold, '#6A4FA0', 'rgba(106,79,160,.12)'],
		cuotas: [I.phCreditCardBold, '#00A88F', 'rgba(0,168,143,.12)'],
		mayorista: [I.phBuildingsBold, '#8A5A3E', 'rgba(138,90,62,.12)'],
		carrito: [I.phSparkleBold, '#C24E25', 'rgba(210,88,47,.12)'],
		cumple: [I.phCakeBold, '#E3A93B', 'rgba(227,169,59,.16)']
	};
	async function automatica(a: (typeof data.automaticas)[number]) {
		if (!(await pedir(`/panel/cupones/automaticas/${a.id}`, 'PUT', { activa: !a.activa }))) return;
		await invalidateAll();
		ui.toast({
			k: a.activa ? 'warn' : 'ok',
			title: `${a.titulo} ${a.activa ? 'apagada' : 'activa'}`,
			desc: a.activa ? 'Deja de aplicarse en el carrito' : 'Ya se aplica sola en el carrito'
		});
	}

	// ---- Relámpago ----
	const r = $derived(data.relampago.vivo);
	const quedan = $derived(
		r && !r.terminada ? Math.max(0, Math.floor((Date.parse(r.termina) - ahora) / 1000)) : 0
	);
	const reloj = $derived([
		[Math.floor(quedan / 3600), 'horas'],
		[Math.floor(quedan / 60) % 60, 'min'],
		[quedan % 60, 'seg']
	] as const);
	async function relampago(a: 'sumar-hora' | 'liberar' | 'terminar' | 'reanudar') {
		if (!(await pedir('/panel/cupones/relampago', 'POST', { accion: a }))) return;
		await invalidateAll();
		if (a === 'sumar-hora')
			ui.toast({ k: 'ok', title: 'Sumaste 1 hora', desc: 'La oferta sigue hasta más tarde' });
		if (a === 'liberar')
			ui.toast({
				k: 'ok',
				title: '20 unidades más',
				desc: `Quedan ${r!.unidades + 20 - r!.vendidos} a precio de oferta`
			});
		if (a === 'terminar')
			ui.toast({
				k: 'warn',
				title: 'Oferta terminada',
				desc: `${r!.producto} vuelve a ${pesos(r!.precio)}`,
				action: 'Deshacer',
				onAction: () => relampago('reanudar')
			});
	}
</script>

<Seo title="Cupones · Panel" description="Cupones y promociones" path="/panel/cupones" noindex />

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Cupones y promociones</h1>
			<div class="c-s">Creá descuentos, mirá cuánto venden y cortalos si no rinden.</div>
		</div>
		<button type="button" class="nuevo" onclick={abrir}
			><Icon icon={I.phPlusBold} size={15} />Nuevo cupón</button
		>
	</div>

	<div class="kpis" use:anim={{ t: 'rise', delay: 45 }}>
		{#each kpis as x (x.t)}
			<div class="kpi">
				<div class="k-t"><Icon icon={x.i} size={15} style="color:{x.c}" />{x.t}</div>
				<div class="k-v">{x.v}</div>
				<div class="k-d" style:color={x.dc}>{x.d}</div>
			</div>
		{/each}
	</div>

	{#if data.consejo}
		<div class="consejo" use:anim={{ t: 'rise', delay: 90 }}>
			<img src="/hornero/grafico.webp" alt="" width="56" height="56" />
			<div class="co-txt">
				<div class="co-t">{data.consejo.titulo}</div>
				<div class="co-d">{data.consejo.texto}</div>
			</div>
			<div class="co-acc">
				<button type="button" class="co-si" onclick={aplicarConsejo}
					>Poner mínimo {pesos(data.consejo.minimo)}</button
				>
				<button type="button" class="co-no" onclick={ahoraNo}>Ahora no</button>
			</div>
		</div>
	{/if}

	<div
		class="tabs"
		role="group"
		aria-label="Ver"
		use:anim={{ t: 'rise', delay: 135 }}
		use:deslizable
	>
		{#each TABS as [id, t, icono, cuantos] (id)}
			<button type="button" aria-pressed={tab === id} onclick={() => (tab = id)}
				><Icon icon={icono} size={15} />{t}<span>{cuantos}</span></button
			>
		{/each}
	</div>

	{#if armando}
		<div class="armador" use:anim={{ t: 'rise' }}>
			<div class="a-izq">
				<div class="a-cab">
					<span>Nuevo cupón</span>
					<button type="button" class="x" aria-label="Cerrar" onclick={() => (armando = false)}
						><Icon icon={I.phXBold} size={14} /></button
					>
				</div>

				<div>
					<div class="et">Para qué es</div>
					<div class="objetivos">
						{#each OBJETIVOS as [id, t, d, icono, p] (id)}
							<button
								type="button"
								class="pill caja"
								aria-pressed={objetivo === id}
								onclick={() => {
									objetivo = id;
									n = { ...n, ...p };
								}}
							>
								<Icon icon={icono} size={20} style="color:#b0441d" />
								<span class="o-t">{t}</span>
								<span class="o-d">{d}</span>
							</button>
						{/each}
					</div>
				</div>

				<div class="dos">
					<div>
						<div class="et">Código</div>
						<div class="codigo">
							<input
								value={n.codigo}
								oninput={(e) =>
									(n.codigo = e.currentTarget.value
										.toUpperCase()
										.replace(/[^A-Z0-9-]/g, '')
										.slice(0, 20))}
								aria-label="Código"
							/>
							<button
								type="button"
								onclick={() => (n.codigo = SUGERIDOS[sugerido++ % SUGERIDOS.length])}
								><Icon icon={I.phSparkleFill} size={14} style="color:#b0441d" />Sugerir</button
							>
						</div>
					</div>
					<div>
						<div class="et">Tipo de descuento</div>
						<div class="pills" role="group" aria-label="Tipo de descuento">
							{#each TIPOS as [id, t, icono] (id)}
								<button
									type="button"
									class="pill"
									aria-pressed={n.tipo === id}
									onclick={() => elegirTipo(id)}><Icon icon={icono} size={14} />{t}</button
								>
							{/each}
						</div>
					</div>
				</div>

				{#if n.tipo === 'pct' || n.tipo === 'fijo'}
					<div>
						<div class="rango-t">
							<span>{n.tipo === 'pct' ? 'Descuento' : 'Monto'}</span><b
								>{n.tipo === 'pct' ? `${n.valor} %` : pesos(n.valor)}</b
							>
						</div>
						<input
							type="range"
							min={n.tipo === 'pct' ? 5 : 500}
							max={n.tipo === 'pct' ? 50 : 20000}
							step={n.tipo === 'pct' ? 1 : 500}
							bind:value={n.valor}
							aria-label="Valor del descuento"
						/>
					</div>
				{/if}

				<div class="dos">
					<div>
						<div class="et">Aplica a</div>
						<div class="pills">
							{#each [['todo', 'Toda la tienda'], ['categoria', 'Una categoría'], ['productos', 'Productos']] as const as [id, t] (id)}
								<button
									type="button"
									class="pill"
									aria-pressed={n.alcance === id}
									onclick={() => (n.alcance = id)}>{t}</button
								>
							{/each}
						</div>
						{#if n.alcance === 'productos'}
							<div class="pills prods">
								{#each data.productos as p (p.id)}
									<button
										type="button"
										class="pill prod"
										aria-pressed={elegidos.has(p.id)}
										onclick={() =>
											elegidos.has(p.id) ? elegidos.delete(p.id) : elegidos.add(p.id)}
										>{#if p.foto}<img
												src={unsplash(p.foto, 60, '&h=60')}
												alt=""
											/>{/if}{p.nombre}</button
									>
								{/each}
							</div>
						{/if}
					</div>
					<div>
						<div class="et">Quién lo puede usar</div>
						<div class="pills">
							{#each [['todos', 'Todos', ''], ['nuevos', 'Primera compra', ''], ['vip', 'VIP', String(data.segmentos.vip)], ['riesgo', 'En riesgo', String(data.segmentos.riesgo)]] as const as [id, t, c] (id)}
								<button
									type="button"
									class="pill"
									aria-pressed={n.quien === id}
									onclick={() => (n.quien = id)}
									>{t}{#if c}<span class="fa">{c}</span>{/if}</button
								>
							{/each}
						</div>
					</div>
				</div>

				<div class="tres">
					<div>
						<div class="rango-t">
							<span>Compra mínima</span><b>{n.minimo ? pesos(n.minimo) : 'Sin mínimo'}</b>
						</div>
						<input
							type="range"
							min="0"
							max="100000"
							step="5000"
							bind:value={n.minimo}
							aria-label="Compra mínima"
						/>
					</div>
					<div>
						<div class="rango-t">
							<span>Usos en total</span><b>{n.limite ? `${n.limite} usos` : 'Sin límite'}</b>
						</div>
						<input
							type="range"
							min="0"
							max="500"
							step="25"
							bind:value={n.limite}
							aria-label="Usos en total"
						/>
					</div>
					<div>
						<div class="et">Usos por cliente</div>
						<div class="seg" role="group" aria-label="Usos por cliente">
							{#each [['1', '1 vez'], ['3', '3 veces'], ['libre', 'Libre']] as const as [id, t] (id)}
								<button
									type="button"
									aria-pressed={n.porCliente === id}
									onclick={() => (n.porCliente = id)}>{t}</button
								>
							{/each}
						</div>
					</div>
				</div>

				<div>
					<div class="et">Vigencia</div>
					<div class="pills">
						{#each VIGENCIAS as [id, t] (id)}
							<button
								type="button"
								class="pill"
								aria-pressed={n.vigencia === id}
								onclick={() => (n.vigencia = id)}>{t}</button
							>
						{/each}
					</div>
				</div>

				<div>
					<div class="et">Dónde vale</div>
					<div class="pills">
						{#each Object.entries(CANAL) as [id, [icono, col, t]] (id)}
							{@const c = id as keyof typeof CANAL}
							<button
								type="button"
								class="pill"
								aria-pressed={n.canales.includes(c)}
								onclick={() =>
									(n.canales = n.canales.includes(c)
										? n.canales.filter((x) => x !== c)
										: [...n.canales, c])}
								><Icon icon={icono} size={15} style="color:{col ?? 'inherit'}" />{t}</button
							>
						{/each}
					</div>
				</div>

				<div class="opciones">
					{#each OPCIONES as [id, icono, t, d] (id)}
						<div class="opcion">
							<Icon icon={icono} size={18} style="color:var(--mu);flex:none" />
							<span class="op-txt"><span class="op-t">{t}</span><span class="op-d">{d}</span></span>
							<Interruptor bind:checked={n.opciones[id]} label={t} />
						</div>
					{/each}
				</div>
			</div>

			<div class="a-der">
				<div class="et">Así lo ve tu cliente</div>
				<div class="ticket">
					<div class="tk-izq">
						<div class="tk-tienda">{data.tienda.toUpperCase()}</div>
						<div class="tk-h">{encabezado(n.tipo, n.valor)}</div>
						<div class="tk-s">{sub}</div>
					</div>
					<div class="tk-corte"></div>
					<div class="tk-der"><span>Código</span><b>{codigo}</b></div>
					<span class="muesca arriba"></span><span class="muesca abajo"></span>
				</div>
				<div class="calcula">
					<div class="ca-t">
						<Icon icon={I.phSparkleFill} size={14} style="color:#b0441d" />Lo que calcula Tincho
					</div>
					{#if est}
						<div class="ca-f">
							<span>{data.ticket ? 'Tu ticket promedio (30 días)' : 'Compra'}</span><b
								>{pesos(est.ticket)}</b
							>
						</div>
						<div class="ca-f"><span>Te cuesta cada uso</span><b>{pesos(est.costo)}</b></div>
					{:else if n.tipo === 'envio'}
						<div class="ca-n">Cada uso te cuesta lo que salga el envío de ese pedido.</div>
					{:else}
						<div class="ca-n">
							Cuando tengas ventas, te calculo cuánto te cuesta cada uso con tu ticket&nbsp;real.
						</div>
					{/if}
				</div>
				<button
					type="button"
					class="crear"
					disabled={creando || codigo.length < 3 || !n.canales.length}
					onclick={() => crear(false)}>Crear {codigo}</button
				>
				<button
					type="button"
					class="programar"
					disabled={creando || codigo.length < 3 || !n.canales.length}
					onclick={() => crear(true)}>Programar para más adelante</button
				>
			</div>
		</div>
	{/if}

	{#if tab === 'cup'}
		<div class="barra" use:anim={{ t: 'rise' }}>
			<label class="buscar"
				><Icon icon={I.phMagnifyingGlass} size={16} /><input
					bind:value={q}
					placeholder="Buscar código o producto"
					aria-label="Buscar cupón"
				/></label
			>
			<div class="filtros" role="group" aria-label="Estado" use:deslizable>
				{#each FILTROS as [id, t, c] (id)}
					<button
						type="button"
						class="pill"
						aria-pressed={filtro === id}
						onclick={() => (filtro = id)}
						><span class="punto" style:background={c}></span>{t}<span class="fa"
							>{id === 'todos'
								? data.cupones.length
								: data.cupones.filter((x) => x.estado === id).length}</span
						></button
					>
				{/each}
			</div>
			<button type="button" class="orden" onclick={() => (orden = (orden + 1) % ORDENES.length)}
				><Icon icon={I.phSortAscendingBold} size={15} />{ORDENES[orden][1]}</button
			>
		</div>
		<div class="cupones">
			{#each lista as c (c.id)}
				{@const [et, sbg, sfg, cab] = ESTADOS[c.estado]}
				{@const pct = c.limite ? Math.min(100, Math.round((c.usos / c.limite) * 100)) : null}
				{@const x = rinde(c)}
				<div class="cupon" class:terminado={c.estado === 'terminado'}>
					<div class="cu-cab" style:background={cab}>
						<div class="cu-izq">
							<div class="cu-cod">
								<span>{c.codigo}</span>
								<button type="button" aria-label="Copiar {c.codigo}" onclick={() => copiar(c)}
									><Icon icon={I.phCopyBold} size={13} /></button
								>
							</div>
							<div class="cu-desc">{c.descripcion}</div>
						</div>
						<div class="cu-corte"></div>
						<div class="cu-grande">{grande(c)}</div>
					</div>
					<div class="cu-cuerpo">
						<div class="cu-estado">
							<span class="chip" style:background={sbg} style:color={sfg}
								><span class="punto" style:background={sfg}></span>{et}</span
							>
							<span class="cu-cuando">{cuando(c, ahora)}</span>
							<span class="cu-canales"
								>{#each c.canales as k (k)}<Icon
										icon={CANAL[k][0]}
										size={15}
										style="color:{CANAL[k][1] ?? 'inherit'}"
									/>{/each}</span
							>
						</div>
						<div>
							<div class="usos">
								<span
									>{c.limite
										? `${c.usos} de ${c.limite} usos`
										: `${c.usos} usos · sin límite`}</span
								><b>{pct === null ? '' : `${pct} %`}</b>
							</div>
							<div class="barra-u">
								<div
									style:width="{pct ?? Math.min(100, c.usos / 2)}%"
									style:background={pct !== null && pct > 85 ? '#E3A93B' : '#D2582F'}
								></div>
							</div>
						</div>
						<div class="cifras">
							<div><span>Vendió</span><b>{c.vendio ? pesos(c.vendio) : '—'}</b></div>
							<div><span>Descontó</span><b>{c.desconto ? pesos(c.desconto) : '—'}</b></div>
							<div>
								<span>Rinde</span><b
									style:color={x >= 6 ? '#2F8F63' : x && x < 4 ? '#B83A2E' : 'var(--tx)'}
									>{x ? veces(x) : '—'}</b
								>
							</div>
						</div>
						{#if c.consejo}<div class="tip"><span>✦</span><span>{c.consejo}</span></div>{/if}
						<div class="cu-acc">
							<button type="button" class="mandar" onclick={() => mandar(c)}
								><Icon icon={I.phWhatsappLogoFill} size={14} />Mandar</button
							>
							<button type="button" class="hist" onclick={() => historia(c)}
								><Icon icon={I.phInstagramLogoBold} size={14} />Historia</button
							>
							<button
								type="button"
								class="dup"
								aria-label="Duplicar {c.codigo}"
								onclick={() => duplicar(c)}><Icon icon={I.phCopySimpleBold} size={14} /></button
							>
							<span class="tog"
								><Interruptor
									checked={c.estado === 'activo'}
									label="Activar o pausar {c.codigo}"
									onchange={() => alternar(c)}
								/></span
							>
						</div>
					</div>
				</div>
			{/each}
		</div>
		{#if !lista.length}
			<div class="vacio">
				<img src="/hornero/lupa.webp" alt="" width="100" height="100" />
				<div>No hay cupones con ese filtro</div>
			</div>
		{/if}
	{:else if tab === 'auto'}
		<div class="auto-d" use:anim={{ t: 'rise' }}>
			Se aplican solas en el carrito, sin que el cliente escriba nada.
		</div>
		<div class="autos">
			{#each data.automaticas as a (a.id)}
				{@const [icono, c, bg] = AUTO[a.id]}
				<div class="auto" class:apagada={!a.activa}>
					<div class="au-cab">
						<span class="au-ico" style:background={bg} style:color={c}
							><Icon icon={icono} size={22} /></span
						>
						<span class="au-txt"
							><span class="au-t">{a.titulo}</span><span class="au-d">{a.detalle}</span></span
						>
						<Interruptor
							checked={a.activa}
							label="Activar {a.titulo}"
							onchange={() => automatica(a)}
						/>
					</div>
					<div class="au-ej">
						<Icon icon={I.phShoppingCartSimpleBold} size={15} style="color:var(--mu)" /><span
							>{a.ejemplo}</span
						>
					</div>
					<div class="au-pie"><span>{a.estado}</span><b>{a.aporte ? pesos(a.aporte) : '—'}</b></div>
				</div>
			{/each}
		</div>
	{:else}
		<div class="flash" use:anim={{ t: 'rise' }}>
			{#if r}
				<div class="vivo">
					<div class="grano"></div>
					<div class="v-et">
						<span class="v-punto" class:off={r.terminada}></span>{r.terminada
							? 'TERMINÓ · OFERTA RELÁMPAGO'
							: 'EN VIVO · OFERTA RELÁMPAGO'}
					</div>
					<div class="v-prod">
						{#if r.foto}<img src={unsplash(r.foto, 200, '&h=200')} alt="" />{/if}
						<div>
							<div class="v-t">{r.producto} · {r.pct} % off</div>
							<div class="v-p"><s>{pesos(r.precio)}</s> · <b>{pesos(r.oferta)}</b></div>
						</div>
					</div>
					<div class="reloj">
						{#each reloj as [v, t] (t)}
							<div><b>{String(v).padStart(2, '0')}</b><span>{t}</span></div>
						{/each}
					</div>
					<div class="v-bar">
						<div class="vb-t">
							<span>Vendidos {r.vendidos} de {r.unidades}</span><b
								>Quedan {r.unidades - r.vendidos}</b
							>
						</div>
						<div class="vb"><div style:width="{(r.vendidos / r.unidades) * 100}%"></div></div>
					</div>
					<div class="v-acc">
						{#if r.terminada}
							<button type="button" class="claro" onclick={() => relampago('reanudar')}
								>Volver a ponerla</button
							>
						{:else}
							<button type="button" class="claro" onclick={() => relampago('sumar-hora')}
								>Sumar 1 hora</button
							>
							<button type="button" onclick={() => relampago('liberar')}>Liberar 20 más</button>
							<button type="button" onclick={() => relampago('terminar')}>Terminar ahora</button>
						{/if}
					</div>
				</div>
			{/if}
			<div class="proximas">
				<div class="px-t">Próximas</div>
				{#each data.relampago.proximas as p (p.titulo)}
					<div class="px">
						{#if p.foto}<img src={unsplash(p.foto, 92, '&h=92')} alt="" />{/if}
						<div class="px-txt">
							<div class="px-n">{p.titulo}</div>
							<div class="px-d">{p.detalle}</div>
						</div>
						<span class="px-c">{p.cuando}</span>
					</div>
				{/each}
				<button
					type="button"
					class="px-nueva"
					onclick={() =>
						ui.toast({
							k: 'info',
							title: 'Elegí el producto',
							desc: 'Programar ofertas relámpago desde acá llega con la tienda online'
						})}
					><Icon icon={I.phLightningBold} size={15} style="color:#b0441d" />Programar oferta
					relámpago</button
				>
			</div>
		</div>
	{/if}
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.cab {
		display: flex;
		align-items: end;
		gap: 12px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 240px;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.nuevo {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 18px;
		border-radius: 999px;
		border: 0;
		background: #b0441d;
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		box-shadow: 0 10px 24px -10px rgba(210, 88, 47, 0.7);
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
	.k-t {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 13px;
		color: var(--mu);
	}
	.k-v {
		margin-top: 4px;
		font-size: 26px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.k-d {
		font-size: 12.5px;
		font-weight: 500;
	}
	.consejo {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 14px 16px;
		border-radius: 20px;
		background: linear-gradient(100deg, #fcede5, #f7e2d3);
		color: #2a140c;
		flex-wrap: wrap;
	}
	.consejo img {
		object-fit: contain;
		flex: none;
	}
	.co-txt {
		flex: 1 1 300px;
		min-width: 0;
	}
	.co-t {
		font-size: 14.5px;
		font-weight: 600;
	}
	.co-d {
		font-size: 13.5px;
		color: #6b5244;
		margin-top: 2px;
		text-wrap: pretty;
	}
	.co-acc {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.co-acc button {
		padding: 9px 14px;
		border-radius: 999px;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.co-si {
		border: 0;
		background: #2a140c;
		color: #fff;
	}
	.co-no {
		border: 1px solid rgba(42, 20, 12, 0.18);
		background: transparent;
		color: #2a140c;
	}
	.tabs {
		display: flex;
		padding: 4px;
		border-radius: 999px;
		background: var(--chip);
		align-self: flex-start;
		max-width: 100%;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.tabs button {
		display: flex;
		align-items: center;
		gap: 7px;
		border: 0;
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
		padding: 9px 15px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
		white-space: nowrap;
	}
	.tabs button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.tabs span {
		color: var(--mu);
		font-weight: 500;
	}
	/* ---- Armador ---- */
	.armador {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 320px;
		gap: 16px;
		align-items: start;
		padding: 18px;
		border-radius: 26px;
		background: var(--sf);
		box-shadow: var(--shd);
		border: 1.5px solid rgba(210, 88, 47, 0.35);
	}
	.a-izq {
		display: flex;
		flex-direction: column;
		gap: 18px;
		min-width: 0;
	}
	.a-cab {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.a-cab span {
		font-size: 19px;
		font-weight: 600;
		letter-spacing: -0.02em;
		flex: 1;
	}
	.x {
		width: 34px;
		height: 34px;
		border-radius: 50%;
		border: 0;
		background: var(--chip);
		color: var(--tx);
		display: grid;
		place-items: center;
	}
	.et {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		margin-bottom: 7px;
	}
	.objetivos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 8px;
	}
	.pill {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		border-radius: 999px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.pill[aria-pressed='true'] {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	.pill.caja {
		flex-direction: column;
		align-items: flex-start;
		gap: 4px;
		padding: 12px;
		border-radius: 16px;
		text-align: left;
		white-space: normal;
	}
	.o-t {
		font-size: 13.5px;
		font-weight: 600;
	}
	.o-d {
		font-size: 12px;
		color: var(--mu);
		line-height: 1.35;
		font-weight: 500;
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 14px;
	}
	.tres {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 14px;
	}
	.codigo {
		display: flex;
		gap: 6px;
	}
	.codigo input {
		flex: 1;
		min-width: 0;
		padding: 11px 14px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: var(--sf);
		font:
			700 15px 'Instrument Sans',
			sans-serif;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--tx);
		outline: 0;
	}
	.codigo button {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 0 12px;
		border-radius: 14px;
		border: 0;
		background: var(--chip);
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.pills {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.pills.prods {
		margin-top: 8px;
	}
	.pill.prod {
		gap: 7px;
		padding: 4px 10px 4px 4px;
		font-size: 12px;
	}
	.pill.prod img {
		width: 24px;
		height: 24px;
		border-radius: 50%;
		object-fit: cover;
	}
	.fa {
		color: var(--fa);
		font-weight: 500;
	}
	.rango-t {
		display: flex;
		justify-content: space-between;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		margin-bottom: 6px;
	}
	.rango-t b {
		color: var(--tx);
		font-size: 15px;
	}
	input[type='range'] {
		width: 100%;
		accent-color: #d2582f;
	}
	.seg {
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
	}
	.seg button {
		flex: 1;
		border: 0;
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		padding: 7px 8px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
	}
	.seg button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.opciones {
		display: flex;
		flex-direction: column;
		gap: 2px;
		border-radius: 18px;
		background: var(--sf3);
		padding: 6px 14px;
	}
	.opcion {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.opcion:last-child {
		border-bottom: 0;
	}
	.op-txt {
		flex: 1;
		min-width: 0;
	}
	.op-t {
		display: block;
		font-size: 13.5px;
		font-weight: 600;
	}
	.op-d {
		display: block;
		font-size: 12px;
		color: var(--mu);
		margin-top: 1px;
		overflow-wrap: anywhere;
	}
	.a-der {
		display: flex;
		flex-direction: column;
		gap: 12px;
		position: sticky;
		top: 84px;
	}
	.a-der .et {
		margin: 0;
	}
	.ticket {
		position: relative;
		display: flex;
		border-radius: 20px;
		overflow: hidden;
		background: #2a140c;
		color: #fbf7f2;
		box-shadow: 0 24px 50px -24px rgba(42, 20, 12, 0.7);
	}
	.tk-izq {
		flex: 1;
		padding: 18px 18px 16px;
		min-width: 0;
	}
	.tk-tienda {
		font-size: 11px;
		font-weight: 700;
		letter-spacing: 0.12em;
		color: #e8875f;
	}
	.tk-h {
		margin-top: 6px;
		font-size: 30px;
		line-height: 1;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.tk-s {
		margin-top: 6px;
		font-size: 12.5px;
		color: #bdb2a8;
		line-height: 1.4;
	}
	.tk-corte {
		width: 0;
		border-left: 2px dashed rgba(251, 247, 242, 0.25);
		margin: 12px 0;
	}
	.tk-der {
		width: 96px;
		flex: none;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 4px;
		padding: 10px;
	}
	.tk-der span {
		font-size: 10px;
		color: #bdb2a8;
	}
	.tk-der b {
		font-size: 13px;
		letter-spacing: 0.05em;
		word-break: break-all;
		text-align: center;
	}
	.muesca {
		position: absolute;
		right: 94px;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: var(--sf);
	}
	.muesca.arriba {
		top: -9px;
	}
	.muesca.abajo {
		bottom: -9px;
	}
	.calcula {
		padding: 14px;
		border-radius: 18px;
		background: var(--sf3);
		display: flex;
		flex-direction: column;
		gap: 9px;
	}
	.ca-t {
		font-size: 13px;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 7px;
	}
	.ca-f {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 13px;
	}
	.ca-f span {
		color: var(--mu);
	}
	.ca-n {
		font-size: 13px;
		line-height: 1.45;
		color: var(--mu);
	}
	.crear {
		padding: 13px 18px;
		border-radius: 999px;
		border: 0;
		background: #b0441d;
		color: #fff;
		font:
			600 14.5px 'Instrument Sans',
			sans-serif;
	}
	.programar {
		padding: 11px 18px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.crear:disabled,
	.programar:disabled {
		opacity: 0.55;
	}
	/* ---- Lista ---- */
	.barra {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		align-items: center;
	}
	.buscar {
		flex: 1 1 220px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 14px;
		border-radius: 999px;
		background: var(--sf);
		box-shadow: var(--sh);
		color: var(--mu);
	}
	.buscar input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font:
			500 14px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
	}
	.filtros {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
		max-width: 100%;
	}
	.filtros .pill {
		background: var(--sf);
	}
	.filtros .pill[aria-pressed='true'] {
		background: rgba(210, 88, 47, 0.08);
	}
	.punto {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		flex: none;
	}
	.orden {
		display: flex;
		align-items: center;
		gap: 7px;
		padding: 9px 14px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.cupones {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 330px), 1fr));
		gap: 14px;
	}
	.cupon {
		display: flex;
		flex-direction: column;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		overflow: hidden;
		transition: opacity 0.25s;
	}
	.cupon.terminado {
		opacity: 0.72;
	}
	.cu-cab {
		position: relative;
		display: flex;
		align-items: stretch;
		color: #fbf7f2;
	}
	.cu-izq {
		flex: 1;
		min-width: 0;
		padding: 14px 16px;
	}
	.cu-cod {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.cu-cod span {
		font-size: 17px;
		font-weight: 700;
		letter-spacing: 0.05em;
	}
	.cu-cod button {
		width: 26px;
		height: 26px;
		border-radius: 8px;
		border: 0;
		background: rgba(255, 255, 255, 0.14);
		color: inherit;
		display: grid;
		place-items: center;
	}
	.cu-desc {
		font-size: 13px;
		opacity: 0.85;
		margin-top: 2px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.cu-corte {
		width: 0;
		border-left: 2px dashed rgba(255, 255, 255, 0.3);
		margin: 10px 0;
	}
	.cu-grande {
		width: 84px;
		flex: none;
		display: grid;
		place-items: center;
		font-size: 22px;
		font-weight: 700;
		letter-spacing: -0.03em;
	}
	.cu-cuerpo {
		padding: 14px 16px;
		display: flex;
		flex-direction: column;
		gap: 11px;
		flex: 1;
	}
	.cu-estado {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.chip {
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 12px;
		font-weight: 600;
		padding: 4px 9px;
		border-radius: 999px;
	}
	.chip .punto {
		width: 6px;
		height: 6px;
	}
	.cu-cuando {
		font-size: 12.5px;
		color: var(--mu);
	}
	.cu-canales {
		margin-left: auto;
		display: flex;
		gap: 4px;
	}
	.usos {
		display: flex;
		justify-content: space-between;
		font-size: 12.5px;
	}
	.usos span {
		color: var(--mu);
	}
	.barra-u {
		margin-top: 5px;
		height: 6px;
		border-radius: 4px;
		background: var(--sf3);
		overflow: hidden;
	}
	.barra-u div {
		height: 100%;
		border-radius: 4px;
	}
	.cifras {
		display: grid;
		grid-template-columns: 1fr 1fr 1fr;
		gap: 8px;
	}
	.cifras span {
		display: block;
		font-size: 11.5px;
		color: var(--mu);
	}
	.cifras b {
		display: block;
		font-size: 15px;
		font-weight: 600;
	}
	.tip {
		display: flex;
		gap: 8px;
		align-items: flex-start;
		padding: 9px 11px;
		border-radius: 12px;
		background: var(--sf3);
		font-size: 12.5px;
		line-height: 1.4;
	}
	.tip span:first-child {
		color: #b0441d;
		font-weight: 700;
	}
	.cu-acc {
		margin-top: auto;
		display: flex;
		align-items: center;
		gap: 6px;
		padding-top: 4px;
	}
	.cu-acc button {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		border-radius: 999px;
		border: 0;
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.mandar {
		background: #e9f8ee;
		color: #1f6f4a;
	}
	.hist {
		background: var(--chip);
		color: var(--tx);
	}
	.cu-acc .dup {
		width: 34px;
		height: 34px;
		padding: 0;
		justify-content: center;
		border-radius: 50%;
		background: var(--chip);
		color: var(--tx);
	}
	.tog {
		margin-left: auto;
		display: flex;
	}
	.vacio {
		padding: 40px;
		text-align: center;
		border-radius: 22px;
		background: var(--sf);
		font-size: 15px;
		font-weight: 600;
	}
	.vacio img {
		object-fit: contain;
	}
	/* ---- Automáticas ---- */
	.auto-d {
		font-size: 14px;
		color: var(--mu);
	}
	.autos {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 300px), 1fr));
		gap: 14px;
	}
	.auto {
		padding: 16px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.auto.apagada {
		opacity: 0.7;
	}
	.au-cab {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.au-ico {
		width: 44px;
		height: 44px;
		border-radius: 14px;
		flex: none;
		display: grid;
		place-items: center;
	}
	.au-txt {
		flex: 1;
		min-width: 0;
	}
	.au-t {
		display: block;
		font-size: 15px;
		font-weight: 600;
	}
	.au-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.au-ej {
		padding: 10px 12px;
		border-radius: 14px;
		background: var(--sf3);
		font-size: 13px;
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.au-ej span {
		flex: 1;
	}
	.au-pie {
		display: flex;
		justify-content: space-between;
		font-size: 12.5px;
		color: var(--mu);
	}
	.au-pie b {
		color: var(--tx);
	}
	/* ---- Relámpago ---- */
	.flash {
		display: grid;
		grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr);
		gap: 16px;
		align-items: start;
	}
	.vivo {
		position: relative;
		overflow: hidden;
		padding: 22px;
		border-radius: 26px;
		background: #2a140c;
		color: #fbf7f2;
		box-shadow: var(--shd);
	}
	.vivo > :not(.grano) {
		position: relative;
	}
	.grano {
		position: absolute;
		inset: 0;
		background: url('/textura/grain.webp');
		opacity: 0.15;
		pointer-events: none;
	}
	.v-et {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.1em;
		color: #e8875f;
	}
	.v-punto {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #e8875f;
		box-shadow: 0 0 0 4px rgba(232, 135, 95, 0.25);
	}
	.v-punto.off {
		background: #bdb2a8;
		box-shadow: none;
	}
	.v-prod {
		margin-top: 10px;
		display: flex;
		gap: 16px;
		align-items: center;
		flex-wrap: wrap;
	}
	.v-prod img {
		width: 84px;
		height: 84px;
		border-radius: 42px 42px 10px 10px;
		object-fit: cover;
	}
	.v-prod > div {
		flex: 1;
		min-width: 180px;
	}
	.v-t {
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.v-p {
		font-size: 14px;
		color: #bdb2a8;
		margin-top: 2px;
	}
	.v-p b {
		color: #fbf7f2;
	}
	.reloj {
		margin-top: 16px;
		display: flex;
		gap: 8px;
	}
	.reloj div {
		flex: 1;
		padding: 12px 6px;
		border-radius: 16px;
		background: rgba(251, 247, 242, 0.08);
		text-align: center;
	}
	.reloj b {
		display: block;
		font-size: 30px;
		font-weight: 600;
		letter-spacing: -0.03em;
		font-variant-numeric: tabular-nums;
	}
	.reloj span {
		font-size: 11.5px;
		color: #bdb2a8;
	}
	.v-bar {
		margin-top: 14px;
	}
	.vb-t {
		display: flex;
		justify-content: space-between;
		font-size: 13px;
	}
	.vb-t span {
		color: #bdb2a8;
	}
	.vb {
		margin-top: 6px;
		height: 8px;
		border-radius: 5px;
		background: rgba(251, 247, 242, 0.12);
		overflow: hidden;
	}
	.vb div {
		height: 100%;
		border-radius: 5px;
		background: #e8875f;
	}
	.v-acc {
		margin-top: 16px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.v-acc button {
		padding: 10px 14px;
		border-radius: 999px;
		border: 1px solid rgba(251, 247, 242, 0.25);
		background: transparent;
		color: #fbf7f2;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
	}
	.v-acc .claro {
		border: 0;
		background: #fbf7f2;
		color: #2a140c;
	}
	.proximas {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.px-t {
		font-size: 15px;
		font-weight: 600;
	}
	.px {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.px img {
		width: 46px;
		height: 46px;
		border-radius: 23px 23px 6px 6px;
		object-fit: cover;
		flex: none;
	}
	.px-txt {
		flex: 1;
		min-width: 0;
	}
	.px-n {
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.px-d {
		font-size: 12.5px;
		color: var(--mu);
	}
	.px-c {
		font-size: 12px;
		font-weight: 700;
		padding: 4px 9px;
		border-radius: 999px;
		background: var(--chip);
		white-space: nowrap;
	}
	.px-nueva {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 12px;
		border-radius: 18px;
		border: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	@media (max-width: 560px) {
		.kpis {
			grid-template-columns: 1fr 1fr;
		}
		.kpi {
			padding: 14px;
		}
		.k-v {
			font-size: 21px;
		}
	}
	@media (max-width: 999px) {
		.armador,
		.flash {
			grid-template-columns: minmax(0, 1fr);
		}
		.a-der {
			position: static;
		}
	}
</style>
