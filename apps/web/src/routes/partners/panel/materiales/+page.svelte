<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { PRUEBA_DIAS } from '#lib/marca/datos.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { linkDe } from '#lib/partners/programa.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { FAMILIAS } from '../../../marca/familias.ts';

	/** Materiales: las piezas de Redes e Imprimibles de /marca, con el link del partner ya puesto. */
	let { data } = $props();
	const ui = getUi();
	const p = $derived(data.panel!);
	const link = $derived(linkDe(p.partner.codigo));
	const familias = FAMILIAS.filter(
		(f) => f.ruta === '/marca/redes' || f.ruta === '/marca/imprimibles'
	);
	const QUE: Record<string, string> = {
		'/marca/redes': 'Para tus posteos y anuncios: el texto de abajo lleva tu link.',
		'/marca/imprimibles': 'Los QR de flyers, afiche, one-pager y roll-up llevan a tu link.'
	};

	const textos = $derived([
		{
			id: 'whatsapp',
			t: 'Para WhatsApp',
			v: `¡Hola! Te recomiendo Tu Local Web para vender online: tienda, Mercado Pago, envíos y factura en un solo lugar. Probalo ${PRUEBA_DIAS} días gratis desde acá: ${link}`
		},
		{
			id: 'instagram',
			t: 'Para Instagram',
			v: `Armamos tiendas con Tu Local Web: catálogo, cobros con Mercado Pago y envíos, todo desde el celular. ${PRUEBA_DIAS} días gratis con nuestro link: ${link}`
		},
		{
			id: 'mail',
			t: 'Para un mail a clientes',
			v: `Para tu tienda online usamos Tu Local Web. Creá tu cuenta con este link y nosotros cargamos todo: ${link}. Si te lo piden, el cupón es ${p.partner.cupon}.`
		}
	]);
	async function copiar(texto: string, que: string) {
		try {
			await navigator.clipboard.writeText(texto);
			ui.toast({ k: 'ok', title: 'Texto copiado', desc: que });
		} catch {
			ui.toast({ k: 'warn', title: 'No se pudo copiar', desc: 'Seleccionalo y copialo a mano' });
		}
	}
</script>

<div class="familias" use:anim={{ t: 'rise' }}>
	{#each familias as f (f.ruta)}
		<div class="tarjeta familia">
			<img src="/hornero/{f.hornero}.png" alt="" width="72" height="72" />
			<div class="f-txt">
				<div class="t-cab">
					<span>{f.titulo}</span><span class="chip-f cuenta">{f.piezas} piezas</span>
				</div>
				<div class="t-d">{f.texto}</div>
				<div class="t-d">{QUE[f.ruta]}</div>
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- /marca con el ref puesto -->
				<a
					class="oscuro abrir"
					href="{f.href}?ref={encodeURIComponent(p.partner.codigo)}"
					target="_blank"
					rel="noopener"><Icon icon={I.phArrowUpRightBold} size={14} />Abrir con mi link</a
				>
			</div>
		</div>
	{/each}
</div>

<div class="tarjeta" use:anim={{ t: 'rise', delay: 40 }}>
	<div class="t-cab"><span>Textos para compartir</span></div>
	<div class="t-d">Con tu link puesto. Copialos tal cual o cambialos a tu&nbsp;manera.</div>
	<div class="filas">
		{#each textos as x (x.id)}
			<div class="opcion">
				<span class="o-txt"><span class="o-t">{x.t}</span><span class="o-d texto">{x.v}</span></span
				>
				<button type="button" class="oscuro" onclick={() => copiar(x.v, x.t)}
					><Icon icon={I.phCopyBold} size={14} />Copiar</button
				>
			</div>
		{/each}
	</div>
</div>

<style>
	.familias {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr));
		gap: 12px;
	}
	.familia {
		display: flex;
		gap: 14px;
		align-items: flex-start;
	}
	.familia img {
		flex: none;
		object-fit: contain;
	}
	.f-txt {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.cuenta {
		background: var(--chip);
	}
	.abrir {
		margin-top: 8px;
		align-self: flex-start;
	}
	.texto {
		margin-top: 2px;
		line-height: 1.45;
		overflow-wrap: anywhere;
	}
	.opcion {
		align-items: flex-start;
	}
	@media (max-width: 759px) {
		.familia img {
			width: 56px;
			height: 56px;
		}
		.opcion {
			flex-wrap: wrap;
		}
		.opcion .oscuro {
			margin-left: auto;
		}
	}
</style>
