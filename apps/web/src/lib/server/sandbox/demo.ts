/**
 * Datos de la tienda de ejemplo de la demo: clientes y pedidos de los últimos 30 días sobre el
 * catálogo real de la plantilla (src/lib/catalogos.ts). Salen siempre iguales para una plantilla
 * (generador con semilla) y se fechan respecto de cuándo se creó el sandbox, así la demo se ve viva
 * pero no cambia mientras se usa. Lo que el visitante modifica se guarda aparte (tabla `cambios`).
 *
 * Canales, medios de pago, envíos y estados son los del panel del handoff.
 */
import { catalogoDe, type Catalogo, type ProductoEjemplo } from '#lib/catalogos.ts';

export type Canal = 'web' | 'wa' | 'ml' | 'ig' | 'pos' | 'tg';
export type MedioPago = 'mp' | 'transf' | 'modo' | 'payway' | 'uala' | 'efectivo';
export type Envio = 'andreani' | 'oca' | 'correo' | 'moto' | 'retiro';
export type EstadoPedido =
	'Pago pendiente' | 'Para preparar' | 'Pagado' | 'Enviado' | 'Entregado' | 'Cancelado';

export type Cliente = {
	id: string;
	nombre: string;
	email: string | null;
	telefono: string | null;
	ciudad: string;
	foto: string | null;
	mayorista: boolean;
	alta: number;
};

export type Pedido = {
	numero: number;
	clienteId: string;
	canal: Canal;
	pago: MedioPago;
	/** "Aprobado", "Acreditada", "Pendiente", "Cobrado". */
	estadoPago: string;
	envio: Envio;
	ciudadEnvio: string;
	estado: EstadoPedido;
	facturado: boolean;
	items: { productoId: string; nombre: string; cantidad: number; precio: number }[];
	total: number;
	/** Milisegundos (epoch). */
	creado: number;
};

export type Demo = { catalogo: Catalogo; clientes: Cliente[]; pedidos: Pedido[] };

