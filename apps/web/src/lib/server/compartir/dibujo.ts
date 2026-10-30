/**
 * El dibujo de la imagen al compartir (1200 × 630), como árbol para satori. Tres composiciones, las
 * mismas familias de las plantillas (editorial, grilla y catálogo), cada una con y sin fotos, y el
 * adorno propio de cada plantilla. Sin fotos queda igual de armada: el nombre grande, la inicial en
 * el acento y, si hay, las categorías con cuántos productos tiene cada una. Todo lo escrito es de la
 * tienda; nada inventado.
 */
import type { Tarjeta, TarjetaProducto, TarjetaTienda } from '#lib/tienda/compartir.ts';
import { alfa, type Estilo, type Tipografia } from './estilos.ts';

export const ANCHO = 1200;
export const ALTO = 630;

/** Una foto ya lista (data URI) con su tamaño. */
export type Foto = { uri: string; ancho: number; alto: number };
/** Ancho real de un texto en una línea, en em (a 1 px), o null si no se midió. */
export type Medir = (texto: string, tipo: Tipografia) => number | null;
export type Recursos = { fotos: Foto[]; logo: Foto | null; medir?: Medir };

/** Lo que mide el dibujo en curso (lo pone `dibujar`). */
let medir: Medir = () => null;

type Estilos = Record<string, string | number | undefined>;
export type Nodo = {
	type: string;
	props: {
		style?: Estilos;
		children?: Hijo | Hijo[];
		src?: string;
		width?: number;
		height?: number;
	};
};
type Hijo = Nodo | string | null | false | undefined;

function div(style: Estilos, ...children: Hijo[]): Nodo {
	const hijos = children.filter((c): c is Nodo | string => !!c);
	return {
		type: 'div',
		props: {
			// Como en CSS con border-box: el padding va adentro del ancho y el alto.
			style: { display: 'flex', boxSizing: 'border-box', ...style },
			children: hijos.length === 1 ? hijos[0] : hijos
		}
	};
}

function img(f: Foto, style: Estilos): Nodo {
	return {
		type: 'img',
		props: { src: f.uri, width: f.ancho, height: f.alto, style: { objectFit: 'cover', ...style } }
	};
}

/**
 * Sin renglones viudos: una palabra corta («de», «y», «la») va pegada a la que sigue, y la última
 * palabra no queda sola si es corta. Pega con espacio duro (satori no corta ahí).
 */
export function sinViudas(texto: string) {
	const p = texto.trim().split(/ +/);
	if (p.length < 2) return texto.trim();
	let out = '';
	for (let i = 0; i < p.length; i++) {
		const sigue = i < p.length - 1;
		// Con tres palabras o más, la última corta tampoco queda sola («Panadería Lili» sí se parte).
		const pegar =
			sigue && (p[i].length <= 3 || (p.length > 2 && i === p.length - 2 && p[i + 1].length <= 5));
		out += p[i] + (sigue ? (pegar ? '\u00a0' : ' ') : '');
	}
	return out;
}

/** Cuántos renglones ocupa el texto cortando por palabras (anchos en px, y el del espacio). */
function renglones(anchos: number[], espacio: number, ancho: number) {
	let n = 1;
	let linea = 0;
	for (const w of anchos) {
		if (w > ancho) return Infinity;
		const con = linea ? linea + espacio + w : w;
		if (con <= ancho) linea = con;
		else {
			n++;
			linea = w;
		}
	}
	return n;
}

/**
 * El tamaño de letra más grande con el que `texto` entra en la caja (en hasta `lineas` renglones),
 * estimando el ancho de cada letra y cortando por palabras: ninguna palabra queda partida.
 */
export function tamano(
	texto: string,
	caja: { ancho: number; alto: number; max: number; min?: number; lineas?: number },
	tipo: Tipografia | Pick<Tipografia, 'ancho' | 'mayusculas'>,
	alturaLinea = 1.02
) {
	const palabras = texto.trim().split(/ +/);
	// Medido con la fuente (con un 4 % de aire) o, si no se pudo, estimado por letra.
	const factor = 0.58 * tipo.ancho * (tipo.mayusculas ? 1.15 : 1);
	const real = 'familia' in tipo ? palabras.map((p) => medir(p, tipo)) : [];
	const espacioReal = 'familia' in tipo ? medir(' ', tipo) : null;
	const em = palabras.map((p, i) => (real[i] != null ? real[i]! * 1.04 : p.length * factor));
	const espacio = espacioReal ?? factor;
	const min = caja.min ?? 28;
	for (let t = caja.max; t > min; t -= 2) {
		const n = renglones(
			em.map((w) => w * t),
			espacio * t,
			caja.ancho
		);
		if (n <= (caja.lineas ?? 3) && n * t * alturaLinea <= caja.alto) return t;
	}
	return min;
}

