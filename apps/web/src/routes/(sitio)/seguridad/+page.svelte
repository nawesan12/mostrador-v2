<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import Etiqueta from '#lib/sitio/Etiqueta.svelte';
	import Faq from '#lib/sitio/Faq.svelte';
	import Titulo from '#lib/sitio/Titulo.svelte';

	/**
	 * Seguridad (design/Mostrador Seguridad.dc.html). Sólo lo que la plataforma hace hoy: sin cifras
	 * de disponibilidad, certificaciones ni funciones que todavía no existen.
	 */
	const DATOS = [
		['Cada hora', 'chequeamos que todo ande'],
		['Todos los días', 'hacemos una copia'],
		['0', 'datos de tarjeta guardados']
	] as const;
	const CAPAS = [
		[
			I.phLockKeyFill,
			'Conexión cifrada',
			'HTTPS en cada página y certificado SSL gratis en tu dominio propio, que se renueva solo.'
		],
		[
			I.phPasswordFill,
			'Claves bien guardadas',
			'Las contraseñas se guardan con argon2id: ni nosotros podemos leerlas.'
		],
		[
			I.phStackFill,
			'Cada tienda, separada',
			'La base de datos aísla cada tienda: una cuenta nunca ve los datos de otra.'
		],
		[
			I.phClockCounterClockwiseFill,
			'Copias de seguridad diarias',
			'Copiamos la base todos los días y guardamos las últimas 14.'
		],
		[
			I.phCreditCardFill,
			'Las tarjetas, en Mercado Pago',
			'Los datos de tarjeta nunca pasan por Tu Local Web: los procesa Mercado Pago.'
		],
		[
			I.phPulseFill,
			'Monitoreo cada hora',
			'Chequeamos la tienda, el panel y los pagos, y lo publicamos en Estado del sistema.'
		]
	] as const;
	const CUMPLIMIENTO = [
		[
			I.phSealCheckFill,
			'#2F8F63',
			'Ley 25.326 de Datos Personales',
			'Cómo tratamos los datos, en la política de privacidad'
		],
		[I.phCreditCardFill, '#2F8F63', 'PCI DSS', 'A través de Mercado Pago, que procesa los pagos']
	] as const;
	const FAQ = [
		[
			'¿Qué pasa si se cae Tu Local Web?',
			'Lo ves en Estado del sistema, que se actualiza cada hora, con lo que pasó y cuándo se resolvió.'
		],
		[
			'¿Pueden ver los datos de mis clientes?',
			'Los datos son tuyos. Sólo los miramos si nos pedís ayuda con algo puntual, y nunca los vendemos ni los compartimos.'
		],
		[
			'¿Me puedo llevar todo si me voy?',
			'Sí. Exportás productos, clientes, pedidos y facturas en CSV o Excel desde Configuración › Mis datos, cuando quieras.'
		],
		[
			'¿Cómo pago con tarjeta sin riesgo?',
			'Tus clientes pagan en Mercado Pago, que guarda y procesa la tarjeta. Tu Local Web sólo recibe el aviso del pago.'
		]
	] as const;
</script>

<Seo
	title="Seguridad · Tu Local Web"
	description="Conexión cifrada, claves bien guardadas, cada tienda separada, copias diarias y monitoreo cada hora, en todos los planes."
	path="/seguridad"
	image="seguridad"
/>

