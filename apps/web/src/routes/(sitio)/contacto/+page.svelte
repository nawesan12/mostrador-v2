<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import Campo from '#lib/sitio/Campo.svelte';
	import Chips from '#lib/sitio/Chips.svelte';
	import Etiqueta from '#lib/sitio/Etiqueta.svelte';
	import { FACTURACION, ORIGEN } from '#lib/sitio/opciones.ts';
	import type { PageProps } from './$types';

	/** Contacto y ventas (design/Mostrador Contacto.dc.html). */
	let { form }: PageProps = $props();
	const err = $derived((form && 'errores' in form ? form.errores : {}) as Record<string, string>);
	const previo = $derived((form && 'errores' in form ? form : {}) as Record<string, string>);
	let otra = $state(false);
	const enviado = $derived(!!form && 'enviado' in form && !otra);
	$effect(() => {
		if (form) otra = false;
	});
	let enviando = $state(false);
	const conEnvio = () => {
		enviando = true;
		return async ({ update }: { update: (o?: { reset?: boolean }) => Promise<void> }) => {
			await update({ reset: false });
			enviando = false;
		};
	};

	const CANALES = [
		[I.phEnvelopeSimpleFill, 'Email', 'ayuda@mostradorapp.com.ar', 'Respuesta en menos de 24 h'],
		[I.phChatsCircleFill, 'Chat del panel', 'Hablá con Hornero desde tu panel', 'Disponible 24/7']
	] as const;
	const CARAS = [
		'1531058240690-006c446962d8',
		'1687422811062-a966b55cb217',
		'1520333789090-1afc82db536a'
	];
</script>

<Seo
	title="Contacto y ventas · Mostrador"
	description="Hablemos de tu negocio: plan a medida y migración gratis si facturás más de $ 20 millones por mes, tenés sucursales o venís de otra plataforma."
	path="/contacto"
/>

