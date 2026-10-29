<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import { COBROS_REALES, CONFIGURACION } from './datos.ts';
	import Etiqueta from './Etiqueta.svelte';
	import { getOnboarding } from './estado.svelte.ts';
	import Logo from './Logo.svelte';
	import PorConfigurar from './PorConfigurar.svelte';

	/**
	 * Paso 6, Cobros, envíos y facturas. En la demo, lo "ya configurado" para confirmar; con la API,
	 * lo que hay que dejar listo antes de vender, con el link a cada pantalla.
	 */
	const o = getOnboarding();
</script>

{#if o.demo}
	<Etiqueta rotulo="Cobros, envíos y facturas" titulo="Ya lo configuramos. Solo confirmá." />
	<div class="grupos">
		{#each CONFIGURACION as g (g.t)}
			<div class="grupo">
				<div class="g-cab">
					<span class="g-ico"><Icon icon={g.i} size={18} /></span><span>{g.t}</span>
				</div>
				<div class="items">
					{#each g.items as it (it.clave)}
						<div class="item">
							<Logo k={it.logo} tam={30} icono={16} radio={9} />
							<span class="i-txt"
								><span class="i-t">{it.t}</span><span class="i-d">{it.d}</span></span
							>
							{#if it.falta}
								<span class="falta">Falta 1 paso</span>
							{:else}
								<Interruptor
									checked={o.e.config[it.clave]}
									label={it.t}
									onchange={() => o.alternarConfig(it.clave)}
								/>
							{/if}
						</div>
					{/each}
				</div>
			</div>
		{/each}
	</div>
{:else}
	<Etiqueta rotulo="Cobros, envíos y facturas" titulo="Lo que falta para vender." />
	<PorConfigurar
		items={COBROS_REALES}
		nota="Sin un medio de pago tu tienda se ve pero no cobra: empezá por Mercado Pago o transferencia."
	/>
{/if}

<style>
	.grupos {
		margin-top: 22px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 360px), 1fr));
		gap: 16px;
	}
	.grupo {
		padding: 18px;
		border-radius: 22px;
		background: #fff;
		box-shadow:
			0 1px 2px rgba(58, 29, 18, 0.06),
			0 20px 44px -30px rgba(58, 29, 18, 0.4);
	}
	.g-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 16px;
		font-weight: 600;
	}
	.g-ico {
		width: 34px;
		height: 34px;
		border-radius: 10px;
		background: #fcede5;
		color: #b0441d;
		display: grid;
		place-items: center;
	}
	.items {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-top: 1px solid rgba(58, 29, 18, 0.07);
	}
	.i-txt {
		flex: 1;
		min-width: 0;
	}
	.i-t {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.i-d {
		display: block;
		font-size: 12.5px;
		color: #7a6252;
		line-height: 1.35;
	}
	.falta {
		font-size: 12px;
		font-weight: 700;
		color: #8a5f0e;
		background: rgba(227, 169, 59, 0.2);
		padding: 4px 9px;
		border-radius: 999px;
		white-space: nowrap;
	}
</style>
