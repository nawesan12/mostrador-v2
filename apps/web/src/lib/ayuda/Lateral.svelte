<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import HorneroAyuda from './HorneroAyuda.svelte';

	/** La columna de la derecha del centro de ayuda: Hornero, una persona y migrar. */
</script>

<aside class="lateral">
	<HorneroAyuda />
	<div class="persona">
		<b>¿Preferís una persona?</b>
		<div class="vias">
			<a href={resolve('/(sitio)/contacto')}
				><Icon icon={I.phEnvelopeSimpleFill} size={18} style="color:#b0441d" />Escribinos · te
				respondemos en el día</a
			>
			<a href={resolve('/(sitio)/estado')}
				><Icon icon={I.phActivityBold} size={18} style="color:#237550" />Estado del sistema</a
			>
		</div>
	</div>
	<a class="migrar" href={resolve('/(sitio)/migrar')}>
		<img src="/hornero/caja.webp" alt="" width="56" height="56" />
		<span><b>¿Venís de otra plataforma?</b><span>Te mudamos gratis en un día</span></span>
		<Icon icon={I.phArrowRightBold} size={15} />
	</a>
</aside>

<style>
	.lateral {
		position: sticky;
		top: 88px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.persona {
		padding: 20px;
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.persona b {
		font-size: 16px;
	}
	.vias {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 14px;
	}
	.vias a {
		display: flex;
		align-items: center;
		gap: 10px;
		color: #2a140c;
		text-decoration: none;
	}
	.migrar {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 18px;
		border-radius: 24px;
		background: #fcede5;
		text-decoration: none;
		color: #2a140c;
	}
	.migrar img {
		object-fit: contain;
	}
	.migrar > span {
		flex: 1;
	}
	.migrar b {
		display: block;
		font-size: 15px;
	}
	.migrar > span > span {
		font-size: 13px;
		color: #5c4536;
	}
	@media (max-width: 860px) {
		.lateral {
			position: static;
		}
	}
</style>
