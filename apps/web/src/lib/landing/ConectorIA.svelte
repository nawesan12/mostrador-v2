<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';

	/** "Tu negocio, desde la IA que ya usás." — conector MCP de Mostrador (handoff: #ia). */
	const PROVEEDORES = [
		['claude', 'Claude', I.siClaude, '#D97757'],
		['chatgpt', 'ChatGPT', I.siOpenai, '#10A37F'],
		['gemini', 'Gemini', I.siGooglegemini, '#4285F4'],
		['mcp', 'Otra IA (MCP)', I.phPlugsConnectedBold, '#8A5A3E']
	] as const;
	let elegido = $state<string>('claude');
	let hecho = $state(false);
	const actual = $derived(PROVEEDORES.find((p) => p[0] === elegido)!);
</script>

<section id="ia">
	<div class="grilla" use:anim={'reveal'}>
		<div>
			<div class="etiq"><span></span>Conector de IA</div>
			<h2>Tu negocio, desde la IA que ya usás.</h2>
			<p class="bajada">
				Conectá Mostrador a Claude, ChatGPT, Gemini o cualquier IA compatible con MCP. Entrás con tu
				cuenta, elegís qué puede hacer y le hablás como a un empleado más.
			</p>
			<div class="provs" role="group" aria-label="Elegí tu IA">
				{#each PROVEEDORES as [k, t, ico, c] (k)}
					{@const on = elegido === k}
					<button
						type="button"
						class:on
						aria-pressed={on}
						onclick={() => {
							elegido = k;
							hecho = false;
						}}
					>
						<Icon icon={ico} size={20} style="color:{c};flex:none" />
						<span class="p-n">{t}</span>
						<Icon
							icon={on ? I.phCheckCircleFill : I.phArrowRightBold}
							size={15}
							style="color:{on ? '#D2582F' : '#A8907F'}"
						/>
					</button>
				{/each}
			</div>
			<div class="seguro">
				<Icon icon={I.phShieldCheckFill} size={16} style="color:#2F8F63" />Reembolsos y descuentos
				siempre te piden confirmación.
			</div>
		</div>
		<div class="chat">
			<div class="cab">
				<Icon icon={actual[2]} size={18} style="color:{actual[3]}" /><b>{actual[1]}</b><span
					class="esp"
				></span><span class="conectado"><span></span>Mostrador conectado</span>
			</div>
			<div class="yo">¿Qué se vendió hoy y qué me falta reponer?</div>
			<div class="tool">
				<Icon icon={I.phPlugsConnectedBold} size={13} />mostrador · ventas_del_dia, stock_bajo
			</div>
			<div class="ia">
				Hoy van <b>$ 412.300</b> en 38 pedidos, 14 por WhatsApp. Se terminan el viernes: Termo 1 L negro
				(3 u.) y Yerba orgánica 1 kg (5 u.).
			</div>
			<div class="yo">Subile 8% al termo y pedile 20 a Cerámica Sur.</div>
			<div class="confirmar">
				<div class="c-t">Confirmá estos cambios en Mostrador</div>
				<div class="c-filas">
					<span><span>Termo 1 L negro</span><b>$ 42.900 → $ 46.300</b></span>
					<span><span>Pedido a Cerámica Sur</span><b>20 u. por WhatsApp</b></span>
				</div>
				<div class="c-acc">
					<button type="button" class="ok" class:hecho onclick={() => (hecho = true)}
						><Icon icon={hecho ? I.phCheckBold : I.phCheckCircleFill} size={14} />{hecho
							? 'Hecho'
							: 'Confirmar'}</button
					>
					<span class="cancelar">Cancelar</span>
				</div>
			</div>
		</div>
	</div>
</section>

<style>
	section {
		max-width: 1240px;
		margin: 0 auto;
		padding: 120px 28px 0;
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 440px), 1fr));
		gap: 48px;
		align-items: center;
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
	.bajada {
		margin: 18px 0 0;
		font-size: 18px;
		line-height: 1.5;
		color: #6b5244;
		max-width: 520px;
		text-wrap: pretty;
	}
	.provs {
		margin-top: 26px;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
		max-width: 520px;
	}
	.provs button {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 14px;
		border-radius: 16px;
		border: 1.5px solid rgba(42, 20, 12, 0.12);
		background: #fff;
		color: #2a140c;
		text-align: left;
		transition:
			border-color 0.2s,
			background 0.2s;
	}
	.provs button.on {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.06);
	}
	.p-n {
		flex: 1;
		min-width: 0;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.seguro {
		margin-top: 16px;
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13.5px;
		color: #8a7466;
	}
	.chat {
		padding: 20px;
		border-radius: 30px;
		background: #fff;
		box-shadow:
			0 40px 80px -40px rgba(42, 20, 12, 0.4),
			0 0 0 1px rgba(42, 20, 12, 0.06);
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.cab {
		display: flex;
		align-items: center;
		gap: 8px;
		padding-bottom: 10px;
		border-bottom: 1px solid rgba(42, 20, 12, 0.06);
	}
	.cab b {
		font-size: 14px;
	}
	.esp {
		flex: 1;
	}
	.conectado {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12px;
		font-weight: 600;
		color: #2f8f63;
	}
	.conectado span {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: #2f8f63;
	}
	.yo {
		align-self: flex-end;
		max-width: 85%;
		padding: 10px 14px;
		border-radius: 18px 18px 4px 18px;
		background: #2a140c;
		color: #fbf7f2;
		font-size: 14px;
	}
	.tool {
		align-self: flex-start;
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 5px 10px;
		border-radius: 999px;
		background: #f6efe6;
		font-size: 11.5px;
		font-weight: 600;
		color: #8a7466;
	}
	.ia {
		align-self: flex-start;
		max-width: 92%;
		padding: 12px 14px;
		border-radius: 18px 18px 18px 4px;
		background: #f6efe6;
		font-size: 14px;
		line-height: 1.5;
		color: #2a140c;
	}
	.confirmar {
		align-self: flex-start;
		width: 92%;
		padding: 12px;
		border-radius: 18px;
		border: 1.5px solid rgba(210, 88, 47, 0.35);
		background: #fff8f3;
	}
	.c-t {
		font-size: 13px;
		font-weight: 600;
		color: #2a140c;
	}
	.c-filas {
		margin-top: 8px;
		display: flex;
		flex-direction: column;
		gap: 5px;
		font-size: 13px;
		color: #5c4536;
	}
	.c-filas > span {
		display: flex;
		justify-content: space-between;
	}
	.c-acc {
		margin-top: 10px;
		display: flex;
		gap: 6px;
	}
	.ok {
		padding: 8px 14px;
		border-radius: 999px;
		border: 0;
		background: #d2582f;
		color: #fff;
		font-size: 13px;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.ok.hecho {
		background: #2f8f63;
	}
	.cancelar {
		padding: 8px 14px;
		border-radius: 999px;
		background: #f3ece4;
		font-size: 13px;
		font-weight: 600;
		color: #5c4536;
	}
</style>
