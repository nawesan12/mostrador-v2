<script lang="ts">
	import CtaTienda from './CtaTienda.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import { phArrowRight } from '#lib/icons.ts';
	import { unsplash } from './formato.ts';

	/*
	 * "De la primera venta a facturar en serio" (handoff). Los precios "desde" del diseño eran de otro
	 * esquema (14.900/29.900/59.900); van los de los planes vigentes (docs/decisiones.md):
	 * WhatsApp arranca en Emprende, ML e IA en Pro, punto de venta y mayoristas en Negocio.
	 */
	const ETAPAS = [
		{
			tag: 'Estoy empezando',
			titulo: 'Tu primera tienda, hoy',
			texto:
				'Cargás productos con una foto, conectás WhatsApp y Mercado Pago, y ya estás vendiendo.',
			img: unsplash('1773146916231-e6aeae932080', 900),
			alt: 'Emprendedora en su taller',
			feats: ['Tienda online', 'WhatsApp', 'Link de pago'],
			precio: '$ 19.900'
		},
		{
			tag: 'Ya vendo y quiero crecer',
			titulo: 'Vendé en todos los canales',
			texto:
				'Mercado Libre, Instagram y la tienda con un solo stock. Campañas y carritos recuperados en automático.',
			img: unsplash('1687422808248-f807f4ea2a2e', 900),
			alt: 'Comerciante revisando pedidos en el celular',
			feats: ['Mercado Libre', 'Campañas', 'Asistente IA'],
			precio: '$ 49.900'
		},
		{
			tag: 'Tengo local y mayoristas',
			titulo: 'Todo el negocio en un panel',
			texto:
				'Punto de venta, listas de precio mayoristas, facturación ARCA y permisos para todo tu equipo.',
			img: unsplash('1556740758-90de374c12ad', 900),
			alt: 'Dueñas de un local',
			feats: ['Punto de venta', 'Mayoristas', 'Facturación'],
			precio: '$ 149.900'
		}
	];
</script>

<section>
	<div class="cab" use:anim={'reveal'}>
		<div class="tit">
			<div class="etiq"><span></span>Crece con vos</div>
			<h2>De la primera venta a facturar en serio.</h2>
		</div>
		<p>
			Empezás con lo justo y vas sumando herramientas a medida que crecés. Sin cambiar de plataforma
			nunca más.
		</p>
	</div>
	<div class="etapas">
		{#each ETAPAS as e (e.tag)}
			<div class="etapa" use:anim={'reveal'}>
				<div class="foto">
					<img data-pxi src={e.img} alt={e.alt} loading="lazy" />
					<span class="tag">{e.tag}</span>
				</div>
				<div class="cuerpo">
					<h3>{e.titulo}</h3>
					<p>{e.texto}</p>
					<div class="feats">
						{#each e.feats as f (f)}<span>{f}</span>{/each}
					</div>
					<div class="pie">
						<span class="desde">desde <b>{e.precio}</b> /mes</span>
						<a href="#precios" class="ir" aria-label="Ver precios"
							><Icon icon={phArrowRight} size={16} /></a
						>
					</div>
				</div>
			</div>
		{/each}
	</div>
	<CtaTienda />
</section>

<style>
	section {
		max-width: 1240px;
		margin: 0 auto;
		padding: 120px var(--lado);
	}
	.cab {
		display: flex;
		justify-content: space-between;
		align-items: end;
		gap: 40px;
		flex-wrap: wrap;
	}
	.tit {
		max-width: 680px;
	}
	.etiq {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #b0441d;
	}
	.etiq span {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #d2582f;
	}
	h2 {
		text-wrap: balance;
		margin: 14px 0 0;
		font-size: clamp(36px, 4.6vw, 60px);
		line-height: 1;
		letter-spacing: -0.04em;
		font-weight: 600;
	}
	.cab > p {
		max-width: 380px;
		margin: 0;
		font-size: 17px;
		line-height: 1.55;
		color: #5c4536;
	}
	.etapas {
		margin-top: 56px;
		display: flex;
		gap: 18px;
		flex-wrap: wrap;
	}
	.etapa {
		flex: 1 1 320px;
		border-radius: 26px;
		overflow: hidden;
		background: #fff;
		box-shadow:
			0 1px 2px rgba(58, 29, 18, 0.06),
			0 20px 40px -24px rgba(58, 29, 18, 0.35);
		display: flex;
		flex-direction: column;
		transition:
			transform 0.3s,
			box-shadow 0.3s;
	}
	.etapa:hover {
		transform: translateY(-4px);
		box-shadow:
			0 1px 2px rgba(58, 29, 18, 0.06),
			0 30px 60px -24px rgba(58, 29, 18, 0.45);
	}
	.foto {
		position: relative;
		height: 260px;
		overflow: hidden;
	}
	.foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
		transition: transform 0.8s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.tag {
		position: absolute;
		left: 16px;
		top: 16px;
		font-size: 13px;
		font-weight: 600;
		padding: 6px 12px;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.92);
		color: #2a140c;
	}
	.cuerpo {
		padding: 24px 26px 26px;
		display: flex;
		flex-direction: column;
		flex: 1;
	}
	h3 {
		margin: 0;
		font-size: 24px;
		letter-spacing: -0.03em;
		font-weight: 600;
	}
	.cuerpo > p {
		margin: 10px 0 0;
		font-size: 15px;
		line-height: 1.55;
		color: #5c4536;
	}
	.feats {
		margin-top: 18px;
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.feats span {
		font-size: 13px;
		font-weight: 500;
		padding: 6px 11px;
		border-radius: 999px;
		background: #f6eee4;
		color: #5c4536;
	}
	.pie {
		margin-top: auto;
		padding-top: 22px;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.desde {
		font-size: 14px;
		color: #7a6252;
	}
	.desde b {
		font-size: 20px;
		color: #2a140c;
		letter-spacing: -0.02em;
	}
	.ir {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: #2a140c;
		color: #fff;
		display: grid;
		place-items: center;
	}
	.ir:hover {
		color: #fff;
	}
</style>
