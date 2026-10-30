/** Tipos del motor de listas (design/PanelListas.dc.html: `CFG` y `row()` de cada pantalla). */
import type { components } from '@mostrador/contrato/esquema';
import type { IconData } from '#lib/icons.ts';
import type { UiState } from '#lib/ui.svelte.ts';

export type ListaGenerica = components['schemas']['ListaGenerica'];
export type NombreLista = components['schemas']['NombreLista'];

/** Cómo se dibuja una fila (las piezas de la fila del handoff). */
export type VistaFila = {
	img?: string | null;
	ini?: string;
	iniBg?: string;
	iniFg?: string;
	/** Radio de la imagen: 50% (personas) o 12px (cosas). */
	radio: string;
	/** Cuadradito abajo a la derecha: marca (canal) o ícono. */
	insignia?: { marca?: string; icon?: IconData; col?: string };
	t: string;
	tags?: { icon: IconData; col: string; t: string }[];
	sub: string;
	spark?: { h: number; c: string }[];
	medidor?: { w: number; c: string; marca?: number };
	pasos?: string[];
	vL: string;
	vC: string;
	vI?: IconData;
	chip?: { t: string; bg: string; fg: string };
	v: string;
	vs: string;
	toggle?: { on: boolean; t: string };
	stepper?: number;
	act?: { t: string; icon: IconData; bg: string; fg: string };
};

/** Lo que la configuración puede hacer desde la lista. */
export type Ctx = {
	accion: (ids: string[], accion: string, valor?: unknown) => Promise<number>;
	ui: UiState;
	refrescar: () => Promise<void>;
};

export type Opcion = { k: string; t: string; marca?: string; icon?: IconData; col?: string };

export type ConfigLista<R extends { id: string }> = {
	/** Nombre de la lista del panel (el backoffice usa los suyos, con `ruta` en MotorLista). */
	lista: NombreLista | 'tiendas-equipo';
	titulo: string;
	unidad: string;
	buscar: string;
	kpi: (l: ListaGenerica) => string;
	primario?: { t: string; icon: IconData; hacer: (ctx: Ctx) => void | Promise<void> };
	tip?: (
		l: ListaGenerica
	) => { texto: string; boton: string; hacer: (ctx: Ctx) => Promise<void> } | null;
	vistas: { id: string; t: string; icon: IconData; color: string; filtros: Partial<FiltrosUrl> }[];
	segmentos: { k: string; t: string; c: string }[];
	/** Sin `opciones`: se usan los valores que vienen en los conteos (los 6 más comunes). */
	facetas: { k: string; t: string; opciones?: Opcion[] }[];
	rango: {
		t: string;
		max: (l: ListaGenerica) => number;
		paso: number;
		formato: (v: number) => string;
	};
	banderas: { k: string; t: string }[];
	ordenes: { k: string; t: string }[];
	fila: (r: R) => VistaFila;
	accionFila?: (r: R, ctx: Ctx) => Promise<void>;
	alternar?: (r: R, on: boolean, ctx: Ctx) => Promise<void>;
	ajustar?: (r: R, delta: number, ctx: Ctx) => Promise<void>;
	lote: { t: string; icon: IconData; hacer: (ids: string[], ctx: Ctx) => Promise<void> }[];
	/** Lo que se ve cuando no hay nada de verdad (sin filtros puestos): qué es y por dónde empezar. */
	vacio?: {
		hornero?: string;
		titulo: string;
		texto: string;
		accion?: { t: string; href?: string; hacer?: (ctx: Ctx) => void | Promise<void> };
	};
};

export type FiltrosUrl = {
	q: string;
	seg: string[];
	facetas: Record<string, string[]>;
	rango: number;
	banderas: string[];
	orden: string;
};
