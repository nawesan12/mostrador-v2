<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { enviarOpinion, varsTema, type TemaOpiniones } from './opiniones.ts';
	import type { PreguntaPublica } from './tipos.ts';

	/**
	 * Las preguntas respondidas de un producto y "Preguntale a la tienda" (nombre, email opcional y
	 * la pregunta). Se ve en la ficha cuando la tienda la responde desde el panel.
	 */
	let {
		slug,
		productoId,
		preguntas,
		tema,
		titulo = 'Preguntas'
	}: {
		slug: string;
		productoId: string;
		preguntas: PreguntaPublica[];
		tema: TemaOpiniones;
		titulo?: string;
	} = $props();

	const uid = $props.id();
	const delProducto = $derived(preguntas.filter((q) => q.productoId === productoId));
	let abierto = $state(false);
	let nombre = $state('');
	let email = $state('');
	let pregunta = $state('');
	let enviando = $state(false);
	let error = $state<string | null>(null);
	let hecho = $state<'con-email' | 'sin-email' | null>(null);

	async function enviar(e: SubmitEvent) {
		e.preventDefault();
		if (nombre.trim().length < 2) return (error = 'Poné tu nombre.');
		if (pregunta.trim().length < 3) return (error = 'Escribí tu pregunta.');
		enviando = true;
		error = await enviarOpinion(slug, 'pregunta', {
			productoId,
			nombre: nombre.trim(),
			email: email.trim(),
			pregunta: pregunta.trim()
		});
		enviando = false;
		if (error) return;
		hecho = email.trim() ? 'con-email' : 'sin-email';
		pregunta = '';
		abierto = false;
	}
</script>

<section class="op" style={varsTema(tema)} aria-labelledby="{uid}-t">
	<h2 id="{uid}-t">{titulo}</h2>
	{#if delProducto.length}
		<ul class="lista">
			{#each delProducto as q, i (i)}
				<li>
					<p class="q"><Icon icon={I.phQuestion} size={16} />{q.pregunta}</p>
					<p class="a"><Icon icon={I.phStorefront} size={16} />{q.respuesta}</p>
				</li>
			{/each}
		</ul>
	{:else}
		<p class="vacio">Nadie preguntó todavía. Si tenés una duda, la tienda te responde acá.</p>
	{/if}

	{#if hecho}
		<p class="listo" role="status">
			<Icon icon={I.phCheckBold} size={16} />¡Listo, la recibimos! {hecho === 'con-email'
				? 'Te avisamos cuando responda.'
				: 'Cuando responda, la ves acá.'}
		</p>
	{/if}
	{#if abierto}
		<form onsubmit={enviar} novalidate>
			<label class="campo"
				><span>Tu pregunta</span><textarea
					bind:value={pregunta}
					maxlength="500"
					rows="3"
					placeholder="¿Tiene garantía? ¿Hay en otro color?"></textarea></label
			>
			<div class="dos">
				<label class="campo"
					><span>Tu nombre</span><input
						bind:value={nombre}
						autocomplete="given-name"
						maxlength="60"
					/></label
				>
				<label class="campo"
					><span>Email <i>(opcional)</i></span><input
						bind:value={email}
						type="email"
						autocomplete="email"
						maxlength="120"
					/></label
				>
			</div>
			<p class="nota">Con tu email, te avisamos cuando responda.</p>
			{#if error}<p class="error" role="alert">{error}</p>{/if}
			<div class="acc">
				<button type="submit" class="pri" disabled={enviando}
					>{enviando ? 'Mandando…' : 'Mandar pregunta'}</button
				><button type="button" class="sec" onclick={() => (abierto = false)}>Cancelar</button>
			</div>
		</form>
	{:else}
		<button
			type="button"
			class="sec abrir"
			aria-expanded="false"
			onclick={() => {
				abierto = true;
				hecho = null;
				error = null;
			}}><Icon icon={I.phChatCircleDotsFill} size={17} />Preguntale a la tienda</button
		>
	{/if}
</section>

<style>
	.op {
		box-sizing: border-box;
		width: 100%;
		padding: var(--op-pad, 0);
		background: var(--op-sf);
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
		font-size: clamp(22px, 4vw, 30px);
		line-height: 1.1;
	}
	p {
		text-wrap: pretty;
	}
	.lista {
		margin: 16px 0 0;
		padding: 0;
		list-style: none;
		border-top: 1px solid var(--op-ln);
	}
	.lista li {
		padding: 14px 0;
		border-bottom: 1px solid var(--op-ln);
	}
	.q,
	.a {
		margin: 0;
		display: flex;
		gap: 8px;
		align-items: baseline;
		font-size: 15px;
		line-height: 1.5;
	}
	.q {
		font-weight: 600;
	}
	.a {
		margin-top: 6px;
		color: var(--op-mu);
	}
	.q :global(svg),
	.a :global(svg),
	.listo :global(svg) {
		flex: none;
		transform: translateY(2px);
	}
	.vacio {
		margin: 10px 0 0;
		font-size: 14.5px;
		line-height: 1.5;
		color: var(--op-mu);
	}
	.listo {
		margin: 14px 0 0;
		display: flex;
		gap: 8px;
		align-items: baseline;
		font-size: 14.5px;
		font-weight: 600;
	}
	form {
		margin-top: 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 200px), 1fr));
		gap: 12px;
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
	input,
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
	input:focus-visible,
	textarea:focus-visible {
		outline: 2px solid var(--op-ac);
		outline-offset: 1px;
	}
	.nota {
		margin: -4px 0 0;
		font-size: 12.5px;
		color: var(--op-mu);
	}
	.acc {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	button {
		min-height: 46px;
		padding: 12px 20px;
		border-radius: var(--op-rb);
		font: 600 14.5px var(--op-f);
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
	}
	.pri {
		border: 0;
		background: var(--op-ac);
		color: var(--op-on);
	}
	.pri:disabled {
		opacity: 0.6;
		cursor: default;
	}
	.sec {
		border: 1px solid var(--op-tx);
		background: transparent;
		color: var(--op-tx);
	}
	.abrir {
		margin-top: 16px;
	}
	.error {
		margin: 0;
		font-size: 13.5px;
		color: var(--op-err);
	}
	button {
		white-space: nowrap;
	}
	@media (max-width: 420px) {
		.acc {
			flex-direction: column;
		}
		.acc button,
		.abrir {
			flex: 1;
			width: 100%;
		}
	}
</style>
