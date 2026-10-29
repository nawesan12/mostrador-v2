<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import Etiqueta from '#lib/sitio/Etiqueta.svelte';
	import Titulo from '#lib/sitio/Titulo.svelte';

	/** Sobre nosotros (design/Mostrador Nosotros.dc.html). Personas y cifras de ejemplo del diseño. */
	const foto = (id: string, extra: string) =>
		`https://images.unsplash.com/photo-${id}?auto=format&fit=crop&q=72&${extra}`;
	const HITOS = [
		[
			'2023',
			'Arrancamos en un garage de Rosario',
			'Tres amigos que vendían yerba online cansados de pagar comisiones.'
		],
		[
			'2024',
			'Las primeras 1.000 tiendas',
			'Sumamos WhatsApp y facturación electrónica, lo que más nos pedían.'
		],
		[
			'2025',
			'Llega Hornero',
			'El asistente de IA que responde, vende y avisa cuando algo se traba.'
		],
		[
			'2026',
			'Más de 40.000 tiendas',
			'Con oficinas en Buenos Aires, Córdoba y Rosario, y un equipo de 80 personas.'
		]
	] as const;
	const VALORES = [
		[
			I.phCurrencyCircleDollarFill,
			'Precio justo',
			'Sin comisiones escondidas ni apps que se pagan aparte.'
		],
		[I.phShieldCheckFill, 'Seguridad de serie', 'Lo que protege tu negocio no es un extra.'],
		[I.phHandHeartFill, 'Gente que atiende', 'Respondemos personas, por WhatsApp, en minutos.'],
		[
			I.phMapTrifoldFill,
			'Hecho para Argentina',
			'Cuotas, ARCA, Correo Argentino y el CP de cada pueblo.'
		]
	] as const;
	const EQUIPO = [
		['1531058240690-006c446962d8', 'Martina Ruiz', 'Cofundadora · Producto'],
		['1520333789090-1afc82db536a', 'Julián Pereyra', 'Cofundador · Tecnología'],
		['1687422811062-a966b55cb217', 'Camila Ríos', 'Atención a comercios'],
		['1545262722-9e0d80a0bc01', 'Federico Luna', 'Pagos y seguridad']
	] as const;
	const BUSQUEDAS = [
		['Ingeniero/a backend sénior', 'Tecnología · Remoto · Argentina'],
		['Diseñador/a de producto', 'Producto · Híbrido · CABA'],
		['Especialista en migraciones', 'Atención · Híbrido · Córdoba'],
		['Ejecutivo/a de cuentas', 'Ventas · Híbrido · Rosario']
	] as const;
</script>

<Seo
	title="Sobre nosotros · Tu Local Web"
	description="Hacemos que vender sea más barato y más simple para el que empieza. Nuestra historia, lo que nos importa y el equipo."
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
				<div class="c-v">40.000+</div>
				<div class="c-t">tiendas que venden con Tu Local Web</div>
			</div>
		</div>
	</div>
</div>

<section class="bloque">
	<div class="ancho">
		<div class="historia">
			<div>
				<Titulo etiqueta="Nuestra historia" titulo="Nació de un problema propio" />
				<p class="h-txt">
					Vendíamos yerba por internet y cada mes pagábamos más en comisiones, apps y planes que en
					envíos. Armamos Tu Local Web para que cualquier emprendedor tenga todo lo que necesita, en un
					solo lugar y por un precio justo.
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
		<Titulo etiqueta="El equipo" titulo="Algunas de las caras de Tu Local Web" />
		<div class="equipo">
			{#each EQUIPO as [id, n, r] (n)}
				<div>
					<div class="cara">
						<img src={foto(id, 'w=400&h=480&crop=faces')} alt="" loading="lazy" />
					</div>
					<div class="e-n">{n}</div>
					<div class="e-r">{r}</div>
				</div>
			{/each}
		</div>
		<div class="trabaja">
			<div>
				<Etiqueta texto="Trabajá con nosotros" color="#F0A58A" />
				<div class="t-t">Buscamos gente que quiera que a los emprendedores les vaya bien.</div>
				<p>
					Sueldo en pesos ajustado por inflación, prepaga, 4 semanas de vacaciones y equipo para
					trabajar desde tu casa.
				</p>
			</div>
			<div class="busquedas">
				{#each BUSQUEDAS as [t, d] (t)}
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- página pública -->
					<a href="/contacto">
						<span class="b-txt"><b>{t}</b><span>{d}</span></span>
						<Icon icon={I.phArrowUpRightBold} size={15} />
					</a>
				{/each}
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
	.equipo {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(46%, 220px), 1fr));
		gap: 12px;
	}
	.cara {
		aspect-ratio: 5 / 6;
		border-radius: 24px;
		overflow: hidden;
		background: #e9ddcf;
	}
	.cara img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.e-n {
		margin-top: 10px;
		font-size: 16px;
		font-weight: 600;
	}
	.e-r {
		font-size: 13.5px;
		color: #7a6252;
	}
	.trabaja {
		margin-top: clamp(48px, 7vw, 80px);
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
