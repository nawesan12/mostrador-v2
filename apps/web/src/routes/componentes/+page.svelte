<script lang="ts">
	import Alerta from '#lib/components/Alerta.svelte';
	import ChipEstado from '#lib/components/ChipEstado.svelte';
	import Hoja from '#lib/components/Hoja.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import Grano from '#lib/components/Grano.svelte';
	import Vacio from '#lib/components/Vacio.svelte';
	import {
		phBellRingingFill,
		phDotsThreeBold,
		phPackageFill,
		phPauseCircleFill,
		phPrinter,
		phSidebarSimpleFill,
		phSparkleFill,
		phWarningFill,
		phX,
		phXBold
	} from '#lib/icons.ts';
	import { BARRA, type ClaveMarca, type Estado } from '#lib/marcas.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	const ui = getUi();

	let hoja = $state(false);
	let confirmar = $state(false);
	let pausarMl = $state(true);
	let periodo = $state<'hoy' | '7d' | '30d'>('7d');
	let tema = $state<'light' | 'dark'>('light');

	const TOASTS: { k: ClaveMarca; title: string; desc: string; action: string }[] = [
		{
			k: 'ok',
			title: 'Pedido #1048 preparado',
			desc: 'Le avisamos a Lucía por WhatsApp',
			action: 'Deshacer'
		},
		{
			k: 'wa',
			title: 'Nuevo pedido por WhatsApp',
			desc: 'Lucía · Termo 1 L negro · $ 42.900',
			action: 'Ver pedido'
		},
		{
			k: 'warn',
			title: 'Stock bajo',
			desc: 'Termo 1 L · Negro: quedan 3 unidades',
			action: 'Reponer'
		},
		{
			k: 'err',
			title: 'Pago rechazado',
			desc: 'La tarjeta de Carla no tiene fondos. Le mandamos otro link.',
			action: 'Ver detalle'
		},
		{
			k: 'ai',
			title: 'Tincho tiene una idea',
			desc: 'Recuperá 12 carritos por $ 1.440',
			action: 'Ver'
		}
	];
	const barra = (k: ClaveMarca) => BARRA[k] ?? (k === 'wa' ? BARRA.info : '#D2582F');

	const ALERTAS = [
		[
			'info',
			'Nueva función: cobros con MODO',
			'Activala en Integraciones y aceptá pagos con apps bancarias.',
			'Activar'
		],
		[
			'ok',
			'Tu tienda está publicada',
			'Tu dominio ya está en línea con certificado de seguridad.',
			'Ver tienda'
		],
		[
			'warn',
			'Tu certificado de ARCA vence en 5 días',
			'Renovalo para seguir emitiendo facturas sin cortes.',
			'Renovar'
		],
		[
			'err',
			'Mercado Libre se desconectó',
			'No estamos sincronizando stock. Volvé a conectar tu cuenta.',
			'Reconectar'
		]
	] as const;

	const ESTADOS_MUESTRA: Estado[] = [
		'Para preparar',
		'Pagado',
		'Pago pendiente',
		'Entregado',
		'Cancelado'
	];
	const SIN_STOCK = [
		'Termo 1 L · Rojo',
		'Mate imperial',
		'Yerbera de cuero',
		'Bombilla pico de loro'
	];

	function probarToast() {
		ui.toast({
			k: 'wa',
			title: 'Nuevo pedido por WhatsApp',
			desc: 'Lucía · Termo 1 L negro · $ 42.900',
			action: 'Ver pedido',
			onAction: () => {
				hoja = true;
			}
		});
	}

	function pausar() {
		confirmar = false;
		ui.undoable(
			'4 productos pausados',
			pausarMl ? 'También en Mercado Libre' : 'Sólo en tu tienda',
			() => {
				ui.toast({ k: 'info', title: 'Listo, volvieron a estar a la venta' });
			}
		);
	}
</script>

<svelte:head>
	<title>Componentes · Tu Local Web</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<Grano />

