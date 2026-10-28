<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { rutaDe } from '#lib/panel/navegacion.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Canales de venta (design/PanelCrecer.dc.html, sCan): cuánto deja cada uno. */
	let { data } = $props();
	const ui = getUi();
	type C = (typeof data.canales)[number];

	const ESTILO: Record<
		C['id'],
		{ icono?: I.IconData; sigla?: string; col?: string; bg: string; barra: string }
	> = {
		web: { icono: I.phGlobeSimpleFill, col: '#D2582F', bg: '#FCEDE5', barra: '#D2582F' },
		wa: { icono: I.logosWhatsappIcon, bg: '#E9F8EE', barra: '#25D366' },
		ml: { sigla: 'ML', col: '#2D3277', bg: '#FFE600', barra: '#E5CF00' },
		pos: { icono: I.phStorefrontFill, col: '#8A5A3E', bg: '#F3E9DF', barra: '#8A5A3E' },
		ig: { icono: I.logosInstagramIcon, bg: '#FCEEF3', barra: '#E1306C' },
		tg: { icono: I.logosTelegram, bg: '#E8F4FB', barra: '#2AABEE' },
		gg: { icono: I.logosGoogleIcon, bg: '#EEF4FD', barra: '#4285F4' },
		tk: { icono: I.logosTiktokIcon, bg: '#F1F1F1', barra: '#111' }
	};
	const total = $derived(data.canales.reduce((s, c) => s + (c.activo ? c.vendido : 0), 0));
	const parte = (c: C) => (c.activo && total ? (c.vendido / total) * 100 : 0);
	const monto = (n: number) =>
		n >= 1e6
			? '$ ' + (n / 1e6).toLocaleString('es-AR', { maximumFractionDigits: 2 }) + ' M'
			: n
				? '$ ' + Math.round(n).toLocaleString('es-AR')
				: '—';

	async function alternar(c: C) {
		if (!c.conectado)
			return ui.toast({
				k: 'info',
				title: `Conectá ${c.nombre}`,
				desc: 'Se hace desde Integraciones en un minuto',
				action: 'Ir',
				onAction: () => {
					// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta del panel
					void goto(rutaDe('Integraciones'));
				}
			});
		const r = await fetch(`/panel/canales/${c.id}`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ activo: !c.activo })
		});
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo.' });
		await invalidateAll();
		ui.toast({
			k: c.activo ? 'warn' : 'ok',
			title: `${c.nombre} ${c.activo ? 'pausado' : 'activado'}`,
			desc: c.activo
				? 'Tus publicaciones quedan ocultas, no se borra nada'
				: 'Sincronizando productos y stock…'
		});
	}
</script>

<Seo title="Canales · Panel" description="Canales de venta" path="/panel/canales" noindex />

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<h1>Canales de venta</h1>
		<div class="c-s">Dónde vendés y cuánto deja cada uno, ya descontadas las comisiones</div>
	</div>

	<div class="tarjeta reparto" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="r-t">Ventas de {data.mes} por canal</div>
		<div class="barra">
			{#each data.canales as c (c.id)}
				{#if parte(c)}<span style:width="{parte(c)}%" style:background={ESTILO[c.id].barra}
					></span>{/if}
			{/each}
		</div>
		<div class="leyenda">
			{#each data.canales as c (c.id)}
				<span
					><span class="cuadro" style:background={ESTILO[c.id].barra}></span>{c.nombre}
					{Math.round(parte(c))}%</span
				>
			{/each}
		</div>
	</div>

	<div class="grilla" use:anim={{ t: 'rise', delay: 120 }}>
		{#each data.canales as c (c.id)}
			{@const e = ESTILO[c.id]}
			{@const estado = c.activo
				? c.vendido
					? 'Conectado y vendiendo'
					: 'Conectado'
				: c.conectado
					? 'Pausado'
					: 'Sin conectar'}
			<div class="tarjeta canal">
				<div class="ca-cab">
					<span class="ico" style:background={e.bg}
						>{#if e.icono}<Icon
								icon={e.icono}
								size={22}
								style="color:{e.col ?? 'inherit'}"
							/>{:else}<span class="sigla" style:color={e.col}>{e.sigla}</span>{/if}</span
					>
					<span class="ca-txt"
						><span class="ca-n">{c.nombre}</span><span class="ca-e" class:on={c.activo}
							><span class="punto"></span>{estado}</span
						></span
					>
					<Interruptor checked={c.activo} label="Activar {c.nombre}" onchange={() => alternar(c)} />
				</div>
				<div class="cifras">
					<div>
						<div class="k">Vendido</div>
						<div class="v">{monto(c.vendido)}</div>
					</div>
					<div>
						<div class="k">Pedidos</div>
						<div class="v">{c.pedidos}</div>
					</div>
					<div>
						<div class="k">Comisión</div>
						<div class="v" style:color={c.comision ? '#9A6A10' : '#2F8F63'}>
							{Math.round(c.comision * 100)}%
						</div>
					</div>
				</div>
				<div class="nota">{c.nota}</div>
			</div>
		{/each}
	</div>
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.tarjeta {
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.reparto {
		padding: 18px;
		border-radius: 24px;
	}
	.r-t {
		font-size: 14px;
		font-weight: 600;
	}
	.barra {
		margin-top: 12px;
		display: flex;
		height: 16px;
		border-radius: 10px;
		overflow: hidden;
		gap: 2px;
	}
	.leyenda {
		margin-top: 10px;
		display: flex;
		gap: 14px;
		flex-wrap: wrap;
		font-size: 12.5px;
		color: var(--mu);
	}
	.leyenda > span {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.cuadro {
		width: 9px;
		height: 9px;
		border-radius: 3px;
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(270px, 1fr));
		gap: 14px;
	}
	.canal {
		padding: 18px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.ca-cab {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.ico {
		width: 42px;
		height: 42px;
		border-radius: 13px;
		display: grid;
		place-items: center;
		flex: none;
	}
	.sigla {
		font-size: 9px;
		font-weight: 800;
	}
	.ca-txt {
		flex: 1;
		min-width: 0;
	}
	.ca-n {
		display: block;
		font-size: 15.5px;
		font-weight: 600;
	}
	.ca-e {
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 12.5px;
		color: var(--fa);
	}
	.ca-e.on {
		color: #237550;
	}
	.punto {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: currentColor;
	}
	.cifras {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 6px;
	}
	.k {
		font-size: 11.5px;
		color: var(--mu);
	}
	.v {
		font-size: 16px;
		font-weight: 600;
	}
	.nota {
		font-size: 12.5px;
		color: var(--mu);
		padding-top: 10px;
		border-top: 1px solid var(--ln);
	}
</style>
