<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';

	/** "Tu Local Web vende. Caja ordena la plata." — la integración con Caja (handoff: #caja). */
	const BENEFICIOS = [
		[
			I.phArrowsLeftRightBold,
			'Sin cargar nada',
			'Las ventas y cobros entran a la actividad del negocio en Caja.'
		],
		[
			I.phReceiptBold,
			'Monotributo al día',
			'Lo que facturás en Tu Local Web suma al tope de tu categoría.'
		],
		[
			I.phCurrencyDollarBold,
			'Ganancia real',
			'Ajustada por inflación o al dólar MEP del día de cada venta.'
		],
		[
			I.phUserSwitchBold,
			'Lo tuyo, aparte',
			'Un retiro para vos se registra como personal, no como gasto.'
		]
	] as const;
	const MOVIMIENTOS = [
		{
			ico: I.logosWhatsappIcon,
			col: undefined,
			t: 'Venta #10480',
			s: 'Lucía · Termo 1 L',
			v: '+$ 44.600',
			vc: '#7FD3A6'
		},
		{
			ico: I.siMercadopago,
			col: '#00B1EA',
			t: 'Liquidación MP',
			s: 'comisión −$ 2.940',
			v: '+$ 182.300',
			vc: '#7FD3A6'
		},
		{
			ico: I.phTruckFill,
			col: '#E8875F',
			t: 'Etiquetas Correo Argentino',
			s: '8 envíos',
			v: '−$ 41.600',
			vc: '#F0A58A'
		},
		{
			ico: I.phReceiptFill,
			col: '#E3A93B',
			t: 'Factura C 0004-812',
			s: 'Autorizada por ARCA',
			v: '$ 44.600',
			vc: undefined
		}
	];
</script>

