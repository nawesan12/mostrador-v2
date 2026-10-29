<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Seo from '#lib/components/Seo.svelte';
	import MotorLista from '#lib/panel/listas/MotorLista.svelte';
	import { configTiendas } from '#lib/staff/tiendas.ts';

	/** Tiendas: la lista del panel con los filtros del equipo; cada fila abre su ficha. */
	let { data } = $props();
	const config = configTiendas();
</script>

<Seo title="Tiendas · Equipo" description="Tiendas de Tu Local Web" path="/staff/tiendas" noindex />

<MotorLista
	{config}
	lista={data.lista}
	ruta="/staff/tiendas/lista"
	exportar={false}
	abrir={(t) => goto(resolve('/staff/(equipo)/tiendas/[id]', { id: t.id }))}
/>
