<script lang="ts">
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { centavos, ESTADOS_TIENDA, fechaCorta, nombrePlan } from '#lib/staff/estado.ts';
	import type { Permiso } from '#lib/staff/permisos.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import type { SubmitFunction } from '$app/forms';

	/** La ficha de una tienda para el equipo: datos, plan, uso, salud, actividad, notas y acciones. */
	let { data } = $props();
	const ui = getUi();
	const f = $derived(data.ficha);
	const t = $derived(f.tienda);
	const e = $derived(ESTADOS_TIENDA[t.estado]);
	const puede = (p: Permiso) => data.sesion.permisos.includes(p);

	type Clave =
		| 'extender_prueba'
		| 'regalar_pack'
		| 'cambiar_plan'
		| 'aplicar_credito'
		| 'pausar'
		| 'reactivar'
		| 'cerrar';
	const ACCIONES: {
		k: Clave;
		t: string;
		icon: typeof I.phStarFill;
		permiso: Permiso;
		peligro?: boolean;
	}[] = [
		{
			k: 'extender_prueba',
			t: 'Extender la prueba',
			icon: I.phHourglassFill,
			permiso: 'tiendas.prueba'
		},
		{ k: 'regalar_pack', t: 'Regalar pack de IA', icon: I.phSparkleFill, permiso: 'tiendas.pack' },
		{ k: 'cambiar_plan', t: 'Cambiar el plan', icon: I.phSwapBold, permiso: 'tiendas.plan' },
		{
			k: 'aplicar_credito',
			t: 'Aplicar crédito',
			icon: I.phCreditCardFill,
			permiso: 'tiendas.credito'
		},
		{
			k: 'pausar',
			t: 'Pausar la tienda',
			icon: I.phPauseFill,
			permiso: 'tiendas.pausar',
			peligro: true
		},
		{
			k: 'reactivar',
			t: 'Reactivar la tienda',
			icon: I.phArrowCounterClockwiseBold,
			permiso: 'tiendas.pausar'
		},
		{
			k: 'cerrar',
			t: 'Cerrar la tienda',
			icon: I.phXCircleFill,
			permiso: 'tiendas.cerrar',
			peligro: true
		}
	];
	const visibles = $derived(
		ACCIONES.filter(
			(a) =>
				puede(a.permiso) &&
				t.estado !== 'cerrada' &&
				(a.k !== 'reactivar' || t.estado === 'pausada') &&
				(a.k !== 'pausar' || t.estado !== 'pausada') &&
				(a.k !== 'extender_prueba' || t.estado === 'prueba')
		)
	);

	let abierta = $state<(typeof ACCIONES)[number] | null>(null);
	let modal = $state(false);
	let enviando = $state(false);
	let error = $state<string | null>(null);
	function abrir(a: (typeof ACCIONES)[number]) {
		abierta = a;
		error = null;
		modal = true;
	}

	/** Lo que vuelve de cada formulario: listo, error de Go o "en la demo no se aplica". */
	const enviar =
		(listo: () => void): SubmitFunction =>
		() => {
			enviando = true;
			error = null;
			return async ({ result, update }) => {
				enviando = false;
				if (result.type === 'success') {
					listo();
					await update();
				} else if (result.type === 'failure') {
					if (result.data?.demo) {
						modal = false;
						ui.toast({
							k: 'info',
							title: 'En la demo no se aplica',
							desc: 'Con la cuenta real queda hecho y en la auditoría'
						});
					} else error = (result.data?.error as string) ?? 'No se pudo.';
				} else await update();
			};
		};

	const NEGOCIO: Record<string, string> = {
		razonSocial: 'Razón social',
		cuit: 'CUIT',
		condicion: 'Condición frente al IVA',
		telefono: 'Teléfono',
		direccion: 'Dirección',
		email: 'Email',
		whatsapp: 'WhatsApp'
	};
	const SUSCRIPCION: Record<string, string> = {
		authorized: 'Activa',
		pending: 'Esperando el primer cobro',
		paused: 'Pausada',
		cancelled: 'Cancelada'
	};
	const QUIEN = {
		persona: 'La tienda',
		hornero: 'Tincho',
		sistema: 'Tu Local Web',
		staff: 'Equipo',
		integracion: 'Integración'
	} as const;
</script>

<Seo
	title="{t.nombre} · Tiendas · Equipo"
	description="Ficha de la tienda"
	path="/staff/tiendas/{t.id}"
	noindex
/>

<a class="volver" href={resolve('/staff/(equipo)/tiendas')}
	><Icon icon={I.phArrowLeftBold} size={13} />Tiendas</a
