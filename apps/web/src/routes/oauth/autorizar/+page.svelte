<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import {
		phCheckBold,
		phCheckCircleFill,
		phInfoFill,
		phLockSimpleFill,
		phMapPinFill,
		phQuestion,
		phShieldCheckFill
	} from '#lib/icons.ts';
	import { IAS } from '#lib/panel/integraciones.ts';

	/**
	 * Consentimiento del OAuth del conector MCP. No está en el handoff: se arma con las piezas de
	 * design/Mostrador Ingresar.dc.html (columna del formulario, nota, casilla, botones y panel lateral).
	 */
	let { data } = $props();
	const ia = $derived(IAS.find((p) => p.id === data.ia.id) ?? IAS[3]);
	let acciones = $state(true);
</script>

<Seo
	title="Conectar una IA · Mostrador"
	description="Permití que tu IA consulte y maneje tu tienda."
	path="/oauth/autorizar"
	noindex
/>

<div class="pag">
	<div class="der">
		<div class="arriba">
			<a href={resolve('/')}><img src="/wordmark-brown.png" alt="mostrador" class="marca" /></a>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- página pública del centro de ayuda -->
			<a class="ayuda" href="/ayuda"><Icon icon={phQuestion} size={16} />Ayuda</a>
		</div>

		<div class="centro">
			<div class="caja">
				<form method="POST" class="col">
					<img src="/hornero/seguridad.png" alt="" class="ave-chica" />
					<div>
						<h1>¿Conectar {data.ia.nombre} con tu&nbsp;Mostrador?</h1>
						<p class="bajada">
							Va a poder trabajar con los datos de
							<b>{data.tienda ?? 'tu tienda'}</b>. Después vuelve a&nbsp;{data.ia.host}.
						</p>
					</div>

					<div class="quien">
						<span class="ia-ico" style:color={ia.col}><Icon icon={ia.icon} size={20} /></span>
						<span class="quien-txt"><b>{data.ia.nombre}</b><span>{data.ia.host}</span></span>
					</div>

					<ul class="permisos">
						<li>
							<Icon icon={phCheckCircleFill} size={18} />
							<span
								><b>Ver ventas, pedidos, productos y clientes</b><br />Sólo lectura,
								siempre&nbsp;activo.</span
							>
						</li>
					</ul>
					{#if data.acciones}
						<label class="check terminos">
							<input type="checkbox" name="acciones" bind:checked={acciones} />
							<span class="caja-check"><Icon icon={phCheckBold} size={11} /></span>
							<span
								><b>También puede hacer cambios</b><br />Cambiar el estado de un pedido, ajustar
								stock y precios, crear cupones y links de pago. Si no lo marcás, sólo
								puede&nbsp;leer.</span
							>
						</label>
					{/if}

					<div class="nota">
						<Icon icon={phInfoFill} size={18} style="color:#C24E25;flex:none;margin-top:1px" />
						<span
							>Podés desconectarla cuando quieras en Integraciones. En la demo, el acceso vence con
							tu tienda de prueba, a las 24&nbsp;horas.</span
						>
					</div>

					<div class="botones">
						<button type="submit" class="sec" name="decision" value="cancelar">Cancelar</button>
						<button type="submit" class="cta" name="decision" value="permitir">Permitir</button>
					</div>
				</form>
			</div>
		</div>

		<div class="legal">
			<span>© 2026 Mostrador</span>
			<!-- eslint-disable svelte/no-navigation-without-resolve -- páginas públicas -->
			<a href="/legales/terminos">Términos</a>
			<a href="/legales/privacidad">Privacidad</a>
			<a href="/estado"><span class="punto"></span>Todos los sistemas funcionan</a>
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
		</div>
	</div>
	<aside class="lado">
		<div class="brillo"></div>
		<div class="chips">
			<span>Ventas</span><span>Pedidos</span><span>Stock</span><span>Cupones</span>
		</div>
		<div class="medio">
			<img class="ave" src="/hornero/saludo.png" alt="" />
			<div class="lema">Preguntale a tu IA cómo viene el&nbsp;día.</div>
		</div>
		<div class="garantias">
			<span><Icon icon={phLockSimpleFill} size={15} />Conexión cifrada</span>
			<span><Icon icon={phShieldCheckFill} size={15} />Desconectás cuando quieras</span>
			<span><Icon icon={phMapPinFill} size={15} />Hecho en Argentina</span>
		</div>
	</aside>
</div>

<style>
	/* Las reglas salen de src/routes/ingresar/+page.svelte (handoff Mostrador Ingresar). */
	.pag {
		min-height: 100vh;
		display: flex;
		flex-wrap: wrap;
		color: #2a140c;
		background: #fbf7f2;
	}
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
	h1 {
		margin: 0;
		font-size: clamp(32px, 3.6vw, 42px);
		line-height: 1.02;
		letter-spacing: -0.04em;
		font-weight: 600;
		text-wrap: balance;
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
	.ave-chica {
		width: 96px;
		height: 96px;
		object-fit: contain;
	}
	/* La IA que pide acceso: el botón "social" de Ingresar, sin acción. */
	.quien {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 13px 16px;
		border-radius: 14px;
		border: 1.5px solid rgba(58, 29, 18, 0.14);
		background: #fff;
	}
	.ia-ico {
		width: 36px;
		height: 36px;
		flex: none;
		border-radius: 11px;
		display: grid;
		place-items: center;
		background: #f4ece2;
	}
	.quien-txt {
		display: flex;
		flex-direction: column;
		min-width: 0;
		font-size: 14.5px;
	}
	.quien-txt span {
		font-size: 12.5px;
		color: #7a6252;
		overflow-wrap: anywhere;
	}
	.permisos {
		list-style: none;
		margin: 0;
		padding: 0;
	}
	.permisos li {
		display: flex;
		gap: 10px;
		align-items: flex-start;
		font-size: 13.5px;
		line-height: 1.5;
		color: #5c4536;
	}
	.permisos li :global(svg) {
		flex: none;
		margin-top: 1px;
		color: #2f8f63;
	}
	.permisos b,
	.terminos b {
		color: #2a140c;
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
	.botones .cta {
		flex: 1;
		min-height: 48px;
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
	.cta {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 8px;
		border: 0;
		background: linear-gradient(180deg, #e06a3e, #d2582f);
		color: #fff;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.25),
			0 12px 26px -12px rgba(210, 88, 47, 0.8);
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
