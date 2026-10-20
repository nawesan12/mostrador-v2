<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';
	import Feed from './Feed.svelte';
	import Vendido from './Vendido.svelte';
	import { cara, unsplash } from './formato.ts';

	/*
	 * Hero del handoff: la composición "panorámica" desde 1180 px (heroDesktop por defecto) y la
	 * apilada en tablet/mobile. Se dibujan las dos y el CSS muestra una: sin saltos al hidratar.
	 */
	// Los CTA llevan al onboarding, que es el registro (handoff: Onboarding Mostrador.dc.html).
	const registro = resolve('/onboarding');
	const ARCO =
		'M3 76a3 3 0 0 1-3-3V5a5 5 0 0 1 5-5h90a5 5 0 0 1 5 5v68a3 3 0 0 1-3 3H82a3 3 0 0 1-3-3V21H21v52a3 3 0 0 1-3 3Z';
	const PRODUCTOS = [
		['1609541994821-d909982e2f1b', 'Mate de cerámica', '$ 18.500'],
		['1675001077188-809370ce3279', 'Yerba orgánica 1 kg', '$ 8.200'],
		['1634741011278-9dc2f5b6fc0b', 'Termo 1 L', '$ 42.900'],
		['1444157545135-c045be691b05', 'Kit matero', '$ 58.400']
	] as const;
	const CARAS = [
		'1531058240690-006c446962d8',
		'1687422811062-a966b55cb217',
		'1764173039506-c5cc7b8ec14b',
		'1520333789090-1afc82db536a'
	];
</script>

