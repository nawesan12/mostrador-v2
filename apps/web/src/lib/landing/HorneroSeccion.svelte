<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';
	import { CHAT } from './datos.ts';
	import { getLanding } from './estado.svelte.ts';
	import { cara, unsplash } from './formato.ts';

	/** Sección "Conocé a Hornero" del handoff: chat demo animado, capacidades y control. */
	const l = getLanding();
	const ARCO =
		'M3 76a3 3 0 0 1-3-3V5a5 5 0 0 1 5-5h90a5 5 0 0 1 5 5v68a3 3 0 0 1-3 3H82a3 3 0 0 1-3-3V21H21v52a3 3 0 0 1-3 3Z';

	const visibles = $derived(CHAT.slice(0, Math.min(l.chat, CHAT.length)));
	const siguiente = $derived(CHAT[l.chat]);
	const escribiendo = $derived(!!siguiente && ['h', 'prod', 'pay'].includes(siguiente.k));

	let ctl = $state({ aprob: true, horario: true, derivar: true });
	const CONTROLES = [
		['aprob', 'Pedir mi aprobación para descuentos'],
		['horario', 'Atender fuera del horario del local'],
		['derivar', 'Pasarme el chat si el cliente se enoja']
	] as const;

	const CAPACIDADES = [
		[
			I.phChatsCircleFill,
			'Atiende y cierra ventas',
			'Responde dudas, recomienda productos, calcula el envío y manda el link de pago.'
		],
		[
			I.phShoppingCartSimpleFill,
			'Recupera lo que se escapa',
			'Carritos abandonados, pagos rechazados y clientes que no vuelven hace meses.'
		],
		[
			I.phCameraFill,
			'Carga productos con una foto',
			'Saca nombre, descripción, categoría y precio sugerido. Vos solo confirmás.'
		],
		[
			I.phLightbulbFilamentFill,
			'Te dice qué hacer hoy',
			'Qué reponer, qué promocionar y qué precio revisar, con tus números reales.'
		]
	] as const;

	const METRICAS = [
		{
			ico: I.phLightningFill,
			bg: 'rgba(47,143,99,.13)',
			col: '#2F8F63',
			b: '8 segundos',
			s: 'tiempo de respuesta',
			pos: 'left:0;top:60px',
			dur: 3600,
			delay: 300,
			px: '-0.22'
		},
		{
			ico: I.phShoppingCartSimpleFill,
			bg: 'rgba(210,88,47,.12)',
			col: '#C24E25',
			b: '$ 384.600',
			s: 'en carritos recuperados',
			pos: 'right:0;top:120px',
			dur: 4400,
			delay: 900,
			px: '-0.15'
		},
		{
			ico: I.phCameraFill,
			bg: 'rgba(227,169,59,.2)',
			col: '#9A6A10',
			b: '24 productos',
			s: 'cargados desde fotos',
			pos: 'left:30px;bottom:50px',
			dur: 3900,
			delay: 500,
			px: '-0.3'
		}
	];

	/** Cada mensaje nuevo entra como en el handoff (translateY 12px, scale .97, 420 ms). */
	function entra(el: HTMLElement) {
		el.animate(
			[
				{ opacity: 0, transform: 'translateY(12px) scale(.97)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 420, easing: 'cubic-bezier(.2,.7,.1,1)' }
		);
	}
</script>

<section id="hornero">
	<div class="arcos"></div>
	<div class="in">
		<div class="cab" use:anim={'reveal'}>
			<div class="sello">
				<Icon icon={I.phSparkleFill} size={15} />Asistente con IA · exclusivo de Mostrador
			</div>
			<h2>Conocé a Hornero. Atiende, vende y te avisa — las 24 horas.</h2>
			<p>
				Arma el negocio como el hornero arma su nido: sin descanso. Contesta por WhatsApp, Instagram
				y Telegram con tu catálogo, tus precios y tu stock reales — y te consulta antes de hacer
				algo importante.
			</p>
		</div>

		<div class="dos">
			<div class="escena" use:anim={'reveal'}>
				<svg class="arco" data-px="0.1" viewBox="0 0 100 76" aria-hidden="true"
					><path fill="#fff" d={ARCO}></path></svg
				>
				<div class="ave" data-px="-0.08" use:anim={{ t: 'float', dur: 4200 }}>
					<div class="ave-in"><img src="/hornero/saludo.png" alt="Hornero" /></div>
				</div>
				{#each METRICAS as m, i (m.b)}
					<div
						class="metrica m{i}"
						style={m.pos}
						data-px={m.px}
						use:anim={{ t: 'float', dur: m.dur, delay: m.delay }}
					>
						<span class="m-ico" style:background={m.bg} style:color={m.col}
							><Icon icon={m.ico} size={17} /></span
						>
						<span class="m-txt"><b>{m.b}</b><br /><span>{m.s}</span></span>
					</div>
				{/each}
				<div
					class="metrica canales"
					data-px="-0.12"
					use:anim={{ t: 'float', dur: 4700, delay: 200 }}
				>
					<Icon icon={I.logosWhatsappIcon} size={20} /><Icon
						icon={I.logosInstagramIcon}
						size={18}
					/><Icon icon={I.logosTelegram} size={18} /><span>Atiende en todos tus canales</span>
				</div>
			</div>

			<div class="chat" use:anim={{ t: 'reveal', delay: 120 }}>
				<div class="chat-cab">
					<img src={cara('1531058240690-006c446962d8')} alt="" />
					<div class="cab-txt">
						<div class="nom">Lucía Fernández <Icon icon={I.logosWhatsappIcon} size={14} /></div>
						<div class="sub">Cliente nueva · llegó por Instagram</div>
					</div>
					<span class="atiende"
						><span class="av"><img src="/hornero/avatar.png" alt="Hornero" /></span>Hornero
						atendiendo</span
					>
				</div>
				<div class="mensajes" aria-live="polite">
					{#each visibles as m, i (i)}
						{#if m.k === 'c'}
							<div class="m-c" use:entra>{m.t}</div>
						{:else if m.k === 'h'}
							<div class="m-h" use:entra>
								<span class="firma"><Icon icon={I.phSparkleFill} size={11} />Hornero</span><span
									class="burbuja">{m.t}</span
								>
							</div>
						{:else if m.k === 'prod'}
							<div class="m-tarjeta" use:entra>
								<div class="prod">
									<img src={unsplash('1573612664822-d7d347da7b80', 500, '&h=260')} alt="" />
									<div class="prod-pie">
										<div>
											<div class="prod-n">Termo 1 L · Negro</div>
											<div class="prod-s">Quedan 3</div>
										</div>
										<span class="prod-p">$ 42.900</span>
									</div>
								</div>
							</div>
						{:else if m.k === 'pay'}
							<div class="m-tarjeta" use:entra>
								<div class="pago">
									<span class="mp"
										><Icon icon={I.siMercadopago} size={20} style="color:#00B1EA" /></span
									>
									<div class="pago-txt">
										<div class="a">Pagar $ 44.600</div>
										<div class="b">Hasta 6 cuotas sin interés</div>
									</div>
									<Icon icon={I.phArrowUpRightBold} size={15} style="color:#b0441d" />
								</div>
							</div>
						{:else}
							<div class="m-sys" use:entra><Icon icon={I.phCheckCircleFill} size={15} />{m.t}</div>
						{/if}
					{/each}
					{#if escribiendo}
						<div class="puntos" aria-label="Hornero está escribiendo">
							<span use:anim={'dots'}></span><span use:anim={{ t: 'dots', delay: 180 }}></span><span
								use:anim={{ t: 'dots', delay: 360 }}
							></span>
						</div>
					{/if}
				</div>
				<div class="chat-pie">
					<span>Hornero contesta solo. Vos podés intervenir cuando quieras.</span>
					<span class="tomar">Tomar la conversación</span>
				</div>
			</div>
		</div>

		<div class="capacidades">
			{#each CAPACIDADES as [ico, t, d], i (t)}
				<div class="cap" use:anim={{ t: 'reveal', delay: i * 60 }}>
					<span class="cap-ico"><Icon icon={ico} size={22} /></span>
					<div class="cap-t">{t}</div>
					<div class="cap-d">{d}</div>
				</div>
			{/each}
		</div>

		<div class="control" use:anim={'reveal'}>
			<div class="ctl-txt">
				<div class="ctl-t">Vos tenés el control</div>
				<div class="ctl-d">
					Hornero nunca inventa precios ni promete lo que no tenés. Elegí hasta dónde llega.
				</div>
			</div>
			<div class="ctl-grilla">
				{#each CONTROLES as [k, t] (k)}
					<button
						type="button"
						class="ctl"
						role="switch"
						aria-checked={ctl[k]}
						onclick={() => (ctl[k] = !ctl[k])}
					>
						<span class="ctl-n">{t}</span>
						<span class="tg" class:on={ctl[k]}><span></span></span>
					</button>
				{/each}
			</div>
		</div>
	</div>
</section>

<style>
	section {
		position: relative;
		overflow: hidden;
		background: linear-gradient(180deg, #fbf7f2 0%, #fceadf 45%, #f7d9c6 100%);
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
	}
	.arcos {
		position: absolute;
		inset: 0;
		background-image: url('/textura/arch-pattern-accent.svg');
		mask-image: radial-gradient(ellipse 60% 70% at 25% 55%, #000, transparent 75%);
	}
	.in {
		position: relative;
		max-width: 1240px;
		margin: 0 auto;
		padding: 120px var(--lado) 110px;
	}
	.cab {
		text-align: center;
		max-width: 860px;
		margin: 0 auto;
	}
	.sello {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #fff;
		background: #d2582f;
		padding: 7px 14px;
		border-radius: 999px;
		box-shadow: 0 10px 24px -12px rgba(210, 88, 47, 0.8);
	}
	h2 {
		text-wrap: balance;
		margin: 22px 0 0;
		font-size: clamp(40px, 5.6vw, 76px);
		line-height: 0.98;
		letter-spacing: -0.045em;
		font-weight: 600;
	}
	.cab p {
		text-wrap: pretty;
		margin: 22px auto 0;
		font-size: 19px;
		line-height: 1.55;
		color: #5c4536;
		max-width: 640px;
	}
	.dos {
		margin-top: 64px;
		display: flex;
		gap: 48px;
		flex-wrap: wrap;
		align-items: center;
	}
	.escena {
		flex: 1 1 420px;
		position: relative;
		min-height: 520px;
		display: grid;
		place-items: center;
	}
	.arco {
		position: absolute;
		width: 92%;
		max-width: 520px;
		opacity: 0.9;
	}
	.ave {
		position: relative;
		width: 360px;
		height: 360px;
		border-radius: 50%;
		background: radial-gradient(circle at 50% 40%, #fff8f2, #f6e2d3);
		box-shadow:
			0 0 0 14px rgba(255, 255, 255, 0.55),
			0 50px 90px -30px rgba(194, 78, 37, 0.55);
	}
	.ave-in {
		position: absolute;
		inset: 0;
		border-radius: 50%;
		overflow: hidden;
	}
	.ave-in img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: contain;
	}
	.metrica {
		position: absolute;
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 14px;
		border-radius: 16px;
		background: #fff;
		box-shadow: 0 18px 40px -18px rgba(58, 29, 18, 0.45);
	}
	.m-ico {
		width: 32px;
		height: 32px;
		border-radius: 10px;
		display: grid;
		place-items: center;
	}
	.m-txt {
		font-size: 13.5px;
		line-height: 1.25;
	}
	.m-txt span {
		color: #7a6252;
	}
	.canales {
		right: 20px;
		bottom: 20px;
		background: #2a140c;
		color: #fff;
		box-shadow: 0 18px 40px -18px rgba(42, 20, 12, 0.7);
	}
	.canales span {
		font-size: 13.5px;
		font-weight: 600;
	}
	.chat {
		flex: 1 1 460px;
		max-width: 560px;
		margin: 0 auto;
		border-radius: 28px;
		background: #fff;
		overflow: hidden;
		box-shadow:
			0 2px 4px rgba(58, 29, 18, 0.06),
			0 40px 80px -30px rgba(58, 29, 18, 0.5);
	}
	.chat-cab {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 16px;
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
	}
	.chat-cab > img {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		object-fit: cover;
	}
	.cab-txt {
		flex: 1;
		min-width: 0;
	}
	.nom {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 15px;
		font-weight: 600;
	}
	.sub {
		font-size: 12.5px;
		color: #7a6252;
	}
	.atiende {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 12.5px;
		font-weight: 600;
		color: #b0441d;
		background: rgba(210, 88, 47, 0.1);
		padding: 6px 10px 6px 6px;
		border-radius: 999px;
		white-space: nowrap;
	}
	.av {
		width: 22px;
		height: 22px;
		border-radius: 50%;
		overflow: hidden;
		background: #fcede5;
		box-shadow: 0 0 0 1.5px #d2582f;
	}
	.av img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.mensajes {
		height: 440px;
		padding: 18px 16px;
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
		gap: 8px;
		overflow: hidden;
		background: #f4ece2 url('/textura/arch-pattern-dark.svg');
	}
	.m-c {
		align-self: flex-start;
		max-width: 78%;
		font-size: 14px;
		line-height: 1.45;
		background: #fff;
		padding: 9px 13px;
		border-radius: 16px 16px 16px 4px;
		box-shadow: 0 1px 1px rgba(0, 0, 0, 0.05);
	}
	.m-h {
		align-self: flex-end;
		max-width: 80%;
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
	.m-tarjeta {
		align-self: flex-end;
		width: 250px;
		border-radius: 16px;
		background: #fce3d6;
		padding: 6px;
	}
	.prod {
		border-radius: 12px;
		background: #fff;
		overflow: hidden;
	}
	.prod img {
		width: 100%;
		height: 110px;
		object-fit: cover;
		display: block;
	}
	.prod-pie {
		padding: 10px 12px;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.prod-n {
		font-size: 13.5px;
		font-weight: 600;
	}
	.prod-s {
		font-size: 12px;
		color: #7a6252;
	}
	.prod-p {
		font-size: 15px;
		font-weight: 600;
		color: #b0441d;
	}
	.pago {
		border-radius: 12px;
		background: #fff;
		padding: 12px;
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.mp {
		width: 38px;
		height: 38px;
		border-radius: 10px;
		background: #e6f7fd;
		display: grid;
		place-items: center;
	}
	.pago-txt {
		flex: 1;
	}
	.pago-txt .a {
		font-size: 13.5px;
		font-weight: 600;
	}
	.pago-txt .b {
		font-size: 12px;
		color: #7a6252;
	}
	.m-sys {
		align-self: center;
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 12.5px;
		font-weight: 600;
		color: #1f6f4a;
		background: #e3f4ea;
		padding: 8px 14px;
		border-radius: 999px;
		margin-top: 4px;
	}
	.puntos {
		align-self: flex-end;
		display: flex;
		gap: 4px;
		padding: 12px 14px;
		border-radius: 16px 16px 4px 16px;
		background: #fce3d6;
	}
	.puntos span {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: #c24e25;
	}
	.chat-pie {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 14px;
		border-top: 1px solid rgba(58, 29, 18, 0.08);
	}
	.chat-pie > span:first-child {
		flex: 1;
		font-size: 13px;
		color: #7a6252;
	}
	.tomar {
		font-size: 13px;
		font-weight: 600;
		padding: 8px 14px;
		border-radius: 999px;
		background: #2a140c;
		color: #fff;
		white-space: nowrap;
	}
	.capacidades {
		margin-top: 72px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 14px;
	}
	.cap {
		padding: 24px;
		border-radius: 22px;
		background: rgba(255, 255, 255, 0.75);
		backdrop-filter: blur(8px);
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.05);
	}
	.cap-ico {
		width: 44px;
		height: 44px;
		border-radius: 13px;
		background: #b0441d;
		color: #fff;
		display: grid;
		place-items: center;
	}
	.cap-t {
		margin-top: 16px;
		font-size: 17px;
		font-weight: 600;
	}
	.cap-d {
		margin-top: 6px;
		font-size: 14.5px;
		line-height: 1.5;
		color: #5c4536;
	}
	.control {
		margin-top: 14px;
		padding: 26px 28px;
		border-radius: 24px;
		background: #2a140c;
		color: #fbf7f2;
		display: flex;
		gap: 32px;
		flex-wrap: wrap;
		align-items: center;
		box-shadow: 0 30px 60px -30px rgba(42, 20, 12, 0.6);
	}
	.ctl-txt {
		flex: 1 1 280px;
	}
	.ctl-t {
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.ctl-d {
		margin-top: 6px;
		font-size: 15px;
		line-height: 1.5;
		color: #d9c8ba;
	}
	.ctl-grilla {
		flex: 2 1 520px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 10px;
	}
	.ctl {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px;
		border-radius: 16px;
		border: 1px solid rgba(251, 247, 242, 0.14);
		background: rgba(251, 247, 242, 0.05);
		color: #fbf7f2;
		text-align: left;
	}
	.ctl-n {
		flex: 1;
		font-size: 14px;
		font-weight: 500;
		line-height: 1.35;
	}
	.tg {
		position: relative;
		width: 42px;
		height: 24px;
		flex: none;
		border-radius: 999px;
		background: #d9cbbe;
		transition: background 0.25s;
	}
	.tg.on {
		background: #d2582f;
	}
	.tg span {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 20px;
		height: 20px;
		border-radius: 50%;
		background: #fff;
		transition: left 0.25s cubic-bezier(0.2, 0.7, 0.1, 1);
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
	}
	.tg.on span {
		left: 20px;
	}
	/* Celular: la escena se compacta para que las métricas no se salgan, y la cabecera del chat
	   pasa el "Hornero atendiendo" abajo en vez de estrujar el nombre. */
	@media (max-width: 640px) {
		.dos {
			margin-top: 40px;
			gap: 28px;
		}
		.escena {
			flex-basis: 100%;
			min-height: 380px;
		}
		.ave {
			width: 250px;
			height: 250px;
			box-shadow:
				0 0 0 10px rgba(255, 255, 255, 0.55),
				0 40px 70px -30px rgba(194, 78, 37, 0.55);
		}
		.metrica {
			padding: 8px 11px;
			gap: 8px;
			font-size: 12.5px;
		}
		.m0 {
			top: 10px !important;
		}
		.m1 {
			top: 110px !important;
		}
		.m2 {
			left: 0 !important;
			bottom: 64px !important;
		}
		.canales {
			right: 0;
			bottom: 0;
		}
		.canales span {
			font-size: 12.5px;
		}
		.chat-cab {
			flex-wrap: wrap;
			row-gap: 8px;
		}
		.cab-txt {
			flex: 1 1 calc(100% - 52px);
		}
		.atiende {
			margin-left: 52px;
		}
	}
</style>
