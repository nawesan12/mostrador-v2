<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { copiar, esTactil, hojaDelSistema, textoProducto } from '#lib/panel/compartir.ts';
	import { imagenProducto } from './compartir.ts';
	import { varsTema, type TemaOpiniones } from './opiniones.ts';
	import type { ProductoPublico, TiendaPublica } from './tipos.ts';

	/**
	 * «Compartir» chico en la ficha del producto, para quien compra: en el celular, la hoja del
	 * sistema con la imagen del producto y «Nombre · $ precio — link»; en la compu, copia el link.
	 */
	let {
		tienda,
		producto,
		tema
	}: { tienda: TiendaPublica; producto: ProductoPublico; tema: TemaOpiniones } = $props();

	let copiado = $state(false);
	let reloj: ReturnType<typeof setTimeout> | undefined;

	async function compartir() {
		// La dirección que está mirando (la de Tu Local Web o el dominio propio), sin la consulta.
		const url = location.origin + location.pathname;
		const texto = textoProducto(producto.nombre, producto.precio, url);
		if (esTactil()) {
			const listo = await hojaDelSistema({
				titulo: producto.nombre,
				texto,
				url,
				imagen: new URL(imagenProducto(tienda, producto)).pathname,
				archivo: producto.slug ?? 'producto'
			});
			if (listo === 'listo') return;
		}
		if (!(await copiar(url))) return;
		copiado = true;
		clearTimeout(reloj);
		reloj = setTimeout(() => (copiado = false), 2500);
	}
</script>

<button type="button" class="compartir" style={varsTema(tema)} onclick={compartir}
	><Icon icon={copiado ? I.phCheckBold : I.phShareNetworkBold} size={15} /><span aria-live="polite"
		>{copiado ? 'Link copiado' : 'Compartir'}</span
	></button
>

<style>
	.compartir {
		display: flex;
		align-items: center;
		justify-content: center;
		align-self: center;
		gap: 7px;
		width: fit-content;
		min-height: 44px;
		margin: 6px auto 0;
		padding: 0 14px;
		border: 0;
		border-radius: 999px;
		background: transparent;
		color: var(--op-mu);
		font: inherit;
		font-size: 13.5px;
		font-weight: 600;
		cursor: pointer;
	}
	.compartir:hover {
		color: var(--op-tx);
	}
</style>