const letra = (t: Tipografia, size: number): Estilos => ({
	fontFamily: t.familia,
	fontWeight: t.peso,
	fontStyle: t.italica ? 'italic' : 'normal',
	textTransform: t.mayusculas ? 'uppercase' : 'none',
	letterSpacing: `${t.tracking}em`,
	fontSize: size
});

/** El nombre de la tienda (o del producto), grande y sin renglones cortos colgando. */
function titular(
	e: Estilo,
	texto: string,
	caja: { ancho: number; alto: number; max: number; lineas?: number },
	extra: Estilos = {}
) {
	texto = sinViudas(texto);
	const size = tamano(texto, caja, e.titulo);
	return div(
		{
			...letra(e.titulo, size),
			lineHeight: e.titulo.mayusculas ? 0.95 : 1.02,
			color: e.tinta,
			maxWidth: caja.ancho,
			textWrap: 'balance',
			...extra
		},
		texto
	);
}

function frase(e: Estilo, texto: string, ancho: number, size = 30, extra: Estilos = {}) {
	if (!texto) return null;
	texto = sinViudas(texto);
	const s = tamano(texto, { ancho, alto: size * 3.6, max: size, min: 22, lineas: 3 }, e.frase, 1.2);
	return div(
		{
			...letra(e.frase, s),
			letterSpacing: 0,
			lineHeight: 1.25,
			color: e.oscura ? alfa(e.tinta, 0.82) : e.suave,
			maxWidth: ancho,
			textWrap: 'balance',
			...extra
		},
		texto
	);
}

/** Texto chico: el detalle (24 productos · Córdoba) o la dirección. */
function chico(e: Estilo, texto: string, extra: Estilos = {}) {
	if (!texto) return null;
	return div(
		{
			fontFamily: e.texto,
			fontWeight: 600,
			fontSize: 22,
			letterSpacing: '0.01em',
			color: e.oscura ? alfa(e.tinta, 0.7) : e.suave,
			...extra
		},
		texto
	);
}

/** La dirección de la tienda con un punto del acento adelante. */
function direccion(e: Estilo, texto: string) {
	return div(
		{ alignItems: 'center', gap: 12 },
		div({ width: 12, height: 12, borderRadius: 999, background: e.acento }),
		chico(e, texto, { color: e.tinta, fontWeight: 600 })
	);
}

/** El logo de la tienda, entrando en la caja sin deformarse. */
function logo(l: Foto | null, maxAncho: number, maxAlto: number) {
	if (!l) return null;
	const k = Math.min(maxAncho / l.ancho, maxAlto / l.alto, 1.6);
	const ancho = Math.round(l.ancho * k);
	const alto = Math.round(l.alto * k);
	return {
		type: 'img',
		props: { src: l.uri, width: ancho, height: alto, style: { width: ancho, height: alto } }
	} satisfies Nodo;
}

/** La inicial del nombre (la primera letra o número). */
export function inicial(nombre: string) {
	return (nombre.match(/[\p{L}\p{N}]/u)?.[0] ?? '·').toUpperCase();
}

/** La inicial grande dentro de una forma del acento (las tarjetas sin fotos). */
function sello(e: Estilo, nombre: string, lado: number, forma: 'circulo' | 'rombo' | 'cuadro') {
	const letraSize = Math.round(lado * 0.56);
	const caja: Estilos = {
		width: lado,
		height: lado,
		background: e.acento,
		alignItems: 'center',
		justifyContent: 'center',
		borderRadius: forma === 'circulo' ? 999 : forma === 'cuadro' ? e.radio * 2 : e.radio,
		transform: forma === 'rombo' ? 'rotate(-8deg)' : undefined
	};
	return div(
		caja,
		div(
			{
				...letra({ ...e.titulo, mayusculas: false }, letraSize),
				color: e.sobre,
				lineHeight: 1,
				transform: forma === 'rombo' ? 'rotate(8deg)' : undefined
			},
			inicial(nombre)
		)
	);
}

