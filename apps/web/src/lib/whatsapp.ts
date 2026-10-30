/**
 * Lleva un número argentino a E.164 sin "+" para WhatsApp (549 + característica + número), igual que
 * `whatsapp.Normalizar` de la API (mismos casos en los tests). Acepta "11 5555-0000",
 * "+54 9 11 5555 0000", "0351 15 555 0000"… Devuelve "" si no sirve.
 */
export function normalizarWhatsapp(s: string): string {
	let n = s.replace(/\D/g, '');
	if (n.length === 13 && n.startsWith('549')) return n;
	if (n.length === 12 && n.startsWith('54')) return '549' + n.slice(2);
	if (n.startsWith('0')) n = n.slice(1);
	// "0351 15 555 0000": característica + 15 + número; el 15 se saca.
	if (n.length === 12) {
		for (const k of [2, 3, 4]) {
			if (n.slice(k, k + 2) === '15') {
				n = n.slice(0, k) + n.slice(k + 2);
				break;
			}
		}
	}
	return n.length === 10 ? '549' + n : '';
}

/** El link de WhatsApp para ese número (null si no sirve). */
export function linkWhatsapp(tel: string, texto?: string): string | null {
	const n = normalizarWhatsapp(tel);
	if (!n) return null;
	return `https://wa.me/${n}${texto ? `?text=${encodeURIComponent(texto)}` : ''}`;
}
