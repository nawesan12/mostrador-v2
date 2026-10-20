<script lang="ts">
	import { Dialog } from 'bits-ui';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { LAYOUT, PLANTILLAS, RUBROS_PLANTILLAS } from '#lib/plantillas.ts';
	import { RECOMENDADAS } from './datos.ts';
	import { getOnboarding } from './estado.svelte.ts';
	import Miniatura from './Miniatura.svelte';

	/** Galería de plantillas del paso Tu tienda (handoff), con filtro por rubro y búsqueda. */
	const o = getOnboarding();
	let rubro = $state<string>('Todos');
	let q = $state('');
	const rubros = $derived(
		RUBROS_PLANTILLAS.filter((r) => r === 'Todos' || PLANTILLAS.some((t) => t.rubro === r))
	);
	const lista = $derived(
		PLANTILLAS.filter(
			(t) =>
				(rubro === 'Todos' || t.rubro === rubro) &&
				(!q.trim() ||
					`${t.nombre} ${t.rubro} ${LAYOUT[t.layout]}`
						.toLowerCase()
						.includes(q.trim().toLowerCase()))
		)
	);
	function elegir(id: string) {
		o.e.plantilla = id;
		o.galeria = false;
		o.guardar();
	}
</script>

<Dialog.Root bind:open={o.galeria}>
	<Dialog.Portal>
		<Dialog.Overlay class="gal-fondo" />
		<Dialog.Content class="gal">
			<div class="cab">
				<div class="cab-t">
					<Dialog.Title class="g-t">Plantillas</Dialog.Title>
					<Dialog.Description class="g-d"
						>Todas usan tus productos, fotos y textos importados.</Dialog.Description
					>
				</div>
				<label class="buscar">
					<Icon icon={I.phMagnifyingGlass} size={15} style="color:#7A6252" />
					<input
						bind:value={q}
						placeholder="Buscar estilo o rubro"
						aria-label="Buscar estilo o rubro"
					/>
				</label>
				<Dialog.Close class="cerrar" aria-label="Cerrar"
					><Icon icon={I.phXBold} size={15} /></Dialog.Close
				>
			</div>
			<div class="rubros" role="group" aria-label="Rubro">
				{#each rubros as r (r)}
					<button type="button" aria-pressed={rubro === r} onclick={() => (rubro = r)}>{r}</button>
				{/each}
			</div>
			<div class="cuerpo">
				<div class="grilla">
					{#each lista as t (t.id)}
						{@const sel = t.id === o.e.plantilla}
						{@const rec = (RECOMENDADAS as readonly string[]).includes(t.id) && !sel}
						<button type="button" class="tarjeta" onclick={() => elegir(t.id)} aria-pressed={sel}>
							<div class="marco" class:sel>
								<Miniatura {t} />
								{#if sel}<span class="elegida"><Icon icon={I.phCheckBold} size={12} />Elegida</span
									>{/if}
								{#if rec}<span class="para"><Icon icon={I.phSparkleFill} size={11} />Para vos</span
									>{/if}
							</div>
							<div class="pie">
								<span class="n">{t.nombre}</span>
								<span class="r">{t.rubro} · {LAYOUT[t.layout]}</span>
							</div>
						</button>
					{/each}
				</div>
				{#if !lista.length}<div class="vacio">No hay plantillas con «{q}».</div>{/if}
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	:global(.gal-fondo) {
		position: fixed;
		inset: 0;
		z-index: 80;
		background: rgba(32, 15, 8, 0.5);
		backdrop-filter: blur(5px);
	}
	:global(.gal) {
		position: fixed;
		inset: 16px;
		z-index: 81;
		border-radius: 28px;
		background: #fbf7f2;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		box-shadow: 0 50px 120px -30px rgba(0, 0, 0, 0.6);
		outline: none;
	}
	:global(.gal[data-state='open']) {
		animation: gal-entra 420ms cubic-bezier(0.16, 1, 0.3, 1) both;
	}
	@keyframes gal-entra {
		from {
			opacity: 0;
			transform: translateY(24px) scale(0.98);
		}
	}
	.cab {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 20px 24px;
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
		flex-wrap: wrap;
	}
	.cab-t {
		flex: 1;
		min-width: 240px;
	}
	:global(.g-t) {
		font-size: 24px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	:global(.g-d) {
		font-size: 13.5px;
		color: #7a6252;
	}
	.buscar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 14px;
		border-radius: 999px;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.12);
		min-width: 220px;
	}
	.buscar input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 14px;
		font-weight: 500;
		color: #2a140c;
	}
	:global(.gal .cerrar) {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		border: 0;
		background: #f1e7dc;
		display: grid;
		place-items: center;
		color: #2a140c;
	}
	.rubros {
		display: flex;
		gap: 6px;
		padding: 14px 24px;
		overflow-x: auto;
		flex: none;
		scrollbar-width: none;
	}
	.rubros button {
		border: 1px solid rgba(58, 29, 18, 0.14);
		font-size: 13px;
		font-weight: 600;
		padding: 8px 14px;
		border-radius: 999px;
		background: #fff;
		color: #2a140c;
		white-space: nowrap;
	}
	.rubros button[aria-pressed='true'] {
		background: #2a140c;
		border-color: #2a140c;
		color: #fff;
	}
	.cuerpo {
		flex: 1;
		overflow: auto;
		padding: 4px 24px 24px;
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
		gap: 16px;
	}
	.tarjeta {
		padding: 0;
		border: 0;
		background: transparent;
		text-align: left;
		color: #2a140c;
		transition: transform 0.3s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.tarjeta:hover {
		transform: translateY(-3px);
	}
	.marco {
		position: relative;
		height: 290px;
		border-radius: 16px;
		overflow: hidden;
		background: #fff;
		box-shadow:
			0 0 0 2px transparent,
			0 20px 40px -24px rgba(58, 29, 18, 0.55);
	}
	.marco.sel {
		box-shadow:
			0 0 0 2px #d2582f,
			0 20px 40px -24px rgba(58, 29, 18, 0.55);
	}
	.elegida,
	.para {
		position: absolute;
		top: 10px;
		display: flex;
		align-items: center;
		gap: 5px;
		font-weight: 700;
		padding: 5px 10px;
		border-radius: 999px;
	}
	.elegida {
		right: 10px;
		font-size: 12px;
		color: #fff;
		background: #d2582f;
	}
	.para {
		left: 10px;
		font-size: 11.5px;
		color: #c24e25;
		background: #fff;
		padding: 5px 9px;
	}
	.pie {
		margin-top: 9px;
		display: flex;
		flex-direction: column;
	}
	.n {
		font-size: 15px;
		font-weight: 600;
	}
	.r {
		font-size: 12.5px;
		color: #7a6252;
	}
	.vacio {
		padding: 60px;
		text-align: center;
		color: #7a6252;
	}
	@media (max-width: 640px) {
		:global(.gal) {
			inset: 8px;
			border-radius: 22px;
		}
		.cab {
			padding: 16px;
		}
		.rubros,
		.cuerpo {
			padding-left: 16px;
			padding-right: 16px;
		}
		.buscar {
			min-width: 0;
			flex: 1 1 100%;
			order: 3;
		}
	}
</style>
