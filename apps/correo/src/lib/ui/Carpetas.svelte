<script lang="ts" module>
	/** Ícono y tono del cuadradito de cada carpeta (handoff: carp()). */
	export const ICONOS: Record<string, [string, number]> = {
		entrada: ['icon-inbox', 45],
		destacados: ['icon-star', 80],
		enviados: ['icon-send', 250],
		borradores: ['icon-file-pen', 60],
		programados: ['icon-clock', 300],
		archivo: ['icon-archive', 150],
		spam: ['icon-shield-alert', 25],
		papelera: ['icon-trash-2', 60]
	};
	/** El color de cada casilla, en el orden del handoff (BOX). */
	/** Cada dominio es una empresa o un producto del grupo: su nombre y su ícono (static/marcas). */
	export const EMPRESAS: Record<string, [string, string]> = {
		'grupovimo.com.ar': ['Grupo Vimo', '/marcas/grupovimo.svg'],
		'tulocalweb.com.ar': ['Tu Local Web', '/marcas/tulocalweb.svg'],
		'corredor.com.ar': ['Corredor', '/marcas/corredor.png'],
		'tengolibre.com.ar': ['Tengo Libre', '/marcas/tengolibre.png'],
		'caigo.com.ar': ['Caigo', '/marcas/caigo.svg']
	};
	export const COLORES = [
		'oklch(0.68 calc(0.19 * var(--ac)) calc(var(--h) + 0))',
		'oklch(0.55 0.13 250)',
		'oklch(0.55 0.12 148)',
		'oklch(0.62 0.08 65)'
	];
</script>

<script lang="ts">
	import type { Bandeja } from '#lib/server/bandeja.ts';
	import { tamano } from '#lib/formato.ts';

	let {
		datos,
		carpeta,
		casilla,
		href,
		onredactar
	}: {
		datos: Pick<Bandeja, 'carpetas' | 'casillas' | 'espacio'>;
		carpeta: string;
		casilla: string | null;
		href: (o: { c?: string; cas?: string | null }) => string;
		onredactar: () => void;
	} = $props();
	/** La propia arriba, sin título; las compartidas, por dominio (una empresa por bloque). */
	const grupos = $derived.by(() => {
		const out: { dominio: string; casillas: typeof datos.casillas }[] = [];
		const propias = datos.casillas.filter((c) => c.personal);
		if (propias.length) out.push({ dominio: '', casillas: propias });
		for (const c of datos.casillas.filter((c) => !c.personal)) {
			const d = c.email.split('@')[1];
			const g = out.find((x) => x.dominio === d);
			if (g) g.casillas.push(c);
			else out.push({ dominio: d, casillas: [c] });
		}
		// Grupo Vimo primero y después los productos, en el orden de EMPRESAS.
		const orden = (d: string) => (d ? Object.keys(EMPRESAS).indexOf(d) + 1 || 99 : 0);
		return out.sort((a, b) => orden(a.dominio) - orden(b.dominio));
	});
	const pct = $derived(
		datos.espacio ? Math.min(100, (datos.espacio.usado / datos.espacio.total) * 100) : 0
	);
</script>

