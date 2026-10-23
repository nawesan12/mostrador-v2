<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { phFileZipBold } from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { descargar, zip } from './exportar.ts';
	import { usarFamilia } from './familia.svelte.ts';

	/** "Descargar todo": un .zip con el PNG de cada pieza de la familia. */
	const fam = usarFamilia();
	const ui = getUi();
	let hecho = $state(-1);

	async function todo() {
		const piezas = fam.piezas;
		hecho = 0;
		try {
			const archivos = [];
			for (const p of piezas) {
				archivos.push({ nombre: `${p.nombre}.png`, blob: await p.png() });
				hecho++;
			}
			descargar(await zip(archivos), `mostrador-${fam.clave}.zip`);
		} catch {
			ui.toast({
				k: 'err',
				title: 'No se pudo armar el .zip',
				desc: 'Probá de nuevo o bajá las piezas de a una.'
			});
		} finally {
			hecho = -1;
		}
	}
</script>

<button type="button" class="btn btn-dark" onclick={todo} disabled={hecho >= 0 || !fam.total}>
	<Icon icon={phFileZipBold} size={17} />
	{#if hecho >= 0}Armando {Math.min(hecho + 1, fam.total)} de {fam.total}…{:else}Descargar todo{/if}
</button>
