<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Mi tienda (design/PanelOperacion.dc.html, sStore): secciones, colores y vista previa en vivo. */
	let { data } = $props();
	const ui = getUi();
	type Seccion = (typeof data.diseno.secciones)[number];

	const SECCIONES = {
		hero: [I.phImageFill, 'Portada'],
		cats: [I.phTagFill, 'Categorías'],
		prods: [I.phStorefrontFill, 'Productos destacados'],
		revs: [I.phStarFill, 'Reseñas'],
		ig: [I.logosInstagramIcon, 'Instagram'],
		wa: [I.phWhatsappLogoFill, 'Botón de WhatsApp']
	} as const;
	const PALETAS = [
		{ bg: '#F3E9DC', ink: '#3B2416', acc: '#6B3A22' },
		{ bg: '#EEF1E8', ink: '#1E2B22', acc: '#2F5D50' },
		{ bg: '#FBF3EC', ink: '#2A140C', acc: '#C24E25' },
		{ bg: '#1E1410', ink: '#F4E9E1', acc: '#E8875F' }
	];
	// Lo que se edita es local y se guarda en cada cambio; al recargar vuelve lo guardado.
	let secciones = $derived<Seccion[]>(data.diseno.secciones.map((s) => ({ ...s })));
	let paleta = $derived(data.diseno.paleta);
	let sinPublicar = $derived(data.sinPublicar);
	const pal = $derived(PALETAS[paleta]);

	// Los guardados van en fila: si dos salen juntos, el último en llegar tiene que ser el último hecho.
	let fila = Promise.resolve();
	function guardar() {
		sinPublicar = true;
		const cuerpo = JSON.stringify({ secciones, paleta });
		fila = fila.then(async () => {
			await fetch('/panel/mi-tienda', {
				method: 'PUT',
				headers: { 'content-type': 'application/json' },
				body: cuerpo
			});
		});
		return fila;
	}
	function subir(i: number) {
		if (!i) return;
		const a = [...secciones];
		[a[i - 1], a[i]] = [a[i], a[i - 1]];
		secciones = a;
		void guardar();
	}
	function alternar(i: number) {
		secciones = secciones.map((s, k) => (k === i ? { ...s, visible: !s.visible } : s));
		void guardar();
	}
	async function publicar() {
		if (!sinPublicar) return;
		await fila;
		await fetch('/panel/mi-tienda/publicar', { method: 'POST' });
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Cambios publicados',
			desc: `${data.direccion} ya muestra la nueva versión`
		});
	}
	const visibles = $derived(secciones.filter((s) => s.visible));
</script>

