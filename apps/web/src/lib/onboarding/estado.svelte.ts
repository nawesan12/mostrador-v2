/**
 * Estado del onboarding (design/Onboarding Mostrador.dc.html). Lo que se elige queda en el navegador
 * (localStorage, como en el handoff: "ob-step" y "ob-tpl" los escribe también la landing) y se manda
 * al servidor al terminar o, con la API, en cada paso.
 */
import { getContext, setContext } from 'svelte';
import { slugDeNombre } from '#lib/tienda/direccion.ts';
import {
	CONEXIONES,
	CONFIGURACION,
	PASOS_DEMO,
	PASOS_REALES,
	type Conexion,
	type PasoId,
	type Seccion,
	type Tono
} from './datos.ts';

export type EstadoConexion = 'off' | 'wait' | 'on';

/** Lo que se guarda (y se manda al servidor). */
export type Elecciones = {
	paso: number;
	plantilla: string;
	paleta: number;
	fuente: number;
	instagram: boolean;
	secciones: Record<Seccion, boolean>;
	objetivos: number[];
	/** Qué vende (id de packages/contrato/rubros.json); con la API pasa a ser el rubro de la tienda. */
	rubro: string | null;
	conexiones: Partial<Record<Conexion, EstadoConexion>>;
	conflicto: string | null;
	config: Record<string, boolean>;
	tono: Tono;
	equipo: { email: string; rol: string }[];
	/** De dónde traer la tienda (sólo con la API: se importa al terminar). */
	importar: { plataforma: 'tiendanube' | 'shopify' | 'woocommerce'; url: string } | null;
};

export const INICIAL: Elecciones = {
	paso: 0,
	plantilla: 'litoral',
	paleta: 0,
	fuente: 0,
	instagram: true,
	secciones: { dest: true, cat: true, ig: true, res: true },
	objetivos: [1, 2],
	rubro: null,
	conexiones: { mp: 'on' },
	conflicto: null,
	config: {
		cuotas: true,
		transf: true,
		cash: true,
		andreani: true,
		retiro: true,
		free: true,
		moto: false,
		pol: true
	},
	tono: 'cercano',
	equipo: [{ email: 'valeria@laesquina.com.ar', rol: 'Dueña' }],
	importar: null
};

/**
 * Con la API (una cuenta real) no hay nada de ejemplo: ni canales conectados, ni equipo inventado.
 * Lo de INICIAL es la ilustración de la demo.
 */
export const INICIAL_REAL: Elecciones = {
	...INICIAL,
	paso: 1,
	instagram: false,
	objetivos: [],
	conexiones: {},
	equipo: [],
	importar: null
};

const CLAVE = 'ob-estado';

function leerGuardado(base: Elecciones): Elecciones {
	try {
		const e = {
			...base,
			...(JSON.parse(localStorage.getItem(CLAVE) ?? '{}') as Partial<Elecciones>)
		};
		// La landing elige plantilla y paso por su cuenta ("Usar esta plantilla").
		const tpl = localStorage.getItem('ob-tpl');
		const paso = localStorage.getItem('ob-step');
		if (tpl) e.plantilla = tpl;
		if (paso !== null) e.paso = Math.max(0, Math.min(7, Number(paso) || 0));
		return e;
	} catch {
		return { ...base };
	}
}

/** Lo que la tienda tiene guardado del onboarding (GET /v1/panel/onboarding). */
export type Guardado = { plantilla?: string; elecciones: Record<string, unknown> } | null;

/**
 * Con la API, lo guardado en la tienda manda: se retoma también desde otro dispositivo y nunca se
 * mezcla con lo que dejó en este navegador otra cuenta (o la demo). Una cuenta recién creada arranca
 * de cero, salvo la plantilla que eligió en la landing antes de registrarse.
 */