<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas públicas fijas -->
<div class="ancho">
	<div class="in">
		<div>
			<Etiqueta texto="Contacto y ventas" />
			<h1>Hablemos de tu negocio.</h1>
			<p class="bajada">
				Si facturás más de $ 20 millones por mes, tenés varias sucursales o venís de otra
				plataforma, te armamos un plan a medida y te migramos gratis.
			</p>
			<div class="canales">
				{#each CANALES as [ico, t, v, d] (t)}
					<div class="canal">
						<Icon icon={ico} size={24} style="color:#b0441d" />
						<div class="ca-t">{t}</div>
						<div class="ca-v">{v}</div>
						<div class="ca-d">{d}</div>
					</div>
				{/each}
			</div>
			<div class="equipo">
				<div class="caras">
					{#each CARAS as id (id)}
						<img
							src="https://images.unsplash.com/photo-{id}?auto=format&fit=crop&w=96&h=96&q=60&crop=faces"
							alt=""
							width="40"
							height="40"
						/>
					{/each}
				</div>
				<div class="eq-t">
					<b>Equipo de ventas en Buenos Aires, Córdoba y Rosario.</b> Te atiende una persona, no un call
					center.
				</div>
			</div>
		</div>

		<div class="tarjeta">
			{#if enviado}
				<div class="exito" role="status">
					<img src="/hornero/celebrando.webp" alt="" width="140" height="140" />
					<div class="ex-t">¡Listo, {form && 'nombre' in form ? form.nombre : ''}!</div>
					<p>Te respondemos en menos de 24 h. Mientras tanto, podés crear tu tienda&nbsp;gratis.</p>
					<div class="ex-acc">
						<a href="/onboarding">Crear mi tienda</a>
						<button type="button" onclick={() => (otra = true)}>Enviar otra consulta</button>
					</div>
				</div>
			{:else}
				<form method="POST" use:enhance={conEnvio} novalidate>
					<div class="f-t">Contanos de tu negocio</div>
					<div class="f-d">Te escribimos en menos de 2 horas hábiles.</div>
					{#if err.form}<div class="err-form" role="alert">{err.form}</div>{/if}
					<div class="campos">
						<Campo
							nombre="nombre"
							etiqueta="Nombre"
							ejemplo="Lucía Fernández"
							autocompletar="name"
							valor={previo.nombre}
							error={err.nombre}
						/>
						<Campo
							nombre="email"
							etiqueta="Email"
							tipo="email"
							ejemplo="lucia@tutienda.com.ar"
							autocompletar="email"
							valor={previo.email}
							error={err.email}
						/>
						<Campo
							nombre="whatsapp"
							etiqueta="WhatsApp"
							tipo="tel"
							ejemplo="11 5555-0000"
							autocompletar="tel"
							modo="tel"
							valor={previo.whatsapp}
							error={err.whatsapp}
						/>
						<Campo
							nombre="web"
							etiqueta="Tu tienda o web actual"
							ejemplo="tutienda.com.ar"
							valor={previo.web}
							error={err.web}
						/>
					</div>
					<Chips
						nombre="facturacion"
						pregunta="¿Cuánto facturás por mes?"
						opciones={FACTURACION}
						elegidas={[previo.facturacion ?? '']}
						error={err.facturacion}
					/>
					<Chips
						nombre="origen"
						pregunta="¿Desde dónde venís?"
						opciones={ORIGEN}
						elegidas={[previo.origen ?? '']}
						error={err.origen}
					/>
					<div class="msj">
						<Campo
							nombre="mensaje"
							etiqueta="¿En qué te ayudamos?"
							filas={3}
							ejemplo="Tengo 2 locales y quiero unificar stock con la tienda online…"
							valor={previo.mensaje}
							error={err.mensaje}
						/>
					</div>
					<button class="enviar" type="submit" disabled={enviando}
						>{enviando ? 'Enviando…' : 'Quiero que me contacten'}</button
					>
					<div class="pie">
						Al enviar aceptás la <a href="/legales/privacidad">política de privacidad</a>.
					</div>
				</form>
			{/if}
		</div>
	</div>
</div>

<style>
	.ancho {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 clamp(16px, 3vw, 28px);
	}
	.in {
		padding: clamp(40px, 6vw, 80px) 0 clamp(56px, 8vw, 100px);
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 400px), 1fr));
		gap: clamp(24px, 4vw, 56px);
		align-items: start;
	}
	h1 {
		margin: 14px 0 0;
		font-size: clamp(42px, 6vw, 80px);
		line-height: 0.96;
		letter-spacing: -0.05em;
		font-weight: 600;
	}
	.bajada {
		margin: 16px 0 0;
		font-size: 17.5px;
		line-height: 1.55;
		color: #5c4536;
		max-width: 480px;
	}
	.canales {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 10px;
	}
	.canal {
		padding: 18px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.ca-t {
		margin-top: 10px;
		font-size: 13px;
		color: #7a6252;
	}
	.ca-v {
		font-size: 15.5px;
		font-weight: 600;
		margin-top: 2px;
		overflow-wrap: anywhere;
	}
	.ca-d {
		font-size: 13px;
		color: #7a6252;
		margin-top: 4px;
	}
	.equipo {
		margin-top: 14px;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 16px 18px;
		border-radius: 20px;
		background: #f4ece2;
	}
	.caras {
		display: flex;
		flex: none;
	}
	.caras img {
		border-radius: 50%;
		object-fit: cover;
		box-shadow: 0 0 0 3px #f4ece2;
	}
	.caras img + img {
		margin-left: -10px;
	}
	.eq-t {
		font-size: 14px;
		line-height: 1.45;
	}
	.tarjeta {
		padding: clamp(20px, 3vw, 30px);
		border-radius: 28px;
		background: #fff;
		box-shadow:
			0 2px 4px rgba(58, 29, 18, 0.05),
			0 30px 60px -30px rgba(58, 29, 18, 0.45);
	}
	.f-t {
		font-size: 20px;
		font-weight: 600;
	}
	.f-d {
		font-size: 14px;
		color: #7a6252;
		margin-top: 4px;
	}
	.campos {
		margin-top: 18px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 10px;
	}
	.msj {
		margin-top: 12px;
	}
	.err-form {
		margin-top: 14px;
		padding: 12px 14px;
		border-radius: 14px;
		background: rgba(184, 58, 46, 0.07);
		border: 1px solid rgba(184, 58, 46, 0.28);
		font-size: 14px;
	}
	.enviar {
		margin-top: 16px;
		width: 100%;
		padding: 16px;
		border: 0;
		border-radius: 999px;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 15.5px;
		font-weight: 600;
		box-shadow: 0 12px 26px -12px rgba(210, 88, 47, 0.8);
	}
	.enviar:disabled {
		opacity: 0.7;
	}
	.pie {
		margin-top: 8px;
		font-size: 12.5px;
		color: #7a6252;
		text-align: center;
	}
	.pie a {
		color: #b0441d;
		text-decoration: underline;
	}
	.exito {
		text-align: center;
		padding: 24px 8px;
	}
	.exito img {
		object-fit: contain;
	}
	.ex-t {
		margin-top: 8px;
		font-size: 22px;
		font-weight: 600;
	}
	.exito p {
		margin: 8px auto 0;
		max-width: 360px;
		font-size: 15px;
		line-height: 1.55;
		color: #5c4536;
	}
	.ex-acc {
		margin-top: 16px;
		display: flex;
		gap: 8px;
		justify-content: center;
		flex-wrap: wrap;
	}
	.ex-acc a {
		padding: 12px 18px;
		border-radius: 999px;
		background: #2a140c;
		color: #fff;
		font-size: 14px;
		font-weight: 600;
	}
	.ex-acc button {
		padding: 12px 18px;
		border-radius: 999px;
		border: 1.5px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		font-size: 14px;
		font-weight: 600;
		color: #2a140c;
	}
</style>
