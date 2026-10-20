/** Plata como en el handoff: "$ 42.900" (es-AR, sin decimales). */
// Espacio duro entre el signo y el número: un precio nunca se parte en dos renglones.
export const fmt = (n: number) => '$\u00a0' + Math.round(n).toLocaleString('es-AR');

/** Foto de Unsplash del handoff (fotos de ejemplo: el README pide reemplazarlas por propias). */
export const unsplash = (id: string, w: number, extra = '') =>
	`https://images.unsplash.com/photo-${id}?auto=format&fit=crop&w=${w}&q=75${extra}`;
export const cara = (id: string) => unsplash(id, 96, '&h=96&crop=faces');
