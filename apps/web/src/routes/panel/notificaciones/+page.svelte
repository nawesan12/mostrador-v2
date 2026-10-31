<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { deslizable } from '#lib/panel/deslizable.ts';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { activo, AVISOS, CANALES_AVISO, type CanalAviso } from '#lib/panel/avisos.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Notificaciones (design/PanelCuenta.dc.html, sNot): qué avisos y por dónde. */
	let { data } = $props();
	const ui = getUi();
	let prefs = $derived({ ...data, matriz: { ...data.matriz } });
	/**
	 * Con la API real, a la tienda le salen hoy tres avisos de venta nueva: el mail (Venta nueva ›
	 * Email, prendido por defecto), el push al iPhone (Venta nueva › Celular, prendido por defecto;
	 * docs/decisiones.md §40) y Telegram si lo vinculó. Y «Sin stock» por mail y push (§41). La matriz se guarda, pero no se promete
	 * un resumen diario, un «no molestar» ni una prueba que no mandan nada (la prueba era un toast
	 * con una venta inventada).
	 */
	const demo = $derived(!!page.data.demo);
	let verMas = $state(false);

	const CANAL: Record<CanalAviso, [I.IconData, string, string]> = {
		panel: [I.phBellFill, '#D2582F', 'Panel'],
		whatsapp: [I.phWhatsappLogoFill, '#25D366', 'WhatsApp'],
		email: [I.phEnvelopeSimpleFill, '#3E6BFD', 'Email'],
		celular: [I.phDeviceMobileFill, '#6A4FA0', 'Celular']
	};
	// Los guardados van en fila, como en Mi tienda.
	let fila = Promise.resolve();
	function guardar() {
		const cuerpo = JSON.stringify(prefs);
		fila = fila.then(async () => {
			await fetch('/panel/notificaciones/preferencias', {
				method: 'PUT',
				headers: { 'content-type': 'application/json' },
				body: cuerpo
			});
		});
	}
	function alternar(aviso: string, canal: CanalAviso) {
		prefs = {
			...prefs,
			matriz: { ...prefs.matriz, [`${aviso}:${canal}`]: !activo(prefs.matriz, aviso, canal) }
		};
		guardar();
	}
	function silencio(v: boolean) {
		prefs = { ...prefs, silencio: v };
		guardar();
		ui.toast({
			k: 'ok',
			title: v ? 'No molestar de 23 a 8 h' : 'No molestar desactivado',
			desc: v ? 'Las ventas igual quedan en el panel' : 'Te avisamos a cualquier hora'
		});
	}
	function resumen(r: 'whatsapp' | 'email' | 'off') {
		prefs = { ...prefs, resumen: r };
		guardar();
	}
</script>

