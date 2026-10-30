<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import TplMini from '#lib/components/TplMini.svelte';
	import * as I from '#lib/icons.ts';
	import { PLANTILLAS } from '#lib/plantillas.ts';
	import { slugDe } from '#lib/slug.ts';
	import { getLanding } from './estado.svelte.ts';

	/** Modal de plantilla de la landing (handoff: tplModal). */
	const l = getLanding();
	// Los CTA llevan al onboarding, que es el registro (handoff: Onboarding Mostrador.dc.html).
	const registro = resolve('/onboarding');
	const DESC = {
		ed: 'Imagen grande y colecciones destacadas. Ideal si tenés buenas fotos y una marca para mostrar.',
		grid: 'Muchos productos a la vista y compra rápida desde el celular. Ideal para catálogos que rotan.',
		cat: 'Categorías a un costado y combos destacados. Ideal para almacenes y tiendas con mucho surtido.'
	} as const;

	let celular = $state(false);
	let pal = $state(0);
	const abierto = $derived(l.tplI !== null);
	const t = $derived(PLANTILLAS[l.tplI ?? 0]);
	const paletas = $derived([
		{ bg: t.bg, tx: t.tx, ac: t.ac },
		{ bg: '#F3E9DC', tx: '#3B2416', ac: '#6B3A22' },
		{ bg: '#EEF1E8', tx: '#1E2B22', ac: '#2F5D50' },
		{ bg: '#141414', tx: '#FFFFFF', ac: '#E8875F' }
	]);

	function mover(d: number) {
		l.tplI = ((l.tplI ?? 0) + d + PLANTILLAS.length) % PLANTILLAS.length;
		pal = 0;
	}
	function usar() {
		// Como el handoff: el onboarding arranca con esta plantilla elegida.
		try {
			localStorage.setItem('ob-tpl', t.id);
			localStorage.setItem('ob-step', '0');
		} catch {
			// Sin almacenamiento: el onboarding arranca con la de su rubro.
		}
	}
</script>

<Dialog.Root
	open={abierto}
	onOpenChange={(o) => {
		if (!o) l.tplI = null;
	}}
