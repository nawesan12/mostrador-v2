<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { CONEXIONES } from './datos.ts';
	import Etiqueta from './Etiqueta.svelte';
	import { getOnboarding } from './estado.svelte.ts';
	import Logo from './Logo.svelte';

	/** Paso 3, Canales: conectar lo que ya se usa, y la invitación a Caja. */
	const o = getOnboarding();
</script>

<div class="cab">
	<div class="cab-t">
		<Etiqueta rotulo="Canales" titulo="Conectá lo que ya usás. Traemos todo de una." />
	</div>
	<button type="button" class="todas" onclick={() => o.conectarRecomendadas()}
		><Icon icon={I.phLightningFill} size={15} style="color:#F5B796" />Conectar las recomendadas</button
	>
</div>
<div class="grilla">
	{#each CONEXIONES as [k, n, d, trae, rec] (k)}
		{@const c = o.e.conexiones[k] ?? 'off'}
		<div class="con" class:on={c === 'on'}>
			<Logo {k} />
			<div class="c-txt">
				<div class="c-n">
					<span>{n}</span>{#if rec}<span class="reco">Recomendado</span>{/if}
				</div>
				<div class="c-d" class:trae={c === 'on'}>{c === 'on' ? trae : d}</div>
			</div>
			{#if c === 'off'}
				<button type="button" class="conectar" onclick={() => o.conectar(k)}>Conectar</button>
			{:else if c === 'wait'}
				<span class="espera"
					><Icon icon={I.phCircleNotchBold} size={15} class="gira" />Conectando</span
				>
			{:else}
				<span class="listo" aria-label="Conectado"><Icon icon={I.phCheckBold} size={14} /></span>
			{/if}
		</div>
	{/each}
</div>
<div class="resumen">
	{o.conectadas} de {CONEXIONES.length} conectadas. Podés sumar más después desde Integraciones.
</div>
<div class="caja">
	<img src="/caja/01-saludo.webp" alt="Caja" width="64" height="64" />
	<div class="cj-txt">
		<div class="cj-t">
			<b>Sumá Caja y sabé cuánto ganás de verdad</b><span>3 MESES DE PRO MAX GRATIS</span>
		</div>
		<div class="cj-d">
			Tus ventas, cobros, comisiones y facturas pasan solos a Caja, separados de tus gastos
			personales. Con el monotributo y el dólar del día ya calculados.
		</div>
	</div>
	<button
		type="button"
		class:conectada={o.e.caja}
		onclick={() => ((o.e.caja = !o.e.caja), o.guardar())}
		><Icon icon={o.e.caja ? I.phCheckCircleFill : I.phPlugFill} size={15} />{o.e.caja
			? 'Caja conectada'
			: 'Conectar Caja'}</button
	>
</div>

<style>
	.cab {
		display: flex;
		align-items: end;
		gap: 20px;
		flex-wrap: wrap;
	}
	.cab-t {
		flex: 1 1 400px;
	}
	.todas {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 18px;
		border-radius: 999px;
		border: 0;
		background: #2a140c;
		color: #fff;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
	}
	.grilla {
		margin-top: 24px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr));
		gap: 12px;
	}
	.con {
		display: flex;
		gap: 14px;
		padding: 16px;
		border-radius: 20px;
		background: #fff;
		border: 1.5px solid rgba(58, 29, 18, 0.08);
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.05);
		transition: border-color 0.3s;
	}
	.con.on {
		border-color: rgba(47, 143, 99, 0.45);
	}
	.c-txt {
		flex: 1;
		min-width: 0;
	}
	.c-n {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 15px;
		font-weight: 600;
		flex-wrap: wrap;
	}
	.reco {
		font-size: 11px;
		font-weight: 700;
		color: #b0441d;
		background: rgba(210, 88, 47, 0.12);
		padding: 2px 7px;
		border-radius: 999px;
	}
	.c-d {
		font-size: 13px;
		color: #7a6252;
		margin-top: 3px;
		line-height: 1.4;
	}
	.c-d.trae {
		color: #237550;
		font-weight: 500;
	}
	.conectar {
		align-self: center;
		padding: 8px 14px;
		border-radius: 999px;
		border: 1px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		color: #2a140c;
		font-size: 13px;
		font-weight: 600;
	}
	.espera {
		align-self: center;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		color: #b0441d;
		white-space: nowrap;
	}
	.espera :global(.gira) {
		animation: gira 0.9s linear infinite;
	}
	@keyframes gira {
		to {
			transform: rotate(360deg);
		}
	}
	.listo {
		align-self: center;
		width: 30px;
		height: 30px;
		flex: none;
		border-radius: 50%;
		background: #2f8f63;
		color: #fff;
		display: grid;
		place-items: center;
	}
	.resumen {
		margin-top: 14px;
		font-size: 13px;
		color: #7a6252;
	}
	.caja {
		margin-top: 16px;
		display: flex;
		align-items: center;
		gap: 18px;
		flex-wrap: wrap;
		padding: 18px 20px;
		border-radius: 24px;
		background: #131211;
		color: #f4eee6;
	}
	.caja img {
		object-fit: contain;
		flex: none;
	}
	.cj-txt {
		flex: 1 1 300px;
		min-width: 0;
	}
	.cj-t {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.cj-t b {
		font-size: 16px;
	}
	.cj-t span {
		font-size: 11px;
		font-weight: 700;
		padding: 3px 8px;
		border-radius: 999px;
		background: #e8875f;
		color: #131211;
		letter-spacing: 0.02em;
	}
	.cj-d {
		font-size: 13.5px;
		color: #bdb2a8;
		margin-top: 4px;
		line-height: 1.45;
	}
	.caja button {
		flex: none;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 18px;
		border-radius: 999px;
		border: 0;
		background: #f4eee6;
		color: #131211;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		transition: background 0.25s;
	}
	.caja button.conectada {
		background: #2f8f63;
		color: #fff;
	}
</style>
