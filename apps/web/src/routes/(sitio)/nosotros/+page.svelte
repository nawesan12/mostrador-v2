<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import Etiqueta from '#lib/sitio/Etiqueta.svelte';
	import Titulo from '#lib/sitio/Titulo.svelte';

	/**
	 * Sobre nosotros (design/Mostrador Nosotros.dc.html). Sólo lo que es cierto: quién lo hace y lo
	 * que la plataforma hace hoy. Sin cifras, historia ni personas de ejemplo.
	 */
	const foto = (id: string, extra: string) =>
		`https://images.unsplash.com/photo-${id}?auto=format&fit=crop&q=72&${extra}`;
	const HITOS = [
		['01', 'Tu tienda, con tu dominio', 'Tu marca y tu dirección, sin comisiones por venta.'],
		['02', 'Cobrás y facturás desde el panel', 'Mercado Pago para cobrar y ARCA para facturar.'],
		['03', 'Tincho te da una mano', 'Responde lo de todos los días con los datos de tu negocio.'],
		['04', 'Tus datos son tuyos', 'Los exportás cuando quieras, en CSV o Excel.']
	] as const;
	const VALORES = [
		[
			I.phCurrencyCircleDollarFill,
			'Precio claro',
			'Un abono por mes, sin comisiones escondidas ni apps que se pagan aparte.'
		],
		[I.phShieldCheckFill, 'Seguridad de serie', 'Lo que protege tu negocio no es un extra.'],
		[I.phHandHeartFill, 'Gente que atiende', 'Te responde una persona del equipo, no un bot.'],
		[
			I.phMapTrifoldFill,
			'Hecho para Argentina',
			'Pesos, cuotas, Mercado Pago, ARCA y la forma en que se vende acá.'
		]
	] as const;
</script>

<Seo
	title="Sobre nosotros · Tu Local Web"
	description="Hacemos que vender sea más barato y más simple para el que empieza. Por qué existe Tu Local Web, lo que nos importa y quién lo hace."
	path="/nosotros"
/>

<div class="ancho">
	<div class="intro">
		<Etiqueta texto="Sobre nosotros" />
		<h1>
			Hacemos que vender sea <span>más barato y más simple</span> para el que empieza.
		</h1>
	</div>
	<div class="mosaico">
		<div class="grande"><img src={foto('1522071820081-009f0129c71c', 'w=1400')} alt="" /></div>
		<div class="costado">
			<div class="chica"><img src={foto('1556761175-5973dc0f32e7', 'w=700')} alt="" /></div>
			<div class="cifra">
				<div class="c-v">Grupo Vimo</div>
				<div class="c-t">hace Tu Local Web, desde Argentina</div>
			</div>
		</div>
	</div>
</div>

