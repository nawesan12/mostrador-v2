/**
 * Tope de Tincho en la demo pública: 15 mensajes por visitante por día (en su sandbox). En la demo no
 * hay IA (responde el motor con respuestas armadas); con la API real los cupos los mide y corta Go
 * (ia_uso, según el plan).
 */
export const POR_VISITANTE = 15;

const diaAr = (d = new Date()) =>
	new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Argentina/Buenos_Aires' }).format(d);

export { diaAr };
