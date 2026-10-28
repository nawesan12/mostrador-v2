<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { TRAE_MP } from './datos.ts';
	import Etiqueta from './Etiqueta.svelte';

	/**
	 * Paso 1, Cuenta. Los botones mandan al ingreso con ese proveedor y el email al registro; en la
	 * demo no hay cuentas y siguen directo al paso 2 (docs/decisiones.md §12).
	 */
	let { demo, seguir, error }: { demo: boolean; seguir: () => void; error?: string } = $props();
	let email = $state('');
</script>

<div class="dos">
	<div class="izq">
		<Etiqueta rotulo="Crear cuenta" titulo="Abrí tu mostrador. No arrancás de cero." grande />
		<p class="bajada">
			Entrá con la cuenta que ya usás y traemos tus datos desde ahí: nombre, CUIT, clientes y
			ventas.
		</p>
		<form class="botones" method="POST" action="?/proveedor">
			<button
				class="mp"
				name="proveedor"
				value="mercadopago"
				type={demo ? 'button' : 'submit'}
				onclick={demo ? seguir : undefined}
			>
				<span class="ico blanco"
					><Icon icon={I.siMercadopago} size={18} style="color:#00B1EA" /></span
				>
				<span class="txt">Continuar con Mercado Pago</span><span class="reco">Recomendado</span>
			</button>
			<button
				class="ml"
				name="proveedor"
				value="mercadolibre"
				type={demo ? 'button' : 'submit'}
				onclick={demo ? seguir : undefined}
			>
				<span class="ico ml-ico">ML</span><span class="txt">Continuar con Mercado Libre</span>
			</button>
			<button
				class="g"
				name="proveedor"
				value="google"
				type={demo ? 'button' : 'submit'}
				onclick={demo ? seguir : undefined}
			>
				<span class="ico beige"><Icon icon={I.logosGoogleIcon} size={16} /></span>
				<span class="txt">Continuar con Google</span>
			</button>
		</form>
		<div class="o"><span></span>o con tu email<span></span></div>
		<form
			class="email"
			method="POST"
			action="?/email"
			onsubmit={(ev) => {
				if (demo) {
					ev.preventDefault();
					seguir();
				}
			}}
		>
			<input
				name="email"
				type="email"
				autocomplete="email"
				placeholder="tu@email.com"
				aria-label="Tu email"
				bind:value={email}
			/>
			<button type="submit">Seguir</button>
		</form>
		{#if error}<div class="error" role="alert">{error}</div>{/if}
		<div class="legal">
			Al continuar aceptás los términos y la política de privacidad. 14 días gratis, sin tarjeta.
		</div>
	</div>
	<div class="trae">
		<div class="t-t">Con Mercado Pago traemos</div>
		<div class="t-l">
			{#each TRAE_MP as [ico, t] (t)}
				<div>
					<span class="t-ico"><Icon icon={ico} size={17} /></span><span class="t-x">{t}</span><Icon
						icon={I.phCheckBold}
						size={14}
						style="color:#237550"
					/>
				</div>
			{/each}
		</div>
		<div class="t-pie">
			<Icon
				icon={I.phLockKeyFill}
				size={13}
				style="color:#b0441d;display:inline;vertical-align:-1px"
			/> Solo leemos lo necesario. Nunca movemos tu plata sin permiso.
		</div>
	</div>
</div>

<style>
	.dos {
		display: flex;
		gap: 40px;
		flex-wrap: wrap;
	}
	.izq {
		flex: 1 1 380px;
		min-width: 0;
	}
	.bajada {
		margin: 14px 0 0;
		font-size: 17px;
		line-height: 1.55;
		color: #5c4536;
	}
	.botones {
		margin-top: 28px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.botones button {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 18px;
		border-radius: 16px;
		border: 0;
		font-size: 15.5px;
		font-weight: 600;
		text-align: left;
	}
	.mp {
		background: #00b1ea;
		color: #fff;
		box-shadow: 0 12px 26px -14px rgba(0, 120, 170, 0.8);
	}
	.ml {
		background: #ffe600;
		color: #2d3277;
	}
	.g {
		border: 1px solid rgba(58, 29, 18, 0.14) !important;
		background: #fff;
		color: #2a140c;
	}
	.ico {
		width: 30px;
		height: 30px;
		flex: none;
		border-radius: 9px;
		display: grid;
		place-items: center;
	}
	.blanco {
		background: #fff;
	}
	.beige {
		background: #f6eee4;
	}
	.ml-ico {
		background: #2d3277;
		color: #ffe600;
		font-size: 9px;
		font-weight: 800;
	}
	.txt {
		flex: 1;
	}
	.reco {
		font-size: 12px;
		font-weight: 700;
		background: rgba(255, 255, 255, 0.22);
		padding: 3px 8px;
		border-radius: 999px;
	}
	.o {
		margin: 22px 0;
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
	.email {
		display: flex;
		gap: 8px;
	}
	.email input {
		flex: 1;
		min-width: 0;
		padding: 14px 16px;
		border-radius: 14px;
		border: 1px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		font-size: 15px;
		font-weight: 500;
		color: #2a140c;
		outline: none;
	}
	.email input:focus {
		border-color: #d2582f;
	}
	.email button {
		padding: 14px 18px;
		border-radius: 14px;
		border: 0;
		background: #2a140c;
		color: #fff;
		font-size: 15px;
		font-weight: 600;
	}
	.error {
		margin-top: 10px;
		font-size: 13.5px;
		color: #b83a2e;
	}
	.legal {
		margin-top: 14px;
		font-size: 12.5px;
		color: var(--marron-2);
		line-height: 1.5;
	}
	.trae {
		flex: 1 1 300px;
		padding: 24px;
		border-radius: 24px;
		background: #fff;
		box-shadow:
			0 1px 2px rgba(58, 29, 18, 0.06),
			0 24px 50px -30px rgba(58, 29, 18, 0.4);
		align-self: flex-start;
	}
	.t-t {
		font-size: 15px;
		font-weight: 600;
	}
	.t-l {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.t-l > div {
		display: flex;
		gap: 12px;
		align-items: center;
	}
	.t-ico {
		width: 34px;
		height: 34px;
		flex: none;
		border-radius: 10px;
		background: #fcede5;
		color: #b0441d;
		display: grid;
		place-items: center;
	}
	.t-x {
		flex: 1;
		font-size: 14px;
	}
	.t-pie {
		margin-top: 18px;
		padding: 12px 14px;
		border-radius: 14px;
		background: #f6eee4;
		font-size: 13px;
		line-height: 1.45;
		color: #5c4536;
	}
</style>
