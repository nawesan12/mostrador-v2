/**
 * Topes de Hornero en la demo pública, para cuidar la cuota de Workers AI: por visitante (30 mensajes
 * por día, en su sandbox) y para toda la demo (1.500 por día).
 */
import { env } from 'cloudflare:workers';

export const POR_VISITANTE = 30;
export const POR_DIA = 1500;

const diaAr = (d = new Date()) =>
	new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Argentina/Buenos_Aires' }).format(d);

/** Suma un mensaje a la demo entera; false si ya se llegó al tope del día. */
export async function cupoDemo(ahora = new Date()) {
	const dia = diaAr(ahora);
	const fila = await env.SANDBOX.prepare(
		`INSERT INTO contadores (nombre, valor) VALUES (?1, 1)
		 ON CONFLICT (nombre) DO UPDATE SET valor = valor + 1 RETURNING valor`
	)
		.bind(`hornero:${dia}`)
		.first<{ valor: number }>();
	// Los días anteriores no se usan más.
	await env.SANDBOX.prepare(`DELETE FROM contadores WHERE nombre LIKE 'hornero:%' AND nombre < ?1`)
		.bind(`hornero:${dia}`)
		.run();
	return (fila?.valor ?? 0) <= POR_DIA;
}

export { diaAr };
