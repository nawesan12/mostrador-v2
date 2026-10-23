<script lang="ts">
	import { invalidateAll, replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { marca } from '#lib/marcas.ts';
	import { anim } from '#lib/panel/anim.ts';
	import {
		CATEGORIAS,
		IAS,
		INTEGRACIONES,
		PERMISOS_IA,
		type AjusteIntegracion,
		type Integracion
	} from '#lib/panel/integraciones.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Integraciones (design/PanelOperacion.dc.html, sInt). "Configurar" no está diseñado: se arma con
	 * el modal y los interruptores del handoff. Mercado Pago es real (OAuth de la tienda); las demás
	 * llegan una por PR y en la demo se simulan.
	 */
	let { data } = $props();
	const ui = getUi();

	let cat = $state<(typeof CATEGORIAS)[number][0]>('todas');
	const lista = $derived(INTEGRACIONES.filter((i) => cat === 'todas' || i.categoria === cat));
	const conectadas = $derived(Object.values(data.estado.conectadas).filter(Boolean).length);
	let autorizando = $state<string | null>(null);
	let abierta = $state<Integracion | null>(null);
	let modal = $state(false);

	// La vuelta del OAuth de Mercado Pago deja ?mercadopago=conectada|error.
	let avisado = false;
	$effect(() => {
		const mp = page.url.searchParams.get('mercadopago');
		const ml = page.url.searchParams.get('mercadolibre');
		if ((!mp && !ml) || avisado) return;
		avisado = true;
		const nombre = mp ? 'Mercado Pago' : 'Mercado Libre';
		ui.toast(
			(mp ?? ml) === 'conectada'
				? {
						k: 'ok',
						title: `${nombre} conectado`,
						desc: mp ? 'Ya podés cobrar con tu cuenta' : 'Publicá productos desde su ficha'
					}
				: { k: 'err', title: `No se conectó ${nombre}`, desc: 'Probá de nuevo en un rato' }
		);
		replaceState(resolve('/panel/integraciones'), {});
	});

	async function ajustar(a: AjusteIntegracion) {
		const r = await fetch('/panel/integraciones/ajustar', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(a)
		});
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			ui.toast({ k: 'warn', title: 'Todavía no', desc: m ?? 'Probá de nuevo.' });
			return false;
		}
		await invalidateAll();
		return true;
	}

	const RUTA = {
		mp: '/panel/integraciones/mercadopago',
		ml: '/panel/integraciones/mercadolibre'
	} as const;
	async function conectarMP(id: 'mp' | 'ml' = 'mp') {
		const r = await fetch(RUTA[id], { method: 'POST' });
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({
				k: 'warn',
				title: id === 'mp' ? 'Mercado Pago' : 'Mercado Libre',
				desc: m ?? 'Probá de nuevo.'
			});
		}
		location.href = ((await r.json()) as { url: string }).url;
	}

	async function desconectarMP(id: 'mp' | 'ml' = 'mp') {
		const r = await fetch(RUTA[id], { method: 'DELETE' });
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo.' });
		modal = false;
		await invalidateAll();
		ui.toast(
			id === 'mp'
				? {
						k: 'info',
						title: 'Mercado Pago desconectado',
						desc: 'La tienda deja de cobrar con tarjeta hasta que lo conectes'
					}
				: {
						k: 'info',
						title: 'Mercado Libre desconectado',
						desc: 'Las publicaciones quedan, pero sin sincronizar stock ni precio'
					}
		);
	}

	async function ajustarML(recargo: number, pausarSinStock: boolean) {
		const r = await fetch(RUTA.ml, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ recargo, pausarSinStock })
		});
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo.' });
		await invalidateAll();
	}

	async function boton(i: Integracion) {
		const on = data.estado.conectadas[i.id];
		if (on) {
			abierta = i;
			modal = true;
			return;
		}
		if (i.id === 'mp' || i.id === 'ml') return conectarMP(i.id);
		if (i.pronto)
			return ui.toast({ k: 'info', title: i.nombre, desc: 'Llega pronto. Te avisamos.' });
		if (await ajustar({ tipo: 'conexion', id: i.id, valor: true }))
			ui.toast({
				k: 'ok',
				title: `${i.nombre} conectada`,
				desc: 'Sincronizando datos, tarda unos segundos'
			});
	}

	async function desconectar(i: Integracion) {
		if (i.id === 'mp' || i.id === 'ml') return desconectarMP(i.id);
		if (!(await ajustar({ tipo: 'conexion', id: i.id, valor: false }))) return;
		modal = false;
		ui.toast({ k: 'info', title: `${i.nombre} desconectada`, desc: 'Podés volver a conectarla' });
	}

	async function ia(id: string, nombre: string) {
		const on = data.estado.ia[id];
		if (on) {
			if (await ajustar({ tipo: 'ia', id, valor: false }))
				ui.toast({
					k: 'info',
					title: `${nombre} desconectado`,
					desc: 'Ya no puede ver ni tocar tu negocio'
				});
			return;
		}
		autorizando = id;
		await new Promise((r) => setTimeout(r, 1300));
		autorizando = null;
		if (await ajustar({ tipo: 'ia', id, valor: true }))
			ui.toast({ k: 'ok', title: `${nombre} conectado`, desc: 'Probá: "¿qué se vendió hoy?"' });
	}

	const fecha = (iso?: string) =>
		iso
			? new Date(iso).toLocaleDateString('es-AR', {
					day: 'numeric',
					month: 'long',
					timeZone: 'America/Argentina/Buenos_Aires'
				})
			: '';