<section class="hero">
	<div class="ancho">
		<div class="hero-in">
			<div class="hero-txt">
				<Etiqueta texto="Seguridad" color="#F0A58A" />
				<h1>Tu negocio, cuidado desde el primer día.</h1>
				<p>
					Lo básico para que tu tienda y tus datos estén seguros viene activado en todos los planes,
					sin pagar&nbsp;extra.
				</p>
				<div class="datos">
					{#each DATOS as [v, t] (t)}<div>
							<div class="d-v">{v}</div>
							<div class="d-t">{t}</div>
						</div>{/each}
				</div>
			</div>
			<img src="/hornero/seguridad.webp" alt="" width="320" height="320" />
		</div>
	</div>
</section>

<section class="bloque">
	<div class="ancho">
		<Titulo etiqueta="Qué incluye" titulo="Seis cosas, todas incluidas" />
		<div class="capas">
			{#each CAPAS as [ico, t, d] (t)}
				<div class="capa">
					<span class="c-ico"><Icon icon={ico} size={22} /></span>
					<div class="c-t">{t}</div>
					<p>{d}</p>
				</div>
			{/each}
		</div>
	</div>
</section>

<section class="bloque">
	<div class="ancho">
		<div class="dos">
			<div class="tarjeta">
				<Etiqueta texto="Cumplimiento" />
				<div class="cumple">
					{#each CUMPLIMIENTO as [ico, col, t, d] (t)}
						<div>
							<Icon icon={ico} size={26} style="color:{col};flex:none" />
							<div>
								<b>{t}</b>
								<div class="cu-d">{d}</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
			<div class="tarjeta oscura">
				<Etiqueta texto="Reportá una vulnerabilidad" color="#F0A58A" />
				<p>
					Si encontrás un problema de seguridad, escribinos a <b>soporte@tulocalweb.com.ar</b> con los
					detalles. Lo lee una persona del equipo y te respondemos&nbsp;personalmente.
				</p>
				<a href="mailto:soporte@tulocalweb.com.ar"
					>Escribir a soporte<Icon icon={I.phArrowRightBold} size={13} /></a
				>
			</div>
		</div>
		<div class="preguntas">
			<Titulo etiqueta="Preguntas" titulo="Lo que más nos preguntan" />
			<div class="faq"><Faq items={FAQ} /></div>
		</div>
	</div>
</section>

<style>
	.ancho {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 clamp(16px, 3vw, 28px);
	}
	.hero {
		background: #2a140c;
		color: #fbf7f2;
	}
	.hero-in {
		padding: clamp(48px, 8vw, 104px) 0;
		display: flex;
		align-items: center;
		gap: 32px;
		flex-wrap: wrap;
	}
	.hero-txt {
		flex: 1 1 520px;
		min-width: 0;
	}
	h1 {
		margin: 14px 0 0;
		font-size: clamp(44px, 6.6vw, 88px);
		line-height: 0.95;
		letter-spacing: -0.05em;
		font-weight: 600;
	}
	.hero-txt > p {
		margin: 18px 0 0;
		font-size: 18px;
		line-height: 1.55;
		color: #d9c8ba;
		max-width: 540px;
	}
	.datos {
		margin-top: 24px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.datos > div {
		padding: 14px 16px;
		border-radius: 18px;
		background: rgba(251, 247, 242, 0.07);
		border: 1px solid rgba(251, 247, 242, 0.1);
	}
	.d-v {
		font-size: 26px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.d-t {
		font-size: 13px;
		color: #d9c8ba;
	}
	.hero-in > img {
		width: clamp(180px, 26vw, 320px);
		height: auto;
		flex: none;
		margin: 0 auto;
	}
	.bloque {
		padding: clamp(56px, 8vw, 104px) 0;
	}
	.beige {
		background: #f4ece2;
	}
	.capas {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr));
		gap: 12px;
	}
	.capa {
		padding: 22px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.c-ico {
		width: 44px;
		height: 44px;
		border-radius: 14px;
		background: #fcede5;
		display: grid;
		place-items: center;
		color: #b0441d;
	}
	.c-t {
		margin-top: 14px;
		font-size: 17px;
		font-weight: 600;
	}
	.capa p {
		margin: 6px 0 0;
		font-size: 14.5px;
		line-height: 1.55;
		color: #5c4536;
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr));
		gap: 12px;
	}
	.tarjeta {
		padding: 26px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.cumple {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.cumple > div {
		display: flex;
		gap: 12px;
		align-items: center;
	}
	.cumple b {
		font-size: 15px;
	}
	.cu-d {
		font-size: 13.5px;
		color: #7a6252;
	}
	.oscura {
		background: #2a140c;
		color: #fbf7f2;
		box-shadow: none;
	}
	.oscura p {
		margin: 14px 0 0;
		font-size: 15.5px;
		line-height: 1.6;
		color: #d9c8ba;
	}
	.oscura b {
		color: #fff;
	}
	.oscura a {
		margin-top: 16px;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 12px 18px;
		border-radius: 999px;
		background: #fff;
		color: #2a140c;
		font-size: 14px;
		font-weight: 600;
	}
	.preguntas {
		margin-top: 40px;
	}
	.faq {
		margin-top: 22px;
	}
	/* Celular: la tabla de tres columnas se apila (la función arriba, las dos respuestas abajo). */
	@media (max-width: 640px) {
		.hero-in > img {
			width: 160px;
		}
		.datos {
			display: grid;
			grid-template-columns: repeat(3, minmax(0, 1fr));
			gap: 6px;
		}
		.datos > div {
			padding: 12px 10px;
		}
		.d-v {
			font-size: 21px;
		}
		.d-t {
			font-size: 12px;
			line-height: 1.3;
		}
	}
</style>
