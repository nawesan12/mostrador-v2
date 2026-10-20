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
