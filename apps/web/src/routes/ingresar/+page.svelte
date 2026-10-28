<script lang="ts">
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { onDestroy } from 'svelte';
	import CookieBanner from '#lib/components/CookieBanner.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import { fuerza } from '#lib/fuerza.ts';
	import {
		logosGoogleIcon,
		logosWhatsappIcon,
		phArrowLeftBold,
		phCheckBold,
		phCheckCircleFill,
		phEye,
		phEyeSlash,
		phInfoFill,
		phLockSimpleFill,
		phMapPinFill,
		phQuestion,
		phShieldCheckFill,
		phStarFill
	} from '#lib/icons.ts';

	/* Recreación de design/Mostrador Ingresar.dc.html: modos login, registro, recuperar y enviado. */
	let { form } = $props();

	type Modo =
		'login' | 'registro' | 'recuperar' | 'enviado' | 'restablecer' | 'whatsapp' | 'enlace';
	const MODOS: Modo[] = [
		'login',
		'registro',
		'recuperar',
		'enviado',
		'restablecer',
		'whatsapp',
		'enlace'
	];
	// Link del mail de recuperación: ?m=restablecer&token=…
	const tokenUrl = $derived(page.url.searchParams.get('token') ?? '');
	const listo = $derived(page.url.searchParams.get('listo') === 'clave');
	const deUrl = $derived.by(() => {
		const m = page.url.searchParams.get('m') as Modo | null;
		return m && MODOS.includes(m) && m !== 'enviado' ? m : 'login';
	});
	const modo = $derived<Modo>((form?.modo as Modo | undefined) ?? deUrl);
	// Errores que vuelven en la URL desde la vuelta de un proveedor (/ingresar/<proveedor>/vuelta).
	const ERRORES_URL: Record<string, string> = {
		cancelado: 'Cancelaste el ingreso. Probá de nuevo o entrá con tu email.',
		estado_invalido: 'El ingreso se cortó o tardó demasiado. Probá de nuevo.',
		sin_email: 'No pudimos saber tu email desde esa cuenta. Probá con tu email.',
		vincular:
			'Ya hay una cuenta con ese email. Ingresá con tu contraseña y conectá esa cuenta desde Mis datos.',
		proveedor_apagado: 'Ese ingreso todavía no está disponible. Probá con tu email.'
	};
	const errorUrl = $derived.by(() => {
		const c = page.url.searchParams.get('error');
		return c ? (ERRORES_URL[c] ?? 'No pudimos completar el ingreso. Probá de nuevo.') : '';
	});
	const err = $derived(
		(form && 'errores' in form
			? form.errores
			: errorUrl && !form
				? { form: errorUrl }
				: {}) as Record<string, string>
	);

	let verClave = $state(false);
	let clave = $state('');
	const f = $derived(fuerza(clave));
	let canal = $state<'email' | 'whatsapp'>('email');
	let enviando = $state(false);

	// "Reenviar en 0:45": cuenta regresiva desde que se mandó el link.
	let resta = $state(45);
	let reloj: ReturnType<typeof setInterval> | undefined;
	$effect(() => {
		if (modo !== 'enviado') return;
		resta = 45;
		clearInterval(reloj);
		reloj = setInterval(() => (resta = Math.max(0, resta - 1)), 1000);
		return () => clearInterval(reloj);
	});
	onDestroy(() => clearInterval(reloj));
	/** Link a otro modo de esta misma pantalla. */
	const aModo = (m: Modo) => `${resolve('/ingresar')}?m=${m}`;
	const mmss = (s: number) => `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;

	const conEnvio = () => {
		enviando = true;
		return async ({ update }: { update: () => Promise<void> }) => {
			await update();
			enviando = false;
		};
	};
	const barra = (i: number) =>
		i < f.nivel
			? i < 2
				? '#D2582F'
				: f.nivel === 4
					? '#2F8F63'
					: '#E3A93B'
			: 'rgba(58,29,18,.12)';
</script>

<Seo
	title="Ingresar · Mostrador"
	description="Entrá a tu panel o creá tu tienda gratis."
	path="/ingresar"
	image="home"
/>

<div class="pag">
	<div class="der">
		<div class="arriba">
			<a href={resolve('/')}><img src="/wordmark-brown.png" alt="mostrador" class="marca" /></a>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- /ayuda llega con las páginas públicas -->
			<a class="ayuda" href="/ayuda"><Icon icon={phQuestion} size={16} />Ayuda</a>
		</div>

		<div class="centro">
			<div class="caja">
				{#if err.form}<div class="err-form" role="alert">{err.form}</div>{/if}

				{#if modo === 'login'}
					<form method="POST" action="?/ingresar" class="col" use:enhance={conEnvio} novalidate>
						<div>
							<h1>Hola de nuevo</h1>
							<p class="bajada">Entrá para ver cómo vienen tus ventas hoy.</p>
						</div>
						{#if listo}
							<div class="nota" role="status">
								<Icon
									icon={phCheckCircleFill}
									size={18}
									style="color:#237550;flex:none;margin-top:1px"
								/>
								Listo, tu contraseña cambió. Ya podés ingresar con la nueva.
							</div>
						{/if}
						<div class="sociales">
							<button
								type="submit"
								formaction="?/social"
								name="proveedor"
								value="google"
								class="social"><Icon icon={logosGoogleIcon} size={18} />Google</button
							>
							<button
								type="submit"
								formaction="?/social"
								name="proveedor"
								value="whatsapp"
								class="social"><Icon icon={logosWhatsappIcon} size={18} />WhatsApp</button
							>
						</div>
						<div class="o"><span></span>o con tu email<span></span></div>
						<label class="campo"
							>Email<span class="in"
								><input
									name="email"
									type="email"
									autocomplete="email"
									placeholder="lucia@laesquina.com.ar"
									defaultValue={form?.email ?? ''}
									aria-invalid={!!err.email}
									aria-describedby={err.email ? 'err-email' : undefined}
								/></span
							></label
						>{#if err.email}<small class="err" id="err-email">{err.email}</small>{/if}
						<label class="campo"
							>Contraseña<span class="in"
								><input
									name="clave"
									type={verClave ? 'text' : 'password'}
									autocomplete="current-password"
									placeholder="••••••••"
									class="con-ojo"
									aria-invalid={!!err.clave}
									aria-describedby={err.clave ? 'err-clave' : undefined}
								/><button
									type="button"
									class="ojo"
									aria-label={verClave ? 'Ocultar contraseña' : 'Mostrar contraseña'}
									onclick={() => (verClave = !verClave)}
									><Icon icon={verClave ? phEyeSlash : phEye} size={19} /></button
								></span
							></label
						>{#if err.clave}<small class="err" id="err-clave">{err.clave}</small>{/if}
						<div class="fila">
							<label class="check"
								><input type="checkbox" name="recordarme" checked /><span class="caja-check"
									><Icon icon={phCheckBold} size={11} /></span
								>Recordarme</label
							>
							<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- otro modo de esta misma pantalla (aModo usa resolve) -->
							<a class="link" href={aModo('recuperar')}>Olvidé mi contraseña</a>
						</div>
						<button type="submit" class="cta" disabled={enviando}>Ingresar</button>
						<div class="pie-form">
							<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- otro modo de esta misma pantalla (aModo usa resolve) -->
							¿No tenés cuenta? <a class="link" href={aModo('registro')}>Creá tu tienda gratis</a>
						</div>
					</form>
				{:else if modo === 'registro'}
					<form method="POST" action="?/registro" class="col reg" use:enhance={conEnvio} novalidate>
						<div>
							<h1>Creá tu tienda gratis</h1>
							<p class="bajada">14 días sin tarjeta. Después elegís el plan, o no.</p>
						</div>
						<div class="sociales">
							<button
								type="submit"
								formaction="?/social"
								name="proveedor"
								value="google"
								class="social"><Icon icon={logosGoogleIcon} size={18} />Google</button
							>
							<button
								type="submit"
								formaction="?/social"
								name="proveedor"
								value="whatsapp"
								class="social"><Icon icon={logosWhatsappIcon} size={18} />WhatsApp</button
							>
						</div>
						<div class="o"><span></span>o con tu email<span></span></div>
						<div class="dos">
							<label class="campo"
								>Tu nombre<span class="in"
									><input
										name="nombre"
										autocomplete="name"
										placeholder="Lucía Fernández"
										defaultValue={form?.nombre ?? ''}
										aria-invalid={!!err.nombre}
										aria-describedby={err.nombre ? 'err-nombre' : undefined}
									/></span
								></label
							>{#if err.nombre}<small class="err" id="err-nombre">{err.nombre}</small>{/if}
							<label class="campo"
								>Nombre de la tienda<span class="in"
									><input
										name="tienda"
										autocomplete="organization"
										placeholder="La Esquina"
										defaultValue={form?.tienda ?? ''}
										aria-invalid={!!err.tienda}
										aria-describedby={err.tienda ? 'err-tienda' : undefined}
									/></span
								></label
							>{#if err.tienda}<small class="err" id="err-tienda">{err.tienda}</small>{/if}
						</div>
						<label class="campo"
							>Email<span class="in"
								><input
									name="email"
									type="email"
									autocomplete="email"
									placeholder="lucia@laesquina.com.ar"
									defaultValue={form?.email ?? ''}
									aria-invalid={!!err.email}
									aria-describedby={err.email ? 'err-email' : undefined}
								/></span
							></label
						>{#if err.email}<small class="err" id="err-email">{err.email}</small>{/if}
						<label class="campo"
							>Contraseña<span class="in"
								><input
									name="clave"
									type={verClave ? 'text' : 'password'}
									autocomplete="new-password"
									placeholder="Mínimo 8 caracteres"
									class="con-ojo"
									bind:value={clave}
									aria-invalid={!!err.clave}
									aria-describedby={err.clave ? 'err-clave fuerza' : 'fuerza'}
								/><button
									type="button"
									class="ojo"
									aria-label={verClave ? 'Ocultar contraseña' : 'Mostrar contraseña'}
									onclick={() => (verClave = !verClave)}
									><Icon icon={verClave ? phEyeSlash : phEye} size={19} /></button
								></span
							></label
						>{#if err.clave}<small class="err" id="err-clave">{err.clave}</small>{/if}
						<div class="barras" aria-hidden="true">
							{#each [0, 1, 2, 3] as i (i)}<span style:background={barra(i)}></span>{/each}
						</div>
						<div class="fuerza" id="fuerza">{f.texto}</div>
						<label class="check terminos"
							><input
								type="checkbox"
								name="terminos"
								aria-invalid={!!err.terminos}
								aria-describedby={err.terminos ? 'err-terminos' : undefined}
							/><span class="caja-check"><Icon icon={phCheckBold} size={11} /></span><span
								>Acepto los
								<!-- eslint-disable svelte/no-navigation-without-resolve -- /legales llega con las páginas públicas -->
								<a href="/legales/terminos">Términos y condiciones</a> y la
								<a href="/legales/privacidad">Política de privacidad</a>.
								<!-- eslint-enable svelte/no-navigation-without-resolve -->
							</span></label
						>
						{#if err.terminos}<small class="err" id="err-terminos">{err.terminos}</small>{/if}
						<button type="submit" class="cta" disabled={enviando}>Crear mi tienda</button>
						<div class="pie-form">
							<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- otro modo de esta misma pantalla (aModo usa resolve) -->
							¿Ya tenés cuenta? <a class="link" href={aModo('login')}>Ingresá</a>
						</div>
					</form>
				{:else if modo === 'recuperar'}
					<form method="POST" action="?/recuperar" class="col" use:enhance={conEnvio} novalidate>
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- otro modo de esta misma pantalla (aModo usa resolve) -->
						<a class="volver" href={aModo('login')}
							><Icon icon={phArrowLeftBold} size={14} />Volver</a
						>
						<img src="/hornero/pensando.png" alt="" class="ave-chica" />
						<div>
							<h1>¿Te olvidaste la contraseña?</h1>
							<p class="bajada">
								Pasa. Te mandamos un link para crear una nueva, por email o por WhatsApp.
							</p>
						</div>
						<input type="hidden" name="canal" value={canal} />
						<div class="seg" role="group" aria-label="Canal">
							<button
								type="button"
								aria-pressed={canal === 'email'}
								onclick={() => (canal = 'email')}>Email</button
							>
							<button
								type="button"
								aria-pressed={canal === 'whatsapp'}
								onclick={() => (canal = 'whatsapp')}>WhatsApp</button
							>
						</div>
						{#if canal === 'email'}
							<label class="campo"
								>Email de tu cuenta<span class="in"
									><input
										name="email"
										type="email"
										autocomplete="email"
										placeholder="lucia@laesquina.com.ar"
										aria-invalid={!!err.email}
										aria-describedby={err.email ? 'err-email' : undefined}
									/></span
								></label
							>{#if err.email}<small class="err" id="err-email">{err.email}</small>{/if}
						{:else}
							<label class="campo"
								>WhatsApp de tu cuenta<span class="in"
									><input
										name="telefono"
										type="tel"
										autocomplete="tel"
										placeholder="11 2345 6789"
										aria-invalid={!!err.telefono}
										aria-describedby={err.telefono ? 'err-telefono' : undefined}
									/></span
								></label
							>{#if err.telefono}<small class="err" id="err-telefono">{err.telefono}</small>{/if}
						{/if}
						<button type="submit" class="cta" disabled={enviando}>Mandame el link</button>
					</form>
				{:else if modo === 'restablecer'}
					<!-- No está en el handoff: se arma con las piezas del modo "recuperar" y el medidor del registro. -->
					<form method="POST" action="?/restablecer" class="col" use:enhance={conEnvio} novalidate>
						<img src="/hornero/seguridad.png" alt="" class="ave-chica" />
						<div>
							<h1>Elegí tu nueva contraseña</h1>
							<p class="bajada">
								Reemplaza a la anterior y cerramos las sesiones abiertas en otros dispositivos.
							</p>
						</div>
						<input type="hidden" name="token" value={tokenUrl} />
						<label class="campo"
							>Contraseña nueva<span class="in"
								><input
									name="clave"
									type={verClave ? 'text' : 'password'}
									autocomplete="new-password"
									placeholder="Mínimo 8 caracteres"
									class="con-ojo"
									bind:value={clave}
									aria-invalid={!!err.clave}
									aria-describedby={err.clave ? 'err-clave fuerza' : 'fuerza'}
								/><button
									type="button"
									class="ojo"
									aria-label={verClave ? 'Ocultar contraseña' : 'Mostrar contraseña'}
									onclick={() => (verClave = !verClave)}
									><Icon icon={verClave ? phEyeSlash : phEye} size={19} /></button
								></span
							></label
						>{#if err.clave}<small class="err" id="err-clave">{err.clave}</small>{/if}
						<div class="barras" aria-hidden="true">
							{#each [0, 1, 2, 3] as i (i)}<span style:background={barra(i)}></span>{/each}
						</div>
						<div class="fuerza" id="fuerza">{f.texto}</div>
						{#if err.token}<small class="err">{err.token}</small>{/if}
						<button type="submit" class="cta" disabled={enviando}>Guardar contraseña</button>
					</form>
				{:else if modo === 'whatsapp'}
					<!-- No está en el handoff: las piezas del modo "recuperar" (canal WhatsApp). -->
					<form method="POST" action="?/enlace" class="col" use:enhance={conEnvio} novalidate>
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- otro modo de esta misma pantalla (aModo usa resolve) -->
						<a class="volver" href={aModo('login')}
							><Icon icon={phArrowLeftBold} size={14} />Volver</a
						>
						<img src="/hornero/whatsapp.png" alt="" class="ave-chica" />
						<div>
							<h1>Entrá con tu WhatsApp</h1>
							<p class="bajada">
								Te mandamos un link al WhatsApp de tu cuenta. Lo tocás y entrás, sin contraseña.
							</p>
						</div>
						<label class="campo"
							>WhatsApp de tu cuenta<span class="in"
								><input
									name="telefono"
									type="tel"
									autocomplete="tel"
									placeholder="11 2345 6789"
									aria-invalid={!!err.telefono}
									aria-describedby={err.telefono ? 'err-telefono' : undefined}
								/></span
							></label
						>{#if err.telefono}<small class="err" id="err-telefono">{err.telefono}</small>{/if}
						<button type="submit" class="cta" disabled={enviando}>Mandame el link</button>
						<p class="pie-nota">
							¿Todavía no cargaste tu WhatsApp? Entrá con tu email y sumalo desde Mis datos.
						</p>
					</form>
				{:else if modo === 'enlace'}
					<!-- Link del WhatsApp: pide un clic, así la vista previa del link no lo consume. -->
					<form method="POST" action="?/entrar" class="col" use:enhance={conEnvio}>
						<img src="/hornero/saludo.png" alt="" class="ave-chica" />
						<div>
							<h1>Entrá a tu cuenta</h1>
							<p class="bajada">Tocá el botón para entrar. El link vale una sola vez.</p>
						</div>
						<input type="hidden" name="token" value={tokenUrl} />
						<button type="submit" class="cta" disabled={enviando}>Entrar a mi panel</button>
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- otro modo de esta misma pantalla (aModo usa resolve) -->
						<a class="volver centro-l" href={aModo('whatsapp')}>Pedir otro link</a>
					</form>
				{:else}
					<form
						method="POST"
						action={form?.tipo === 'ingreso' ? '?/enlace' : '?/recuperar'}
						class="col enviado"
						use:enhance={conEnvio}
					>
						<img src="/hornero/exito.png" alt="" class="ave-exito" />
						<div>
							<h1>{form?.canal === 'whatsapp' ? 'Revisá tu WhatsApp' : 'Revisá tu email'}</h1>
							<p class="bajada">
								Te mandamos un link a <b>{form?.a}</b>. Vence en {form?.tipo === 'ingreso'
									? 15
									: 30} minutos.
							</p>
						</div>
						<div class="nota">
							<Icon icon={phInfoFill} size={18} style="color:#b0441d;flex:none;margin-top:1px" />
							{#if form?.canal === 'whatsapp'}
								¿No te llegó? Revisá que el número sea el de tu cuenta. Si seguís sin verlo, podemos
								mandártelo por email.
							{:else}
								¿No te llegó? Fijate en spam o promociones. Si seguís sin verlo, podemos mandártelo
								por WhatsApp.
							{/if}
						</div>
						<input type="hidden" name="canal" value={form?.canal ?? 'email'} />
						<input
							type="hidden"
							name={form?.canal === 'whatsapp' ? 'telefono' : 'email'}
							value={form?.a ?? ''}
						/>
						<div class="botones">
							<button type="submit" class="sec" disabled={resta > 0 || enviando}
								>{resta > 0 ? `Reenviar en ${mmss(resta)}` : 'Reenviar'}</button
							>
							<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- otro modo de esta misma pantalla (aModo usa resolve) -->
							<a class="oscuro" href={aModo('login')}>Volver a ingresar</a>
						</div>
					</form>
				{/if}
			</div>
		</div>

		<div class="legal">
			<span>© 2026 Mostrador</span>
			<!-- eslint-disable svelte/no-navigation-without-resolve -- /legales y /estado llegan con las páginas públicas -->
			<a href="/legales/terminos">Términos</a>
			<a href="/legales/privacidad">Privacidad</a>
			<a href="/estado"><span class="punto"></span>Todos los sistemas funcionan</a>
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
		</div>
	</div>
	<aside class="lado">
		<div class="brillo"></div>
		<div class="chips">
			<span>Tienda online</span><span>WhatsApp</span><span>Punto de venta</span><span
				>Facturación</span
			>
		</div>
		<div class="medio">
			<img class="ave" src="/hornero/saludo.png" alt="" />
			<div class="lema">Todo lo de las otras, por menos plata y más seguro.</div>
			<div class="testi">
				<div class="testi-cab">
					<img
						src="https://images.unsplash.com/photo-1562322140-8baeececf3df?auto=format&fit=crop&w=96&h=96&q=60&crop=faces"
						alt=""
						width="40"
						height="40"
					/>
					<div class="testi-quien">
						<b>Florencia Paz</b>
						<div>Mate &amp; Co · Posadas</div>
					</div>
					<span class="estrellas">
						{#each [1, 2, 3, 4, 5] as n (n)}<Icon icon={phStarFill} size={13} />{/each}
					</span>
				</div>
				<p>
					“Me mudé desde Tiendanube en una tarde y ahora pago la mitad. Hornero me contesta los
					WhatsApp a la noche.”
				</p>
			</div>
		</div>
		<div class="garantias">
			<span><Icon icon={phLockSimpleFill} size={15} />Conexión cifrada</span>
			<span><Icon icon={phShieldCheckFill} size={15} />Verificación en dos pasos</span>
			<span><Icon icon={phMapPinFill} size={15} />Hecho en Argentina</span>
		</div>
	</aside>
</div>

<CookieBanner />

<style>
	.pag {
		min-height: 100vh;
		display: flex;
		flex-wrap: wrap;
		color: #2a140c;
		background: #fbf7f2;
	}
	/* Panel izquierdo */
	.lado {
		flex: 1 1 420px;
		min-height: min(100vh, 900px);
		position: relative;
		overflow: hidden;
		background: #2a140c url('/textura/arch-pattern-dark.svg');
		color: #fbf7f2;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		padding: clamp(28px, 4vw, 56px);
	}
	.brillo {
		position: absolute;
		right: -12%;
		top: -10%;
		width: 70%;
		aspect-ratio: 1;
		border-radius: 50%;
		background: radial-gradient(closest-side, rgba(210, 88, 47, 0.45), rgba(210, 88, 47, 0));
	}
	.chips {
		position: relative;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.chips span {
		padding: 7px 12px;
		border-radius: 999px;
		background: rgba(251, 247, 242, 0.08);
		border: 1px solid rgba(251, 247, 242, 0.12);
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
	}
	.medio {
		position: relative;
	}
	.ave {
		width: clamp(150px, 18vw, 220px);
		height: auto;
		display: block;
	}
	.lema {
		text-wrap: balance;
		margin-top: 8px;
		font-size: clamp(30px, 3.4vw, 46px);
		line-height: 1.02;
		letter-spacing: -0.04em;
		font-weight: 600;
		max-width: 460px;
	}
	.testi {
		margin-top: 22px;
		max-width: 420px;
		padding: 16px;
		border-radius: 22px;
		background: #fbf7f2;
		color: #2a140c;
		box-shadow: 0 30px 60px -24px rgba(0, 0, 0, 0.6);
	}
	.testi-cab {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.testi-cab img {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		object-fit: cover;
	}
	.testi-quien {
		min-width: 0;
	}
	.testi-quien b {
		font-size: 14.5px;
	}
	.testi-quien div {
		font-size: 12.5px;
		color: #7a6252;
		white-space: nowrap;
	}
	.estrellas {
		margin-left: auto;
		display: flex;
		flex: none;
		color: #b0441d;
	}
	.testi p {
		margin: 10px 0 0;
		font-size: 14.5px;
		line-height: 1.5;
	}
	.garantias {
		position: relative;
		display: flex;
		gap: 16px;
		flex-wrap: wrap;
		font-size: 13px;
		color: #d9c8ba;
	}
	.garantias span {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	/* Columna del formulario */
	.der {
		flex: 1 1 480px;
		display: flex;
		flex-direction: column;
		padding: clamp(20px, 3vw, 36px) clamp(20px, 5vw, 72px);
	}
	.arriba {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.marca {
		height: 21px;
		width: auto;
		display: block;
	}
	.ayuda {
		font-size: 14px;
		font-weight: 600;
		color: #5c4536;
		display: flex;
		align-items: center;
		gap: 6px;
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
	.col.reg {
		gap: 14px;
	}
	h1 {
		margin: 0;
		font-size: clamp(32px, 3.6vw, 42px);
		line-height: 1.02;
		letter-spacing: -0.04em;
		font-weight: 600;
	}
	.bajada {
		margin: 8px 0 0;
		font-size: 15.5px;
		line-height: 1.5;
		color: #5c4536;
	}
	.bajada b {
		color: #2a140c;
	}
	.sociales {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}
	.social {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 9px;
		padding: 13px;
		border-radius: 999px;
		border: 1.5px solid rgba(58, 29, 18, 0.14);
		background: #fff;
		font-size: 14.5px;
		font-weight: 600;
		color: #2a140c;
	}
	.o {
		display: flex;
		align-items: center;
		gap: 12px;
		font-size: 13px;
		color: var(--marron-2);
	}
	.o span {
		flex: 1;
		height: 1px;
		background: rgba(58, 29, 18, 0.12);
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 10px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
	}
	.in {
		position: relative;
		display: flex;
	}
	.in input {
		flex: 1;
		min-width: 0;
		padding: 14px 16px;
		border-radius: 14px;
		border: 1.5px solid rgba(58, 29, 18, 0.14);
		background: #fff;
		font-size: 15.5px;
		font-weight: 500;
		color: #2a140c;
		outline: none;
	}
	.in input:focus {
		border-color: #d2582f;
	}
	.in input[aria-invalid='true'] {
		border-color: #b83a2e;
	}
	.in input.con-ojo {
		padding-right: 46px;
	}
	.ojo {
		position: absolute;
		right: 8px;
		top: 50%;
		transform: translateY(-50%);
		width: 34px;
		height: 34px;
		border: 0;
		border-radius: 10px;
		background: transparent;
		display: grid;
		place-items: center;
		color: #7a6252;
	}
	small.err {
		margin-top: -10px;
		font-size: 12.5px;
		font-weight: 500;
		color: #b83a2e;
	}
	.fila {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 10px;
		font-size: 14px;
	}
	.check {
		display: flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
	}
	.check input {
		position: absolute;
		opacity: 0;
		width: 1px;
		height: 1px;
	}
	.caja-check {
		width: 18px;
		height: 18px;
		flex: none;
		border-radius: 5px;
		border: 1.5px solid rgba(58, 29, 18, 0.3);
		color: transparent;
		display: grid;
		place-items: center;
	}
	.check input:checked + .caja-check {
		background: #2a140c;
		border-color: #2a140c;
		color: #fff;
	}
	.check input:focus-visible + .caja-check {
		outline: 2px solid #d2582f;
		outline-offset: 2px;
	}
	.terminos {
		align-items: flex-start;
		gap: 10px;
		font-size: 13.5px;
		line-height: 1.5;
		color: #5c4536;
	}
	.terminos .caja-check {
		margin-top: 1px;
	}
	.terminos a {
		font-weight: 600;
		text-decoration: underline;
	}
	.link {
		border: 0;
		background: transparent;
		padding: 0;
		font-size: 14px;
		font-weight: 600;
		color: #b0441d;
	}
	.cta {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 8px;
		padding: 16px;
		border: 0;
		border-radius: 999px;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 15.5px;
		font-weight: 600;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.25),
			0 12px 26px -12px rgba(210, 88, 47, 0.8);
	}
	.cta:disabled {
		opacity: 0.7;
		cursor: progress;
	}
	.pie-form {
		text-align: center;
		font-size: 14.5px;
		color: #5c4536;
	}
	.barras {
		display: flex;
		gap: 4px;
	}
	.barras span {
		flex: 1;
		height: 4px;
		border-radius: 2px;
		transition: background 0.25s;
	}
	.fuerza {
		font-size: 12.5px;
		color: #7a6252;
		margin-top: -6px;
	}
	.volver {
		align-self: flex-start;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 14px;
		font-weight: 600;
		color: #5c4536;
	}
	.ave-chica {
		width: 96px;
		height: 96px;
		object-fit: contain;
	}
	.ave-exito {
		width: 110px;
		height: 110px;
		object-fit: contain;
	}
	.seg {
		display: flex;
		padding: 4px;
		border-radius: 999px;
		background: #f4ece2;
	}
	.seg button {
		flex: 1;
		text-align: center;
		padding: 10px;
		border: 0;
		border-radius: 999px;
		background: transparent;
		font-size: 14px;
		font-weight: 600;
		color: #5c4536;
	}
	.seg button[aria-pressed='true'] {
		background: #fff;
		color: #2a140c;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.1);
	}
	.enviado {
		align-items: flex-start;
	}
	.nota {
		width: 100%;
		padding: 14px 16px;
		border-radius: 16px;
		background: #f4ece2;
		font-size: 14px;
		line-height: 1.5;
		color: #5c4536;
		display: flex;
		gap: 10px;
	}
	.botones {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		width: 100%;
	}
	.botones .sec,
	.botones .oscuro {
		flex: 1;
		padding: 14px;
		border-radius: 999px;
		font-size: 14.5px;
		font-weight: 600;
		text-align: center;
		white-space: nowrap;
	}
	.sec {
		border: 1.5px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		color: #2a140c;
	}
	.sec:disabled {
		cursor: default;
	}
	.oscuro {
		border: 0;
		background: #2a140c;
		color: #fff;
	}
	.oscuro:hover {
		color: #fff;
	}
	.pie-nota {
		margin: 0;
		font-size: 13px;
		line-height: 1.5;
		color: #7a6252;
		text-align: center;
	}
	.centro-l {
		align-self: center;
	}
	.err-form {
		margin-bottom: 16px;
		padding: 12px 14px;
		border-radius: 14px;
		background: rgba(184, 58, 46, 0.07);
		border: 1px solid rgba(184, 58, 46, 0.28);
		font-size: 14px;
		color: #2a140c;
	}
	.legal {
		display: flex;
		gap: 14px;
		flex-wrap: wrap;
		font-size: 12.5px;
		color: #7a6252;
	}
	.legal a {
		color: #7a6252;
		text-decoration: underline;
		display: flex;
		align-items: center;
		gap: 5px;
	}
	.punto {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: #3fa877;
	}
</style>
