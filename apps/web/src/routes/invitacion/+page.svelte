<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import { ErrorClerk, iniciar, type ProveedorClerk } from '#lib/clerk.ts';
	import { logosApple, logosGoogleIcon, phInfoFill, phUsersThreeFill } from '#lib/icons.ts';
	import { NOMBRE_ROL, QUE_HACE } from '#lib/panel/permisos.ts';

	/**
	 * Aceptar una invitación al equipo de una tienda (docs/decisiones.md §37). Misma estética que
	 * /ingresar. Con sesión: «Sumarme». Si el email ya tiene cuenta: ingresar y volver acá. Si no:
	 * crear la cuenta (con clave o con Google/Apple), sin tienda propia.
	 */
	let { data, form } = $props();
	let enviando = $state(false);
	let errSocial = $state('');
	const inv = $derived(data.invitacion);
	const volver = $derived(
		data.token
			? `${resolve('/ingresar')}?volver=${encodeURIComponent(`/invitacion?token=${data.token}`)}`
			: ''
	);
	const vence = $derived(
		inv
			? new Date(inv.vence).toLocaleDateString('es-AR', {
					day: 'numeric',
					month: 'long',
					timeZone: 'America/Argentina/Buenos_Aires'
				})
			: ''
	);

	async function seguirCon(proveedor: ProveedorClerk) {
		if (!data.clerk) return;
		enviando = true;
		errSocial = '';
		try {
			location.assign(
				await iniciar(data.clerk, proveedor, `${location.origin}${resolve('/ingresar/clerk')}`)
			);
		} catch (x) {
			enviando = false;
			errSocial =
				x instanceof ErrorClerk && x.codigo === 'too_many_requests'
					? 'Hubo demasiados intentos. Probá de nuevo en unos minutos.'
					: 'No pudimos conectar con Google o Apple. Probá de nuevo o creá la cuenta con una contraseña.';
		}
	}
	const conEnvio = () => {
		enviando = true;
		return async ({ update }: { update: () => Promise<void> }) => {
			await update();
			enviando = false;
		};
	};
</script>

<Seo
	title="Invitación · Tu Local Web"
	description="Sumate al equipo de una tienda en Tu Local Web."
	path="/invitacion"
	noindex
/>

