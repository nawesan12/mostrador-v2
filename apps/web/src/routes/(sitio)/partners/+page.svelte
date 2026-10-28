<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import Campo from '#lib/sitio/Campo.svelte';
	import Chips from '#lib/sitio/Chips.svelte';
	import Etiqueta from '#lib/sitio/Etiqueta.svelte';
	import Titulo from '#lib/sitio/Titulo.svelte';
	import { ESPECIALIDADES } from '#lib/sitio/opciones.ts';
	import type { PageProps } from './$types';

	/** Programa de partners (design/Mostrador Partners.dc.html). */
	let { data, form }: PageProps = $props();
	const err = $derived((form && 'errores' in form ? form.errores : {}) as Record<string, string>);
	const previo = $derived(
		(form && 'errores' in form ? form : {}) as Record<string, string> & {
			especialidades?: string[];
		}
	);
	const enviado = $derived(!!form && 'enviado' in form);
	let enviando = $state(false);
	const conEnvio = () => {
		enviando = true;
		return async ({ update }: { update: (o?: { reset?: boolean }) => Promise<void> }) => {
			await update({ reset: false });
			enviando = false;
		};
	};

	const NIVELES = [
		{
			n: 'Partner',
			d: 'Empezá hoy',
			f: [
				'20 % de comisión recurrente por 12 meses',
				'Cuenta de prueba ilimitada',
				'Materiales de venta'
			],
			req: 'Sin requisitos',
			tono: 'claro'
		},
		{
			n: 'Partner Plata',
			d: 'Desde 5 tiendas activas',
			f: [
				'25 % de comisión recurrente',
				'Aparecés en el directorio',
				'Soporte prioritario por WhatsApp',
				'Acceso anticipado a novedades'
			],
			req: '5 tiendas',
			tono: 'oscuro'
		},
		{
			n: 'Partner Oro',
			d: 'Desde 20 tiendas activas',
			f: [
				'30 % de comisión recurrente',
				'Destacado en el directorio y la landing',
				'Gerente de cuenta dedicado',
				'Co-marketing y leads derivados'
			],
			req: '20 tiendas',
			tono: 'terracota'
		}
	] as const;
	const DIRECTORIO = [
		['1522202176988-66273c2fd55f', 'Estudio Brote', 'Partner Oro', 'Diseño y branding · CABA', 38],
		['1552664730-d307ca884978', 'Nodo Digital', 'Partner Oro', 'Migraciones y SEO · Córdoba', 38],
		[
			'1517245386807-bb43f82c33c4',
			'Pampa Code',
			'Partner Plata',
			'Desarrollo e integraciones · Rosario',
			14
		]
	] as const;
</script>

<Seo
	title="Programa de partners · Mostrador"
	description="Para agencias, diseñadores, desarrolladores y contadores: recomendás Mostrador y cobrás hasta 30 % del plan cada mes."
	path="/partners"
	image="partners"
/>

<section class="hero">
	<div class="ancho">
		<div class="hero-in">
			<div class="hero-txt">
				<Etiqueta texto="Programa de partners" color="#F0A58A" />
				<h1>Armás tiendas. Ganás cada mes.</h1>
				<p>
					Para agencias, diseñadores, desarrolladores y contadores. Recomendás Mostrador a tus
					clientes y cobrás una parte del plan mientras sigan activos.
				</p>
				<div class="ctas">
					<a class="pri" href="#sumate">Quiero ser partner</a>
					<a class="sec" href="#directorio">Buscar un partner</a>
				</div>
			</div>
			<img src="/hornero/celebrando.png" alt="" width="300" height="300" />
		</div>
	</div>
</section>

