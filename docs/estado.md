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
| #22 | Hornero con Workers AI (ver abajo) |
| #21 | Resumen (dashboard) completo, meta del mes (migración Go 00008), demo con 60 días de pedidos, stock, carritos, reseñas y chats |

## Hornero (PR #22)

- Panel lateral (`lib/panel/Hornero.svelte`, handoff `ovAi`): saludo con ventas de hoy contra ayer, ideas (recuperar carritos, pausar sin stock, promo de fin de semana) con Aprobar / Editar / Deshacer, charla con chips y streaming. Se abre desde el rail, la barra de abajo, el Resumen y el ⌘K (opción 0; Enter va a la sección si lo escrito coincide con una).
- Servidor: `lib/server/hornero/` y `routes/panel/hornero/`. Modelo `@cf/google/gemma-4-26b-a4b-it` con `chat_template_kwargs.enable_thinking = false` (si razona, se gasta los tokens y responde vacío).
- **Límite gratis de Workers AI (10.000 neurons por día, toda la cuenta):** topes de 15 preguntas por visitante y 200 por día para toda la demo (`cupo.ts`). Pasado el tope de la demo, contesta con las respuestas armadas, sin llamar al modelo. Pedido de Nahuel: probar el modelo real sólo dentro del límite gratis, con pocas preguntas.
- Al modelo se le pasa "ayer completo" (`ayerDiaCompleto`), no "ayer a esta hora": con ese dato confundía el total del día.
- En modo API (Go) todavía no hay topes por plan: llegan con Cuenta / Plan.

## Lo que sigue (orden del plan)

Pedidos (vistas, embudo, facetas, lista virtual, barra masiva, detalle, crear pedido, link de pago) → motor de listas (Clientes, Productos, Inventario, Facturación, Envíos) → Ventas (bandeja WhatsApp, Carritos, Devoluciones, POS, Reseñas) → Listas de precio → Cupones → Crecer → Operación → Cuenta → storefront y 16 plantillas → integraciones reales → MCP + OAuth → Marca → endurecimiento.

## Trampas que ya mordieron

- **`wrangler dev` con el binding de IA** pide credenciales al arrancar (en CI no hay). El script `preview` corre con `--local --var HORNERO:simulado`; no sacarlo.
- **Deploy manual:** el job `deploy` del CI no hace nada (no hay `CLOUDFLARE_API_TOKEN` en GitHub). Después de mergear: `cd apps/web && bun run build && bunx wrangler deploy`. Preview: `bunx wrangler versions upload --preview-alias pr-N`.
- **`bun scripts/icons.ts ph:nombre`** reescribe `src/lib/icons.ts` sin formato: correr `bunx prettier --write src/lib/icons.ts` y revisar que el diff sólo sume íconos.
- **Tests con miles de `expect`** en loops se pasan de los 5 s: juntar los errores y hacer un solo `expect`.
- **Hora de Argentina:** UTC−3 fijo; `inicioDelDia` y `horaAr` en `sandbox/tablero.ts`, formatos en `lib/panel/formato.ts`.
- **Handoff con desbordes en mobile** (leyenda del gráfico, aviso de ARCA): se corrige y se anota en el PR; nunca desbordar ni dejar palabras solas en el último renglón.
