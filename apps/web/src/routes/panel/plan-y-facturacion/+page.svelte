<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { MOTIVOS } from '#lib/panel/permisos.ts';
	import { PLANES, precioAnual } from '#lib/panel/planes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Plan y facturación (design/PanelCuenta.dc.html, sPlan): planes, uso, medio de pago y cancelar. */
	let { data } = $props();
	const ui = getUi();

	const fecha = (iso: string) =>
		new Date(iso).toLocaleDateString('es-AR', {
			day: 'numeric',
			month: 'long',
			timeZone: 'America/Argentina/Buenos_Aires'
		});
	let ciclo = $derived(data.ciclo);
	const anual = $derived(ciclo === 'anio');
	const indice = (id: string) => PLANES.findIndex((p) => p.id === id);
	const actual = $derived(PLANES[indice(data.actual)]);

	async function pedir(url: string, method: string, cuerpo?: unknown) {
		const r = await fetch(url, {
			method,
			headers: cuerpo === undefined ? {} : { 'content-type': 'application/json' },
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!r.ok) ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo.' });
		return r.ok ? r : null;
	}
	async function cambiar(id: (typeof PLANES)[number]['id'], c = ciclo) {
		const antes = data.actual;
		const antesCiclo = data.ciclo;
		const r0 = await pedir('/panel/plan', 'PUT', { plan: id, ciclo: c });
		if (!r0) return false;
		// Plan pago con el abono real: se carga la tarjeta en Mercado Pago.
		if (r0.status === 200) {
			location.href = ((await r0.json()) as { url: string }).url;
			return false;
		}
		await invalidateAll();
		return { antes, antesCiclo };
	}
	async function elegir(p: (typeof PLANES)[number]) {
		if (p.id === data.actual) return;
		const sube = indice(p.id) > indice(data.actual);
		const r = await cambiar(p.id);
		if (!r) return;
		ui.toast({
			k: 'ok',
			title: `Cambiaste al plan ${p.nombre}`,
			desc: data.enPrueba
				? 'Se cobra cuando termina tu prueba'
				: sube
					? 'Ya tenés todo activo. Cobramos la diferencia prorrateada.'
					: 'Se aplica desde el próximo período.',
			action: 'Deshacer',
			onAction: () => void cambiar(r.antes, r.antesCiclo)
		});
	}
	async function cicloA(c: 'mes' | 'anio') {
		ciclo = c;
		await cambiar(data.actual, c);
	}

	// ---- Uso ----
	const tope = $derived(actual.hornero + data.uso.packs * 2000);
	const USO = $derived([
		['Productos', data.uso.productos, null],
		['Personas en el equipo', data.uso.personas, actual.personas],
		['Respuestas de Hornero con IA', data.uso.hornero, tope],
		['Respuestas automáticas sin IA', data.uso.sinIa, null]
	] as const);
	async function pack() {
		if (!(await pedir('/panel/cuenta/pack', 'POST'))) return;
		await invalidateAll();
		ui.toast({
			k: 'ai',
			title: '2.000 respuestas más para Hornero',
			desc: 'Se suman a este mes · $ 4.900 en tu próxima factura',
			action: 'Deshacer',
			onAction: async () => {
				await pedir('/panel/cuenta/pack', 'DELETE');
				await invalidateAll();
			}
		});
	}

	// ---- Frenar y cancelar ----
	async function aGratis() {
		const r = await cambiar('gratis');
		if (!r) return;
		ui.toast({
			k: 'info',
			title: `Pasás al plan Gratis el ${fecha(data.proximoCobro)}`,
			desc: 'Conservás productos, clientes y pedidos',
			action: 'Deshacer',
			onAction: () => void cambiar(r.antes, r.antesCiclo)
		});
	}
	let cancelando = $state(false);
	let paso = $state(0);
	let motivo = $state(0);
	const OFERTAS = $derived([
		[
			'Te dejamos 3 meses a mitad de precio',
			`Sabemos que no siempre es fácil. Seguí con todo por ${pesos(actual.precio / 2)}/mes durante 3 meses.`,
			I.phPercentBold,
			'50 % off por 3 meses',
			'Se aplica en tu próxima factura',
			'Quiero el descuento'
		],
		[
			'Contanos qué te falta',
			'Muchas funciones se pueden activar o están por salir. Una persona del equipo te escribe hoy.',
			I.phChatsCircleBold,
			'Llamada de 15 minutos',
			'Te escribimos por WhatsApp',
			'Que me contacten'
		],
		[
			'Mejor pausala',
			'Tu tienda queda guardada y la reactivás cuando quieras, sin cargar nada de nuevo.',
			I.phPauseCircleBold,
			'Pausa hasta 3 meses',
			'Gratis: pasás al plan Gratis y conservás todo',
			'Pausar en vez de cancelar'
		],
		[
			'Antes de irte, compará',
			'Te mostramos cuánto pagarías en la otra plataforma con tus ventas reales.',
			I.phScalesBold,
			'Comparación con tus números',
			'Comisiones, apps y abono',
			'Ver la comparación'
		],
		[
			'Te ayudamos igual',
			'Si algo no anduvo bien, queremos saberlo.',
			I.phHandHeartBold,
			'Primer mes gratis si te quedás',
			'Sin compromiso',
			'Me quedo un mes más'
		]
	] as const);
	const oferta = $derived(OFERTAS[motivo]);
	function abrirCancelar() {
		paso = 0;
		motivo = 0;
		cancelando = true;
	}
	async function aceptarOferta() {
		if (motivo === 2) {
			cancelando = false;
			return aGratis();
		}
		if (!(await pedir('/panel/cuenta/cancelacion', 'POST', { motivo, accion: 'oferta' }))) return;
		cancelando = false;
		ui.toast({ k: 'ok', title: '¡Gracias por quedarte!', desc: `${oferta[3]} · ${oferta[4]}` });
	}
	async function confirmar() {
		const r = await pedir('/panel/cuenta/cancelacion', 'POST', { motivo, accion: 'cancelar' });
		if (!r) return;
		const { hasta } = (await r.json()) as { hasta: string };
		cancelando = false;
		await invalidateAll();
		ui.toast({
			k: 'warn',
			title: 'Plan cancelado',
			desc: `Tu tienda sigue online hasta el ${fecha(hasta)}`,
			action: 'Deshacer',
			onAction: async () => {
				await pedir('/panel/cuenta/cancelacion', 'DELETE');
				await invalidateAll();
			}
		});
	}
	const CONSECUENCIAS = $derived([
		[I.phCalendarCheck, `Seguís usando Mostrador hasta el ${fecha(data.proximoCobro)}`, '#2F8F63'],
		[I.phStorefront, 'Después la tienda queda offline, pero no se borra', 'var(--mu)'],
		[I.phDatabase, 'Guardamos todo 90 días por si volvés', 'var(--mu)'],
		[I.phXCircle, 'No te volvemos a cobrar', '#2F8F63']
	] as const);
</script>

<Seo
	title="Plan y facturación · Panel"
	description="Tu plan de Mostrador"
	path="/panel/plan-y-facturacion"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Plan y facturación</h1>
			<div class="c-s">
				Estás en <b>plan {actual.nombre} {anual ? 'anual' : 'mensual'}</b>{data.enPrueba
					? ' (en prueba)'
					: ''} ·
				{data.cancelado
					? `se da de baja el ${fecha(data.cancelado.hasta)}`
					: `próximo cobro el ${fecha(data.proximoCobro)}`}
			</div>
		</div>
		<div class="ciclos" role="group" aria-label="Ciclo">
			<button type="button" aria-pressed={!anual} onclick={() => cicloA('mes')}>Mensual</button>
			<button type="button" aria-pressed={anual} onclick={() => cicloA('anio')}
				>Anual<span>−25 %</span></button
			>
		</div>
	</div>

	<div class="planes" use:anim={{ t: 'rise', delay: 50 }}>
		{#each PLANES as p (p.id)}
			{@const cur = p.id === data.actual}
			{@const oscuro = p.id === 'pro'}
			{@const precio = anual ? precioAnual(p.precio) : p.precio}
			<div class="plan" class:oscuro class:cur>
				{#if cur}<span class="tuyo">Tu plan</span>{/if}
				<div>
					<div class="p-n">{p.nombre}</div>
					<div class="p-d">{p.lema}</div>
				</div>
				<div class="p-precio">
					<b>{p.precio ? pesos(precio) : 'Gratis'}</b>{#if p.precio}<span>/ mes</span>{/if}
				</div>
				{#if anual && p.precio}<div class="p-anual">
						{pesos(precio * 12)} por año · ahorrás {pesos(p.precio * 12 * 0.25)}
					</div>{/if}
				<div class="p-incluye">
					{#each p.incluye as f (f)}<div><Icon icon={I.phCheckBold} size={14} />{f}</div>{/each}
				</div>
				<button type="button" class="p-btn" class:actual={cur} onclick={() => elegir(p)}
					>{cur
						? 'Tu plan actual'
						: indice(p.id) > indice(data.actual)
							? `Pasar a ${p.nombre}`
							: `Bajar a ${p.nombre}`}</button
				>
			</div>
		{/each}
	</div>

	<div class="dos" use:anim={{ t: 'rise', delay: 100 }}>
		<div class="tarjeta uso">
			<div class="u-cab">
				<span>Uso de este mes</span>
				<button type="button" onclick={pack}
					><Icon icon={I.phSparkleFill} size={13} style="color:#b0441d" />+2.000 respuestas · $
					4.900</button
				>
			</div>
			{#each USO as [t, v, max] (t)}
				{@const w = max ? Math.min(100, (v / max) * 100) : 18}
				<div>
					<div class="u-f">
						<span>{t}</span><span
							><b>{v.toLocaleString('es-AR')}</b> de {max
								? max.toLocaleString('es-AR')
								: 'ilimitado'}</span
						>
					</div>
					<div class="u-b">
						<div
							style:width="{w}%"
							style:background={w >= 100 ? '#B83A2E' : w > 80 ? '#E3A93B' : '#D2582F'}
						></div>
					</div>
				</div>
			{/each}
		</div>
		<div class="tarjeta medio">
			<div class="u-cab">
				<span>Medio de pago</span>
				<button
					type="button"
					class="link"
					onclick={() =>
						ui.toast({
							k: 'info',
							title: 'Cambiar medio de pago',
							desc: 'Se hace en Mercado Pago cuando se conecte el cobro del abono'
						})}>Cambiar</button
				>
			</div>
			{#if data.medio}
				<div class="tarjeta-credito">
					<span class="visa"><Icon icon={I.logosVisa} size={30} /></span>
					<div>
						<div class="t-n">•••• {data.medio.ultimos}</div>
						<div class="t-v">Vence {data.medio.vence} · débito automático</div>
					</div>
				</div>
			{/if}
			<div class="fact">
				<span>Datos de facturación</span>
				<span
					><b>{data.facturacion.razonSocial}</b><br /><span
						>CUIT {data.facturacion.cuit} · {data.facturacion.tipo}</span
					></span
				>
			</div>
		</div>
	</div>

	<div class="tarjeta facturas" use:anim={{ t: 'rise', delay: 150 }}>
		<div class="f-cab"><span>Tus facturas de Mostrador</span></div>
		{#each data.facturas as f (f.numero)}
			<div class="factura">
				<span class="f-ico"><Icon icon={I.phReceipt} size={16} /></span>
				<span class="f-txt"
					><span class="f-n">{f.numero}</span><span class="f-d">{f.detalle}</span></span
				>
				<span class="pagada">Pagada</span>
				<b>{pesos(f.total)}</b>
			</div>
		{:else}
			<div class="vacio">
				{data.enPrueba
					? `Estás en la prueba: la primera factura llega el ${fecha(data.proximoCobro)}.`
					: 'Todavía no hay facturas.'}
			</div>
		{/each}
	</div>

	<div class="frenar" use:anim={{ t: 'rise', delay: 200 }}>
		<Icon icon={I.phPauseCircle} size={26} style="color:var(--mu)" />
		<div class="fr-txt">
			<div class="fr-t">{data.cancelado ? 'Tu plan está cancelado' : '¿Necesitás frenar?'}</div>
			<div class="fr-d">
				{data.cancelado
					? `Seguís hasta el ${fecha(data.cancelado.hasta)}. Si cambiás de idea, lo reactivás acá.`
					: 'Podés pasar al plan Gratis sin perder nada, o cancelar. Tus datos quedan guardados.'}
			</div>
		</div>
		{#if data.cancelado}
			<button
				type="button"
				class="fr-a"
				onclick={async () => {
					await pedir('/panel/cuenta/cancelacion', 'DELETE');
					await invalidateAll();
					ui.toast({ k: 'ok', title: 'Plan reactivado', desc: 'Seguís como hasta ahora' });
				}}>Reactivar</button
			>
		{:else}
			{#if data.actual !== 'gratis'}<button type="button" class="fr-a" onclick={aGratis}
					>Pasar a Gratis</button
				>{/if}
			<button type="button" class="fr-c" onclick={abrirCancelar}>Cancelar plan</button>
		{/if}
	</div>
</div>

<Modal bind:open={cancelando} titulo="Cancelar el plan" ancho={520} padding={24}>
	<div class="pasos">
		{#each [0, 1, 2] as i (i)}<span class:on={i <= paso}></span>{/each}
	</div>
	{#if paso === 0}
		<div class="m-cab">
			<img src="/hornero/preocupado.webp" alt="" width="64" height="64" />
			<div>
				<b>¿Por qué te vas?</b>
				<div class="m-s">Nos ayuda a mejorar. Es opcional.</div>
			</div>
		</div>
		<div class="motivos" role="radiogroup" aria-label="Motivo">
			{#each MOTIVOS as m, i (m)}
				<button type="button" role="radio" aria-checked={motivo === i} onclick={() => (motivo = i)}
					><span class="radio"></span>{m}</button
				>
			{/each}
		</div>
		<div class="m-acc">
			<button type="button" class="b-sec" onclick={() => (cancelando = false)}>Me quedo</button>
			<button type="button" class="b-osc" onclick={() => (paso = 1)}>Continuar</button>
		</div>
	{:else if paso === 1}
		<b class="m-t">{oferta[0]}</b>
		<div class="m-d">{oferta[1]}</div>
		<div class="oferta">
			<Icon icon={oferta[2]} size={28} style="color:#b0441d" />
			<div>
				<b>{oferta[3]}</b>
				<div class="m-s">{oferta[4]}</div>
			</div>
		</div>
		<div class="m-acc envolver">
			<button type="button" class="b-pri" onclick={aceptarOferta}>{oferta[5]}</button>
			<button type="button" class="b-sec" onclick={() => (paso = 2)}>Cancelar igual</button>
		</div>
	{:else}
		<b class="m-t">Confirmá la cancelación</b>
		<div class="consecuencias">
			{#each CONSECUENCIAS as [icono, t, c] (t)}
				<div><Icon icon={icono} size={17} style="color:{c};flex:none;margin-top:1px" />{t}</div>
			{/each}
		</div>
		<div class="m-acc">
			<button type="button" class="b-sec" onclick={() => (cancelando = false)}>Volver</button>
			<button type="button" class="b-roja" onclick={confirmar}>Cancelar mi plan</button>
		</div>
	{/if}
</Modal>

<style>
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
		text-wrap: pretty;
	}
	.c-s b {
		color: var(--tx);
	}
	.ciclos {
		display: flex;
		padding: 4px;
		border-radius: 999px;
		background: var(--chip);
	}
	.ciclos button {
		border: 0;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		padding: 8px 14px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
		white-space: nowrap;
	}
	.ciclos button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.ciclos span {
		margin-left: 6px;
		font-size: 11px;
		font-weight: 700;
		color: #237550;
	}
	.planes {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 250px), 1fr));
		gap: 12px;
	}
	.plan {
		position: relative;
		padding: 20px;
		border-radius: 24px;
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--shd);
		border: 2px solid transparent;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.plan.oscuro {
		background: #2a140c;
		color: #fbf7f2;
	}
	.plan.cur {
		border-color: #d2582f;
	}
	.tuyo {
		position: absolute;
		top: -11px;
		left: 18px;
		padding: 4px 10px;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
		font-size: 11.5px;
		font-weight: 700;
	}
	.p-n {
		font-size: 18px;
		font-weight: 600;
	}
	.p-d {
		font-size: 13.5px;
		opacity: 0.75;
		margin-top: 2px;
	}
	.p-precio {
		display: flex;
		align-items: baseline;
		gap: 6px;
	}
	.p-precio b {
		font-size: 32px;
		letter-spacing: -0.04em;
	}
	.p-precio span,
	.p-anual {
		font-size: 13.5px;
		opacity: 0.75;
	}
	.p-anual {
		font-size: 12.5px;
		margin-top: -8px;
		opacity: 0.8;
	}
	.p-incluye {
		display: flex;
		flex-direction: column;
		gap: 7px;
	}
	.p-incluye div {
		display: flex;
		gap: 8px;
		font-size: 13.5px;
		line-height: 1.4;
	}
	.p-incluye :global(svg) {
		flex: none;
		margin-top: 2px;
		color: #b0441d;
	}
	.oscuro .p-incluye :global(svg) {
		color: #f0a58a;
	}
	.p-btn {
		margin-top: auto;
		padding: 12px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font:
			600 14px 'Instrument Sans',
			sans-serif;
	}
	.oscuro .p-btn {
		background: #b0441d;
		color: #fff;
	}
	.p-btn.actual {
		background: transparent;
		color: var(--mu);
		border: 1.5px solid var(--ln2);
	}
	.oscuro .p-btn.actual {
		color: #fbf7f2;
		border-color: rgba(251, 247, 242, 0.3);
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr));
		gap: 12px;
		align-items: start;
	}
	.tarjeta {
		padding: 18px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.uso,
	.medio {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.u-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 10px;
		font-size: 16px;
		font-weight: 600;
	}
	.u-cab button {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 7px 12px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.u-cab button.link {
		border: 0;
		padding: 0;
		font-size: 13px;
		color: #b0441d;
	}
	.u-f {
		display: flex;
		justify-content: space-between;
		gap: 8px;
		font-size: 13.5px;
	}
	.u-f span:last-child {
		color: var(--mu);
		white-space: nowrap;
	}
	.u-f b {
		color: var(--tx);
	}
	.u-b {
		margin-top: 6px;
		height: 7px;
		border-radius: 4px;
		background: var(--chip);
		overflow: hidden;
	}
	.u-b div {
		height: 100%;
		border-radius: 4px;
	}
	.tarjeta-credito {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px;
		border-radius: 16px;
		background: linear-gradient(135deg, #2a140c, #5a3424);
		color: #fff;
	}
	.visa {
		display: grid;
		place-items: center;
		background: #fff;
		border-radius: 5px;
		padding: 3px 5px;
	}
	.t-n {
		font-size: 15px;
		font-weight: 600;
		letter-spacing: 0.06em;
	}
	.t-v {
		font-size: 12.5px;
		opacity: 0.75;
	}
	.fact {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 10px;
		font-size: 13.5px;
	}
	.fact > span:first-child,
	.fact span span {
		color: var(--mu);
	}
	.fact > span:last-child {
		text-align: right;
	}
	.facturas {
		padding: 0;
		overflow: hidden;
	}
	.f-cab {
		padding: 14px 18px;
		border-bottom: 1px solid var(--ln);
		font-size: 16px;
		font-weight: 600;
	}
	.factura {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 18px;
		border-bottom: 1px solid var(--ln);
		flex-wrap: wrap;
	}
	.f-ico {
		width: 34px;
		height: 34px;
		border-radius: 10px;
		background: var(--chip);
		display: grid;
		place-items: center;
	}
	.f-txt {
		flex: 1;
		min-width: 160px;
	}
	.f-n {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.f-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.pagada {
		font-size: 12px;
		font-weight: 700;
		padding: 3px 9px;
		border-radius: 999px;
		background: rgba(47, 143, 99, 0.12);
		color: #237550;
	}
	.vacio {
		padding: 18px;
		font-size: 14px;
		color: var(--mu);
	}
	.frenar {
		display: flex;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
		padding: 16px 18px;
		border-radius: 22px;
		border: 1px dashed var(--ln2);
	}
	.fr-txt {
		flex: 1;
		min-width: 220px;
	}
	.fr-t {
		font-size: 15px;
		font-weight: 600;
	}
	.fr-d {
		font-size: 13.5px;
		color: var(--mu);
	}
	.fr-a {
		padding: 10px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.fr-c {
		padding: 10px 16px;
		border-radius: 999px;
		border: 0;
		background: transparent;
		color: #b83a2e;
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.pasos {
		display: flex;
		gap: 6px;
		margin-bottom: 16px;
	}
	.pasos span {
		flex: 1;
		height: 4px;
		border-radius: 2px;
		background: var(--ln2);
	}
	.pasos span.on {
		background: #d2582f;
	}
	.m-cab {
		display: flex;
		gap: 12px;
		align-items: center;
	}
	.m-cab img {
		object-fit: contain;
	}
	.m-cab b,
	.m-t {
		font-size: 19px;
	}
	.m-s {
		font-size: 13.5px;
		color: var(--mu);
	}
	.motivos {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.motivos button {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 14px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			500 14px 'Instrument Sans',
			sans-serif;
		text-align: left;
	}
	.radio {
		width: 18px;
		height: 18px;
		border-radius: 50%;
		border: 2px solid var(--ln2);
		display: grid;
		place-items: center;
		flex: none;
	}
	.motivos button[aria-checked='true'] {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.06);
	}
	.motivos button[aria-checked='true'] .radio {
		border-color: #d2582f;
	}
	.motivos button[aria-checked='true'] .radio::after {
		content: '';
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #d2582f;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14px;
		line-height: 1.5;
		color: var(--mu);
	}
	.oferta {
		margin-top: 14px;
		padding: 16px;
		border-radius: 18px;
		background: linear-gradient(135deg, rgba(210, 88, 47, 0.12), rgba(210, 88, 47, 0.04));
		border: 1px solid rgba(210, 88, 47, 0.25);
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.oferta b {
		font-size: 15px;
	}
	.consecuencias {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 14px;
	}
	.consecuencias div {
		display: flex;
		gap: 10px;
		align-items: flex-start;
	}
	.m-acc {
		margin-top: 16px;
		display: flex;
		gap: 8px;
	}
	.m-acc.envolver {
		flex-wrap: wrap;
	}
	.m-acc button {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
	}
	.envolver .b-pri {
		flex: 1 1 180px;
	}
	.envolver .b-sec {
		flex: 1 1 140px;
	}
	.b-sec {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.b-osc {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.b-pri {
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
	}
	.b-roja {
		border: 0;
		background: #b83a2e;
		color: #fff;
	}
</style>
