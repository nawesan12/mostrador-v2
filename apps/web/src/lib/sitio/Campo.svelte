<script lang="ts">
	/** Campo de los formularios del sitio (handoff): rótulo, input o textarea y el error debajo. */
	let {
		nombre,
		etiqueta,
		ejemplo = '',
		tipo = 'text',
		valor = '',
		error,
		filas,
		autocompletar,
		modo
	}: {
		nombre: string;
		etiqueta: string;
		ejemplo?: string;
		tipo?: 'text' | 'email' | 'tel' | 'url';
		valor?: string;
		error?: string;
		/** Con filas es un textarea. */
		filas?: number;
		autocompletar?: AutoFill;
		modo?: 'text' | 'email' | 'tel' | 'url' | 'numeric';
	} = $props();
	const id = $props.id();
</script>

<div class="campo">
	<label for={id}>{etiqueta}</label>
	{#if filas}
		<textarea
			{id}
			name={nombre}
			rows={filas}
			placeholder={ejemplo}
			defaultValue={valor}
			aria-invalid={error ? 'true' : undefined}
			aria-describedby={error ? `${id}-err` : undefined}></textarea>
	{:else}
		<input
			{id}
			name={nombre}
			type={tipo}
			placeholder={ejemplo}
			defaultValue={valor}
			autocomplete={autocompletar}
			inputmode={modo}
			aria-invalid={error ? 'true' : undefined}
			aria-describedby={error ? `${id}-err` : undefined}
		/>
	{/if}
	{#if error}<small id="{id}-err">{error}</small>{/if}
</div>

<style>
	.campo {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
	}
	label {
		font-size: 13px;
		font-weight: 600;
	}
	input,
	textarea {
		padding: 13px 14px;
		border-radius: 12px;
		border: 1.5px solid rgba(58, 29, 18, 0.14);
		background: #fbf7f2;
		font-size: 15px;
		font-weight: 500;
		color: #2a140c;
		outline: none;
		width: 100%;
		transition:
			border-color 0.15s,
			box-shadow 0.15s;
	}
	textarea {
		resize: vertical;
	}
	input:focus,
	textarea:focus {
		border-color: #d2582f;
		box-shadow: 0 0 0 4px rgba(210, 88, 47, 0.12);
	}
	[aria-invalid='true'] {
		border-color: #b83a2e;
	}
	small {
		font-size: 12.5px;
		font-weight: 500;
		color: #b83a2e;
	}
</style>
