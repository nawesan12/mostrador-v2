<script lang="ts">
	/**
	 * Opciones como chips (handoff): radio si se elige una, checkbox si se eligen varias. Son inputs
	 * de verdad, así el formulario anda sin JavaScript y se navega con el teclado.
	 */
	let {
		nombre,
		pregunta,
		opciones,
		varias = false,
		elegidas = [],
		error
	}: {
		nombre: string;
		pregunta: string;
		opciones: readonly (readonly [string, string])[];
		varias?: boolean;
		elegidas?: string[];
		error?: string;
	} = $props();
	const uid = $props.id();
</script>

<fieldset aria-describedby={error ? `${uid}-err` : undefined}>
	<legend>{pregunta}</legend>
	<div class="chips">
		{#each opciones as [v, t] (v)}
			<label>
				<input
					type={varias ? 'checkbox' : 'radio'}
					name={nombre}
					value={v}
					defaultChecked={elegidas.includes(v)}
				/><span>{t}</span>
			</label>
		{/each}
	</div>
	{#if error}<small id="{uid}-err">{error}</small>{/if}
</fieldset>

<style>
	fieldset {
		border: 0;
		margin: 12px 0 0;
		padding: 0;
		min-width: 0;
	}
	legend {
		padding: 0;
		font-size: 13px;
		font-weight: 600;
	}
	.chips {
		margin-top: 8px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	input {
		position: absolute;
		opacity: 0;
		width: 1px;
		height: 1px;
	}
	span {
		display: block;
		padding: 9px 13px;
		border-radius: 999px;
		border: 1.5px solid rgba(58, 29, 18, 0.14);
		background: #fff;
		font-size: 13.5px;
		font-weight: 600;
		white-space: nowrap;
		cursor: pointer;
		transition:
			border-color 0.15s,
			background 0.15s;
	}
	input:checked + span {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	input:focus-visible + span {
		box-shadow: 0 0 0 4px rgba(210, 88, 47, 0.18);
	}
	small {
		display: block;
		margin-top: 6px;
		font-size: 12.5px;
		font-weight: 500;
		color: #b83a2e;
	}
</style>
