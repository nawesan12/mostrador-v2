/**
 * Topes de Hornero en la demo pública, para no salir nunca del límite gratis de Workers AI (10.000
 * neurons por día para toda la cuenta; cada pregunta manda unos 3.000 tokens de contexto): 15 por
 * visitante por día (en su sandbox) y 200 por día para toda la demo. Pasado el tope de la demo,
 * Hornero sigue respondiendo con las respuestas armadas (simulado), sin llamar al modelo.
 */
import { env } from 'cloudflare:workers';

export const POR_VISITANTE = 15;
export const POR_DIA = 200;

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