const lienzo = (e: Estilo, ...hijos: Hijo[]) =>
	div(
		{
			width: ANCHO,
			height: ALTO,
			position: 'relative',
			background: e.fondo,
			color: e.tinta,
			fontFamily: e.texto,
			overflow: 'hidden'
		},
		...hijos
	);

/** El marco fino (Ceibo, Malbec, Hora). */
const marco = (e: Estilo) =>
	div({
		position: 'absolute',
		left: 26,
		top: 26,
		width: ANCHO - 52,
		height: ALTO - 52,
		border: `2px solid ${e.oscura ? alfa(e.acento, 0.55) : alfa(e.tinta, 0.18)}`
	});

/** Arriba del nombre: el logo o el detalle. */
function encabezado(e: Estilo, t: TarjetaTienda, r: Recursos, centrado = false) {
	const l = logo(r.logo, 320, 84);
	if (l) return div({ justifyContent: centrado ? 'center' : 'flex-start' }, l);
	// Sin nada arriba, un hueco: el nombre queda en el medio y no pegado arriba.
	if (!t.detalle) return div({ height: 1 });
	return chico(e, t.detalle, {
		textTransform: 'uppercase',
		letterSpacing: '0.14em',
		fontSize: 19,
		color: e.oscura ? e.acento : e.suave
	});
}

/** Debajo del nombre en las que tienen logo arriba: el detalle (si no se mostró ya). */
const pie = (e: Estilo, t: TarjetaTienda, r: Recursos) => (r.logo ? chico(e, t.detalle) : null);

// ---- Editorial: Ceibo, Malbec, Quebracho, Rosedal, Taller, Torno ----

function editorial(e: Estilo, t: TarjetaTienda, r: Recursos) {
	const [foto] = r.fotos;
	if (!foto) {
		const ancho = 920;
		return lienzo(
			e,
			e.adorno === 'marco' && marco(e),
			e.adorno === 'arco' &&
				div({
					position: 'absolute',
					left: 170,
					top: 40,
					width: 860,
					height: 760,
					borderTopLeftRadius: 430,
					borderTopRightRadius: 430,
					background: alfa(e.acento, 0.1)
				}),
			div(
				{
					position: 'absolute',
					left: 0,
					top: 0,
					width: ANCHO,
					height: ALTO,
					flexDirection: 'column',
					alignItems: 'center',
					justifyContent: 'center',
					gap: 22,
					padding: '70px 140px',
					textAlign: 'center'
				},
				encabezado(e, t, r, true),
				titular(
					e,
					t.nombre,
					{ ancho, alto: 270, max: 132 },
					{ textAlign: 'center', justifyContent: 'center' }
				),
				div(
					{ alignItems: 'center', gap: 14 },
					div({ width: 70, height: 2, background: e.acento }),
					div({ width: 9, height: 9, borderRadius: 999, background: e.acento }),
					div({ width: 70, height: 2, background: e.acento })
				),
				frase(e, t.frase, 840, 34, { textAlign: 'center', justifyContent: 'center' }),
				pie(e, t, r)
			),
			div(
				{ position: 'absolute', left: 0, bottom: 52, width: ANCHO, justifyContent: 'center' },
				direccion(e, t.direccion)
			)
		);
	}
	const arco = e.adorno === 'arco';
	const enMarco = e.adorno === 'marco';
	const fotoAncho = arco ? 470 : enMarco ? 520 : 560;
	const fotoAlto = arco ? ALTO - 70 : enMarco ? ALTO - 104 : ALTO;
	const izq = ANCHO - fotoAncho - (arco ? 60 : enMarco ? 52 : 0);
	return lienzo(
		e,
		enMarco && marco(e),
		div(
			{
				position: 'absolute',
				left: 0,
				top: 0,
				width: izq,
				height: ALTO,
				flexDirection: 'column',
				justifyContent: 'space-between',
				padding: '72px 56px 64px 80px'
			},
			encabezado(e, t, r),
			div(
				{ flexDirection: 'column', gap: 20 },
				titular(e, t.nombre, { ancho: izq - 136, alto: 260, max: 112 }),
				e.adorno === 'linea' && div({ width: 84, height: 4, background: e.acento }),
				frase(e, t.frase, izq - 136, 30),
				pie(e, t, r)
			),
			direccion(e, t.direccion)
		),
		div(
			{
				position: 'absolute',
				right: arco ? 60 : enMarco ? 52 : 0,
				top: arco ? 70 : enMarco ? 52 : 0,
				width: fotoAncho,
				height: fotoAlto,
				overflow: 'hidden',
				...(arco
					? { borderTopLeftRadius: fotoAncho / 2, borderTopRightRadius: fotoAncho / 2 }
					: { borderRadius: e.radio > 4 ? e.radio : 0 })
			},
			img(foto, {
				width: fotoAncho,
				height: fotoAlto,
				...(arco ? { borderTopLeftRadius: fotoAncho / 2, borderTopRightRadius: fotoAncho / 2 } : {})
			})
		),
		e.adorno === 'linea' &&
			div({ position: 'absolute', left: izq, top: 0, width: 6, height: ALTO, background: e.acento })
	);
}

