<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { getShell } from '#lib/panel/shell.svelte.ts';

	/** Ayuda del panel (design/PanelOperacion.dc.html, sHelp): Tincho, guías y contacto. */
	let { data } = $props();
	const shell = getShell();
	let pregunta = $state('');

	const art = (categoria: string, articulo: string) =>
		resolve('/(sitio)/ayuda/[categoria]/[articulo]', { categoria, articulo });
	const GUIAS = [
		[
			I.phRocketLaunchFill,
			'Primeros pasos',
			'5 minutos para tu primera venta',
			'#FCEDE5',
			'#C24E25',
			art('primeros-pasos', 'crear-tu-tienda')
		],
		[
			I.phReceiptFill,
			'Facturar con ARCA',
			'Certificado, tipos A, B y C',
			'#E6ECF7',
			'#1F3A6E',
			resolve('/(sitio)/ayuda/[categoria]', { categoria: 'facturacion' })
		],
		[
			I.phHandshakeFill,
			'Vender en Mercado Libre',
			'Precios, stock y preguntas',
			'#FFF6C2',
			'#6B5B00',
			art('productos', 'mercado-libre')
		],
		[
			I.phWhatsappLogoFill,
			'WhatsApp y Tincho',
			'Qué responde y cuándo te avisa',
			'#E9F8EE',
			'#1F6F4A',
			art('whatsapp', 'que-responde-hornero')
		],
		[
			I.phTruckFill,
			'Envíos y etiquetas',
			'Andreani, OCA y moto propia',
			'#FDF3DD',
			'#9A6A10',
			art('pedidos', 'envios-por-codigo-postal')
		],
		[
			I.phArrowsLeftRightBold,
			'Migrar desde otra plataforma',
			'Te lo hacemos gratis',
			'#EFE9F7',
			'#6A4FA0',
			resolve('/(sitio)/migrar')
		]
	] as const;
	const disponible = $derived(
		data.estado?.disponibilidad == null
			? 'Mirá cómo anda cada servicio'
			: `${data.estado.todoBien ? 'Todo funcionando' : 'Hay un servicio con problemas'} · ${data.estado.disponibilidad.toLocaleString('es-AR', { maximumFractionDigits: 2 })}% en 90 días`
	);
</script>

<Seo title="Ayuda · Panel" description="Ayuda de Tu Local Web" path="/panel/ayuda" noindex />

<div class="pagina">
	<div class="hero" use:anim={{ t: 'rise' }}>
		<div class="h-txt">
			<h1>¿En qué te ayudo?</h1>
			<form
				class="buscar"
				onsubmit={(e) => {
					e.preventDefault();
					shell.hornero(pregunta.trim() || 'Cómo hago una factura A');
					pregunta = '';
				}}
			>
				<Icon icon={I.phMagnifyingGlass} size={17} style="color:#7A6252" />
				<input
					bind:value={pregunta}
					placeholder="Ej: cómo hago una factura A"
					aria-label="Tu pregunta"
				/>
				<button>Preguntar</button>
			</form>
		</div>
		<img src="/hornero/lupa.webp" alt="" width="130" height="130" />
	</div>

	<div class="guias" use:anim={{ t: 'rise', delay: 60 }}>
		{#each GUIAS as [icono, t, d, bg, fg, href] (t)}
			<a class="guia" {href}>
				<span class="g-ico" style:background={bg} style:color={fg}
					><Icon icon={icono} size={19} /></span
				>
				<span><span class="g-t">{t}</span><span class="g-d">{d}</span></span>
			</a>
		{/each}
	</div>

	<div class="contactos" use:anim={{ t: 'rise', delay: 120 }}>
		<button type="button" class="contacto" onclick={() => shell.hornero()}>
			<Icon icon={I.phSparkleFill} size={24} style="color:#b0441d" />
			<div class="c-t">Tincho</div>
			<div class="c-d">Responde al instante, 24 h</div>
		</button>
		<a class="contacto" href={resolve('/(sitio)/contacto')}>
			<Icon icon={I.phEnvelopeSimpleFill} size={24} style="color:#b0441d" />
			<div class="c-t">Soporte</div>
			<div class="c-d">Personas reales · te respondemos en el día</div>
		</a>
		<a class="contacto" href={resolve('/(sitio)/estado')}>
			<Icon
				icon={I.phActivityBold}
				size={24}
				style="color:{data.estado?.todoBien === false ? '#B83A2E' : '#2F8F63'}"
			/>
			<div class="c-t">Estado del sistema</div>
			<div class="c-d">{disponible}</div>
		</a>
	</div>
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.hero {
		position: relative;
		overflow: hidden;
		padding: 28px;
		border-radius: 26px;
		background: linear-gradient(140deg, #d2582f, #b9481f);
		color: #fff;
		display: flex;
		align-items: center;
		gap: 20px;
		flex-wrap: wrap;
	}
	.h-txt {
		position: relative;
		flex: 1 1 300px;
	}
	h1 {
		margin: 0;
		font-size: 32px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.buscar {
		margin-top: 14px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 6px 6px 16px;
		border-radius: 999px;
		background: #fff;
	}
	.buscar input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font:
			500 15px 'Instrument Sans',
			sans-serif;
		color: #2a140c;
	}
	.buscar button {
		padding: 10px 16px;
		border-radius: 999px;
		border: 0;
		background: #2a140c;
		color: #fff;
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.hero img {
		position: relative;
		object-fit: contain;
	}
	.guias,
	.contactos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
		gap: 12px;
	}
	.guia {
		text-align: left;
		padding: 16px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		color: var(--tx);
		display: flex;
		gap: 12px;
		align-items: flex-start;
		text-decoration: none;
	}
	.g-ico {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 12px;
		display: grid;
		place-items: center;
	}
	.g-t {
		display: block;
		font-size: 14.5px;
		font-weight: 600;
	}
	.g-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		margin-top: 2px;
	}
	.contacto {
		text-align: left;
		padding: 18px;
		border-radius: 20px;
		border: 0;
		background: var(--sf);
		box-shadow: var(--shd);
		color: var(--tx);
		text-decoration: none;
		font: inherit;
	}
	.c-t {
		margin-top: 10px;
		font-size: 15px;
		font-weight: 600;
	}
	.c-d {
		font-size: 13px;
		color: var(--mu);
	}
	@media (max-width: 560px) {
		.hero {
			padding: 22px;
		}
		.hero img {
			width: 96px;
			height: 96px;
		}
	}
</style>