<section class="bloque">
	<div class="ancho">
		<Titulo etiqueta="Niveles" titulo="Cuanto más tiendas, más ganás" />
		<div class="niveles">
			{#each NIVELES as nv (nv.n)}
				<div class="nivel {nv.tono}">
					<div>
						<div class="nv-n">{nv.n}</div>
						<div class="nv-d">{nv.d}</div>
					</div>
					<div class="nv-f">
						{#each nv.f as f (f)}<div>
								<Icon icon={I.phCheckBold} size={16} style="flex:none;margin-top:2px" />{f}
							</div>{/each}
					</div>
					<div class="nv-req">Requisito: {nv.req}</div>
				</div>
			{/each}
		</div>
		<div class="ejemplo">
			<Icon icon={I.phCalculatorFill} size={30} style="color:#C24E25;flex:none" />
			<div class="ej-txt">
				<b>Ejemplo: 12 tiendas en plan Pro</b>
				<div>Nivel Plata, 25&nbsp;% de comisión sobre $&nbsp;49.900 por mes</div>
			</div>
			<div class="ej-v">
				<div class="v">$&nbsp;149.700</div>
				<div class="d">por mes, mientras sigan activas</div>
			</div>
		</div>
	</div>
</section>

<section id="directorio" class="bloque beige">
	<div class="ancho">
		<Titulo etiqueta="Directorio" titulo="¿Necesitás ayuda? Contratá un partner" />
		<div class="directorio">
			{#each DIRECTORIO as [img, n, nivel, d, tiendas] (n)}
				<div class="partner">
					<div class="foto">
						<img
							src="https://images.unsplash.com/photo-{img}?auto=format&fit=crop&w=700&h=394&q=72"
							alt=""
							loading="lazy"
						/>
					</div>
					<div class="p-in">
						<div class="p-cab">
							<b>{n}</b><span class="nivel-chip" class:oro={nivel === 'Partner Oro'}>{nivel}</span>
						</div>
						<div class="p-d">{d}</div>
						<div class="p-val">
							<Icon icon={I.phStarFill} size={14} style="color:#D2582F" /><b>4,9</b><span
								>· {tiendas} tiendas armadas</span
							>
						</div>
					</div>
				</div>
			{/each}
		</div>
	</div>
</section>

<section id="sumate" class="bloque">
	<div class="ancho">
		<div class="sumate">
			<div>
				<Titulo etiqueta="Sumate" titulo="Tu primera tienda, esta semana" />
				<p class="s-txt">
					Te damos acceso al panel de partners, una cuenta de prueba sin límite y una capacitación
					de una hora con el equipo.
				</p>
				{#if data.demo}
					<form id="ejemplo" class="ver-ejemplo" method="POST" action="/partners/panel/ejemplo">
						<button class="btn btn-secondary">Ver el panel de partners de ejemplo</button>
					</form>
				{/if}
			</div>
			<div class="tarjeta">
				{#if enviado}
					<div class="exito" role="status">
						<img src="/hornero/exito.png" alt="" width="120" height="120" />
						<div class="ex-t">¡Bienvenido al programa!</div>
						<p>
							Te mandamos el acceso al panel de partners por email. Revisá también la carpeta de
							spam.
						</p>
					</div>
				{:else}
					<form method="POST" use:enhance={conEnvio} novalidate>
						{#if err.form}<div class="err-form" role="alert">{err.form}</div>{/if}
						<div class="campos">
							<Campo
								nombre="nombre"
								etiqueta="Nombre"
								ejemplo="Juan Gómez"
								autocompletar="name"
								valor={previo.nombre}
								error={err.nombre}
							/>
							<Campo
								nombre="email"
								etiqueta="Email"
								tipo="email"
								ejemplo="juan@agencia.com"
								autocompletar="email"
								valor={previo.email}
								error={err.email}
							/>
							<Campo
								nombre="agencia"
								etiqueta="Agencia o estudio"
								ejemplo="Estudio Brote"
								autocompletar="organization"
								valor={previo.agencia}
								error={err.agencia}
							/>
							<Campo
								nombre="web"
								etiqueta="Web o portfolio"
								ejemplo="estudiobrote.com.ar"
								valor={previo.web}
								error={err.web}
							/>
						</div>
						<Chips
							nombre="especialidades"
							pregunta="¿Qué hacés?"
							opciones={ESPECIALIDADES}
							varias
							elegidas={previo.especialidades ?? []}
							error={err.especialidades}
						/>
						<button class="enviar" type="submit" disabled={enviando}
							>{enviando ? 'Enviando…' : 'Postularme'}</button
						>
					</form>
				{/if}
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
	.hero {
		background: #2a140c url('/textura/arch-pattern-dark.svg');
		color: #fbf7f2;
	}
	.hero-in {
		padding: clamp(48px, 8vw, 104px) 0;
		display: flex;
		align-items: center;
		gap: 28px;
		flex-wrap: wrap;
	}
	.hero-txt {
		flex: 1 1 520px;
		min-width: 0;
	}
	h1 {
		margin: 14px 0 0;
		font-size: clamp(44px, 6.6vw, 92px);
		line-height: 0.94;
		letter-spacing: -0.05em;
		font-weight: 600;
	}
	.hero-txt > p {
		margin: 16px 0 0;
		font-size: 18px;
		line-height: 1.55;
		color: #d9c8ba;
		max-width: 520px;
	}
	.ctas {
		margin-top: 24px;
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}
	.ctas a {
		padding: 15px 22px;
		border-radius: 999px;
		font-size: 15px;
		font-weight: 600;
		color: #fff;
	}
	.pri {
		background: #d2582f;
	}
	.sec {
		border: 1.5px solid rgba(251, 247, 242, 0.3);
		color: #fbf7f2 !important;
	}
	.ver-ejemplo {
		margin-top: 18px;
		scroll-margin-top: 90px;
	}
	.hero-in > img {
		width: clamp(170px, 24vw, 300px);
		height: auto;
		flex: none;
		margin: 0 auto;
	}
	.bloque {
		padding: clamp(56px, 8vw, 104px) 0;
		scroll-margin-top: 70px;
	}
	.beige {
		background: #f4ece2;
	}
	.niveles {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 280px), 1fr));
		gap: 12px;
	}
	.nivel {
		padding: 24px;
		border-radius: 28px;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		display: flex;
		flex-direction: column;
		gap: 14px;
		--linea: rgba(255, 255, 255, 0.2);
	}
	.claro {
		background: #fff;
		color: #2a140c;
		--linea: rgba(58, 29, 18, 0.1);
	}
	.oscuro {
		background: #2a140c;
		color: #fbf7f2;
	}
	.terracota {
		background: #d2582f;
		color: #fff;
	}
	.nv-n {
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.nv-d {
		font-size: 14px;
		opacity: 0.75;
		margin-top: 2px;
	}
	.nv-f {
		display: flex;
		flex-direction: column;
		gap: 9px;
	}
	.nv-f div {
		display: flex;
		gap: 9px;
		font-size: 14.5px;
		line-height: 1.45;
	}
	.nv-req {
		margin-top: auto;
		padding-top: 12px;
		border-top: 1px solid var(--linea);
		font-size: 13px;
		opacity: 0.8;
	}
	.ejemplo {
		margin-top: 14px;
		padding: 20px 22px;
		border-radius: 24px;
		background: #f4ece2;
		display: flex;
		align-items: center;
		gap: 18px;
		flex-wrap: wrap;
	}
	.ej-txt {
		flex: 1;
		min-width: 240px;
	}
	.ej-txt b {
		font-size: 16px;
	}
	.ej-txt div {
		font-size: 14px;
		color: #5c4536;
	}
	.ej-v {
		text-align: right;
	}
	.ej-v .v {
		font-size: 30px;
		font-weight: 600;
		letter-spacing: -0.03em;
		color: #c24e25;
	}
	.ej-v .d {
		font-size: 12.5px;
		color: #7a6252;
	}
	.directorio {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr));
		gap: 12px;
	}
	.partner {
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		overflow: hidden;
	}
	.foto {
		aspect-ratio: 16 / 9;
		background: #e9ddcf;
	}
	.foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.p-in {
		padding: 18px;
	}
	.p-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 10px;
	}
	.p-cab b {
		font-size: 17px;
	}
	.nivel-chip {
		padding: 4px 10px;
		border-radius: 999px;
		background: #efeae4;
		color: #5c4536;
		font-size: 12px;
		font-weight: 700;
		white-space: nowrap;
	}
	.nivel-chip.oro {
		background: #fcede5;
		color: #c24e25;
	}
	.p-d {
		font-size: 14px;
		color: #7a6252;
		margin-top: 4px;
	}
	.p-val {
		margin-top: 12px;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13.5px;
	}
	.p-val span {
		color: #7a6252;
	}
	.sumate {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 380px), 1fr));
		gap: clamp(20px, 4vw, 48px);
		align-items: start;
	}
	.s-txt {
		margin: 16px 0 0;
		font-size: 16.5px;
		line-height: 1.6;
		color: #5c4536;
	}
	.tarjeta {
		padding: clamp(20px, 3vw, 28px);
		border-radius: 28px;
		background: #fff;
		box-shadow:
			0 2px 4px rgba(58, 29, 18, 0.05),
			0 30px 60px -30px rgba(58, 29, 18, 0.45);
	}
	.campos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 10px;
	}
	.err-form {
		margin-bottom: 14px;
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
		background: #2a140c;
		color: #fff;
		font-size: 15px;
		font-weight: 600;
	}
	.enviar:disabled {
		opacity: 0.7;
	}
	.exito {
		text-align: center;
		padding: 20px 8px;
	}
	.exito img {
		object-fit: contain;
	}
	.ex-t {
		margin-top: 8px;
		font-size: 21px;
		font-weight: 600;
	}
	.exito p {
		margin: 8px auto 0;
		max-width: 340px;
		font-size: 15px;
		line-height: 1.55;
		color: #5c4536;
	}
	@media (max-width: 640px) {
		.hero-in > img {
			width: 150px;
		}
		.ej-v {
			text-align: left;
		}
	}
</style>
