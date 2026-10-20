<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';

	/** "Tu negocio no se cae. Y nada se pierde." (handoff: #seguridad). */
	const GARANTIAS = [
		[
			I.phShieldCheckFill,
			'Doble verificación obligatoria',
			'Para vos y cada persona de tu equipo.'
		],
		[
			I.phArrowCounterClockwiseBold,
			'Deshacer hasta 30 días',
			'Borraste algo: lo recuperás en un clic.'
		],
		[I.phDatabaseFill, 'Copias cada hora', 'Cifradas y guardadas en otra región.'],
		[I.phUserFocusFill, 'Permisos por persona', 'El cajero cobra, no ve la facturación.'],
		[
			I.phFingerprintFill,
			'Registro de cada cambio',
			'Quién cambió un precio, cuándo y desde dónde.'
		],
		[I.phLockKeyFill, 'Nunca guardamos tarjetas', 'Los pagos se procesan fuera de tu tienda.']
	] as const;
	// 60 días; el 38 tuvo una demora (diseño). Cuando exista /estado, sale del monitor real.
	const UPTIME = Array.from({ length: 60 }, (_, i) => (i === 37 ? '#E3A93B' : '#2F8F63'));
</script>

<section id="seguridad">
	<div class="in">
		<div class="txt" use:anim={'reveal'}>
			<div class="etiq"><span></span>Seguridad</div>
			<h2>Tu negocio no se cae. Y nada se pierde.</h2>
			<p>Lo que en otras plataformas es opcional, acá viene activado desde el primer día.</p>
			<div class="estado">
				<div class="e-cab">
					<span class="e-ok"
						><span class="punto"><span class="pulso" use:anim={'pulse'}></span><span></span></span
						>Todos los sistemas funcionando</span
					>
					<span class="e-dias">últimos 90 días</span>
				</div>
				<div class="barras">
					{#each UPTIME as c, i (i)}<span style:background={c}></span>{/each}
				</div>
			</div>
		</div>
		<div class="grilla">
			{#each GARANTIAS as [ico, t, d] (t)}
				<div class="g" use:anim={'reveal'}>
					<span class="g-ico"><Icon icon={ico} size={22} /></span>
					<div class="g-t">{t}</div>
					<div class="g-d">{d}</div>
				</div>
			{/each}
		</div>
	</div>
</section>

<style>
	section {
		background: #f4ece2;
		border-top: 1px solid rgba(58, 29, 18, 0.08);
	}
	.in {
		max-width: 1240px;
		margin: 0 auto;
		padding: 120px var(--lado);
		display: flex;
		gap: 56px;
		flex-wrap: wrap;
	}
	.txt {
		flex: 1 1 380px;
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
		max-width: 440px;
	}
	.estado {
		margin-top: 32px;
		padding: 20px;
		border-radius: 20px;
		background: #fff;
		box-shadow: 0 16px 36px -24px rgba(58, 29, 18, 0.4);
	}
	.e-cab {
		display: flex;
		gap: 12px;
		justify-content: space-between;
		align-items: center;
	}
	.e-ok {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
	}
	.punto {
		position: relative;
		width: 8px;
		height: 8px;
	}
	.punto span {
		position: absolute;
		inset: 0;
		border-radius: 50%;
		background: #2f8f63;
	}
	.e-dias {
		white-space: nowrap;
		font-size: 13px;
		color: #7a6252;
	}
	.barras {
		margin-top: 14px;
		display: flex;
		gap: 2px;
		height: 30px;
	}
	.barras span {
		flex: 1;
		border-radius: 2px;
	}
	.grilla {
		flex: 1 1 480px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 12px;
		align-self: start;
	}
	.g {
		background: #fff;
		padding: 24px;
		border-radius: 20px;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.05);
	}
	.g-ico {
		width: 44px;
		height: 44px;
		border-radius: 13px;
		background: #fcede5;
		display: grid;
		place-items: center;
		color: #d2582f;
	}
	.g-t {
		margin-top: 16px;
		font-size: 16.5px;
		font-weight: 600;
	}
	.g-d {
		margin-top: 6px;
		font-size: 14.5px;
		line-height: 1.5;
		color: #5c4536;
	}
</style>
