<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import { marca } from '#lib/marcas.ts';
	import type { ConexionCorreo } from '#lib/server/correo-argentino.ts';
	import { PROVINCIAS, nombreProvincia } from '#lib/tienda/provincias.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Integraciones › Correo Argentino. El handoff no diseña la conexión: se arma con las piezas del
	 * modal de Integraciones y del asistente de ARCA (cabecera, secciones, campos, notas y acciones).
	 * Con la cuenta de MiCorreo de la tienda: email y contraseña (sólo para validar), el origen y el
	 * paquete típico; las credenciales de API, si Tu Local Web no las tiene o la tienda usa las suyas.
	 */
	let {
		open = $bindable(false),
		correo,
		demo
	}: { open?: boolean; correo: ConexionCorreo | null; demo: boolean } = $props();
	const ui = getUi();
	const m = marca('correo');

	let email = $state('');
	let clave = $state('');
	let propias = $state(false);
	let apiUsuario = $state('');
	let apiClave = $state('');
	let cp = $state('');
	let provincia = $state('');
	let localidad = $state('');
	let calle = $state('');
	let altura = $state('');
	let piso = $state('');
	let depto = $state('');
	let peso = $state('');
	let alto = $state('');
	let ancho = $state('');
	let largo = $state('');
	let error = $state<string | null>(null);
	let ocupado = $state(false);

	const conectada = $derived(!!correo?.conectada);
	const pideApi = $derived(!!correo?.pideCredenciales || propias);

	// Al abrir, lo que ya está guardado (o vacío para conectar).
	let abierto = false;
	$effect(() => {
		if (open && !abierto) {
			const o = correo?.origen;
			const p = correo?.paquete;
			email = correo?.email ?? '';
			clave = apiUsuario = apiClave = '';
			propias = false;
			cp = o?.cp ?? '';
			provincia = o?.provincia ?? '';
			localidad = o?.localidad ?? '';
			calle = o?.calle ?? '';
			altura = o?.altura ?? '';
			piso = o?.piso ?? '';
			depto = o?.depto ?? '';
			peso = p ? String(p.peso) : '';
			alto = p ? String(p.alto) : '';
			ancho = p ? String(p.ancho) : '';
			largo = p ? String(p.largo) : '';
			error = null;
		}
		abierto = open;
	});

	const entero = (s: string) => (/^\d+$/.test(s.trim()) ? Number(s.trim()) : NaN);

	/** Valida lo mismo que Go antes de mandar; devuelve el error o el origen y el paquete. */
	function armar() {
		if (!/^\d{4}$/.test(cp.trim())) return 'El código postal de origen tiene 4 números.';
		if (!provincia) return 'Elegí la provincia de origen.';
		if (!localidad.trim() || !calle.trim() || !altura.trim())
			return 'Completá la dirección de donde salen los paquetes.';
		const pk = {
			peso: entero(peso),
			alto: entero(alto),
			ancho: entero(ancho),
			largo: entero(largo)
		};
		if (!(pk.peso >= 1 && pk.peso <= 25000)) return 'El peso por unidad va de 1 a 25.000 gramos.';
		if (![pk.alto, pk.ancho, pk.largo].every((l) => l >= 1 && l <= 150))
			return 'Cada medida de la caja va de 1 a 150 cm.';
		return {
			origen: {
				cp: cp.trim(),
				provincia,
				localidad: localidad.trim(),
				calle: calle.trim(),
				altura: altura.trim(),
				piso: piso.trim(),
				depto: depto.trim()
			},
			paquete: pk
		};
	}

	async function enviar(metodo: 'POST' | 'PUT' | 'DELETE', cuerpo?: unknown) {
		const r = await fetch('/panel/integraciones/correo', {
			method: metodo,
			headers: { 'content-type': 'application/json' },
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!r.ok) {
			const msg = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return msg ?? 'No pudimos hablar con el servidor. Probá de nuevo.';
		}
		return null;
	}

	async function conectar(e: SubmitEvent) {
		e.preventDefault();
		error = null;
		if (!email.trim() || !clave) return (error = 'Poné el email y la contraseña de MiCorreo.');
		if (pideApi && (!apiUsuario.trim() || !apiClave))
			return (error = 'Completá el usuario y la contraseña de la API.');
		const d = armar();
		if (typeof d === 'string') return (error = d);
		ocupado = true;
		try {
			error = await enviar('POST', {
				email: email.trim(),
				clave,
				api: pideApi ? { usuario: apiUsuario.trim(), clave: apiClave } : null,
				...d
			});
		} catch {
			error = 'No pudimos conectarnos. Revisá la conexión y probá de nuevo.';
		} finally {
			ocupado = false;
		}
		if (error) return;
		clave = apiClave = '';
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Correo Argentino conectado', desc: 'Ya cotiza en tu tienda' });
	}

	async function guardar(e: SubmitEvent) {
		e.preventDefault();
		error = null;
		const d = armar();
		if (typeof d === 'string') return (error = d);
		ocupado = true;
		try {
			error = await enviar('PUT', d);
		} catch {
			error = 'No pudimos conectarnos. Revisá la conexión y probá de nuevo.';
		} finally {
			ocupado = false;
		}
		if (error) return;
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Envíos guardados', desc: 'Las próximas cotizaciones los usan' });
	}

	async function desconectar() {
		const e = await enviar('DELETE').catch(() => 'No pudimos conectarnos. Probá de nuevo.');
		if (e) return ui.toast({ k: 'err', title: 'No se pudo', desc: e });
		open = false;
		await invalidateAll();
		ui.toast({
			k: 'info',
			title: 'Correo Argentino desconectado',
			desc: 'La tienda deja de ofrecerlo hasta que lo conectes'
		});
	}

	const fecha = (iso: string) =>
		new Date(iso).toLocaleDateString('es-AR', {
			day: 'numeric',
			month: 'long',
			timeZone: 'America/Argentina/Buenos_Aires'
		});
	// En la demo, "es una tienda de ejemplo" no es un error: va con el estilo de las notas.
	const aviso = (e: string) => (demo && e.includes('tienda de ejemplo') ? 'm-cuenta' : 'm-error');
</script>

{#snippet envios()}
	<div class="m-sec">Desde dónde despachás</div>
	<div class="m-campos dos">
		<label class="field"
			>Código postal<input
				class="input"
				bind:value={cp}
				inputmode="numeric"
				maxlength="4"
				autocomplete="postal-code"
			/></label
		>
		<label class="field"
			>Provincia<select class="input" bind:value={provincia}>
				<option value="" disabled>Elegí</option>
				{#each PROVINCIAS as [c, n] (c)}<option value={c}>{n}</option>{/each}
			</select></label
		>
		<label class="field ancho"
			>Localidad<input
				class="input"
				bind:value={localidad}
				maxlength="80"
				autocomplete="address-level2"
			/></label
		>
		<label class="field ancho"
			>Calle<input
				class="input"
				bind:value={calle}
				maxlength="80"
				autocomplete="address-line1"
			/></label
		>
		<label class="field">Altura<input class="input" bind:value={altura} maxlength="10" /></label>
		<div class="pd">
			<label class="field">Piso<input class="input" bind:value={piso} maxlength="3" /></label>
			<label class="field">Depto<input class="input" bind:value={depto} maxlength="3" /></label>
		</div>
	</div>
	<div class="m-sec">Paquete típico</div>
	<div class="m-d">
		Correo cotiza por peso y medidas. Poné lo que pesa un producto embalado y la caja que más usás:
		el peso se multiplica por las unidades del&nbsp;pedido.
	</div>
	<div class="m-campos cuatro">
		<label class="field"
			>Peso (g)<input class="input" bind:value={peso} inputmode="numeric" maxlength="5" /></label
		>
		<label class="field"
			>Alto (cm)<input class="input" bind:value={alto} inputmode="numeric" maxlength="3" /></label
		>
		<label class="field"
			>Ancho (cm)<input class="input" bind:value={ancho} inputmode="numeric" maxlength="3" /></label
		>
		<label class="field"
			>Largo (cm)<input class="input" bind:value={largo} inputmode="numeric" maxlength="3" /></label
		>
	</div>
{/snippet}

<Modal bind:open titulo="Correo Argentino" ancho={520} padding={24}>
	<div class="m-cab">
		<span class="logo grande" style:background={m.bg}
			>{#if m.icon}<Icon icon={m.icon} size={26} style="color:{m.col ?? 'inherit'}" />{:else}<span
					class="sigla"
					style:color={m.col}>{m.sigla}</span
				>{/if}</span
		>
		<div>
			<div class="m-t">Correo Argentino</div>
			<div class="m-st" class:apagado={!conectada}>
				{#if conectada && correo}● Conectada{#if correo.desde}&nbsp;desde el {fecha(
							correo.desde
						)}{/if}{:else}Sin conectar{/if}
			</div>
		</div>
	</div>

	{#if conectada && correo}
		<div class="m-cuenta">
			Cuenta de MiCorreo <b>{correo.email}</b> (cliente {correo.cliente}){#if correo.propias}, con
				tus credenciales de API{/if}. Sale de {correo.origen?.localidad},
			{nombreProvincia(correo.origen?.provincia ?? '')}.
			{#if correo.pruebas}<br />Está en el ambiente de pruebas de Correo: las cotizaciones no son
				las&nbsp;reales.{/if}
		</div>
		<div class="m-sec">Cómo se usa</div>
		<ol class="m-pasos">
			<li>
				<span class="n">1</span><span
					>Quien compra elige Correo a domicilio o a sucursal, con el precio que cotiza&nbsp;Correo.</span
				>
			</li>
			<li>
				<span class="n">2</span><span
					>En el pedido tocá «Generar envío»: queda cargado en tu&nbsp;MiCorreo.</span
				>
			</li>
			<li>
				<span class="n">3</span><span
					>En MiCorreo lo pagás e imprimís el rótulo. Pegá el número de seguimiento en el pedido y
					quien compra lo&nbsp;ve.</span
				>
			</li>
		</ol>
		<form onsubmit={guardar}>
			{@render envios()}
			{#if error}<div class={aviso(error)} role="alert">{error}</div>{/if}
			<div class="m-acc">
				<button type="button" class="b-sec rojo" onclick={desconectar}>Desconectar</button>
				<button class="b-pri" disabled={ocupado}>{ocupado ? 'Guardando…' : 'Guardar'}</button>
			</div>
		</form>
	{:else}
		<div class="m-d">
			Conectá tu cuenta de MiCorreo y tu tienda cotiza el envío con el precio real de Correo, a
			domicilio y a sucursal. Los envíos se cargan en tu cuenta y los pagás&nbsp;allá.
		</div>
		{#if demo}
			<div class="m-cuenta">
				Es una tienda de ejemplo: Correo Argentino se conecta con una cuenta de MiCorreo real,
				cuando tu tienda esté en&nbsp;marcha.
			</div>
		{/if}
		<div class="m-sec">Dónde se saca</div>
		<ol class="m-pasos">
			<li>
				<span class="n">1</span><span
					>Creá tu cuenta en
					<a href="https://micorreo.correoargentino.com.ar" target="_blank" rel="noopener"
						>micorreo.correoargentino.com.ar</a
					> (con DNI o CUIT) y completá tus datos de&nbsp;remitente.</span
				>
			</li>
			<li>
				<span class="n">2</span><span
					>Usá acá el mismo email y contraseña con los que entrás a MiCorreo. La contraseña sólo
					sirve para validar la cuenta: no se&nbsp;guarda.</span
				>
			</li>
			{#if pideApi}
				<li>
					<span class="n">3</span><span
						>Pedile a Correo Argentino el usuario y la contraseña de la API de MiCorreo (son
						distintos de los de tu cuenta). Se guardan&nbsp;cifrados.</span
					>
				</li>
			{/if}
		</ol>
		<form onsubmit={conectar}>
			<div class="m-sec">Tu cuenta de MiCorreo</div>
			<div class="m-campos">
				<label class="field"
					>Email<input
						class="input"
						type="email"
						bind:value={email}
						maxlength="120"
						autocomplete="username"
					/></label
				>
				<label class="field"
					>Contraseña<input
						class="input"
						type="password"
						bind:value={clave}
						maxlength="200"
						autocomplete="current-password"
					/></label
				>
				{#if !correo?.pideCredenciales}
					<label class="tilde"
						><input type="checkbox" bind:checked={propias} />Tengo credenciales de API propias</label
					>
				{/if}
				{#if pideApi}
					<label class="field"
						>Usuario de la API<input
							class="input"
							bind:value={apiUsuario}
							maxlength="120"
							autocomplete="off"
						/></label
					>
					<label class="field"
						>Contraseña de la API<input
							class="input"
							type="password"
							bind:value={apiClave}
							maxlength="200"
							autocomplete="new-password"
						/></label
					>
				{/if}
			</div>
			{@render envios()}
			{#if error}<div class={aviso(error)} role="alert">{error}</div>{/if}
			<div class="m-acc">
				<button type="button" class="b-sec" onclick={() => (open = false)}>Cancelar</button>
				<button class="b-pri" disabled={ocupado}>{ocupado ? 'Validando…' : 'Conectar'}</button>
			</div>
		</form>
	{/if}
</Modal>

<style>
	/* Piezas del modal de Integraciones y del asistente de ARCA (Arca.svelte). */
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
		color: #237550;
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
		text-wrap: pretty;
	}
	.m-campos {
		margin-top: 10px;
		display: grid;
		gap: 12px;
	}
	.m-campos.dos {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}
	.m-campos.cuatro {
		grid-template-columns: repeat(4, minmax(0, 1fr));
	}
	.m-campos .ancho {
		grid-column: 1 / -1;
	}
	.pd {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 12px;
	}
	@media (max-width: 420px) {
		.m-campos.cuatro {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
	.tilde {
		display: flex;
		align-items: center;
		gap: 8px;
		min-height: 44px;
		font-size: 13.5px;
		color: var(--mu);
		cursor: pointer;
	}
	.m-cuenta {
		margin-top: 12px;
		padding: 12px 14px;
		border-radius: 14px;
		background: var(--sf2);
		font-size: 13px;
		line-height: 1.5;
		color: var(--mu);
		text-wrap: pretty;
	}
	.m-cuenta b {
		color: var(--tx);
	}
	.m-error {
		margin-top: 10px;
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
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
		text-wrap: pretty;
	}
	.m-pasos a {
		color: inherit;
		text-decoration: underline;
		overflow-wrap: anywhere;
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
	.b-sec.rojo {
		color: #b83a2e;
	}
	.b-pri {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.b-pri:disabled {
		opacity: 0.6;
	}
</style>
