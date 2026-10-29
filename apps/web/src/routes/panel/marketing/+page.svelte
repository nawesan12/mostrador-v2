<script lang="ts">
	import { resolve } from '$app/paths';
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Marketing (design/PanelCrecer.dc.html, sMkt): campaña por WhatsApp con el costo exacto. */
	let { data } = $props();
	const ui = getUi();

	let audiencia = $state(0);
	let iMensaje = $state(0);
	// El texto que se edita arranca con la propuesta de Tincho y cambia al reescribir.
	let mensaje = $derived(data.mensajes[iMensaje % data.mensajes.length] ?? '');
	let foto = $state(true);
	let boton = $state(true);
	let cupon = $state(true);
	const aud = $derived(data.audiencias[audiencia]);
	const costo = $derived(aud.cantidad * data.costoMensaje);
	const vista = $derived(mensaje.replaceAll('{nombre}', 'Lucía'));
	let enviando = $state(false);

	async function enviar(programar: boolean) {
		if (enviando) return;
		enviando = true;
		const r = await fetch('/panel/marketing/campanas', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				audiencia: aud.id,
				mensaje,
				foto,
				boton,
				cupon: cupon && !!data.cupon,
				programar
			})
		});
		enviando = false;
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({ k: 'err', title: 'No se mandó la campaña', desc: m ?? 'Probá de nuevo.' });
		}
		await invalidateAll();
		if (programar)
			ui.toast({
				k: 'info',
				title: 'Campaña programada',
				desc: 'Sábado 10:00 · el mejor horario de tus clientes'
			});
		else
			ui.toast({
				k: 'wa',
				title: 'Campaña en camino',
				desc: `${aud.cantidad} mensajes · ${pesos(costo)} · te avisamos las ventas`
			});
	}
	const ADJUNTOS = $derived([
		['foto', I.phImageFill, 'Foto'],
		['boton', I.phCursorClickFill, 'Botón al producto'],
		...(data.cupon ? ([['cupon', I.phTicketFill, `Cupón ${data.cupon}`]] as const) : [])
	] as const);
	const adjunto = (k: string) => (k === 'foto' ? foto : k === 'boton' ? boton : cupon);
	function alternar(k: string) {
		if (k === 'foto') foto = !foto;
		else if (k === 'boton') boton = !boton;
		else cupon = !cupon;
	}
</script>

