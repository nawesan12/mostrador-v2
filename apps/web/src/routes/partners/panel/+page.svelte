<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Vacio from '#lib/components/Vacio.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { NIVELES, cuandoSePaga, nivel, nombreMes, resumen } from '#lib/partners/programa.ts';
	import { ESTADO_PAGO } from '#lib/partners/vista.ts';

	/** Inicio del panel de partners: el mes, el próximo pago, el nivel y lo que hay que hacer. */
	let { data } = $props();
	const p = $derived(data.panel!);
	const r = $derived(resumen(p));
	const actual = $derived(nivel(p.partner.nivelEfectivo));

	const IR = {
		tiendas: '/partners/panel/tiendas',
		comisiones: '/partners/panel/comisiones',
		perfil: '/partners/panel/perfil'
	} as const;
	const ICONO = {
		tiendas: I.phStorefrontFill,
		comisiones: I.phHandCoinsFill,
		perfil: I.phIdentificationCardFill
	} as const;
</script>

<div class="kpis" use:anim={{ t: 'rise' }}>
	<div class="kpi grande">
		<div class="k-t">
			<Icon icon={I.phHandCoinsFill} size={15} style="color:#b0441d" />Comisión del mes
		</div>
		<div class="k-v">{pesos(r.estimado)}</div>
		<div class="k-d">
			Estimada · {r.generan}
			{r.generan === 1 ? 'tienda genera' : 'tiendas generan'} comisión
		</div>
	</div>
	<div class="kpi">
		<div class="k-t">
			<Icon icon={I.phCalendarCheckFill} size={15} style="color:#237550" />Próximo pago
		</div>
		<div class="k-v">{pesos(r.proximo.monto)}</div>
		<div class="k-d">{ESTADO_PAGO[r.proximo.estado].t} · {r.proximo.cuando}</div>
	</div>
	<div class="kpi">
		<div class="k-t">
			<Icon icon={I.phStorefrontFill} size={15} style="color:#237550" />Tiendas activas
		</div>
		<div class="k-v">{r.activas}</div>
		<div class="k-d">Con plan pago</div>
	</div>
	<div class="kpi">
		<div class="k-t">
			<Icon icon={I.phHourglassMediumFill} size={15} style="color:#C98A12" />En prueba
		</div>
		<div class="k-v">{r.prueba}</div>
		<div class="k-d">Sus 14 días gratis</div>
	</div>
	<div class="kpi">
		<div class="k-t"><Icon icon={I.phTrendDownBold} size={15} style="color:#B83A2E" />Perdidas</div>
		<div class="k-v">{r.perdidas}</div>
		<div class="k-d">Sin plan pago o de baja</div>
	</div>
</div>

<div class="dos" use:anim={{ t: 'rise', delay: 40 }}>
	<div class="tarjeta">
		<div class="t-cab">
			<span>Tu nivel</span>
			<span class="chip-f nivel" class:oro={actual.id === 'oro'}
				>{actual.nombre} · {actual.pct}&nbsp;%</span
			>
		</div>
		{#if r.progreso}
			<div class="progreso">
				<span
					class="medidor"
					role="progressbar"
					aria-label="Tiendas activas para {r.progreso.siguiente.nombre}"
					aria-valuemin={0}
					aria-valuemax={r.progreso.siguiente.tiendas}
					aria-valuenow={r.activas}><span style:width="{r.progreso.pct}%"></span></span
				>
				<div class="p-n">
					<b>{r.activas} de {r.progreso.siguiente.tiendas}</b> tiendas activas
				</div>
			</div>
			<div class="t-d">
				Te {r.progreso.faltan === 1
					? 'falta 1 tienda activa'
					: `faltan ${r.progreso.faltan} tiendas activas`}
				para {r.progreso.siguiente.nombre}: {r.progreso.siguiente.pct}&nbsp;% sobre todas
				tus&nbsp;tiendas.
			</div>
		{:else}
			<div class="t-d">
				Llegaste al nivel más alto: 30&nbsp;% de comisión y destacado en el&nbsp;directorio.
			</div>
		{/if}
		<div class="niveles">
			{#each NIVELES as n (n.id)}
				<div class="nv" class:on={n.id === actual.id}>
					<b>{n.nombre}</b><span
						>{n.pct}&nbsp;% · {n.tiendas ? `desde ${n.tiendas} tiendas` : 'sin requisitos'}</span
					>
				</div>
			{/each}
		</div>
	</div>

	<div class="tarjeta">
		<div class="t-cab"><span>Para hacer</span></div>
		{#if r.tareas.length}
			<div class="filas">
				{#each r.tareas as t (t.id)}
					<a class="opcion tarea" href={resolve(IR[t.ir])}>
						<span class="o-ico"><Icon icon={ICONO[t.ir]} size={18} /></span>
						<span class="o-txt"><span class="o-t">{t.texto}</span></span>
						<Icon icon={I.phCaretRightBold} size={14} style="color:var(--mu)" />
					</a>
				{/each}
			</div>
		{:else}
			<div class="vacio-in">
				<Vacio
					hornero="exito"
					titulo="Todo al día"
					texto="Cuando una tienda tuya necesite algo, te avisamos acá."
				/>
			</div>
		{/if}
	</div>
</div>

{#if p.pagos.length}
	<div class="tarjeta" use:anim={{ t: 'rise', delay: 80 }}>
		<div class="t-cab">
			<span>Últimos pagos</span>
			<a class="mas" href={resolve('/partners/panel/comisiones')}
				>Ver todo <Icon icon={I.phArrowRightBold} size={13} /></a
			>
		</div>
		<div class="filas">
			{#each p.pagos.slice(0, 3) as g (g.mes)}
				{@const c = ESTADO_PAGO[g.estado]}
				<div class="f-in">
					<span class="ini" style:background={c.bg} style:color={c.fg}
						><Icon icon={I.phReceiptFill} size={18} /></span
					>
					<span class="quien">
						<span class="q-t">{nombreMes(g.mes)}</span>
						<span class="q-s"
							>{g.detalle.length}
							{g.detalle.length === 1 ? 'tienda' : 'tiendas'} · {cuandoSePaga(g, p.hoy)}</span
						>
					</span>
					<span class="chip-f solo-ancho" style:background={c.bg} style:color={c.fg}>{c.t}</span>
					<span class="valor"><span class="v-v">{pesos(g.monto)}</span></span>
				</div>
			{/each}
		</div>
	</div>
{/if}

<style>
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr));
		gap: 12px;
		align-items: start;
	}
	.nivel {
		background: #efeae4;
		color: #5c4536;
	}
	.nivel.oro {
		background: #fcede5;
		color: #b0441d;
	}
	.progreso {
		margin-top: 14px;
	}
	.p-n {
		margin-top: 8px;
		font-size: 13.5px;
	}
	.t-d {
		margin-top: 4px;
	}
	.niveles {
		margin-top: 14px;
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 6px;
	}
	.nv {
		padding: 10px 12px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		display: flex;
		flex-direction: column;
		gap: 2px;
		font-size: 13px;
	}
	.nv span {
		color: var(--mu);
		font-size: 12px;
	}
	.nv.on {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.06);
	}
	.tarea {
		color: var(--tx);
		text-decoration: none;
	}
	.vacio-in {
		margin-top: 8px;
	}
	@media (max-width: 759px) {
		.grande {
			grid-column: 1 / -1;
		}
		.niveles {
			grid-template-columns: minmax(0, 1fr);
		}
		.solo-ancho {
			display: none;
		}
	}
</style>
