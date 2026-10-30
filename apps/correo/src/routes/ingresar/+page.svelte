<script lang="ts">
	import { enhance } from '$app/forms';
	import Marca from '#lib/ui/Marca.svelte';

	let { data, form } = $props();
	let enviando = $state(false);
	const m = $derived(data.marca);
</script>

<svelte:head><title>Correo · {m}</title></svelte:head>

<main>
	<div class="marca">
		<Marca marca={m} />
		<div class="tx">
			<div class="t">Correo</div>
			<div class="m">{m}</div>
		</div>
	</div>
	<form
		class="panel vidrio"
		method="POST"
		use:enhance={() => {
			enviando = true;
			return async ({ update }) => {
				await update({ reset: false });
				enviando = false;
			};
		}}
	>
		<h1>Entrá a tu casilla</h1>
		<label>
			<span>Casilla</span>
			<span class="campo vidrio">
				<i class="icon-at-sign"></i>
				<input
					name="email"
					type="email"
					autocomplete="username"
					placeholder="nombre@grupovimo.com.ar"
					value={form?.email ?? ''}
					required
				/>
			</span>
		</label>
		<label>
			<span>Clave</span>
			<span class="campo vidrio">
				<i class="icon-lock"></i>
				<input name="clave" type="password" autocomplete="current-password" required />
			</span>
		</label>
		{#if form?.error}<p class="error" role="alert">{form.error}</p>{/if}
		<button class="entrar" disabled={enviando}>
			<i class="icon-log-in"></i>{enviando ? 'Entrando…' : 'Entrar'}
		</button>
	</form>
</main>

<style>
	main {
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 28px;
		padding: calc(24px + env(safe-area-inset-top)) 16px calc(24px + env(safe-area-inset-bottom));
	}
	.marca {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.tx {
		line-height: 1.2;
	}
	.t {
		font-size: 19px;
		font-weight: 600;
	}
	.m {
		font-size: 14px;
		color: var(--mut);
	}
	.panel {
		width: min(100%, 440px);
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 28px;
		border-radius: 16px;
		background: var(--panel);
		box-shadow: var(--panelSh);
	}
	h1 {
		margin: 0 0 4px;
		font-size: 26px;
		font-weight: 600;
		line-height: 1.3;
		letter-spacing: -0.01em;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 15px;
		color: var(--mut);
	}
	.campo {
		height: 50px;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 0 16px;
		border-radius: 12px;
		background: var(--srch);
		color: var(--mut);
		font-size: 17px;
	}
	.campo:focus-within {
		box-shadow: 0 0 0 2px var(--acc);
	}
	.campo i {
		font-size: 20px;
	}
	input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		color: var(--fg);
		font-size: 17px;
	}
	input::placeholder {
		color: var(--t3);
	}
	.error {
		margin: 0;
		padding: 10px 14px;
		border-radius: 10px;
		background: var(--prBg);
		color: var(--prT);
		font-size: 15px;
	}
	.entrar {
		margin-top: 6px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		height: 58px;
		border-radius: 16px;
		background: var(--acc);
		color: #fff;
		font-size: 18px;
		font-weight: 600;
		box-shadow:
			0 2px 0 oklch(0.45 calc(0.17 * var(--ac)) calc(var(--h) - 6)),
			0 8px 20px -8px oklch(0.45 calc(0.17 * var(--ac)) var(--h) / 0.55);
	}
	.entrar i {
		font-size: 21px;
	}
	.entrar:disabled {
		opacity: 0.7;
	}
</style>