<Seo
	title="Marketing · Panel"
	description="Campañas por WhatsApp y email"
	path="/panel/marketing"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Marketing</h1>
			<div class="c-s">
				Campañas por WhatsApp, email y cupones, con el costo exacto antes de enviar
			</div>
		</div>
	</div>

	<div class="arriba" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="tarjeta nueva">
			<div class="n-cab">
				<span class="n-ico"><Icon icon={I.logosWhatsappIcon} size={20} /></span><span
					>Nueva campaña</span
				>
			</div>
			<div>
				<div class="et">¿A quién?</div>
				<div class="pills" role="group" aria-label="¿A quién?">
					{#each data.audiencias as a, i (a.id)}
						<button
							type="button"
							class="pill"
							aria-pressed={audiencia === i}
							onclick={() => (audiencia = i)}>{a.nombre} <span>{a.cantidad}</span></button
						>
					{/each}
				</div>
			</div>
			<div>
				<div class="et fila">
					Mensaje<button type="button" class="reescribir" onclick={() => iMensaje++}
						>✦ Reescribir con Tincho</button
					>
				</div>
				<textarea bind:value={mensaje} maxlength="1000" aria-label="Mensaje"></textarea>
			</div>
			<div class="pills">
				{#each ADJUNTOS as [k, icono, t] (k)}
					<button
						type="button"
						class="pill chica"
						aria-pressed={adjunto(k)}
						onclick={() => alternar(k)}
						><Icon icon={icono} size={14} style="color:#b0441d" />{t}</button
					>
				{/each}
			</div>
			<div class="numeros">
				<span
					><span class="nu-t">Costo exacto</span><span class="nu-v">{pesos(costo)}</span><span
						class="nu-d">{aud.cantidad} mensajes de marketing × {pesos(data.costoMensaje)}</span
					></span
				>
				<span
					><span class="nu-t">Ventas estimadas</span><span class="nu-v verde"
						>{data.retorno ? pesos(costo * data.retorno) : '—'}</span
					><span class="nu-d">según tus campañas anteriores</span></span
				>
			</div>
			<div class="acc">
				<button
					type="button"
					class="programar"
					disabled={enviando || !aud.cantidad}
					onclick={() => enviar(true)}>Programar</button
				>
				<button
					type="button"
					class="enviar"
					disabled={enviando || !aud.cantidad || mensaje.trim().length < 10}
					onclick={() => enviar(false)}>Enviar a {aud.cantidad} personas</button
				>
			</div>
		</div>

		<div class="tel-col">
			<div class="tel">
				<div class="pantalla">
					<div class="t-cab">
						{#if data.tienda.foto}<img src={unsplash(data.tienda.foto, 80, '&h=80')} alt="" />{/if}
						<span>{data.tienda.nombre}</span><Icon
							icon={I.phSealCheckFill}
							size={14}
							style="color:#237550"
						/>
					</div>
					<div class="chat">
						{#key iMensaje}
							<div class="burbuja" use:anim={{ t: 'rise' }}>
								{#if foto && data.producto?.foto}<img
										src={unsplash(data.producto.foto, 480, '&h=260')}
										alt=""
									/>{/if}
								<div class="b-txt">{vista}</div>
								{#if boton && data.producto}<div class="b-btn">
										🛍️ Ver {data.producto.nombre.toLowerCase()}
									</div>{/if}
							</div>
						{/key}
					</div>
				</div>
			</div>
			<span class="t-pie">Vista previa en WhatsApp</span>
		</div>
	</div>

	<div class="campanas" use:anim={{ t: 'rise', delay: 120 }}>
		{#each data.campanas as c (c.id)}
			<div class="tarjeta campana">
				<div class="ca-cab">
					<Icon icon={c.canal === 'wa' ? I.logosWhatsappIcon : I.phEnvelopeSimpleFill} size={18} />
					<span class="ca-t">{c.titulo}</span>
					<span class="estado" class:activa={c.estado === 'Activa' || c.estado === 'Enviando'}
						>{c.estado}</span
					>
				</div>
				<div class="ca-d">{c.detalle}</div>
				<div class="metricas">
					{#each c.metricas as m (m.k)}
						<div>
							<div class="m-k">{m.k}</div>
							<div class="m-v">{m.v}</div>
						</div>
					{/each}
				</div>
			</div>
		{/each}
	</div>

	<div class="tarjeta cupones" use:anim={{ t: 'rise', delay: 160 }}>
		<div class="cu-cab">
			<span>Cupones activos</span>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta del panel con query -->
			<a href="{resolve('/panel/cupones')}?nuevo">+ Nuevo cupón</a>
		</div>
		<div class="cu-lista">
			{#each data.cupones as c (c.codigo)}
				<div class="cupon">
					<span class="codigo">{c.codigo}</span>
					<span class="cu-txt"
						><span class="cu-t">{c.descripcion}</span><span class="cu-u">{c.usos}</span></span
					>
				</div>
			{:else}
				<div class="vacio">No tenés cupones activos.</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
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
	.arriba {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 300px;
		gap: 16px;
		align-items: start;
	}
	.tarjeta {
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.nueva {
		padding: 20px;
		border-radius: 24px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.n-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 17px;
		font-weight: 600;
	}
	.n-ico {
		width: 38px;
		height: 38px;
		border-radius: 12px;
		background: #e9f8ee;
		display: grid;
		place-items: center;
	}
	.et {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		margin-bottom: 6px;
	}
	.et.fila {
		display: flex;
		justify-content: space-between;
	}
	.reescribir {
		border: 0;
		background: transparent;
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		color: #b0441d;
	}
	.pills {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.pill {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 8px 13px;
		border-radius: 999px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.pill span {
		opacity: 0.6;
	}
	.pill[aria-pressed='true'] {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	.pill.chica {
		gap: 6px;
		padding: 7px 11px;
		border-width: 1px;
		font-size: 12.5px;
	}
	textarea {
		width: 100%;
		min-height: 96px;
		padding: 12px;
		border-radius: 14px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font:
			500 14px/1.5 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
		resize: vertical;
		box-sizing: border-box;
	}
	.numeros {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 14px;
		border-radius: 16px;
		background: var(--sf3);
		flex-wrap: wrap;
	}
	.numeros > span {
		flex: 1 1 160px;
	}
	.nu-t,
	.nu-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.nu-d {
		font-size: 12px;
	}
	.nu-v {
		display: block;
		font-size: 24px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.nu-v.verde {
		color: #237550;
	}
	.acc {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.programar {
		padding: 12px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.enviar {
		flex: 1;
		padding: 12px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
	}
	.acc button:disabled {
		opacity: 0.55;
	}
	.tel-col {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
	}
	.tel {
		width: 260px;
		border-radius: 38px;
		padding: 9px;
		background: #2a140c;
		box-shadow: 0 30px 60px -30px rgba(42, 20, 12, 0.7);
	}
	.pantalla {
		border-radius: 30px;
		overflow: hidden;
		background: #efe7dc;
	}
	.t-cab {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 26px 12px 10px;
		background: #fff;
		color: #2a140c;
		font-size: 12.5px;
		font-weight: 600;
	}
	.t-cab img {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		object-fit: cover;
	}
	.t-cab span {
		flex: 1;
	}
	.chat {
		padding: 12px 10px 16px;
		min-height: 300px;
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
	}
	.burbuja {
		align-self: flex-start;
		max-width: 92%;
		border-radius: 12px 12px 12px 3px;
		background: #fff;
		color: #1f2a20;
		overflow: hidden;
		box-shadow: 0 1px 1px rgba(0, 0, 0, 0.06);
	}
	.burbuja img {
		width: 100%;
		height: 110px;
		object-fit: cover;
		display: block;
	}
	.b-txt {
		padding: 8px 10px;
		font-size: 12.5px;
		line-height: 1.45;
		white-space: pre-wrap;
	}
	.b-btn {
		border-top: 1px solid #eee;
		padding: 8px;
		text-align: center;
		font-size: 12.5px;
		font-weight: 600;
		color: #0b8db8;
	}
	.t-pie {
		font-size: 12.5px;
		color: var(--mu);
	}
	.campanas {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
		gap: 12px;
	}
	.campana {
		padding: 16px;
	}
	.ca-cab {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.ca-t {
		flex: 1;
		font-size: 15px;
		font-weight: 600;
	}
	.estado {
		font-size: 11.5px;
		font-weight: 700;
		padding: 3px 9px;
		border-radius: 999px;
		background: var(--chip);
		color: var(--mu);
	}
	.estado.activa {
		background: rgba(47, 143, 99, 0.13);
		color: #237550;
	}
	.ca-d {
		font-size: 12.5px;
		color: var(--mu);
		margin-top: 3px;
	}
	.metricas {
		margin-top: 12px;
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 6px;
	}
	.m-k {
		font-size: 11.5px;
		color: var(--mu);
	}
	.m-v {
		font-size: 16px;
		font-weight: 600;
	}
	.cupones {
		padding: 18px;
	}
	.cu-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		font-size: 16px;
		font-weight: 600;
	}
	.cu-cab a {
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
		color: #b0441d;
		text-decoration: none;
	}
	.cu-lista {
		margin-top: 10px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 10px;
	}
	.cupon {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px;
		border-radius: 16px;
		border: 1.5px dashed var(--ln2);
	}
	.codigo {
		font-size: 15px;
		font-weight: 700;
		letter-spacing: 0.04em;
		padding: 6px 10px;
		border-radius: 10px;
		background: rgba(210, 88, 47, 0.1);
		color: #b0441d;
	}
	.cu-txt {
		flex: 1;
		min-width: 0;
	}
	.cu-t {
		display: block;
		font-size: 13.5px;
		font-weight: 600;
	}
	.cu-u {
		display: block;
		font-size: 12px;
		color: var(--mu);
	}
	.vacio {
		color: var(--mu);
		font-size: 14px;
	}
	@media (max-width: 999px) {
		.arriba {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
