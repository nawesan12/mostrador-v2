<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import * as I from '#lib/icons.ts';
	import { marca } from '#lib/marcas.ts';
	import { pedidoCertificado } from '#lib/panel/csr.ts';
	import { cuitValido, formatoCuit, limpiarCuit } from '#lib/panel/cuit.ts';
	import type { ArcaConfig, ArcaPrueba } from '#lib/server/datos/tipos.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Integraciones › ARCA › Configurar. El handoff no diseña este asistente: se arma con las piezas
	 * del modal de Integraciones (cabecera, secciones, pasos numerados, opciones con interruptor) y
	 * la barra de pasos de la cancelación del plan. Cuatro pasos: datos fiscales, pedido de
	 * certificado, certificado y prueba con ajustes.
	 */
	let {
		open = $bindable(false),
		arca,
		demo
	}: { open?: boolean; arca: ArcaConfig; demo: boolean } = $props();
	const ui = getUi();
	const m = marca('arca');

	const TITULOS = ['Datos fiscales', 'Pedido de certificado', 'Certificado', 'Probar y ajustar'];
	let paso = $state(0);
	let cuit = $state('');
	let razon = $state('');
	let condicion = $state<'monotributo' | 'responsable_inscripto'>('monotributo');
	let pv = $state('');
	let ambiente = $state<'homologacion' | 'produccion'>('homologacion');
	let csr = $state('');
	let pem = $state('');
	let error = $state<string | null>(null);
	let ocupado = $state(false);
	let prueba = $state<ArcaPrueba | null>(null);
	let pvAjuste = $state('');

	// Al abrir: conectada va directo a probar y ajustar; si no, a los datos (con lo que haya).
	let abierto = false;
	$effect(() => {
		if (open && !abierto) {
			paso = arca.conectada ? 3 : 0;
			cuit = arca.cuit ? formatoCuit(arca.cuit) : '';
			razon = arca.razonSocial;
			condicion =
				arca.condicion === 'responsable_inscripto' ? 'responsable_inscripto' : 'monotributo';
			pv = arca.puntoVenta ? String(arca.puntoVenta) : '';
			pvAjuste = pv;
			ambiente = arca.ambiente;
			error = null;
			prueba = null;
			pem = '';
		}
		abierto = open;
	});

	const vence = $derived(arca.certificadoVence ? Date.parse(arca.certificadoVence) : null);
	const diasParaVencer = $derived(
		vence === null ? null : Math.ceil((vence - Date.now()) / 86_400_000)
	);
	const fecha = (ms: number) =>
		new Date(ms).toLocaleDateString('es-AR', {
			day: 'numeric',
			month: 'long',
			year: 'numeric',
			timeZone: 'America/Argentina/Buenos_Aires'
		});

	async function pedir(metodo: 'POST' | 'PUT' | 'PATCH', cuerpo: unknown, ruta = '') {
		const r = await fetch(`/panel/integraciones/arca${ruta}`, {
			method: metodo,
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(cuerpo)
		});
		if (!r.ok) {
			const msg = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return { error: msg ?? 'No pudimos hablar con el servidor. Probá de nuevo.' };
		}
		return r.status === 204 ? {} : ((await r.json()) as Record<string, unknown>);
	}

	async function generar(e: SubmitEvent) {
		e.preventDefault();
		error = null;
		const c = limpiarCuit(cuit);
		const n = Number(pv);
		if (!cuitValido(c)) return (error = 'Revisá el CUIT: el dígito verificador no da.');
		if (!razon.trim()) return (error = 'Poné la razón social como figura en ARCA.');
		if (!Number.isInteger(n) || n < 1 || n > 99998)
			return (error = 'El punto de venta va de 1 a 99998.');
		ocupado = true;
		try {
			const alta = { cuit: c, razonSocial: razon.trim(), condicion, puntoVenta: n, ambiente };
			const r = await pedir('POST', alta);
			if ('error' in r) return (error = String(r.error));
			// En la demo el pedido lo arma este navegador; con la cuenta, Go (y guarda la clave).
			csr = 'pedido' in r ? String(r.pedido) : await pedidoCertificado(c, alta.razonSocial);
			paso = 1;
			if (!demo) await invalidateAll();
		} catch {
			error = 'No pudimos armar el pedido. Probá de nuevo.';
		} finally {
			ocupado = false;
		}
	}

	function descargar() {
		const url = URL.createObjectURL(new Blob([csr], { type: 'application/pkcs10' }));
		const a = document.createElement('a');
		a.href = url;
		a.download = `mostrador-${limpiarCuit(cuit)}.csr`;
		a.click();
		setTimeout(() => URL.revokeObjectURL(url), 1000);
	}

	async function leerArchivo(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const f = input.files?.[0];
		if (!f) return;
		if (f.size > 16_384) return (error = 'Ese archivo es muy grande para ser un certificado.');
		pem = (await f.text()).trim();
		input.value = '';
	}

	async function guardarCertificado(e: SubmitEvent) {
		e.preventDefault();
		error = null;
		if (!pem.trim()) return (error = 'Pegá el certificado o subí el archivo .crt.');
		ocupado = true;
		const r = await pedir('PUT', { certificado: pem.trim() });
		ocupado = false;
		if ('error' in r) return (error = String(r.error));
		await invalidateAll();
		paso = 3;
		ui.toast({ k: 'arca', title: 'ARCA conectada', desc: 'Probá la conexión para confirmar' });
	}

	async function probar() {
		ocupado = true;
		const r = await fetch('/panel/integraciones/arca/probar', { method: 'POST' });
		prueba = r.ok
			? ((await r.json()) as ArcaPrueba)
			: { ok: false, ultimo: 0, detalle: 'No pudimos hablar con el servidor. Probá de nuevo.' };
		ocupado = false;
	}

	async function alCobrar(v: boolean) {
		const r = await pedir('PATCH', { facturarAlCobrar: v });
		if ('error' in r) return ui.toast({ k: 'err', title: 'No se guardó', desc: String(r.error) });
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: v ? 'Facturar al cobrar, prendido' : 'Facturar al cobrar, apagado',
			desc: v ? 'Cada venta cobrada sale facturada' : 'Las ventas quedan sin facturar',
			action: 'Deshacer',
			onAction: async () => {
				await pedir('PATCH', { facturarAlCobrar: !v });
				await invalidateAll();
			}
		});
	}

	async function guardarPv() {
		const n = Number(pvAjuste);
		if (!Number.isInteger(n) || n < 1 || n > 99998)
			return ui.toast({ k: 'warn', title: 'Punto de venta', desc: 'Va de 1 a 99998' });
		const r = await pedir('PATCH', { puntoVenta: n });
		if ('error' in r) return ui.toast({ k: 'err', title: 'No se guardó', desc: String(r.error) });
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Punto de venta guardado', desc: `Las próximas salen con el ${n}` });
	}

	const NB = '\u00a0';
	// En la demo, "es una tienda de ejemplo" no es un error: va con el estilo de las notas.
	const aviso = (e: string) => (demo && e.includes('tienda de ejemplo') ? 'm-cuenta' : 'm-error');
	const PEM_EJEMPLO = '-----BEGIN CERTIFICATE-----\n…\n-----END CERTIFICATE-----';

	const PASOS_ARCA = {
		homologacion: [
			'Entrá a arca.gob.ar con tu clave fiscal. Si no ves «WSASS – Autoservicio de Acceso a APIs de Homologación», sumalo desde el Administrador de Relaciones de Clave Fiscal › Adherir servicio.',
			'En WSASS tocá «Nuevo certificado», poné mostrador como nombre (es el alias) y pegá el contenido del archivo .csr que bajaste.',
			'Tocá «Crear DN y obtener certificado» y guardá el texto que aparece como un archivo .crt, o copialo para pegarlo acá.',
			'En «Crear autorización a servicio» elegí el alias mostrador, tu CUIT y el servicio wsfe (Facturación electrónica). Confirmá y volvé acá.'
		],
		produccion: [
			'Entrá a arca.gob.ar con tu clave fiscal (nivel 3) y abrí «Administración de Certificados Digitales». Si no está, sumalo desde el Administrador de Relaciones de Clave Fiscal › Adherir servicio.',
			'Tocá «Agregar alias», poné mostrador, subí el archivo .csr y confirmá con «Agregar alias».',
			'Entrá al alias con «Ver» y descargá el certificado: es el archivo .crt.',
			'En el «Administrador de Relaciones de Clave Fiscal» tocá «Nueva relación», buscá ARCA › WebServices › Facturación electrónica y, en Representante, elegí el computador fiscal mostrador. Confirmá y volvé acá.'
		]
	} as const;
