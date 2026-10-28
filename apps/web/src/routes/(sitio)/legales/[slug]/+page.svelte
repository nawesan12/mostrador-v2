<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { LEGALES } from '#lib/legales.ts';
	import type { PageProps } from './$types';

	/** Las 8 páginas legales del handoff (design/Legal *.dc.html), con sus bloques propios. */
	let { data, form }: PageProps = $props();
	const legal = $derived(data.legal);

	const err = $derived((form && 'errores' in form ? form.errores : {}) as Record<string, string>);
	const previo = $derived((form && 'email' in form ? form : {}) as Record<string, string>);
	const codigo = $derived(form && 'codigo' in form ? (form.codigo as string) : null);
	let enviando = $state(false);
	let otro = $state(false);
	$effect(() => {
		// Un envío nuevo vuelve a mostrar el resultado.
		if (form) otro = false;
	});

	const conEnvio = () => {
		enviando = true;
		return async ({ update }: { update: (o?: { reset?: boolean }) => Promise<void> }) => {
			await update({ reset: false });
			enviando = false;
		};
	};
	// "En esta página" abierto en escritorio; en el celular arranca plegado.
	let indiceAbierto = $state(true);
	$effect(() => {
		if (matchMedia('(max-width: 760px)').matches) indiceAbierto = false;
	});
	const configurar = () => window.dispatchEvent(new CustomEvent('mostrador:configurar-cookies'));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve, svelte/no-at-html-tags -- links a rutas públicas fijas y anclas; el HTML es el texto de los legales del handoff, que scripts/legales.ts deja sólo con <b> y <a> -->
<Seo
	title="{legal.titulo} · Mostrador"
	description={legal.descripcion}
	path="/legales/{legal.slug}"
/>

