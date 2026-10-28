<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Dominio (design/PanelCuenta.dc.html, sDom): dónde se ve la tienda y conectar uno propio. */
	let { data } = $props();
	const ui = getUi();
	let dominio = $state('');
	const verificando = $derived(data.dominios.find((d) => d.estado === 'verificando'));
	const principal = $derived(data.dominios.find((d) => d.principal)?.nombre ?? data.gratis);

	// Mientras haya uno verificando, se vuelve a mirar cada 20 s (sólo con la pantalla abierta).
	$effect(() => {
		if (!verificando) return;
		const t = setInterval(() => void invalidateAll(), 20_000);
		return () => clearInterval(t);
	});

	async function pedir(url: string, method: string, cuerpo?: unknown) {
		const r = await fetch(url, {
			method,
			headers: cuerpo === undefined ? {} : { 'content-type': 'application/json' },
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			ui.toast({ k: 'err', title: 'No se pudo', desc: m ?? 'Probá de nuevo.' });
		}
		return r.ok;
	}
	async function conectar(e: SubmitEvent) {
		e.preventDefault();
		const d = dominio.trim();
		if (!(await pedir('/panel/dominio', 'POST', { dominio: d }))) return;
		dominio = '';
		await invalidateAll();
		ui.toast({
			k: 'info',
			title: `Buscando los registros de ${d}`,
			desc: 'Puede tardar hasta 1 hora en propagarse'
		});
	}
	async function hacerPrincipal(n: string) {
		if (!(await pedir(`/panel/dominio/${encodeURIComponent(n)}`, 'PUT'))) return;
		await invalidateAll();
		ui.toast({ k: 'ok', title: `${n} es tu dominio principal`, desc: 'Los demás redirigen ahí' });
	}
	async function quitar(n: string) {
		if (!(await pedir(`/panel/dominio/${encodeURIComponent(n)}`, 'DELETE'))) return;
		await invalidateAll();
		ui.toast({ k: 'ok', title: `${n} quitado`, desc: 'La tienda sigue en tu dirección gratis' });
	}
	async function copiar(v: string) {
		try {
			await navigator.clipboard.writeText(v);
		} catch {
			// Sin permiso: igual queda a la vista.
		}
		ui.toast({ k: 'ok', title: 'Copiado', desc: v });
	}
	const CHEQUEO = {
		ok: [I.phCheckCircleFill, '#2F8F63', 'listo'],
		esperando: [I.phHourglassMediumFill, '#E3A93B', 'revisamos cada minuto'],
		pendiente: [I.phCircleDashed, 'var(--mu)', 'después del anterior']
	} as const;
	const sugerencias = $derived.by(() => {
		const base = data.gratis.split('/').pop() ?? 'tutienda';
		return [`${base}.com.ar`, `${base}.ar`, `${base}shop.com.ar`];
	});
</script>

<Seo title="Dominio · Panel" description="El dominio de tu tienda" path="/panel/dominio" noindex />

<div class="pagina">
	<div use:anim={{ t: 'rise' }}>
		<h1>Dominio</h1>
		<div class="c-s">
			Tu tienda se ve en <b>{principal}</b>{data.dominios.length > 1
				? ' y en tus dominios propios'
				: ''}
		</div>
	</div>

	<div class="lista" use:anim={{ t: 'rise', delay: 50 }}>
		{#each data.dominios as d (d.nombre)}
			{@const gratis = d.nombre === data.gratis}
			<div class="dominio">
				<span class="ico" class:principal={d.principal}
					><Icon icon={gratis ? I.phStorefrontFill : I.phGlobeHemisphereWestFill} size={20} /></span
				>
				<span class="txt">
					<span class="n"
						>{d.nombre}{#if d.principal}<span class="badge">Principal</span>{/if}</span
					>
					<span class="d">{d.detalle}</span>
				</span>
				<span class="estado" class:espera={d.estado !== 'activo'}
					><Icon
						icon={d.estado === 'activo' ? I.phLockSimpleFill : I.phHourglassMediumFill}
						size={13}
					/>{d.estado === 'activo' ? 'SSL activo' : 'Verificando'}</span
				>
				{#if !d.principal && d.estado === 'activo'}
					<button type="button" class="sec" onclick={() => hacerPrincipal(d.nombre)}
						>Hacer principal</button
					>
				{/if}
				{#if !gratis}
					<button
						type="button"
						class="quitar"
						aria-label="Quitar {d.nombre}"
						onclick={() => quitar(d.nombre)}><Icon icon={I.phTrash} size={16} /></button
					>
				{/if}
			</div>
		{/each}
	</div>

	<div class="dos" use:anim={{ t: 'rise', delay: 100 }}>
		<div class="tarjeta">
			<div class="t-t">Conectar un dominio</div>
			<form class="campo" onsubmit={conectar}>
				<input
					bind:value={dominio}
					placeholder="tutienda.com.ar"
					aria-label="Tu dominio"
					maxlength="253"
				/>
				<button disabled={dominio.trim().length < 4}>Conectar</button>
			</form>
			<div class="ayuda">
				Agregá estos registros en tu proveedor (NIC Argentina, DonWeb, GoDaddy…):
			</div>
			<div class="registros">
				{#each data.registros as r (r.tipo + r.nombre)}
					<div class="registro">
						<b>{r.tipo}</b>
						<span
							><span class="r-n">Nombre: {r.nombre}</span><span class="r-v">{r.valor}</span></span
						>
						<button type="button" aria-label="Copiar {r.valor}" onclick={() => copiar(r.valor)}
							><Icon icon={I.phCopy} size={14} /></button
						>
					</div>
				{/each}
			</div>
			{#if verificando?.chequeos}
				<div class="chequeos">
					{#each verificando.chequeos as c (c.titulo)}
						{@const [icono, col, t] = CHEQUEO[c.estado]}
						<div>
							<Icon icon={icono} size={18} style="color:{col}" /><span>{c.titulo}</span><span
								class="cs">{t}</span
							>
						</div>
					{/each}
				</div>
			{/if}
		</div>
		<div class="col">
			<div class="ssl">
				<img src="/hornero/seguridad.png" alt="" width="54" height="54" />
				<div>
					<b>SSL gratis y automático</b>
					<div>Renovamos el certificado solos cada 90 días</div>
				</div>
			</div>
			<div class="tarjeta">
				<div class="t-t chico">¿Todavía no tenés dominio?</div>
				<div class="t-d">
					Registralo en NIC Argentina (.com.ar y .ar) y después conectalo acá en 5 minutos.
				</div>
				<div class="comprar">
					{#each sugerencias as n (n)}
						<div>
							<span class="cn">{n}</span>
							<a
								href="https://nic.ar/es/dominios/busqueda?dominio={encodeURIComponent(n)}"
								target="_blank"
								rel="noopener">Buscar en NIC</a
							>
						</div>
					{/each}
				</div>
			</div>
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
		overflow-wrap: anywhere;
	}
	.c-s b {
		color: var(--tx);
	}
	.lista {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.dominio {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 16px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		flex-wrap: wrap;
	}
	.ico {
		width: 40px;
		height: 40px;
		border-radius: 12px;
		background: var(--chip);
		color: var(--mu);
		display: grid;
		place-items: center;
		flex: none;
	}
	.ico.principal {
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}
	.txt {
		flex: 1 1 200px;
		min-width: 0;
	}
	.n {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
		font-size: 15px;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.badge {
		font-size: 11px;
		font-weight: 700;
		padding: 2px 8px;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
	}
	.d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.estado {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12.5px;
		font-weight: 700;
		padding: 4px 10px;
		border-radius: 999px;
		background: rgba(47, 143, 99, 0.12);
		color: #237550;
	}
	.estado.espera {
		background: rgba(227, 169, 59, 0.18);
		color: #8a5f0e;
	}
	.sec {
		padding: 8px 12px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
	}
	.quitar {
		width: 32px;
		height: 32px;
		border-radius: 10px;
		border: 0;
		background: transparent;
		color: var(--mu);
		display: grid;
		place-items: center;
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr));
		gap: 12px;
		align-items: start;
	}
	.col {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.tarjeta {
		padding: 20px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.t-t {
		font-size: 16px;
		font-weight: 600;
	}
	.t-t.chico {
		font-size: 15.5px;
	}
	.t-d {
		font-size: 13.5px;
		color: var(--mu);
		margin-top: 4px;
	}
	.campo {
		margin-top: 12px;
		display: flex;
		padding: 5px;
		border-radius: 999px;
		background: var(--sf3);
	}
	.campo input {
		flex: 1;
		min-width: 0;
		padding: 9px 14px;
		border: 0;
		background: transparent;
		font:
			500 15px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
	}
	.campo button {
		padding: 10px 16px;
		border: 0;
		border-radius: 999px;
		background: var(--tx);
		color: var(--sf);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.campo button:disabled {
		opacity: 0.5;
	}
	.ayuda {
		margin-top: 14px;
		font-size: 13px;
		color: var(--mu);
	}
	.registros {
		margin-top: 8px;
		border-radius: 16px;
		border: 1px solid var(--ln);
		overflow: hidden;
	}
	.registro {
		display: grid;
		grid-template-columns: 70px minmax(0, 1fr) auto;
		gap: 10px;
		align-items: center;
		padding: 10px 12px;
		border-bottom: 1px solid var(--ln);
		font-size: 13.5px;
	}
	.registro:last-child {
		border-bottom: 0;
	}
	.r-n {
		display: block;
		font-size: 12px;
		color: var(--mu);
	}
	.r-v {
		display: block;
		overflow-wrap: anywhere;
		font-weight: 600;
	}
	.registro button {
		width: 30px;
		height: 30px;
		border-radius: 9px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		display: grid;
		place-items: center;
	}
	.chequeos {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.chequeos div {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 13.5px;
	}
	.chequeos span:nth-child(2) {
		flex: 1;
	}
	.cs {
		font-size: 12px;
		color: var(--mu);
	}
	.ssl {
		padding: 20px;
		border-radius: 24px;
		background: #2a140c;
		color: #fbf7f2;
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.ssl img {
		object-fit: contain;
	}
	.ssl b {
		font-size: 15.5px;
	}
	.ssl div div {
		font-size: 13px;
		color: #d9c8ba;
	}
	.comprar {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.comprar > div {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 14px;
		background: var(--sf2);
		font-size: 14px;
	}
	.cn {
		flex: 1;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.comprar a {
		padding: 7px 12px;
		border-radius: 999px;
		background: var(--tx);
		color: var(--sf);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
		text-decoration: none;
	}
</style>
