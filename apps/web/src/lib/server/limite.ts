/**
 * Límite por ventana, como `ratelimit.ts` de Caja: suma un uso de `clave` en una ventana de
 * `ventanaSeg` segundos y dice si todavía está dentro de `max`. Una sola consulta atómica (upsert);
 * la ventana se reinicia sola al vencer. Vive en la D1 del sandbox (tabla `limites`).
 */
import { lt, sql } from 'drizzle-orm';
import type { DbSandbox } from './sandbox/sandbox.ts';
import { limites } from './sandbox/schema.ts';

export async function golpe(
	db: DbSandbox,
	clave: string,
	max: number,
	ventanaSeg: number,
	ahora = Date.now()
): Promise<boolean> {
	const reinicia = ahora + ventanaSeg * 1000;
	const [fila] = await db
		.insert(limites)
		.values({ clave, cuenta: 1, reiniciaEn: reinicia })
		.onConflictDoUpdate({
			target: limites.clave,
			set: {
				cuenta: sql`case when ${limites.reiniciaEn} < ${ahora} then 1 else ${limites.cuenta} + 1 end`,
				reiniciaEn: sql`case when ${limites.reiniciaEn} < ${ahora} then ${reinicia} else ${limites.reiniciaEn} end`
			}
		})
		.returning({ cuenta: limites.cuenta });
	return (fila?.cuenta ?? 0) <= max;
}

/** Limpieza del cron: las ventanas vencidas no se usan más. */
export const purgarLimites = (db: DbSandbox, ahora = Date.now()) =>
	db.delete(limites).where(lt(limites.reiniciaEn, ahora));
