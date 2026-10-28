<script lang="ts">
	import { enhance, type SubmitFunction } from '$app/forms';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { ESTADOS_CONSULTA, fechaCorta, habilesHasta } from '#lib/staff/estado.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Una consulta: el mensaje, sus datos, el plazo legal (arrepentimiento) y las respuestas. */
	let { data } = $props();
	const ui = getUi();
	const c = $derived(data.consulta);
	const e = $derived(ESTADOS_CONSULTA[c.estado]);
	const puede = $derived(data.sesion.permisos.includes('consultas.responder'));
	const dias = $derived(
		c.tipo === 'arrepentimiento' && c.plazo ? habilesHasta(c.plazo, Date.now()) : null
	);
	let error = $state<string | null>(null);

	const enviar =
		(titulo: string): SubmitFunction =>
		() =>
		async ({ result, update }) => {
			error = null;
			if (result.type === 'failure') {
				if (result.data?.demo)
					ui.toast({
						k: 'info',
						title: 'En la demo no se aplica',
						desc: 'Con la cuenta real sale el mail y queda en la auditoría'
					});
				else error = (result.data?.error as string) ?? 'No se pudo.';
				return;
			}
			if (result.type === 'success') ui.toast({ k: 'ok', title: titulo });
			await update();
		};

	const DATOS = $derived(
		[
			['Email', c.email],
			['WhatsApp', c.whatsapp],
			['Plan', c.plan],
			['Web', c.web],
			['Facturación', c.facturacion],
			['Llegó desde', c.origen],
			['Documento', c.documento]
		].filter(([, v]) => v) as [string, string][]
	);
</script>

<Seo
	title="{c.asunto || c.nombre} · Consultas · Equipo"
	description="Consulta"
	path="/staff/consultas/{c.tipo}/{c.id}"
	noindex
/>

<a class="volver" href={resolve('/staff/(equipo)/consultas')}
	><Icon icon={I.phArrowLeftBold} size={13} />Consultas</a
>
<div class="cab">
	<h1>{c.tipo === 'arrepentimiento' ? `Arrepentimiento ${c.codigo}` : c.asunto || 'Consulta'}</h1>
	<span class="chip" style:background={e.bg} style:color={e.c}>{e.t}</span>
</div>
<p class="sub">{c.nombre} · {fechaCorta(c.creadaEn)}</p>

{#if dias !== null && c.estado !== 'resuelta'}
	<p class="plazo" class:urgente={dias <= 2}>
		<Icon icon={I.phWarningFill} size={16} />Ley 24.240: hay que devolver el dinero hasta el
		{fechaCorta(c.plazo ?? '')}
		({dias < 0
			? 'el plazo venció'
			: dias === 0
				? 'vence hoy'
				: `faltan ${dias} ${dias === 1 ? 'día hábil' : 'días hábiles'}`}).
	</p>
{/if}

<div class="grilla">
	<section class="tarjeta">
		<div class="t-t">Mensaje</div>
		<p class="mensaje">{c.mensaje}</p>
		{#each DATOS as [k, v] (k)}
			<div class="fila"><span class="f-s dato">{k}</span><span class="v">{v}</span></div>
		{/each}
	</section>

	<section class="tarjeta">
		<div class="t-t">Estado</div>
		{#if puede}
			<form
				method="POST"
				action="?/estado"
				class="form"
				use:enhance={enviar('Consulta actualizada')}
			>
				<label class="campo"
					>Estado<select name="estado"
						>{#each Object.entries(ESTADOS_CONSULTA) as [k, v] (k)}<option
								value={k}
								selected={k === c.estado}>{v.t}</option
							>{/each}</select
					></label
				>
				<label class="campo"
					>Asignada a<select name="asignadoA"
						><option value="">Sin asignar</option>{#each data.miembros as m (m.id)}<option
								value={m.id}
								selected={m.id === c.asignadoA}>{m.nombre}</option
							>{/each}</select
					></label
				>
				<button type="submit" class="borde">Guardar</button>
			</form>
		{:else}
			<p class="sub">Tu rol puede leer las consultas, no cambiarlas.</p>
		{/if}
	</section>

	<section class="tarjeta ancha">
		<div class="t-t">Respuestas</div>
		{#each c.respuestas as r (r.id)}
			<div class="respuesta">
				<b>{r.asunto}</b>
				<p>{r.texto}</p>
				<span class="f-s">{r.autor} · {fechaCorta(r.creadaEn)}</span>
			</div>
		{:else}
			<div class="vacio">Todavía no le respondimos.</div>
		{/each}
		{#if puede}
			<form
				method="POST"
				action="?/responder"
				class="form"
				use:enhance={enviar('Respuesta enviada por mail')}
			>
				<label class="campo"
					>Asunto<input
						name="asunto"
						maxlength="140"
						value="Re: {c.asunto || 'tu consulta en Mostrador'}"
					/></label
				>
				<label class="campo"
					>Respuesta para {c.email}<textarea
						name="texto"
						maxlength="5000"
						placeholder="Hola {c.nombre.split(' ')[0]}, …"></textarea></label
				>
				{#if error}<p class="aviso" role="alert">{error}</p>{/if}
				<button type="submit" class="boton"
					><Icon icon={I.phPaperPlaneTiltFill} size={15} />Responder por mail</button
				>
			</form>
		{/if}
	</section>
</div>

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
	.plazo {
		margin: 0;
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
		padding: 12px 14px;
		border-radius: 14px;
		background: var(--chip);
		font-size: 14px;
	}
	.plazo.urgente {
		background: rgba(184, 58, 46, 0.1);
		color: #b83a2e;
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr));
		gap: 14px;
		align-items: start;
	}
	.ancha {
		grid-column: 1 / -1;
	}
	.mensaje {
		margin: 0 0 8px;
		font-size: 14.5px;
		line-height: 1.55;
		white-space: pre-wrap;
	}
	.dato {
		flex: 1;
	}
	.v {
		overflow-wrap: anywhere;
		text-align: right;
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: 10px;
		align-items: flex-start;
		margin-top: 10px;
	}
	.form .campo {
		align-self: stretch;
	}
	.respuesta {
		padding: 12px 0;
		border-bottom: 1px solid var(--ln);
	}
	.respuesta p {
		margin: 4px 0;
		font-size: 14px;
		line-height: 1.5;
		white-space: pre-wrap;
	}
	.aviso {
		margin: 0;
		padding: 10px 12px;
		border-radius: 12px;
		background: #fcede5;
		color: #a8401c;
		font-size: 13.5px;
	}
</style>
