<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { phUploadSimpleBold, phXBold } from '#lib/icons.ts';

	/**
	 * Campos editables arriba de las piezas que llevan datos de una persona o de un evento: lo que se
	 * escribe se ve al instante en las piezas (y en sus PNG y PDF). No se guarda en ningún lado.
	 */
	let {
		titulo,
		ayuda,
		campos,
		valores = $bindable(),
		foto = false
	}: {
		titulo: string;
		ayuda: string;
		campos: [clave: string, etiqueta: string][];
		valores: Record<string, string>;
		/** Suma "Foto": la imagen queda en el navegador (URL local). */
		foto?: boolean;
	} = $props();

	const id = $props.id();

	function elegir(e: Event) {
		const f = (e.currentTarget as HTMLInputElement).files?.[0];
		if (!f) return;
		if (valores.foto) URL.revokeObjectURL(valores.foto);
		valores.foto = URL.createObjectURL(f);
	}
	function quitar() {
		URL.revokeObjectURL(valores.foto);
		valores.foto = '';
	}
</script>

<div class="campos">
	<div class="cab">
		<b>{titulo}</b>
		<span>{ayuda}</span>
	</div>
	<div class="grilla">
		{#each campos as [clave, etiqueta] (clave)}
			<label class="field" for="{id}-{clave}"
				>{etiqueta}<input id="{id}-{clave}" class="input" bind:value={valores[clave]} /></label
			>
		{/each}
		{#if foto}
			<div class="field">
				Foto
				<div class="foto">
					<label class="btn btn-secondary">
						<Icon icon={phUploadSimpleBold} size={16} />{valores.foto ? 'Cambiar' : 'Elegir foto'}
						<input type="file" accept="image/*" onchange={elegir} />
					</label>
					{#if valores.foto}
						<button type="button" class="icon-btn" onclick={quitar} aria-label="Quitar la foto"
							><Icon icon={phXBold} size={16} /></button
						>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.campos {
		max-width: 1400px;
		padding: 20px;
		border-radius: var(--r-card);
		background: #fff;
		box-shadow: 0 18px 40px -24px rgba(42, 20, 12, 0.4);
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.cab {
		display: flex;
		flex-wrap: wrap;
		gap: 4px 12px;
		align-items: baseline;
	}
	.cab b {
		font-size: 17px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.cab span {
		font-size: 14px;
		color: var(--marron);
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
		gap: 12px;
	}
	.foto {
		display: flex;
		gap: 8px;
		align-items: center;
		min-height: 44px;
	}
	.foto input {
		position: absolute;
		width: 1px;
		height: 1px;
		opacity: 0;
	}
	.foto label {
		position: relative;
		cursor: pointer;
	}
	.foto label:has(input:focus-visible) {
		outline: 2px solid var(--terracota);
		outline-offset: 2px;
	}
</style>
