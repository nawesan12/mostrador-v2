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
| #21 | Resumen (dashboard) completo, meta del mes (migración Go 00008), demo con 60 días de pedidos, stock, carritos, reseñas y chats |

## En curso: PR #22, Hornero (rama `pr-22-hornero`, pusheada, sin PR abierto)

Hecho en la rama:
- Panel lateral de Hornero (`lib/panel/Hornero.svelte`, handoff `ovAi`): saludo con ventas de hoy contra ayer, ideas (recuperar carritos, pausar sin stock, promo de fin de semana) con Aprobar / Editar / Deshacer, charla con chips y streaming.
- Se abre desde el rail, la barra de abajo, el Resumen (avatar, "Ver las 3 ideas", "Responder con Hornero") y el ⌘K ("Preguntale a Hornero", opción 0).
- Servidor: `lib/server/hornero/` (contexto para el modelo, ideas, respuestas simuladas, stream de Workers AI, cupo) y rutas `routes/panel/hornero/+server.ts` (GET inicio, POST charla) y `ideas/[idea]/+server.ts`.
- Modelo: `@cf/google/gemma-4-26b-a4b-it` (binding `AI`, `remote: true`). Var `HORNERO`: `workers-ai` en producción, `simulado` en el preview local.
- Cupos de la demo: 30 mensajes por visitante por día (en `cambios`) y 1.500 por día en total (`contadores`, clave `hornero:AAAA-MM-DD`).
- Contrato: `GET /v1/panel/hornero/contexto`, `POST /v1/panel/hornero/ideas/{idea}`, y en `Tablero.hornero` se sumaron `audiencia` e `ideasHechas`. Go: contexto vacío e ideas 404 hasta que existan esos dominios.
- Las ideas aprobadas se guardan en el sandbox (entidad `hornero`); los carritos recordados del Resumen salen de la misma idea.
- vitest: 49 pasan. Check, lint y Go con `-race` pasan.

Falta:
1. Arreglar el e2e «aprobar una idea queda hecho y se puede deshacer» (`e2e/hornero.e2e.ts`). Los otros 2 e2e de Hornero pasan en desktop y mobile. Falla así:
   - desktop: después de «Deshacer», el botón «Aprobar» no vuelve a aparecer (línea 64);
   - mobile: después de «Aprobar», no aparece «Hecho» (línea 61).
   Sospecha: el `cargar()` de `Hornero.svelte` o el `invalidateAll()` del layout no refrescan las ideas (o el layout las cierra con `cerrarTodo`). Después, correr la suite completa.
2. Probar el modelo real con pocas llamadas (`bun run dev` con el binding remoto), verificar el formato del stream y que responda en rioplatense sin inventar.
3. Comparar el panel contra el handoff a 1440 y barrer 375 px.
4. Abrir el PR, pasar CI, mergear y desplegar (preview y deploy a mano).

## Lo que sigue (orden del plan)

Pedidos (vistas, embudo, facetas, lista virtual, barra masiva, detalle, crear pedido, link de pago) → motor de listas (Clientes, Productos, Inventario, Facturación, Envíos) → Ventas (bandeja WhatsApp, Carritos, Devoluciones, POS, Reseñas) → Listas de precio → Cupones → Crecer → Operación → Cuenta → storefront y 16 plantillas → integraciones reales → MCP + OAuth → Marca → endurecimiento.

## Trampas que ya mordieron

- **`wrangler dev` con el binding de IA** pide credenciales al arrancar (en CI no hay). El script `preview` corre con `--local --var HORNERO:simulado`; no sacarlo.
- **Deploy manual:** el job `deploy` del CI no hace nada (no hay `CLOUDFLARE_API_TOKEN` en GitHub). Después de mergear: `cd apps/web && bun run build && bunx wrangler deploy`. Preview: `bunx wrangler versions upload --preview-alias pr-N`.
- **`bun scripts/icons.ts ph:nombre`** reescribe `src/lib/icons.ts` sin formato: correr `bunx prettier --write src/lib/icons.ts` y revisar que el diff sólo sume íconos.
- **Tests con miles de `expect`** en loops se pasan de los 5 s: juntar los errores y hacer un solo `expect`.
- **Hora de Argentina:** UTC−3 fijo; `inicioDelDia` y `horaAr` en `sandbox/tablero.ts`, formatos en `lib/panel/formato.ts`.
- **Handoff con desbordes en mobile** (leyenda del gráfico, aviso de ARCA): se corrige y se anota en el PR; nunca desbordar ni dejar palabras solas en el último renglón.
