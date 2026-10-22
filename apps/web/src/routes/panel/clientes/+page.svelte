<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import FichaCliente from '#lib/panel/listas/FichaCliente.svelte';
	import MotorLista from '#lib/panel/listas/MotorLista.svelte';
	import { configClientes, escribir, type Cliente } from '#lib/panel/listas/clientes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Clientes (PanelListas, pantalla Clientes) con su ficha, el alta y el etiquetado. */
	let { data } = $props();
	const ui = getUi();

	let alta = $state(false);
	let etiquetar = $state<string[]>([]);
	let etiquetando = $state(false);
	let ficha = $state(false);
	let elegido = $state<Cliente | null>(null);
	const config = configClientes(
		() => (alta = true),
		(ids) => {
			etiquetar = ids;
			etiquetando = true;
		}
	);

	async function accion(ids: string[], accion: string, valor?: unknown) {
		const r = await fetch('/panel/listas/clientes', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ ids, accion, valor })
		});
		await invalidateAll();
		return r.ok ? ((await r.json()) as { hechos: number }).hechos : 0;
	}

	let nombre = $state('');
	let telefono = $state('');
	let email = $state('');
	let ciudad = $state('');
	let errorAlta = $state<string | null>(null);
	async function crear(e: SubmitEvent) {
		e.preventDefault();
		if (nombre.trim().length < 2) {
			errorAlta = 'Poné el nombre del cliente.';
			return;
		}
		if (telefono && telefono.replace(/\D/g, '').length < 8) {
			errorAlta = 'El WhatsApp tiene que tener la característica: 11 5555-1234.';
			return;
		}
		const n = await accion([], 'crear', { nombre, telefono, email, ciudad });
		if (!n) {
			errorAlta = 'No pudimos guardarlo. Revisá los datos.';
			return;
		}
		ui.toast({ k: 'ok', title: `${nombre.trim()} ya es cliente`, desc: 'Lo encontrás en Nuevos' });
		nombre = telefono = email = ciudad = '';
		errorAlta = null;
		alta = false;
	}

	let etiqueta = $state('');
	async function aplicarEtiqueta(e: SubmitEvent) {
		e.preventDefault();
		if (!etiquetar.length || !etiqueta.trim()) return;
		const n = await accion(etiquetar, 'etiquetar', etiqueta);
		ui.toast({
			k: 'ok',
			title: `Etiqueta «${etiqueta.trim()}»`,
			desc: `${n} ${n === 1 ? 'cliente' : 'clientes'}`
		});
		etiqueta = '';
		etiquetando = false;
	}
</script>

<Seo title="Clientes · Panel" description="Clientes de tu tienda" path="/panel/clientes" noindex />

<MotorLista
	{config}
	lista={data.lista}
	abrir={(c: Cliente) => {
		elegido = c;
		ficha = true;
	}}
/>

<FichaCliente
	bind:open={ficha}
	cliente={elegido}
	escribir={(c) => {
		if (!escribir(c))
			ui.toast({ k: 'warn', title: 'Sin WhatsApp cargado', desc: 'Sumalo desde su ficha' });
	}}
	cupon={async (c) => {
		await accion([c.id], 'cupon');
		elegido = { ...c, cupon: true };
		ui.toast({
			k: 'ok',
			title: `Cupón del 10% para ${c.nombre.split(' ')[0]}`,
			desc: 'Enviado por WhatsApp · vence en 7 días'
		});
	}}
/>

<Modal bind:open={alta} titulo="Nuevo cliente" ancho={460}>
	<form onsubmit={crear}>
		<div class="m-ico"><Icon icon={I.phUserPlusFill} size={24} /></div>
		<div class="m-t">Nuevo cliente</div>
		<div class="m-d">Para cargar a alguien que compró en el local o por teléfono.</div>
		<div class="m-campos">
			<label class="field"
				>Nombre y apellido<input
					class="input"
					bind:value={nombre}
					maxlength="80"
					autocomplete="off"
				/></label
			>
			<label class="field"
				>WhatsApp<input
					class="input"
					bind:value={telefono}
					inputmode="tel"
					placeholder="11 5555-1234"
					maxlength="30"
				/></label
			>
			<label class="field"
				>Email<input class="input" bind:value={email} type="email" maxlength="120" /></label
			>
			<label class="field">Ciudad<input class="input" bind:value={ciudad} maxlength="60" /></label>
		</div>
		{#if errorAlta}<div class="m-error" role="alert">{errorAlta}</div>{/if}
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (alta = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary">Guardar cliente</button>
		</div>
	</form>
</Modal>

<Modal bind:open={etiquetando} titulo="Etiquetar clientes" ancho={420}>
	<form onsubmit={aplicarEtiqueta}>
		<div class="m-ico violeta"><Icon icon={I.phTagFill} size={24} /></div>
		<div class="m-t">
			Etiquetar {etiquetar.length}
			{etiquetar.length === 1 ? 'cliente' : 'clientes'}
		</div>
		<div class="m-d">
			Sirve para buscarlos y para mandarles campañas: «Mayoristas zona sur», «Feria de mayo».
		</div>
		<div class="m-campos">
			<label class="field"
				>Etiqueta<input
					class="input"
					bind:value={etiqueta}
					maxlength="30"
					autocomplete="off"
				/></label
			>
		</div>
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (etiquetando = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary" disabled={!etiqueta.trim()}>Etiquetar</button>
		</div>
	</form>
</Modal>

<style>
	.m-ico {
		width: 48px;
		height: 48px;
		display: grid;
		place-items: center;
		border-radius: 15px;
		background: rgba(210, 88, 47, 0.12);
		color: #c24e25;
	}
	.m-ico.violeta {
		background: rgba(106, 79, 160, 0.12);
		color: #6a4fa0;
	}
	.m-t {
		margin-top: 16px;
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14.5px;
		line-height: 1.5;
		color: var(--mu);
	}
	.m-campos {
		margin-top: 16px;
		display: grid;
		gap: 12px;
	}
	.m-error {
		margin-top: 10px;
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	.m-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
</style>
