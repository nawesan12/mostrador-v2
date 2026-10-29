<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import * as I from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { anim } from './anim.ts';
	import { pesos, pesosCortos } from './formato.ts';

	/** "Meta de octubre" (handoff): anillo, cuánto va, proyección y "Editar". */
	let {
		mes,
		objetivo,
		actual,
		proyeccion,
		carritos
	}: { mes: string; objetivo: number; actual: number; proyeccion: number; carritos: number } =
		$props();
	const ui = getUi();
	const parte = $derived(objetivo ? Math.min(1, actual / objetivo) : 0);
	const extra = $derived(
		carritos ? ` Recuperar los ${carritos} carritos abandonados te deja más cerca.` : ''
	);
	let editar = $state(false);
	let error = $state<string | null>(null);
	let enviando = $state(false);
</script>

<div class="meta" use:anim={{ t: 'rise', delay: 120 }}>
	<div class="luz"></div>
	<div class="cab">
		<span class="t">Meta de {mes}</span>
		<button type="button" class="editar" onclick={() => ((error = null), (editar = true))}
			>{objetivo ? 'Editar' : 'Poner meta'}</button
		>
	</div>
	<div class="anillo">
		<svg viewBox="0 0 120 120" aria-hidden="true">
			<circle cx="60" cy="60" r="52" fill="none" stroke="rgba(251,247,242,.1)" stroke-width="10"
			></circle>
			{#key parte}
				<circle
					use:anim={{ t: 'ring', target: parte }}
					cx="60"
					cy="60"
					r="52"
					fill="none"
					stroke="url(#meta-anillo)"
					stroke-width="10"
					stroke-linecap="round"
					stroke-dasharray="326.7"
					stroke-dashoffset={326.7 * (1 - parte)}
				></circle>
			{/key}
			<defs
				><linearGradient id="meta-anillo" x1="0" x2="1"
					><stop offset="0" stop-color="#E8875F"></stop><stop offset="1" stop-color="#D2582F"
					></stop></linearGradient
				></defs
			>
		</svg>
		<div class="centro">
			<div>
				<div class="pct">{objetivo ? Math.round((actual / objetivo) * 100) : 0}%</div>
				<div class="de">
					{#if objetivo}{pesosCortos(actual)} de {pesosCortos(objetivo)}{:else}{pesosCortos(actual)} vendidos{/if}
				</div>
			</div>
		</div>
	</div>
	<div class="nota">
		{#if objetivo}
			A este ritmo cerrás el mes en <b>{pesosCortos(proyeccion)}</b>.{extra}
		{:else}
			Poné una meta para el mes y te muestro cómo venís y a qué ritmo cerrás.
		{/if}
	</div>
</div>

<Modal bind:open={editar} titulo="Meta de {mes}" ancho={420}>
	<form
		method="POST"
		action="?/meta"
		use:enhance={() => {
			enviando = true;
			return async ({ result, update }) => {
				enviando = false;
				if (result.type === 'failure') {
					error = (result.data?.meta as string) ?? 'No pudimos guardar la meta.';
					return;
				}
				await update({ reset: false });
				editar = false;
				ui.toast({
					k: 'ok',
					title: 'Meta guardada',
					desc: `Te vamos mostrando cómo venís en ${mes}.`
				});
			};
		}}
	>
		<div class="m-ico"><Icon icon={I.phTarget} size={24} /></div>
		<div class="m-t">Meta de {mes}</div>
		<div class="m-d">Cuánto querés vender este mes, sumando todos los canales.</div>
		<label class="field m-campo">
			<span>Meta en pesos</span>
			<input
				class="input"
				name="objetivo"
				inputmode="numeric"
				autocomplete="off"
				value={objetivo ? pesos(objetivo) : ''}
				placeholder="$ 20.000.000"
				aria-invalid={error ? 'true' : undefined}
				aria-describedby={error ? 'meta-error' : undefined}
			/>
		</label>
		{#if error}<div class="m-error" id="meta-error">{error}</div>{/if}
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (editar = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary" disabled={enviando}>Guardar meta</button>
		</div>
	</form>
</Modal>

<style>
	.meta {
		position: relative;
		overflow: hidden;
		border-radius: 24px;
		background: #2a140c;
		color: #fbf7f2;
		padding: 24px;
		display: flex;
		flex-direction: column;
		box-shadow: var(--shd);
	}
	.luz {
		position: absolute;
		inset: 0;
		background: radial-gradient(360px 260px at 100% 0%, rgba(210, 88, 47, 0.35), transparent 70%);
	}
	.cab {
		position: relative;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.t {
		font-size: 15px;
		font-weight: 600;
	}
	.editar {
		border: 0;
		background: transparent;
		padding: 0;
		font-size: 13px;
		font-weight: 600;
		color: #f5b796;
	}
	.anillo {
		position: relative;
		margin: 18px auto;
		width: 190px;
		height: 190px;
	}
	.anillo svg {
		width: 100%;
		height: 100%;
		transform: rotate(-90deg);
	}
	.centro {
		position: absolute;
		inset: 0;
		display: grid;
		place-items: center;
		text-align: center;
	}
	.pct {
		font-size: 40px;
		font-weight: 600;
		letter-spacing: -0.05em;
	}
	.de {
		font-size: 12.5px;
		color: #d9c8ba;
	}
	.nota {
		position: relative;
		margin-top: auto;
		padding: 14px;
		border-radius: 16px;
		background: rgba(251, 247, 242, 0.08);
		border: 1px solid rgba(251, 247, 242, 0.12);
		font-size: 13.5px;
		line-height: 1.45;
		color: #eadbcf;
	}
	.nota b {
		color: #fff;
	}
	.m-ico {
		width: 48px;
		height: 48px;
		display: grid;
		place-items: center;
		border-radius: 15px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}
	.m-t {
		margin-top: 16px;
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14.5px;
		line-height: 1.5;
		color: var(--mu);
	}
	.m-campo {
		margin-top: 16px;
	}
	.m-error {
		margin-top: 8px;
		font-size: 13px;
		font-weight: 500;
		color: #b83a2e;
	}
	.m-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
</style>