<div class="pagina" data-theme={tema}>
	<div class="etiqueta">Sistema de componentes</div>
	<h1>Avisos, modales y hojas</h1>
	<p class="bajada">
		Cómo habla Tu Local Web cuando pasa algo. Todo se puede probar en vivo dentro del panel.
	</p>

	<div class="probar">
		<button type="button" class="btn btn-dark" onclick={probarToast}>
			<Icon icon={phBellRingingFill} size={16} />Toast
		</button>
		<button type="button" class="btn btn-secondary" onclick={() => (hoja = true)}>
			<Icon icon={phSidebarSimpleFill} size={16} />Hoja
		</button>
		<button type="button" class="btn btn-secondary" onclick={() => (confirmar = true)}>
			<Icon icon={phWarningFill} size={16} style="color:#8a5f0e" />Confirmación
		</button>
		<Segmentado
			label="Tema"
			opciones={[
				['light', 'Claro'],
				['dark', 'Oscuro']
			]}
			bind:value={tema}
		/>
	</div>

	<div class="grilla">
		<section>
			<h2>Toasts</h2>
			<p class="nota">
				Abajo a la derecha, máximo 3 a la vez, se van solos a los 5 s. Siempre con una acción útil.
			</p>
			<div class="lista">
				{#each TOASTS as t (t.title)}
					<div class="toast-muestra">
						<MarcaIcono k={t.k} />
						<div class="cuerpo">
							<div class="t">{t.title}</div>
							<div class="d">{t.desc}</div>
							<span class="accion">{t.action}</span>
						</div>
						<Icon icon={phX} size={15} style="color:#9A8576" />
						<span class="vida" style:background={barra(t.k)}></span>
					</div>
				{/each}
			</div>
		</section>

		<section>
			<h2>Alertas en página</h2>
			<p class="nota">
				Para lo que no se puede ignorar. Se quedan hasta que se resuelven o se descartan.
			</p>
			<div class="lista">
				{#each ALERTAS as [tipo, titulo, texto, accion] (titulo)}
					<Alerta
						{tipo}
						{titulo}
						{texto}
						{accion}
						onaccion={() => ui.toast({ k: tipo, title: accion, desc: titulo })}
					/>
				{/each}
			</div>

			<h2 class="separado">Estados vacíos</h2>
			<div class="vacios">
				<Vacio
					hornero="horno"
					titulo="Todavía no hay pedidos"
					texto="Compartí tu tienda por WhatsApp y Tincho atiende."
				>
					<button type="button" class="btn btn-primary btn-sm">Compartir tienda</button>
				</Vacio>
				<Vacio
					hornero="lupa"
					titulo="Sin resultados para «termo azul»"
					texto="Probá con otra palabra o preguntale a Tincho."
				>
					<button type="button" class="btn btn-soft btn-sm">Limpiar búsqueda</button>
				</Vacio>
			</div>
		</section>
	</div>

	<h2 class="separado">Botones y estados</h2>
	<div class="botones card">
		<button type="button" class="btn btn-primary">Principal</button>
		<button type="button" class="btn btn-dark">Oscuro</button>
		<button type="button" class="btn btn-secondary">Secundario</button>
		<button type="button" class="btn btn-soft">Suave</button>
		<button type="button" class="btn btn-danger">Destructivo</button>
		<span class="sep" aria-hidden="true"></span>
		{#each ESTADOS_MUESTRA as e (e)}<ChipEstado estado={e} />{/each}
	</div>

	<h2 class="separado">Controles</h2>
	<div class="botones card">
		<Segmentado
			label="Período"
			opciones={[
				['hoy', 'Hoy'],
				['7d', '7 días'],
				['30d', '30 días']
			]}
			bind:value={periodo}
		/>
		<label class="fila-toggle">
			<Interruptor label="También pausar en Mercado Libre" bind:checked={pausarMl} />
			<span>También pausar en Mercado Libre</span>
		</label>
		<span class="sep" aria-hidden="true"></span>
		{#each ['wa', 'mp', 'ml', 'ig', 'tg', 'arca', 'andreani', 'oca', 'modo', 'uala', 'payway'] as k (k)}
			<MarcaIcono {k} size={34} radius={11} />
		{/each}
	</div>
</div>

<Hoja bind:open={hoja} titulo="Pedido #1048">
	{#snippet cabecera()}
		<div class="hc">
			<div class="hc-fila">
				<span class="hc-titulo">Pedido #1048</span>
				<ChipEstado estado="Para preparar" />
			</div>
			<div class="hc-sub">Hoy 19:38 · por WhatsApp</div>
		</div>
		<button type="button" class="icon-btn" aria-label="Imprimir"
			><Icon icon={phPrinter} size={17} /></button
		>
		<button type="button" class="icon-btn" aria-label="Cerrar" onclick={() => (hoja = false)}>
			<Icon icon={phXBold} size={15} />
		</button>
	{/snippet}
	<div class="nota-hornero">
		<Icon icon={phSparkleFill} size={16} style="color:#b0441d;margin-top:2px" />
		<div><b>Tincho:</b> pidió que vaya envuelto para regalo. Ya lo anoté en la etiqueta.</div>
	</div>
	<div class="bloque">
		<div class="bloque-t">Productos</div>
		{#each [['Termo 1 L · Negro', '1 u', '$ 42.900'], ['Bombilla de acero', '1 u', '$ 1.700']] as [n, q, t] (n)}
			<div class="item">
				<span class="item-ico"><Icon icon={phPackageFill} size={19} /></span>
				<span class="item-n"><b>{n}</b><small>{q}</small></span>
				<b>{t}</b>
			</div>
		{/each}
		<div class="total"><span>Total</span><span>$ 44.600</span></div>
	</div>
	{#snippet pie()}
		<button
			type="button"
			class="btn btn-primary btn-lg"
			style="flex:1"
			onclick={() => (hoja = false)}
		>
			Marcar como preparado
		</button>
		<button type="button" class="icon-btn grande" aria-label="Más acciones">
			<Icon icon={phDotsThreeBold} size={18} />
		</button>
	{/snippet}
</Hoja>

<Modal bind:open={confirmar} titulo="¿Pausar 4 productos sin stock?">
	<div class="conf-ico"><Icon icon={phPauseCircleFill} size={26} /></div>
	<div class="conf-t">¿Pausar 4 productos sin stock?</div>
	<div class="conf-d">
		Se ocultan de tu tienda hasta que cargues stock. Podés deshacerlo durante 30 días.
	</div>
	<div class="conf-lista">
		{#each SIN_STOCK as p (p)}<div><span>{p}</span><span>0 u</span></div>{/each}
	</div>
	<label class="conf-ml">
		<MarcaIcono k="ml" size={26} radius={7} />
		<span>También pausar en Mercado Libre</span>
		<Interruptor label="También pausar en Mercado Libre" bind:checked={pausarMl} />
	</label>
	<div class="conf-acc">
		<button type="button" class="btn btn-secondary" onclick={() => (confirmar = false)}
			>Cancelar</button
		>
		<button type="button" class="btn btn-primary" onclick={pausar}>Pausar productos</button>
	</div>
</Modal>

<style>
	.pagina {
		max-width: 1240px;
		margin: 0 auto;
		padding: 56px 28px 140px;
		background: var(--crema);
		color: var(--tx);
		min-height: 100vh;
	}
	.pagina[data-theme='dark'] {
		background: var(--bg);
		max-width: none;
	}
	h1 {
		margin: 12px 0 0;
		font-size: clamp(36px, 4.6vw, 56px);
		line-height: 1;
		letter-spacing: -0.04em;
		font-weight: 600;
	}
	.bajada {
		margin: 14px 0 0;
		max-width: 620px;
		font-size: 17px;
		line-height: 1.55;
		color: var(--mu);
	}
	.probar {
		margin-top: 22px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		align-items: center;
	}
	.grilla {
		margin-top: 56px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(360px, 100%), 1fr));
		gap: 40px;
	}
	h2 {
		margin: 0;
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.separado {
		margin-top: 36px;
	}
	.nota {
		margin: 4px 0 0;
		font-size: 14px;
		color: var(--fa);
	}
	.lista {
		margin-top: 18px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.toast-muestra {
		position: relative;
		overflow: hidden;
		max-width: 380px;
		display: flex;
		gap: 12px;
		padding: 14px 12px 16px 14px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow:
			0 1px 0 #fff inset,
			0 20px 44px -20px rgba(42, 20, 12, 0.45),
			0 0 0 1px rgba(58, 29, 18, 0.07);
	}
	.toast-muestra .cuerpo {
		flex: 1;
		min-width: 0;
	}
	.toast-muestra .t {
		font-size: 14px;
		font-weight: 600;
	}
	.toast-muestra .d {
		margin-top: 2px;
		font-size: 13px;
		line-height: 1.4;
		color: var(--mu);
	}
	.toast-muestra .accion {
		display: inline-block;
		line-height: 1.2;
		margin-top: 10px;
		padding: 6px 11px;
		border-radius: 999px;
		background: #f4ece2;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--tostado);
	}
	.toast-muestra .vida {
		position: absolute;
		left: 0;
		bottom: 0;
		height: 3px;
		width: 62%;
	}
	.vacios {
		margin-top: 14px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(220px, 100%), 1fr));
		gap: 10px;
	}
	.botones {
		margin-top: 14px;
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
		align-items: center;
		padding: 22px;
	}
	.sep {
		width: 1px;
		height: 28px;
		margin: 0 6px;
		background: var(--ln);
	}
	.fila-toggle {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 14px;
		font-weight: 500;
	}

	/* Hoja de muestra */
	.hc {
		flex: 1;
		min-width: 0;
	}
	.hc-fila {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}
	.hc-titulo {
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.hc-sub {
		margin-top: 2px;
		font-size: 13px;
		color: var(--mu);
	}
	.nota-hornero {
		display: flex;
		gap: 10px;
		padding: 12px 14px;
		border-radius: 16px;
		background: rgba(210, 88, 47, 0.1);
		border: 1px solid rgba(210, 88, 47, 0.22);
		font-size: 13.5px;
		line-height: 1.45;
	}
	.bloque {
		padding: 14px 16px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.bloque-t {
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	.item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
		font-size: 14px;
	}
	.item-ico {
		width: 40px;
		height: 40px;
		flex: none;
		display: grid;
		place-items: center;
		border-radius: 11px;
		background: var(--sf3);
		color: var(--terracota-oscuro);
	}
	.item-n {
		flex: 1;
		min-width: 0;
		display: grid;
	}
	.item-n small {
		font-size: 12.5px;
		color: var(--mu);
	}
	.total {
		display: flex;
		justify-content: space-between;
		padding-top: 10px;
		font-size: 17px;
		font-weight: 600;
	}
	.grande {
		width: 46px;
		height: 46px;
		background: transparent;
		border: 1px solid var(--ln2);
	}

	/* Confirmación */
	.conf-ico {
		width: 48px;
		height: 48px;
		display: grid;
		place-items: center;
		border-radius: 15px;
		background: rgba(227, 169, 59, 0.2);
		color: #8a5f0e;
	}
	.conf-t {
		margin-top: 16px;
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.conf-d {
		margin-top: 6px;
		font-size: 14.5px;
		line-height: 1.5;
		color: var(--mu);
	}
	.conf-lista {
		margin-top: 16px;
		border-radius: 16px;
		border: 1px solid var(--ln);
		overflow: hidden;
	}
	.conf-lista div {
		display: flex;
		justify-content: space-between;
		padding: 10px 14px;
		font-size: 13.5px;
		border-bottom: 1px solid var(--ln);
	}
	.conf-lista div:last-child {
		border-bottom: 0;
	}
	.conf-lista span:last-child {
		color: var(--fa);
	}
	.conf-ml {
		margin-top: 14px;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 14px;
		border-radius: 14px;
		background: var(--sf3);
		font-size: 14px;
		font-weight: 500;
	}
	.conf-ml span {
		flex: 1;
	}
	.conf-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
		flex-wrap: wrap;
	}

	@media (max-width: 760px) {
		.pagina {
			padding: 32px 16px 120px;
		}
		.botones {
			padding: 16px;
		}
	}
</style>
