# Estado del trabajo

Última actualización: 20/10/2026. Plan completo: `~/.claude/plans/claude-vamos-con-algo-encapsulated-alpaca.md`.

## Hecho (mergeado y desplegado en mostrador.iamnaw.dev)

| PR | Qué |
|---|---|
| #1–#9 | Bootstrap, design system (`/componentes`), fundación Go (roles, RLS, outbox), sandbox D1 + contrato, auth, landing parte 1 |
| #10–#11 | Landing parte 2 y arreglos de mobile |
| #12–#16 | Legales (arrepentimiento con código ARR), Seguridad, Nosotros, 404, formularios (contacto, partners, novedades), `/migrar`, `/estado` con monitor real (cron de `mostrador-jobs`) |
| #17 | Onboarding de 8 pasos (es el registro) |
| #18 | Ingreso con Google, Mercado Pago, Mercado Libre y link por WhatsApp (apagados hasta tener credenciales) |
| #19 | Datos de la demo: catálogo real de las 16 plantillas (`scripts/catalogos.ts`), clientes y pedidos sembrados |
| #20 | Armazón del panel: rail, barra lateral, topbar, barra de abajo, notificaciones, ⌘K, tema y modo Simple/Completo |
| #28 | Crear pedido (cliente existente o nuevo, productos, entrega; ya pagado o con link) y Link de pago: modal del handoff (Resumen, ⌘K, bandeja, crear pedido), página pública `/p/[código]` sin cookie, al pagar el pedido pasa a preparar. ⌘K con las acciones del handoff. Go: tabla `links` con RLS + `app.tienda_de_link` (definer con política por código); pagar da 409 hasta conectar Mercado Pago. D1: migración 0004 (índice por clave) |
| #27 | Ventas en el local (punto de venta: grilla, ticket, medios de pago, dividir el pago; crea el pedido y descuenta stock), Devoluciones (aprobar, ofrecer cambio, deshacer) y Reseñas y preguntas (respuesta de Hornero para editar y publicar; resumen por estrellas). Go: pedido nuevo real con stock en la misma transacción |
| #26 | Bandeja (conversaciones de WhatsApp, Instagram, Telegram y Mercado Libre; hilo, sugerencia de Hornero armada con los datos, tomar/devolver, respuestas rápidas, ficha del cliente) y Carritos abandonados (KPIs, recordar uno o todos, automáticos). Go: automáticos reales (00011); bandeja y carritos vacíos hasta conectar canales |
| #25 | Inventario (KPIs, movimientos de hoy, contar stock, pedido a proveedores por WhatsApp, códigos de barras Code 39 para imprimir), Facturación (comprobantes desde pedidos facturados, factura manual A/B) y Envíos (etapas, etiquetas, reclamos). Go: inventario y envíos reales, movimientos de hoy |
| #24 | Motor de listas (filtros en la URL, segmentos y facetas, lista virtual, lote, sugerencia de Hornero, exportar CSV) con Clientes (ficha, alta, etiquetas, cupón) y Productos (editor lateral de PanelCatalogo, alta, importar CSV, cambiar precio, publicar/pausar, mejorar descripción con Hornero). Go: tabla `productos` (00010), Clientes desde pedidos, filtro genérico |
| #23 | Pedidos: vistas, embudo, filtros en la URL, lista virtual, acciones por fila y en lote, hoja de detalle, exportar CSV. Go: tabla `pedidos` (migración 00009) con RLS, lista con facetas, detalle, estados y lote |
| #22 | Hornero con Workers AI (ver abajo) |
| #21 | Resumen (dashboard) completo, meta del mes (migración Go 00008), demo con 60 días de pedidos, stock, carritos, reseñas y chats |

## Hornero (PR #22)

- Panel lateral (`lib/panel/Hornero.svelte`, handoff `ovAi`): saludo con ventas de hoy contra ayer, ideas (recuperar carritos, pausar sin stock, promo de fin de semana) con Aprobar / Editar / Deshacer, charla con chips y streaming. Se abre desde el rail, la barra de abajo, el Resumen y el ⌘K (opción 0; Enter va a la sección si lo escrito coincide con una).
- Servidor: `lib/server/hornero/` y `routes/panel/hornero/`. Modelo `@cf/google/gemma-4-26b-a4b-it` con `chat_template_kwargs.enable_thinking = false` (si razona, se gasta los tokens y responde vacío).
- **Límite gratis de Workers AI (10.000 neurons por día, toda la cuenta):** topes de 15 preguntas por visitante y 200 por día para toda la demo (`cupo.ts`). Pasado el tope de la demo, contesta con las respuestas armadas, sin llamar al modelo. Pedido de Nahuel: probar el modelo real sólo dentro del límite gratis, con pocas preguntas.
- Al modelo se le pasa "ayer completo" (`ayerDiaCompleto`), no "ayer a esta hora": con ese dato confundía el total del día.
- En modo API (Go) todavía no hay topes por plan: llegan con Cuenta / Plan.

## Estados de un pedido

El circuito es el de PanelPedidos del handoff: Sin pagar → Para preparar → Listo para despachar → En camino → Entregado (más Cancelado). Sólo se puede avanzar una etapa, volver una (deshacer) o cancelar. El Resumen, el menú y Hornero leen los mismos pedidos con los cambios del visitante (`DatosSandbox.demo()`).

## Lo que sigue (orden del plan)

Listas de precio → Cupones → Crecer → Operación → Cuenta → Ventas (bandeja WhatsApp, Carritos, Devoluciones, POS, Reseñas) → Listas de precio → Cupones → Crecer → Operación → Cuenta → storefront y 16 plantillas → integraciones reales → MCP + OAuth → Marca → endurecimiento.

## Trampas que ya mordieron

- **`e.currentTarget` después de un `await`** es null: guardarlo antes (pasó en Importar CSV).
- **Tests que dependen de la hora** (saludo de Hornero): contemplar mañana, tarde y noche.

- **Tests de Go sin Postgres se saltan (SKIP) en local:** antes de correrlos, `bun run db:up`; confirmar con `go test -v ./... | grep -c '^--- SKIP'` que dé 0. En CI no se saltan (fallan).

- **`wrangler dev` con el binding de IA** pide credenciales al arrancar (en CI no hay). El script `preview` corre con `--local --var HORNERO:simulado`; no sacarlo.
- **Deploy manual:** el job `deploy` del CI no hace nada (no hay `CLOUDFLARE_API_TOKEN` en GitHub). Después de mergear: `cd apps/web && bun run build && bunx wrangler deploy`. Preview: `bunx wrangler versions upload --preview-alias pr-N`.
- **`bun scripts/icons.ts ph:nombre`** reescribe `src/lib/icons.ts` sin formato: correr `bunx prettier --write src/lib/icons.ts` y revisar que el diff sólo sume íconos.
- **Tests con miles de `expect`** en loops se pasan de los 5 s: juntar los errores y hacer un solo `expect`.
- **Hora de Argentina:** UTC−3 fijo; `inicioDelDia` y `horaAr` en `sandbox/tablero.ts`, formatos en `lib/panel/formato.ts`.
- **Handoff con desbordes en mobile** (leyenda del gráfico, aviso de ARCA): se corrige y se anota en el PR; nunca desbordar ni dejar palabras solas en el último renglón.
