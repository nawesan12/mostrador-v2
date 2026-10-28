/**
 * Roles y permisos del backoffice del equipo (/staff). Es la misma matriz que Go
 * (`internal/dominio/equipo`) y la tabla `equipo.permisos`: los tests de los dos lados la comparan
 * contra `apps/api/internal/dominio/equipo/testdata/permisos.json`.
 */
import type { components } from '@mostrador/contrato/esquema';
import * as I from '#lib/icons.ts';
import type { IconData } from '#lib/icons.ts';

export type Rol = components['schemas']['RolEquipo'];
export type Permiso = components['schemas']['PermisoEquipo'];

const SOPORTE: Permiso[] = [
	'tiendas.ver',
	'tiendas.mirar',
	'tiendas.notas',
	'tiendas.prueba',
	'tiendas.pack',
	'consultas.ver',
	'consultas.responder'
];
const ADMINISTRACION: Permiso[] = [
	...SOPORTE.slice(0, 5),
	'tiendas.plan',
	'tiendas.credito',
	'tiendas.pausar',
	...SOPORTE.slice(5),
	'partners.ver',
	'partners.decidir',
	'facturacion.ver',
	'auditoria.ver'
];
const DUENIO: Permiso[] = [
	...ADMINISTRACION.slice(0, 8),
	'tiendas.cerrar',
	...ADMINISTRACION.slice(8)
];

export const PERMISOS: Record<Rol, Permiso[]> = {
	soporte: SOPORTE,
	administracion: ADMINISTRACION,
	duenio: DUENIO
};

export const ROLES: { id: Rol; nombre: string; lema: string }[] = [
	{ id: 'soporte', nombre: 'Soporte', lema: 'Lee tiendas y responde' },
	{ id: 'administracion', nombre: 'Administración', lema: 'Además, facturación y partners' },
	{ id: 'duenio', nombre: 'Dueño', lema: 'Todo, incluido dar de baja' }
];
export const nombreRol = (r: Rol) => ROLES.find((x) => x.id === r)?.nombre ?? r;

/** Navegación del backoffice: cada sección con el permiso que la abre. */
export const SECCIONES: {
	id: string;
	t: string;
	ruta: string;
	icon: IconData;
	permiso: Permiso;
}[] = [
	{ id: 'inicio', t: 'Inicio', ruta: '/staff', icon: I.phHouseSimpleFill, permiso: 'tiendas.ver' },
	{
		id: 'tiendas',
		t: 'Tiendas',
		ruta: '/staff/tiendas',
		icon: I.phStorefrontFill,
		permiso: 'tiendas.ver'
	},
	{
		id: 'consultas',
		t: 'Consultas',
		ruta: '/staff/consultas',
		icon: I.phTrayFill,
		permiso: 'consultas.ver'
	},
	{
		id: 'partners',
		t: 'Partners',
		ruta: '/staff/partners',
		icon: I.phHandshakeFill,
		permiso: 'partners.ver'
	},
	{
		id: 'facturacion',
		t: 'Facturación',
		ruta: '/staff/facturacion',
		icon: I.phReceiptFill,
		permiso: 'facturacion.ver'
	},
	{
		id: 'auditoria',
		t: 'Auditoría',
		ruta: '/staff/auditoria',
		icon: I.phShieldCheckFill,
		permiso: 'auditoria.ver'
	}
];

export const seccionesDe = (permisos: readonly Permiso[]) =>
	SECCIONES.filter((s) => permisos.includes(s.permiso));
