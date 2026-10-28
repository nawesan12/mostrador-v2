<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import * as I from '#lib/icons.ts';
	import { marca } from '#lib/marcas.ts';
	import { SITE } from '#lib/site.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { getShell } from './shell.svelte.ts';

	/** Crear link de pago (handoff, ovLink): monto, concepto, cuotas y cómo le llega por WhatsApp. */
	const shell = getShell();
	const ui = getUi();
	const HOST = SITE.replace(/^https?:\/\//, '');

	let monto = $state('');
	let concepto = $state('');
	let cuotas = $state(true);
	let creando = $state(false);
	// Cada vez que se abre, toma lo precargado.
	$effect(() => {
		if (!shell.link) return;
		const p = shell.linkPrevio;
		monto = p.monto ? p.monto.toLocaleString('es-AR') : '';
		concepto = p.concepto ?? '';
		cuotas = true;
	});
	const valor = $derived(Number(monto.replace(/\D/g, '')) || 0);
	const para = $derived(shell.linkPrevio.para?.split(' ')[0]);
	/** El monto se escribe con puntos de miles mientras se tipea. */
	function alEscribir(e: Event & { currentTarget: HTMLInputElement }) {
		const n = Number(e.currentTarget.value.replace(/\D/g, '').slice(0, 9));
		monto = n ? n.toLocaleString('es-AR') : '';
	}

	async function crear() {
		if (!valor || concepto.trim().length < 2 || creando) return;
		creando = true;
		const r = await fetch('/panel/links', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				monto: valor,
				concepto,
				cuotas,
				pedido: shell.linkPrevio.pedido ?? null
			})
		});
		creando = false;
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({ k: 'err', title: 'No se creó el link', desc: m ?? 'Probá de nuevo.' });
		}
		const { codigo } = (await r.json()) as { codigo: string };
		const url = `${SITE}/p/${codigo}`;
		try {
			await navigator.clipboard.writeText(url);
		} catch {
			// Sin permiso de portapapeles: el link igual queda en el aviso.
		}
		const tel = shell.linkPrevio.telefono?.replace(/\D/g, '') ?? '';
		const texto = `¡Hola${para ? ` ${para}` : ''}! Te paso el link para pagar $ ${valor.toLocaleString('es-AR')}: ${url}`;
		shell.link = false;
		ui.toast({
			k: 'link',
			title: 'Link copiado',
			desc: `${HOST}/p/${codigo} · $ ${valor.toLocaleString('es-AR')}`,
			action: 'Enviar por WhatsApp',
			onAction: () => {
				window.open(`https://wa.me/${tel}?text=${encodeURIComponent(texto)}`, '_blank', 'noopener');
			}
		});
	}
	const MEDIOS = ['mp', 'visa', 'master', 'modo', 'uala'];
</script>

<Modal bind:open={shell.link} titulo="Crear link de pago" ancho={560} padding={0}>
	<form
		onsubmit={(e) => {
			e.preventDefault();
			void crear();
		}}
	>
		<div class="cab">
			<span>Crear link de pago</span>
			<button type="button" class="x" aria-label="Cerrar" onclick={() => (shell.link = false)}
				><Icon icon={I.phXBold} size={14} /></button
			>
		</div>
		<div class="cuerpo">
			<label class="monto">
				<span class="m-t">Monto</span>
				<span class="m-v"
					><span class="sig">$</span><input
						value={monto}
						oninput={alEscribir}
						inputmode="numeric"
						placeholder="0"
						aria-label="Monto"
					/></span
				>
			</label>
			<label class="fila">
				<Icon icon={I.phTextAlignLeft} size={17} style="color:var(--mu)" />
				<input
					bind:value={concepto}
					maxlength="140"
					placeholder="Qué se cobra (lo ve quien paga)"
					aria-label="Concepto"
				/>
			</label>
			<div class="fila">
				<Icon icon={I.phCreditCard} size={17} style="color:var(--mu)" />
				<span class="f-t">{cuotas ? 'Hasta 6 cuotas sin interés' : 'Un solo pago'}</span>
				<Interruptor bind:checked={cuotas} label="Cuotas sin interés" />
			</div>
			<div class="acepta">
				<span>Acepta</span>
				{#each MEDIOS as k (k)}
					{@const m = marca(k)}
					<span class="medio" style:background={m.bg}
						>{#if m.icon}<Icon
								icon={m.icon}
								size={18}
								style="color:{m.col ?? 'inherit'}"
							/>{:else}<span class="sigla" style:color={m.col}>{m.sigla}</span>{/if}</span
					>
				{/each}
			</div>
			<div class="wa">
				<div class="wa-t">Así le llega por WhatsApp</div>
				<div class="burbuja">
					¡Hola{para ? ` ${para}` : ''}! Te paso el link para pagar <b>$ {monto || '0'}</b>:<br
					/><span class="url">{HOST}/p/…</span>
				</div>
			</div>
		</div>
		<div class="pie">
			<button type="button" class="btn btn-secondary" onclick={() => (shell.link = false)}
				>Cancelar</button
			>
			<button class="crear" disabled={!valor || concepto.trim().length < 2 || creando}
				><Icon icon={I.phCopyBold} size={15} />Crear y copiar link</button
			>
		</div>
	</form>
</Modal>

<style>
	.cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 20px 22px 0;
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.x {
		width: 34px;
		height: 34px;
		border: 0;
		border-radius: 50%;
		background: var(--chip);
		color: var(--tx);
		display: grid;
		place-items: center;
	}
	.cuerpo {
		padding: 18px 22px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.monto {
		display: block;
		padding: 16px 18px;
		border-radius: 18px;
		background: var(--sf3);
	}
	.m-t {
		display: block;
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	.m-v {
		display: flex;
		align-items: baseline;
		gap: 6px;
		margin-top: 4px;
	}
	.sig {
		font-size: 34px;
		font-weight: 600;
		color: var(--mu);
	}
	.m-v input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font:
			600 40px 'Instrument Sans',
			sans-serif;
		letter-spacing: -0.04em;
		color: var(--tx);
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 14px;
		border-radius: 14px;
		border: 1px solid var(--ln2);
	}
	.fila input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 14px;
		color: var(--tx);
	}
	.f-t {
		flex: 1;
		font-size: 14px;
	}
	.acepta {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.acepta > span:first-child {
		font-size: 13px;
		color: var(--mu);
		margin-right: 4px;
	}
	.medio {
		width: 34px;
		height: 26px;
		border-radius: 7px;
		display: grid;
		place-items: center;
		box-shadow: var(--sh);
	}
	.sigla {
		font-size: 7px;
		font-weight: 800;
	}
	.wa {
		padding: 14px;
		border-radius: 18px;
		background: #efe7dc url('/textura/arch-pattern-dark.svg');
	}
	.wa-t {
		font-size: 12px;
		font-weight: 600;
		color: #7a6252;
		margin-bottom: 8px;
	}
	.burbuja {
		margin-left: auto;
		max-width: 330px;
		padding: 10px 12px;
		border-radius: 14px 14px 4px 14px;
		background: #d8f3df;
		color: #1f2a20;
		font-size: 13.5px;
		line-height: 1.45;
	}
	.url {
		color: #1f6f4a;
		font-weight: 600;
	}
	.pie {
		padding: 14px 22px 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
	.crear {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 20px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.3);
	}
	.crear:disabled {
		opacity: 0.55;
	}
</style>
