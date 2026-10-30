<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { porcentaje } from '#lib/sitio/estado.ts';
	import type { PageProps } from './$types';

	/**
	 * Estado del sistema (design/Mostrador Estado.dc.html) con los datos del monitor horario del VPS.
	 * Lo que el diseño no contempla (algo caído, días sin datos, sin incidentes) se arma con sus
	 * mismas piezas y colores.
	 */
	let { data, form }: PageProps = $props();
	const e = $derived(data.estado);
	const caidos = $derived(e.componentes.filter((c) => c.ok === false));
	const sinDatos = $derived(e.componentes.every((c) => c.ok === null));
	// Días con chequeos: al principio son menos que el rango elegido, y se dice así.
	const medidos = $derived(
		Math.max(0, ...e.componentes.map((c) => c.barras.filter((b) => b.estado !== 'sin').length))
	);
	const periodo = $derived(
		medidos < data.dias
			? `en ${medidos === 1 ? 'el día' : `los ${medidos} días`} que llevamos midiendo`
			: `en los últimos ${data.dias} días`
	);
	const conFallas = $derived(
		e.componentes.some((c) => c.barras.some((b) => b.estado === 'falla' || b.estado === 'caida'))
	);
	const abiertos = $derived(e.incidentes.filter((i) => i.fin === null && i.tipo === 'incidente'));

	const hace = (ms: number | null) => {
		if (!ms) return 'todavía sin chequeos';
		const min = Math.max(1, Math.round((data.ahora - ms) / 60_000));
		if (min < 60) return `hace ${min} ${min === 1 ? 'minuto' : 'minutos'}`;
		const h = Math.round(min / 60);
		return `hace ${h} ${h === 1 ? 'hora' : 'horas'}`;
	};
	const fecha = (ms: number) =>
		new Intl.DateTimeFormat('es-AR', {
			day: 'numeric',
			month: 'short',
			year: 'numeric',
			timeZone: 'America/Argentina/Buenos_Aires'
		})
			.format(ms)
			.replace('.', '')
			.replaceAll(' de ', ' ');
	const COLOR = { ok: '#3FA877', falla: '#E3A93B', caida: '#D2582F', sin: '#E7DED3' } as const;
	const TITULO = {
		ok: 'Sin incidentes',
		falla: 'Incidente',
		caida: 'Caída',
		sin: 'Sin datos'
	} as const;
	const diaLargo = (d: string) =>
		new Intl.DateTimeFormat('es-AR', { day: 'numeric', month: 'long', timeZone: 'UTC' }).format(
			Date.parse(`${d}T12:00:00Z`)
		);

	let enviando = $state(false);
	const conEnvio = () => {
		enviando = true;
		return async ({ update }: { update: (o?: { reset?: boolean }) => Promise<void> }) => {
			await update({ reset: false });
			enviando = false;
		};
	};
</script>

<Seo
	title="Estado del sistema · Tu Local Web"
	description="Disponibilidad de las tiendas, el panel, los pagos, WhatsApp, ARCA y Mercado Libre, medida cada hora. Incidentes y avisos."
	path="/estado"
/>
<svelte:head>
	<link
		rel="alternate"
		type="application/rss+xml"
		title="Estado de Tu Local Web"
		href="/estado/rss.xml"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- rango por query y feed RSS -->
