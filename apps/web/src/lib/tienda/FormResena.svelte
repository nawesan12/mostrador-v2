<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { enviarOpinion, varsTema, type TemaOpiniones } from './opiniones.ts';

	/**
	 * "¿Cómo te fue? Dejá tu reseña": estrellas de 1 a 5 (radios: se eligen con las flechas), nombre y
	 * texto. Con el código del pedido, la reseña queda con compra. La plantilla le pasa su tema.
	 */
	let {
		slug,
		tema,
		codigo = undefined,
		productoId = undefined,
		nombre: inicial = '',
		titulo = '¿Cómo te fue? Dejá tu reseña',
		bajada = 'Tu opinión ayuda a quien compra después. Se publica con tu nombre.'
	}: {
		slug: string;
		tema: TemaOpiniones;
		codigo?: string;
		productoId?: string;
		nombre?: string;
		titulo?: string;
		bajada?: string;
	} = $props();

	const uid = $props.id();
	let estrellas = $state(0);
	let sobre = $state(0);
	// Arranca con el nombre del pedido; quien escribe lo puede cambiar.
	let autor = $derived(inicial);
	let texto = $state('');
	let enviando = $state(false);
	let error = $state<string | null>(null);
	let hecho = $state(false);
	const ETIQUETAS = ['Muy mal', 'Mal', 'Más o menos', 'Bien', '¡Excelente!'];

	async function enviar(e: SubmitEvent) {
		e.preventDefault();
		if (!estrellas) return (error = 'Elegí cuántas estrellas.');
		if (autor.trim().length < 2) return (error = 'Poné tu nombre.');
		enviando = true;
		error = await enviarOpinion(slug, 'resena', {
			codigo,
			productoId,
			autor: autor.trim(),
			estrellas,
			texto: texto.trim()
		});
		enviando = false;
		hecho = !error;
	}
</script>

<section class="op" style={varsTema(tema)} aria-labelledby="{uid}-t">
	{#if hecho}
		<div class="gracias" role="status">
			<span class="ok"><Icon icon={I.phCheckBold} size={20} /></span>
			<div>
				<h2 id="{uid}-t">¡Gracias por tu reseña!</h2>
				<p>Ya se ve en la tienda. Nos ayuda un montón.</p>
			</div>
		</div>
	{:else}
		<h2 id="{uid}-t">{titulo}</h2>
		{#if bajada}<p class="bajada">{bajada}</p>{/if}
		<form onsubmit={enviar} novalidate>
			<fieldset class="estrellas">
				<legend>¿Cuántas estrellas le das?</legend>
				<div class="fila" role="presentation" onmouseleave={() => (sobre = 0)}>
					{#each [1, 2, 3, 4, 5] as n (n)}
						<input
							type="radio"
							id="{uid}-e{n}"
							name="{uid}-estrellas"
							value={n}
							bind:group={estrellas}
						/><label
							for="{uid}-e{n}"
							class:on={(sobre || estrellas) >= n}
							onmouseenter={() => (sobre = n)}
							><Icon icon={(sobre || estrellas) >= n ? I.phStarFill : I.phStar} size={28} /><span
								class="sr">{n === 1 ? '1 estrella' : `${n} estrellas`}</span
							></label
						>
					{/each}
					<span class="que" aria-hidden="true">{ETIQUETAS[(sobre || estrellas) - 1] ?? ''}</span>
				</div>
			</fieldset>
			<label class="campo"
				><span>Tu nombre</span><input
					bind:value={autor}
					autocomplete="given-name"
					maxlength="60"
					required
				/></label
			>
			<label class="campo"
				><span>Contanos <i>(opcional)</i></span><textarea
					bind:value={texto}
					maxlength="1000"
					rows="3"
					placeholder="¿Qué te pareció? ¿Llegó bien?"></textarea></label
			>
			{#if error}<p class="error" role="alert">{error}</p>{/if}
			<button type="submit" disabled={enviando}
				>{enviando ? 'Publicando…' : 'Publicar mi reseña'}</button
			>
		</form>
	{/if}
</section>

<style>
	.op {
		box-sizing: border-box;
		width: 100%;
		max-width: 560px;
		margin: 0 auto;
		padding: clamp(18px, 4vw, 28px);
		background: var(--op-sf);
		border: 1px solid var(--op-ln);
		border-radius: var(--op-r);
		color: var(--op-tx);
		font-family: var(--op-f);
		text-align: left;
	}
	.op * {
		box-sizing: border-box;
	}
	h2 {
		margin: 0;
		font-family: var(--op-tf);
		font-weight: var(--op-tw);
		text-transform: var(--op-tt);
		letter-spacing: var(--op-ls);
		font-size: clamp(22px, 5vw, 28px);
		line-height: 1.1;
		text-wrap: balance;
	}
	p {
		text-wrap: pretty;
	}
	.bajada {
		margin: 8px 0 0;
		font-size: 14px;
		line-height: 1.5;
		color: var(--op-mu);
	}
	form {
		margin-top: 18px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	fieldset {
		margin: 0;
		padding: 0;
		border: 0;
		min-width: 0;
	}
	legend {
		padding: 0;
		font-size: 13.5px;
		font-weight: 600;
	}
	.fila {
		margin-top: 4px;
		display: flex;
		align-items: center;
		flex-wrap: wrap;
	}
	.fila input {
		position: absolute;
		opacity: 0;
		width: 1px;
		height: 1px;
		pointer-events: none;
	}
	.fila label {
		display: grid;
		place-items: center;
		width: 44px;
		height: 44px;
		color: var(--op-mu);
		cursor: pointer;
		border-radius: var(--op-r);
	}
	.fila label.on {
		color: var(--op-st);
	}
	.fila input:focus-visible + label {
		outline: 2px solid var(--op-ac);
		outline-offset: -2px;
	}
	.que {
		margin-left: 6px;
		font-size: 13.5px;
		color: var(--op-mu);
	}
	.sr {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
	}
	.campo i {
		font-style: normal;
		font-weight: 400;
		color: var(--op-mu);
	}
	input:not([type='radio']),
	textarea {
		width: 100%;
		min-height: 44px;
		padding: 11px 12px;
		border: 1px solid var(--op-ln);
		border-radius: var(--op-r);
		background: var(--op-campo);
		color: var(--op-tx);
		font: 400 15px var(--op-f);
		resize: vertical;
	}
	input:not([type='radio']):focus-visible,
	textarea:focus-visible {
		outline: 2px solid var(--op-ac);
		outline-offset: 1px;
	}
	button {
		align-self: flex-start;
		min-height: 46px;
		padding: 12px 22px;
		border: 0;
		border-radius: var(--op-rb);
		background: var(--op-ac);
		color: var(--op-on);
		font: 600 14.5px var(--op-f);
		cursor: pointer;
	}
	button:disabled {
		opacity: 0.6;
		cursor: default;
	}
	.error {
		margin: 0;
		font-size: 13.5px;
		color: var(--op-err);
	}
	.gracias {
		display: flex;
		gap: 14px;
		align-items: flex-start;
	}
	.gracias p {
		margin: 6px 0 0;
		font-size: 14.5px;
		color: var(--op-mu);
	}
	.ok {
		flex: none;
		display: grid;
		place-items: center;
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: var(--op-ac);
		color: var(--op-on);
	}
	@media (max-width: 420px) {
		button {
			align-self: stretch;
		}
	}
</style>