</script>

<Modal bind:open titulo="Configurar ARCA" ancho={520} padding={24}>
	<div class="pasos" aria-hidden="true">
		{#each TITULOS as t, i (t)}<span class:on={i <= paso}></span>{/each}
	</div>
	<div class="m-cab">
		<span class="logo grande" style:background={m.bg}
			><span class="sigla" style:color={m.col}>{m.sigla}</span></span
		>
		<div>
			<div class="m-t">ARCA</div>
			<div class="m-st" class:apagado={!arca.conectada}>
				{arca.conectada
					? `● Conectada en ${arca.ambiente === 'produccion' ? 'producción' : 'homologación'}`
					: 'Factura electrónica sin conectar'}
			</div>
		</div>
	</div>
	<div class="m-sec">Paso {paso + 1} de 4 · {TITULOS[paso]}</div>

	{#if paso === 0}
		<form onsubmit={generar}>
			<div class="m-campos">
				<label class="field"
					>CUIT<input
						class="input"
						value={cuit}
						oninput={(e) => (cuit = formatoCuit(e.currentTarget.value))}
						inputmode="numeric"
						autocomplete="off"
						placeholder="20-12345678-9"
						maxlength="13"
						aria-invalid={limpiarCuit(cuit).length === 11 && !cuitValido(limpiarCuit(cuit))}
					/></label
				>
				{#if limpiarCuit(cuit).length === 11 && !cuitValido(limpiarCuit(cuit))}
					<div class="m-error" role="alert">El dígito verificador no da: revisá el&nbsp;CUIT.</div>
				{/if}
				<label class="field"
					>Razón social<input
						class="input"
						bind:value={razon}
						maxlength="64"
						autocomplete="organization"
						placeholder="Como figura en ARCA"
					/></label
				>
				<div class="field">
					Condición frente al IVA
					<Segmentado
						opciones={[
							['monotributo', 'Monotributo'],
							['responsable_inscripto', 'Responsable inscripto']
						]}
						bind:value={condicion}
						label="Condición frente al IVA"
					/>
					<span class="ayuda"
						>{condicion === 'monotributo'
							? 'Emitís factura C, sin IVA discriminado.'
							: 'Emitís factura A a otros inscriptos y B a consumidores finales.'}</span
					>
				</div>
				<label class="field"
					>Punto de venta<input
						class="input"
						bind:value={pv}
						inputmode="numeric"
						maxlength="5"
						placeholder="Por ejemplo: 4"
					/>
					<span class="ayuda"
						>Tiene que ser de tipo «Factura electrónica – Web Services». Lo creás en ARCA ›
						Administración de puntos de venta y&nbsp;domicilios.</span
					></label
				>
				<div class="field">
					Ambiente
					<Segmentado
						opciones={[
							['homologacion', 'Homologación'],
							['produccion', 'Producción']
						]}
						bind:value={ambiente}
						label="Ambiente"
					/>
				</div>
			</div>
			<div class="m-cuenta">
				{#if ambiente === 'homologacion'}
					<b>Homologación</b> es el ambiente de prueba de ARCA: las facturas no tienen validez fiscal
					y el certificado sale en un minuto. Ideal para ver que todo ande antes de&nbsp;facturar.
				{:else}
					<b>Producción</b> emite facturas reales, con CAE válido. El certificado se pide con tu clave
					fiscal y cada venta cobrada sale facturada a tu&nbsp;nombre.
				{/if}
			</div>
			{#if arca.conectada && !demo}
				<div class="m-aviso" role="note">
					<Icon icon={I.phWarningCircleFill} size={18} />
					<span
						>Si generás un pedido nuevo, el certificado actual deja de servir hasta que cargues el
						que te dé&nbsp;ARCA.</span
					>
				</div>
			{/if}
			{#if error}<div class={aviso(error)} role="alert">{error}</div>{/if}
			<div class="m-acc">
				{#if arca.conectada}
					<button type="button" class="b-sec" onclick={() => ((paso = 3), (error = null))}
						>Volver</button
					>
				{:else}
					<button type="button" class="b-sec" onclick={() => (open = false)}>Cancelar</button>
				{/if}
				<button class="b-pri" disabled={ocupado}>{ocupado ? 'Generando…' : 'Generar pedido'}</button
				>
			</div>
		</form>
	{:else if paso === 1}
		<div class="m-d">
			Descargá tu pedido de certificado y subilo en ARCA. Lleva tu CUIT y tu razón social; la clave
			{demo
				? 'se generó en tu navegador y no se guarda.'
				: 'queda guardada y cifrada en Mostrador.'}
		</div>
		<button type="button" class="b-pri descargar" onclick={descargar}
			><Icon icon={I.phDownloadSimpleBold} size={17} />Descargar pedido (.csr)</button
		>
		{#if demo}
			<div class="m-cuenta">
				Es una tienda de ejemplo: el pedido es real, pero como la clave no se guarda, el certificado
				que te dé ARCA no se puede cargar&nbsp;acá.
			</div>
		{/if}
		<div class="m-sec">
			En ARCA · {ambiente === 'produccion' ? 'producción' : 'homologación (WSASS)'}
		</div>
		<ol class="m-pasos">
			{#each PASOS_ARCA[ambiente] as p, i (i)}
				<li><span class="n">{i + 1}</span>{p}</li>
			{/each}
		</ol>
		<div class="m-acc">
			<button type="button" class="b-sec" onclick={() => (paso = 0)}>Atrás</button>
			<button type="button" class="b-pri" onclick={() => ((paso = 2), (error = null))}
				>Ya tengo el certificado</button
			>
		</div>
	{:else if paso === 2}
		<form onsubmit={guardarCertificado}>
			<div class="m-d">
				Pegá el certificado que te dio ARCA o subí el archivo .crt. Revisamos que sea del pedido que
				bajaste y de tu&nbsp;CUIT.
			</div>
			<label class="field pegar"
				>Certificado<textarea
					class="textarea"
					bind:value={pem}
					rows="6"
					spellcheck="false"
					placeholder={PEM_EJEMPLO}></textarea></label
			>
			<label class="b-sec archivo"
				><Icon icon={I.phFileTextFill} size={16} />Subir archivo .crt<input
					type="file"
					accept=".crt,.pem,.cer,application/x-x509-ca-cert"
					onchange={leerArchivo}
				/></label
			>
			{#if error}<div class={aviso(error)} role="alert">{error}</div>{/if}
			<div class="m-acc">
				<button type="button" class="b-sec" onclick={() => ((paso = 1), (error = null))}
					>Atrás</button
				>
				<button class="b-pri" disabled={ocupado}
					>{ocupado ? 'Revisando…' : 'Guardar certificado'}</button
				>
			</div>
		</form>
	{:else}
		{#if diasParaVencer !== null && diasParaVencer < 30}
			<div class="m-aviso" role="alert">
				<Icon icon={I.phWarningCircleFill} size={18} />
				<span
					>{diasParaVencer <= 0
						? 'El certificado venció: las ventas quedan sin facturar hasta que cargues uno nuevo.'
						: `El certificado vence en ${diasParaVencer} ${diasParaVencer === 1 ? 'día' : 'días'}. Renovalo antes: generá un pedido nuevo y cargá el certificado que te${NB}dé${NB}ARCA.`}</span
				>
			</div>
		{/if}
		<div class="m-cuenta">
			CUIT <b>{formatoCuit(arca.cuit)}</b> · {arca.razonSocial} · {arca.condicion ===
			'responsable_inscripto'
				? 'Responsable inscripto'
				: 'Monotributo'}.
			{#if vence !== null}Certificado válido hasta el <b>{fecha(vence)}</b>.{/if}
		</div>
		<div class="m-sec">Conexión</div>
		<div class="m-ops">
			<div class="m-op">
				<span class="g-txt"
					><span class="g-t">Probar conexión</span><span class="g-d"
						>Pide el acceso a ARCA y el último número&nbsp;autorizado</span
					></span
				>
				<button type="button" class="t-btn" disabled={ocupado} onclick={probar}
					>{ocupado ? 'Probando…' : 'Probar'}</button
				>
			</div>
		</div>
		{#if prueba}
			<div class="resultado" class:ok={prueba.ok} role="status">
				<Icon icon={prueba.ok ? I.phCheckCircleFill : I.phInfoFill} size={18} />
				<span>{prueba.detalle}</span>
			</div>
		{/if}
		<div class="m-sec">Ajustes</div>
		<div class="m-ops">
			<div class="m-op">
				<span class="g-txt"
					><span class="g-t">Punto de venta</span><span class="g-d"
						>El de «Factura electrónica – Web&nbsp;Services»</span
					></span
				>
				<input
					class="input pv"
					bind:value={pvAjuste}
					inputmode="numeric"
					maxlength="5"
					aria-label="Punto de venta"
				/>
				<button
					type="button"
					class="t-btn on"
					disabled={pvAjuste === String(arca.puntoVenta)}
					onclick={guardarPv}>Guardar</button
				>
			</div>
			<div class="m-op">
				<span class="g-txt"
					><span class="g-t">Facturar al cobrar</span><span class="g-d"
						>Si está apagado, facturás a&nbsp;mano</span
					></span
				>
				<Interruptor
					label="Facturar al cobrar"
					checked={arca.facturarAlCobrar}
					onchange={(v) => void alCobrar(v)}
				/>
			</div>
		</div>
		<div class="m-acc">
			<button type="button" class="b-sec" onclick={() => ((paso = 0), (error = null))}
				>Renovar el certificado</button
			>
			<button type="button" class="b-pri" onclick={() => (open = false)}>Listo</button>
		</div>
	{/if}
</Modal>

<style>
	/* Piezas del modal de Integraciones (+page.svelte) y la barra de pasos de Plan y facturación. */
	.pasos {
		display: flex;
		gap: 6px;
		margin-bottom: 16px;
	}
	.pasos span {
		flex: 1;
		height: 4px;
		border-radius: 2px;
		background: var(--ln2);
	}
	.pasos span.on {
		background: #d2582f;
	}
	.m-cab {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.logo {
		width: 52px;
		height: 52px;
		flex: none;
		border-radius: 16px;
		display: grid;
		place-items: center;
	}
	.sigla {
		font-size: 8.5px;
		font-weight: 800;
	}
	.m-t {
		font-size: 19px;
		font-weight: 600;
	}
	.m-st {
		font-size: 13px;
		font-weight: 600;
		color: #2f8f63;
	}
	.m-st.apagado {
		color: var(--fa);
	}
	.m-sec {
		margin-top: 16px;
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--fa);
	}
	.m-d {
		margin-top: 8px;
		font-size: 14px;
		line-height: 1.5;
		color: var(--mu);
	}
	.m-campos {
		margin-top: 10px;
		display: grid;
		gap: 12px;
	}
	.ayuda {
		font-size: 12px;
		font-weight: 500;
		line-height: 1.45;
		color: var(--fa);
	}
	.m-cuenta {
		margin-top: 12px;
		padding: 12px 14px;
		border-radius: 14px;
		background: var(--sf2);
		font-size: 13px;
		line-height: 1.5;
		color: var(--mu);
	}
	.m-cuenta b {
		color: var(--tx);
	}
	.m-aviso {
		margin-top: 14px;
		display: flex;
		gap: 8px;
		align-items: flex-start;
		padding: 12px 14px;
		border-radius: 14px;
		background: rgba(227, 169, 59, 0.16);
		color: #9a6a10;
		font-size: 13.5px;
		line-height: 1.45;
	}
	.m-aviso :global(svg) {
		flex: none;
		margin-top: 1px;
	}
	.m-error {
		margin-top: 10px;
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	.m-campos .m-error {
		margin-top: -6px;
	}
	.m-pasos {
		margin: 8px 0 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.m-pasos li {
		display: flex;
		gap: 10px;
		align-items: flex-start;
		font-size: 14px;
		line-height: 1.45;
	}
	.m-pasos .n {
		width: 22px;
		height: 22px;
		flex: none;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: var(--tx);
		color: var(--sf);
		font-size: 12px;
		font-weight: 700;
	}
	.m-ops {
		margin-top: 4px;
		display: flex;
		flex-direction: column;
	}
	.m-op {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.m-op:last-child {
		border-bottom: 0;
	}
	.g-txt {
		flex: 1;
		min-width: 0;
	}
	.g-t {
		display: block;
		font-size: 13.5px;
		font-weight: 600;
		color: var(--tx);
	}
	.g-d {
		display: block;
		font-size: 12px;
		color: var(--fa);
	}
	.t-btn {
		min-height: 44px;
		padding: 8px 14px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
		white-space: nowrap;
	}
	.t-btn.on {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.t-btn:disabled {
		opacity: 0.5;
		cursor: default;
	}
	.input.pv {
		width: 76px;
		flex: none;
		text-align: center;
	}
	.resultado {
		margin-top: 8px;
		display: flex;
		gap: 8px;
		align-items: flex-start;
		padding: 12px 14px;
		border-radius: 14px;
		background: var(--sf2);
		color: var(--mu);
		font-size: 13.5px;
		line-height: 1.45;
	}
	.resultado.ok {
		background: rgba(47, 143, 99, 0.1);
		color: #2f8f63;
	}
	.resultado :global(svg) {
		flex: none;
		margin-top: 1px;
	}
	.pegar {
		margin-top: 12px;
	}
	.pegar textarea {
		font:
			12px/1.4 ui-monospace,
			monospace;
		resize: vertical;
	}
	.m-acc {
		margin-top: 18px;
		display: flex;
		gap: 8px;
	}
	.m-acc button {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.b-sec {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.b-pri {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.b-pri:disabled {
		opacity: 0.6;
	}
	.descargar,
	.archivo {
		margin-top: 12px;
		width: 100%;
		min-height: 48px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		border-radius: 999px;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.archivo {
		position: relative;
	}
	.archivo input {
		position: absolute;
		width: 1px;
		height: 1px;
		opacity: 0;
		pointer-events: none;
	}
</style>
