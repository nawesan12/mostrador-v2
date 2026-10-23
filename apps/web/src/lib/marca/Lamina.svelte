<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { phFilePdfBold, phImageBold } from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { onMount, type Snippet } from 'svelte';
	import { ESCALA_IMPRENTA, aPng, descargar, imprimir } from './exportar.ts';
	import { usarFamilia } from './familia.svelte.ts';

	/**
	 * Una pieza de marca a su tamaño real. En pantalla se muestra escalada (la escala del handoff, o
	 * menos si no entra en el ancho), pero el PNG y el PDF salen de la pieza sin escalar.
	 */
	let {
		nombre,
		ancho,
		alto,
		mm,
		escala = 1,
		radio = 0,
		sombra = '',
		escalaPng,
		impresion = 1,
		sangrado = 3,
		compacta = false,
		children
	}: {
		/** Nombre del archivo, sin extensión. */
		nombre: string;
		/** Medida en px CSS (las piezas en mm también: 1 mm = 96 / 25,4 px). */
		ancho: number;
		/** Sin alto, la pieza crece con su contenido (firmas de email). */
		alto?: number;
		/** Medida de imprenta en mm: habilita el PDF. */
		mm?: [number, number];
		/** Escala en pantalla del handoff (las portadas de redes, al 40 %). */
		escala?: number;
		radio?: number;
		sombra?: string;
		/** Multiplicador del PNG; por defecto 300 ppp para imprenta y 1 para pantalla. */
		escalaPng?: number;
		/** Cuántas veces más grande es la pieza real que el diseño (roll-up 1:5 → 5). */
		impresion?: number;
		sangrado?: number;
		/** Sólo el ícono en los botones (tapas de destacadas, que miden 84 px). */
		compacta?: boolean;
		children: Snippet;
	} = $props();

	const fam = usarFamilia();
	const ui = getUi();
	let real = $state<HTMLDivElement>();
	let medido = $state(0);
	let ocupado = $state<'' | 'png' | 'pdf'>('');

	const s = $derived(fam.ancho > 0 ? Math.min(escala, fam.ancho / ancho) : escala);
	const h = $derived(alto ?? medido);
	const multiplo = $derived(escalaPng ?? (mm ? ESCALA_IMPRENTA : 1));
	const pieza = () => real!.firstElementChild as HTMLElement;
	const generar = () => aPng(pieza(), multiplo);

	onMount(() => fam.registrar({ nombre, png: generar }));

	async function png() {
		ocupado = 'png';
		try {
			descargar(await generar(), `${nombre}.png`);
		} catch {
			ui.toast({
				k: 'err',
				title: 'No se pudo armar el PNG',
				desc: 'Probá de nuevo; si sigue, recargá la página.'
			});
		} finally {
			ocupado = '';
		}
	}

	async function pdf() {
		if (!mm) return;
		ocupado = 'pdf';
		try {
			await imprimir(pieza(), { mm, sangrado, escala: impresion, titulo: nombre });
		} finally {
			ocupado = '';
		}
	}
</script>

<div class="lamina" data-lamina={nombre}>
	<div
		class="marco"
		style:width="{ancho * s}px"
		style:height="{h * s}px"
		style:border-radius={radio ? `${radio}px` : null}
		style:box-shadow={sombra || null}
	>
		<div
			class="real"
			bind:this={real}
			bind:offsetHeight={medido}
			style:width="{ancho}px"
			style:height={alto ? `${alto}px` : null}
			style:transform={s === 1 ? null : `scale(${s})`}
		>
			{@render children()}
		</div>
	</div>
	<div class="acciones" class:compacta>
		<button
			type="button"
			class="btn btn-secondary btn-sm"
			onclick={png}
			disabled={!!ocupado}
			aria-label={compacta ? `Descargar ${nombre} en PNG` : undefined}
			><Icon icon={phImageBold} size={15} />{#if !compacta}{ocupado === 'png'
					? 'Armando…'
					: 'PNG'}{/if}</button
		>
		{#if mm}
			<button
				type="button"
				class="btn btn-secondary btn-sm"
				onclick={pdf}
				disabled={!!ocupado}
				title="Se abre la impresión: elegí «Guardar como PDF»"
				><Icon icon={phFilePdfBold} size={15} />PDF</button
			>
		{/if}
	</div>
</div>

<style>
	.lamina {
		display: flex;
		flex-direction: column;
		gap: 10px;
		max-width: 100%;
	}
	.marco {
		position: relative;
		flex: none;
	}
	.real {
		position: absolute;
		left: 0;
		top: 0;
		transform-origin: 0 0;
	}
	.acciones {
		display: flex;
		gap: 6px;
	}
	.acciones.compacta {
		justify-content: center;
	}
	.acciones .btn {
		background: rgba(255, 255, 255, 0.7);
	}
	@media (max-width: 760px) {
		.acciones .btn {
			min-height: 44px;
			min-width: 44px;
		}
	}
</style>