<nav class="col" aria-label="Carpetas">
	<button class="redactar" onclick={onredactar}><i class="icon-pencil"></i>Redactar</button>
	<div class="panel vidrio">
		{#each datos.carpetas as c (c.id)}
			{@const [ic, tono] = ICONOS[c.id]}
			{@const act = carpeta === c.id}
			<a
				class="carpeta"
				class:act
				href={href({ c: c.id })}
				aria-current={act ? 'page' : undefined}
				data-sveltekit-reset="false"
			>
				<span
					class="sq"
					style:background={act ? 'var(--acc)' : `oklch(0.66 0.13 ${tono} / 0.18)`}
					style:color={act ? '#fff' : `var(--sq${tono})`}><i class={ic}></i></span
				>
				<span class="l">{c.l}</span>
				{#if c.n}<span class="n">{c.n}</span>{/if}
			</a>
		{/each}
		{#if datos.casillas.length > 1}
			<div class="sep"></div>
			<div class="tit">Casillas</div>
			{#each grupos as g (g.dominio)}
				{#if g.dominio}
					{@const e = EMPRESAS[g.dominio]}
					<div class="dom">
						{#if e}<img src={e[1]} alt="" width="18" height="18" />{/if}{e?.[0] ?? g.dominio}
					</div>
				{/if}
				{#each g.casillas as c (c.cuenta)}
					{@const act = casilla === c.cuenta}
					{@const i = datos.casillas.indexOf(c)}
					<a
						class="casilla"
						class:act
						href={href({ cas: act ? null : c.cuenta })}
						aria-current={act ? 'true' : undefined}
						data-sveltekit-reset="false"
					>
						<span class="ini" style:background={COLORES[i % COLORES.length]}
							>{c.email.charAt(0).toUpperCase()}</span
						>
						<span class="tx"
							><span class="nom">{c.nombre}</span><span class="dir">{c.email.split('@')[0]}@</span
							></span
						>
						{#if c.sinLeer}<span class="n">{c.sinLeer}</span>{/if}
					</a>
				{/each}
			{/each}
		{/if}
		{#if datos.espacio}
			<div class="espacio">
				<div class="barra"><div style:width="{Math.max(pct, 1)}%"></div></div>
				<div>{tamano(datos.espacio.usado)} de {tamano(datos.espacio.total)} usados</div>
			</div>
		{/if}
	</div>
</nav>

<style>
	.col {
		width: 272px;
		flex: none;
		display: flex;
		flex-direction: column;
		padding: 4px 14px 16px 16px;
		min-height: 0;
	}
	.redactar {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		height: 58px;
		flex: none;
		border-radius: 16px;
		background: var(--acc);
		color: #fff;
		font-size: 18px;
		font-weight: 600;
		box-shadow:
			0 2px 0 oklch(0.45 calc(0.17 * var(--ac)) calc(var(--h) - 6)),
			0 8px 20px -8px oklch(0.45 calc(0.17 * var(--ac)) var(--h) / 0.55);
	}
	.redactar i {
		font-size: 21px;
	}
	.panel {
		margin-top: 14px;
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		padding: 8px;
		border-radius: 16px;
		background: var(--panel);
		box-shadow: var(--panelSh);
	}
	.carpeta {
		display: flex;
		align-items: center;
		gap: 12px;
		height: 46px;
		flex: none;
		padding: 0 10px 0 8px;
		border-radius: 10px;
		font-size: 16px;
		font-weight: 500;
		color: var(--fg);
		text-decoration: none;
	}
	.carpeta.act {
		background: var(--itemAct);
		box-shadow: var(--itemSh);
		font-weight: 600;
	}
	.carpeta:not(.act):hover,
	.casilla:not(.act):hover {
		background: var(--sel);
	}
	.sq {
		width: 32px;
		height: 32px;
		flex: none;
		border-radius: 9px;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.sq i {
		font-size: 17px;
	}
	.l {
		flex: 1;
	}
	.n {
		min-width: 26px;
		height: 24px;
		padding: 0 8px;
		border-radius: 999px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font-size: 14px;
		font-weight: 600;
		background: var(--mu);
		color: var(--fg);
	}
	.carpeta.act .n {
		background: var(--acc);
		color: #fff;
	}
	.sep {
		height: 1px;
		flex: none;
		background: var(--ls);
		margin: 10px 8px;
	}
	.tit {
		padding: 2px 10px 6px;
		font-size: 13px;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--mut);
	}
	.dom {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 10px 4px;
		font-size: 13px;
		font-weight: 600;
		color: var(--mut);
	}
	.dom img {
		border-radius: 5px;
		object-fit: contain;
	}
	.casilla {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 50px;
		flex: none;
		padding: 4px 10px 4px 8px;
		border-radius: 10px;
		color: var(--fg);
		text-decoration: none;
	}
	.casilla.act {
		background: var(--itemAct);
		box-shadow: var(--itemSh);
	}
	.ini {
		width: 32px;
		height: 32px;
		flex: none;
		border-radius: 9px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #fff;
		font-size: 14px;
		font-weight: 700;
	}
	.tx {
		flex: 1;
		min-width: 0;
		line-height: 1.25;
	}
	.nom {
		display: block;
		font-size: 16px;
		font-weight: 500;
	}
	.dir {
		display: block;
		font-size: 13px;
		color: var(--mut);
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.espacio {
		margin-top: auto;
		padding: 10px 10px 4px;
		font-size: 14px;
		color: var(--mut);
	}
	.barra {
		height: 6px;
		border-radius: 3px;
		background: var(--mu);
		overflow: hidden;
		margin-bottom: 6px;
	}
	.barra div {
		height: 100%;
		border-radius: 3px;
		background: var(--acc);
	}
</style>
