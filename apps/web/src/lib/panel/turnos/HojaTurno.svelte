<script lang="ts">
	import Hoja from '#lib/components/Hoja.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { duracion, cuandoLargo } from '#lib/turnos/formato.ts';
	import { ESTADO_TURNO, waDe, type EstadoTurno, type Turno } from '../turnos.ts';

	/**
	 * Detalle de un turno (no está en el handoff: con las piezas de la hoja de pedido): cliente,
	 * cuándo y con quién, seña, y las acciones confirmar / atendido / ausente / cancelar / devolver
	 * la seña.
	 */
	let {
		open = $bindable(false),
		turno,
		cambiar,
		devolver
	}: {
		open?: boolean;
		turno: Turno | null;
		cambiar: (t: Turno, estado: EstadoTurno) => Promise<void>;
		devolver: (t: Turno) => Promise<void>;
	} = $props();
	let ocupado = $state(false);
	const accion = async (f: () => Promise<void>) => {
		ocupado = true;
		try {
			await f();
		} finally {
			ocupado = false;
		}
	};
	const minutos = (t: Turno) => Math.round((Date.parse(t.fin) - Date.parse(t.inicio)) / 60_000);
</script>

<Hoja bind:open titulo={turno ? `Turno de ${turno.cliente.nombre}` : 'Turno'} ancho={480}>
	{#snippet cabecera()}
		{#if turno}
			{@const e = ESTADO_TURNO[turno.estado]}
			<div class="cab">
				<div class="c-txt">
					<div class="c-fila">
						<span class="c-n">{turno.servicio}</span><span
							class="chip-estado"
							style:--chip-bg={e.c[0]}
							style:--chip-fg={e.c[1]}>{turno.vencido ? 'Venció la seña' : e.t}</span
						>
					</div>
					<div class="c-s">{cuandoLargo(turno.inicio)} · con {turno.profesional}</div>
				</div>
				<button
					type="button"
					class="redondo"
					title="Cerrar"
					aria-label="Cerrar"
					onclick={() => (open = false)}><Icon icon={I.phXBold} size={15} /></button
				>
			</div>
		{/if}
	{/snippet}

	{#if turno}
		{@const wa = waDe(turno.cliente.telefono)}
		<div class="cuerpo">
			<div class="tarjeta fila">
				<span class="av">{turno.cliente.nombre.slice(0, 1).toUpperCase()}</span>
				<div class="f-txt">
					<div class="f-n">{turno.cliente.nombre}</div>
					<div class="f-s">
						{[turno.cliente.telefono, turno.cliente.email].filter(Boolean).join(' · ') ||
							'Sin datos de contacto'}
					</div>
				</div>
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- WhatsApp, link externo -->
				{#if wa}<a class="wa" href={wa} target="_blank" rel="noopener noreferrer"
						><Icon icon={I.logosWhatsappIcon} size={14} />Escribir</a
					>{/if}
			</div>

			<div class="tarjeta">
				<div class="t">Turno</div>
				<div class="suma"><span>Servicio</span><span>{turno.servicio}</span></div>
				<div class="suma"><span>Duración</span><span>{duracion(minutos(turno))}</span></div>
				<div class="suma"><span>Profesional</span><span>{turno.profesional}</span></div>
				<div class="suma">
					<span>Reservado</span><span
						>{turno.origen === 'tienda' ? 'Desde la tienda' : 'Desde el panel'}</span
					>
				</div>
				<div class="total"><span>Precio</span><span>{pesos(turno.precio)}</span></div>
			</div>

			<div class="tarjeta fila">
				<span class="i-ico"><Icon icon={I.phCurrencyCircleDollarFill} size={20} /></span>
				<span class="f-txt">
					<span class="f-n">{turno.sena ? `Seña de ${pesos(turno.sena)}` : 'Sin seña'}</span>
					<span class="f-s"
						>{turno.senaDevuelta
							? 'Devuelta por Mercado Pago'
							: turno.senaPagada
								? 'Pagada con Mercado Pago'
								: turno.sena
									? turno.vencido
										? 'No se pagó a tiempo: el horario quedó libre'
										: 'Esperando el pago'
									: 'Se paga todo en el local'}</span
					>
				</span>
			</div>

			{#if turno.nota}
				<div class="nota">
					<Icon
						icon={I.phNotePencilFill}
						size={16}
						style="color:#C24E25;margin-top:2px;flex:none"
					/>
					<div>{turno.nota}</div>
				</div>
			{/if}

			{#if turno.estado === 'cancelado' && turno.senaPagada && !turno.senaDevuelta}
				<div class="nota">
					<Icon icon={I.phWarningFill} size={16} style="color:#9A6A10;margin-top:2px;flex:none" />
					<div>
						Cancelar no devuelve la seña sola. Si corresponde, devolvela desde acá con
						Mercado&nbsp;Pago.
					</div>
				</div>
			{/if}
		</div>
	{/if}

	{#snippet pie()}
		{#if turno}
			<div class="acciones">
				{#if turno.estado === 'reservado'}
					<button
						type="button"
						class="btn btn-primary"
						disabled={ocupado}
						onclick={() => accion(() => cambiar(turno!, 'confirmado'))}>Confirmar</button
					>
				{/if}
				{#if turno.estado === 'confirmado' || turno.estado === 'reservado'}
					<button
						type="button"
						class="btn btn-dark"
						disabled={ocupado}
						onclick={() => accion(() => cambiar(turno!, 'atendido'))}>Atendido</button
					>
					<button
						type="button"
						class="btn btn-secondary"
						disabled={ocupado}
						onclick={() => accion(() => cambiar(turno!, 'ausente'))}>No vino</button
					>
					<button
						type="button"
						class="btn btn-secondary peligro"
						disabled={ocupado}
						onclick={() => accion(() => cambiar(turno!, 'cancelado'))}>Cancelar turno</button
					>
				{/if}
				{#if turno.senaPagada && !turno.senaDevuelta}
					<button
						type="button"
						class="btn btn-secondary"
						disabled={ocupado}
						onclick={() => accion(() => devolver(turno!))}>Devolver la seña</button
					>
				{/if}
			</div>
		{/if}
	{/snippet}
</Hoja>

<style>
	.cab {
		display: flex;
		align-items: flex-start;
		gap: 10px;
	}
	.c-txt {
		flex: 1;
		min-width: 0;
	}
	.c-fila {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.c-n {
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.c-s {
		margin-top: 3px;
		font-size: 13.5px;
		color: var(--mu);
	}
	.redondo {
		width: 36px;
		height: 36px;
		flex: none;
		display: grid;
		place-items: center;
		border: 0;
		border-radius: 50%;
		background: var(--chip);
		color: var(--tx);
	}
	.cuerpo {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.tarjeta {
		padding: 14px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.av,
	.i-ico {
		width: 40px;
		height: 40px;
		flex: none;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: var(--chip);
		font-weight: 600;
	}
	.i-ico {
		border-radius: 12px;
		color: #2f8f63;
		background: rgba(47, 143, 99, 0.12);
	}
	.f-txt {
		flex: 1;
		min-width: 0;
	}
	.f-n {
		display: block;
		font-size: 14.5px;
		font-weight: 600;
	}
	.f-s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		overflow-wrap: anywhere;
	}
	.wa {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		color: var(--tx);
		font-size: 13px;
		font-weight: 600;
		text-decoration: none;
		white-space: nowrap;
	}
	.t {
		font-size: 14px;
		font-weight: 600;
		margin-bottom: 6px;
	}
	.suma,
	.total {
		display: flex;
		justify-content: space-between;
		gap: 10px;
		padding: 5px 0;
		font-size: 13.5px;
	}
	.suma span:first-child {
		color: var(--mu);
	}
	.suma span:last-child {
		text-align: right;
	}
	.total {
		margin-top: 4px;
		padding-top: 10px;
		border-top: 1px solid var(--ln);
		font-size: 15px;
		font-weight: 600;
	}
	.nota {
		display: flex;
		gap: 10px;
		padding: 12px 14px;
		border-radius: 16px;
		background: rgba(227, 169, 59, 0.12);
		font-size: 13.5px;
		line-height: 1.45;
	}
	.acciones {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		width: 100%;
	}
	.acciones .btn {
		flex: 1 1 auto;
	}
	.peligro {
		color: var(--error);
	}
</style>
