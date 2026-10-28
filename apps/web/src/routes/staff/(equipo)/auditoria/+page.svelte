<script lang="ts">
	import { resolve } from '$app/paths';
	import Seo from '#lib/components/Seo.svelte';
	import { nombreRol } from '#lib/staff/permisos.ts';

	/** Auditoría del equipo: cada lectura de una tienda o consulta y cada acción, con su autor. */
	let { data } = $props();
	const hora = (iso: string) =>
		new Intl.DateTimeFormat('es-AR', {
			timeZone: 'America/Argentina/Buenos_Aires',
			day: 'numeric',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit'
		}).format(new Date(iso));
	const acciones = $derived([...new Set(data.auditoria.registros.map((r) => r.accion))].sort());
</script>

<Seo
	title="Auditoría · Equipo"
	description="Auditoría del equipo"
	path="/staff/auditoria"
	noindex
/>

<div class="cab"><h1>Auditoría</h1></div>

<form method="GET" class="filtros">
	<label class="campo"
		>Quién<select name="miembro" onchange={(e) => e.currentTarget.form?.requestSubmit()}
			><option value="">Todo el equipo</option>{#each data.auditoria.miembros as m (m.id)}<option
					value={m.id}
					selected={m.id === data.filtros.miembro}>{m.nombre} · {nombreRol(m.rol)}</option
				>{/each}</select
		></label
	>
	<label class="campo"
		>Qué<select name="accion" onchange={(e) => e.currentTarget.form?.requestSubmit()}
			><option value="">Todas las acciones</option>{#each acciones as a (a)}<option
					value={a}
					selected={a === data.filtros.accion}>{a.replaceAll('_', ' ')}</option
				>{/each}</select
		></label
	>
</form>

<section class="tarjeta">
	{#each data.auditoria.registros as r (r.id)}
		<div class="fila">
			<span class="cuando f-s">{hora(r.cuando)}</span>
			<span class="f-txt"
				><span>{r.texto}</span><span class="f-s"
					>{r.miembro.nombre}{#if r.tiendaId && r.tienda}
						· <a href={resolve('/staff/(equipo)/tiendas/[id]', { id: r.tiendaId })}>{r.tienda}</a
						>{/if}</span
				></span
			>
		</div>
	{:else}
		<div class="vacio">Nada con esos filtros.</div>
	{/each}
</section>

<style>
	.filtros {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 220px), 1fr));
		gap: 10px;
	}
	.cuando {
		width: 110px;
		flex: none;
		font-variant-numeric: tabular-nums;
	}
	.f-txt a {
		color: inherit;
		font-weight: 600;
	}
	@media (max-width: 520px) {
		.fila {
			flex-direction: column;
			align-items: flex-start;
			gap: 2px;
		}
	}
</style>
