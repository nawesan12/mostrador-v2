import { describe, expect, it } from 'vitest';
import casos from '@mostrador/contrato/casos/importar/empretienda.json';
import esperado from '@mostrador/contrato/casos/importar/empretienda.esperado.json';
import { fichaDeHtml, resultadoEmpretienda, type FichaEmpretienda } from './empretienda.ts';

describe('importar: Empretienda (mismos casos que Go)', () => {
	it('fichas → productos con variantes reales', () => {
		expect(resultadoEmpretienda(casos.fichas as FichaEmpretienda[])).toEqual(esperado);
	});

	it('saca la ficha de la página', () => {
		const html =
			'<script>\nvar p_nombre = "Mate";\nvar stock = [{"s_cantidad":2,"s_precio":5000,"valoratributo":[]}];\nvar cdn_products = "https://cdn/";\nvar imagenes = [{"idImagenes":1,"i_link":"a.jpg"}];\n</script><div class="product-vip__description"><p>Calabaza</p></div>';
		const f = fichaDeHtml(html, 'https://t.empretienda.com.ar/mates/mate');
		expect(f?.p_nombre).toBe('Mate');
		expect(f?.stock).toHaveLength(1);
		expect(f?.descripcion).toBe('<p>Calabaza</p>');
		expect(fichaDeHtml('<html>no es una ficha</html>', 'https://x')).toBeNull();
	});
});