<section class="hero">
	<div class="ancho">
		<div class="hero-in">
			<div class="hero-txt">
				<div class="miga">
					<a href="/legales/terminos">Legales</a><span>/</span><span class="actual"
						>{legal.miga}</span
					>
				</div>
				<h1>{legal.titulo}</h1>
				<div class="chips">
					{#each legal.chips as c (c)}<span>{c}</span>{/each}
				</div>
			</div>
			<img src="/hornero/{legal.ilustracion}.webp" alt="" width="170" height="170" />
		</div>
	</div>
</section>

<div class="ancho">
	<div class="aviso">
		<Icon icon={I.phWarningCircleFill} size={18} style="flex:none" />
		<span>Texto de ejemplo para el diseño. Revisalo con un abogado antes de publicarlo.</span>
	</div>
	<div class="cuerpo">
		<aside>
			<div class="grupo">
				<div class="grupo-t">Legales</div>
				<nav class="legales" aria-label="Páginas legales">
					{#each LEGALES as l (l.slug)}
						<a
							href="/legales/{l.slug}"
							class:on={l.slug === legal.slug}
							aria-current={l.slug === legal.slug ? 'page' : undefined}>{l.corto}</a
						>
					{/each}
				</nav>
			</div>
			<details class="grupo indice" bind:open={indiceAbierto}>
				<summary class="grupo-t">En esta página</summary>
				<div class="toc">
					{#each legal.indice as [id, t] (id)}<a href="#{id}">{t}</a>{/each}
				</div>
			</details>
		</aside>

		<article>
			{#each legal.bloques as b, i (i)}
				{#if b.t === 'resumen'}
					<div class="resumen" class:arriba={i === 0}>
						<div class="resumen-t"><Icon icon={I.phSparkleFill} size={15} />En pocas palabras</div>
						<ul class="vinetas">
							{#each b.items as it (it)}<li>
									<span class="punto"></span><span>{@html it}</span>
								</li>{/each}
						</ul>
					</div>
				{:else if b.t === 'seccion'}
					<section id={b.id}>
						<h2><span>{b.n}</span>{b.titulo}</h2>
						{#each b.cuerpo as c, j (j)}
							{#if 'p' in c}
								<p>{@html c.p}</p>
							{:else}
								<ul class="vinetas">
									{#each c.ul as it (it)}<li>
											<span class="punto"></span><span>{@html it}</span>
										</li>{/each}
								</ul>
							{/if}
						{/each}
					</section>
				{:else if b.t === 'organismos'}
					<div class="organismos">
						{#each b.items as o (o.titulo)}
							<a
								href={o.href}
								target={o.href.startsWith('http') ? '_blank' : undefined}
								rel={o.href.startsWith('http') ? 'noopener' : undefined}
							>
								<Icon icon={o.icono} size={24} style="color:#b0441d" />
								<b>{o.titulo}</b>
								<span class="o-txt">{o.texto}</span>
								<span class="o-ir">Ir<Icon icon={I.phArrowUpRightBold} size={12} /></span>
							</a>
						{/each}
					</div>
				{:else if b.t === 'creditos'}
					<div class="creditos">
						{#each b.items as c (c.umbral)}
							<div>
								<div class="c-et">{c.etiqueta}</div>
								<div class="c-um">{c.umbral}</div>
								<div class="c-cr">{c.credito}</div>
								<div class="c-pie">{c.pie}</div>
							</div>
						{/each}
					</div>
				{:else if b.t === 'cookies'}
					<div class="tabla" role="table" aria-label="Cookies que usamos">
						<div class="fila cab" role="row">
							<span role="columnheader">Cookie</span><span role="columnheader">Tipo</span><span
								role="columnheader">Para qué</span
							><span role="columnheader">Duración</span>
						</div>
						{#each b.filas as f (f.nombre)}
							<div class="fila" role="row">
								<span role="cell" class="nombre">{f.nombre}</span>
								<span role="cell" class="tipo" style:color={f.color}>{f.tipo}</span>
								<span role="cell">{f.para}</span>
								<span role="cell">{f.dura}</span>
							</div>
						{/each}
					</div>
				{:else if b.t === 'configurar-cookies'}
					<button type="button" class="configurar" onclick={configurar}
						><Icon icon={I.phSlidersHorizontalBold} size={15} />Configurar mis cookies</button
					>
				{:else if b.t === 'arrepentimiento'}
					<div class="revocar">
						{#if codigo && !otro}
							<div class="exito" role="status">
								<img src="/hornero/exito.webp" alt="" width="96" height="96" />
								<div class="exito-txt">
									<b>Recibimos tu pedido</b>
									<p>
										Tu código de trámite es <b class="codigo">{codigo}</b>. Cancelamos el débito
										automático y te devolvemos el total en hasta 10 días hábiles.
									</p>
									<button type="button" onclick={() => (otro = true)}>Hacer otro pedido</button>
								</div>
							</div>
						{:else}
							<form method="POST" action="?/revocar" use:enhance={conEnvio} novalidate>
								<div class="rv-cab">
									<span class="rv-ico"><Icon icon={I.phArrowCounterClockwiseBold} size={19} /></span
									>
									<b>Revocar mi contratación</b>
								</div>
								{#if err.form}<div class="err-form" role="alert">{err.form}</div>{/if}
								<div class="campos">
									{#each [['email', 'Email de la cuenta', 'lucia@laesquina.com.ar', 'email', 'email'], ['nombre', 'Nombre y apellido', 'Lucía Fernández', 'text', 'name'], ['documento', 'CUIT o DNI', '27-38412905-4', 'text', 'off'], ['plan', 'Plan contratado', 'Pro mensual · 14 de octubre de 2026', 'text', 'off']] as [n, t, ej, tipo, ac] (n)}
										<div class="campo">
											<label for="rv-{n}">{t}</label>
											<input
												id="rv-{n}"
												name={n}
												type={tipo}
												autocomplete={ac as AutoFill}
												inputmode={n === 'documento' ? 'numeric' : undefined}
												placeholder={ej}
												defaultValue={previo[n] ?? ''}
												aria-invalid={err[n] ? 'true' : undefined}
												aria-describedby={err[n] ? `rv-err-${n}` : undefined}
											/>
											{#if err[n]}<small class="err" id="rv-err-{n}">{err[n]}</small>{/if}
										</div>
									{/each}
								</div>
								<div class="campo">
									<label for="rv-motivo">Motivo (opcional)</label>
									<textarea
										id="rv-motivo"
										name="motivo"
										rows="3"
										placeholder="Contanos si querés, no es obligatorio"
										defaultValue={previo.motivo ?? ''}
										aria-invalid={err.motivo ? 'true' : undefined}
										aria-describedby={err.motivo ? 'rv-err-motivo' : undefined}></textarea>
									{#if err.motivo}<small class="err" id="rv-err-motivo">{err.motivo}</small>{/if}
								</div>
								<input
									class="trampa"
									name="web"
									tabindex="-1"
									autocomplete="off"
									aria-hidden="true"
								/>
								<button class="enviar" type="submit" disabled={enviando}
									>{enviando ? 'Enviando…' : 'Enviar pedido de revocación'}</button
								>
								<div class="rv-pie">
									Sin costo y sin dar explicaciones. Te llega la confirmación por email.
								</div>
							</form>
						{/if}
					</div>
				{/if}
			{/each}

			<div class="cierre">
				<img src="/hornero/hablando.webp" alt="" width="56" height="56" />
				<div class="cierre-txt">
					<b>{legal.cierre.titulo}</b>
					<div>{legal.cierre.texto}</div>
				</div>
				<a href="/contacto">Contactanos</a>
			</div>
		</article>
	</div>
</div>

<style>
	.ancho {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 clamp(16px, 3vw, 28px);
	}
	.hero {
		background: #f4ece2 url('/textura/arch-pattern-light.svg');
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
	}
	.hero-in {
		padding: clamp(36px, 6vw, 72px) 0 clamp(28px, 4vw, 44px);
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		gap: 24px;
		flex-wrap: wrap;
	}
	.hero-txt {
		min-width: 0;
		flex: 1 1 520px;
	}
	.miga {
		display: flex;
		gap: 8px;
		font-size: 13.5px;
		color: #7a6252;
	}
	.miga a {
		color: #7a6252;
		text-decoration: underline;
	}
	.miga .actual {
		color: #2a140c;
	}
	h1 {
		margin: 14px 0 0;
		font-size: clamp(40px, 6vw, 76px);
		line-height: 0.98;
		letter-spacing: -0.045em;
		font-weight: 600;
	}
	.chips {
		margin-top: 16px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		font-size: 13.5px;
	}
	.chips span {
		padding: 6px 11px;
		border-radius: 999px;
		background: #fff;
		white-space: nowrap;
	}
	.hero-in > img {
		width: clamp(110px, 14vw, 170px);
		height: auto;
		flex: none;
	}
	.aviso {
		margin-top: 22px;
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 16px;
		border-radius: 14px;
		background: rgba(227, 169, 59, 0.14);
		color: #7a4f0a;
		font-size: 13.5px;
	}
	.cuerpo {
		display: flex;
		gap: clamp(24px, 4vw, 56px);
		align-items: flex-start;
		flex-wrap: wrap;
		padding: 28px 0 clamp(56px, 8vw, 100px);
	}
	aside {
		flex: 0 1 250px;
		min-width: 200px;
		position: sticky;
		top: 88px;
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.grupo-t {
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: #7a6252;
		margin-bottom: 8px;
	}
	.legales {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.legales a {
		padding: 9px 12px;
		border-radius: 12px;
		font-size: 14px;
		font-weight: 500;
		color: #5c4536;
	}
	.legales a:hover {
		background: rgba(58, 29, 18, 0.04);
		color: #2a140c;
	}
	.legales a.on {
		font-weight: 600;
		color: #2a140c;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.08);
	}
	.indice summary {
		list-style: none;
		pointer-events: none;
	}
	.indice summary::-webkit-details-marker {
		display: none;
	}
	.toc {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding-left: 12px;
		border-left: 2px solid rgba(58, 29, 18, 0.1);
	}
	.toc a {
		font-size: 13.5px;
		color: #5c4536;
	}
	.toc a:hover {
		color: #b0441d;
	}
	article {
		flex: 1 1 520px;
		min-width: 0;
		max-width: 780px;
	}
	.resumen {
		margin-top: 22px;
		padding: 20px 22px;
		border-radius: 22px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.resumen.arriba {
		margin-top: 0;
	}
	.resumen-t {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		font-weight: 700;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: #b0441d;
	}
	.vinetas {
		margin: 12px 0 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.vinetas li {
		display: flex;
		gap: 10px;
		font-size: 15.5px;
		line-height: 1.6;
		color: #4a3428;
	}
	.punto {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #d2582f;
		flex: none;
		margin-top: 10px;
	}
	article section {
		padding-top: 34px;
		scroll-margin-top: 84px;
	}
	h2 {
		margin: 0;
		font-size: clamp(21px, 2.2vw, 25px);
		font-weight: 600;
		letter-spacing: -0.02em;
		display: flex;
		gap: 10px;
		align-items: baseline;
	}
	h2 span {
		color: #b0441d;
		font-size: 0.8em;
	}
	article p {
		margin: 12px 0 0;
		font-size: 16px;
		line-height: 1.7;
		color: #4a3428;
	}
	article :global(a) {
		color: #b0441d;
	}
	.organismos {
		margin-top: 18px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 8px;
	}
	.organismos a {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 16px;
		border-radius: 18px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		color: #2a140c;
	}
	.organismos b {
		font-size: 15px;
	}
	.o-txt {
		font-size: 13px;
		color: #7a6252;
	}
	.o-ir {
		font-size: 13px;
		font-weight: 600;
		color: #b0441d;
		display: flex;
		align-items: center;
		gap: 4px;
	}
	.creditos {
		margin-top: 18px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 8px;
	}
	.creditos > div {
		padding: 16px;
		border-radius: 18px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.c-et {
		font-size: 13px;
		color: #7a6252;
	}
	.c-um {
		font-size: 16px;
		font-weight: 600;
		margin-top: 2px;
	}
	.c-cr {
		margin-top: 10px;
		font-size: 28px;
		font-weight: 600;
		letter-spacing: -0.03em;
		color: #b0441d;
	}
	.c-pie {
		font-size: 12.5px;
		color: #7a6252;
	}
	.tabla {
		margin-top: 22px;
		border-radius: 20px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		overflow: hidden;
	}
	.fila {
		display: grid;
		grid-template-columns: minmax(0, 1.3fr) minmax(0, 0.8fr) minmax(0, 1.6fr) minmax(0, 0.7fr);
		gap: 10px;
		padding: 12px 16px;
		font-size: 14px;
		border-top: 1px solid rgba(58, 29, 18, 0.07);
	}
	.fila span {
		min-width: 0;
		overflow-wrap: anywhere;
	}
	.fila.cab {
		font-size: 12px;
		background: #f4ece2;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: #7a6252;
		border-top: 0;
	}
	.nombre,
	.tipo {
		font-weight: 600;
	}
	.configurar {
		margin-top: 14px;
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 12px 18px;
		border: 0;
		border-radius: 999px;
		background: #2a140c;
		color: #fff;
		font-size: 14px;
		font-weight: 600;
	}
	.revocar {
		margin-top: 22px;
		padding: clamp(18px, 3vw, 26px);
		border-radius: 24px;
		background: #fff;
		box-shadow:
			0 2px 4px rgba(58, 29, 18, 0.05),
			0 24px 50px -30px rgba(58, 29, 18, 0.4);
	}
	.rv-cab {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.rv-cab b {
		font-size: 18px;
	}
	.rv-ico {
		width: 38px;
		height: 38px;
		border-radius: 12px;
		background: #fcede5;
		display: grid;
		place-items: center;
		color: #b0441d;
	}
	.campos {
		margin-top: 16px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 10px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.campos + .campo {
		margin-top: 10px;
	}
	.campo label {
		font-size: 13px;
		font-weight: 600;
	}
	.campo input,
	.campo textarea {
		padding: 12px 14px;
		border-radius: 12px;
		border: 1.5px solid rgba(58, 29, 18, 0.14);
		background: #fbf7f2;
		font: 500 15px var(--fuente, inherit);
		color: #2a140c;
		outline: none;
		width: 100%;
	}
	.campo textarea {
		resize: vertical;
	}
	.campo input:focus,
	.campo textarea:focus {
		border-color: #d2582f;
		box-shadow: 0 0 0 4px rgba(210, 88, 47, 0.12);
	}
	.campo [aria-invalid='true'] {
		border-color: #b83a2e;
	}
	small.err {
		font-size: 12.5px;
		font-weight: 500;
		color: #b83a2e;
	}
	.err-form {
		margin-top: 14px;
		padding: 12px 14px;
		border-radius: 14px;
		background: rgba(184, 58, 46, 0.07);
		border: 1px solid rgba(184, 58, 46, 0.28);
		font-size: 14px;
	}
	.trampa {
		position: absolute;
		left: -9999px;
		width: 1px;
		height: 1px;
		opacity: 0;
	}
	.enviar {
		margin-top: 14px;
		width: 100%;
		padding: 15px;
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
	.rv-pie {
		margin-top: 8px;
		font-size: 12.5px;
		color: #7a6252;
		text-align: center;
	}
	.exito {
		display: flex;
		gap: 16px;
		align-items: center;
		flex-wrap: wrap;
	}
	.exito img {
		object-fit: contain;
	}
	.exito-txt {
		flex: 1;
		min-width: 220px;
	}
	.exito-txt > b {
		font-size: 20px;
	}
	.exito-txt p {
		margin: 6px 0 0;
		font-size: 15px;
		line-height: 1.55;
		color: #5c4536;
	}
	.codigo {
		color: #2a140c;
		white-space: nowrap;
	}
	.exito-txt button {
		margin-top: 10px;
		border: 0;
		background: transparent;
		padding: 0;
		font-size: 13.5px;
		font-weight: 600;
		color: #b0441d;
		text-decoration: underline;
	}
	.cierre {
		margin-top: 40px;
		padding: 18px 20px;
		border-radius: 20px;
		background: #2a140c;
		color: #fbf7f2;
		display: flex;
		gap: 14px;
		align-items: center;
		flex-wrap: wrap;
	}
	.cierre img {
		object-fit: contain;
	}
	.cierre-txt {
		flex: 1;
		min-width: 200px;
	}
	.cierre-txt b {
		font-size: 15.5px;
	}
	.cierre-txt div {
		font-size: 14px;
		color: #d9c8ba;
		margin-top: 2px;
	}
	.cierre a {
		padding: 11px 16px;
		border-radius: 999px;
		background: #fff;
		color: #2a140c;
		font-size: 14px;
		font-weight: 600;
	}

	/* Celular: el índice lateral sticky taparía el texto. Las páginas legales pasan a una fila de
	   chips deslizable y "En esta página" se pliega. */
	@media (max-width: 760px) {
		aside {
			position: static;
			flex: 1 1 100%;
			min-width: 0;
			gap: 10px;
		}
		.legales {
			flex-direction: row;
			overflow-x: auto;
			scrollbar-width: none;
			margin: 0 calc(-1 * clamp(16px, 3vw, 28px));
			padding: 2px clamp(16px, 3vw, 28px) 4px;
			gap: 6px;
		}
		.legales::-webkit-scrollbar {
			display: none;
		}
		.legales a {
			flex: none;
			white-space: nowrap;
			background: #fff;
			border: 1px solid rgba(58, 29, 18, 0.1);
			border-radius: 999px;
			padding: 8px 13px;
		}
		.legales a.on {
			background: #2a140c;
			color: #fff;
			border-color: #2a140c;
		}
		.indice {
			padding: 12px 14px;
			border-radius: 16px;
			background: #fff;
			box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		}
		.indice summary {
			pointer-events: auto;
			display: flex;
			justify-content: space-between;
			margin: 0;
			cursor: pointer;
		}
		.indice summary::after {
			content: '+';
			font-size: 16px;
			line-height: 1;
		}
		.indice[open] summary::after {
			content: '–';
		}
		.indice[open] .toc {
			margin-top: 10px;
		}
		.creditos {
			grid-template-columns: repeat(3, minmax(0, 1fr));
			gap: 6px;
		}
		.creditos > div {
			padding: 12px 10px;
		}
		.c-um {
			font-size: 14px;
		}
		.c-cr {
			font-size: 23px;
		}
		.hero-in {
			flex-wrap: nowrap;
			align-items: flex-start;
			gap: 12px;
		}
		.hero-in > img {
			width: 64px;
		}
		.fila {
			grid-template-columns: minmax(0, 1fr) auto;
			grid-template-areas: 'nombre tipo' 'para para' 'dura dura';
			gap: 2px 12px;
		}
		.fila.cab {
			display: none;
		}
		.fila > :nth-child(1) {
			grid-area: nombre;
		}
		.fila > :nth-child(2) {
			grid-area: tipo;
		}
		.fila > :nth-child(3) {
			grid-area: para;
		}
		.fila > :nth-child(4) {
			grid-area: dura;
			font-size: 12.5px;
			color: #7a6252;
		}
		/* Sin la cabecera de la tabla, la duración necesita su rótulo. */
		.fila > :nth-child(4)::before {
			content: 'Duración: ';
		}
	}
</style>
