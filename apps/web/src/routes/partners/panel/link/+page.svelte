<script lang="ts">
	import { renderSVG } from 'uqr';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { linkDe, nombreMes } from '#lib/partners/programa.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Link y cupón: el link único, su QR para descargar, el cupón y cómo rinden. */
	let { data } = $props();
	const ui = getUi();
	const p = $derived(data.panel!);
	const link = $derived(linkDe(p.partner.codigo));
	// El QR con los colores de las piezas de marca (Qr.svelte de /marca).
	const svg = $derived(
		renderSVG(link, { border: 4, ecc: 'M', whiteColor: '#FBF7F2', blackColor: '#2A140C' })
	);
	const qr = $derived('data:image/svg+xml;charset=utf-8,' + encodeURIComponent(svg));

	const mes = $derived(p.hoy.slice(0, 7));
	const clicsMes = $derived(p.clics.find((c) => c.mes === mes)?.n ?? 0);
	const altasMes = $derived(p.tiendas.filter((t) => t.alta.startsWith(mes)).length);
	const conversiones = $derived(p.tiendas.filter((t) => t.estado === 'activa').length);
	const tasa = $derived(p.tiendas.length ? Math.round((conversiones / p.tiendas.length) * 100) : 0);
	const maxClics = $derived(Math.max(1, ...p.clics.map((c) => c.n)));

	async function copiar(texto: string, que: string) {
		try {
			await navigator.clipboard.writeText(texto);
			ui.toast({ k: 'ok', title: `${que} copiado`, desc: texto });
		} catch {
			ui.toast({ k: 'warn', title: 'No se pudo copiar', desc: 'Seleccionalo y copialo a mano' });
		}
	}

	function bajar(href: string, nombre: string) {
		const a = document.createElement('a');
		a.href = href;
		a.download = nombre;
		a.click();
	}
	/** PNG de 1024 px: el SVG dibujado en un canvas. */
	function bajarPng() {
		const img = new Image();
		img.onload = () => {
			const c = document.createElement('canvas');
			c.width = c.height = 1024;
			const ctx = c.getContext('2d');
			if (!ctx) return;
			ctx.imageSmoothingEnabled = false;
			ctx.drawImage(img, 0, 0, 1024, 1024);
			bajar(c.toDataURL('image/png'), `qr-${p.partner.codigo}.png`);
		};
		img.src = qr;
	}
	const wa = $derived(
		`https://wa.me/?text=${encodeURIComponent(`Armá tu tienda online en Tu Local Web, 14 días gratis: ${link}`)}`
	);
</script>

<div class="dos" use:anim={{ t: 'rise' }}>
	<div class="tarjeta">
		<div class="t-cab"><span>Tu link</span></div>
		<div class="t-d">
			Cada tienda que se crea desde acá queda a tu nombre, aunque vuelva días&nbsp;después.
		</div>
		<div class="copiable">
			<span class="texto" data-testid="link-referido">{link}</span>
			<button type="button" class="oscuro" onclick={() => copiar(link, 'Link')}
				><Icon icon={I.phCopyBold} size={14} />Copiar</button
			>
		</div>
		<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- compartir por WhatsApp -->
		<a class="mas" href={wa} target="_blank" rel="noopener"
			><Icon icon={I.phWhatsappLogoFill} size={15} />Compartir por WhatsApp</a
		>

		<div class="t-cab cupon-t"><span>Tu cupón</span></div>
		<div class="t-d">
			Para quien llega sin el link: lo pone al crear la cuenta y la tienda queda a tu&nbsp;nombre.
		</div>
		<div class="copiable">
			<span class="texto cupon" data-testid="cupon">{p.partner.cupon}</span>
			<button type="button" class="oscuro" onclick={() => copiar(p.partner.cupon, 'Cupón')}
				><Icon icon={I.phCopyBold} size={14} />Copiar</button
			>
		</div>
	</div>

	<div class="tarjeta qr">
		<div class="t-cab"><span>QR para imprimir</span></div>
		<img src={qr} alt="QR de tu link de partner" width="200" height="200" />
		<div class="acc">
			<button type="button" class="oscuro" onclick={bajarPng}
				><Icon icon={I.phDownloadSimpleBold} size={14} />PNG</button
			>
			<button
				type="button"
				class="btn btn-secondary btn-sm"
				onclick={() => bajar(qr, `qr-${p.partner.codigo}.svg`)}
				><Icon icon={I.phDownloadSimpleBold} size={14} />SVG para imprenta</button
			>
		</div>
	</div>
