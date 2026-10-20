<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';
	import { getLanding } from './estado.svelte.ts';
	import { unsplash } from './formato.ts';

	/** "Te mudás en una tarde." (handoff: #migrar), con la barra de importación en vivo. */
	const l = getLanding();
	const PASOS = [
		['1', 'Conectás Tiendanube, Empretienda o Mercado Shops', '#2A140C'],
		['2', 'Revisás lo importado antes de publicar', '#2A140C'],
		['3', 'Apuntás tu dominio. Listo.', '#D2582F']
	] as const;
</script>

<section id="migrar">
	<div class="foto" use:anim={'reveal'}>
		<img
			src={unsplash('1773146916231-e6aeae932080', 1000)}
			data-pxi
			alt="Emprendedora preparando pedidos"
			loading="lazy"
		/>
		<div class="progreso">
			<div class="p-cab">
				<span>Importando desde Tiendanube</span><span class="pct">{l.mig}%</span>
			</div>
			<div class="barra"><div style:width="{l.mig}%"></div></div>
			<div class="items">
				<span
					><Icon icon={I.phCheckCircleFill} size={14} style="color:#2F8F63" />1.204 productos</span
				>
				<span
					><Icon icon={I.phCheckCircleFill} size={14} style="color:#2F8F63" />3.880 clientes</span
				>
				<span><Icon icon={I.phCircleNotch} size={14} style="color:#C24E25" />Pedidos</span>
			</div>
		</div>
	</div>
	<div class="txt" use:anim={'reveal'}>
		<div class="etiq"><span></span>Migración asistida</div>
		<h2>Te mudás en una tarde.</h2>
		<p>
			Traemos productos, fotos, clientes y pedidos. Tu dominio sigue siendo tuyo y no perdés
			posicionamiento en Google.
		</p>
		<div class="pasos">
			{#each PASOS as [n, t, bg] (n)}
				<div class="paso"><span style:background={bg}>{n}</span><span class="p-t">{t}</span></div>
			{/each}
		</div>
	</div>
</section>

<style>
	section {
		max-width: 1240px;
		margin: 0 auto;
		padding: 120px 28px;
		display: flex;
		gap: 56px;
		flex-wrap: wrap;
		align-items: center;
	}
	.foto {
		flex: 1 1 440px;
		position: relative;
		min-height: 440px;
		border-radius: 26px;
		overflow: hidden;
		box-shadow: 0 24px 50px -30px rgba(58, 29, 18, 0.5);
	}
	.foto img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.progreso {
		position: absolute;
		left: 20px;
		right: 20px;
		bottom: 20px;
		padding: 20px;
		border-radius: 20px;
		background: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(8px);
	}
	.p-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		font-size: 14px;
		font-weight: 600;
	}
	.pct {
		color: #c24e25;
	}
	.barra {
		margin-top: 12px;
		height: 8px;
		border-radius: 8px;
		background: #efe6da;
		overflow: hidden;
	}
	.barra div {
		height: 100%;
		border-radius: 8px;
		background: linear-gradient(90deg, #d2582f, #e8875f);
		transition: width 0.8s;
	}
	.items {
		margin-top: 12px;
		display: flex;
		gap: 16px;
		flex-wrap: wrap;
		font-size: 13px;
		color: #5c4536;
	}
	.items span {
		display: flex;
		align-items: center;
		gap: 5px;
	}
	.txt {
		flex: 1 1 420px;
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
	.txt > p {
		margin: 20px 0 0;
		font-size: 17px;
		line-height: 1.55;
		color: #5c4536;
		max-width: 460px;
	}
	.pasos {
		margin-top: 28px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.paso {
		display: flex;
		gap: 14px;
		align-items: center;
		padding: 14px 16px;
		border-radius: 16px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.05);
	}
	.paso > span:first-child {
		width: 32px;
		height: 32px;
		flex: none;
		border-radius: 50%;
		color: #fff;
		display: grid;
		place-items: center;
		font-size: 14px;
		font-weight: 600;
	}
	.p-t {
		font-size: 15.5px;
		font-weight: 500;
	}
</style>
