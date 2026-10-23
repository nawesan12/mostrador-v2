# Adaptar una plantilla con diseño propio

S = /private/tmp/claude-501/-Users-nahuelsantillan-Trabajo-vimo/cb219c60-3378-4580-9592-49566588ebb1/scratchpad/plantillas

Escribís UN componente Svelte 5 (runes, TypeScript) por plantilla: `S/<Nombre sin espacios>.svelte` (ej. `S/MigaClasica.svelte`). NO toques el repositorio `~/Trabajo/vimo/mostrador-v2` (sólo leelo; para compilar usá un worktree propio, ver abajo). Escribí únicamente en `S/`.

## Contexto

Mostrador es una plataforma de tiendas online argentinas (SvelteKit 5). Hay dos plantillas terminadas que son tu MODELO EXACTO de cómo se enchufan los datos reales a un marcado del handoff (leelas desde `origin/main`):

- `apps/web/src/lib/tienda/plantillas/Ceibo.svelte` (modelo principal: leelo ENTERO antes de empezar) y `apps/web/src/lib/tienda/plantillas/Motor.svelte`.
- Tipos: `apps/web/src/lib/tienda/tipos.ts`; `carrito.svelte.ts`, `envio.ts`, `titulo.ts` (partirTitulo), en `apps/web/src/lib/tienda/`.
- Diseño: `design/Plantilla <Nombre>.dc.html` (leé el marcado y el `<script type="text/x-dc">` entero: datos de ejemplo, renderVals, páginas).
- Conversor a Svelte: `cd apps/web && python3 scripts/plantilla.py "Plantilla <Nombre>" > S/_conv-<nombre>.txt`. OJO: el disco no distingue mayúsculas; nunca guardes el convertido con el mismo nombre que el componente.

## Reglas (obligatorias)

1. Mismas props y API que Ceibo (`tienda, pagina, producto, categoria, orden, pedido, error`), mismas páginas (inicio, catalogo, producto, carrito, checkout, nosotros, listo). Se intercambia con Ceibo sin tocar las rutas.
2. Fidelidad: conservá el marcado y TODOS los estilos del diseño (colores, tipografías, radios, espaciados); sólo cambiás valores de ejemplo por datos de `tienda`, igual que Ceibo. Si el diseño no tiene alguna de las 7 páginas, armala con piezas del propio diseño, nunca con estilos inventados.
3. Nada de datos de ejemplo: nombre, portada (título partido con partirTitulo, lema, foto), categorías, productos, más vendidos, reseñas y promedio, cobros, envíos, fotos, whatsapp y dirección salen de `tienda`. Lo que no tiene datos (variantes/talles, contadores inventados, cuentas regresivas, newsletter, seguidores, formularios sin destino, sellos de medios de pago que no están en tienda.cobros) NO se dibuja.
4. Funcionalidad idéntica a Ceibo: links reales a /t/{slug}/..., catálogo con chips + orden + "Ver más" de a 24, ficha con agregar al carrito (deshabilitado sin stock o en vitrina), WhatsApp, envío por CP, acordeón, relacionados; carrito +/−/quitar con barra de envío gratis; checkout `<form method="POST" use:enhance>` con los MISMOS `name` que Ceibo y el mismo manejo de enhance; página listo; nosotros; footer con los 4 legales y "Tienda creada con mostrador"; WhatsApp flotante; aviso de vitrina. Copiá la lógica del `<script>` de Ceibo.
5. Funciones propias de la plantilla (turnos, seña, reserva de drops, +18, club, menú gastronómico, configurador, comparar, factura A…): si se resuelven 100 % en el navegador sin mentir (ej. aviso +18 recordado en localStorage con try/catch, comparar productos ya cargados), hacelas. Si necesitan backend que no existe, NO las simules: no las dibujes y listalas en tu respuesta.
6. Sin marco de la herramienta de diseño ni barras de páginas/dispositivos. Responsive con `<svelte:window bind:innerWidth={ancho} />`, celular < 820 px; a 375 px nada se desborda horizontalmente.
7. Íconos: `import Icon from '#lib/components/Icon.svelte'` y `import * as I from '#lib/icons.ts'` (camelCase: ph:arrow-right-bold → I.phArrowRightBold, logos:whatsapp-icon → I.logosWhatsappIcon, simple-icons:x → I.siX). Si falta uno en icons.ts, listalo en la respuesta (no edites icons.ts) y usá uno parecido que exista.
8. Fuentes: `<svelte:head>` que cargue sólo las Google Fonts de la plantilla (Instrument Sans e Instrument Serif ya vienen del layout).
9. Calidad: `<!-- eslint-disable svelte/no-navigation-without-resolve -->` como en Ceibo, claves en todos los `{#each}`, nada de `any`, imports relativos `../carrito.svelte.ts`, `../envio.ts`, `../tipos.ts`, `../titulo.ts` (vivirá en src/lib/tienda/plantillas/). Textos en español rioplatense; en textos fijos largos, `&nbsp;` antes de la última palabra (sin palabras sueltas en el último renglón).
10. Comentario de cabecera corto como el de Ceibo.

## Verificación

`git -C ~/Trabajo/vimo/mostrador-v2 worktree add --detach S/wt-<nombre> origin/main`, después `cd S/wt-<nombre>/apps/web && bun install`, copiá tu componente a src/lib/tienda/plantillas/ y corré `bunx svelte-check --threshold error` y `bunx eslint <archivo>` hasta que no tenga errores. Al terminar: `git -C ~/Trabajo/vimo/mostrador-v2 worktree remove --force S/wt-<nombre>`. Releé el archivo buscando valores de ejemplo que hayan quedado.

## Qué devolver (corto)

Ruta de cada archivo, qué páginas/secciones se dibujan, qué sacaste por no tener datos, funciones propias hechas y pendientes por necesitar backend, íconos faltantes.
