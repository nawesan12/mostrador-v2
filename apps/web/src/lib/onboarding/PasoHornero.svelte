<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { TONOS, type Tono } from './datos.ts';
	import Etiqueta from './Etiqueta.svelte';
	import { getOnboarding } from './estado.svelte.ts';

	/** Paso 7, Hornero: el tono, dónde atiende y la vista previa de una respuesta. */
	const o = getOnboarding();
	const tonos = Object.entries(TONOS) as [Tono, (typeof TONOS)[Tono]][];

	// Se anima cada vez que cambia (como en el handoff).
	let burbuja = $state<HTMLElement>();
	$effect(() => {
		void o.e.tono;
		burbuja?.animate(
			[
				{ opacity: 0, transform: 'translateY(8px) scale(.97)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 380, easing: 'cubic-bezier(.2,.7,.1,1)' }
		);
	});
</script>

<div class="dos">
	<div class="izq">
		<Etiqueta
			rotulo="Tu asistente"
			titulo="Hornero ya conoce tu negocio. ¿Cómo querés que hable?"
		/>
		<div class="sub">Tono</div>
		<div class="tonos" role="group" aria-label="Tono">
			{#each tonos as [id, [n]] (id)}
				<button
					type="button"
					aria-pressed={o.e.tono === id}
					onclick={() => ((o.e.tono = id), o.guardar())}>{n}</button
				>
			{/each}
		</div>
		<div class="sub">Dónde atiende</div>
		<div class="canales">
			<span
				><Icon icon={I.logosWhatsappIcon} size={16} />WhatsApp<Icon
					icon={I.phCheckBold}
					size={13}
					style="color:#237550"
				/></span
			>
			<span
				><Icon icon={I.logosInstagramIcon} size={15} />Instagram<Icon
					icon={I.phCheckBold}
					size={13}
					style="color:#237550"
				/></span
			>
			<span
				><span class="ml">ML</span>Preguntas de ML<Icon
					icon={I.phCheckBold}
					size={13}
					style="color:#237550"
				/></span
			>
			<span class="off"><Icon icon={I.logosTelegram} size={15} />Telegram</span>
		</div>
		<div class="regla">
			<Icon icon={I.logosGoogleMaps} size={20} />
			<span
				><b>Horario del local:</b> lun a sáb de 8 a 20 h<br /><span class="r-d"
					>Fuera de ese horario, Hornero atiende y vos ves todo a la mañana.</span
				></span
			>
		</div>
		<div class="regla">
			<Icon icon={I.phSealPercentFill} size={20} style="color:#b0441d" />
			<span
				><b>Descuentos:</b> puede ofrecer hasta 10% sin preguntarte<br /><span class="r-d"
					>Más que eso, te pide aprobación por WhatsApp.</span
				></span
			>
		</div>
	</div>
	<div class="vista">
		<div class="v-cab">
			<img
				src="https://images.unsplash.com/photo-1531058240690-006c446962d8?auto=format&fit=crop&w=96&h=96&q=70&crop=faces"
				alt=""
				width="34"
				height="34"
			/>
			<span>Vista previa</span><Icon icon={I.logosWhatsappIcon} size={16} />
		</div>
		<div class="chat">
			<div class="cliente">Hola! ¿Tienen el termo de 1 L en negro?</div>
			<div class="resp" bind:this={burbuja}>
				<span class="firma"><Icon icon={I.phSparkleFill} size={11} />Hornero</span>
				<span class="burbuja">{TONOS[o.e.tono][1]}</span>
			</div>
		</div>
	</div>
</div>

<style>
	.dos {
		display: flex;
		gap: 28px;
		flex-wrap: wrap;
	}
	.izq {
		flex: 1 1 420px;
		min-width: 0;
	}
	.sub {
		margin-top: 22px;
		font-size: 14px;
		font-weight: 600;
	}
	.sub + .canales {
		margin-top: 8px;
	}
	.tonos {
		margin-top: 8px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.tonos button {
		padding: 11px 16px;
		border-radius: 999px;
		border: 1.5px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		color: #2a140c;
		font-size: 14px;
		font-weight: 600;
	}
	.tonos button[aria-pressed='true'] {
		background: #2a140c;
		border-color: #2a140c;
		color: #fff;
	}
	.canales {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.canales > span {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 14px;
		border-radius: 999px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.08);
		font-size: 14px;
		font-weight: 500;
	}
	.canales .off {
		color: #7a6252;
	}
	.ml {
		font-size: 8px;
		font-weight: 800;
		padding: 2px 4px;
		border-radius: 4px;
		background: #ffe600;
		color: #2d3277;
	}
	.regla {
		margin-top: 12px;
		padding: 14px 16px;
		border-radius: 16px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		display: flex;
		gap: 12px;
		align-items: center;
		font-size: 14px;
		line-height: 1.4;
	}
	.sub + .canales + .regla {
		margin-top: 20px;
	}
	.r-d {
		color: #7a6252;
		font-size: 13px;
	}
	.vista {
		flex: 1 1 320px;
		align-self: flex-start;
		border-radius: 24px;
		background: #fff;
		overflow: hidden;
		box-shadow:
			0 2px 4px rgba(58, 29, 18, 0.06),
			0 30px 60px -30px rgba(58, 29, 18, 0.5);
	}
	.v-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 14px;
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
		font-size: 14px;
		font-weight: 600;
	}
	.v-cab img {
		border-radius: 50%;
		object-fit: cover;
	}
	.v-cab span {
		flex: 1;
	}
	.chat {
		padding: 16px 14px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		background: #f4ece2 url('/textura/arch-pattern-dark.svg');
		min-height: 260px;
		justify-content: flex-end;
	}
	.cliente {
		align-self: flex-start;
		max-width: 82%;
		font-size: 14px;
		line-height: 1.45;
		background: #fff;
		padding: 9px 13px;
		border-radius: 16px 16px 16px 4px;
	}
	.resp {
		align-self: flex-end;
		max-width: 86%;
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 3px;
	}
	.firma {
		font-size: 11px;
		font-weight: 600;
		color: #b0441d;
		display: flex;
		align-items: center;
		gap: 4px;
	}
	.burbuja {
		font-size: 14px;
		line-height: 1.45;
		background: #fce3d6;
		padding: 9px 13px;
		border-radius: 16px 16px 4px 16px;
	}
</style>