</div>

<div class="kpis" use:anim={{ t: 'rise', delay: 40 }}>
	<div class="kpi">
		<div class="k-t">
			<Icon icon={I.phCursorClickFill} size={15} style="color:#b0441d" />Clics este mes
		</div>
		<div class="k-v">{clicsMes.toLocaleString('es-AR')}</div>
		<div class="k-d">Visitas a tu link</div>
	</div>
	<div class="kpi">
		<div class="k-t"><Icon icon={I.phUserPlusFill} size={15} style="color:#3E6BFD" />Altas</div>
		<div class="k-v">{p.tiendas.length}</div>
		<div class="k-d">{altasMes} este mes</div>
	</div>
	<div class="kpi">
		<div class="k-t">
			<Icon icon={I.phCheckCircleFill} size={15} style="color:#237550" />Conversiones
		</div>
		<div class="k-v">{conversiones}</div>
		<div class="k-d">Con plan pago</div>
	</div>
	<div class="kpi">
		<div class="k-t"><Icon icon={I.phTrendUpBold} size={15} style="color:#237550" />Conversión</div>
		<div class="k-v">{tasa}&nbsp;%</div>
		<div class="k-d">De las altas a plan pago</div>
	</div>
</div>

{#if p.clics.length}
	<div class="tarjeta" use:anim={{ t: 'rise', delay: 80 }}>
		<div class="t-cab"><span>Clics por mes</span></div>
		<div class="filas">
			{#each p.clics as c (c.mes)}
				<div class="mes">
					<span class="m-n">{nombreMes(c.mes)}</span>
					<span class="medidor"><span style:width="{(c.n / maxClics) * 100}%"></span></span>
					<span class="v-v">{c.n.toLocaleString('es-AR')}</span>
				</div>
			{/each}
		</div>
	</div>
{/if}

<style>
	.dos {
		display: grid;
		grid-template-columns: minmax(0, 1.6fr) minmax(0, 1fr);
		gap: 12px;
		align-items: start;
	}
	.copiable {
		margin-top: 10px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 6px 6px 12px;
		border-radius: 14px;
		border: 1px solid var(--ln2);
		background: var(--bg);
	}
	.texto {
		flex: 1;
		min-width: 0;
		font-size: 14px;
		font-weight: 500;
		overflow-wrap: anywhere;
	}
	.cupon {
		font-weight: 700;
		letter-spacing: 0.04em;
	}
	.cupon-t {
		margin-top: 18px;
	}
	.qr {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
	}
	.qr .t-cab {
		align-self: stretch;
	}
	.qr img {
		margin-top: 12px;
		width: 200px;
		height: 200px;
		border-radius: 14px;
	}
	.acc {
		margin-top: 12px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		justify-content: center;
	}
	.mes {
		display: grid;
		grid-template-columns: 140px minmax(0, 1fr) 60px;
		align-items: center;
		gap: 12px;
		padding: 9px 0;
		font-size: 13.5px;
	}
	.mes .v-v {
		text-align: right;
	}
	@media (max-width: 899px) {
		.dos {
			grid-template-columns: minmax(0, 1fr);
		}
	}
	@media (max-width: 759px) {
		.mes {
			grid-template-columns: 110px minmax(0, 1fr) 44px;
		}
	}
</style>
