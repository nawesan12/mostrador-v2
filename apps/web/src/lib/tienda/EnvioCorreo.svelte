<script lang="ts">
	/**
	 * Correo Argentino en el checkout de cualquier plantilla: avisa mientras cotiza (o si no pudo) y,
	 * con Correo elegido, pide la provincia y la localidad (a domicilio) o la sucursal (a sucursal).
	 * Toma los colores y la letra de la plantilla (currentColor e inherit).
	 */
	import { estadoCorreo, esCorreo, reintentarCorreo } from './envio.svelte.ts';
	import { PROVINCIAS, provinciaDeCP } from './provincias.ts';
	import type { OpcionEnvio, TiendaPublica } from './tipos.ts';

	type Sucursal = {
		codigo: string;
		nombre: string;
		direccion: string;
		localidad: string;
		cp: string;
		horario: string;
	};

	let {
		tienda,
		cp,
		envio
	}: {
		tienda: Pick<TiendaPublica, 'slug' | 'envios' | 'direccion' | 'vitrina'>;
		cp: string;
		envio: OpcionEnvio['id'] | null;
	} = $props();

	const estado = $derived(estadoCorreo(cp, tienda));
	let provincia = $state('');
	let localidad = $state('');
	let sucursal = $state('');
	let sucursales = $state<Sucursal[] | null>(null);
	let cargando = $state(false);
	let fallo = $state(false);

	// El CP de CABA o del conurbano ya dice la provincia; si cambia el CP, se vuelve a sugerir.
	let sugerida = '';
	$effect(() => {
		const p = provinciaDeCP(cp) ?? '';
		if (p !== sugerida) {
			sugerida = p;
			if (p) provincia = p;
		}
	});

	// Sucursales de la provincia, las más cercanas al CP primero.
	let pedido = 0;
	$effect(() => {
		if (envio !== 'correo-sucursal' || !provincia) return;
		const n = ++pedido;
		const q = new URLSearchParams({ provincia, cp: cp.trim() });
		cargando = true;
		fallo = false;
		fetch(`/t/${encodeURIComponent(tienda.slug)}/correo/sucursales?${q}`)
			.then(async (r) => {
				if (!r.ok) throw new Error(String(r.status));
				const l = (await r.json()) as Sucursal[];
				if (n !== pedido) return;
				sucursales = l;
				if (!l.some((s) => s.codigo === sucursal)) sucursal = l[0]?.codigo ?? '';
			})
			.catch(() => {
				if (n === pedido) {
					sucursales = null;
					fallo = true;
				}
			})
			.finally(() => {
				if (n === pedido) cargando = false;
			});
	});

	const elegida = $derived(sucursales?.find((s) => s.codigo === sucursal) ?? null);
</script>

{#if estado === 'cotizando'}
	<p class="aviso" aria-live="polite">Cotizando con Correo Argentino…</p>
{:else if estado === 'error'}
	<p class="aviso" aria-live="polite">
		No pudimos cotizar con Correo Argentino.
		<button type="button" class="link" onclick={() => reintentarCorreo(cp, tienda)}
			>Probar de nuevo</button
		>
	</p>
{/if}

{#if esCorreo(envio)}
	<div class="correo">
		<label class="campo">
			<span>Provincia</span>
			<select name="provincia" bind:value={provincia} required>
				<option value="" disabled>Elegí la provincia</option>
				{#each PROVINCIAS as [c, n] (c)}<option value={c}>{n}</option>{/each}
			</select>
		</label>
		{#if envio === 'correo-domicilio'}
			<label class="campo">
				<span>Localidad</span>
				<input
					name="localidad"
					bind:value={localidad}
					maxlength="80"
					autocomplete="address-level2"
					required
				/>
			</label>
		{:else}
			<label class="campo ancho">
				<span>Sucursal donde retirás</span>
				{#if cargando}
					<span class="nota">Buscando sucursales…</span>
				{:else if fallo}
					<span class="nota">No pudimos traer las sucursales. Probá con otra provincia.</span>
				{:else if sucursales && !sucursales.length}
					<span class="nota">No hay sucursales para retirar en esta provincia.</span>
				{:else if sucursales}
					<select name="sucursal" bind:value={sucursal} required>
						{#each sucursales as s (s.codigo)}<option value={s.codigo}
								>{s.nombre} · {s.direccion}</option
							>{/each}
					</select>
				{:else}
					<span class="nota">Elegí la provincia para ver las sucursales.</span>
				{/if}
			</label>
			{#if elegida}
				<p class="detalle">
					{elegida.direccion}, {elegida.localidad}{#if elegida.horario}
						· {elegida.horario}{/if}
				</p>
			{/if}
		{/if}
	</div>
{/if}

<style>
	.aviso {
		margin: 8px 0 0;
		font-size: 13px;
		opacity: 0.75;
		text-wrap: pretty;
	}
	.link {
		padding: 0;
		border: 0;
		background: none;
		font: inherit;
		color: inherit;
		text-decoration: underline;
		cursor: pointer;
		min-height: 44px;
	}
	.correo {
		margin-top: 10px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 200px), 1fr));
		gap: 10px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 10px 12px;
		border: 1.5px solid color-mix(in srgb, currentColor 22%, transparent);
		border-radius: 10px;
		min-width: 0;
	}
	.ancho {
		grid-column: 1 / -1;
	}
	.campo > span:first-child {
		font-size: 11.5px;
		opacity: 0.7;
	}
	.campo select,
	.campo input {
		width: 100%;
		min-width: 0;
		min-height: 28px;
		padding: 0;
		border: 0;
		background: transparent;
		font: inherit;
		font-size: 14.5px;
		color: inherit;
		outline: none;
	}
	.campo select option {
		color: #1d1d1b;
	}
	.nota {
		font-size: 13.5px;
		opacity: 0.75;
		text-wrap: pretty;
	}
	.detalle {
		grid-column: 1 / -1;
		margin: 0;
		font-size: 12.5px;
		opacity: 0.75;
		text-wrap: pretty;
	}
</style>
