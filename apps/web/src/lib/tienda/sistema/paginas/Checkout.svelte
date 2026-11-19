<script lang="ts">
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { faltantes } from '../../faltantes.ts';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import CampoCupon from '../../CampoCupon.svelte';
	import EnvioCorreo from '../../EnvioCorreo.svelte';
	import { carritoDe } from '../../carrito.svelte.ts';
	import { combosDe, descuentoCombos } from '../../combos.ts';
	import { CuponCheckout, totales } from '../../cupon.svelte.ts';
	import { disenoDeTienda } from '../migrar.ts';
	import { cotizar, cpValido } from '../../envio.svelte.ts';
	import { envia } from '../../envio.ts';
	import { camposFactura, OPCIONALES } from '../../factura.ts';
	import type { OpcionEnvio, TiendaPublica } from '../../tipos.ts';
	import { conTransferencia, precio } from '../formato.ts';
	import { cuotasPara, ofreceDe, textoCuotas, vende } from '../cobros.ts';
	import BotonFinalizar from '../marco/BotonFinalizar.svelte';
	import Tienda from '../marco/Tienda.svelte';
	import Boton from '../primitivas/Boton.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * El checkout (A/B/C-celular › pago): 1 Datos, 2 Entrega y 3 Pago en una sola página, con el paso
	 * en el que se está marcado arriba; al costado (abajo en el celular), el pedido con el cupón y la
	 * cuenta, igual que la calcula el servidor (cupon.svelte.ts › totales). Factura A si la tienda la
	 * emite, «Quiero recibir novedades», Correo Argentino con provincia y sucursal (EnvioCorreo). El
	 * botón de pagar no se apaga por faltar algo: dice qué falta. Mercado Pago es otro sitio: se
	 * navega de verdad.
	 */
	let { tienda, error = null }: { tienda: TiendaPublica; error?: string | null } = $props();

	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const carrito = $derived(carritoDe(tienda.slug));
	const lineas = $derived(carrito.detalle(tienda.productos, tienda.mayorista ?? null));
	const subtotal = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));

	let email = $state('');
	let nombre = $state('');
	let telefono = $state('');
	let cp = $state('');
	let direccion = $state('');
	let envio = $state<OpcionEnvio['id'] | null>(null);
	// Arranca en Mercado Pago cuando está (como el handoff, A-celular): el descuento por
	// transferencia ya se ve en su opción.
	let pago = $state<'transferencia' | 'mercadopago'>(
		untrack(() => (tienda.cobros.mercadoPago ? 'mercadopago' : 'transferencia'))
	);
	let enviando = $state(false);
	let faltaEntrega = $state(false);
	// El aviso cuando el servidor no respondió (el formulario queda como estaba).
	let falloEnvio = $state<string | null>(null);
	// Si la tienda manda a algún lado (moto o Correo): si no, ni código postal ni dirección.
	const enviaTienda = $derived(envia(tienda.envios));

	const opciones = $derived(
		cpValido(cp)
			? cotizar(cp, subtotal, tienda)
			: tienda.envios.retiro
				? cotizar('0000', subtotal, tienda)
				: []
	);
	const elegida = $derived(opciones.find((o) => o.id === envio) ?? null);
	// Una sola forma de entrega posible (sólo retiro): queda elegida y se muestra como dato.
	const unica = $derived(!enviaTienda && opciones.length === 1 ? opciones[0] : null);
	$effect(() => {
		if (unica && envio !== unica.id) envio = unica.id;
		else if (envio && !opciones.some((o) => o.id === envio)) envio = null;
	});
	$effect(() => {
		if (envio) faltaEntrega = false;
	});

	const cupon = new CuponCheckout(() => tienda.slug);
	const cuenta = $derived(
		totales({
			subtotal,
			envio: elegida?.precio ?? 0,
			pago,
			transferencia: tienda.cobros.transferencia,
			cupon,
			combo: descuentoCombos(
				combosDe(disenoDeTienda(tienda)),
				lineas.map((l) => ({ ...l, precio: l.producto.precio }))
			).total
		})
	);

	const pagos = $derived([
		...(tienda.cobros.mercadoPago
			? [
					{
						id: 'mercadopago' as const,
						t: 'Mercado Pago',
						d: cuotasPara(ofreceDe(tienda), cuenta.total)
							? `Tarjetas, hasta ${ofreceDe(tienda).cuotas} cuotas sin interés, o dinero en cuenta`
							: 'Tarjetas de crédito y débito, o dinero en cuenta'
					}
				]
			: []),
		...(tienda.cobros.datosTransferencia
			? [
					{
						id: 'transferencia' as const,
						t: tienda.cobros.transferencia
							? `Transferencia · ${tienda.cobros.transferencia} % off`
							: 'Transferencia',
						d: 'Te mostramos el alias al confirmar'
					}
				]
			: [])
	]);
	$effect(() => {
		if (pagos.length && !pagos.some((x) => x.id === pago)) pago = pagos[0].id;
	});

	const CAMPOS = [
		['email', 'Email', 'email', 'email'],
		['nombre', 'Nombre y apellido', 'text', 'name'],
		['telefono', 'Teléfono', 'tel', 'tel'],
		['dni', 'DNI (opcional)', 'text', 'off']
	] as const;
	const factura = $derived(camposFactura(tienda));

	// En qué paso va (los de arriba): datos completos, entrega elegida, pago.
	const datosOk = $derived(
		/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()) &&
			nombre.trim().length >= 2 &&
			/^[\d\s+()-]{8,20}$/.test(telefono.trim())
	);
	const entregaOk = $derived(
		!!envio && (envio === 'retiro' || (cpValido(cp) && direccion.trim().length >= 4))
	);
	const paso = $derived(!datosOk ? 0 : !entregaOk ? 1 : 2);
	const PASOS = ['Datos', 'Entrega', 'Pago'];

	// Sin Mercado Pago ni transferencia no hay checkout: se pide por WhatsApp (BotonFinalizar).
	const sinCobro = $derived(!tienda.vitrina && !vende(tienda));
	// Lo agotado mientras tanto: se dice arriba del resumen y no se manda (pagar fallaría).
	const faltan = $derived(faltantes(carrito.lineas, tienda.productos));
	// Mercado Pago cobra ahí mismo; con transferencia se confirma y después se transfiere (C-cafe).
	const textoPagar = $derived(
		tienda.vitrina
			? 'Es una tienda de ejemplo'
			: cuenta.total <= 0
				? 'Confirmar pedido'
				: pago === 'mercadopago'
					? `Pagar ${precio(cuenta.total)}`
					: `Confirmar pedido · ${precio(cuenta.total)}`
	);
	// Los datos: el mensaje debajo de cada campo (no el globo del navegador, uno por vez y en su idioma).
	let tocado = $state(false);
	const errores = $derived({
		email: /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()) ? '' : 'Poné un email válido.',
		nombre: nombre.trim().length >= 2 ? '' : 'Poné tu nombre y apellido.',
		telefono: /^[\d\s+()-]{8,20}$/.test(telefono.trim())
			? ''
			: 'Poné un teléfono con característica.',
		cp:
			!enviaTienda || envio === 'retiro' || cpValido(cp)
				? ''
				: 'Poné los 4 números del código postal.',
		direccion:
			!envio || envio === 'retiro' || direccion.trim().length >= 4
				? ''
				: 'Poné la calle y el número.'
	} as Record<string, string>);
	const error1 = (n: string) => (tocado ? (errores[n] ?? '') : '');
	const extra = $derived(
		pago === 'mercadopago' && tienda.cobros.mercadoPago
			? textoCuotas(ofreceDe(tienda), cuenta.total)
			: ''
	);
	let entrega = $state<HTMLElement>();
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="checkout" sinWhatsApp>
	<div class="tv-ancho tv-pag tv-checkout">
		{#if page.data.pendienteMP}
			<!-- Volvió atrás desde Mercado Pago: ese pedido sigue reservado; pagarlo no arma otro. -->
			<div class="tv-confirma-caja tv-tarjeta tv-checkout-pendiente" role="status">
				<p><b>Tenés un pedido esperando el pago.</b> Pagalo con Mercado Pago o empezá uno nuevo.</p>
				<Boton href={page.data.pendienteMP}>Ver mi pedido</Boton>
			</div>
		{/if}
		{#if !lineas.length}
			<div class="tv-carrito-vacio tv-tarjeta">
				<Icon icon={I.phHandbag} size={28} />
				<p class="tv-lead">Tu carrito está vacío.</p>
				<Boton href="{base}/catalogo">Ver el catálogo</Boton>
			</div>
		{:else if sinCobro}
			<!-- Sin medios de cobro no hay formulario: el pedido va por WhatsApp (o se consulta). -->
			<div class="tv-carrito-vacio tv-tarjeta">
				<Icon icon={I.phChatCenteredText} size={28} />
				<p class="tv-lead">Esta tienda coordina el pago por mensaje.</p>
				<div class="tv-checkout-wa"><BotonFinalizar /></div>
			</div>
		{:else}
			<ol class="tv-pasos" aria-label="Pasos de la compra">
				{#each PASOS as p, i (p)}
					<li
						class:tv-paso--hecho={i < paso}
						class:tv-paso--ahora={i === paso}
						aria-current={i === paso ? 'step' : undefined}
					>
						<span>{i + 1}</span>{p}{#if i < paso}<Icon icon={I.phCheckBold} size={12} />{/if}
					</li>
				{/each}
			</ol>
			<form
				class="tv-checkout-cuerpo"
				method="POST"
				novalidate
				use:enhance={({ cancel, formElement }) => {
					falloEnvio = null;
					tocado = true;
					const mal = ['email', 'nombre', 'telefono', 'cp', 'direccion'].find((n) => errores[n]);
					if (mal || !envio || faltan.length) {
						cancel();
						if (!envio) faltaEntrega = true;
						const el = mal
							? formElement.querySelector<HTMLInputElement>(`[name="${mal}"]`)
							: !envio
								? entrega
								: null;
						el?.scrollIntoView({ behavior: 'smooth', block: 'center' });
						if (el instanceof HTMLInputElement) el.focus({ preventScroll: true });
						return;
					}
					enviando = true;
					return async ({ result, update }) => {
						enviando = false;
						// Mercado Pago es otro sitio: se navega de verdad.
						if (result.type === 'redirect' && result.location.startsWith('https://'))
							return void (location.href = result.location);
						// El servidor no respondió: el formulario queda como estaba (nada de la página de error).
						if (result.type === 'error') {
							falloEnvio = 'No pudimos confirmar el pedido. No se te cobró nada: probá de nuevo.';
							return;
						}
						if (result.type === 'redirect') carrito.vaciar();
						await update({ reset: false, navigate: false });
					};
				}}
			>
				<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
				<div class="tv-checkout-pasos">
					<section class="tv-checkout-paso" aria-labelledby="tv-co-datos">
						<h2 class="tv-h3" id="tv-co-datos">Tus datos</h2>
						<div class="tv-campos">
							{#each CAMPOS as [n, l, tipo, ac] (n)}
								<div class="tv-campo">
									<label for="tv-co-{n}">{l}</label>
									{#if n === 'email'}<input
											id="tv-co-{n}"
											name={n}
											type={tipo}
											autocomplete={ac}
											maxlength="120"
											aria-invalid={!!error1('email')}
											aria-describedby={error1('email') ? 'tv-co-email-e' : undefined}
											bind:value={email}
										/>
									{:else if n === 'nombre'}<input
											id="tv-co-{n}"
											name={n}
											type={tipo}
											autocomplete={ac}
											maxlength="80"
											aria-invalid={!!error1('nombre')}
											aria-describedby={error1('nombre') ? 'tv-co-nombre-e' : undefined}
											bind:value={nombre}
										/>
									{:else if n === 'telefono'}<input
											id="tv-co-{n}"
											name={n}
											type={tipo}
											autocomplete={ac}
											maxlength="20"
											aria-invalid={!!error1('telefono')}
											aria-describedby={error1('telefono') ? 'tv-co-telefono-e' : undefined}
											bind:value={telefono}
										/>
									{:else}<input
											id="tv-co-{n}"
											name={n}
											type={tipo}
											autocomplete={ac}
											maxlength="12"
											inputmode="numeric"
										/>{/if}
									{#if error1(n)}<p class="tv-error tv-campo-error" id="tv-co-{n}-e">
											{error1(n)}
										</p>{/if}
								</div>
							{/each}
							{#each factura as [n, l, tipo, ac] (n)}
								<div class="tv-campo">
									<label for="tv-co-{n}">{l}</label>
									<input
										id="tv-co-{n}"
										name={n}
										type={tipo}
										autocomplete={ac}
										required={!OPCIONALES.has(n)}
										maxlength="120"
									/>
								</div>
							{/each}
						</div>
						<label class="tv-casilla tv-checkout-nl"
							><input type="checkbox" name="novedades" /><span>Quiero recibir novedades</span
							></label
						>
					</section>

					<section class="tv-checkout-paso" aria-labelledby="tv-co-entrega" bind:this={entrega}>
						<h2 class="tv-h3" id="tv-co-entrega">¿Cómo lo recibís?</h2>
						{#if unica}
							<!-- Una sola forma (sólo retiro): elegida y como dato, con dónde y cuándo una sola vez. -->
							<div class="tv-radio tv-tarjeta tv-radio--on tv-radio--dato">
								<input type="hidden" name="envio" value={unica.id} />
								<span class="tv-radio-txt"
									><b>{unica.nombre}</b><span class="tv-chico tv-suave">{unica.detalle}</span></span
								>
								<b class="tv-radio-precio tv-ok">Gratis</b>
							</div>
						{:else}
							{#if enviaTienda}
								<div class="tv-campos tv-campos--envio">
									<div class="tv-campo">
										<label for="tv-co-cp">Código postal</label>
										<input
											id="tv-co-cp"
											name="cp"
											bind:value={cp}
											inputmode="numeric"
											maxlength="4"
											autocomplete="postal-code"
											aria-invalid={!!error1('cp')}
											aria-describedby={error1('cp') ? 'tv-co-cp-e' : undefined}
										/>
										{#if error1('cp')}<p class="tv-error tv-campo-error" id="tv-co-cp-e">
												{error1('cp')}
											</p>{/if}
									</div>
								</div>
							{/if}
							<div class="tv-radios" role="radiogroup" aria-label="Forma de entrega">
								{#each opciones as s (s.id)}
									<label class="tv-radio tv-tarjeta" class:tv-radio--on={envio === s.id}>
										<input type="radio" name="envio" value={s.id} bind:group={envio} />
										<span class="tv-radio-txt"
											><b>{s.nombre}</b><span class="tv-chico tv-suave">{s.detalle}</span></span
										>
										<b class="tv-radio-precio" class:tv-ok={!s.precio}
											>{s.precio ? precio(s.precio) : 'Gratis'}</b
										>
									</label>
								{:else}
									<p class="tv-suave tv-chico">
										Poné tu código postal para ver cómo te&nbsp;llega.
									</p>
								{/each}
							</div>
							{#if enviaTienda && envio && envio !== 'retiro'}
								<div class="tv-campos">
									<div class="tv-campo">
										<label for="tv-co-dir">Dirección (calle, número, piso)</label>
										<input
											id="tv-co-dir"
											name="direccion"
											bind:value={direccion}
											maxlength="160"
											autocomplete="street-address"
											aria-invalid={!!error1('direccion')}
											aria-describedby={error1('direccion') ? 'tv-co-dir-e' : undefined}
										/>
										{#if error1('direccion')}<p class="tv-error tv-campo-error" id="tv-co-dir-e">
												{error1('direccion')}
											</p>{/if}
									</div>
								</div>
							{:else}
								<input type="hidden" name="direccion" value="" />
							{/if}
						{/if}
						{#if !enviaTienda}<input type="hidden" name="cp" value="" />{/if}
						{#if faltaEntrega}<p class="tv-error" role="alert">
								Elegí cómo te llega para seguir.
							</p>{/if}
						{#if !unica}<EnvioCorreo {tienda} {cp} {envio} {opciones} />{/if}
					</section>

					<section class="tv-checkout-paso" aria-labelledby="tv-co-pago">
						<h2 class="tv-h3" id="tv-co-pago">¿Cómo pagás?</h2>
						<!-- Un cupón que cubre todo: no hay nada que cobrar (Go no pasa por Mercado Pago). -->
						{#if cuenta.total <= 0}<p class="tv-ok tv-chico">
								No hay nada que pagar: tu cupón cubre todo.
							</p>{/if}
						<div
							class="tv-radios"
							role="radiogroup"
							aria-label="Forma de pago"
							hidden={cuenta.total <= 0}
						>
							{#each pagos as s (s.id)}
								<label class="tv-radio tv-tarjeta" class:tv-radio--on={pago === s.id}>
									<input type="radio" name="pago" value={s.id} bind:group={pago} />
									<span class="tv-radio-txt"
										><b>{s.t}</b><span class="tv-chico tv-suave">{s.d}</span></span
									>
								</label>
							{:else}
								<p class="tv-suave tv-chico">
									Esta tienda todavía no cargó cómo cobrar. Escribile para coordinar.
								</p>
							{/each}
						</div>
					</section>

					{#if falloEnvio || error}<p class="tv-error tv-error--caja" role="alert">
							{falloEnvio ?? error}
						</p>{/if}
					{#if faltan.length}<p class="tv-error tv-error--caja" role="alert">
							Algo de tu carrito se agotó mientras tanto. <a class="tv-link" href="{base}/carrito"
								>Revisá el carrito</a
							> para seguir.
						</p>{/if}
					<div class="tv-solo-compu">
						<button
							class="tv-btn tv-btn--primario tv-btn--bloque tv-checkout-pagar"
							disabled={!pagos.length || enviando || tienda.vitrina}
							><Icon icon={I.phLockSimpleFill} size={16} />{enviando
								? 'Un momento…'
								: textoPagar}</button
						>
					</div>
				</div>

				<aside class="tv-checkout-resumen tv-tarjeta" aria-label="Tu pedido">
					<div class="tv-checkout-resumen-cab">
						<h2 class="tv-h3">Tu pedido</h2>
						<a class="tv-link" href="{base}/carrito">Cambiar</a>
					</div>
					<ul class="tv-checkout-lineas">
						{#each lineas as l (l.clave)}
							<li>
								<Foto
									src={l.producto.foto}
									alt=""
									nombre={l.producto.nombre}
									ancho={160}
									proporcion="1/1"
								/>
								<span
									>{l.producto.nombre}<small class="tv-suave"
										>{l.variante ? `${l.variante} · ` : ''}× {l.cantidad}</small
									></span
								>
								<span class="tv-checkout-sub">{precio(l.subtotal)}</span>
							</li>
						{/each}
					</ul>
					<CampoCupon
						{cupon}
						lineas={carrito.lineas}
						{email}
						tinta="var(--t-texto)"
						suave="var(--t-suave)"
						borde="var(--t-borde)"
						fondo="var(--t-fondo)"
						acento="var(--t-texto)"
						sobreAcento="var(--t-fondo)"
						radio="var(--t-r-boton)"
						fuente="var(--t-f-texto)"
					/>
					<dl class="tv-cuentas">
						<div>
							<dt>Subtotal</dt>
							<dd>{precio(subtotal)}</dd>
						</div>
						<div>
							<dt>Envío{elegida ? ` · ${elegida.nombre}` : ''}</dt>
							<dd>{elegida ? (cuenta.envio ? precio(cuenta.envio) : 'Gratis') : '—'}</dd>
						</div>
						{#if cuenta.descCombo}<div class="tv-ok">
								<dt>Combo</dt>
								<dd>− {precio(cuenta.descCombo)}</dd>
							</div>{/if}
						{#if cuenta.descCupon}<div class="tv-ok">
								<dt>Cupón {cupon.aplicado}</dt>
								<dd>− {precio(cuenta.descCupon)}</dd>
							</div>{/if}
						{#if cuenta.transf}<div class="tv-ok">
								<dt>Transferencia −{tienda.cobros.transferencia} %</dt>
								<dd>− {precio(cuenta.transf)}</dd>
							</div>{/if}
					</dl>
					<div class="tv-cuentas-total"><span>Total</span><b>{precio(cuenta.total)}</b></div>
					{#if extra}<p class="tv-chico tv-suave">o {extra}</p>{/if}
					{#if !cuenta.transf && tienda.cobros.transferencia > 0 && tienda.cobros.datosTransferencia && pago !== 'transferencia'}<p
							class="tv-chico tv-suave"
						>
							{precio(
								conTransferencia(
									subtotal - cuenta.descCombo - cuenta.descCupon,
									tienda.cobros.transferencia
								) + cuenta.envio
							)} pagando con transferencia
						</p>{/if}
					<p class="tv-chico tv-suave tv-carrito-seguro">
						<Icon icon={I.phArrowUUpLeft} size={16} /><a
							class="tv-link"
							href="{base}/arrepentimiento">Botón de arrepentimiento</a
						>&nbsp;· 10 días para&nbsp;devolverlo
					</p>
				</aside>

				<div class="tv-reserva-barra tv-solo-movil" aria-hidden="true"></div>
				<div class="tv-ficha-barra tv-solo-movil">
					<button
						class="tv-btn tv-btn--primario tv-btn--bloque"
						disabled={!pagos.length || enviando || tienda.vitrina}
						>{enviando ? 'Un momento…' : textoPagar}</button
					>
				</div>
			</form>
		{/if}
	</div>
</Tienda>
