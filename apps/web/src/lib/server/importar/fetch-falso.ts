/** Para los tests: un fetch que responde según path y query, sin salir a la red. */
export type Ruta = unknown | { cuerpo: unknown; cabeceras: Record<string, string> };

export const conCabeceras = (cuerpo: unknown, cabeceras: Record<string, string>) => ({
	cuerpo,
	cabeceras
});

/** Busca primero path + query y después sólo el path; lo que no está, 404. Anota cada pedido. */
export function fetchFalso(rutas: Record<string, Ruta>, pedidos: string[] = []): typeof fetch {
	return (async (url: string) => {
		pedidos.push(url);
		const u = new URL(url);
		const r = rutas[u.pathname + u.search] ?? rutas[u.pathname];
		if (r === undefined) return new Response('no', { status: 404 });
		const conC = r && typeof r === 'object' && 'cuerpo' in r && 'cabeceras' in r;
		const { cuerpo, cabeceras } = conC
			? (r as { cuerpo: unknown; cabeceras: Record<string, string> })
			: { cuerpo: r, cabeceras: {} };
		return new Response(typeof cuerpo === 'string' ? cuerpo : JSON.stringify(cuerpo), {
			headers: cabeceras
		});
	}) as typeof fetch;
}
