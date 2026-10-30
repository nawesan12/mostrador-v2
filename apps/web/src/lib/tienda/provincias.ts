/** Las provincias con el código de MiCorreo (la letra del CPA). Gemelo de correoargentino.Provincias. */
export const PROVINCIAS = [
	['C', 'Ciudad de Buenos Aires'],
	['B', 'Buenos Aires'],
	['K', 'Catamarca'],
	['H', 'Chaco'],
	['U', 'Chubut'],
	['X', 'Córdoba'],
	['W', 'Corrientes'],
	['E', 'Entre Ríos'],
	['P', 'Formosa'],
	['Y', 'Jujuy'],
	['L', 'La Pampa'],
	['F', 'La Rioja'],
	['M', 'Mendoza'],
	['N', 'Misiones'],
	['Q', 'Neuquén'],
	['R', 'Río Negro'],
	['A', 'Salta'],
	['J', 'San Juan'],
	['D', 'San Luis'],
	['Z', 'Santa Cruz'],
	['S', 'Santa Fe'],
	['G', 'Santiago del Estero'],
	['V', 'Tierra del Fuego'],
	['T', 'Tucumán']
] as const;

export type CodigoProvincia = (typeof PROVINCIAS)[number][0];

export const nombreProvincia = (c: string) => PROVINCIAS.find(([k]) => k === c)?.[1] ?? c;

/**
 * La provincia que se puede dar por segura con sólo el CP de 4 números: CABA (1000 a 1499) y el
 * conurbano (1600 a 1999). Para el resto, quien compra la elige (los rangos se cruzan entre provincias).
 */
export function provinciaDeCP(cp: string): CodigoProvincia | null {
	const n = Number(cp.trim());
	if (!/^\d{4}$/.test(cp.trim())) return null;
	if (n >= 1000 && n <= 1499) return 'C';
	if (n >= 1600 && n <= 1999) return 'B';
	return null;
}
