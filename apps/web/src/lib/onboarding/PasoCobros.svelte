<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import { COBROS_REALES, CONFIGURACION } from './datos.ts';
	import Etiqueta from './Etiqueta.svelte';
	import { getOnboarding } from './estado.svelte.ts';
	import Logo from './Logo.svelte';
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import Arca from '../../routes/panel/integraciones/Arca.svelte';
	import type { ArcaConfig, EstadoEntregas, Negocio } from '#lib/server/datos/tipos.ts';

	/**
	 * Paso 6, Cobros, envíos y facturas. En la demo, lo "ya configurado" para confirmar. Con la API,
	 * cada cosa se carga acá mismo (sin salir del onboarding) y nada es obligatorio para seguir.
	 */
	type Cobros = {
		negocio: Negocio | null;
		entregas: EstadoEntregas | null;
		arca: ArcaConfig | null;
	} | null;
	let { cobros = null }: { cobros?: Cobros } = $props();
	const o = getOnboarding();

	let abierto = $state<'transferencia' | 'entregas' | null>(null);
	let arcaAbierta = $state(false);
	let guardando = $state(false);
	let error = $state<{ donde: string; texto: string } | null>(null);
	const n = $derived(cobros?.negocio);
	const en = $derived(cobros?.entregas);
	const tieneTransf = $derived(!!(n?.alias || n?.cbu));
	let retiro = $state(true);
	let moto = $state(false);
	$effect.pre(() => {
		retiro = en?.retiro ?? true;
		moto = en?.moto ?? false;
	});

	const guardar = (donde: string) => () => {
		guardando = true;
		error = null;
		return async ({
			result,
			update
		}: {
			result: { type: string; data?: Record<string, unknown> };
			update: (o?: { reset?: boolean }) => Promise<void>;
		}) => {
			guardando = false;
			if (result.type === 'failure') {
				error = { donde, texto: String(result.data?.error ?? 'No se pudo guardar.') };
				return;
			}
			await update({ reset: false });
			await invalidateAll();
			abierto = null;
		};
	};
</script>

