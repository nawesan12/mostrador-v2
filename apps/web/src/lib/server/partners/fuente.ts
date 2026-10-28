/**
 * El panel de partners desde la API Go (DATA_SOURCE=api), con la sesión de la persona y la clave
 * del BFF. La demo (DATA_SOURCE=sandbox) implementa lo mismo en demo.ts.
 */
import type { CobroPartner, PanelPartner, PerfilPartner } from '#lib/partners/programa.ts';
import { ClienteBff, ErrorApi } from '../bff.ts';

export interface FuentePartners {
	/** El panel, o null si la cuenta todavía no es partner. */
	panel(): Promise<PanelPartner | null>;
	guardarPerfil(p: PerfilPartner): Promise<PerfilPartner>;
	guardarCobro(c: CobroPartner): Promise<CobroPartner>;
	subirFactura(mes: string, url: string): Promise<void>;
}

export class PartnersApi extends ClienteBff implements FuentePartners {
	constructor(
		base: string,
		claveBFF: string,
		cliente: { ip: string; agente: string },
		private token?: string
	) {
		super(base, claveBFF, cliente);
	}
	async panel() {
		try {
			return await this.pedir<PanelPartner>('GET', '/v1/partners/yo', undefined, this.token);
		} catch (e) {
			if (e instanceof ErrorApi && e.estado === 404) return null;
			throw e;
		}
	}
	guardarPerfil(p: PerfilPartner) {
		return this.pedir<PerfilPartner>('PUT', '/v1/partners/yo/perfil', p, this.token);
	}
	guardarCobro(c: CobroPartner) {
		return this.pedir<CobroPartner>('PUT', '/v1/partners/yo/cobro', c, this.token);
	}
	subirFactura(mes: string, url: string) {
		return this.pedir<void>(
			'PUT',
			`/v1/partners/yo/pagos/${encodeURIComponent(mes)}/factura`,
			{ url },
			this.token
		);
	}
	/** Un clic en el link (/empezar?ref=…). Sin sesión. */
	clic(ref: string) {
		return this.pedir<void>('POST', '/v1/partners/clics', { ref });
	}
}
