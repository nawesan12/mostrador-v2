<script lang="ts">
	import Hoja from '#lib/components/Hoja.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import type { components } from '@mostrador/contrato/esquema';
	import { pesos } from '../formato.ts';

	type Producto = components['schemas']['Producto'];

	/**
	 * Editor de producto (design/PanelCatalogo.dc.html, `edOpen`): fotos, nombre, descripción (con
	 * "Mejorar con Hornero"), precio por canal, stock, dónde se publica y cómo aparece en Google.
	 * Con `producto` null es el alta.
	 */
	let {
		open = $bindable(false),
		producto,
		tienda,
		dominio,
		guardado
	}: {
		open?: boolean;
		producto: Producto | null;
		tienda: string;
		dominio: string;
		guardado: (p: Producto, nuevo: boolean) => void;
	} = $props();
	const ui = getUi();

	let nombre = $state('');
	let descripcion = $state('');
	let precio = $state(0);
	let costo = $state(0);
	let stock = $state(0);
	let categoria = $state('');
	let canales = $state<Record<string, boolean>>({ web: true, ml: false, ig: false, google: false });
	let error = $state<string | null>(null);
	let guardando = $state(false);
	let mejorando = $state(false);

	$effect(() => {
		if (!open) return;
		const p = producto;
		nombre = p?.nombre ?? '';
		descripcion = p?.descripcion ?? '';
		precio = p?.precio ?? 0;
		costo = p?.costo ?? 0;
		stock = p?.stock ?? 0;
		categoria = p?.categoria ?? '';
		canales = { web: true, ml: false, ig: false, google: false, ...(p?.canales ?? {}) };
		error = null;
	});

	const margen = $derived(precio ? Math.round(((precio - costo) / precio) * 100) : 0);
	const slug = $derived(
		nombre
			.normalize('NFD')
			.replace(/[̀-ͯ]/g, '')
			.toLowerCase()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/^-|-$/g, '') || 'producto'
	);
	const PRECIOS = $derived([
		['Tienda y local', precio, 'Precio base'],
		['Mercado Libre', Math.round((precio * 1.1) / 10) * 10, '+10% por comisión'],
		['Mayorista', Math.round((precio * 0.75) / 10) * 10, 'Desde 10 u']
	] as const);
	const CANALES = [
		['web', I.phGlobeSimpleFill, 'Tienda online', '#D2582F'],
		['ml', I.phHandshakeFill, 'Mercado Libre', '#2D3277'],
		['ig', I.logosInstagramIcon, 'Instagram Shopping', 'inherit'],
		['google', I.logosGoogleIcon, 'Google Shopping', 'inherit']
	] as const;
	const numero = (s: string) => Math.max(0, Number(s.replace(/\D/g, '')) || 0);

	async function mejorar() {
		if (nombre.trim().length < 2) {
			error = 'Poné el nombre así Hornero sabe de qué se trata.';
			return;
		}
		mejorando = true;
		const r = await fetch('/panel/hornero/mejorar', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ nombre, descripcion })
		});
		mejorando = false;
		if (r.ok) {
			const anterior = descripcion;
			descripcion = ((await r.json()) as { texto: string }).texto;
			ui.toast({
				k: 'ai',
				title: 'Descripción mejorada',
				desc: 'Revisala antes de guardar',
				action: 'Deshacer',
				onAction: () => {
					descripcion = anterior;
				}
			});
		}
	}

	async function guardar() {
		error = null;
		if (nombre.trim().length < 2) {
			error = 'Poné un nombre de al menos 2 letras.';
			return;
		}
		guardando = true;
		const r = await fetch(`/panel/productos/${producto?.id ?? 'nuevo'}`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				nombre,
				descripcion,
				categoria: categoria.trim() || null,
				precio,
				costo,
				stock,
				foto: producto?.foto ?? null,
				canales
			})
		});
		guardando = false;
		if (!r.ok) {
			error =
				((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
				'No pudimos guardarlo.';
			return;
		}
		open = false;
		guardado((await r.json()) as Producto, !producto);
	}

	const fotoPronto = () =>
		ui.toast({
			k: 'info',
			title: 'Subir fotos llega con el almacenamiento',
			desc: 'Mientras, el producto queda con la foto que tiene.'
		});
</script>

<Hoja bind:open titulo={producto ? 'Editar producto' : 'Nuevo producto'} ancho={560}>
	{#snippet cabecera()}
		<div class="cab">
			<span class="c-t">{producto ? 'Editar producto' : 'Nuevo producto'}</span>
			<button type="button" class="redondo" aria-label="Cerrar" onclick={() => (open = false)}
				><Icon icon={I.phXBold} size={14} /></button
			>
		</div>
	{/snippet}

	<div class="cuerpo">
		<div class="fotos">
			{#if producto?.foto}
				<img class="f-1" src={unsplash(producto.foto, 500, '&h=500')} alt="" />
				<img src={unsplash(producto.foto, 200, '&h=200')} alt="" style="object-position:20% 30%" />
				<img src={unsplash(producto.foto, 200, '&h=200')} alt="" style="object-position:80% 70%" />
			{:else}
				<button type="button" class="f-1 f-vacia" onclick={fotoPronto}
					><Icon icon={I.phCameraPlusFill} size={28} /><span>Sumá la primera foto</span></button
				>
			{/if}
			<button type="button" class="f-mas" aria-label="Sumar una foto" onclick={fotoPronto}
				><Icon icon={I.phPlusBold} size={18} /></button
			>
			<button type="button" class="f-magia" onclick={fotoPronto}
				><Icon icon={I.phMagicWandFill} size={16} />Mejorar fondo</button
			>
		</div>

		<div class="tarjeta campos">
			<label class="campo"
				>Nombre<input
					bind:value={nombre}
					maxlength="120"
					aria-invalid={error && nombre.trim().length < 2 ? 'true' : undefined}
				/></label
			>
			<div class="campo">
				<span class="c-fila"
					><label for="desc-prod">Descripción</label><button
						type="button"
						class="hornero"
						onclick={mejorar}
						disabled={mejorando}
						>{mejorando ? 'Hornero está escribiendo…' : '✦ Mejorar con Hornero'}</button
					></span
				>
				<textarea id="desc-prod" bind:value={descripcion} maxlength="4000"></textarea>
			</div>
			<label class="campo"
				>Categoría<input
					bind:value={categoria}
					maxlength="60"
					placeholder="Por ejemplo: Mates"
				/></label
			>
		</div>

		<div class="tarjeta">
			<div class="t">Precio por canal</div>
			<div class="precios">
				{#each PRECIOS as [t, v, nota], i (t)}
					<div class="precio">
						<div class="p-t">{t}</div>
						{#if i === 0}
							<input
								class="p-in"
								inputmode="numeric"
								aria-label="Precio base"
								value={pesos(precio)}
								onchange={(e) => (precio = numero(e.currentTarget.value))}
							/>
						{:else}
							<div class="p-v">{pesos(v)}</div>
						{/if}
						<div class="p-n">{nota}</div>
					</div>
				{/each}
			</div>
			<div class="margen">
				Margen <b class:bajo={margen < 25}>{margen}%</b> · costo
				<input
					class="costo"
					inputmode="numeric"
					aria-label="Costo"
					value={pesos(costo)}
					onchange={(e) => (costo = numero(e.currentTarget.value))}
				/>
			</div>
		</div>

		<div class="tarjeta">
			<div class="t">Variantes y stock</div>
			<div class="variantes">
				<div class="variante">
					<span class="v-color"></span><span class="v-n">Única</span>
					<input
						class="v-s"
						inputmode="numeric"
						aria-label="Stock"
						value={stock}
						onchange={(e) => (stock = numero(e.currentTarget.value))}
					/><span class="v-u">u</span>
				</div>
			</div>
		</div>

		<div class="tarjeta canales">
			<div class="t">Dónde se publica</div>
			{#each CANALES as [k, icon, t, col] (k)}
				<div class="canal">
					<Icon {icon} size={18} style="color:{col}" />
					<span>{t}</span>
					<Interruptor label={t} bind:checked={canales[k]} />
				</div>
			{/each}
		</div>

		<div class="tarjeta google">
			<div class="g-t">Así aparece en Google</div>
			<div class="g-u">{dominio} › {slug}</div>
			<div class="g-n">{nombre || 'Nombre del producto'} | {tienda}</div>
			<div class="g-d">{descripcion || 'La descripción aparece acá.'}</div>
		</div>

		{#if error}<div class="error" role="alert">{error}</div>{/if}
	</div>

	{#snippet pie()}
		<div class="pie">
			<button type="button" class="cancelar" onclick={() => (open = false)}>Cancelar</button>
			<button type="button" class="guardar" onclick={guardar} disabled={guardando}
				>Guardar y publicar</button
			>
		</div>
	{/snippet}
</Hoja>

<style>
	.cab {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.c-t {
		flex: 1;
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.redondo {
		width: 36px;
		height: 36px;
		border: 0;
		border-radius: 50%;
		background: var(--chip);
		color: var(--tx);
		display: grid;
		place-items: center;
	}
	.cuerpo {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.fotos {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr;
		gap: 8px;
	}
	.fotos img {
		width: 100%;
		aspect-ratio: 1;
		border-radius: 12px;
		object-fit: cover;
	}
	.fotos .f-1 {
		grid-row: span 2;
		border-radius: 16px;
	}
	.f-vacia {
		aspect-ratio: auto;
		width: 100%;
		height: 100%;
		border: 1.5px dashed var(--ln2);
		background: var(--sf);
		color: var(--mu);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
	}
	.f-mas {
		aspect-ratio: 1;
		border-radius: 12px;
		border: 1.5px dashed var(--ln2);
		background: transparent;
		display: grid;
		place-items: center;
		color: var(--mu);
	}
	.f-magia {
		aspect-ratio: 1;
		border-radius: 12px;
		border: 0;
		background: rgba(210, 88, 47, 0.1);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 2px;
		color: #c24e25;
		font-size: 11px;
		font-weight: 600;
		text-align: center;
		padding: 4px;
	}
	.tarjeta {
		padding: 14px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.campos {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.campo input,
	.campo textarea {
		padding: 11px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font-size: 15px;
		font-weight: 500;
		color: var(--tx);
		outline: none;
	}
	.campo textarea {
		min-height: 72px;
		font-size: 14px;
		line-height: 1.5;
		resize: vertical;
	}
	.campo input:focus,
	.campo textarea:focus {
		border-color: #d2582f;
	}
	.c-fila {
		display: flex;
		justify-content: space-between;
	}
	.hornero {
		border: 0;
		background: transparent;
		padding: 0;
		font-size: 12.5px;
		font-weight: 600;
		color: #c24e25;
	}
	.t {
		font-size: 14px;
		font-weight: 600;
	}
	.precios {
		margin-top: 10px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 8px;
	}
	.precio {
		padding: 10px 12px;
		border-radius: 12px;
		background: var(--sf3);
	}
	.p-t {
		font-size: 12px;
		color: var(--mu);
	}
	.p-v,
	.p-in {
		font-size: 18px;
		font-weight: 600;
	}
	.p-in {
		width: 100%;
		border: 0;
		border-bottom: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--tx);
		padding: 0;
		outline: none;
	}
	.p-in:focus {
		border-bottom-color: #d2582f;
	}
	.p-n {
		font-size: 11.5px;
		color: var(--fa);
	}
	.margen {
		margin-top: 10px;
		font-size: 12.5px;
		color: var(--mu);
		display: flex;
		align-items: center;
		gap: 4px;
		flex-wrap: wrap;
	}
	.margen b {
		color: #2f8f63;
	}
	.margen b.bajo {
		color: #c24e25;
	}
	.costo {
		width: 110px;
		border: 0;
		border-bottom: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 12.5px;
		font-weight: 600;
		outline: none;
	}
	.variantes {
		margin-top: 8px;
	}
	.variante {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 9px 0;
		border-bottom: 1px solid var(--ln);
	}
	.v-color {
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: var(--sf3);
		box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.12);
	}
	.v-n {
		flex: 1;
		font-size: 14px;
	}
	.v-s {
		width: 64px;
		text-align: right;
		border: 0;
		border-bottom: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 13px;
		font-weight: 600;
		outline: none;
	}
	.v-u {
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	.canales {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.canales .t {
		margin-bottom: 4px;
	}
	.canal {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 7px 0;
	}
	.canal span {
		flex: 1;
		font-size: 14px;
	}
	.g-t {
		font-size: 12px;
		color: var(--mu);
	}
	.g-u {
		margin-top: 6px;
		font-size: 12.5px;
		color: #2f8f63;
		overflow-wrap: anywhere;
	}
	.g-n {
		font-size: 16px;
		color: #3e6bfd;
		font-weight: 500;
	}
	.g-d {
		font-size: 13px;
		color: var(--mu);
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.error {
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	.pie {
		width: 100%;
		display: flex;
		gap: 8px;
	}
	.cancelar {
		padding: 13px 18px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 14px;
		font-weight: 600;
	}
	.guardar {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #e06a3e, #d2582f);
		color: #fff;
		font-size: 14.5px;
		font-weight: 600;
	}
</style>
