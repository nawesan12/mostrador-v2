<script lang="ts" module>
	export type LineaListo = {
		nombre: string;
		foto?: string | null;
		variante?: string | null;
		cantidad: number;
		precio: number;
	};
	export type PedidoListo = {
		numero: number;
		total: number;
		pago: string;
		nombre: string;
		/** Los datos para transferir que cargó la tienda (sólo con transferencia). */
		alias: string | null;
		cbu?: string | null;
		titular?: string | null;
		/** El código del pedido: con él, la reseña queda con compra. */
		codigo?: string;
		/** En qué está (EstadoPedido): arma el título y los pasos. */
		estado?: string;
		/** retiro, moto, correo-domicilio, correo-sucursal… */
		envio?: string;
		seguimiento?: string | null;
		/** Ya está pago (Mercado Pago acreditado o la transferencia confirmada). */
		pagado?: boolean;
		/** Con Mercado Pago y todavía sin pagar (Mercado Pago no armó el pago o lo rechazó): se puede pagar acá. */
		porPagar?: boolean;
		/** Volvió de Mercado Pago con el pago rechazado. */
		rechazado?: boolean;
		items?: LineaListo[];
		subtotal?: number;
		costoEnvio?: number;
		entrega?: { direccion: string; cp: string; localidad?: string; sucursal?: string } | null;
		/** La API no respondió: sólo el número (sale del código); el tiempo real lo completa. */
		sinDatos?: boolean;
	};
</script>