// ---- Grilla: Andes, Mostaza, Pampa, Patio, Hora, Tijera, Lente ----

const FORMA: Record<string, 'circulo' | 'rombo' | 'cuadro'> = {
	circulo: 'circulo',
	sol: 'circulo',
	rombo: 'rombo',
	bloque: 'cuadro',
	franja: 'cuadro',
	marco: 'circulo'
};

/** Las fotos en mosaico dentro de la caja: una grande, dos a la par o una alta y dos apiladas. */
function mosaico(e: Estilo, fotos: Foto[], ancho: number, alto: number) {
	const gap = 14;
	const radio = Math.min(e.radio, 28);
	const tile = (f: Foto, w: number, h: number) =>
		div(
			{ width: w, height: h, borderRadius: radio, overflow: 'hidden' },
			img(f, { width: w, height: h })
		);
	if (fotos.length === 1) return tile(fotos[0], ancho, alto);
	if (fotos.length === 2)
		return div(
			{ gap },
			tile(fotos[0], (ancho - gap) / 2, alto),
			tile(fotos[1], (ancho - gap) / 2, alto)
		);
	const grande = Math.round((ancho - gap) * 0.58);
	const resto = ancho - gap - grande;
	return div(
		{ gap },
		tile(fotos[0], grande, alto),
		div(
			{ flexDirection: 'column', gap },
			tile(fotos[1], resto, (alto - gap) / 2),
			tile(fotos[2], resto, (alto - gap) / 2)
		)
	);
}

function grilla(e: Estilo, t: TarjetaTienda, r: Recursos) {
	const fotos = r.fotos.slice(0, 3);
	if (!fotos.length) {
		const forma = FORMA[e.adorno] ?? 'circulo';
		const lado = 330;
		return lienzo(
			e,
			e.adorno === 'marco' && marco(e),
			e.adorno === 'sol' &&
				div({
					position: 'absolute',
					right: -170,
					top: -170,
					width: 760,
					height: 760,
					borderRadius: 999,
					background: alfa(e.acento, 0.16)
				}),
			div(
				{
					position: 'absolute',
					right: 96,
					top: e.adorno === 'franja' ? 0 : 96,
					width: lado,
					height: e.adorno === 'franja' ? ALTO - 96 : lado
				},
				e.adorno === 'franja'
					? div(
							{
								width: lado,
								height: ALTO - 96,
								background: e.acento,
								alignItems: 'flex-end',
								justifyContent: 'center',
								paddingBottom: 30
							},
							div({ ...letra(e.titulo, 230), color: e.sobre, lineHeight: 1 }, inicial(t.nombre))
						)
					: sello(e, t.nombre, lado, forma)
			),
			div(
				{
					position: 'absolute',
					left: 0,
					top: 0,
					width: 720,
					height: ALTO,
					flexDirection: 'column',
					justifyContent: 'space-between',
					padding: '76px 0 64px 84px'
				},
				encabezado(e, t, r),
				div(
					{ flexDirection: 'column', gap: 20 },
					titular(e, t.nombre, { ancho: 620, alto: 290, max: 128 }),
					frase(e, t.frase, 600, 30),
					pie(e, t, r)
				),
				direccion(e, t.direccion)
			)
		);
	}
	const caja = { x: 600, y: 48, ancho: ANCHO - 600 - 48, alto: ALTO - 96 };
	return lienzo(
		e,
		e.adorno === 'marco' && marco(e),
		(e.adorno === 'circulo' || e.adorno === 'sol') &&
			div({
				position: 'absolute',
				right: -120,
				top: -150,
				width: 520,
				height: 520,
				borderRadius: 999,
				background: e.acento
			}),
		e.adorno === 'rombo' &&
			div({
				position: 'absolute',
				left: 600,
				top: -80,
				width: 200,
				height: 820,
				background: e.acento,
				transform: 'rotate(14deg)'
			}),
		div(
			{ position: 'absolute', left: caja.x, top: caja.y, width: caja.ancho, height: caja.alto },
			mosaico(e, fotos, caja.ancho, caja.alto)
		),
		div(
			{
				position: 'absolute',
				left: 0,
				top: 0,
				width: 560,
				height: ALTO,
				flexDirection: 'column',
				justifyContent: 'space-between',
				padding: '70px 20px 60px 72px'
			},
			encabezado(e, t, r),
			div(
				{ flexDirection: 'column', gap: 18 },
				titular(e, t.nombre, { ancho: 460, alto: 270, max: 104 }),
				e.adorno === 'franja' && div({ width: 140, height: 12, background: e.acento }),
				e.adorno === 'bloque' &&
					div({ width: 56, height: 8, borderRadius: 8, background: e.acento }),
				frase(e, t.frase, 450, 27),
				pie(e, t, r)
			),
			direccion(e, t.direccion)
		)
	);
}

