<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Seo from '#lib/components/Seo.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { hace } from '#lib/panel/listas/clientes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Reseñas y preguntas (design/PanelVentas.dc.html): responder las preguntas de Mercado Libre y de
	 * la tienda online, y moderar las reseñas (con compra o sin compra; ocultar, nunca borrar).
	 */
	let { data } = $props();
	const ui = getUi();
	let tab = $state<'q' | 'r'>('q');
	const pendientes = $derived(data.preguntas.filter((p) => !p.respuesta).length);
	let editando = $state<string | null>(null);
	let texto = $state('');
	const deTienda = (id: string) => id.startsWith('t-');

	async function publicar(id: string, respuesta: string, producto: string) {
		const r = await fetch(`/panel/resenas-y-preguntas/${id}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ respuesta })
		});
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se publicó', desc: 'Revisá la respuesta.' });
		editando = null;
		await invalidateAll();
		ui.toast(
			deTienda(id)
				? { k: 'ok', title: 'Respuesta publicada', desc: `${producto} · Tu tienda online` }
				: { k: 'ml', title: 'Respuesta publicada', desc: `${producto} · Mercado Libre` }
		);
	}

	// Reseñas: ocultar o mostrar (con Deshacer) y responder en público.
	let respondiendo = $state<string | null>(null);
	let respuesta = $state('');
	async function accion(id: string, a: 'ocultar' | 'mostrar' | 'responder', r = '') {
		const res = await fetch(`/panel/resenas-y-preguntas/resena/${id}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ accion: a, respuesta: r })
		});
		if (!res.ok) {
			ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo en un rato.' });
			return false;
		}
		await invalidateAll();
		return true;
	}
	async function ocultar(id: string, oculta: boolean) {
		if (!(await accion(id, oculta ? 'mostrar' : 'ocultar'))) return;
		ui.undoable(
			oculta ? 'Reseña visible otra vez' : 'Reseña oculta',
			oculta ? 'Se ve de nuevo en tu tienda.' : 'No se ve en tu tienda. No se borró.',
			() => {
				accion(id, oculta ? 'ocultar' : 'mostrar');
			}
		);
	}
	async function responderResena(id: string) {
		if (!(await accion(id, 'responder', respuesta))) return;
		respondiendo = null;
		ui.toast({
			k: 'ok',
			title: respuesta.trim() ? 'Respuesta publicada' : 'Respuesta quitada',
			desc: respuesta.trim() ? 'Se ve debajo de la reseña, en tu tienda.' : undefined
		});
	}
	const dias = (iso: string) => (Date.now() - Date.parse(iso)) / 86_400_000;
	const cuando = (iso: string) => {
		const h = (Date.now() - Date.parse(iso)) / 3_600_000;
		return h < 24 ? `hace ${Math.max(1, Math.round(h))} h` : hace(dias(iso));
	};
	const maxEstrellas = $derived(Math.max(1, ...data.porEstrellas));
</script>

