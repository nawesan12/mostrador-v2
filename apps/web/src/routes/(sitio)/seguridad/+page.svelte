<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import Etiqueta from '#lib/sitio/Etiqueta.svelte';
	import Faq from '#lib/sitio/Faq.svelte';
	import Titulo from '#lib/sitio/Titulo.svelte';

	/** Seguridad (design/Mostrador Seguridad.dc.html). */
	const DATOS = [
		['99,98 %', 'disponibilidad 12 meses'],
		['30 días', 'de copias restaurables'],
		['0', 'datos de tarjeta guardados']
	] as const;
	const CAPAS = [
		[
			I.phLockKeyFill,
			'Cifrado de punta a punta',
			'TLS 1.3 en cada conexión y AES-256 para los datos guardados. Certificado SSL gratis en tu dominio.'
		],
		[
			I.phPasswordFill,
			'Verificación en dos pasos',
			'Con app de autenticación o WhatsApp, incluida en todos los planes. Obligatoria para dueños de tienda.'
		],
		[
			I.phClockCounterClockwiseFill,
			'Copias de seguridad diarias',
			'Guardamos 30 días de historia. Restaurás un producto borrado o toda la tienda vos mismo, en un clic.'
		],
		[
			I.phShieldWarningFill,
			'Antifraude en cada venta',
			'Detectamos compras sospechosas antes de despachar y te avisamos con el motivo.'
		],
		[
			I.phCreditCardFill,
			'Pagos certificados PCI',
			'Los datos de tarjeta nunca pasan por Tu Local Web: los procesan Mercado Pago y Payway.'
		],
		[
			I.phPulseFill,
			'Monitoreo 24/7',
			'Un equipo de guardia y alertas automáticas. El 99,98 % de disponibilidad en los últimos 12 meses.'
		]
	] as const;
	const COMPARADO = [
		['Verificación en dos pasos', 'Incluida', 'Depende del plan'],
		['Restaurar copias vos mismo', 'Sí, 30 días', 'Por soporte'],
		['Alertas de inicio de sesión', 'Email y WhatsApp', 'Solo email'],
		['Roles y permisos por persona', 'Ilimitados', 'Limitados'],
		['Registro de actividad', '12 meses', '30 días o nada'],
		['Antifraude en pedidos', 'Incluido', 'Pago aparte']
	] as const;
	const CUMPLIMIENTO = [
		[I.phSealCheckFill, '#2F8F63', 'Ley 25.326 de Datos Personales', 'Base inscripta ante la AAIP'],
		[I.phCreditCardFill, '#2F8F63', 'PCI DSS', 'A través de nuestros procesadores de pago'],
		[I.phCertificateFill, '#9A6A10', 'ISO 27001', 'Certificación en proceso · previsto para 2027']
	] as const;
	const FAQ = [
		[
			'¿Qué pasa si se cae Tu Local Web en pleno Hot Sale?',
			'Tenemos servidores en dos zonas distintas y escalamos solos con el tráfico. Si aun así hay una caída, lo ves en Estado del sistema y te acreditamos parte del plan según la política de Reembolsos.'
		],
		[
			'¿Pueden ver los datos de mis clientes?',
			'Solo el personal de soporte que vos autorices desde el panel, por un tiempo limitado y con registro de cada acceso.'
		],
		[
			'¿Me puedo llevar todo si me voy?',
			'Sí. Exportás productos, clientes, pedidos y facturas en CSV o Excel desde Configuración › Mis datos, cuando quieras.'
		],
		[
			'¿Cómo me protegen de compras fraudulentas?',
			'Analizamos cada pedido con más de 40 señales (IP, dirección, historial, monto) y te marcamos los riesgosos antes de despachar.'
		]
	] as const;
</script>

<Seo
	title="Seguridad · Tu Local Web"
	description="Cifrado, verificación en dos pasos, copias diarias y antifraude incluidos en todos los planes."
	path="/seguridad"
	image="seguridad"
/>

<section class="hero">
	<div class="ancho">
		<div class="hero-in">
			<div class="hero-txt">
				<Etiqueta texto="Seguridad" color="#F0A58A" />
				<h1>Tu negocio no se cae. Y nada se pierde.</h1>
				<p>
					Lo que en otras plataformas es opcional o cuesta extra, en Tu Local Web viene activado
					desde el primer día.
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
		<Titulo etiqueta="Qué incluye" titulo="Seis capas, todas incluidas" />
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

<section class="bloque beige">
	<div class="ancho">
		<Titulo etiqueta="Comparado" titulo="Lo que otras cobran aparte" />
		<div class="tabla" role="table" aria-label="Seguridad: Tu Local Web contra otras plataformas">
			<div class="fila cab" role="row">
				<span role="columnheader">Función</span>
				<span role="columnheader" class="t-m"
					><img src="/mark.svg" alt="" width="16" />Tu Local Web</span
				>
				<span role="columnheader" class="t-o">Otras plataformas</span>
			</div>
			{#each COMPARADO as [f, m, o] (f)}
				<div class="fila" role="row">
					<span role="cell" class="t-f">{f}</span>
					<span role="cell" class="t-si"><Icon icon={I.phCheckCircleFill} size={16} />{m}</span>
					<span role="cell" class="t-no">{o}</span>
				</div>
			{/each}
		</div>
		<div class="nota">
			Comparación orientativa con planes básicos de plataformas de e-commerce en Argentina, octubre
			de 2026.
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
					Si encontrás un problema de seguridad, escribinos a <b>seguridad@tulocalweb.com.ar</b>.
					Respondemos en 24 horas, no tomamos acciones legales contra investigaciones de buena fe y
					reconocemos los hallazgos válidos con recompensas de hasta USD 2.000.
				</p>
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- página pública -->
				<a href="/contacto">Ver el programa<Icon icon={I.phArrowRightBold} size={13} /></a>
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
	.tabla {
		margin-top: 28px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		overflow: hidden;
	}
	.fila {
		display: grid;
		grid-template-columns: minmax(0, 1.6fr) minmax(0, 1fr) minmax(0, 1fr);
		gap: 12px;
		padding: 14px 20px;
		border-top: 1px solid rgba(58, 29, 18, 0.07);
		font-size: 14.5px;
		align-items: center;
	}
	.fila.cab {
		background: #2a140c;
		color: #fff;
		font-size: 13px;
		font-weight: 600;
		border-top: 0;
	}
	.t-m {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.t-o {
		opacity: 0.7;
	}
	.t-f {
		font-weight: 500;
	}
	.t-si {
		display: flex;
		align-items: center;
		gap: 6px;
		font-weight: 600;
		color: #237550;
	}
	.t-no {
		color: #7a6252;
	}
	.nota {
		margin-top: 8px;
		font-size: 12.5px;
		color: #7a6252;
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
		.fila {
			grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
			gap: 4px 12px;
			padding: 13px 16px;
		}
		.fila .t-f {
			grid-column: 1 / -1;
		}
		.fila.cab > :first-child {
			display: none;
		}
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