<!-- Escritorio (≥ 1180 px): composición panorámica -->
<section class="desk">
	<div class="d-luz"></div>
	<div class="d-arcos" data-px="0.3"></div>
	<div class="d-in">
		<a class="pildora" href="#hornero" use:anim={'rise'}>
			<span class="avatar"><img src="/hornero/avatar.png" alt="" /></span>Conocé a Hornero, tu
			asistente con IA<Icon icon={I.phArrowRight} size={14} />
		</a>
		<h1 use:anim={{ t: 'rise', delay: 80 }}>
			Todo para vender, en un solo lugar. <span>Más seguro y por menos.</span>
		</h1>
		<p class="bajada" use:anim={{ t: 'rise', delay: 160 }}>
			Tienda online, WhatsApp, Mercado Libre, local y facturación. Para el que recién arranca y para
			el que ya factura millones — sin comisión por venta.
		</p>
		<div class="ctas" use:anim={{ t: 'rise', delay: 240 }}>
			<a class="cta-pri" href={registro}
				>Crear mi tienda gratis <Icon icon={I.phArrowRightBold} size={16} /></a
			>
			<a class="cta-sec" href="#plantillas"
				><Icon icon={I.phSquaresFour} size={17} />Ver plantillas</a
			>
		</div>
		<div class="checks" use:anim={{ t: 'fade', delay: 380 }}>
			<span
				><Icon icon={I.phCheckCircleFill} size={16} style="color:#2F8F63" />14 días gratis, sin
				tarjeta</span
			>
			<span
				><Icon icon={I.phCheckCircleFill} size={16} style="color:#2F8F63" />0% de comisión por venta</span
			>
			<span
				><Icon icon={I.phCheckCircleFill} size={16} style="color:#2F8F63" />Te migramos gratis</span
			>
		</div>

		<div class="escena">
			<svg class="arco" data-px="0.12" viewBox="0 0 100 76" aria-hidden="true"
				><defs
					><linearGradient id="hA-arch" x1="0" x2="0" y1="0" y2="1"
						><stop offset="0" stop-color="#E8875F"></stop><stop offset="1" stop-color="#C24E25"
						></stop></linearGradient
					></defs
				><path fill="url(#hA-arch)" d={ARCO}></path></svg
			>

			<div class="foto-local" data-px="0.2">
				<div class="foto-in" use:anim={{ t: 'rise', delay: 200 }}>
					<img src={unsplash('1556745753-b2904692b3cd', 700)} alt="Comerciante en su local" />
					<div class="foto-sombra"></div>
					<div class="foto-txt">
						<div class="quien">Valeria · Café La Esquina</div>
						<div class="dice">“Ahora vendo mientras atiendo el local”</div>
					</div>
				</div>
			</div>

			<div class="tienda" data-px="-0.04">
				<div class="navegador" use:anim={{ t: 'rise', delay: 120 }}>
					<div class="barra-nav">
						<span style:background="#E0A595"></span><span style:background="#E7CB8F"></span><span
							style:background="#A9CBA8"
						></span>
						<span class="url"
							><Icon
								icon={I.phLockSimpleFill}
								size={11}
								style="color:#2F8F63"
							/>laesquina.com.ar</span
						>
					</div>
					<div class="sitio">
						<div class="sitio-cab">
							<img src={unsplash('1556745753-b2904692b3cd', 80, '&h=80')} alt="" />
							<b>Café La Esquina</b><span class="esp"></span>
							<span class="tenue">Yerbas</span><span class="tenue">Mates</span><span class="tenue"
								>Café</span
							><span class="tenue">Regalos</span>
							<span class="bolsa"
								><Icon icon={I.phHandbag} size={19} /><span class="cant">2</span></span
							>
						</div>
						<div class="sitio-hero">
							<img data-pxi src={unsplash('1687920492018-e21cad9023d9', 1400)} alt="" />
							<div class="sitio-sombra"></div>
							<div class="sitio-txt">
								<div class="sup">Nueva cosecha</div>
								<div class="tit">El mate de cada mañana</div>
								<div class="ver">Ver productos</div>
							</div>
						</div>
						<div class="sitio-prods">
							{#each PRODUCTOS as [id, n, p] (id)}
								<div>
									<div class="p-foto"><img src={unsplash(id, 300, '&h=300')} alt="" /></div>
									<div class="p-n">{n}</div>
									<div class="p-p">{p}</div>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>

			<div class="celu" data-px="-0.16">
				<div class="celu-in" use:anim={{ t: 'rise', delay: 260 }}>
					<div class="pantalla">
						<div class="chat-cab">
							<img src={cara('1531058240690-006c446962d8')} alt="" /><span class="quien"
								>Lucía F.</span
							><Icon icon={I.logosWhatsappIcon} size={15} />
						</div>
						<div class="chat">
							<div class="m-c">¿Tenés el termo en negro?</div>
							<div class="m-h">
								<span class="firma">✦ Hornero</span><span class="burbuja"
									>¡Sí! Quedan 3. Te llega el jueves 📦</span
								>
							</div>
							<div class="m-pago">
								<div class="pago-in">
									<span class="mp"
										><Icon icon={I.siMercadopago} size={16} style="color:#00B1EA" /></span
									>
									<span class="pago-txt"
										><span class="a">Pagar $ 44.600</span><span class="b">6 cuotas sin interés</span
										></span
									>
								</div>
							</div>
							<div class="m-ok">
								<Icon icon={I.phCheckCircleFill} size={12} />Pago aprobado · Pedido #1049
							</div>
						</div>
					</div>
				</div>
			</div>

			<div class="feed-d" data-px="-0.3"><Feed /></div>
			<div class="vend-d" data-px="-0.22"><Vendido /></div>
			<div class="ave-d" data-px="-0.1">
				<div class="flota" use:anim={{ t: 'float', dur: 3800 }}>
					<img src="/hornero/saludo.png" alt="Hornero saludando" />
				</div>
			</div>
		</div>
	</div>
	<div style="height:80px"></div>
</section>

<!-- Tablet y mobile (< 1180 px) -->
<section class="mob">
	<div class="m-luz"></div>
	<div class="m-arcos" data-px="0.3"></div>
	<div class="m-in">
		<div class="m-texto">
			<a class="pildora" href="#asistente" use:anim={'rise'}>
				<span class="nuevo">Nuevo</span>
				Asistente con IA que vende por vos
				<Icon icon={I.phArrowRight} size={14} />
			</a>
			<h1 class="m-h1" use:anim={{ t: 'rise', delay: 80 }}>
				Todo lo que tu negocio necesita para vender. <span>Más seguro y por menos.</span>
			</h1>
			<p class="m-bajada" use:anim={{ t: 'rise', delay: 160 }}>
				Tienda online, WhatsApp, Mercado Libre, local y facturación en un solo lugar. Para el que
				recién arranca y para el que ya factura millones — sin comisión por venta.
			</p>
			<div class="m-ctas" use:anim={{ t: 'rise', delay: 240 }}>
				<a class="cta-pri m" href={registro}
					>Probar gratis 14 días <Icon icon={I.phArrowRightBold} size={16} /></a
				>
				<a class="cta-sec m" href="#migrar">Me paso desde Tiendanube</a>
			</div>
			<div class="prueba" use:anim={{ t: 'fade', delay: 400 }}>
				<div class="caras">
					{#each CARAS as c (c)}<img src={cara(c)} alt="" />{/each}
				</div>
				<div class="estrellas">
					<span>★★★★★</span><br />Emprendedores y comercios de todo el país
				</div>
			</div>
		</div>

		<div class="m-escena">
			<svg class="m-arco" use:anim={'fade'} data-px="-0.14" viewBox="0 0 100 76" aria-hidden="true"
				><defs
					><linearGradient id="v3-arch" x1="0" x2="1" y1="0" y2="1"
						><stop offset="0" stop-color="#E8875F"></stop><stop offset="1" stop-color="#C24E25"
						></stop></linearGradient
					></defs
				><path fill="url(#v3-arch)" d={ARCO}></path></svg
			>
			<div class="m-foto" use:anim={{ t: 'rise', delay: 120 }} data-px="0.05">
				<img
					src={unsplash('1556745753-b2904692b3cd', 1200)}
					data-pxi
					alt="Comerciante atendiendo en su local"
				/>
				<div class="m-foto-sombra"></div>
				<div class="m-foto-txt">
					<div class="a">Café La Esquina · Rosario</div>
					<div class="b">Vende en el local, por WhatsApp y en Mercado Libre</div>
				</div>
			</div>
			<div class="m-hornero" use:anim={{ t: 'float', dur: 4200, delay: 400 }}>
				<span class="m-av"><img src="/hornero/avatar.png" alt="Hornero" /></span>
				<span class="m-hor-txt"
					><b>Hornero</b> respondió 41 consultas<br /><span>mientras atendías el local</span></span
				>
			</div>
			<div class="m-feed" data-px="-0.2">
				<Feed
					grande
					sombra="0 1px 0 #fff inset,0 18px 40px -16px rgba(58,29,18,.45),0 0 0 1px rgba(58,29,18,.06)"
				/>
			</div>
			<div class="m-vend"><Vendido /></div>
		</div>
	</div>
</section>

<style>
	/* ----- Escritorio ----- */
	.desk {
		position: relative;
		overflow: hidden;
	}
	.mob {
		display: none;
		position: relative;
	}
	@media (max-width: 1179px) {
		.desk {
			display: none;
		}
		.mob {
			display: block;
		}
	}
	.d-luz {
		position: absolute;
		inset: 0;
		background: radial-gradient(1100px 700px at 50% 0%, rgba(232, 135, 95, 0.2), transparent 65%);
	}
	.d-arcos,
	.m-arcos {
		position: absolute;
		inset: 0;
		background-image: url('/textura/arch-pattern-dark.svg');
		mask-image: linear-gradient(180deg, #000, transparent 70%);
	}
	.d-in {
		position: relative;
		max-width: 1320px;
		margin: 0 auto;
		padding: 64px 40px 0;
		text-align: center;
	}
	.pildora {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		padding: 6px 14px 6px 6px;
		border-radius: 999px;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.1);
		color: #2a140c;
		font-size: 13.5px;
		font-weight: 500;
		white-space: nowrap;
		max-width: 100%;
		box-shadow: 0 4px 14px -6px rgba(58, 29, 18, 0.2);
	}
	.avatar {
		width: 26px;
		height: 26px;
		border-radius: 50%;
		overflow: hidden;
		background: #fcede5;
	}
	.avatar img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	h1 {
		text-wrap: balance;
		margin: 26px auto 0;
		max-width: 1080px;
		font-size: clamp(56px, 6.4vw, 92px);
		line-height: 0.95;
		letter-spacing: -0.05em;
		font-weight: 600;
	}
	h1 span,
	.m-h1 span {
		color: #d2582f;
	}
	.bajada {
		text-wrap: pretty;
		margin: 24px auto 0;
		max-width: 640px;
		font-size: 19.5px;
		line-height: 1.55;
		color: #5c4536;
	}
	.ctas {
		margin-top: 32px;
		display: flex;
		gap: 12px;
		justify-content: center;
		flex-wrap: wrap;
	}
	.cta-pri {
		display: flex;
		align-items: center;
		gap: 10px;
		color: #fff;
		background: linear-gradient(180deg, #e06a3e, #d2582f 60%, #bf4b22);
		font-size: 16px;
		font-weight: 600;
		padding: 17px 28px;
		border-radius: 999px;
		white-space: nowrap;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			inset 0 -2px 0 rgba(0, 0, 0, 0.12),
			0 16px 34px -14px rgba(210, 88, 47, 0.75);
	}
	.cta-pri:hover {
		color: #fff;
		transform: translateY(-1px);
	}
	.cta-sec {
		display: flex;
		align-items: center;
		gap: 10px;
		color: #2a140c;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.14);
		font-size: 16px;
		font-weight: 600;
		padding: 17px 24px;
		border-radius: 999px;
		white-space: nowrap;
		box-shadow: 0 6px 16px -10px rgba(58, 29, 18, 0.3);
	}
	.cta-sec:hover {
		color: #2a140c;
		border-color: rgba(58, 29, 18, 0.3);
	}
	.checks {
		margin-top: 22px;
		display: flex;
		gap: 22px;
		justify-content: center;
		flex-wrap: wrap;
		font-size: 14px;
		color: #7a6252;
	}
	.checks span {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.escena {
		position: relative;
		margin: 64px auto 0;
		height: 600px;
		max-width: 1240px;
	}
	.arco {
		position: absolute;
		left: 50%;
		top: 40px;
		width: 900px;
		margin-left: -450px;
	}
	.foto-local {
		position: absolute;
		left: 0;
		top: 150px;
		width: 290px;
		height: 360px;
		z-index: 1;
	}
	.foto-in {
		position: relative;
		width: 100%;
		height: 100%;
		border-radius: 26px;
		overflow: hidden;
		transform: rotate(-5deg);
		box-shadow:
			0 30px 60px -24px rgba(58, 29, 18, 0.55),
			0 0 0 6px #fbf7f2;
	}
	.foto-in img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.foto-sombra {
		position: absolute;
		inset: 0;
		background: linear-gradient(180deg, transparent 55%, rgba(42, 20, 12, 0.7));
	}
	.foto-txt {
		position: absolute;
		left: 18px;
		right: 18px;
		bottom: 16px;
		color: #fff;
		text-align: left;
	}
	.quien {
		font-size: 12.5px;
		opacity: 0.85;
	}
	.dice {
		font-size: 15px;
		font-weight: 600;
		line-height: 1.25;
	}
	.tienda {
		position: absolute;
		left: 50%;
		top: 0;
		width: 740px;
		margin-left: -370px;
		z-index: 2;
	}
	.navegador {
		border-radius: 20px;
		overflow: hidden;
		background: #fff;
		box-shadow:
			0 2px 4px rgba(58, 29, 18, 0.08),
			0 50px 100px -40px rgba(42, 20, 12, 0.6),
			0 0 0 1px rgba(58, 29, 18, 0.08);
		text-align: left;
	}
	.barra-nav {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 14px;
		background: #f1e9df;
	}
	.barra-nav > span:not(.url) {
		width: 10px;
		height: 10px;
		border-radius: 50%;
	}
	.url {
		flex: 1;
		margin: 0 60px;
		padding: 5px 12px;
		border-radius: 999px;
		background: #fff;
		font-size: 12.5px;
		color: #5c4536;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
	}
	.sitio {
		background: #f3e9dc;
		color: #3b2416;
	}
	.sitio-cab {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 22px;
	}
	.sitio-cab img {
		width: 30px;
		height: 30px;
		border-radius: 50%;
		object-fit: cover;
	}
	.sitio-cab b {
		font-size: 16px;
		letter-spacing: -0.02em;
	}
	.esp {
		flex: 1;
	}
	.tenue {
		font-size: 13px;
		opacity: 0.7;
	}
	.bolsa {
		position: relative;
	}
	.cant {
		position: absolute;
		top: -4px;
		right: -6px;
		width: 14px;
		height: 14px;
		border-radius: 50%;
		background: #c24e25;
		color: #fff;
		font-size: 9px;
		font-weight: 700;
		display: grid;
		place-items: center;
	}
	.sitio-hero {
		position: relative;
		margin: 0 18px;
		height: 230px;
		border-radius: 16px;
		overflow: hidden;
	}
	.sitio-hero img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.sitio-sombra {
		position: absolute;
		inset: 0;
		background: linear-gradient(90deg, rgba(30, 15, 8, 0.7), transparent 65%);
	}
	.sitio-txt {
		position: absolute;
		left: 28px;
		top: 50%;
		transform: translateY(-50%);
		color: #fff;
		max-width: 300px;
	}
	.sup {
		font-size: 12px;
		font-weight: 600;
		opacity: 0.85;
	}
	.tit {
		margin-top: 4px;
		font-size: 32px;
		font-weight: 700;
		line-height: 1;
		letter-spacing: -0.03em;
	}
	.ver {
		margin-top: 14px;
		display: inline-block;
		font-size: 13px;
		font-weight: 600;
		padding: 9px 16px;
		border-radius: 999px;
		background: #6b3a22;
	}
	.sitio-prods {
		padding: 18px 18px 22px;
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 14px;
	}
	.p-foto {
		aspect-ratio: 1;
		border-radius: 12px;
		overflow: hidden;
	}
	.p-foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.p-n {
		margin-top: 7px;
		font-size: 13px;
		font-weight: 600;
	}
	.p-p {
		font-size: 13px;
		font-weight: 700;
		color: #6b3a22;
	}
	.celu {
		position: absolute;
		right: 0;
		top: 110px;
		width: 270px;
		z-index: 3;
	}
	.celu-in {
		border-radius: 40px;
		padding: 9px;
		background: #2a140c;
		transform: rotate(4deg);
		box-shadow: 0 40px 80px -30px rgba(42, 20, 12, 0.75);
	}
	.pantalla {
		border-radius: 32px;
		overflow: hidden;
		background: #efe7dc;
		text-align: left;
	}
	.chat-cab {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 30px 12px 10px;
		background: #fff;
	}
	.chat-cab img {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		object-fit: cover;
	}
	.chat-cab .quien {
		flex: 1;
		font-size: 12.5px;
		font-weight: 600;
		opacity: 1;
	}
	.chat {
		padding: 12px 10px 14px;
		display: flex;
		flex-direction: column;
		gap: 6px;
		height: 380px;
		justify-content: flex-end;
		background-image: url('/textura/arch-pattern-dark.svg');
	}
	.m-c {
		align-self: flex-start;
		max-width: 82%;
		font-size: 12px;
		line-height: 1.4;
		background: #fff;
		padding: 7px 10px;
		border-radius: 12px 12px 12px 3px;
	}
	.m-h {
		align-self: flex-end;
		max-width: 86%;
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 2px;
	}
	.firma {
		font-size: 9.5px;
		font-weight: 600;
		color: #c24e25;
	}
	.burbuja {
		font-size: 12px;
		line-height: 1.4;
		background: #fce3d6;
		padding: 7px 10px;
		border-radius: 12px 12px 3px 12px;
	}
	.m-pago {
		align-self: flex-end;
		width: 190px;
		border-radius: 12px;
		background: #fce3d6;
		padding: 5px;
	}
	.pago-in {
		border-radius: 9px;
		background: #fff;
		padding: 8px;
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.mp {
		width: 30px;
		height: 30px;
		border-radius: 8px;
		background: #e6f7fd;
		display: grid;
		place-items: center;
	}
	.pago-txt {
		flex: 1;
	}
	.pago-txt .a {
		display: block;
		font-size: 11.5px;
		font-weight: 600;
	}
	.pago-txt .b {
		display: block;
		font-size: 10px;
		color: #7a6252;
	}
	.m-ok {
		align-self: center;
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 10.5px;
		font-weight: 600;
		color: #1f6f4a;
		background: #e3f4ea;
		padding: 5px 10px;
		border-radius: 999px;
	}
	.feed-d {
		position: absolute;
		left: 230px;
		top: 40px;
		width: 300px;
		z-index: 4;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.vend-d {
		position: absolute;
		right: 250px;
		bottom: 30px;
		z-index: 4;
	}
	.ave-d {
		position: absolute;
		left: 200px;
		bottom: -10px;
		width: 190px;
		height: 190px;
		z-index: 5;
	}
	.flota {
		width: 100%;
		height: 100%;
	}
	.flota img {
		width: 100%;
		height: 100%;
		object-fit: contain;
		filter: drop-shadow(0 20px 24px rgba(42, 20, 12, 0.3));
	}

	/* ----- Tablet y mobile ----- */
	.m-luz {
		position: absolute;
		inset: 0;
		background:
			radial-gradient(900px 600px at 80% 10%, rgba(232, 135, 95, 0.22), transparent 65%),
			radial-gradient(700px 500px at 0% 90%, rgba(227, 169, 59, 0.12), transparent 60%);
	}
	.m-arcos {
		mask-image: linear-gradient(180deg, #000, transparent 80%);
	}
	.m-in {
		position: relative;
		max-width: 1240px;
		margin: 0 auto;
		padding: 72px var(--lado) 96px;
		display: flex;
		gap: 56px;
		flex-wrap: wrap;
		align-items: center;
	}
	/* Sin min-width: 0, como el handoff: la píldora (nowrap) define el ancho mínimo de la columna. */
	.m-texto {
		flex: 1 1 500px;
	}
	.nuevo {
		font-size: 12px;
		font-weight: 700;
		color: #fff;
		background: #d2582f;
		padding: 3px 9px;
		border-radius: 999px;
	}
	.m-h1 {
		margin: 26px 0 0;
		max-width: none;
		font-size: clamp(44px, 6vw, 82px);
		line-height: 0.98;
		letter-spacing: -0.045em;
	}
	.m-bajada {
		text-wrap: pretty;
		margin: 26px 0 0;
		max-width: 560px;
		font-size: 19px;
		line-height: 1.55;
		color: #5c4536;
	}
	.m-ctas {
		margin-top: 34px;
		display: flex;
		gap: 12px;
		flex-wrap: wrap;
	}
	.cta-pri.m,
	.cta-sec.m {
		flex: none;
		padding: 16px 26px;
	}
	.cta-sec.m {
		padding: 16px 24px;
	}
	.prueba {
		margin-top: 30px;
		display: flex;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
	}
	.caras {
		display: flex;
	}
	.caras img {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		object-fit: cover;
		border: 2px solid #fbf7f2;
	}
	.caras img + img {
		margin-left: -10px;
	}
	.estrellas {
		font-size: 14px;
		color: #5c4536;
		line-height: 1.35;
	}
	.estrellas span {
		color: #e3a93b;
		letter-spacing: 1px;
	}
	.m-escena {
		flex: 1 1 480px;
		position: relative;
		min-height: 600px;
		min-width: 0;
	}
	.m-arco {
		position: absolute;
		right: -40px;
		top: -10px;
		width: 88%;
		opacity: 0.95;
	}
	.m-foto {
		position: absolute;
		left: 8%;
		right: 4%;
		top: 60px;
		bottom: 40px;
		border-radius: 28px;
		overflow: hidden;
		box-shadow:
			0 2px 4px rgba(58, 29, 18, 0.08),
			0 30px 60px -20px rgba(58, 29, 18, 0.45),
			0 0 0 6px #fbf7f2;
	}
	.m-foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.m-foto-sombra {
		position: absolute;
		inset: 0;
		background: linear-gradient(180deg, transparent 55%, rgba(42, 20, 12, 0.55));
	}
	.m-foto-txt {
		position: absolute;
		left: 22px;
		bottom: 20px;
		color: #fff;
	}
	.m-foto-txt .a {
		font-size: 13px;
		opacity: 0.85;
	}
	.m-foto-txt .b {
		font-size: 17px;
		font-weight: 600;
	}
	.m-hornero {
		position: absolute;
		left: -20px;
		bottom: 120px;
		z-index: 4;
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 16px 8px 8px;
		border-radius: 999px;
		background: #fff;
		box-shadow: 0 18px 40px -16px rgba(58, 29, 18, 0.5);
	}
	.m-av {
		width: 44px;
		height: 44px;
		flex: none;
		border-radius: 50%;
		overflow: hidden;
		background: #fcede5;
		box-shadow: 0 0 0 2px #d2582f;
	}
	.m-av img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.m-hor-txt {
		font-size: 13px;
		line-height: 1.3;
		white-space: nowrap;
	}
	.m-hor-txt span {
		color: #7a6252;
	}
	.m-feed {
		position: absolute;
		left: -30px;
		top: 10px;
		width: 300px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		z-index: 3;
	}
	.m-vend {
		position: absolute;
		right: -10px;
		bottom: 0;
		z-index: 3;
	}
</style>