<div class="ancho">
	<div class="arriba">
		<h1>Estado del sistema</h1>
		<div class="general" class:problema={caidos.length > 0} class:neutro={sinDatos}>
			<span class="g-ico">
				<Icon
					icon={caidos.length ? I.phWarningBold : sinDatos ? I.phHourglassBold : I.phCheckBold}
					size={22}
				/>
			</span>
			<div class="g-txt">
				<div class="g-t">
					{#if caidos.length}
						Hay problemas en {caidos.map((c) => c.nombre).join(', ')}
					{:else if sinDatos}
						Todavía estamos midiendo
					{:else}
						Todos los sistemas funcionan
					{/if}
				</div>
				<div class="g-d">
					Actualizado {hace(e.actualizado)} · {porcentaje(e.disponibilidad)} de disponibilidad
					{periodo}
				</div>
			</div>
			<nav class="rango" aria-label="Período">
				{#each [90, 30, 7] as d (d)}
					<a
						href="?dias={d}"
						aria-current={data.dias === d ? 'true' : undefined}
						data-sveltekit-noscroll>{d} días</a
					>
				{/each}
			</nav>
		</div>
	</div>
</div>

<div class="ancho">
	<div class="lista">
		{#each e.componentes as c (c.id)}
			<div class="comp">
				<div class="c-cab">
					<span class="c-n"
						><span
							class="punto"
							style:background={c.ok === null ? '#C9B9A9' : c.ok ? '#3FA877' : '#E3A93B'}
						></span>{c.nombre}</span
					>
					<span class="c-p"
						>{c.ok === null ? 'Sin datos todavía' : porcentaje(c.disponibilidad)}</span
					>
				</div>
				<div class="barras" class:muchas={data.dias === 90}>
					{#each c.barras as b (b.dia)}
						<span title="{diaLargo(b.dia)}: {TITULO[b.estado]}" style:background={COLOR[b.estado]}
						></span>
					{/each}
				</div>
				<div class="c-pie"><span>Hace {data.dias} días</span><span>Hoy</span></div>
			</div>
		{/each}
	</div>
</div>

<div class="ancho">
	<div class="abajo">
		<div>
			<div class="h-t">Incidentes recientes</div>
			<div class="incidentes">
				{#each e.incidentes as i (i.id)}
					{@const estado =
						i.tipo === 'mantenimiento'
							? i.fin
								? ['Completado', '#3E6BFD']
								: ['En curso', '#3E6BFD']
							: i.fin
								? ['Resuelto', '#2F8F63']
								: ['En curso', '#C2861B']}
					<div class="inc">
						<div class="i-cab">
							<span class="i-f">{fecha(i.inicio)}</span>
							<span class="i-e" style:background="{estado[1]}1A" style:color={estado[1]}
								>{estado[0]}</span
							>
						</div>
						<div class="i-t">{i.titulo}</div>
						<p>{i.detalle}</p>
					</div>
				{:else}
					<div class="inc vacio">
						<Icon icon={I.phCheckCircleFill} size={22} style="color:#237550;flex:none" />
						<div>
							{#if conFallas}
								<div class="i-t">Sin incidentes informados</div>
								<p>
									Hubo chequeos fallidos aislados, marcados en las barras. Si algo se cae, aparece
									acá con lo que pasó y cuándo se&nbsp;resolvió.
								</p>
							{:else}
								<div class="i-t">Sin incidentes {periodo}</div>
								<p>Si algo falla, aparece acá con lo que pasó y cuándo se&nbsp;resolvió.</p>
							{/if}
						</div>
					</div>
				{/each}
				{#if abiertos.length}<span class="sr">Hay {abiertos.length} incidentes en curso.</span>{/if}
			</div>
		</div>
		<div class="avisos">
			<img src="/hornero/desenchufado.webp" alt="" width="84" height="84" />
			<div class="a-t">Enterate antes que tus clientes</div>
			<p>Te avisamos por email o WhatsApp si algo falla y cuando se resuelve.</p>
			{#if form && 'suscripto' in form}
				<div class="a-ok" role="status">
					<Icon icon={I.phCheckCircleFill} size={16} />Listo, te avisamos por {form.suscripto ===
					'email'
						? 'email'
						: 'WhatsApp'}.
				</div>
			{:else}
				<form method="POST" action="?/suscribir" use:enhance={conEnvio} novalidate>
					<input
						name="contacto"
						placeholder="Tu email o WhatsApp"
						aria-label="Tu email o WhatsApp"
						autocomplete="email"
						defaultValue={form && 'contacto' in form ? String(form.contacto) : ''}
						aria-invalid={form && 'error' in form ? 'true' : undefined}
						aria-describedby={form && 'error' in form ? 'est-err' : undefined}
					/>
					<button type="submit" disabled={enviando}>Suscribirme</button>
				</form>
				{#if form && 'error' in form}<small id="est-err">{form.error}</small>{/if}
			{/if}
			<div class="a-pie">También por <a href="/estado/rss.xml">RSS</a></div>
		</div>
	</div>
</div>

<style>
	.ancho {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 clamp(16px, 3vw, 28px);
	}
	.arriba {
		padding: clamp(40px, 6vw, 72px) 0 0;
	}
	h1 {
		margin: 0;
		font-size: clamp(40px, 5.6vw, 72px);
		line-height: 1;
		letter-spacing: -0.045em;
		font-weight: 600;
	}
	.general {
		margin-top: 24px;
		display: flex;
		align-items: center;
		gap: 16px;
		flex-wrap: wrap;
		padding: 22px 24px;
		border-radius: 24px;
		background: #e6f4ec;
		border: 1px solid rgba(47, 143, 99, 0.25);
		--fuerte: #2f8f63;
		--titulo: #1f5c40;
		--texto: #3f6e58;
		--halo: rgba(47, 143, 99, 0.15);
	}
	.general.problema {
		background: #fff4dc;
		border-color: rgba(227, 169, 59, 0.4);
		--fuerte: #e3a93b;
		--titulo: #7a4f0a;
		--texto: #8a6420;
		--halo: rgba(227, 169, 59, 0.2);
	}
	.general.neutro {
		background: #f4ece2;
		border-color: rgba(58, 29, 18, 0.12);
		--fuerte: #9a8576;
		--titulo: #2a140c;
		--texto: #5c4536;
		--halo: rgba(58, 29, 18, 0.08);
	}
	.g-ico {
		width: 46px;
		height: 46px;
		border-radius: 50%;
		background: var(--fuerte);
		color: #fff;
		display: grid;
		place-items: center;
		flex: none;
		box-shadow: 0 0 0 8px var(--halo);
	}
	.g-txt {
		flex: 1;
		min-width: 220px;
	}
	.g-t {
		font-size: 21px;
		font-weight: 600;
		color: var(--titulo);
	}
	.g-d {
		font-size: 14px;
		color: var(--texto);
		margin-top: 2px;
	}
	.rango {
		display: flex;
		padding: 4px;
		border-radius: 999px;
		background: #fff;
		font-size: 13px;
		font-weight: 600;
	}
	.rango a {
		padding: 7px 12px;
		border-radius: 999px;
		color: #2a140c;
		white-space: nowrap;
	}
	.rango a[aria-current='true'] {
		background: #2a140c;
		color: #fff;
	}
	.lista {
		margin-top: 16px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		padding: 6px 22px;
	}
	.comp {
		padding: 18px 0;
	}
	.comp + .comp {
		border-top: 1px solid rgba(58, 29, 18, 0.07);
	}
	.c-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		margin-bottom: 10px;
	}
	.c-n {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 15.5px;
		font-weight: 600;
	}
	.punto {
		width: 9px;
		height: 9px;
		border-radius: 50%;
		flex: none;
	}
	.c-p {
		font-size: 13.5px;
		color: #7a6252;
		white-space: nowrap;
	}
	.barras {
		display: flex;
		gap: 2px;
		height: 30px;
	}
	.barras span {
		flex: 1;
		min-width: 0;
		border-radius: 2px;
	}
	.c-pie {
		display: flex;
		justify-content: space-between;
		margin-top: 6px;
		font-size: 11.5px;
		color: var(--marron-2);
	}
	.abajo {
		padding: clamp(40px, 6vw, 64px) 0 clamp(56px, 8vw, 100px);
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 360px), 1fr));
		gap: clamp(18px, 3vw, 36px);
		align-items: start;
	}
	.h-t {
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.incidentes {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.inc {
		padding: 18px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.inc.vacio {
		display: flex;
		gap: 12px;
	}
	.i-cab {
		display: flex;
		justify-content: space-between;
		gap: 10px;
		flex-wrap: wrap;
	}
	.i-f {
		font-size: 13px;
		color: #7a6252;
	}
	.i-e {
		font-size: 12px;
		font-weight: 700;
		padding: 3px 9px;
		border-radius: 999px;
	}
	.i-t {
		margin-top: 6px;
		font-size: 16px;
		font-weight: 600;
	}
	.vacio .i-t {
		margin-top: 0;
	}
	.inc p {
		margin: 6px 0 0;
		font-size: 14.5px;
		line-height: 1.55;
		color: #5c4536;
	}
	.sr {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
	}
	.avisos {
		position: sticky;
		top: 88px;
		padding: 24px;
		border-radius: 24px;
		background: #2a140c;
		color: #fbf7f2;
	}
	.avisos img {
		object-fit: contain;
	}
	.a-t {
		margin-top: 8px;
		font-size: 20px;
		font-weight: 600;
	}
	.avisos > p {
		margin: 6px 0 0;
		font-size: 14.5px;
		line-height: 1.55;
		color: #d9c8ba;
	}
	form {
		margin-top: 14px;
		display: flex;
		padding: 5px;
		border-radius: 999px;
		background: rgba(251, 247, 242, 0.08);
	}
	form input {
		flex: 1;
		min-width: 0;
		padding: 10px 12px;
		border: 0;
		background: transparent;
		font-size: 14.5px;
		font-weight: 500;
		color: #fff;
		outline: none;
	}
	form input::placeholder {
		color: #bfae9c;
	}
	form button {
		padding: 10px 16px;
		border: 0;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
		font-size: 14px;
		font-weight: 600;
	}
	small {
		display: block;
		margin-top: 8px;
		font-size: 12.5px;
		color: #f5b796;
	}
	.a-ok {
		margin-top: 14px;
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #7fd3a6;
	}
	.a-pie {
		margin-top: 12px;
		font-size: 12.5px;
		color: #bfae9c;
	}
	.a-pie a {
		color: #f5b796;
		text-decoration: underline;
	}
	/* Celular: 90 barras en 340 px no entran con 2 px de espacio; se juntan. Y el aviso no queda fijo. */
	@media (max-width: 640px) {
		.lista {
			padding: 4px 16px;
		}
		.barras.muchas {
			gap: 1px;
		}
		.avisos {
			position: static;
		}
		.general {
			padding: 18px;
		}
	}
</style>