<div class="pag">
	<div class="arriba">
		<a href={resolve('/')}><img src="/wordmark-brown.webp" alt="Tu Local Web" class="marca" /></a>
	</div>
	<div class="centro">
		<div class="caja">
			{#if !inv}
				<div class="col">
					<h1>Este link ya no sirve</h1>
					<p class="bajada">
						La invitación venció, se canceló o ya se usó. Pedile a la dueña de la tienda que te la
						reenvíe desde Usuarios y permisos.
					</p>
					<a class="cta" href={resolve('/ingresar')}>Ir a ingresar</a>
				</div>
			{:else}
				<div class="col">
					<span class="icono"><Icon icon={phUsersThreeFill} size={22} /></span>
					<div>
						<h1>Sumate a {inv.tienda}</h1>
						<p class="bajada">
							{inv.invita ? `${inv.invita} te invitó` : 'Te invitaron'} al equipo con el rol
							<b>{NOMBRE_ROL[inv.rol]}</b>: {QUE_HACE[inv.rol].toLowerCase()}.
						</p>
					</div>
					{#if form?.error}<div class="err-form" role="alert">{form.error}</div>{/if}

					{#if data.sesion && data.sesion.email.toLowerCase() !== inv.email.toLowerCase()}
						<!-- Sólo la cuenta del email invitado puede sumarse (Go lo exige). -->
						<div class="nota">
							<Icon icon={phInfoFill} size={18} style="flex:none;margin-top:1px" />
							<span
								>Entraste como <b>{data.sesion.email}</b>, pero la invitación es para
								<b>{inv.email}</b>. Salí y entrá con esa&nbsp;cuenta.</span
							>
						</div>
						<form method="POST" action="?/salir">
							<input type="hidden" name="token" value={data.token} />
							<button class="cta">Salir y entrar con {inv.email}</button>
						</form>
					{:else if data.sesion}
						<div class="nota">
							<Icon icon={phInfoFill} size={18} style="flex:none;margin-top:1px" />
							<span>Entraste como <b>{data.sesion.email}</b>.</span>
						</div>
						<form method="POST" action="?/aceptar" use:enhance={conEnvio}>
							<input type="hidden" name="token" value={data.token} />
							<button class="cta" disabled={enviando}
								>{enviando ? 'Sumándote…' : `Sumarme a ${inv.tienda}`}</button
							>
						</form>
					{:else if inv.tieneCuenta}
						<p class="bajada">
							<b>{inv.email}</b> ya tiene cuenta en Tu Local Web. Ingresá con ella y te sumamos a la tienda.
						</p>
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- /ingresar resuelto arriba, con la vuelta -->
						<a class="cta" href={volver}>Ingresar para sumarme</a>
					{:else}
						{#if data.clerk}
							<div class="sociales">
								<button
									type="button"
									class="social"
									disabled={enviando}
									onclick={() => seguirCon('google')}
									><Icon icon={logosGoogleIcon} size={18} />Seguir con Google</button
								>
								<button
									type="button"
									class="social"
									disabled={enviando}
									onclick={() => seguirCon('apple')}
									><Icon icon={logosApple} size={18} />Seguir con Apple</button
								>
							</div>
							{#if errSocial}<div class="err-form" role="alert">{errSocial}</div>{/if}
							<div class="o"><span></span>o con una contraseña<span></span></div>
						{/if}
						<form method="POST" action="?/crear" class="col" use:enhance={conEnvio}>
							<input type="hidden" name="token" value={data.token} />
							<label class="campo"
								>Email<span class="in"
									><input value={inv.email} readonly aria-readonly="true" /></span
								></label
							>
							<label class="campo"
								>Tu nombre<span class="in"
									><input
										name="nombre"
										value={form?.nombre ?? ''}
										autocomplete="name"
										maxlength="80"
										required
									/></span
								></label
							>
							<label class="campo"
								>Contraseña<span class="in"
									><input
										name="clave"
										type="password"
										autocomplete="new-password"
										minlength="8"
										maxlength="200"
										required
									/></span
								></label
							>
							<button class="cta" disabled={enviando}
								>{enviando ? 'Creando tu cuenta…' : 'Crear mi cuenta y sumarme'}</button
							>
						</form>
					{/if}
					<p class="pie">El link sirve una sola vez y vence el {vence}.</p>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.pag {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		padding: clamp(20px, 3vw, 36px) clamp(16px, 5vw, 72px);
		color: #2a140c;
		background: #fbf7f2;
	}
	.marca {
		height: 21px;
		width: auto;
		display: block;
	}
	.centro {
		flex: 1;
		display: flex;
		align-items: center;
		padding: clamp(28px, 5vw, 56px) 0;
	}
	.caja {
		width: 100%;
		max-width: 440px;
		margin: 0 auto;
	}
	.col {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}
	.icono {
		width: 48px;
		height: 48px;
		border-radius: 16px;
		display: grid;
		place-items: center;
		background: #f4ece2;
		color: #b0441d;
	}
	h1 {
		margin: 0;
		font-size: clamp(30px, 3.6vw, 40px);
		line-height: 1.04;
		letter-spacing: -0.04em;
		font-weight: 600;
		text-wrap: balance;
	}
	.bajada {
		margin: 8px 0 0;
		font-size: 15.5px;
		line-height: 1.5;
		color: #5c4536;
		text-wrap: pretty;
		overflow-wrap: anywhere;
	}
	.bajada b {
		color: #2a140c;
	}
	.nota {
		padding: 14px 16px;
		border-radius: 16px;
		background: #f4ece2;
		font-size: 14px;
		line-height: 1.5;
		color: #5c4536;
		display: flex;
		gap: 10px;
		overflow-wrap: anywhere;
	}
	.err-form {
		padding: 12px 14px;
		border-radius: 14px;
		background: rgba(184, 58, 46, 0.08);
		color: #8f2a20;
		font-size: 14px;
		line-height: 1.45;
	}
	.sociales {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
		gap: 8px;
	}
	.social {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 9px;
		min-height: 48px;
		padding: 13px;
		border-radius: 999px;
		border: 1.5px solid rgba(58, 29, 18, 0.14);
		background: #fff;
		font: inherit;
		font-size: 14.5px;
		font-weight: 600;
		color: #2a140c;
		white-space: nowrap;
	}
	.o {
		display: flex;
		align-items: center;
		gap: 12px;
		font-size: 13px;
		color: #7a6252;
	}
	.o span {
		flex: 1;
		height: 1px;
		background: rgba(58, 29, 18, 0.12);
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
	}
	.in {
		display: flex;
	}
	.in input {
		flex: 1;
		min-width: 0;
		padding: 14px 16px;
		border-radius: 14px;
		border: 1.5px solid rgba(58, 29, 18, 0.14);
		background: #fff;
		font: inherit;
		font-size: 15.5px;
		font-weight: 500;
		color: #2a140c;
		outline: none;
	}
	.in input:focus {
		border-color: #d2582f;
	}
	.in input[readonly] {
		background: #f4ece2;
		color: #5c4536;
	}
	.cta {
		text-decoration: none;
		display: flex;
		justify-content: center;
		align-items: center;
		width: 100%;
		min-height: 52px;
		padding: 16px;
		border: 0;
		border-radius: 999px;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font: inherit;
		font-size: 15.5px;
		font-weight: 600;
		text-align: center;
	}
	.cta:disabled {
		opacity: 0.7;
		cursor: progress;
	}
	.pie {
		margin: 0;
		text-align: center;
		font-size: 14px;
		color: #5c4536;
	}
</style>