// ---- Catálogo: Miga, Miga clásica, Yerbal, Litoral ----

/** La nota escrita a mano de Miga, con el detalle real. */
const nota = (e: Estilo, texto: string) =>
	texto
		? div(
				{
					fontFamily: 'Caveat',
					fontWeight: 700,
					fontSize: 40,
					color: e.acento,
					transform: 'rotate(-3deg)'
				},
				texto
			)
		: null;

/** Las categorías como carta: nombre, puntos y cuántos productos tiene. */
function carta(e: Estilo, cats: TarjetaTienda['categorias'], ancho: number) {
	return div(
		{
			flexDirection: 'column',
			width: ancho,
			padding: '34px 38px',
			gap: 16,
			borderRadius: Math.min(e.radio + 6, 30),
			background: alfa(e.tinta, 0.06)
		},
		...cats.map((c) =>
			div(
				{ alignItems: 'flex-end', gap: 12, fontSize: 28, fontWeight: 600, color: e.tinta },
				div(
					{ fontFamily: e.titulo.familia, fontWeight: e.titulo.peso, letterSpacing: '-0.01em' },
					c.nombre
				),
				div({
					flexGrow: 1,
					height: 2,
					marginBottom: 9,
					borderBottom: `3px dotted ${alfa(e.tinta, 0.3)}`
				}),
				div({ color: e.acento, fontWeight: 700 }, String(c.cantidad))
			)
		)
	);
}