<section id="caja">
	<div class="bloque" use:anim={'reveal'}>
		<div class="txt">
			<div class="duo">
				<img src="/mark.svg" alt="Tu Local Web" width="30" height="30" /><span class="mas">+</span
				><img src="/caja/01-saludo.webp" alt="Caja" width="40" height="40" class="masc" /><span
					class="nuevo">Nuevo · Integración con Caja</span
				>
			</div>
			<h2>Tu Local Web vende. Caja ordena la plata.</h2>
			<p class="bajada">
				Cada venta, liquidación, comisión, envío y factura pasa sola a Caja, separada de tus gastos
				personales. Sabés cuánto ganaste de verdad, en pesos de hoy o en dólares.
			</p>
			<div class="beneficios">
				{#each BENEFICIOS as [ico, t, d] (t)}
					<div class="ben">
						<Icon icon={ico} size={20} style="color:#E8875F;flex:none;margin-top:2px" />
						<div>
							<b>{t}</b>
							<div class="ben-d">{d}</div>
						</div>
					</div>
				{/each}
			</div>
			<div class="ia">
				<div class="ia-cab">
					<b>Y manejala desde tu IA</b><span class="mcp">MCP</span><span class="esp"></span><span
						class="logos"
						><span title="Claude"><Icon icon={I.siClaude} size={18} /></span><span title="ChatGPT"
							><Icon icon={I.siOpenai} size={18} /></span
						></span
					>
				</div>
				<div class="ia-d">
					Conectá Caja a Claude, ChatGPT o la IA que uses, con MCP o como conector. Le preguntás por
					tu plata y carga movimientos por vos.
				</div>
				<div class="ia-chat">
					<div class="yo">¿Cuánto me quedó limpio de Tu Local Web este mes, en dólares?</div>
					<div class="caja-r">
						<span class="via">Caja · vía conector</span>US$ 4.412 al MEP. Vendiste $ 9,1 M y se
						fueron $ 2,3 M en comisiones, envíos y campañas.
					</div>
				</div>
			</div>
			<div class="acc">
				<a href="https://caja.iamnaw.dev" target="_blank" rel="noopener"
					>Conocé Caja<Icon icon={I.phArrowUpRightBold} size={14} /></a
				>
				<span>3 meses de Caja Pro Max con cualquier plan de Tu Local Web</span>
			</div>
		</div>
		<div class="flujo">
			<div class="col">
				<div class="col-t">En Tu Local Web</div>
				{#each MOVIMIENTOS as m (m.t)}
					<div class="mov">
						<Icon icon={m.ico} size={18} style={m.col ? `color:${m.col}` : ''} />
						<div class="mov-txt">
							<div class="a">{m.t}</div>
							<div class="b">{m.s}</div>
						</div>
						<b style:color={m.vc}>{m.v}</b>
					</div>
				{/each}
			</div>
			<div class="flechas">
				<Icon icon={I.phArrowRightBold} size={18} /><Icon
					icon={I.phArrowRightBold}
					size={18}
					style="opacity:.6"
				/><Icon icon={I.phArrowRightBold} size={18} style="opacity:.3" />
			</div>
			<div class="tarjeta">
				<div class="t-cab">
					<img src="/caja/05-exito.webp" alt="" width="34" height="34" />
					<div>
						<div class="t-n">Caja · Actividad</div>
						<div class="t-s">Negocio · octubre</div>
					</div>
				</div>
				<div class="t-l">Ganancia del mes</div>
				<div class="t-v">$ 6.812.400</div>
				<div class="t-l2">≈ US$ 4.412 al MEP</div>
				<div class="t-mono"><span>Monotributo · Cat. D</span><b>70% del tope</b></div>
				<div class="t-barra"><div></div></div>
				<div class="t-tip">Las comisiones subieron 12%. ¿Pasamos más ventas a transferencia?</div>
			</div>
		</div>
	</div>
</section>

<style>
	section {
		max-width: 1240px;
		margin: 0 auto;
		padding: 120px var(--lado) 40px;
	}
	.bloque {
		border-radius: 36px;
		background: #131211;
		color: #f4eee6;
		padding: clamp(20px, 5vw, 64px);
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 420px), 1fr));
		gap: 48px;
		align-items: center;
		overflow: hidden;
		position: relative;
	}
	.txt,
	.flujo {
		position: relative;
		z-index: 1;
	}
	.duo {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.mas {
		font-size: 20px;
		color: #8a7f76;
	}
	.masc {
		object-fit: contain;
	}
	.nuevo {
		margin-left: 6px;
		font-size: 13px;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: #e8875f;
	}
	h2 {
		margin: 22px 0 0;
		font-size: clamp(34px, 4.6vw, 58px);
		line-height: 1;
		letter-spacing: -0.04em;
		font-weight: 600;
		text-wrap: balance;
	}
	.bajada {
		margin: 18px 0 0;
		font-size: 18px;
		line-height: 1.5;
		color: #bdb2a8;
		max-width: 520px;
		text-wrap: pretty;
	}
	.beneficios {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 14px;
	}
	.ben {
		display: flex;
		gap: 10px;
	}
	.ben b {
		font-size: 15px;
	}
	.ben-d {
		font-size: 14px;
		color: #9c9188;
		margin-top: 2px;
	}
	.ia {
		margin-top: 22px;
		padding: 16px;
		border-radius: 20px;
		background: #1f1c1a;
		border: 1px solid rgba(244, 238, 230, 0.08);
	}
	.ia-cab {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.ia-cab b {
		font-size: 15px;
	}
	.mcp {
		font-size: 11px;
		font-weight: 700;
		padding: 3px 8px;
		border-radius: 999px;
		background: rgba(232, 135, 95, 0.16);
		color: #e8875f;
	}
	.esp {
		flex: 1;
	}
	.logos {
		display: flex;
		gap: 6px;
		color: #bdb2a8;
	}
	.ia-d {
		font-size: 13.5px;
		color: #9c9188;
		margin-top: 4px;
		line-height: 1.45;
	}
	.ia-chat {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.yo {
		align-self: flex-end;
		max-width: 85%;
		padding: 9px 13px;
		border-radius: 16px 16px 4px 16px;
		background: #e8875f;
		color: #131211;
		font-size: 13px;
		font-weight: 500;
	}
	.caja-r {
		align-self: flex-start;
		max-width: 90%;
		padding: 9px 13px;
		border-radius: 16px 16px 16px 4px;
		background: #2b2724;
		font-size: 13px;
		line-height: 1.45;
	}
	.via {
		font-size: 11px;
		font-weight: 600;
		color: #8a7f76;
		display: block;
		margin-bottom: 2px;
	}
	.acc {
		margin-top: 30px;
		display: flex;
		gap: 10px;
		flex-wrap: wrap;
	}
	.acc a {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 13px 22px;
		border-radius: 999px;
		background: #f4eee6;
		color: #131211;
		font-weight: 600;
		font-size: 15px;
	}
	.acc span {
		display: flex;
		align-items: center;
		padding: 13px 4px;
		font-size: 14px;
		color: #9c9188;
	}
	.flujo {
		display: grid;
		grid-template-columns: 1fr auto 1fr;
		gap: 12px;
		align-items: center;
	}
	.col {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.col-t {
		font-size: 12px;
		font-weight: 600;
		color: #8a7f76;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		margin-bottom: 2px;
	}
	.mov {
		padding: 12px 14px;
		border-radius: 16px;
		background: #1f1c1a;
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.mov-txt {
		flex: 1;
		min-width: 0;
	}
	.mov-txt .a {
		font-size: 13px;
		font-weight: 600;
	}
	.mov-txt .b {
		font-size: 12px;
		color: #8a7f76;
	}
	.mov b {
		font-size: 13px;
	}
	.flechas {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		color: #e8875f;
	}
	.tarjeta {
		padding: 16px;
		border-radius: 22px;
		background: #f4eee6;
		color: #131211;
		box-shadow: 0 30px 60px -20px rgba(0, 0, 0, 0.6);
	}
	.t-cab {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.t-cab img {
		object-fit: contain;
	}
	.t-n {
		font-size: 13px;
		font-weight: 700;
	}
	.t-s {
		font-size: 11.5px;
		color: #6e655d;
	}
	.t-l {
		margin-top: 12px;
		font-size: 12px;
		color: #6e655d;
	}
	.t-v {
		font-size: 28px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.t-l2 {
		font-size: 12px;
		color: #6e655d;
	}
	.t-mono {
		margin-top: 12px;
		font-size: 12px;
		color: #6e655d;
		display: flex;
		justify-content: space-between;
	}
	.t-mono b {
		color: #131211;
	}
	.t-barra {
		margin-top: 5px;
		height: 7px;
		border-radius: 4px;
		background: #e2d8cd;
		overflow: hidden;
	}
	.t-barra div {
		width: 70%;
		height: 100%;
		background: #d2582f;
		border-radius: 4px;
	}
	.t-tip {
		margin-top: 12px;
		padding: 9px 10px;
		border-radius: 12px;
		background: #eae1d6;
		font-size: 12px;
		line-height: 1.4;
	}
	/* El handoff deja el flujo en 3 columnas y a 375 px la tarjeta de Caja queda cortada: se apila. */
	@media (max-width: 560px) {
		.flujo {
			grid-template-columns: 1fr;
		}
		.flechas {
			flex-direction: row;
			justify-content: center;
			transform: rotate(90deg);
		}
	}
</style>
