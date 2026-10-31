<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import {
		activarPush,
		desactivarPush,
		estadoPush,
		probarPush,
		type EstadoPush
	} from '#lib/panel/app-celular.svelte.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Notificaciones › «Push en este celular» (docs/decisiones.md §42): activa el push web en este
	 * aparato y manda un aviso de prueba. En el iPhone sólo anda con el panel instalado en la pantalla
	 * de inicio (iOS 16.4 o más nuevo), y se dice así.
	 */
	const ui = getUi();
	let estado = $state<EstadoPush>('cargando');
	let clave = '';
	let ocupado = $state(false);

	onMount(async () => {
		const nada = { clave: '', suscripta: false };
		const srv: typeof nada = await fetch('/panel/notificaciones/push')
			.then((r) => (r.ok ? (r.json() as Promise<typeof nada>) : nada))
			.catch(() => nada);
		clave = srv.clave;
		estado = await estadoPush(srv).catch(() => 'sin-soporte' as const);
	});

	const TEXTO: Record<EstadoPush, string> = {
		cargando: 'Revisando este celular…',
		'no-disponible': 'Todavía no está disponible. Mientras, te avisamos por email.',
		'instalar-primero':
			'En el iPhone, los avisos llegan sólo con Tu\u00a0Local\u00a0Web instalada: en Safari, Compartir › Agregar a inicio. Después abrila desde el ícono y activalos acá.',
		'ios-viejo':
			'Para recibir avisos, el iPhone tiene que tener iOS 16.4 o más nuevo. Actualizalo en Ajustes › General.',
		'sin-soporte': 'Este navegador no recibe avisos. Probá con Chrome o con el panel instalado.',
		bloqueado:
			'Bloqueaste los avisos de Tu\u00a0Local\u00a0Web. Habilitalos en los ajustes del celular, en Notificaciones, y volvé acá.',
		apagado: 'Te avisamos cada venta al instante en este celular, aunque el panel esté cerrado.',
		prendido: 'Activado. Las ventas nuevas te llegan a este celular.'
	};

	async function activar() {
		ocupado = true;
		const r = await activarPush(clave);
		ocupado = false;
		if (r === 'prendido') {
			estado = 'prendido';
			ui.toast({ k: 'ok', title: 'Avisos activados', desc: 'Probá con un aviso de prueba' });
		} else if (r === 'bloqueado') {
			estado = 'bloqueado';
		} else {
			ui.toast({
				k: 'err',
				title: 'No pudimos activar los avisos',
				desc: 'Probá de nuevo en un rato'
			});
		}
	}

	async function desactivar() {
		ocupado = true;
		const ok = await desactivarPush().catch(() => false);
		ocupado = false;
		if (ok) {
			estado = 'apagado';
			ui.toast({ k: 'ok', title: 'Avisos desactivados en este celular' });
		} else ui.toast({ k: 'err', title: 'No pudimos desactivarlos', desc: 'Probá de nuevo' });
	}

	async function probar() {
		ocupado = true;
		const ok = await probarPush().catch(() => false);
		ocupado = false;
		ui.toast(
			ok
				? { k: 'ok', title: 'Aviso de prueba en camino', desc: 'Te llega en unos segundos' }
				: { k: 'err', title: 'No pudimos mandarlo', desc: 'Activá los avisos y probá de nuevo' }
		);
	}
</script>

<div class="push" aria-busy={estado === 'cargando'}>
	<span class="a-ico"><Icon icon={I.phBellRingingFill} size={20} /></span>
	<span class="a-txt"
		><b>Push en este celular</b><span class="d" aria-live="polite"
			>{TEXTO[estado].replace(/ (\S+)$/, '\u00a0$1')}</span
		></span
	>
	{#if estado === 'apagado'}
		<button type="button" class="btn btn-primary btn-sm" disabled={ocupado} onclick={activar}
			>Activar</button
		>
	{/if}
	{#if estado === 'prendido'}
		<div class="acciones">
			<button type="button" class="btn btn-dark btn-sm" disabled={ocupado} onclick={probar}
				><Icon icon={I.phBellRingingBold} size={15} />Mandame un aviso de prueba</button
			>
			<button type="button" class="btn btn-secondary btn-sm" disabled={ocupado} onclick={desactivar}
				>Desactivar</button
			>
		</div>
	{/if}
</div>

<style>
	.push {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: 12px;
		min-height: 64px;
		padding: 12px 18px;
		border-top: 1px solid var(--ln);
	}
	.a-ico {
		flex: none;
		width: 38px;
		height: 38px;
		border-radius: 12px;
		background: var(--sf3);
		color: #6a4fa0;
		display: grid;
		place-items: center;
	}
	.a-txt {
		flex: 1;
		min-width: min(100%, 220px);
		font-size: 14px;
	}
	.a-txt b {
		font-weight: 600;
	}
	.d {
		display: block;
		font-size: 12px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.acciones {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	/* Celular: los botones abajo, alineados con el texto. */
	@media (max-width: 760px) {
		.acciones,
		.push > .btn {
			margin-left: 50px;
		}
	}
</style>
