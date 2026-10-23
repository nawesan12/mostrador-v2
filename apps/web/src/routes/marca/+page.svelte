<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import { phArrowRightBold } from '#lib/icons.ts';
	import { sinViuda } from '#lib/marca/datos.ts';
	import { FAMILIAS } from './familias.ts';
</script>

<Seo
	title="Marca · Mostrador"
	description="Papelería, redes, imprimibles e imágenes para compartir de Mostrador, listas para exportar."
	path="/marca"
	noindex
/>

<section class="lienzo">
	<header style="max-width:1400px">
		<img src="/wordmark-brown.png" alt="mostrador" style="height:26px;display:block" />
		<h1 style="margin:22px 0 0;font-size:56px;line-height:1;letter-spacing:-.04em;font-weight:600">
			Marca
		</h1>
		<p style="margin:12px 0 0;font-size:18px;color:#6B5244;max-width:680px;text-wrap:pretty">
			{sinViuda(
				'Todas las piezas de Mostrador a su medida real. Cada una se baja en PNG; las de imprenta, también en PDF con 3 mm de sangrado.'
			)}
		</p>
	</header>

	<div class="grilla">
		{#each FAMILIAS as f (f.ruta)}
			<a class="familia" href={f.href}>
				<span class="arco"><img src="/hornero/{f.hornero}.png" alt="" /></span>
				<span class="txt">
					<b>{f.titulo}</b>
					<span>{sinViuda(f.texto)}</span>
					<span class="pie">{f.piezas} piezas <Icon icon={phArrowRightBold} size={15} /></span>
				</span>
			</a>
		{/each}
	</div>
</section>

<style>
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 300px), 1fr));
		gap: 20px;
		max-width: 1400px;
	}
	/* Las tarjetas de la home de marca, con el arco terracota del handoff (avatar de Hornero). */
	.familia {
		display: flex;
		flex-direction: column;
		border-radius: 22px;
		background: #fbf7f2;
		color: #2a140c;
		overflow: hidden;
		box-shadow: 0 18px 40px -24px rgba(42, 20, 12, 0.45);
		transition: transform 0.2s var(--ease);
	}
	.familia:hover {
		transform: translateY(-2px);
		color: #2a140c;
	}
	.arco {
		position: relative;
		height: 170px;
		background: #d2582f url(/textura/arch-pattern-accent.svg) center/22px;
		display: grid;
		place-items: end center;
	}
	.arco img {
		width: 150px;
		height: 150px;
		object-fit: contain;
		margin-bottom: -6px;
	}
	.txt {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 18px 20px 20px;
	}
	.txt b {
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.txt span {
		font-size: 14px;
		line-height: 1.5;
		color: #6b5244;
	}
	.txt .pie {
		margin-top: 6px;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-weight: 600;
		color: #c24e25;
	}
</style>