{#if o.demo}
	<Etiqueta rotulo="Cobros, envíos y facturas" titulo="Ya lo configuramos. Solo confirmá." />
	<div class="grupos">
		{#each CONFIGURACION as g (g.t)}
			<div class="grupo">
				<div class="g-cab">
					<span class="g-ico"><Icon icon={g.i} size={18} /></span><span>{g.t}</span>
				</div>
				<div class="items">
					{#each g.items as it (it.clave)}
						<div class="item">
							<Logo k={it.logo} tam={30} icono={16} radio={9} />
							<span class="i-txt"
								><span class="i-t">{it.t}</span><span class="i-d">{it.d}</span></span
							>
							{#if it.falta}
								<span class="falta">Falta 1 paso</span>
							{:else}
								<Interruptor
									checked={o.e.config[it.clave]}
									label={it.t}
									onchange={() => o.alternarConfig(it.clave)}
								/>
							{/if}
						</div>
					{/each}
				</div>
			</div>
		{/each}
	</div>
{:else}
	<Etiqueta rotulo="Cobros, envíos y facturas" titulo="Lo que falta para vender." />
	<div class="reales">
		<!-- Transferencia -->
		<div class="real" class:on={tieneTransf}>
			<div class="r-fila">
				<Logo k={COBROS_REALES[0].k} tam={36} icono={18} radio={11} />
				<span class="r-txt"
					><span class="r-t">Transferencia</span><span class="r-d"
						>{tieneTransf
							? `Alias ${n?.alias || '—'}${n?.titular ? ` · ${n.titular}` : ''}`
							: COBROS_REALES[0].d}</span
					></span
				>
				<button
					type="button"
					class="r-con"
					class:sec={tieneTransf}
					onclick={() => (abierto = abierto === 'transferencia' ? null : 'transferencia')}
					>{abierto === 'transferencia' ? 'Cerrar' : tieneTransf ? 'Editar' : 'Cargar'}</button
				>
			</div>
			{#if abierto === 'transferencia'}
				<form
					method="POST"
					action="?/transferencia"
					class="r-form"
					use:enhance={guardar('transferencia')}
				>
					<label
						><span>Alias</span><input
							name="alias"
							value={n?.alias ?? ''}
							placeholder="mi.tienda.mp"
							autocomplete="off"
						/></label
					>
					<label
						><span>CBU o CVU <em>(opcional)</em></span><input
							name="cbu"
							value={n?.cbu ?? ''}
							inputmode="numeric"
							placeholder="22 números"
							autocomplete="off"
						/></label
					>
					<label
						><span>Titular <em>(opcional)</em></span><input
							name="titular"
							value={n?.titular ?? ''}
							placeholder="Nombre y apellido o razón social"
						/></label
					>
					{#if error?.donde === 'transferencia'}<p class="r-err" role="alert">{error.texto}</p>{/if}
					<button class="r-guardar" disabled={guardando}
						>{guardando ? 'Guardando…' : 'Guardar'}</button
					>
				</form>
			{/if}
		</div>

		<!-- Envíos y retiro -->
		<div class="real" class:on={!!en}>
			<div class="r-fila">
				<Logo k={COBROS_REALES[1].k} tam={36} icono={18} radio={11} />
				<span class="r-txt"
					><span class="r-t">Envíos y retiro</span><span class="r-d"
						>{en
							? [
									en.retiro && 'Retiro en el local',
									en.moto && 'Moto',
									en.correo && 'Correo Argentino'
								]
									.filter(Boolean)
									.join(' · ') || 'Elegí cómo entregás'
							: COBROS_REALES[1].d}</span
					></span
				>
				<button
					type="button"
					class="r-con sec"
					onclick={() => (abierto = abierto === 'entregas' ? null : 'entregas')}
					>{abierto === 'entregas' ? 'Cerrar' : 'Editar'}</button
				>
			</div>
			{#if abierto === 'entregas'}
				<form method="POST" action="?/entregas" class="r-form" use:enhance={guardar('entregas')}>
					<label class="r-check"
						><input type="checkbox" name="retiro" bind:checked={retiro} />Retiro en el local</label
					>
					{#if retiro}
						<label
							><span>Dirección de retiro <em>(opcional)</em></span><input
								name="retiroDireccion"
								value={en?.retiroDireccion ?? ''}
								placeholder="Si la dejás vacía, va la del negocio"
							/></label
						>
						<label
							><span>Horario <em>(opcional)</em></span><input
								name="retiroHorario"
								value={en?.retiroHorario ?? ''}
								placeholder="Lunes a viernes de 9 a 18"
							/></label
						>
					{/if}
					<label class="r-check"
						><input type="checkbox" name="moto" bind:checked={moto} />Envío en moto</label
					>
					{#if moto}
						<label
							><span>Adónde llega</span><input
								name="motoZona"
								value={en?.motoZona ?? ''}
								placeholder="Barrios o localidades"
							/></label
						>
						<label
							><span>Costo <em>(opcional)</em></span><input
								name="motoCosto"
								value={en?.motoCosto ? String(en.motoCosto) : ''}
								inputmode="numeric"
								placeholder="$ 0"
							/></label
						>
						<label
							><span>Cuánto tarda <em>(opcional)</em></span><input
								name="motoDemora"
								value={en?.motoDemora ?? ''}
								placeholder="En el día"
							/></label
						>
					{/if}
					<p class="r-ayuda">Correo Argentino se conecta desde Integraciones, cuando quieras.</p>
					{#if error?.donde === 'entregas'}<p class="r-err" role="alert">{error.texto}</p>{/if}
					<button class="r-guardar" disabled={guardando}
						>{guardando ? 'Guardando…' : 'Guardar'}</button
					>
				</form>
			{/if}
		</div>

		<!-- Factura electrónica -->
		<div class="real" class:on={!!cobros?.arca?.conectada}>
			<div class="r-fila">
				<Logo k={COBROS_REALES[2].k} tam={36} icono={18} radio={11} />
				<span class="r-txt"
					><span class="r-t">Factura electrónica</span><span class="r-d"
						>{cobros?.arca?.conectada
							? `CUIT ${cobros.arca.cuit} · conectada`
							: COBROS_REALES[2].d}</span
					></span
				>
				{#if cobros?.arca}
					<button
						type="button"
						class="r-con"
						class:sec={cobros.arca.conectada}
						onclick={() => (arcaAbierta = true)}
						>{cobros.arca.conectada ? 'Ver' : 'Conectar'}</button
					>
				{/if}
			</div>
		</div>

		<!-- Cambios y devoluciones -->
		<div class="real on">
			<div class="r-fila">
				<Logo k={COBROS_REALES[3].k} tam={36} icono={18} radio={11} />
				<span class="r-txt"
					><span class="r-t">Cambios y devoluciones</span><span class="r-d"
						>Tu tienda ya muestra el botón de arrepentimiento (10 días, por ley).</span
					></span
				>
				<span class="r-ok"><Icon icon={I.phCheckBold} size={14} />Incluido</span>
			</div>
		</div>
	</div>
	<p class="r-nota">
		Nada de esto es obligatorio para seguir: lo completás ahora o después desde el&nbsp;panel.
	</p>
	{#if cobros?.arca}<Arca bind:open={arcaAbierta} arca={cobros.arca} demo={false} />{/if}
{/if}

<style>
	.grupos {
		margin-top: 22px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 360px), 1fr));
		gap: 16px;
	}
	.grupo {
		padding: 18px;
		border-radius: 22px;
		background: #fff;
		box-shadow:
			0 1px 2px rgba(58, 29, 18, 0.06),
			0 20px 44px -30px rgba(58, 29, 18, 0.4);
	}
	.g-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 16px;
		font-weight: 600;
	}
	.g-ico {
		width: 34px;
		height: 34px;
		border-radius: 10px;
		background: #fcede5;
		color: #b0441d;
		display: grid;
		place-items: center;
	}
	.items {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-top: 1px solid rgba(58, 29, 18, 0.07);
	}
	.i-txt {
		flex: 1;
		min-width: 0;
	}
	.i-t {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.i-d {
		display: block;
		font-size: 12.5px;
		color: #7a6252;
		line-height: 1.35;
	}
	.falta {
		font-size: 12px;
		font-weight: 700;
		color: #8a5f0e;
		background: rgba(227, 169, 59, 0.2);
		padding: 4px 9px;
		border-radius: 999px;
		white-space: nowrap;
	}
	.reales {
		margin-top: 22px;
		display: grid;
		gap: 10px;
	}
	.real {
		padding: 14px 16px;
		border-radius: 18px;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.1);
		color: #2a140c;
	}
	.real.on {
		border-color: rgba(47, 143, 99, 0.4);
	}
	.r-fila {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.r-txt {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.r-t {
		font-size: 15px;
		font-weight: 600;
	}
	.r-d {
		font-size: 13.5px;
		color: #6b5244;
		line-height: 1.35;
		overflow-wrap: anywhere;
	}
	.r-con {
		flex: none;
		padding: 10px 16px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 14px;
		font-weight: 600;
	}
	.r-con.sec {
		background: transparent;
		color: #b0441d;
		box-shadow: inset 0 0 0 1px rgba(176, 68, 29, 0.35);
	}
	.r-ok {
		flex: none;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 14px;
		font-weight: 600;
		color: #2f8f63;
	}
	.r-form {
		margin-top: 14px;
		padding-top: 14px;
		border-top: 1px solid rgba(58, 29, 18, 0.08);
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.r-form label {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
	}
	.r-form em {
		font-style: normal;
		font-weight: 400;
		color: #8a7262;
	}
	.r-form input:not([type='checkbox']) {
		padding: 12px 14px;
		border-radius: 12px;
		border: 1px solid rgba(58, 29, 18, 0.16);
		background: #fbf7f2;
		font: inherit;
		font-size: 16px;
		font-weight: 400;
		color: #2a140c;
	}
	.r-form .r-check {
		flex-direction: row;
		align-items: center;
		gap: 10px;
		font-size: 15px;
	}
	.r-check input {
		width: 20px;
		height: 20px;
		accent-color: #b0441d;
	}
	.r-ayuda {
		margin: 0;
		font-size: 13px;
		color: #6b5244;
	}
	.r-err {
		margin: 0;
		font-size: 13.5px;
		color: #b83a2e;
	}
	.r-guardar {
		align-self: flex-start;
		padding: 12px 22px;
		border-radius: 999px;
		border: 0;
		background: #2a140c;
		color: #fff;
		font-size: 15px;
		font-weight: 600;
	}
	.r-guardar:disabled {
		opacity: 0.6;
	}
	.r-nota {
		margin: 14px 0 0;
		font-size: 14px;
		color: #6b5244;
	}
</style>
