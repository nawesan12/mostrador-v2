import { sql, type SQL, type SQLWrapper } from 'drizzle-orm';

/**
 * `col IN (...)` con un solo parámetro (la lista como JSON), como en Caja: D1 permite hasta 100
 * parámetros por consulta y el `inArray` de drizzle se rompe con listas largas.
 */
export function inArray(column: SQLWrapper, values: readonly unknown[]): SQL {
	if (!values.length) return sql`0`;
	return sql`${column} in (select value from json_each(${JSON.stringify(values)}))`;
}
