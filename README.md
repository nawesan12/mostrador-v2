# Mostrador

Plataforma para crear y operar tiendas online en Argentina: landing, onboarding que precarga la
tienda, panel de administración, plantillas por rubro, Hornero (IA) e integración con Caja.

Demo: https://mostrador.iamnaw.dev

## Piezas

```
apps/web     SvelteKit en Cloudflare Workers: sitio, panel, tiendas (/t/<tienda>) y conector MCP
apps/api     API en Go + Postgres 17 (para el VPS; hoy local y CI)
apps/jobs    Worker de cron que llama a la app cada 15 minutos
design/      Handoff de diseño (fuente de verdad visual)
docs/        Convenciones y decisiones
infra/       Postgres y buzón de mails para desarrollo
```

## Correr

```bash
bun install
bun run dev              # app en http://localhost:5173
bun run db:up            # Postgres 17 en :5440 y Mailpit en :8026
bun run api:bootstrap    # roles y base mostrador_dev (una vez)
bun run api:migrar       # migraciones
bun run api:dev          # API Go en :4000
bun run api:worker       # worker del outbox
```

## Verificar

```bash
cd apps/web && bun run lint && bun run check && bun run test && bun run build && bun run test:e2e
cd apps/api && go vet ./... && go test -race ./...
```

Antes de tocar código: [`docs/CONVENCIONES.md`](docs/CONVENCIONES.md) y [`docs/decisiones.md`](docs/decisiones.md).
Las credenciales se listan en `CREDENCIALES.example.md`; la copia con valores (`CREDENCIALES.md`) no va al repo.