<Seo
	title="Reseñas y preguntas · Panel"
	description="Reseñas de tu tienda"
	path="/panel/resenas-y-preguntas"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Reseñas y preguntas</h1>
			<div class="c-s">De tu tienda, Google y Mercado Libre</div>
		</div>
		<div class="tabs" role="group" aria-label="Ver">
			<button type="button" aria-pressed={tab === 'q'} onclick={() => (tab = 'q')}
				>Preguntas{#if pendientes}{` · ${pendientes}`}{/if}</button
			>
			<button type="button" aria-pressed={tab === 'r'} onclick={() => (tab = 'r')}>Reseñas</button>
		</div>
	</div>

	{#if tab === 'q'}
		<div class="preguntas" use:anim={{ t: 'rise', delay: 60 }}>
			{#each data.preguntas as q (q.id)}
				<div class="tarjeta">
					<div class="q-cab">
						{#if q.foto}<img src={unsplash(q.foto, 96, '&h=96')} alt="" />{/if}
						<span class="q-txt"
							><span class="q-s"
								>{#if deTienda(q.id)}<span class="tn">Tienda</span>{:else}<span class="ml">ML</span
									>{/if}{q.producto} · {cuando(q.creada)}</span
							><span class="q-q">{q.pregunta}</span></span
						>
					</div>
					{#if q.respuesta}
						<div class="publicada">
							<Icon icon={I.phCheckCircleFill} size={15} />Publicada en {deTienda(q.id)
								? 'tu tienda'
								: 'Mercado Libre'}
						</div>
						<div class="r-txt">{q.respuesta}</div>
					{:else if !q.sugerida}
						<!-- Las de la tienda online no traen sugerencia: se responden escribiendo. -->
						<div class="sugerida propia">
							<label class="s-t" for="r-{q.id}">Tu respuesta (se ve en la ficha)</label>
							<textarea
								id="r-{q.id}"
								value={editando === q.id ? texto : ''}
								oninput={(e) => {
									editando = q.id;
									texto = e.currentTarget.value;
								}}
								maxlength="1000"
								placeholder="¡Hola! Sí, …"></textarea>
							<div class="s-acc">
								<button
									type="button"
									class="pub"
									disabled={editando !== q.id || texto.trim().length < 2}
									onclick={() => publicar(q.id, texto, q.producto)}>Publicar</button
								>
							</div>
						</div>
					{:else}
						<div class="sugerida">
							<div class="s-t">✦ Respuesta de Hornero</div>
							{#if editando === q.id}
								<textarea bind:value={texto} aria-label="Respuesta" maxlength="2000"></textarea>
							{:else}
								<div class="s-txt">{q.sugerida}</div>
							{/if}
							<div class="s-acc">
								<button
									type="button"
									class="pub"
									onclick={() => publicar(q.id, editando === q.id ? texto : q.sugerida, q.producto)}
									>Publicar</button
								>
								{#if editando !== q.id}
									<button
										type="button"
										class="ed"
										onclick={() => {
											editando = q.id;
											texto = q.sugerida;
										}}>Editar</button
									>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{:else}
				<div class="vacio">No hay preguntas sin responder.</div>
			{/each}
		</div>
	{:else}
		<div class="resenas" use:anim={{ t: 'rise', delay: 60 }}>
			<div class="tarjeta resumen">
				<div class="r-prom">
					<span class="n"
						>{(data.promedio ?? 0).toLocaleString('es-AR', { minimumFractionDigits: 1 })}</span
					><span class="estrellas">★★★★★</span>
				</div>
				<div class="r-s">{data.total} reseñas · {data.recomendaria}% recomendaría</div>
				<div class="barras">
					{#each data.porEstrellas as c, i (i)}
						<div class="barra">
							<span class="b-n">{5 - i}</span>
							<div class="b-f"><div style:width="{(c / maxEstrellas) * 100}%"></div></div>
							<span class="b-c">{c}</span>
						</div>
					{/each}
				</div>
			</div>
			<div class="lista">
				{#each data.resenas.slice(0, 30) as r (r.id)}
					{@const sospechosa = r.estrellas === 1 && !r.conCompra}
					<div class="tarjeta resena" class:oculta={r.oculta} class:sospechosa>
						<div class="re-cab">
							<span class="estrellas" aria-label="{r.estrellas} estrellas"
								>{'★'.repeat(r.estrellas)}{'☆'.repeat(5 - r.estrellas)}</span
							><span class="re-q">{r.autor} · {r.origen}</span><span class="re-h"
								>{hace(dias(r.creada))}</span
							>
						</div>
						<div class="re-chips">
							{#if r.conCompra}<span class="chip-estado ok"
									><Icon icon={I.phSealCheckFill} size={13} />Compra verificada</span
								>{:else}<span class="chip-estado sin">Sin compra</span>{/if}
							{#if r.oculta}<span class="chip-estado oc"
									><Icon icon={I.phEyeSlash} size={13} />Oculta en la tienda</span
								>{/if}
							{#if r.producto}<span class="re-p">{r.producto}</span>{/if}
						</div>
						{#if r.texto}<div class="re-t">{r.texto}</div>{/if}
						{#if sospechosa && !r.oculta}
							<div class="aviso">
								Una estrella y sin compra: no hay un pedido detrás. No se borra sola; si no es de un
								cliente, podés ocultarla.
							</div>
						{/if}
						{#if respondiendo === r.id}
							<div class="sugerida propia">
								<label class="s-t" for="rr-{r.id}">Tu respuesta pública</label>
								<textarea id="rr-{r.id}" bind:value={respuesta} maxlength="1000"></textarea>
								<div class="s-acc">
									<button type="button" class="pub" onclick={() => responderResena(r.id)}
										>Publicar</button
									><button type="button" class="ed" onclick={() => (respondiendo = null)}
										>Cancelar</button
									>
								</div>
							</div>
						{:else if r.respuesta}
							<div class="re-r"><b>Tu respuesta</b>{r.respuesta}</div>
						{/if}
						{#if respondiendo !== r.id}
							<div class="re-acc">
								<button
									type="button"
									class="ed"
									onclick={() => {
										respondiendo = r.id;
										respuesta = r.respuesta ?? '';
									}}
									><Icon icon={I.phChatCircleDotsFill} size={14} />{r.respuesta
										? 'Editar respuesta'
										: 'Responder públicamente'}</button
								><button type="button" class="ed" onclick={() => ocultar(r.id, r.oculta)}
									><Icon icon={r.oculta ? I.phEye : I.phEyeSlash} size={14} />{r.oculta
										? 'Mostrar'
										: 'Ocultar'}</button
								>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	{/if}
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
	.tabs {
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
	}
	.tabs button {
		border: 0;
		font-size: 13px;
		font-weight: 600;
		padding: 8px 14px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
		white-space: nowrap;
	}
	.tabs button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.preguntas,
	.lista {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.tarjeta {
		padding: 16px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.q-cab {
		display: flex;
		gap: 12px;
		align-items: flex-start;
	}
	.q-cab img {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		object-fit: cover;
		flex: none;
	}
	.q-txt {
		flex: 1;
		min-width: 0;
	}
	.q-s {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12.5px;
		color: var(--mu);
	}
	.ml {
		font-size: 8px;
		font-weight: 800;
		padding: 2px 4px;
		border-radius: 4px;
		background: #ffe600;
		color: #2d3277;
	}
	.q-q {
		display: block;
		font-size: 15px;
		font-weight: 600;
		margin-top: 3px;
	}
	.sugerida {
		margin-top: 12px;
		padding: 12px;
		border-radius: 14px;
		background: rgba(210, 88, 47, 0.08);
		border: 1px solid rgba(210, 88, 47, 0.2);
	}
	.s-t {
		font-size: 12px;
		font-weight: 600;
		color: #c24e25;
		margin-bottom: 4px;
	}
	.s-txt {
		font-size: 14px;
		line-height: 1.45;
	}
	.sugerida textarea {
		width: 100%;
		min-height: 80px;
		padding: 10px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 14px;
		line-height: 1.45;
		resize: vertical;
	}
	.s-acc {
		margin-top: 10px;
		display: flex;
		gap: 6px;
	}
	.s-acc button {
		padding: 8px 14px;
		border-radius: 999px;
		font-size: 13px;
		font-weight: 600;
	}
	.pub {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.ed {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.publicada {
		margin-top: 10px;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		color: #2f8f63;
	}
	.r-txt {
		margin-top: 6px;
		font-size: 13.5px;
		line-height: 1.45;
		color: var(--mu);
	}
	.resenas {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 16px;
		align-items: start;
	}
	.resumen {
		padding: 20px;
	}
	.r-prom {
		display: flex;
		align-items: baseline;
		gap: 10px;
	}
	.r-prom .n {
		font-size: 52px;
		font-weight: 600;
		letter-spacing: -0.05em;
	}
	.estrellas {
		color: #e3a93b;
		letter-spacing: 2px;
	}
	.r-prom .estrellas {
		font-size: 18px;
	}
	.r-s {
		font-size: 13.5px;
		color: var(--mu);
	}
	.barras {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 7px;
	}
	.barra {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 13px;
	}
	.b-n {
		width: 14px;
	}
	.b-f {
		flex: 1;
		height: 8px;
		border-radius: 6px;
		background: var(--sf3);
		overflow: hidden;
	}
	.b-f div {
		height: 100%;
		border-radius: 6px;
		background: #e3a93b;
	}
	.b-c {
		width: 36px;
		text-align: right;
		color: var(--mu);
	}
	.resena {
		padding: 14px 16px;
		border-radius: 20px;
	}
	.re-cab {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.re-cab .estrellas {
		letter-spacing: 1px;
	}
	.re-q {
		flex: 1;
		font-size: 13px;
		color: var(--mu);
	}
	.re-h {
		font-size: 12px;
		color: var(--fa);
	}
	.re-t {
		margin-top: 6px;
		font-size: 14px;
		line-height: 1.5;
	}
	.tn {
		font-size: 8px;
		font-weight: 800;
		padding: 2px 4px;
		border-radius: 4px;
		background: var(--tx);
		color: var(--sf);
		text-transform: uppercase;
	}
	.propia {
		background: var(--sf2, rgba(127, 127, 127, 0.06));
		border-color: var(--ln2);
	}
	.propia .s-t {
		display: block;
		color: var(--tx);
	}
	.pub:disabled {
		opacity: 0.5;
	}
	.re-chips {
		margin-top: 8px;
		display: flex;
		align-items: center;
		gap: 6px;
		flex-wrap: wrap;
	}
	.re-chips .chip-estado::before {
		display: none;
	}
	.chip-estado.ok {
		--chip-bg: rgba(47, 143, 99, 0.12);
		--chip-fg: #2f8f63;
	}
	.chip-estado.sin {
		--chip-bg: var(--chip);
		--chip-fg: var(--mu);
	}
	.chip-estado.oc {
		--chip-bg: var(--chip);
		--chip-fg: var(--tx);
	}
	.re-p {
		font-size: 12.5px;
		color: var(--mu);
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.resena.oculta > :not(.re-acc) {
		opacity: 0.55;
	}
	.resena.sospechosa {
		box-shadow:
			inset 0 0 0 1.5px rgba(210, 88, 47, 0.45),
			var(--shd);
	}
	.aviso {
		margin-top: 10px;
		padding: 10px 12px;
		border-radius: 12px;
		background: rgba(210, 88, 47, 0.08);
		color: #c24e25;
		font-size: 13px;
		line-height: 1.45;
		text-wrap: pretty;
	}
	.re-r {
		margin-top: 10px;
		padding: 10px 12px;
		border-left: 2px solid var(--ln2);
		font-size: 13.5px;
		line-height: 1.45;
		color: var(--mu);
	}
	.re-r b {
		display: block;
		color: var(--tx);
		font-size: 12.5px;
	}
	.re-acc {
		margin-top: 12px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.re-acc button {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		min-height: 36px;
		padding: 8px 14px;
		border-radius: 999px;
		font-size: 13px;
		font-weight: 600;
	}
	@media (max-width: 760px) {
		.re-acc button,
		.s-acc button {
			min-height: 44px;
		}
	}
	.vacio {
		padding: 28px;
		text-align: center;
		color: var(--mu);
	}
</style>