<section class="bloque">
	<div class="ancho">
		<div class="historia">
			<div>
				<Titulo etiqueta="Por qué existe" titulo="Vender no debería costar tanto" />
				<p class="h-txt">
					Para vender online hoy se pagan comisiones, apps sueltas y planes que no conversan entre
					sí. Armamos Tu Local Web para que un comercio tenga todo en un solo lugar: tienda, cobros,
					facturas, pedidos y stock, por un abono&nbsp;claro.
				</p>
			</div>
			<div class="hitos">
				{#each HITOS as [a, t, d] (a)}
					<div class="hito">
						<span>{a}</span>
						<div>
							<b>{t}</b>
							<div class="hi-d">{d}</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
</section>

<section class="bloque beige">
	<div class="ancho">
		<Titulo etiqueta="Lo que nos importa" titulo="Cuatro cosas que no negociamos" />
		<div class="valores">
			{#each VALORES as [ico, t, d] (t)}
				<div class="valor">
					<Icon icon={ico} size={28} style="color:#b0441d" />
					<div class="v-t">{t}</div>
					<p>{d}</p>
				</div>
			{/each}
		</div>
	</div>
</section>

<section class="bloque">
	<div class="ancho">
		<div class="trabaja">
			<div>
				<Etiqueta texto="Quiénes somos" color="#F0A58A" />
				<div class="t-t">Tu Local Web es un producto de Grupo Vimo.</div>
				<p>
					Somos un equipo chico que hace productos digitales en Argentina. Si querés escribirnos,
					hablar de tu negocio o trabajar con nosotros, estamos en&nbsp;contacto.
				</p>
			</div>
			<div class="busquedas">
				<!-- eslint-disable svelte/no-navigation-without-resolve -- páginas públicas -->
				<a href="/contacto">
					<span class="b-txt"><b>Escribinos</b><span>Te responde una persona del equipo</span></span
					>
					<Icon icon={I.phArrowUpRightBold} size={15} />
				</a>
				<a href="https://grupovimo.com.ar" rel="noopener">
					<span class="b-txt"><b>Grupo Vimo</b><span>grupovimo.com.ar</span></span>
					<Icon icon={I.phArrowUpRightBold} size={15} />
				</a>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
			</div>
		</div>
	</div>
</section>

<style>
	.ancho {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 clamp(16px, 3vw, 28px);
	}
	.intro {
		padding: clamp(48px, 7vw, 104px) 0 0;
		max-width: 980px;
	}
	h1 {
		margin: 14px 0 0;
		font-size: clamp(44px, 7vw, 100px);
		line-height: 0.92;
		letter-spacing: -0.055em;
		font-weight: 600;
	}
	h1 span {
		color: #b0441d;
	}
	.mosaico {
		margin-top: clamp(28px, 4vw, 48px);
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 10px;
	}
	.grande {
		aspect-ratio: 16 / 9;
		border-radius: 28px;
		overflow: hidden;
		background: #e9ddcf;
	}
	.costado {
		display: grid;
		grid-template-rows: 1fr 1fr;
		gap: 10px;
	}
	.chica {
		border-radius: 28px;
		overflow: hidden;
		background: #e9ddcf;
	}
	.mosaico img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.cifra {
		border-radius: 28px;
		background: #b0441d;
		color: #fff;
		padding: 20px;
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
	}
	.c-v {
		font-size: clamp(32px, 4vw, 52px);
		font-weight: 600;
		letter-spacing: -0.04em;
		line-height: 1;
	}
	.c-t {
		font-size: 14px;
		opacity: 0.9;
	}
	.bloque {
		padding: clamp(56px, 8vw, 104px) 0;
	}
	.beige {
		background: #f4ece2;
	}
	.historia {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 360px), 1fr));
		gap: clamp(20px, 4vw, 56px);
	}
	.h-txt {
		margin: 16px 0 0;
		font-size: 17px;
		line-height: 1.65;
		color: #5c4536;
	}
	.hitos {
		display: flex;
		flex-direction: column;
	}
	.hito {
		display: grid;
		grid-template-columns: 72px 1fr;
		gap: 16px;
		padding: 16px 0;
	}
	.hito + .hito {
		border-top: 1px solid rgba(58, 29, 18, 0.1);
	}
	.hito > span {
		font-size: 22px;
		font-weight: 600;
		color: #b0441d;
		letter-spacing: -0.02em;
	}
	.hito b {
		font-size: 16.5px;
	}
	.hi-d {
		font-size: 14.5px;
		color: #5c4536;
		margin-top: 3px;
		line-height: 1.5;
	}
	.valores {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 240px), 1fr));
		gap: 12px;
	}
	.valor {
		padding: 22px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.v-t {
		margin-top: 12px;
		font-size: 17px;
		font-weight: 600;
	}
	.valor p {
		margin: 6px 0 0;
		font-size: 14.5px;
		line-height: 1.55;
		color: #5c4536;
	}
	.trabaja {
		padding: clamp(22px, 4vw, 40px);
		border-radius: 32px;
		background: #2a140c;
		color: #fbf7f2;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr));
		gap: 24px;
		align-items: start;
	}
	.t-t {
		margin-top: 12px;
		font-size: clamp(28px, 3.4vw, 42px);
		font-weight: 600;
		letter-spacing: -0.035em;
		line-height: 1.05;
		text-wrap: balance;
	}
	.trabaja p {
		margin: 12px 0 0;
		font-size: 15px;
		color: #d9c8ba;
		line-height: 1.55;
	}
	.busquedas {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.busquedas a {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 16px;
		border-radius: 18px;
		background: rgba(251, 247, 242, 0.06);
		border: 1px solid rgba(251, 247, 242, 0.1);
		color: #fbf7f2;
	}
	.b-txt {
		flex: 1;
		min-width: 0;
	}
	.b-txt b {
		display: block;
		font-size: 15px;
	}
	.b-txt span {
		font-size: 13px;
		color: #d9c8ba;
	}
	/* Celular: el mosaico 2:1 queda en tiras finitas; la foto grande arriba y las otras dos al lado. */
	@media (max-width: 640px) {
		.mosaico {
			grid-template-columns: 1fr;
		}
		.costado {
			grid-template-rows: none;
			grid-template-columns: 1fr 1fr;
			min-height: 150px;
		}
		.mosaico > div,
		.chica,
		.cifra {
			border-radius: 22px;
		}
		.hito {
			grid-template-columns: 58px 1fr;
			gap: 12px;
		}
	}
</style>
