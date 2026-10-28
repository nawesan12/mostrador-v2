<script lang="ts">
	import { enhance, type SubmitFunction } from '$app/forms';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import { ESPECIALIDADES, NIVELES, centavos, fechaCorta } from '#lib/staff/estado.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Partners: postulaciones (aprobar o rechazar con motivo) y los activos con su comisión. */
	let { data } = $props();
	const ui = getUi();
	const puede = $derived(data.sesion.permisos.includes('partners.decidir'));
	const nuevas = $derived(data.partners.postulaciones.filter((p) => p.estado === 'nueva'));
	const resueltas = $derived(data.partners.postulaciones.filter((p) => p.estado !== 'nueva'));

	let rechazar = $state<{ id: string; nombre: string } | null>(null);
	let modal = $state(false);
	let error = $state<string | null>(null);
	const enviar: SubmitFunction =
		() =>
		async ({ result, update }) => {
			error = null;
			if (result.type === 'failure') {
				if (result.data?.demo) {
					modal = false;
					ui.toast({
						k: 'info',
						title: 'En la demo no se aplica',
						desc: 'Con la cuenta real le llega el mail'
					});
				} else error = (result.data?.error as string) ?? 'No se pudo.';
				return;
			}
			if (result.type === 'success') {
				modal = false;
				ui.toast({
					k: 'ok',
					title: result.data?.decision === 'aprobar' ? 'Partner aprobado' : 'Postulación rechazada',
					desc: 'Le avisamos por mail'
				});
			}
			await update();
		};
</script>

<Seo title="Partners · Equipo" description="Programa de partners" path="/staff/partners" noindex />

<div class="cab"><h1>Partners</h1></div>

<section class="tarjeta">
	<div class="t-t">Postulaciones nuevas</div>
	{#each nuevas as p (p.id)}
		<div class="fila">
			<span class="f-txt"
				><b>{p.agencia || p.nombre}</b><span class="f-s"
					>{p.nombre} · {p.email}{p.web ? ` · ${p.web}` : ''} · {fechaCorta(p.creadaEn)}</span
				><span class="esp"
					>{#each p.especialidades as e (e)}<span class="chip" style:background="var(--chip)"
							>{ESPECIALIDADES[e] ?? e}</span
						>{/each}</span
				></span
			>
			{#if puede}
				<div class="botones">
					<form method="POST" action="?/decidir" use:enhance={enviar}>
						<input type="hidden" name="id" value={p.id} /><input
							type="hidden"
							name="decision"
							value="aprobar"
						/>
						<button type="submit" class="boton">Aprobar</button>
					</form>
					<button
						type="button"
						class="borde"
						onclick={() => {
							rechazar = { id: p.id, nombre: p.agencia || p.nombre };
							error = null;
							modal = true;
						}}>Rechazar</button
					>
				</div>
			{/if}
		</div>
	{:else}
		<div class="vacio">No hay postulaciones nuevas.</div>
	{/each}
</section>

<section class="tarjeta">
	<div class="t-t">Partners activos</div>
	{#each data.partners.activos as a (a.codigo)}
		<div class="fila">
			<span class="f-txt"
				><b>{a.nombre}</b><span class="f-s"
					>Código {a.codigo} · {a.activas} de {a.tiendas}
					{a.tiendas === 1 ? 'tienda activa' : 'tiendas activas'}</span
				></span
			>
			<span class="chip" style:background="rgba(201,138,18,.16)" style:color="#8A5A08"
				>{NIVELES[a.nivel]?.t ?? a.nivel} · {NIVELES[a.nivel]?.pct ?? 0} %</span
			>
			<span class="monto">{centavos(a.comision)}<span class="f-s">por mes</span></span>
		</div>
	{:else}
		<div class="vacio">Todavía no hay partners activos.</div>
	{/each}
</section>

{#if resueltas.length}
	<section class="tarjeta">
		<div class="t-t">Resueltas</div>
		{#each resueltas as p (p.id)}
			<div class="fila">
				<span class="f-txt"
					><b>{p.agencia || p.nombre}</b><span class="f-s"
						>{p.estado === 'aprobada'
							? 'Aprobada'
							: `Rechazada${p.motivo ? `: ${p.motivo}` : ''}`}</span
					></span
				>
			</div>
		{/each}
	</section>
{/if}

<Modal bind:open={modal} titulo="Rechazar la postulación" ancho={440}>
	{#if rechazar}
		<form method="POST" action="?/decidir" class="modal" use:enhance={enviar}>
			<h2>Rechazar a {rechazar.nombre}</h2>
			<input type="hidden" name="id" value={rechazar.id} /><input
				type="hidden"
				name="decision"
				value="rechazar"
			/>
			<label class="campo"
				>Motivo (le llega por mail)<textarea name="motivo" maxlength="500" required
				></textarea></label
			>
			{#if error}<p class="aviso" role="alert">{error}</p>{/if}
			<div class="pie">
				<button type="button" class="borde" onclick={() => (modal = false)}>Cancelar</button>
				<button type="submit" class="boton">Rechazar</button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.fila {
		flex-wrap: wrap;
	}
	.esp {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin-top: 4px;
	}
	.botones {
		display: flex;
		gap: 8px;
	}
	.monto {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		font-weight: 600;
	}
	.modal {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.modal h2 {
		margin: 0;
		font-size: 20px;
		font-weight: 600;
	}
	.pie {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}
	.aviso {
		margin: 0;
		padding: 10px 12px;
		border-radius: 12px;
		background: #fcede5;
		color: #a8401c;
		font-size: 13.5px;
	}
</style>
