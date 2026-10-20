<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { BENEFICIOS_CAJA, QR, RESUMEN } from './datos.ts';
	import { getOnboarding } from './estado.svelte.ts';

	/**
	 * Paso 8, ¡Listo!: confeti, lo que quedó armado, lo que falta, Caja, el QR y el equipo. "Ir a mi
	 * panel" manda las elecciones al servidor (form del padre) y entra al panel.
	 */
	let { terminar }: { terminar: () => void } = $props();
	const o = getOnboarding();
	let copiado = $state(false);
	let invitacion = $state('');

	function confeti(host: HTMLElement) {
		if (matchMedia('(prefers-reduced-motion: reduce)').matches) return;
		const cols = ['#D2582F', '#E8875F', '#E3A93B', '#2F8F63', '#2A140C'];
		const t = setTimeout(() => {
			for (let i = 0; i < 70; i++) {
				const p = document.createElement('span');
				const w = 6 + Math.random() * 6;
				p.style.cssText = `position:absolute;left:0;top:0;width:${w}px;height:${w * 0.45}px;border-radius:2px;background:${cols[i % 5]}`;
				host.appendChild(p);
				const a = Math.random() * Math.PI * 2;
				const d = 160 + Math.random() * 260;
				p.animate(
					[
						{ transform: 'translate(0,0) rotate(0)', opacity: 1 },
						{
							transform: `translate(${Math.cos(a) * d}px,${Math.sin(a) * d * 0.6 + 260}px) rotate(${Math.random() * 720}deg)`,
							opacity: 0
						}
					],
					{
						duration: 1600 + Math.random() * 1200,
						easing: 'cubic-bezier(.15,.6,.3,1)',
						fill: 'forwards'
					}
				).onfinish = () => p.remove();
			}
		}, 300);
		return { destroy: () => clearTimeout(t) };
	}
	async function copiar() {
		try {
			await navigator.clipboard.writeText('https://laesquina.com.ar');
		} catch {
			// Sin permiso de portapapeles: igual se muestra el aviso, el link está a la vista.
		}
		copiado = true;
		setTimeout(() => (copiado = false), 1800);
	}
	const iniciales = (e: string) => e.slice(0, 2).toUpperCase();
</script>