function leerDeLaTienda(g: Guardado): Elecciones {
	const el = g?.elecciones ?? {};
	const e: Elecciones = { ...INICIAL_REAL };
	if (!Object.keys(el).length) {
		try {
			const tpl = localStorage.getItem('ob-tpl');
			if (tpl) e.plantilla = tpl;
		} catch {
			// Sin almacenamiento: la plantilla del rubro.
		}
		return e;
	}
	if (g?.plantilla) e.plantilla = g.plantilla;
	if (typeof el.rubro === 'string') e.rubro = el.rubro;
	if (el.secciones && typeof el.secciones === 'object')
		e.secciones = { ...e.secciones, ...(el.secciones as Elecciones['secciones']) };
	const imp = el.importar as Elecciones['importar'];
	if (imp && typeof imp === 'object' && typeof imp.plataforma === 'string')
		e.importar = { plataforma: imp.plataforma, url: String(imp.url ?? '') };
	const i = PASOS_REALES.indexOf(el.paso as PasoId);
	e.paso = i > 0 ? i : 1;
	return e;
}

export class Onboarding {
	/** La demo (sandbox) muestra el onboarding ilustrado; con la API, sólo lo que pasa de verdad. */
	demo = $state(true);
	e = $state<Elecciones>({ ...INICIAL });
	/** Paso 1: cuántas fuentes ya "aparecieron" (ARCA, Instagram, Maps). */
	revelado = $state(0);
	/** Paso 3: avance de la importación, de 0 a 100. */
	importado = $state(0);
	dispositivo = $state<'desk' | 'mob'>('desk');
	galeria = $state(false);
	consejo = $state(false);
	/**
	 * Paso 2 con la API: el nombre y la dirección de la tienda (decisiones §38). `tienda` es lo
	 * guardado; nombre y slug, lo que se está escribiendo (se guardan al seguir).
	 */
	tienda = $state<{ nombre: string; slug: string } | null>(null);
	nombre = $state('');
	slug = $state('');
	/** Si tocó la dirección a mano: entonces ya no la arma el nombre. */
	slugTocado = $state(false);
	estadoDireccion = $state<'actual' | 'revisando' | 'ok' | 'mal'>('actual');
	errorTienda = $state('');
	guardandoTienda = $state(false);
	#timers: ReturnType<typeof setInterval>[] = [];

	get paso() {
		return this.e.paso;
	}
	/** Los pasos de este recorrido: los 8 del handoff en la demo, menos con una cuenta real. */
	get pasos() {
		return this.demo ? PASOS_DEMO : PASOS_REALES;
	}
	get id(): PasoId {
		return this.pasos[this.e.paso] ?? 'listo';
	}
	get ultimo() {
		return this.pasos.length - 1;
	}
	get conectadas() {
		return CONEXIONES.filter(([k]) => this.e.conexiones[k] === 'on').length;
	}
	get importacionLista() {
		return this.importado >= 100;
	}
	/** Shopify y WooCommerce se traen desde el link de la tienda: sin link no hay qué importar. */
	get importacionValida() {
		const i = this.e.importar;
		return !i || i.plataforma === 'tiendanube' || i.url.trim().length >= 4;
	}
	errorImportacion = $state('');

	/** Con la API y sesión, `guardado` es lo de la tienda; sin sesión, siempre el paso 1 (Cuenta). */
	cargar(guardado?: Guardado) {
		if (this.demo) this.e = leerGuardado(INICIAL);
		else if (guardado !== undefined) this.e = leerDeLaTienda(guardado);
		else this.e = { ...INICIAL_REAL, paso: 0 };
		if (!this.demo) {
			this.revelado = 3;
			this.importado = 100;
		}
		this.e.paso = Math.min(this.e.paso, this.ultimo);
		this.entrar(this.e.paso);
	}

	/** Con la API, lo que manda lo elegido a la tienda (lo pone la página). */
	alCambiar: (() => void) | null = null;
	#espera: ReturnType<typeof setTimeout> | undefined;