</script>

<Seo
	title="Integraciones · Panel"
	description="Integraciones de tu tienda"
	path="/panel/integraciones"
	noindex
/>

<div class="pagina" data-mp-marketplace-page="checkout-pro">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Integraciones</h1>
			<div class="sub">{conectadas} conectadas · todas incluidas en tu&nbsp;plan</div>
		</div>
	</div>

	<section class="ia" use:anim={{ t: 'rise', delay: 50 }}>
		<div class="ia-cab">
			<img src="/hornero/whatsapp.png" alt="" width="52" height="52" />
			<div class="ia-txt">
				<div class="ia-t">Manejá Mostrador desde tu IA<span class="mcp">MCP</span></div>
				<div class="ia-d">
					Entrás con tu cuenta de la IA y listo. Preguntale por ventas, stock o pedidos, y pedile
					que cargue, edite o responda por&nbsp;vos.
				</div>
			</div>
		</div>
		<div class="provs">
			{#each IAS as p (p.id)}
				{@const on = data.estado.ia[p.id]}
				{@const espera = autorizando === p.id}
				<button
					type="button"
					class="prov"
					class:on
					onclick={() => ia(p.id, p.nombre)}
					disabled={espera}
				>
					<span class="p-ico" style:color={p.col}><Icon icon={p.icon} size={19} /></span>
					<span class="p-txt">
						<span class="p-t">{p.nombre}</span>
						<span class="p-st"
							>{on ? 'Conectado' : espera ? 'Autorizando…' : 'Conectar con tu cuenta'}</span
						>
					</span>
					<Icon icon={on ? I.phCheckCircleFill : I.phArrowRightBold} size={16} />
				</button>
			{/each}
		</div>
		<div class="perms">
			{#each PERMISOS_IA as g (g.id)}
				{@const on = data.estado.permisosIa[g.id]}
				<button
					type="button"
					class="perm"
					role="switch"
					aria-checked={on}
					disabled={g.id === 'leer'}
					onclick={() => ajustar({ tipo: 'permisoIa', id: g.id, valor: !on })}
				>
					<Icon icon={g.icon} size={17} />
					<span class="g-txt"><span class="g-t">{g.t}</span><span class="g-d">{g.d}</span></span>
					<span class="tg" class:on><span></span></span>
				</button>
			{/each}
		</div>
	</section>

	<div class="cats" role="group" aria-label="Categoría" use:anim={{ t: 'rise', delay: 100 }}>
		{#each CATEGORIAS as [id, t] (id)}
			<button
				type="button"
				class:on={cat === id}
				aria-pressed={cat === id}
				onclick={() => (cat = id)}>{t}</button
			>
		{/each}
	</div>

	<div class="grilla" use:anim={{ t: 'rise', delay: 150 }}>
		{#each lista as i (i.id)}
			{@const on = data.estado.conectadas[i.id]}
			{@const m = marca(i.marca)}
			<div class="tarjeta" class:on>
				<div class="t-cab">
					<span class="logo" style:background={m.bg}>
						{#if m.icon}<Icon
								icon={m.icon}
								size={22}
								style="color:{m.col ?? 'inherit'}"
							/>{:else}<span class="sigla" style:color={m.col}>{m.sigla}</span>{/if}
					</span>
					<span class="t-txt"
						><span class="t-n">{i.nombre}</span><span class="t-d">{i.detalle}</span></span
					>
				</div>
				<div class="t-pie">
					<span class="t-st" class:on
						>{on ? '● Conectada' : i.pronto ? 'Próximamente' : 'Disponible'}</span
					>
					<button
						type="button"
						class="t-btn"
						class:on
						data-mp-marketplace-connect={i.id === 'mp' && !on ? 'oauth' : undefined}
						onclick={() => boton(i)}
					>
						{on ? 'Configurar' : 'Conectar'}
					</button>
				</div>
			</div>
		{/each}
	</div>
</div>

<Modal
	bind:open={modal}
	titulo={abierta ? `Configurar ${abierta.nombre}` : 'Configurar'}
	ancho={480}
	padding={24}
>
	{#if abierta}
		{@const m = marca(abierta.marca)}
		{@const mp = abierta.id === 'mp'}
		<div class="m-cab">
			<span class="logo grande" style:background={m.bg}>
				{#if m.icon}<Icon icon={m.icon} size={26} style="color:{m.col ?? 'inherit'}" />{:else}<span
						class="sigla"
						style:color={m.col}>{m.sigla}</span
					>{/if}
			</span>
			<div>
				<div class="m-t">{abierta.nombre}</div>
				<div class="m-st">
					● Conectada{#if mp && data.mp.desde}&nbsp;desde el {fecha(data.mp.desde)}{/if}
				</div>
			</div>
		</div>
		{#if mp && data.mp.reautorizar}
			<div class="m-aviso" role="alert">
				<Icon icon={I.phWarningCircleFill} size={18} />
				<span
					>Mercado Pago pidió que vuelvas a conectar la cuenta. Hasta entonces no se
					puede&nbsp;cobrar.</span
				>
			</div>
		{/if}
		<div class="m-sec">Qué hace</div>
		<ul class="m-hace">
			{#each abierta.hace as h (h)}
				<li><Icon icon={I.phCheckCircleFill} size={16} />{h}</li>
			{/each}
		</ul>
		{#if mp && data.mp.usuario}
			<div class="m-cuenta">
				Cuenta de Mercado Pago <b>#{data.mp.usuario}</b>. La plata de cada venta va directo ahí:
				Mostrador no cobra&nbsp;comisión.
			</div>
		{/if}
		{#if abierta.id === 'ml'}
			<div class="m-cuenta">
				Cuenta <b>{data.ml.apodo}</b> · {data.ml.publicaciones}
				{data.ml.publicaciones === 1 ? 'publicación' : 'publicaciones'} desde Mostrador.
				{#if data.ml.preguntas}
					<a href={resolve('/panel/resenas-y-preguntas')}
						>{data.ml.preguntas}
						{data.ml.preguntas === 1 ? 'pregunta' : 'preguntas'} sin&nbsp;responder</a
					>{/if}
			</div>
			<div class="m-sec">Precio en Mercado Libre</div>
			<div class="recargos" role="group" aria-label="Recargo por comisión">
				{#each [0, 5, 10, 15, 20] as r (r)}
					<button
						type="button"
						class:on={data.ml.recargo === r}
						aria-pressed={data.ml.recargo === r}
						onclick={() => ajustarML(r, data.ml.pausarSinStock)}>{r ? `+${r}%` : 'Igual'}</button
					>
				{/each}
			</div>
			<div class="m-ops">
				<div class="m-op">
					<span class="g-txt"
						><span class="g-t">Pausar sin stock</span><span class="g-d"
							>Las publicaciones se pausan solas y vuelven cuando hay</span
						></span
					>
					<Interruptor
						label="Pausar sin stock"
						checked={data.ml.pausarSinStock}
						onchange={(v) => void ajustarML(data.ml.recargo, v)}
					/>
				</div>
			</div>
		{:else if abierta.opciones.length}
			<div class="m-sec">Opciones</div>
			<div class="m-ops">
				{#each abierta.opciones as o (o.id)}
					{@const id = abierta.id}
					<div class="m-op">
						<span class="g-txt"><span class="g-t">{o.t}</span><span class="g-d">{o.d}</span></span>
						<Interruptor
							label={o.t}
							checked={data.estado.opciones[id]?.[o.id] ?? o.def}
							onchange={(v) => void ajustar({ tipo: 'opcion', id, clave: o.id, valor: v })}
						/>
					</div>
				{/each}
			</div>
		{/if}
		<div class="m-acc">
			<button type="button" class="b-sec rojo" onclick={() => abierta && desconectar(abierta)}
				>Desconectar</button
			>
			{#if (mp && data.mp.reautorizar) || (abierta.id === 'ml' && data.ml.reautorizar)}
				<button type="button" class="b-pri" onclick={() => conectarMP(mp ? 'mp' : 'ml')}
					>Volver a conectar</button
				>
			{:else}
				<button type="button" class="b-pri" onclick={() => (modal = false)}>Listo</button>
			{/if}
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
	.sub {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.ia {
		padding: 20px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.ia-cab {
		display: flex;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
	}
	.ia-cab img {
		width: 52px;
		height: 52px;
		object-fit: contain;
		flex: none;
	}
	.ia-txt {
		flex: 1 1 280px;
		min-width: 0;
	}
	.ia-t {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 17px;
		font-weight: 600;
	}
	.mcp {
		font-size: 11px;
		font-weight: 700;
		padding: 3px 8px;
		border-radius: 999px;
		background: rgba(210, 88, 47, 0.12);
		color: #c24e25;
	}
	.ia-d {
		font-size: 13.5px;
		color: var(--mu);
		margin-top: 2px;
	}
	.provs {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 10px;
	}
	.prov {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 14px;
		border-radius: 16px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--mu);
		cursor: pointer;
		font-family: inherit;
		text-align: left;
	}
	.prov.on {
		border-color: rgba(47, 143, 99, 0.5);
		background: rgba(47, 143, 99, 0.06);
		color: #2f8f63;
	}
	.p-ico {
		width: 36px;
		height: 36px;
		flex: none;
		border-radius: 11px;
		display: grid;
		place-items: center;
		background: var(--sf3);
	}
	.p-txt {
		flex: 1;
		min-width: 0;
	}
	.p-t {
		display: block;
		font-size: 14px;
		font-weight: 600;
		color: var(--tx);
	}
	.p-st {
		display: block;
		font-size: 12px;
		font-weight: 600;
	}
	.perms {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 4px 18px;
		padding-top: 4px;
		border-top: 1px solid var(--ln);
	}
	.perm {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 0;
		border: 0;
		background: transparent;
		cursor: pointer;
		font-family: inherit;
		color: var(--mu);
		text-align: left;
	}
	.perm:disabled {
		cursor: default;
	}
	.g-txt {
		flex: 1;
		min-width: 0;
	}
	.g-t {
		display: block;
		font-size: 13.5px;
		font-weight: 600;
		color: var(--tx);
	}
	.g-d {
		display: block;
		font-size: 12px;
		color: var(--fa);
	}
	.tg {
		position: relative;
		width: 36px;
		height: 21px;
		flex: none;
		border-radius: 999px;
		background: #d9cbbe;
		transition: background 0.25s;
	}
	.tg.on {
		background: #d2582f;
	}
	.tg span {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 17px;
		height: 17px;
		border-radius: 50%;
		background: #fff;
		transition: left 0.25s;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
	}
	.tg.on span {
		left: 17px;
	}
	.cats {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.cats button {
		border: 0;
		cursor: pointer;
		white-space: nowrap;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		padding: 9px 14px;
		border-radius: 999px;
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--sh);
	}
	.cats button.on {
		background: var(--tx);
		color: var(--sf);
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
		gap: 12px;
	}
	.tarjeta {
		padding: 16px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 12px;
		border: 1.5px solid transparent;
		transition: border-color 0.3s;
	}
	.tarjeta.on {
		border-color: rgba(47, 143, 99, 0.35);
	}
	.t-cab {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.logo {
		width: 42px;
		height: 42px;
		flex: none;
		border-radius: 12px;
		display: grid;
		place-items: center;
	}
	.logo.grande {
		width: 52px;
		height: 52px;
		border-radius: 16px;
	}
	.sigla {
		font-size: 8.5px;
		font-weight: 800;
	}
	.t-txt {
		flex: 1;
		min-width: 0;
	}
	.t-n {
		display: block;
		font-size: 15px;
		font-weight: 600;
	}
	.t-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.t-pie {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.t-st {
		flex: 1;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--fa);
	}
	.t-st.on {
		color: #2f8f63;
	}
	.t-btn {
		padding: 8px 14px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
		white-space: nowrap;
	}
	.t-btn.on {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.m-cab {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.m-t {
		font-size: 19px;
		font-weight: 600;
	}
	.m-st {
		font-size: 13px;
		font-weight: 600;
		color: #2f8f63;
	}
	.m-aviso {
		margin-top: 14px;
		display: flex;
		gap: 8px;
		align-items: flex-start;
		padding: 12px 14px;
		border-radius: 14px;
		background: rgba(227, 169, 59, 0.16);
		color: #9a6a10;
		font-size: 13.5px;
		line-height: 1.45;
	}
	.m-sec {
		margin-top: 16px;
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--fa);
	}
	.m-hace {
		margin: 8px 0 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 7px;
	}
	.m-hace li {
		display: flex;
		gap: 8px;
		align-items: flex-start;
		font-size: 14px;
		line-height: 1.4;
	}
	.m-hace li :global(svg) {
		flex: none;
		margin-top: 2px;
		color: #2f8f63;
	}
	.m-cuenta {
		margin-top: 12px;
		padding: 12px 14px;
		border-radius: 14px;
		background: var(--sf2);
		font-size: 13px;
		line-height: 1.5;
		color: var(--mu);
	}
	.m-cuenta b {
		color: var(--tx);
	}
	.m-cuenta a {
		color: #c24e25;
		font-weight: 600;
	}
	.recargos {
		margin-top: 8px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.recargos button {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		padding: 8px 14px;
		border-radius: 999px;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.recargos button.on {
		background: var(--tx);
		color: var(--sf);
		border-color: var(--tx);
	}
	.m-ops {
		margin-top: 4px;
		display: flex;
		flex-direction: column;
	}
	.m-op {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.m-op:last-child {
		border-bottom: 0;
	}
	.m-acc {
		margin-top: 18px;
		display: flex;
		gap: 8px;
	}
	.m-acc button {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.b-sec {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.b-sec.rojo {
		color: #b83a2e;
	}
	.b-pri {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
</style>