>
	<Dialog.Portal>
		<Dialog.Overlay class="tm-fondo" />
		<Dialog.Content class="tm" data-lenis-prevent>
			<Dialog.Title class="sr-only">Plantilla {t.nombre}</Dialog.Title>
			<div class="prev">
				<div class="prev-cab">
					<div class="disp" role="group" aria-label="Dispositivo">
						<button type="button" aria-pressed={!celular} onclick={() => (celular = false)}
							><Icon icon={I.phDesktop} size={15} />Escritorio</button
						>
						<button type="button" aria-pressed={celular} onclick={() => (celular = true)}
							><Icon icon={I.phDeviceMobile} size={15} />Celular</button
						>
					</div>
					<span class="esp"></span>
					<button
						type="button"
						class="flecha"
						aria-label="Plantilla anterior"
						onclick={() => mover(-1)}><Icon icon={I.phCaretLeftBold} size={14} /></button
					>
					<button
						type="button"
						class="flecha"
						aria-label="Plantilla siguiente"
						onclick={() => mover(1)}><Icon icon={I.phCaretRightBold} size={14} /></button
					>
				</div>
				<div class="escenario">
					<div class="marco" class:celular>
						<div class="pantalla">
							<div class="zoom" style:zoom={celular ? 1.15 : 1.9}>
								<TplMini {t} paleta={pal ? paletas[pal] : null} />
							</div>
						</div>
					</div>
				</div>
			</div>
			<div class="lado">
				<div class="lado-cab">
					<div>
						<div class="rubro">{t.rubro}</div>
						<div class="nombre">{t.nombre}</div>
					</div>
					<Dialog.Close class="cerrar" aria-label="Cerrar"
						><Icon icon={I.phXBold} size={14} /></Dialog.Close
					>
				</div>
				<div class="desc">{DESC[t.layout]}</div>
				<div>
					<div class="sub">Probá otros colores</div>
					<div class="pals">
						{#each paletas as p, i (i)}
							<button
								type="button"
								class:on={pal === i}
								aria-label="Paleta {i + 1}"
								aria-pressed={pal === i}
								style:background="linear-gradient(135deg,{p.bg} 50%,{p.ac} 50%)"
								onclick={() => (pal = i)}
							></button>
						{/each}
					</div>
				</div>
				<div class="feats">
					{#each ['Checkout con Mercado Pago y cuotas', 'Botón de WhatsApp y catálogo', 'Optimizada para Google y celular', 'Editor de arrastrar y soltar'] as f (f)}
						<span><Icon icon={I.phCheckCircleFill} size={16} style="color:#237550" />{f}</span>
					{/each}
				</div>
				<div class="esp"></div>
				<!-- La plantilla va en la URL además del almacenamiento: llega aunque el navegador lo bloquee. -->
				<a class="usar" href="{registro}?plantilla={t.id}" onclick={usar}
					>Usar esta plantilla <Icon icon={I.phArrowRightBold} size={15} /></a
				>
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- vitrina pública de ejemplo -->
				<a class="ver" href="/t/{slugDe(t.tienda)}" target="_blank" rel="noopener"
					>Ver la tienda de ejemplo <Icon icon={I.phArrowUpRightBold} size={14} /></a
				>
				<div class="nota">Gratis en todos los planes · la cambiás cuando quieras</div>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	:global(.tm-fondo) {
		position: fixed;
		inset: 0;
		z-index: 96;
		background: rgba(32, 15, 8, 0.55);
		backdrop-filter: blur(6px);
	}
	:global(.tm-fondo[data-state='open']) {
		animation: tm-fade 260ms both;
	}
	:global(.tm) {
		position: fixed;
		left: 50%;
		top: 50%;
		z-index: 97;
		width: min(1100px, calc(100vw - 40px));
		height: min(720px, calc(100vh - 40px));
		translate: -50% -50%;
		border-radius: 28px;
		background: #fbf7f2;
		color: #2a140c;
		display: flex;
		overflow: hidden;
		box-shadow: 0 50px 120px -30px rgba(0, 0, 0, 0.7);
		outline: none;
	}
	:global(.tm[data-state='open']) {
		animation: tm-entra 340ms cubic-bezier(0.2, 0.7, 0.1, 1) both;
	}
	@keyframes tm-fade {
		from {
			opacity: 0;
		}
	}
	@keyframes tm-entra {
		from {
			opacity: 0;
			transform: translateY(10px) scale(0.96);
		}
	}
	.prev {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		background: #efe6da;
		display: flex;
		flex-direction: column;
	}
	.prev-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 14px 18px;
	}
	.disp {
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: rgba(58, 29, 18, 0.08);
	}
	.disp button {
		display: flex;
		align-items: center;
		gap: 6px;
		border: 0;
		font-size: 13px;
		font-weight: 600;
		padding: 7px 13px;
		border-radius: 999px;
		background: transparent;
		color: #2a140c;
	}
	.disp button[aria-pressed='true'] {
		background: #fff;
	}
	.esp {
		flex: 1;
	}
	.flecha {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		border: 0;
		background: #fff;
		display: grid;
		place-items: center;
		color: #2a140c;
	}
	.escenario {
		flex: 1;
		min-height: 0;
		display: grid;
		place-items: center;
		padding: 0 24px 24px;
	}
	.marco {
		width: 100%;
		height: 100%;
		border-radius: 16px;
		padding: 0;
		background: transparent;
		box-shadow: 0 30px 70px -30px rgba(42, 20, 12, 0.6);
		transition:
			width 0.45s cubic-bezier(0.2, 0.7, 0.1, 1),
			height 0.45s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.marco.celular {
		width: 300px;
		height: 600px;
		max-height: 100%;
		border-radius: 42px;
		padding: 10px;
		background: #2a140c;
	}
	.pantalla {
		width: 100%;
		height: 100%;
		border-radius: 16px;
		overflow: hidden;
		background: #fff;
	}
	.celular .pantalla {
		border-radius: 32px;
	}
	.zoom {
		height: 100%;
	}
	.lado {
		width: 330px;
		flex: none;
		padding: 26px;
		display: flex;
		flex-direction: column;
		gap: 16px;
		overflow: auto;
	}
	.lado-cab {
		display: flex;
		justify-content: space-between;
		align-items: start;
	}
	.rubro {
		font-size: 13px;
		color: #7a6252;
	}
	.nombre {
		font-size: 30px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.lado :global(.cerrar) {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		border: 0;
		background: #f1e7dc;
		display: grid;
		place-items: center;
		color: #2a140c;
	}
	.desc {
		font-size: 14.5px;
		line-height: 1.55;
		color: #5c4536;
	}
	.sub {
		font-size: 13px;
		font-weight: 600;
		margin-bottom: 8px;
	}
	.pals {
		display: flex;
		gap: 8px;
	}
	.pals button {
		width: 40px;
		height: 40px;
		border-radius: 12px;
		border: 2px solid transparent;
		padding: 0;
	}
	.pals button.on {
		border-color: #d2582f;
	}
	.feats {
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 14px;
	}
	.feats span {
		display: flex;
		gap: 8px;
		align-items: center;
	}
	.usar {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		color: #fff;
		font-size: 15.5px;
		font-weight: 600;
		padding: 15px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #ee7a4a 0%, #d2582f 55%, #bc4920 100%);
		box-shadow: var(--inflado-naranja);
	}
	.ver {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		margin-top: 10px;
		padding: 12px;
		border-radius: 999px;
		font-size: 14.5px;
		font-weight: 600;
		color: #b0441d;
		text-decoration: none;
		box-shadow: inset 0 0 0 1px rgba(176, 68, 29, 0.3);
	}
	.usar:hover {
		color: #fff;
	}
	.nota {
		text-align: center;
		font-size: 12.5px;
		color: #7a6252;
	}
	/* < 860 px: vista previa arriba (54 %) y el panel abajo (46 %), como el handoff. */
	@media (max-width: 859px) {
		:global(.tm) {
			flex-direction: column;
		}
		.prev {
			min-height: 54%;
		}
		.lado {
			width: 100%;
			padding: 18px;
			max-height: 46%;
		}
	}
</style>