<Seo title="Notificaciones · Panel" description="Tus avisos" path="/panel/notificaciones" noindex />

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Notificaciones</h1>
			<div class="c-s">Elegí qué te avisamos y por dónde</div>
		</div>
		{#if demo}
			<div class="silencio">
				<Icon icon={I.phMoonStarsFill} size={16} style="color:#6A4FA0" />No molestar de 23 a 8 h
				<Interruptor checked={prefs.silencio} label="No molestar de 23 a 8 h" onchange={silencio} />
			</div>
		{/if}
	</div>

	{#if !demo}
		<!-- Lo que hoy sale de verdad va arriba y a la vista; el resto de la matriz se guarda para cuando
		     esté, detrás de «Más avisos». -->
		<div class="tarjeta" use:anim={{ t: 'rise', delay: 30 }}>
			<div class="activo">
				<span class="a-ico" style:color={CANAL.email[1]}
					><Icon icon={CANAL.email[0]} size={20} /></span
				>
				<span class="a-txt"
					><b>Venta nueva por email</b><span class="d"
						>Por transferencia apenas entra el pedido; por Mercado Pago cuando se acredita. Llega al
						email de <a href="{resolve('/panel/configuracion')}?tab=negocio">tu negocio</a> o, si no hay,
						al de tu&nbsp;cuenta.</span
					></span
				>
				<Interruptor
					checked={activo(prefs.matriz, 'venta', 'email')}
					label="Venta nueva por email"
					onchange={() => alternar('venta', 'email')}
				/>
			</div>
			<div class="activo">
				<span class="a-ico" style:color={CANAL.celular[1]}
					><Icon icon={CANAL.celular[0]} size={20} /></span
				>
				<span class="a-txt"
					><b>Push en el iPhone</b><span class="d"
						>«Nueva venta» al instante, en el iPhone de quien entró a la app de
						Tu&nbsp;Local&nbsp;Web. El monto lo ve sólo quien tiene acceso a&nbsp;finanzas.</span
					></span
				>
				<Interruptor
					checked={activo(prefs.matriz, 'venta', 'celular')}
					label="Venta nueva por push en el iPhone"
					onchange={() => alternar('venta', 'celular')}
				/>
			</div>
			<div class="activo">
				<span class="a-ico" style:color="#B83A2E"
					><Icon icon={I.phWarningCircleFill} size={20} /></span
				>
				<span class="a-txt"
					><b>Sin stock</b><span class="d"
						>Cuando una venta deja un producto en 0, con un botón para reponer. Uno por día como
						mucho, a vos y a quien edita&nbsp;productos.</span
					><span class="canales-a">
						{#each ['email', 'celular'] as const as c (c)}
							<span class="c-a"
								><Icon icon={CANAL[c][0]} size={15} style="color:{CANAL[c][1]}" />{c === 'celular'
									? 'iPhone'
									: CANAL[c][2]}<Interruptor
									checked={activo(prefs.matriz, 'sinstock', c)}
									label="Sin stock por {c === 'celular' ? 'push en el iPhone' : 'email'}"
									onchange={() => alternar('sinstock', c)}
								/></span
							>
						{/each}
					</span></span
				>
			</div>
			<a class="activo" href={resolve('/panel/integraciones')}>
				<span class="a-ico" style:color="#229ED9"><Icon icon={I.logosTelegram} size={20} /></span>
				<span class="a-txt"
					><b>Venta nueva por Telegram</b><span class="d">Se vincula desde Integraciones</span
					></span
				>
				<Icon icon={I.phCaretRight} size={14} style="color:var(--fa)" />
			</a>
		</div>
		<button type="button" class="mas" aria-expanded={verMas} onclick={() => (verMas = !verMas)}
			>{verMas ? 'Ocultar' : 'Más avisos'}<span class="d"
				>Stock bajo, envíos, ARCA y más: los guardamos para cuando&nbsp;salgan</span
			><Icon icon={verMas ? I.phCaretUpBold : I.phCaretDownBold} size={14} /></button
		>
	{/if}

	{#if demo || verMas}
		<div class="tarjeta" use:anim={{ t: 'rise', delay: 50 }}>
			<div class="desliza" use:deslizable>
				<div class="tabla">
					<div class="fila cab-t">
						<span>Aviso</span>
						{#each CANALES_AVISO as c (c)}
							{@const [icono, col, t] = CANAL[c]}
							<span class="canal"><Icon icon={icono} size={18} style="color:{col}" />{t}</span>
						{/each}
					</div>
					{#each AVISOS as g (g.grupo)}
						<div class="grupo">{g.grupo}</div>
						{#each g.filas as f (f.id)}
							<div class="fila">
								<span><b>{f.t}</b><span class="d">{f.d}</span></span>
								{#each CANALES_AVISO as c (c)}
									<span class="celda"
										><span class="c-mob" aria-hidden="true"
											><Icon icon={CANAL[c][0]} size={15} style="color:{CANAL[c][1]}" />{CANAL[
												c
											][2]}</span
										><Interruptor
											checked={activo(prefs.matriz, f.id, c)}
											label="{f.t} por {CANAL[c][2]}"
											onchange={() => alternar(f.id, c)}
										/></span
									>
								{/each}
							</div>
						{/each}
					{/each}
				</div>
			</div>
		</div>
	{/if}

	{#if demo}
		<div class="dos" use:anim={{ t: 'rise', delay: 100 }}>
			<div class="caja">
				<div class="t">Resumen diario</div>
				<div class="d2">
					Tincho te manda a las 9 h cuánto vendiste ayer, qué pedidos esperan y qué se está quedando
					sin stock.
				</div>
				<div class="pills" role="group" aria-label="Resumen diario">
					{#each [['whatsapp', 'Por WhatsApp'], ['email', 'Por email'], ['off', 'No quiero']] as const as [k, t] (k)}
						<button type="button" aria-pressed={prefs.resumen === k} onclick={() => resumen(k)}
							>{t}</button
						>
					{/each}
				</div>
			</div>
			<div class="caja">
				<div class="t">Probar avisos</div>
				<div class="d2">Te mandamos un aviso de prueba para que veas cómo llega.</div>
				<button
					type="button"
					class="probar"
					onclick={() =>
						ui.toast({
							k: 'ok',
							title: 'Nueva venta · $ 44.600',
							desc: 'Aviso de prueba: así te llega cada venta al panel'
						})}><Icon icon={I.phBellRingingBold} size={15} />Mandar prueba</button
				>
			</div>
		</div>
	{/if}
</div>

<style>
	.activo {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 64px;
		padding: 12px 18px;
		color: var(--tx);
		text-decoration: none;
	}
	.canales-a {
		display: flex;
		flex-wrap: wrap;
		gap: 8px 20px;
		margin-top: 10px;
	}
	.c-a {
		display: inline-flex;
		align-items: center;
		gap: 7px;
		font-size: 13px;
		font-weight: 600;
	}
	.activo + .activo {
		border-top: 1px solid var(--ln);
	}
	.a-ico {
		flex: none;
		width: 38px;
		height: 38px;
		border-radius: 12px;
		background: var(--sf3);
		display: grid;
		place-items: center;
	}
	.a-txt {
		flex: 1;
		min-width: 0;
		font-size: 14px;
	}
	.a-txt b {
		font-weight: 600;
	}
	.a-txt a {
		color: var(--terracota);
		font-weight: 600;
	}
	.mas {
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: 48px;
		padding: 10px 18px;
		border-radius: 18px;
		border: 1px solid var(--ln);
		background: transparent;
		color: var(--tx);
		font: inherit;
		font-size: 14px;
		font-weight: 600;
		text-align: left;
		cursor: pointer;
	}
	.mas .d {
		flex: 1;
		font-weight: 400;
	}
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.cab {
		display: flex;
		align-items: end;
		gap: 12px;
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
	.silencio {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 14px;
		border-radius: 999px;
		background: var(--sf);
		box-shadow: var(--shd);
		font-size: 13.5px;
	}
	.tarjeta {
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		overflow: hidden;
	}
	.desliza {
		overflow-x: auto;
	}
	.tabla {
		min-width: 560px;
	}
	.fila {
		display: grid;
		grid-template-columns: minmax(200px, 2fr) repeat(4, minmax(70px, 1fr));
		padding: 10px 18px;
		border-top: 1px solid var(--ln);
		align-items: center;
		font-size: 14px;
	}
	.fila b {
		font-weight: 600;
	}
	.d {
		display: block;
		font-size: 12px;
		color: var(--mu);
	}
	.cab-t {
		border-top: 0;
		padding: 12px 18px;
		background: var(--sf2);
		font-size: 12px;
		font-weight: 700;
		color: var(--mu);
	}
	.cab-t > span:first-child {
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}
	.canal {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
	}
	.grupo {
		padding: 10px 18px;
		font-size: 12.5px;
		font-weight: 700;
		color: var(--mu);
		border-top: 1px solid var(--ln);
	}
	.celda {
		display: grid;
		place-items: center;
	}
	.c-mob {
		display: none;
	}
	/* Celular: cada aviso es una tarjeta con sus cuatro canales debajo, con nombre e ícono (sin
	   tabla ancha ni scroll escondido). */
	@media (max-width: 760px) {
		.desliza {
			overflow: visible;
		}
		.tabla {
			min-width: 0;
		}
		.cab-t {
			display: none;
		}
		.grupo {
			padding: 14px 16px 6px;
		}
		.fila {
			grid-template-columns: repeat(4, minmax(0, 1fr));
			row-gap: 10px;
			padding: 12px 16px 14px;
		}
		.fila > span:first-child {
			grid-column: 1 / -1;
		}
		.celda {
			gap: 6px;
		}
		.c-mob {
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 2px;
			font-size: 11.5px;
			font-weight: 600;
			color: var(--mu);
		}
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr));
		gap: 12px;
	}
	.caja {
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.t {
		font-size: 15.5px;
		font-weight: 600;
	}
	.d2 {
		font-size: 13.5px;
		color: var(--mu);
	}
	.pills {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.pills button {
		padding: 8px 13px;
		border-radius: 999px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 13px 'Instrument Sans',
			sans-serif;
	}
	.pills button[aria-pressed='true'] {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	.probar {
		align-self: flex-start;
		display: flex;
		align-items: center;
		gap: 7px;
		padding: 10px 16px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
</style>
