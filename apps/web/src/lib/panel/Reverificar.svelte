<script lang="ts">
	import Modal from '#lib/components/Modal.svelte';
	import { reverificacion } from '#lib/panel/reverificar.svelte.ts';

	/**
	 * «Confirmá que sos vos» antes de cambiar adónde llega la plata. Con contraseña, se pide ésa;
	 * sin contraseña (Google, Apple, link de WhatsApp), el código de la app autenticadora si la tiene
	 * o uno que llega al mail. Se monta una vez en el layout del panel.
	 */
	type Metodos = { clave: boolean; totp: boolean; email: string };
	let metodos = $state<Metodos | null>(null);
	let modo = $state<'clave' | 'codigo'>('clave');
	let valor = $state('');
	let error = $state('');
	let enviando = $state(false);
	let mandado = $state(false);

	$effect(() => {
		if (!reverificacion.abierta) return;
		valor = '';
		error = '';
		mandado = false;
		metodos = null;
		void fetch('/panel/cuenta/reverificar')
			.then((r) => (r.ok ? (r.json() as Promise<Metodos>) : null))
			.catch(() => null)
			.then((m) => {
				metodos = m ?? { clave: false, totp: false, email: '' };
				modo = metodos.clave ? 'clave' : 'codigo';
				if (!metodos.clave && !metodos.totp) void mandarCodigo();
			});
	});

	async function mandarCodigo() {
		error = '';
		const r = await fetch('/panel/cuenta/reverificar', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ accion: 'codigo' })
		}).catch(() => null);
		if (r?.ok) mandado = true;
		else error = 'No pudimos mandarte el código. Esperá un momento y probá de nuevo.';
	}

	async function confirmar(e: SubmitEvent) {
		e.preventDefault();
		if (!valor.trim()) {
			error = modo === 'clave' ? 'Escribí tu contraseña.' : 'Escribí el código.';
			return;
		}
		enviando = true;
		error = '';
		const r = await fetch('/panel/cuenta/reverificar', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(modo === 'clave' ? { clave: valor } : { codigo: valor.trim() })
		}).catch(() => null);
		enviando = false;
		const p = (await r?.json().catch(() => null)) as { token?: string; message?: string } | null;
		if (r?.ok && p?.token) {
			reverificacion.terminar(p.token);
			return;
		}
		error = p?.message ?? 'No pudimos verificarte. Probá de nuevo.';
	}
</script>

<Modal
	bind:open={() => reverificacion.abierta, (v) => !v && reverificacion.terminar(null)}
	titulo="Confirmá que sos vos"
	ancho={420}
>
	<form class="rv" onsubmit={confirmar} novalidate>
		<div class="rv-t">Confirmá que sos vos</div>
		<p class="rv-p">
			Vas a cambiar adónde te llega la plata. Te avisamos por mail con cada cambio.
		</p>
		{#if !metodos}
			<div class="rv-carga" aria-hidden="true"></div>
		{:else}
			{#if modo === 'clave'}
				<label class="rv-l" for="rv-in">Tu contraseña</label>
				<input
					id="rv-in"
					class="input"
					type="password"
					autocomplete="current-password"
					bind:value={valor}
					aria-invalid={!!error}
					aria-describedby={error ? 'rv-err' : undefined}
				/>
			{:else}
				<label class="rv-l" for="rv-in">
					{mandado
						? `El código que te mandamos a ${metodos.email}`
						: metodos.totp
							? 'El código de tu app autenticadora'
							: 'El código que te mandamos por mail'}
				</label>
				<input
					id="rv-in"
					class="input rv-cod"
					inputmode="numeric"
					autocomplete="one-time-code"
					maxlength="12"
					placeholder="123 456"
					bind:value={valor}
					aria-invalid={!!error}
					aria-describedby={error ? 'rv-err' : undefined}
				/>
			{/if}
			{#if error}<p id="rv-err" class="rv-err" role="alert">{error}</p>{/if}
			<div class="rv-otro">
				{#if modo === 'clave'}
					<button
						type="button"
						class="rv-link"
						onclick={() => ((modo = 'codigo'), (valor = ''), void mandarCodigo())}
						>Mejor mandame un código por mail</button
					>
				{:else if !mandado}
					<button type="button" class="rv-link" onclick={mandarCodigo}
						>Mandame un código por mail</button
					>
				{:else}
					<button type="button" class="rv-link" onclick={mandarCodigo}>Mandármelo de nuevo</button>
				{/if}
			</div>
			<div class="rv-acc">
				<button
					type="button"
					class="btn btn-secondary"
					onclick={() => reverificacion.terminar(null)}>Cancelar</button
				>
				<button type="submit" class="btn btn-primary" disabled={enviando}
					>{enviando ? 'Verificando…' : 'Confirmar'}</button
				>
			</div>
		{/if}
	</form>
</Modal>

<style>
	.rv {
		display: grid;
		gap: 10px;
	}
	.rv-t {
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
		color: var(--tx);
	}
	.rv-p {
		margin: 0 0 6px;
		color: var(--mu);
		font-size: 14.5px;
		line-height: 1.45;
		text-wrap: balance;
	}
	.rv-l {
		font-size: 13.5px;
		font-weight: 600;
		color: var(--tx);
	}
	.rv-cod {
		letter-spacing: 0.12em;
		font-variant-numeric: tabular-nums;
	}
	.rv-err {
		margin: 0;
		color: #b0441d;
		font-size: 13.5px;
	}
	.rv-otro {
		min-height: 22px;
	}
	.rv-link {
		padding: 0;
		border: 0;
		background: none;
		color: var(--tx);
		font: inherit;
		font-size: 13.5px;
		text-decoration: underline;
		text-underline-offset: 3px;
		cursor: pointer;
		min-height: 44px;
	}
	.rv-acc {
		display: flex;
		justify-content: flex-end;
		gap: 10px;
		margin-top: 4px;
	}
	.rv-carga {
		height: 120px;
		border-radius: 14px;
		background: var(--chip);
	}
</style>