>
<div class="cab">
	<h1>{t.nombre}</h1>
	<span class="chip" style:background={e.bg} style:color={e.c}>{e.t}</span>
	{#if puede('tiendas.mirar') && t.estado !== 'cerrada'}
		<form method="POST" action="?/mirar" use:enhance={enviar(() => {})}>
			<button type="submit" class="borde"
				><Icon icon={I.phEyeFill} size={15} />Ver como la dueña</button
			>
		</form>
	{/if}
</div>
<p class="sub">
	{t.slug} · {t.rubro} · plantilla {t.plantilla} · alta el {fechaCorta(t.creadaEn)}{t.partner
		? ` · la trajo ${t.partner}`
		: ''}
</p>
{#if t.motivoBaja}<p class="baja">Motivo de la baja: {t.motivoBaja}</p>{/if}

<div class="grilla">
	<section class="tarjeta">
		<div class="t-t">Negocio y titular</div>
		<div class="fila">
			<span class="f-txt"
				><b>{f.titular?.nombre ?? 'Sin titular'}</b><span class="f-s">{f.titular?.email ?? ''}</span
				></span
			>
		</div>
		{#each Object.entries(f.negocio) as [k, v] (k)}
			{#if v}<div class="fila">
					<span class="f-s dato">{NEGOCIO[k] ?? k}</span><span>{v}</span>
				</div>{/if}
		{/each}
		<div class="fila">
			<span class="f-s dato">Personas en la cuenta</span><span>{f.personas}</span>
		</div>
	</section>

	<section class="tarjeta">
		<div class="t-t">Plan y suscripción</div>
		<div class="fila">
			<span class="f-txt"
				><b>{nombrePlan(f.plan.actual)} · {f.plan.ciclo === 'anio' ? 'anual' : 'mensual'}</b><span
					class="f-s">{f.plan.precio ? `${centavos(f.plan.precio)} por mes` : 'Sin cargo'}</span
				></span
			>
		</div>
		{#if f.plan.enPrueba}<div class="fila">
				<span class="f-s dato">Prueba hasta</span><span>{fechaCorta(f.plan.pruebaHasta)}</span>
			</div>{/if}
		<div class="fila">
			<span class="f-s dato">Próximo cobro</span><span
				>{f.plan.proximoCobro ? fechaCorta(f.plan.proximoCobro) : '—'}</span
			>
		</div>
		<div class="fila">
			<span class="f-s dato">Suscripción de Mercado Pago</span><span class="corto"
				>{f.plan.suscripcion
					? (SUSCRIPCION[f.plan.suscripcion] ?? f.plan.suscripcion)
					: 'Sin suscripción'}</span
			>
		</div>
		<div class="fila">
			<span class="f-s dato">Crédito a favor</span><span>{centavos(f.plan.credito)}</span>
		</div>
	</section>

	<section class="tarjeta">
		<div class="t-t">Uso del mes</div>
		<div class="kpis chicos">
			<div><b>{f.uso.productos}</b><span class="f-s">productos</span></div>
			<div><b>{f.uso.pedidosMes}</b><span class="f-s">pedidos</span></div>
			<div><b>{f.uso.packsMes}</b><span class="f-s">packs de IA</span></div>
		</div>
		<div class="t-t sep">Integraciones</div>
		{#each f.integraciones as i (i.id)}
			<div class="fila">
				<span class="f-txt">{i.nombre}</span><span
					class="chip"
					style:background={i.salud === 'ok' ? 'rgba(47,143,99,.13)' : 'rgba(184,58,46,.1)'}
					style:color={i.salud === 'ok' ? '#2F8F63' : '#B83A2E'}
					>{i.salud === 'ok' ? 'Funciona' : 'Reconectar'}</span
				>
			</div>
		{:else}
			<div class="vacio">Sin integraciones conectadas.</div>
		{/each}
		{#if f.dominios.length}
			<div class="t-t sep">Dominio propio</div>
			{#each f.dominios as d (d.nombre)}
				<div class="fila">
					<span class="f-txt">{d.nombre}</span><span class="f-s"
						>{d.verificado ? 'Verificado' : 'Sin verificar'}{d.principal
							? ' · principal'
							: ''}</span
					>
				</div>
			{/each}
		{/if}
	</section>

	{#if visibles.length}
		<section class="tarjeta">
			<div class="t-t">Acciones</div>
			<div class="acciones">
				{#each visibles as a (a.k)}
					<button type="button" class="borde" class:peligro={a.peligro} onclick={() => abrir(a)}
						><Icon icon={a.icon} size={15} />{a.t}</button
					>
				{/each}
			</div>
		</section>
	{/if}

	<section class="tarjeta">
		<div class="t-t">Actividad reciente</div>
		{#each f.actividad as a (a.id)}
			<div class="fila">
				<span class="f-txt"
					>{a.accion}<span class="f-s">{QUIEN[a.quien]} · {fechaCorta(a.cuando)}</span></span
				>
			</div>
		{:else}
			<div class="vacio">Sin actividad.</div>
		{/each}
	</section>

	<section class="tarjeta">
		<div class="t-t">Notas del equipo</div>
		{#each f.notas as n (n.id)}
			<div class="fila">
				<span class="f-txt"
					>{n.texto}<span class="f-s">{n.autor} · {fechaCorta(n.creadaEn)}</span></span
				>
			</div>
		{:else}
			<div class="vacio">Todavía no hay notas.</div>
		{/each}
		{#if puede('tiendas.notas')}
			<form
				method="POST"
				action="?/nota"
				class="nota"
				use:enhance={enviar(() => ui.toast({ k: 'ok', title: 'Nota guardada' }))}
			>
				<label class="campo"
					>Nota interna<textarea name="texto" maxlength="1000" placeholder="Sólo la ve el equipo"
					></textarea></label
				>
				<button type="submit" class="borde" disabled={enviando}
					><Icon icon={I.phNotePencilFill} size={15} />Guardar nota</button
				>
			</form>
		{/if}
	</section>
</div>

<Modal bind:open={modal} titulo={abierta?.t ?? 'Acción'} ancho={440}>
	{#if abierta}
		<form
			method="POST"
			action="?/accion"
			class="modal"
			use:enhance={enviar(() => {
				modal = false;
				ui.toast({ k: 'ok', title: `${abierta?.t}: listo`, desc: 'Quedó en la auditoría' });
			})}
		>
			<h2>{abierta.t}</h2>
			<input type="hidden" name="accion" value={abierta.k} />
			{#if abierta.k === 'extender_prueba'}
				<label class="campo"
					>Días más de prueba<input
						name="dias"
						type="number"
						min="1"
						max="30"
						value="7"
						required
					/></label
				>
			{:else if abierta.k === 'regalar_pack'}
				<label class="campo"
					>Packs de IA<input
						name="cantidad"
						type="number"
						min="1"
						max="5"
						value="1"
						required
					/></label
				>
			{:else if abierta.k === 'cambiar_plan'}
				<label class="campo"
					>Plan<select name="plan"
						>{#each ['gratis', 'emprende', 'pro', 'negocio'] as p (p)}<option
								value={p}
								selected={p === f.plan.actual}>{nombrePlan(p)}</option
							>{/each}</select
					></label
				>
			{:else if abierta.k === 'aplicar_credito'}
				<label class="campo"
					>Crédito en pesos<input
						name="pesos"
						type="number"
						min="1"
						max="1500000"
						required
						oninput={(ev) => {
							const m = ev.currentTarget.form?.elements.namedItem(
								'monto'
							) as HTMLInputElement | null;
							if (m) m.value = String(Math.round(Number(ev.currentTarget.value) * 100));
						}}
					/></label
				>
				<input type="hidden" name="monto" />
			{/if}
			{#if ['aplicar_credito', 'pausar', 'cerrar', 'reactivar'].includes(abierta.k)}
				<label class="campo"
					>Motivo<textarea
						name="motivo"
						maxlength="500"
						required={abierta.k !== 'reactivar'}
						placeholder="Queda en la auditoría"></textarea></label
				>
			{/if}
			{#if abierta.k === 'cerrar'}
				<p class="aviso">
					Cerrar es definitivo: la tienda deja de vender y la dueña pierde el acceso. Para
					confirmar, escribí <b>{t.slug}</b>.
				</p>
				<label class="campo"
					>Dirección de la tienda<input name="confirmacion" autocomplete="off" required /></label
				>
			{/if}
			{#if error}<p class="aviso" role="alert">{error}</p>{/if}
			<div class="pie">
				<button type="button" class="borde" onclick={() => (modal = false)}>Cancelar</button>
				<button type="submit" class="boton" class:rojo={abierta.peligro} disabled={enviando}
					>{abierta.t}</button
				>
			</div>
		</form>
	{/if}
</Modal>

<style>
	.volver {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		color: var(--mu);
		font-size: 13.5px;
		font-weight: 600;
		text-decoration: none;
	}
	.baja {
		margin: 0;
		padding: 10px 14px;
		border-radius: 12px;
		background: var(--chip);
		font-size: 14px;
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr));
		gap: 14px;
		align-items: start;
	}
	.dato {
		flex: 1;
	}
	.corto {
		max-width: 50%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.chicos {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 8px;
	}
	.chicos div {
		display: flex;
		flex-direction: column;
		padding: 10px 12px;
		border-radius: 14px;
		background: var(--sf2);
	}
	.chicos b {
		font-size: 20px;
	}
	.sep {
		margin-top: 14px;
	}
	.acciones {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	.peligro {
		color: #b83a2e;
	}
	.nota {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		align-items: flex-start;
	}
	.nota .campo {
		align-self: stretch;
	}
	.modal {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.modal h2 {
		margin: 0;
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.aviso {
		margin: 0;
		padding: 10px 12px;
		border-radius: 12px;
		background: #fcede5;
		color: #a8401c;
		font-size: 13.5px;
		line-height: 1.45;
	}
	.pie {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}
	.rojo {
		background: #b83a2e;
	}
</style>