<div class="todo">
	<div class="confeti" use:confeti></div>
	<div class="anillo">
		<svg viewBox="0 0 120 120" aria-hidden="true"
			><circle cx="60" cy="60" r="52" fill="none" stroke="#F1E7DC" stroke-width="10"
			></circle><circle
				cx="60"
				cy="60"
				r="52"
				fill="none"
				stroke="#D2582F"
				stroke-width="10"
				stroke-linecap="round"
				stroke-dasharray="326.7"
				stroke-dashoffset="19.6"
			></circle></svg
		>
		<div class="ave"><img src="/hornero/celebrando.png" alt="Hornero celebrando" /></div>
	</div>
	<h1>Tu mostrador está 94% listo.</h1>
	<p class="bajada">Y no tuviste que cargar nada a mano. Ya podés vender por todos tus canales.</p>
	<div class="resumen">
		{#each RESUMEN as [v, t] (t)}<div>
				<div class="r-v">{v}</div>
				<div class="r-t">{t}</div>
			</div>{/each}
	</div>
	<div class="faltan">
		<div class="f-t">Te quedan 2 cosas, cuando quieras</div>
		<div class="f-fila">
			<span class="arca">ARCA</span><span class="f-x"
				>Subir el certificado para facturar <span>· 2 minutos</span></span
			><button type="button" class="f-link" onclick={terminar}>Hacerlo</button>
		</div>
		<div class="f-fila">
			<span class="corona"><Icon icon={I.phCrownSimpleFill} size={15} /></span><span class="f-x"
				>Elegir tu plan <span>· te quedan 14 días gratis</span></span
			>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- ancla de la landing -->
			<a class="f-link" href="/#planes">Ver planes</a>
		</div>
	</div>
	<div class="caja">
		<div class="cj-cab">
			<img src="/caja/05-exito.png" alt="Caja" width="56" height="56" />
			<div class="cj-txt">
				<div class="cj-t">
					{o.e.caja
						? 'Caja ya está ordenando tu plata'
						: 'Tu negocio vende. Caja te dice cuánto ganás.'}
				</div>
				<div class="cj-d">
					{o.e.caja
						? 'Conectado · 3 meses de Pro Max activos · las ventas de hoy ya están ahí'
						: 'Con tu cuenta de Mostrador tenés 3 meses de Caja Pro Max gratis'}
				</div>
			</div>
			{#if o.e.caja}
				<a class="cj-btn" href="https://caja.iamnaw.dev" target="_blank" rel="noopener"
					>Abrir Caja<Icon icon={I.phArrowUpRightBold} size={13} /></a
				>
			{:else}
				<button type="button" class="cj-btn" onclick={() => ((o.e.caja = true), o.guardar())}
					><Icon icon={I.phPlugFill} size={14} />Conectar Caja</button
				>
			{/if}
		</div>
		<div class="cj-ben">
			{#each BENEFICIOS_CAJA as [ico, t, d] (t)}
				<div>
					<Icon icon={ico} size={20} style="color:#E8875F;flex:none;margin-top:1px" />
					<div>
						<div class="b-t">{t}</div>
						<div class="b-d">{d}</div>
					</div>
				</div>
			{/each}
		</div>
	</div>
	<div class="dos">
		<div class="online">
			<div class="qr" aria-hidden="true">
				{#each QR as q, i (i)}<span class:lleno={q}></span>{/each}
			</div>
			<div class="on-txt">
				<div class="on-t">Tu tienda ya está en línea</div>
				<div class="on-d">laesquina.com.ar · imprimí el QR para el mostrador</div>
				<div class="on-acc">
					<button type="button" onclick={copiar}
						><Icon icon={copiado ? I.phCheckBold : I.phCopyBold} size={13} />{copiado
							? 'Copiado'
							: 'Copiar link'}</button
					>
					<a
						class="wa"
						href="https://wa.me/?text={encodeURIComponent(
							'Mirá mi tienda: https://laesquina.com.ar'
						)}"
						target="_blank"
						rel="noopener"><Icon icon={I.phWhatsappLogoFill} size={13} />Compartir</a
					>
				</div>
			</div>
		</div>
		<div class="equipo">
			<div class="eq-t">Sumá a tu equipo</div>
			<div class="eq-d">Cada uno con sus permisos y doble verificación.</div>
			<div class="eq-l">
				{#each o.e.equipo as m (m.email)}
					<div>
						<span class="ini">{iniciales(m.email)}</span><span class="mail">{m.email}</span><span
							class="rol">{m.rol}</span
						>
					</div>
				{/each}
			</div>
			<form
				class="invitar"
				onsubmit={(ev) => {
					ev.preventDefault();
					o.invitar(invitacion);
					invitacion = '';
				}}
			>
				<input
					type="email"
					bind:value={invitacion}
					placeholder="email de tu socio o empleado"
					aria-label="Email de tu socio o empleado"
				/>
				<button type="submit">Invitar</button>
			</form>
		</div>
	</div>
	<div class="ctas">
		<button type="button" class="pri" onclick={terminar}
			>Ir a mi panel <Icon icon={I.phArrowRightBold} size={16} /></button
		>
		<button type="button" class="sec" onclick={terminar}>Ver mi tienda</button>
	</div>
</div>

<style>
	.todo {
		position: relative;
		text-align: center;
		max-width: 760px;
		margin: 0 auto;
	}
	.confeti {
		position: absolute;
		left: 50%;
		top: 60px;
		width: 0;
		height: 0;
	}
	.anillo {
		position: relative;
		width: 200px;
		height: 200px;
		margin: 0 auto;
	}
	.anillo svg {
		width: 100%;
		height: 100%;
		transform: rotate(-90deg);
	}
	.ave {
		position: absolute;
		inset: 22px;
		border-radius: 50%;
		overflow: hidden;
		background: #fcede5;
	}
	.ave img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: contain;
	}
	h1 {
		margin: 22px 0 0;
		font-size: clamp(34px, 4.2vw, 56px);
		line-height: 1;
		letter-spacing: -0.045em;
		font-weight: 600;
	}
	.bajada {
		margin: 14px auto 0;
		font-size: 17px;
		line-height: 1.55;
		color: #5c4536;
		max-width: 520px;
	}
	.resumen {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 10px;
	}
	.resumen > div {
		padding: 16px 12px;
		border-radius: 18px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.r-v {
		font-size: 28px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.r-t {
		font-size: 13px;
		color: #7a6252;
		margin-top: 2px;
	}
	.faltan {
		margin-top: 20px;
		padding: 16px 18px;
		border-radius: 20px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		text-align: left;
	}
	.f-t {
		font-size: 14.5px;
		font-weight: 600;
	}
	.f-fila {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-top: 1px solid rgba(58, 29, 18, 0.07);
	}
	.f-t + .f-fila {
		margin-top: 8px;
	}
	.arca,
	.corona {
		width: 30px;
		height: 30px;
		flex: none;
		border-radius: 9px;
		display: grid;
		place-items: center;
	}
	.arca {
		background: #1f3a6e;
		font-size: 7px;
		font-weight: 800;
		color: #fff;
	}
	.corona {
		background: #fcede5;
		color: #c24e25;
	}
	.f-x {
		flex: 1;
		font-size: 14px;
	}
	.f-x span {
		color: #7a6252;
	}
	.f-link {
		font-size: 13.5px;
		font-weight: 600;
		color: #c24e25;
		background: none;
		border: 0;
		padding: 0;
	}
	.caja {
		margin-top: 14px;
		padding: 20px;
		border-radius: 24px;
		background: #131211;
		color: #f4eee6;
		text-align: left;
	}
	.cj-cab {
		display: flex;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
	}
	.cj-cab img {
		object-fit: contain;
		flex: none;
	}
	.cj-txt {
		flex: 1 1 260px;
		min-width: 0;
	}
	.cj-t {
		font-size: 17px;
		font-weight: 600;
	}
	.cj-d {
		font-size: 13px;
		color: #a99e94;
		margin-top: 2px;
	}
	.cj-btn {
		flex: none;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border-radius: 999px;
		border: 0;
		background: #f4eee6;
		color: #131211;
		font-size: 13.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.cj-ben {
		margin-top: 16px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 10px;
	}
	.cj-ben > div {
		display: flex;
		gap: 10px;
		padding: 12px;
		border-radius: 16px;
		background: #1f1c1a;
	}
	.b-t {
		font-size: 14px;
		font-weight: 600;
	}
	.b-d {
		font-size: 12.5px;
		color: #9c9188;
		margin-top: 2px;
		line-height: 1.4;
	}
	.dos {
		margin-top: 14px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr));
		gap: 12px;
		text-align: left;
	}
	.online {
		padding: 16px;
		border-radius: 20px;
		background: #2a140c;
		color: #fbf7f2;
		display: flex;
		gap: 14px;
		align-items: center;
	}
	.qr {
		width: 84px;
		height: 84px;
		flex: none;
		border-radius: 14px;
		background: #fbf7f2;
		padding: 8px;
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 2px;
	}
	.qr span {
		border-radius: 1.5px;
	}
	.qr .lleno {
		background: #2a140c;
	}
	.on-txt {
		flex: 1;
		min-width: 0;
	}
	.on-t {
		font-size: 15px;
		font-weight: 600;
	}
	.on-d {
		font-size: 12.5px;
		color: #d9c8ba;
		margin-top: 2px;
	}
	.on-acc {
		margin-top: 10px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.on-acc button,
	.on-acc a {
		display: flex;
		align-items: center;
		gap: 5px;
		padding: 7px 12px;
		border-radius: 999px;
		border: 0;
		background: rgba(251, 247, 242, 0.12);
		color: #fff;
		font-size: 12.5px;
		font-weight: 600;
	}
	.on-acc .wa {
		background: #25d366;
	}
	.equipo {
		padding: 16px;
		border-radius: 20px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.eq-t {
		font-size: 15px;
		font-weight: 600;
	}
	.eq-d {
		font-size: 12.5px;
		color: #7a6252;
		margin-top: 2px;
	}
	.eq-l {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.eq-l > div {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 13.5px;
	}
	.ini {
		width: 28px;
		height: 28px;
		flex: none;
		border-radius: 50%;
		background: #f1e7dc;
		display: grid;
		place-items: center;
		font-size: 11px;
		font-weight: 700;
		color: #7a6252;
	}
	.mail {
		flex: 1;
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.rol {
		font-size: 12px;
		font-weight: 600;
		padding: 3px 9px;
		border-radius: 999px;
		background: #f4ece2;
		color: #5c4536;
	}
	.invitar {
		margin-top: 10px;
		display: flex;
		gap: 6px;
	}
	.invitar input {
		flex: 1;
		min-width: 0;
		padding: 9px 12px;
		border-radius: 12px;
		border: 1px solid rgba(58, 29, 18, 0.16);
		font-size: 13.5px;
		font-weight: 500;
		outline: none;
		color: #2a140c;
		background: #fff;
	}
	.invitar button {
		padding: 9px 14px;
		border-radius: 12px;
		border: 0;
		background: #2a140c;
		color: #fff;
		font-size: 13px;
		font-weight: 600;
	}
	.ctas {
		margin-top: 24px;
		display: flex;
		gap: 10px;
		justify-content: center;
		flex-wrap: wrap;
	}
	.ctas button {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 16px;
		font-weight: 600;
		border-radius: 999px;
	}
	.pri {
		color: #fff;
		border: 0;
		background: linear-gradient(180deg, #e06a3e, #d2582f);
		padding: 16px 26px;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			0 16px 34px -14px rgba(210, 88, 47, 0.75);
	}
	.sec {
		color: #2a140c;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.14);
		padding: 16px 24px;
	}
	@media (max-width: 560px) {
		.resumen {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
