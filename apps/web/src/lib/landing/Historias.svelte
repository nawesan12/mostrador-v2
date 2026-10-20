<script lang="ts">
	import { anim } from '#lib/anim.ts';
	import { unsplash } from './formato.ts';

	/** "Negocios reales, creciendo de verdad." (handoff: #clientes). Testimonios de ejemplo del diseño. */
	const HISTORIAS = [
		{
			img: unsplash('1556745753-b2904692b3cd', 800),
			q: 'Antes anotaba los pedidos de WhatsApp en un cuaderno. Hoy Hornero responde de noche y yo me levanto con ventas.',
			quien: 'Valeria Sosa',
			negocio: 'Café La Esquina · Rosario',
			etiqueta: 'Ventas por WhatsApp',
			metrica: '×3 en 6 meses'
		},
		{
			img: unsplash('1687422810663-c316494f725a', 800),
			q: 'Me pasé desde Tiendanube un sábado a la tarde. El lunes ya estaba vendiendo y dejé de pagar comisión.',
			quien: 'Julián Pereyra',
			negocio: 'Taller Pereyra · Mendoza',
			etiqueta: 'Ahorro en comisiones',
			metrica: '$ 2,1 M al año'
		},
		{
			img: unsplash('1764173039506-c5cc7b8ec14b', 800),
			q: 'Tengo el local, la web y vendo a revendedoras. Antes eran tres sistemas, ahora es uno.',
			quien: 'Norma Ruiz',
			negocio: 'Ñandú Indumentaria · CABA',
			etiqueta: 'Facturación mensual',
			metrica: '$ 9,4 M'
		}
	];
</script>

<section id="clientes">
	<div class="cab" use:anim={'reveal'}>
		<div class="etiq"><span></span>Historias</div>
		<h2>Negocios reales, creciendo de verdad.</h2>
	</div>
	<div class="lista">
		{#each HISTORIAS as h (h.quien)}
			<figure use:anim={'reveal'}>
				<img data-pxi src={h.img} alt={h.quien} loading="lazy" />
				<div class="sombra"></div>
				<div class="metrica">
					<div class="m-l">{h.etiqueta}</div>
					<div class="m-v">{h.metrica}</div>
				</div>
				<figcaption>
					<div class="q">“{h.q}”</div>
					<div class="quien">{h.quien}</div>
					<div class="negocio">{h.negocio}</div>
				</figcaption>
			</figure>
		{/each}
	</div>
</section>

<style>
	section {
		max-width: 1240px;
		margin: 0 auto;
		padding: 120px 28px;
	}
	.cab {
		max-width: 720px;
	}
	.etiq {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #d2582f;
	}
	.etiq span {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #d2582f;
	}
	h2 {
		margin: 14px 0 0;
		font-size: clamp(36px, 4.6vw, 60px);
		line-height: 1;
		letter-spacing: -0.04em;
		font-weight: 600;
		text-wrap: balance;
	}
	.lista {
		margin-top: 56px;
		display: flex;
		gap: 18px;
		flex-wrap: wrap;
	}
	figure {
		flex: 1 1 320px;
		margin: 0;
		position: relative;
		min-height: 520px;
		border-radius: 26px;
		overflow: hidden;
		box-shadow: 0 24px 50px -30px rgba(58, 29, 18, 0.5);
	}
	figure img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.sombra {
		position: absolute;
		inset: 0;
		background: linear-gradient(180deg, rgba(42, 20, 12, 0) 30%, rgba(42, 20, 12, 0.9));
	}
	.metrica {
		position: absolute;
		left: 18px;
		top: 18px;
		padding: 10px 14px;
		border-radius: 14px;
		background: rgba(255, 255, 255, 0.94);
		color: #2a140c;
	}
	.m-l {
		font-size: 12px;
		color: #7a6252;
	}
	.m-v {
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	figcaption {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		padding: 26px;
		color: #fff;
	}
	.q {
		font-size: 19px;
		line-height: 1.4;
		font-weight: 500;
		letter-spacing: -0.01em;
		text-wrap: pretty;
	}
	.quien {
		margin-top: 16px;
		font-size: 14px;
		font-weight: 600;
	}
	.negocio {
		font-size: 13px;
		color: #e6d6c8;
	}
</style>