/** Generador con semilla (mulberry32). */
function azar(semilla: number) {
	let a = semilla >>> 0;
	return () => {
		a = (a + 0x6d2b79f5) >>> 0;
		let t = a;
		t = Math.imul(t ^ (t >>> 15), t | 1);
		t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}
const hash = (s: string) =>
	[...s].reduce((h, c) => Math.imul(h ^ c.charCodeAt(0), 16777619), 2166136261) >>> 0;

const NOMBRES = [
	'Lucía',
	'Martín',
	'Norma',
	'Carla',
	'Julián',
	'Valeria',
	'Sofía',
	'Tomás',
	'Camila',
	'Federico',
	'Florencia',
	'Matías',
	'Agustina',
	'Nicolás',
	'Paula',
	'Diego',
	'Julieta',
	'Gonzalo',
	'Micaela',
	'Santiago',
	'Belén',
	'Lautaro',
	'Rocío',
	'Facundo',
	'Antonella',
	'Joaquín',
	'Milagros',
	'Ezequiel'
];
const APELLIDOS = [
	'Fernández',
	'Gómez',
	'Ruiz',
	'Pereyra',
	'Sosa',
	'Acosta',
	'Paz',
	'Luna',
	'Ríos',
	'Martínez',
	'López',
	'Díaz',
	'Romero',
	'Álvarez',
	'Torres',
	'Benítez',
	'Medina',
	'Herrera',
	'Castro',
	'Molina'
];
const CIUDADES = [
	'CABA',
	'Córdoba',
	'Rosario',
	'Mendoza',
	'La Plata',
	'Mar del Plata',
	'Tucumán',
	'Salta',
	'Neuquén',
	'Posadas',
	'Lanús',
	'Quilmes',
	'San Isidro',
	'Bahía Blanca'
];
/** Caras del handoff (Unsplash). */
const FOTOS = [
	'1531058240690-006c446962d8',
	'1687422811062-a966b55cb217',
	'1764173039506-c5cc7b8ec14b',
	'1520333789090-1afc82db536a',
	'1687422810663-c316494f725a',
	'1556745753-b2904692b3cd',
	'1545262722-9e0d80a0bc01',
	'1562322140-8baeececf3df'
];
const MAYORISTAS = [
	'Distribuidora Sur',
	'Almacén El Puente',
	'Mayorista Norte',
	'Comercial Andina'
];

function elegir<T>(r: () => number, pesos: [T, number][]): T {
	const total = pesos.reduce((s, [, p]) => s + p, 0);
	let x = r() * total;
	for (const [v, p] of pesos) if ((x -= p) < 0) return v;
	return pesos[pesos.length - 1][0];
}

const MIN = 60_000;
const DIA = 1440 * MIN;
const PRIMER_NUMERO = 1048;

function generar(plantilla: string, base: number): Demo | null {
	const catalogo = catalogoDe(plantilla);
	if (!catalogo) return null;
	const r = azar(hash(plantilla));
	const sinAcento = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();

	const clientes: Cliente[] = [];
	const usados = new Set<string>();
	while (clientes.length < 44) {
		const n = NOMBRES[Math.floor(r() * NOMBRES.length)];
		const a = APELLIDOS[Math.floor(r() * APELLIDOS.length)];
		if (usados.has(n + a)) continue;
		usados.add(n + a);
		const i = clientes.length;
		clientes.push({
			id: `c${i + 1}`,
			nombre: `${n} ${a}`,
			email: r() < 0.85 ? `${sinAcento(n)}.${sinAcento(a)}@gmail.com` : null,
			telefono:
				r() < 0.8
					? `+54 9 11 ${String(4000 + Math.floor(r() * 5999))}-${String(1000 + Math.floor(r() * 8999))}`
					: null,
			ciudad: CIUDADES[Math.floor(r() * CIUDADES.length)],
			foto: i < FOTOS.length ? FOTOS[i] : null,
			mayorista: false,
			alta: base - Math.floor(r() * 400) * DIA
		});
	}
	for (const [i, m] of MAYORISTAS.entries())
		clientes.push({
			id: `m${i + 1}`,
			nombre: m,
			email: `compras@${sinAcento(m).replace(/[^a-z]/g, '')}.com.ar`,
			telefono: null,
			ciudad: CIUDADES[i * 3],
			foto: null,
			mayorista: true,
			alta: base - (200 + i * 40) * DIA
		});
	const minoristas = clientes.filter((c) => !c.mayorista);

	// Pedidos de los últimos 30 días, más seguidos hoy (como el feed del handoff).
	const pedidos: Pedido[] = [];
	let t = base - 4 * MIN;
	let numero = PRIMER_NUMERO;
	const productos = catalogo.productos;
	const unidad = (p: ProductoEjemplo) => (p.precio >= 100_000 ? 1 : p.precio >= 20_000 ? 2 : 4);
	while (t > base - 30 * DIA && numero > 0) {
		const edad = base - t;
		const canal = elegir<Canal>(r, [
			['web', 30],
			['wa', 26],
			['ml', 18],
			['ig', 12],
			['pos', 10],
			['tg', 4]
		]);
		const mayorista = canal === 'pos' && r() < 0.15;
		const cliente = mayorista
			? clientes[clientes.length - 1 - Math.floor(r() * MAYORISTAS.length)]
			: minoristas[Math.floor(r() * minoristas.length)];
		const cuantos = mayorista
			? 2
			: elegir(r, [
					[1, 55],
					[2, 32],
					[3, 13]
				]);
		const items: Pedido['items'] = [];
		for (let k = 0; k < cuantos; k++) {
			const p = productos[Math.floor(r() * productos.length)];
			if (items.some((x) => x.productoId === p.id)) continue;
			const cantidad = mayorista
				? 6 * (1 + Math.floor(r() * 6))
				: 1 + Math.floor(r() * unidad(p) * 0.5);
			items.push({
				productoId: p.id,
				nombre: p.nombre,
				cantidad,
				precio: mayorista ? Math.round(p.precio * 0.82) : p.precio
			});
		}
		const pago: MedioPago =
			canal === 'pos'
				? elegir(r, [
						['efectivo', 50],
						['mp', 30],
						['transf', 20]
					])
				: canal === 'ml'
					? 'mp'
					: elegir(r, [
							['mp', 55],
							['transf', 20],
							['modo', 10],
							['payway', 8],
							['uala', 7]
						]);
		const envio: Envio =
			canal === 'pos'
				? mayorista
					? 'moto'
					: 'retiro'
				: elegir(r, [
						['andreani', 45],
						['oca', 15],
						['correo', 12],
						['moto', 12],
						['retiro', 16]
					]);
		// Los recientes están en curso; los viejos, entregados (algunos cancelados).
		let estado: EstadoPedido;
		if (canal === 'pos' && !mayorista) estado = 'Entregado';
		else if (edad < 6 * 3600_000)
			estado = elegir(r, [
				['Para preparar', 50],
				['Pagado', 22],
				['Pago pendiente', 14],
				['Enviado', 14]
			]);
		else if (edad < 3 * DIA)
			estado = elegir(r, [
				['Enviado', 50],
				['Entregado', 35],
				['Para preparar', 10],
				['Cancelado', 5]
			]);
		else
			estado = elegir(r, [
				['Entregado', 94],
				['Cancelado', 6]
			]);
		if (pago === 'transf' && estado === 'Para preparar' && r() < 0.3) estado = 'Pago pendiente';
		const estadoPago =
			estado === 'Pago pendiente'
				? 'Pendiente'
				: pago === 'transf'
					? 'Acreditada'
					: pago === 'efectivo'
						? 'Cobrado'
						: 'Aprobado';
		pedidos.push({
			numero,
			clienteId: cliente.id,
			canal,
			pago,
			estadoPago,
			envio,
			ciudadEnvio: envio === 'retiro' ? catalogo.ciudad : cliente.ciudad,
			estado,
			facturado: estado !== 'Pago pendiente' && estado !== 'Cancelado' && r() < 0.9,
			items,
			total: items.reduce((s, i) => s + i.precio * i.cantidad, 0),
			creado: t
		});
		numero--;
		// Hoy, cada 30 a 75 minutos; antes, más espaciado (unos 20 por día).
		t -= edad < DIA ? (30 + r() * 45) * MIN : (40 + r() * 100) * MIN;
	}
	return { catalogo, clientes, pedidos };
}

// Por isolate: generar es barato (unos 600 pedidos), pero se repite en cada pedido del panel.
const cache = new Map<string, Demo | null>();

/** Los datos de ejemplo de una plantilla, fechados respecto de `base` (creación del sandbox). */
export function demoDe(plantilla: string, base: number): Demo | null {
	const clave = `${plantilla}:${base}`;
	if (!cache.has(clave)) {
		if (cache.size > 50) cache.clear();
		cache.set(clave, generar(plantilla, base));
	}
	return cache.get(clave)!;
}