<Seo title="Mi tienda · Panel" description="Editor de la tienda" path="/panel/mi-tienda" noindex />

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Mi tienda</h1>
			<div class="c-s">Plantilla {data.plantilla} · {data.direccion}</div>
		</div>
		<a class="cambiar" href="{resolve('/')}#plantillas"
			><Icon icon={I.phSwapBold} size={14} />Cambiar plantilla</a
		>
		<button type="button" class="publicar" class:listo={!sinPublicar} onclick={publicar}
			>{sinPublicar ? 'Publicar cambios' : 'Todo publicado'}</button
		>
	</div>

	<div class="editor" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="panelito">
			<div class="p-t">Secciones</div>
			{#each secciones as s, i (s.id)}
				{@const [icono, t] = SECCIONES[s.id]}
				<div class="sec">
					<Icon icon={I.phDotsSixVerticalBold} size={15} style="color:var(--fa)" />
					<Icon icon={icono} size={16} style="color:#C24E25" />
					<span class="s-t" class:oculta={!s.visible}>{t}</span>
					<button type="button" aria-label="Subir {t}" disabled={i === 0} onclick={() => subir(i)}
						><Icon icon={I.phCaretUpBold} size={12} /></button
					>
					<button
						type="button"
						aria-label="{s.visible ? 'Ocultar' : 'Mostrar'} {t}"
						onclick={() => alternar(i)}
						><Icon icon={s.visible ? I.phEye : I.phEyeSlash} size={14} /></button
					>
				</div>
			{/each}
			<div class="p-t colores">Colores</div>
			<div class="paletas" role="group" aria-label="Colores">
				{#each PALETAS as p, i (i)}
					<button
						type="button"
						aria-label="Paleta {i + 1}"
						aria-pressed={paleta === i}
						style:background="linear-gradient(135deg,{p.bg} 50%,{p.acc} 50%)"
						onclick={() => {
							paleta = i;
							void guardar();
						}}
					></button>
				{/each}
			</div>
		</div>

		<div class="navegador">
			<div class="barra">
				<span style:background="#E0A595"></span><span style:background="#E7CB8F"></span><span
					style:background="#A9CBA8"
				></span>
				<span class="url">{data.direccion}</span>
			</div>
			<div class="sitio" style:background={pal.bg} style:color={pal.ink}>
				<div class="s-cab">
					<b>{data.vista.tienda}</b><span class="esp"></span><span class="s-cats"
						>{data.vista.categorias.slice(0, 3).join(' · ')}</span
					><Icon icon={I.phHandbag} size={17} />
				</div>
				{#each visibles as s (s.id)}
					{#if s.id === 'hero'}
						<div class="hero">
							{#if data.vista.portada.foto}<img
									src={data.vista.portada.foto.startsWith('http')
										? data.vista.portada.foto
										: unsplash(data.vista.portada.foto, 1000)}
									alt=""
								/>{/if}
							<div class="sombra"></div>
							<div class="lema">{data.vista.portada.titulo}</div>
						</div>
					{:else if s.id === 'prods'}
						<div class="prods">
							{#each data.vista.productos as p, i (i)}
								<div>
									<div class="foto">
										{#if p.foto}<img src={unsplash(p.foto, 300)} alt="" />{/if}
									</div>
									<div class="pn">{p.nombre}</div>
									<div class="pp" style:color={pal.acc}>{pesos(p.precio)}</div>
								</div>
							{/each}
						</div>
					{:else if s.id === 'cats'}
						<div class="cats">
							{#each data.vista.categorias as c (c)}<span style:background={pal.acc}>{c}</span
								>{/each}
						</div>
					{:else if s.id === 'revs'}
						{#if data.vista.resena}
							<div class="revs">
								<span class="estrellas">★★★★★</span>
								{data.vista.resena.promedio.toLocaleString('es-AR')} · “{data.vista.resena.texto}”
							</div>
						{/if}
					{:else if s.id === 'ig'}
						<div class="ig">
							{#each data.vista.fotos as f, i (i)}<img src={unsplash(f, 200)} alt="" />{/each}
						</div>
					{:else}
						<div class="wa">
							<Icon icon={I.phWhatsappLogoFill} size={16} />¿Dudas? Escribinos, contesta Hornero al
							toque
						</div>
					{/if}
				{/each}
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
		overflow-wrap: anywhere;
	}
	.cambiar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 13.5px;
		font-weight: 600;
		box-shadow: var(--sh);
		white-space: nowrap;
		text-decoration: none;
	}
	.publicar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 20px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #e06a3e, #d2582f);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.publicar.listo {
		background: #c9b6a6;
	}
	.editor {
		display: grid;
		grid-template-columns: 300px minmax(0, 1fr);
		gap: 16px;
		align-items: start;
	}
	.panelito {
		padding: 14px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.p-t {
		font-size: 14px;
		font-weight: 600;
		padding: 2px 4px 6px;
	}
	.p-t.colores {
		margin-top: 8px;
		padding-bottom: 2px;
	}
	.sec {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 10px;
		border-radius: 14px;
		background: var(--sf);
		border: 1px solid var(--ln);
	}
	.s-t {
		flex: 1;
		font-size: 13.5px;
		font-weight: 600;
	}
	.s-t.oculta {
		opacity: 0.45;
	}
	.sec button {
		width: 26px;
		height: 26px;
		border: 0;
		border-radius: 8px;
		background: transparent;
		color: var(--mu);
		display: grid;
		place-items: center;
	}
	.sec button:disabled {
		opacity: 0.3;
	}
	.paletas {
		display: flex;
		gap: 8px;
		padding: 0 4px;
	}
	.paletas button {
		width: 36px;
		height: 36px;
		border-radius: 11px;
		border: 2px solid transparent;
		padding: 0;
	}
	.paletas button[aria-pressed='true'] {
		border-color: #d2582f;
	}
	.navegador {
		border-radius: 20px;
		overflow: hidden;
		background: #fff;
		box-shadow:
			0 30px 60px -30px rgba(42, 20, 12, 0.55),
			0 0 0 1px var(--ln);
	}
	.barra {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 9px 12px;
		background: #f1e9df;
	}
	.barra > span:not(.url) {
		width: 9px;
		height: 9px;
		border-radius: 50%;
	}
	.url {
		flex: 1;
		min-width: 0;
		margin: 0 30px;
		padding: 3px 10px;
		border-radius: 999px;
		background: #fff;
		font-size: 11.5px;
		color: #5c4536;
		text-align: center;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.sitio {
		transition:
			background 0.4s,
			color 0.4s;
		padding-bottom: 14px;
	}
	.s-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 16px;
	}
	.s-cab b {
		font-size: 15px;
	}
	.esp {
		flex: 1;
	}
	.s-cats {
		font-size: 12px;
		opacity: 0.7;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		min-width: 0;
	}
	.hero {
		position: relative;
		margin: 0 12px 12px;
		height: 170px;
		border-radius: 14px;
		overflow: hidden;
	}
	.hero img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.sombra {
		position: absolute;
		inset: 0;
		background: linear-gradient(90deg, rgba(0, 0, 0, 0.6), transparent 70%);
	}
	.lema {
		position: absolute;
		left: 18px;
		top: 50%;
		transform: translateY(-50%);
		color: #fff;
		font-size: 22px;
		font-weight: 700;
		line-height: 1.05;
		max-width: 220px;
	}
	.prods {
		padding: 0 12px 12px;
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 8px;
	}
	.foto {
		aspect-ratio: 1;
		border-radius: 10px;
		overflow: hidden;
		background: rgba(0, 0, 0, 0.05);
	}
	.foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.pn {
		margin-top: 4px;
		font-size: 11px;
		font-weight: 600;
	}
	.pp {
		font-size: 11px;
		font-weight: 700;
	}
	.cats {
		padding: 0 12px 12px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.cats span {
		font-size: 12px;
		font-weight: 600;
		padding: 7px 12px;
		border-radius: 999px;
		color: #fff;
	}
	.revs {
		margin: 0 12px 12px;
		padding: 12px;
		border-radius: 12px;
		background: rgba(0, 0, 0, 0.05);
		font-size: 12.5px;
		line-height: 1.4;
	}
	.estrellas {
		color: #e3a93b;
	}
	.ig {
		padding: 0 12px 12px;
		display: grid;
		grid-template-columns: repeat(6, 1fr);
		gap: 4px;
	}
	.ig img {
		width: 100%;
		aspect-ratio: 1;
		object-fit: cover;
		border-radius: 6px;
	}
	.wa {
		margin: 0 12px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 12px;
		border-radius: 12px;
		background: #25d366;
		color: #fff;
		font-size: 12.5px;
		font-weight: 600;
	}
	@media (max-width: 999px) {
		.editor {
			grid-template-columns: minmax(0, 1fr);
		}
	}
	@media (max-width: 560px) {
		.prods {
			grid-template-columns: repeat(2, 1fr);
		}
		.ig {
			grid-template-columns: repeat(3, 1fr);
		}
	}
</style>
