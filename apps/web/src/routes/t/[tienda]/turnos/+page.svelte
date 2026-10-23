<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data, form } = $props();
</script>

<Seo
	title="Reservá tu turno · {data.tienda.nombre}"
	description="Elegí servicio, profesional y horario en {data.tienda.nombre}."
	path="/t/{data.tienda.slug}/turnos"
	noindex={false}
/>
<Plantilla tienda={data.tienda} pagina="turnos" error={form?.error ?? null} />
