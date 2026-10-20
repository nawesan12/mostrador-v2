# Decisiones

Lo que el código da por sentado, y por qué. Se agrega una entrada por decisión, nunca se reescribe.

## 1. Front siempre en Workers, API en Go para el VPS (20/10/2026)
La demo pública corre entera en Cloudflare con un sandbox por visitante en D1. La app real usa una
API en Go con Postgres 17 en un VPS, por concurrencia: RLS para aislar tiendas, reserva de stock con
`UPDATE` condicional, outbox en la misma transacción. D1 no tiene transacciones interactivas ni RLS.
El navegador sólo habla con SvelteKit (BFF); SvelteKit habla con Go server-to-server.

## 2. Una sola app SvelteKit
Landing, panel, tiendas y conector MCP comparten despliegue. Las tiendas van por path (`/t/<tienda>`)
en `mostrador.iamnaw.dev`; con dominio propio, por subdominio, resuelto en el hook `reroute`.

## 3. La impronta de Caja
Mismo gestor (bun), mismas convenciones y optimizaciones, y el código de Caja se reutiliza cuando
sirve. Excepciones pedidas: Playwright E2E en cada PR e íconos Phosphor.

## 4. Contrato único entre la demo y la API
`packages/contrato/openapi.yaml` define cada operación; una suite de contrato corre los mismos casos
contra el sandbox y contra la API Go, para que la demo no muestre algo que la app real no hace.

## 5. El aislamiento entre tiendas lo garantiza Postgres (PR #4)
Cada tabla con `tienda_id` pasa por `app.aplicar_rls`: RLS habilitada y **forzada** (alcanza al
dueño), política por tienda y lectura de staff. La tienda vive en `set_config('app.tienda_id', …, true)`,
local a la transacción: nunca un `SET` de sesión, que con un pool compartido filtraría una tienda a
otra. Sin tienda fijada, `app.tienda_actual()` es NULL y no se ve nada.
Roles: `mostrador_migrador` (dueño del esquema), `mostrador_app` (sin bypass, sin ser dueña de nada),
`mostrador_worker` (sólo el outbox) y `mostrador_staff` (sólo lectura). Tres tests lo sostienen: el
estructural (rompe si una tabla queda sin RLS; se verificó agregando una a propósito), el de cruces
entre dos tiendas y el de arquitectura (pgxpool sólo en `internal/plataforma/db`).

## 6. Outbox y worker en Go
El evento se inserta en la misma transacción que el cambio. El worker toma lotes con
`FOR UPDATE SKIP LOCKED` (varios workers no se pisan: probado con 4 en paralelo sobre 300 eventos),
reintenta con espera creciente (2, 4, 8… s, techo 1 h) y aparta el evento después de 8 intentos con
su error. La entrega es al menos una vez: los manejadores tienen que ser idempotentes. La app puede
insertar en el outbox pero no marcarlo.

## 7. Idempotency-Key en las escrituras
La clave se reserva en su propia transacción antes de ejecutar: un segundo pedido igual recibe 409
"en curso" o, si ya terminó, la respuesta guardada (`Idempotent-Replayed: true`). La misma clave con
otro cuerpo es 422. Un 5xx libera la clave para poder reintentar.

## 8. Errores como problem+json
Todos los errores salen como `application/problem+json` con `code` estable y un `detail` que le
sirve a una persona. Un panic en un handler es un 500 con ese formato, no un proceso caído.