<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import FormResena from '../../FormResena.svelte';
	import SenaPedido from '../../SenaPedido.svelte';
	import type { TiendaPublica } from '../../tipos.ts';
	import { waHref } from '../datos.ts';
	import { precio, sinViuda } from '../formato.ts';
	import Tienda from '../marco/Tienda.svelte';
	import { disenoDeTienda } from '../migrar.ts';
	import { temaDe } from '../paginas.ts';
	import { estiloFinal } from '../presets.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * La página del pedido (la que se guarda y se vuelve a abrir, y la del «Seguí tu pedido» del
	 * mail), igual para los 17 presets: el título y el ícono según en qué está (recién comprado,
	 * falta pagar, rechazado, preparando, listo, en camino, entregado o cancelado), los pasos, los
	 * datos para transferir con «Copiar», lo que se compró, la entrega con el seguimiento y el
	 * WhatsApp de la tienda. El tiempo real la actualiza sola (depends 'tienda:pedido-publico').
	 */
	let { tienda, pedido }: { tienda: TiendaPublica; pedido: PedidoListo } = $props();
	const diseno = $derived(disenoDeTienda(tienda));
	const dir = $derived(estiloFinal(tienda.plantilla, diseno.paleta, diseno.estilo).direccion);
	const tema = $derived(temaDe(dir));
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const retiro = $derived(pedido.envio === 'retiro');
	const estado = $derived(pedido.estado ?? 'Sin pagar');
	const cancelado = $derived(estado === 'Cancelado');
	const recibido = $derived(estado === 'Entregado' || estado === 'Retirado');
	const transfiere = $derived(
		pedido.pago === 'transferencia' && !pedido.pagado && !cancelado && estado === 'Sin pagar'
	);

	// Los pasos: el retiro no pasa por «En camino».
	const PASOS = $derived(
		retiro
			? ['Recibido', 'Preparando', 'Para retirar', 'Retirado']
			: ['Recibido', 'Preparando', 'En camino', 'Entregado']
	);
	const paso = $derived(
		({
			'Sin pagar': 0,
			'Para preparar': 1,
			'Listo para despachar': retiro ? 2 : 1,
			'En camino': 2,
			Entregado: 3,
			Retirado: 3
		})[estado] ?? 0
	);

	const cab = $derived.by((): { ico: I.IconData; titulo: string; texto: string; alerta?: boolean } => {
		const n = `#${pedido.numero}`;
		if (pedido.sinDatos)
			return {
				ico: I.phCheckBold,
				titulo: 'Recibimos tu pedido',
				texto: `Tu pedido ${n} quedó guardado. En un momento se actualiza esta página.`
			};
		if (cancelado)
			return {
				ico: I.phXBold,
				titulo: 'Este pedido se canceló',
				texto: `El pedido ${n} ya no sigue. Si lo pagaste, escribinos y te devolvemos la plata.`,
				alerta: true
			};
		if (pedido.porPagar && pedido.rechazado)
			return {
				ico: I.phWarningBold,
				titulo: 'El pago no se completó',
				texto: `Tu pedido ${n} sigue guardado 24 horas. Probá de nuevo, con otra tarjeta o medio.`,
				alerta: true
			};
		if (pedido.porPagar)
			return {
				ico: I.phClockBold,
				titulo: 'Falta pagar tu pedido',
				texto: `Te guardamos el pedido ${n} 24 horas para que lo pagues con Mercado Pago.`
			};
		if (transfiere)
			return {
				ico: I.phCheckBold,
				titulo: `¡Gracias, ${pedido.nombre}!`,
				texto: `Recibimos tu pedido ${n}. Transferí ${precio(pedido.total)} y lo preparamos apenas se acredite.`
			};
		if (estado === 'Para preparar' || estado === 'Sin pagar')
			return {
				ico: I.phCheckBold,
				titulo: `¡Gracias, ${pedido.nombre}!`,
				texto: `Tu pedido ${n} ya está pago y lo estamos preparando. Te escribimos por mail con cada novedad.`
			};
		if (estado === 'Listo para despachar')
			return retiro
				? {
						ico: I.phStorefrontBold,
						titulo: 'Ya podés pasar a buscarlo',
						texto: `Tu pedido ${n} está listo para retirar.`
					}
				: {
						ico: I.phPackageBold,
						titulo: 'Tu pedido está listo',
						texto: `Preparamos el pedido ${n}: sale en el próximo envío.`
					};
		if (estado === 'En camino')
			return {
				ico: I.phTruckBold,
				titulo: 'Tu pedido está en camino',
				texto: `El pedido ${n} ya salió.`
			};
		return {
			ico: I.phCheckBold,
			titulo: retiro ? 'Retiraste tu pedido' : 'Tu pedido llegó',
			texto: `¡Gracias por comprar! Contanos qué te pareció.`
		};
	});

	const waComprobante = $derived(
		waHref(tienda, `¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)
	);
	const waConsulta = $derived(
		waHref(tienda, `¡Hola! Te escribo por el pedido #${pedido.numero}.`)
	);

	// Los datos para transferir, cada uno con «Copiar» (desde el celular se pega en la app del banco).
	const datos = $derived(
		[
			pedido.alias ? ['Alias', pedido.alias] : null,
			pedido.cbu ? ['CBU/CVU', pedido.cbu] : null,
			pedido.titular ? ['Titular', pedido.titular] : null,
			['Monto', precio(pedido.total), String(Math.round(pedido.total))]
		].filter((x): x is string[] => !!x)
	);
	let copiado = $state('');
	async function copiar(etiqueta: string, valor: string) {
		try {
			await navigator.clipboard.writeText(valor);
			copiado = etiqueta;
			setTimeout(() => {
				if (copiado === etiqueta) copiado = '';
			}, 2000);
		} catch {
			copiado = '';
		}
	}

	const items = $derived(pedido.items ?? []);
	const donde = $derived.by(() => {
		if (retiro) {
			const d = tienda.envios.retiroDireccion?.trim();
			const h = tienda.envios.retiroHorario?.trim();
			return {
				titulo: 'Retiro en el local',
				lineas: d
					? [d, h].filter(Boolean)
					: [waConsulta ? 'Te pasamos la dirección por WhatsApp' : 'Te escribimos para coordinar']
			};
		}
		const e = pedido.entrega;
		if (!e) return null;
		return {
			titulo: e.sucursal ? 'Retirás en la sucursal' : 'Lo mandamos a',
			lineas: [e.sucursal || e.direccion, [e.localidad, e.cp && `CP ${e.cp}`].filter(Boolean).join(' · ')].filter(
				Boolean
			)
		};
	});
	const esCorreo = $derived(pedido.envio?.startsWith('correo') ?? false);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="listo">
	<section class="tv-ancho tv-pag tv-confirma">
		<span class="tv-confirma-ico" class:tv-confirma-ico--alerta={cab.alerta}
			><Icon icon={cab.ico} size={28} /></span
		>
		<h1 class="tv-h1 tv-pag-titulo">{cab.titulo}</h1>
		<p class="tv-lead">{sinViuda(cab.texto)}</p>

		{#if !cancelado && !pedido.porPagar && !pedido.sinDatos}
			<ol class="tv-pasos tv-confirma-pasos" aria-label="En qué está tu pedido">
				{#each PASOS as p, i (p)}
					<li
						class:tv-paso--hecho={i < paso}
						class:tv-paso--ahora={i === paso}
						aria-current={i === paso ? 'step' : undefined}
					>
						<span>{#if i < paso}<Icon icon={I.phCheckBold} size={12} />{:else}{i + 1}{/if}</span>{p}
					</li>
				{/each}
			</ol>
		{/if}

		<SenaPedido {tema} />

		{#if pedido.porPagar}
			<form class="tv-confirma-caja tv-tarjeta" method="POST" action="?/pagar">
				<p><b>Falta pagar {precio(pedido.total)}</b> con Mercado&nbsp;Pago.</p>
				<button class="tv-btn tv-btn--primario tv-btn--bloque" type="submit"
					>Pagar con Mercado&nbsp;Pago</button
				>
			</form>
		{:else if transfiere}
			<div class="tv-confirma-caja tv-tarjeta tv-transf" aria-labelledby="tv-transf-t">
				<h2 class="tv-h3" id="tv-transf-t">Datos para transferir</h2>
				{#if datos.length > 1}
					<dl class="tv-transf-datos">
						{#each datos as [etiqueta, valor, copia] (etiqueta)}
							<div class="tv-transf-fila">
								<dt>{etiqueta}</dt>
								<dd class:tv-confirma-alias={etiqueta !== 'Titular' && etiqueta !== 'Monto'}>
									{valor}
								</dd>
								<button
									type="button"
									class="tv-btn tv-btn--secundario tv-btn--chico"
									aria-label="Copiar {etiqueta.toLowerCase()}"
									onclick={() => copiar(etiqueta, copia ?? valor)}
									>{copiado === etiqueta ? 'Copiado' : 'Copiar'}</button
								>
							</div>
						{/each}
					</dl>
				{:else}
					<p>Te mandamos los datos por mail.</p>
				{/if}
				<p class="tv-oculto" aria-live="polite">{copiado ? `${copiado} copiado` : ''}</p>
				{#if waComprobante}
					<Boton href={waComprobante} tipo="wa" bloque externo
						><Icon icon={I.phWhatsappLogo} size={18} />Mandanos el comprobante</Boton
					>
				{:else}
					<p class="tv-chico tv-suave">Mandanos el comprobante respondiendo el mail del pedido.</p>
				{/if}
			</div>
		{/if}

		{#if items.length}
			<div class="tv-confirma-caja tv-tarjeta">
				<h2 class="tv-h3">Tu pedido #{pedido.numero}</h2>
				<ul class="tv-checkout-lineas">
					{#each items as l, i (i)}
						<li>
							<Foto src={l.foto ?? null} alt="" nombre={l.nombre} ancho={160} proporcion="1/1" />
							<span
								>{l.nombre}<small class="tv-suave"
									>{l.variante ? `${l.variante} · ` : ''}× {l.cantidad}</small
								></span
							>
							<span class="tv-checkout-sub">{precio(l.precio * l.cantidad)}</span>
						</li>
					{/each}
				</ul>
				<dl class="tv-cuentas">
					{#if pedido.subtotal !== undefined}<div>
							<dt>Subtotal</dt>
							<dd>{precio(pedido.subtotal)}</dd>
						</div>{/if}
					{#if pedido.costoEnvio !== undefined && !retiro}<div>
							<dt>Envío</dt>
							<dd>{pedido.costoEnvio ? precio(pedido.costoEnvio) : 'Gratis'}</dd>
						</div>{/if}
				</dl>
				<div class="tv-cuentas-total"><span>Total</span><b>{precio(pedido.total)}</b></div>
			</div>
		{/if}

		{#if donde && !cancelado}
			<div class="tv-confirma-caja tv-tarjeta">
				<h2 class="tv-h3">{donde.titulo}</h2>
				{#each donde.lineas as l (l)}<p>{l}</p>{/each}
				{#if pedido.seguimiento}
					<p class="tv-chico">
						Código de seguimiento: <b class="tv-confirma-alias">{pedido.seguimiento}</b>
					</p>
					{#if esCorreo}<Boton
							href="https://www.correoargentino.com.ar/formularios/e-commerce"
							tipo="secundario"
							externo>Seguir el envío en Correo&nbsp;Argentino</Boton
						>{/if}
				{/if}
			</div>
		{/if}

		<div class="tv-botones tv-confirma-botones">
			{#if waConsulta && !transfiere}
				<Boton href={waConsulta} tipo={cancelado ? 'primario' : 'secundario'} externo
					><Icon icon={I.phWhatsappLogo} size={18} />Escribile a la tienda</Boton
				>
			{/if}
			<Boton href={base || '/'} tipo={cancelado && waConsulta ? 'secundario' : 'primario'}
				>Volver a la tienda</Boton
			>
		</div>
		{#if recibido}
			<div class="tv-confirma-resena">
				<FormResena slug={tienda.slug} {tema} codigo={pedido.codigo} nombre={pedido.nombre} />
			</div>
		{/if}
	</section>
</Tienda>
