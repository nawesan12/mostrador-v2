# Mostrador · convenciones de código

Leé esto antes de tocar código. Sigue la impronta de Caja (`~/Projects/caja/docs/CONVENCIONES.md`):
la misma lógica, las mismas optimizaciones y, siempre que sirva, el mismo código. Antes de escribir
algo, buscalo en Caja (integraciones, sesiones, rate limit, caché, cripto, fechas, montos).

## Piezas

- `apps/web`: SvelteKit 3 (Svelte 5 con runes) en Cloudflare Workers. Landing, panel, tiendas
  (`/t/<tienda>`), onboarding y conector MCP son **una sola app**. Siempre vive en Workers.
- `apps/api`: API en Go + Postgres 17 para el VPS. Por ahora sólo local y CI.
- `apps/jobs`: Worker de cron cada 15 min → `POST /api/internal/cron` de la app (como `caja-jobs`).
- `DATA_SOURCE`: `sandbox` (demo pública, datos por visitante en D1, se borran a las 24 h) o `api`
  (la API Go). Las páginas no saben cuál está activo: leen de `#lib/server/datos/`.
- Gestor de paquetes: **bun**.

## Diseño

- La fuente de verdad visual es `design/` (handoff de alta fidelidad). Leé `design/README.md`.
- `python3 scripts/screen.py` lista las pantallas con etiqueta; los `.dc.html` se abren servidos
  por HTTP desde `design/` (`python3 -m http.server`).
- Pixel-perfect con **datos reales**. Breakpoints del handoff: ~760 px (mobile), ~1000 px (tablet),
  1280 px+ (desktop). Hit targets ≥ 44 px en mobile. Cero scroll horizontal a 375 px.
- Íconos **Phosphor** (como el handoff) con un `Icon.svelte` de lista cerrada, como en Caja.
  Logos de marcas como SVG locales. Nada de Tailwind: estilos por token en `app.css` y `<style>`.
- Hornero (sprites en `static/hornero/`): sólo en vacíos, carga, error, éxito, consejo y onboarding.

## Design system (`apps/web/src/lib`)

- Tokens y primitivas en `src/app.css`: `.btn` (`-primary`, `-dark`, `-secondary`, `-soft`, `-danger`,
  `-sm`, `-lg`, `-block`), `.icon-btn`, `.chip-estado`, `.card`, `.etiqueta`, `.input`, `.field`.
  Tema del panel con `data-theme="dark"` sobre el contenedor.
- Íconos: `#lib/icons.ts` (generado: `bun scripts/icons.ts ph:nombre` para sumar uno) + `Icon.svelte`.
- Identidad de canales, pagos y avisos: `#lib/marcas.ts` (`MARCAS`, `ESTADOS`) + `MarcaIcono.svelte`.
- Avisos: `getUi().toast({ k, title, desc, action, onAction })` y `getUi().undoable(...)`.
- `Modal`, `Hoja`, `Alerta`, `Vacio`, `Interruptor`, `Segmentado`, `ChipEstado`, `Grano` en
  `#lib/components/`. Animaciones de entrada con `use:anim` (`#lib/anim.ts`).
- Catálogo vivo en `/componentes`.

## Imports

- Alias `#lib/...` **con extensión**: `#lib/server/csrf.ts`.
- SvelteKit 3: `$app/state`, `$app/env`. Bindings y secretos: `import { env } from 'cloudflare:workers'`.

## Datos

- Toda consulta filtra por tienda (`storeId`), como Caja filtra por `spaceId`. En Postgres además
  la garantiza RLS; en el sandbox de D1, el código.
- Montos en enteros de centavos (`amountMinor`), fechas de negocio `YYYY-MM-DD` en hora de Argentina.
- `Promise.all` para consultas independientes; nada de consultas dentro de loops.
- Formularios: form actions + `use:enhance`, validación con `valibot`, errores con `fail(400, …)`.
- D1 (sandbox): mismas reglas que Caja — `inArray` con JSON, `insertMany` en tandas, `db.batch`
  como única atomicidad, tope de 100 parámetros, `explain query plan` antes de mergear, índices que
  empiezan por la columna de tienda, triggers de versión para la caché de páginas.
- Escrituras que no hace falta esperar: dentro de `waitUntil(...)`.
- Límites y precios: una sola fuente (`#lib/plans.ts`). Nunca un precio escrito a mano.

## Textos

- Español rioplatense (vos), claro y corto. Ortografía completa (tildes, ñ, ¿?).
- Vacío con la primera acción; cargando con esqueletos; error que diga qué pasó y cómo seguir.
- Toda acción destructiva o de cambio de estado: toast con "Deshacer".

## Verificar (en cada PR)

- `bun run lint && bun run check && bun run test` en `apps/web`.
- `bun run test:e2e`: Playwright contra el build servido por wrangler, en desktop (1440) y mobile (375).
- `go vet ./... && go test -race ./...` en `apps/api`.
- Capturas a 1440 y 375 px en la descripción del PR, comparadas contra el `.dc.html`.

## Git

- Una rama por PR (`pr-NN-tema`), CI verde y merge a `main`. `main` despliega solo a
  `mostrador.iamnaw.dev` cuando están los secretos de Cloudflare.
- `CREDENCIALES.md` nunca va al repo (está en `.gitignore`); la plantilla es `CREDENCIALES.example.md`.