function catalogo(e: Estilo, t: TarjetaTienda, r: Recursos) {
	const fotos = r.fotos.slice(0, 4);
	if (fotos.length >= 3) {
		const gap = 16;
		const ancho = Math.floor((ANCHO - 112 - gap * (fotos.length - 1)) / fotos.length);
		const alto = 280;
		return lienzo(
			e,
			div(
				{
					position: 'absolute',
					left: 56,
					top: 52,
					width: ANCHO - 112,
					justifyContent: 'space-between',
					alignItems: 'flex-start',
					gap: 40
				},
				div(
					{ flexDirection: 'column', gap: 12, maxWidth: 720 },
					logo(r.logo, 260, 64),
					titular(e, t.nombre, { ancho: 720, alto: 170, max: 92, lineas: 2 }),
					frase(e, t.frase, 700, 26)
				),
				div(
					{ flexDirection: 'column', alignItems: 'flex-end', gap: 10, paddingTop: 8 },
					e.adorno === 'nota' ? nota(e, t.detalle) : chico(e, t.detalle)
				)
			),
			div(
				{ position: 'absolute', left: 56, bottom: 100, gap },
				...fotos.map((f) =>
					div(
						{ width: ancho, height: alto, borderRadius: e.radio, overflow: 'hidden' },
						img(f, { width: ancho, height: alto })
					)
				)
			),
			div({ position: 'absolute', left: 56, bottom: 40 }, direccion(e, t.direccion))
		);
	}
	const [foto] = fotos;
	const derecha = ANCHO - 640;
	const ladoDerecho: Hijo = foto
		? div(
				{ width: 480, height: ALTO - 112, borderRadius: e.radio + 8, overflow: 'hidden' },
				img(foto, { width: 480, height: ALTO - 112 })
			)
		: t.categorias.length >= 2
			? carta(e, t.categorias, 480)
			: sello(e, t.nombre, 380, 'circulo');
	return lienzo(
		e,
		div(
			{
				position: 'absolute',
				left: 0,
				top: 0,
				width: 640,
				height: ALTO,
				flexDirection: 'column',
				justifyContent: 'space-between',
				padding: '68px 40px 60px 72px'
			},
			(r.logo
				? logo(r.logo, 300, 80)
				: e.adorno === 'nota'
					? nota(e, t.detalle)
					: chico(e, t.detalle)) ?? div({ height: 1 }),
			div(
				{ flexDirection: 'column', gap: 18 },
				titular(e, t.nombre, { ancho: 528, alto: 280, max: 118 }),
				frase(e, t.frase, 520, 30),
				r.logo && (e.adorno === 'nota' ? nota(e, t.detalle) : chico(e, t.detalle))
			),
			direccion(e, t.direccion)
		),
		div(
			{
				position: 'absolute',
				left: 640,
				top: 0,
				width: derecha,
				height: ALTO,
				alignItems: 'center',
				justifyContent: 'center',
				paddingRight: 56
			},
			ladoDerecho
		)
	);
}

// ---- La foto propia: va sola, entera ----

function propia(e: Estilo, f: Foto) {
	return lienzo(e, img(f, { width: ANCHO, height: ALTO }));
}

// ---- Producto ----

function producto(e: Estilo, p: TarjetaProducto, r: Recursos) {
	const [foto] = r.fotos;
	const lado = ALTO;
	const ancho = ANCHO - lado;
	const izquierda: Hijo = foto
		? img(foto, { width: lado, height: ALTO })
		: div(
				{
					width: lado,
					height: ALTO,
					alignItems: 'center',
					justifyContent: 'center',
					background: alfa(e.acento, 0.14)
				},
				sello(e, p.nombre, 300, FORMA[e.adorno] ?? 'circulo')
			);
	return lienzo(
		e,
		div({ position: 'absolute', left: 0, top: 0, width: lado, height: ALTO }, izquierda),
		div(
			{
				position: 'absolute',
				left: lado,
				top: 0,
				width: ancho,
				height: ALTO,
				flexDirection: 'column',
				justifyContent: 'space-between',
				padding: '60px 56px 56px 60px'
			},
			logo(r.logo, 280, 70) ??
				div(
					{ ...letra({ ...e.titulo, mayusculas: false }, 30), lineHeight: 1.1, color: e.tinta },
					p.tienda
				),
			div(
				{ flexDirection: 'column', gap: 26 },
				titular(e, p.nombre, { ancho: ancho - 116, alto: 250, max: 80 }),
				div(
					{ alignItems: 'center', gap: 18, flexWrap: 'wrap' },
					div(
						{
							padding: '10px 24px',
							borderRadius: Math.max(e.radio, 6),
							background: e.acento,
							color: e.sobre,
							fontFamily: e.texto,
							fontWeight: 700,
							fontSize: 44
						},
						p.precio
					),
					p.anterior &&
						div(
							{
								fontFamily: e.texto,
								fontWeight: 600,
								fontSize: 30,
								color: e.oscura ? alfa(e.tinta, 0.6) : e.suave,
								textDecoration: 'line-through'
							},
							p.anterior
						)
				)
			),
			direccion(e, p.direccion)
		)
	);
}

/** El árbol completo de la tarjeta. */
export function dibujar(tarjeta: Tarjeta, e: Estilo, r: Recursos): Nodo {
	medir = r.medir ?? (() => null);
	if (tarjeta.tipo === 'producto') return producto(e, tarjeta, r);
	if (tarjeta.propia && r.fotos[0]) return propia(e, r.fotos[0]);
	if (e.comp === 'ed') return editorial(e, tarjeta, r);
	if (e.comp === 'grid') return grilla(e, tarjeta, r);
	return catalogo(e, tarjeta, r);
}
