# Variantes y galería de fotos — especificación común

Repo `~/Trabajo/vimo/mostrador-v2`. Leé `CLAUDE.md`, `docs/CONVENCIONES.md`, `docs/decisiones.md`.

## Qué es

Un producto puede tener **opciones** (hasta 3: "Talle", "Color", "Presentación"…; cada una con hasta 20 valores) y entonces se vende por **variante** (una combinación de valores: Talle M · Negro), cada una con su stock, SKU, y opcionalmente su precio y su foto. Además, el producto tiene una **galería** de hasta 8 fotos (la primera es la de portada; `foto` sigue existiendo como `fotos[0]` para compatibilidad). Un producto sin opciones funciona exactamente como hoy.

## Modelo (nombres exactos; JSON en camelCase)

- `Producto.opciones: { nombre: string; valores: string[] }[]` (0–3 opciones, nombre 1–20 chars, valores 1–20 chars, 1–20 valores, sin repetidos).
- `Producto.fotos: string[]` (0–8, https o ids de foto como hoy; `foto` = `fotos[0] ?? null`).
- `Variante: { id: string; valores: Record<string, string>; sku: string; precio: number | null; stock: number; foto: string | null; activa: boolean }`. `valores` tiene exactamente una clave por opción del producto, con un valor de esa opción. `precio: null` = usa el precio del producto. `foto` debe ser una de `fotos` (o null).
- `Producto.variantes: Variante[]` (vacío si no tiene opciones). Al guardar opciones se generan/ajustan las combinaciones: las nuevas arrancan con stock 0 y activa=true; las que dejan de existir se borran (sus pedidos ya guardaron el texto).
- Con opciones, el stock del producto (`stock`) es la **suma** del stock de sus variantes activas (en Go lo mantiene un trigger; en el sandbox se calcula), y el precio "desde" es el mínimo de las variantes activas con stock (o el del producto).
- Ítem de pedido: suma `varianteId: string | null` y `variante: string | null` (texto legible "Talle M · Negro", copiado al momento de la compra). Al vender, el stock se descuenta de la **variante** (UPDATE condicional, no sobrevende), y nunca se acepta una variante inactiva o de otro producto. Un producto con opciones **exige** `varianteId`.
- Carrito de la tienda: la línea se identifica por `productoId` + `varianteId`.
- `TiendaPublica.productos[]` (ProductoPublico) suma `opciones`, `variantes` (sólo activas: `{ id, valores, precio, stock, foto }`) y `fotos`.
- `PedidoNuevo.items[]`: `{ productoId, varianteId?, cantidad }` (panel y tienda).

## Reglas de UI

- **Tienda (13 plantillas)**: en la ficha, un selector por opción con las piezas del propio diseño (casi todos los handoff tenían talles/colores/presentaciones: `pdSizes`, swatches, etc. en `design/Plantilla <Nombre>.dc.html`; si una no lo tenía, armalo con chips de la plantilla). Valores sin stock en la combinación elegida: tachados/deshabilitados. Precio, stock, foto y "Agregar" siguen a la variante elegida; sin elegir todas las opciones, "Agregar" pide elegir ("Elegí un talle"). Galería: miniaturas de `fotos` como en el diseño (Ceibo, Andes, Rosedal… las tenían). Carrito y checkout muestran el texto de la variante. En las tarjetas del catálogo, "desde $" si las variantes tienen precios distintos y "Agotado" sólo si todas están sin stock.
- **Panel**: en el editor de producto (`EditorProducto.svelte`), "Fotos" (galería: subir varias, ordenar, elegir portada) y "Opciones" (agregar Talle/Color/…, valores como chips), con la grilla de variantes (valores, SKU, precio, stock, activa) y "Aplicar a todas" para precio/stock. Inventario y la lista de productos muestran el stock total y, al abrir, por variante. Crear pedido y Ventas en el local eligen variante.
- **Mercado Libre**: un producto con variantes se publica con `variations` (attribute_combinations por cada opción; si una opción no mapea a un atributo de la categoría, usar el nombre tal cual como atributo personalizado si ML lo acepta; si no, no publicar y avisar con un mensaje claro). Stock por variación.

## Demo (sandbox)

Los catálogos de ejemplo de moda, calzado, cosmética, gastronomía y regionales suman opciones verosímiles (talles S–XL, números 35–44, colores, presentaciones 250 g / 500 g / 1 kg…) con stock por variante, para que la demo las muestre. Lo del visitante se guarda como cambios.