	guardar() {
		// Con la API lo guarda la tienda (un momento después del último cambio, para no mandar uno
		// por toque); el navegador sólo recuerda la demo.
		if (!this.demo) {
			clearTimeout(this.#espera);
			this.#espera = setTimeout(() => this.alCambiar?.(), 800);
			return;
		}
		try {
			localStorage.setItem(CLAVE, JSON.stringify(this.e));
			localStorage.setItem('ob-step', String(this.e.paso));
			localStorage.setItem('ob-tpl', this.e.plantilla);
		} catch {
			// Sin almacenamiento (modo privado): el recorrido vale para esta visita.
		}
	}

	ir(paso: number) {
		const p = Math.max(0, Math.min(this.ultimo, paso));
		if (p === this.e.paso) return;
		this.e.paso = p;
		this.guardar();
		this.entrar(p);
		window.scrollTo({ top: 0 });
	}

	/** Lo que arranca solo al llegar a cada paso (como en el handoff). */
	entrar(paso: number) {
		this.detener();
		// Las animaciones de "encontramos" e "importando" son de la demo.
		if (!this.demo) return;
		if (paso === 1) {
			this.revelado = 0;
			this.#timers.push(
				setInterval(() => {
					if (this.revelado >= 3) return this.detener();
					this.revelado += 1;
				}, 950)
			);
		}
		if (paso === 3 && this.importado < 100)
			this.#timers.push(
				setInterval(() => {
					this.importado = Math.min(100, this.importado + 1.1);
					if (this.importado >= 100) this.detener();
				}, 70)
			);
	}

	detener() {
		clearTimeout(this.#espera);
		for (const t of this.#timers) clearInterval(t);
		this.#timers = [];
	}

	conectar(k: Conexion) {
		if (this.e.conexiones[k] && this.e.conexiones[k] !== 'off') return;
		this.e.conexiones[k] = 'wait';
		setTimeout(
			() => {
				this.e.conexiones[k] = 'on';
				this.guardar();
			},
			900 + Math.random() * 700
		);
	}

	conectarRecomendadas() {
		CONEXIONES.forEach(([k, , , , rec], i) => {
			if (rec && (this.e.conexiones[k] ?? 'off') === 'off')
				setTimeout(() => this.conectar(k), i * 260);
		});
	}

	/** La tienda guardada: lo que se escribe arranca de ahí. */
	fijarTienda(t: { nombre: string; slug: string } | null) {
		this.tienda = t;
		this.nombre = t?.nombre ?? '';
		this.slug = t?.slug ?? '';
		this.slugTocado = false;
		this.errorTienda = '';
	}

	/** Mientras no toque la dirección, la arma el nombre (con el nombre guardado, la guardada). */
	cambiarNombre(nombre: string) {
		this.nombre = nombre;
		this.errorTienda = '';
		if (this.slugTocado || !this.tienda) return;
		this.slug =
			nombre.trim() === this.tienda.nombre
				? this.tienda.slug
				: slugDeNombre(nombre) || this.tienda.slug;
	}

	/** El rubro propone su plantilla como punto de partida (se cambia en «Tu tienda»). */
	elegirRubro(id: string, plantilla: string) {
		this.e.rubro = id;
		this.e.plantilla = plantilla;
		this.guardar();
	}

	alternarObjetivo(i: number) {
		this.e.objetivos = this.e.objetivos.includes(i)
			? this.e.objetivos.filter((x) => x !== i)
			: [...this.e.objetivos, i].sort((a, b) => a - b);
		this.guardar();
	}

	alternarConfig(clave: string) {
		const op = CONFIGURACION.flatMap((g) => g.items).find((x) => x.clave === clave);
		if (!op || op.falta) return;
		this.e.config[clave] = !this.e.config[clave];
		this.guardar();
	}

	invitar(email: string) {
		const e = email.trim();
		if (!e || this.e.equipo.some((m) => m.email === e)) return;
		this.e.equipo = [...this.e.equipo, { email: e, rol: 'Ventas' }];
		this.guardar();
	}
}

const KEY = Symbol('onboarding');
export const setOnboarding = (o: Onboarding) => setContext(KEY, o);
export const getOnboarding = () => getContext<Onboarding>(KEY);
