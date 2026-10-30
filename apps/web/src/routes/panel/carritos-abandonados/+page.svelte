<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { cara, unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { hace, iniciales } from '#lib/panel/listas/clientes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Carritos abandonados (design/PanelVentas.dc.html): lo que quedó, lo recuperado y los recordatorios. */
	let { data } = $props();
	const ui = getUi();
	const k = $derived(data.carritos);
	const pendientes = $derived(k.carritos.filter((x) => !x.recordado));
	const ticket = $derived(k.abiertos ? k.monto / k.abiertos : 0);
	/**
	 * Con la API real, los recordatorios van por mail y no cuestan nada (WhatsApp Business todavía no
	 * se conecta). Los de WhatsApp y el segundo aviso con descuento sólo existen en la demo.
	 */
	const demo = $derived(!!page.data.demo);

	async function post(ruta: string, cuerpo: object = {}) {
		const r = await fetch(`/panel/carritos-abandonados/${ruta}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(cuerpo)
		});
		await invalidateAll();
		return r;
	}
	async function todos() {
		const r = await post('todos');
		const { enviados } = (await r.json()) as { enviados: number };
		ui.toast({
			k: demo ? 'wa' : 'ok',
			title: enviados
				? `${enviados} ${enviados === 1 ? 'recordatorio' : 'recordatorios'} en camino`
				: 'Ya estaban todos recordados',
			desc: !enviados
				? 'No se mandó nada de nuevo'
				: demo
					? `Costo total ${pesos(enviados * k.costoMensaje)} · te avisamos cuando compren`
					: 'Por mail, con el link a su carrito'
		});
	}
	async function uno(c: (typeof k.carritos)[number]) {
		const r = await post(c.id);
		if (r.ok)
			ui.toast({
				k: demo ? 'wa' : 'ok',
				title: 'Recordatorio enviado',
				desc: demo
					? `${c.nombre} · ${pesos(k.costoMensaje)} de WhatsApp`
					: `${c.nombre} · por mail, con el link a su carrito`
			});
	}
	const TODOS = [
		['hora', 'Recordatorio por WhatsApp a la hora'],
		['dia', 'Segundo aviso a las 24 h con 5% off'],
		['email', 'Email al tercer día']
	] as const;
	const AUTOS = $derived(
		demo ? TODOS : ([['email', 'Mail al tercer día con el link a su carrito']] as const)
	);
	async function automatico(clave: (typeof TODOS)[number][0], on: boolean) {
		await post('automaticos', { ...k.automaticos, [clave]: on });
		ui.toast({
			k: 'ok',
			title: on ? 'Recordatorio activado' : 'Recordatorio apagado',
			desc: AUTOS.find((a) => a[0] === clave)?.[1] ?? ''
		});
	}
</script>

<Seo
	title="Carritos abandonados · Panel"
	description="Carritos de tu tienda"
	path="/panel/carritos-abandonados"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Carritos abandonados</h1>
			<div class="c-s">Gente que estuvo a un paso de comprarte</div>
		</div>
		<button type="button" class="primario" disabled={!pendientes.length} onclick={todos}
			>{#if demo}<Icon icon={I.phWhatsappLogoFill} size={15} />Recordar a todos · {pesos(
					pendientes.length * k.costoMensaje
				)}{:else}<Icon icon={I.phEnvelopeSimpleFill} size={15} />Recordar a todos por mail{/if}</button
		>
	</div>

	<div class="kpis" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="oscura">
			<div class="o-in">
				<div class="o-t">En carritos abiertos</div>
				<div class="o-v">{pesos(k.monto)}</div>
				<div class="o-s">
					{k.abiertos}
					{k.abiertos === 1 ? 'carrito' : 'carritos'} · ticket promedio {pesos(ticket)}
				</div>
			</div>
		</div>
		<div class="tarjeta">
			<div class="t-t">Recuperado este mes</div>
			<div class="t-v">{pesos(k.recuperado)}</div>
			<div class="barra"><div style:width="{k.tasa}%"></div></div>
			<div class="t-s">{k.tasa}% de recuperación</div>
		</div>
		<div class="tarjeta autos">
			<div class="a-t">Automático</div>
			{#each AUTOS as [clave, t] (clave)}
				<div class="auto">
					<span>{t}</span>
					<Interruptor
						label={t}
						checked={k.automaticos[clave]}
						onchange={(on: boolean) => automatico(clave, on)}
					/>
				</div>
			{/each}
		</div>
	</div>

	<div class="carritos" use:anim={{ t: 'rise', delay: 120 }}>
		{#each k.carritos as c (c.id)}
			<div class="carrito">
				{#if c.foto}<img class="av" src={cara(c.foto)} alt="" />{:else}<span class="av ini"
						>{iniciales(c.nombre)}</span
					>{/if}
				<span class="quien"
					><span class="n">{c.nombre}</span><span class="s"
						>{hace((Date.now() - Date.parse(c.creado)) / 86_400_000)} · desde {c.desde}</span
					></span
				>
				<span class="fotos"
					>{#each c.fotos as f, i (i)}{#if f}<img
								src={unsplash(f, 80, '&h=80')}
								alt=""
							/>{/if}{/each}</span
				>
				<span class="monto"
					><span class="v">{pesos(c.total)}</span><span class="q"
						>{c.cantidad} {c.cantidad === 1 ? 'producto' : 'productos'}</span
					></span
				>
				{#if c.recordado}
					<span class="hecho"><Icon icon={I.phCheckCircleFill} size={16} />Recordado</span>
				{:else}
					<button type="button" class="recordar" onclick={() => uno(c)}
						><Icon
							icon={demo ? I.logosWhatsappIcon : I.phEnvelopeSimpleFill}
							size={14}
						/>Recordar</button
					>
				{/if}
			</div>
		{:else}
			<div class="vacio">
				No hay carritos abandonados. Cuando alguien deje su mail en el checkout y no termine la
				compra, aparece acá.
			</div>
		{/each}
	</div>
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.cab {
		display: flex;
		align-items: end;
		gap: 14px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 220px;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.primario {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 20px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			0 10px 24px -10px rgba(210, 88, 47, 0.8);
	}
	.primario:disabled {
		background: var(--chip);
		color: var(--mu);
		box-shadow: none;
	}
	.kpis {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 12px;
	}
	.oscura {
		position: relative;
		overflow: hidden;
		padding: 20px;
		border-radius: 22px;
		background: #2a140c;
		color: #fbf7f2;
		box-shadow: var(--shd);
	}
	.o-in {
		position: relative;
	}
	.o-t {
		font-size: 13px;
		color: #d9c8ba;
	}
	.o-v,
	.t-v {
		font-size: 34px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.o-s {
		font-size: 13px;
		color: #f5b796;
		margin-top: 2px;
	}
	.tarjeta {
		padding: 20px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.t-t {
		font-size: 13px;
		color: var(--mu);
	}
	.t-v {
		color: #237550;
	}
	.barra {
		margin-top: 8px;
		height: 8px;
		border-radius: 8px;
		background: var(--sf3);
		overflow: hidden;
	}
	.barra div {
		height: 100%;
		border-radius: 8px;
		background: #2f8f63;
	}
	.t-s {
		font-size: 12.5px;
		color: var(--mu);
		margin-top: 6px;
	}
	.autos {
		padding: 16px 18px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.a-t {
		font-size: 14px;
		font-weight: 600;
		margin-bottom: 4px;
	}
	.auto {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 7px 0;
	}
	.auto span {
		flex: 1;
		font-size: 13.5px;
		line-height: 1.3;
	}
	.carritos {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.carrito {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 14px 16px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		flex-wrap: wrap;
	}
	.av {
		width: 44px;
		height: 44px;
		border-radius: 50%;
		object-fit: cover;
		flex: none;
	}
	.ini {
		display: grid;
		place-items: center;
		background: var(--sf3);
		color: var(--mu);
		font-size: 13px;
		font-weight: 700;
	}
	.quien {
		flex: 1 1 180px;
		min-width: 0;
	}
	.n {
		display: block;
		font-size: 15px;
		font-weight: 600;
	}
	.s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.fotos {
		display: flex;
		padding-left: 8px;
	}
	.fotos img {
		width: 40px;
		height: 40px;
		border-radius: 10px;
		object-fit: cover;
		margin-left: -8px;
		box-shadow: 0 0 0 2px var(--sf);
	}
	.monto {
		min-width: 90px;
		text-align: right;
	}
	.v {
		display: block;
		font-size: 16px;
		font-weight: 600;
		white-space: nowrap;
	}
	.q {
		display: block;
		font-size: 12px;
		color: var(--mu);
	}
	.recordar {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 9px 14px;
		border-radius: 999px;
		border: 0;
		background: #e9f8ee;
		color: #1f6f4a;
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
	}
	.hecho {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		color: #237550;
		white-space: nowrap;
	}
	.vacio {
		padding: 28px;
		text-align: center;
		color: var(--mu);
		font-size: 14px;
	}
</style>
