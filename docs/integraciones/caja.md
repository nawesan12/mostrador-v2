# Integración con Caja (caja.iamnaw.dev)

Lo construyó la sesión de Caja (PR #69 de Caja, en producción desde el 20/10/2026). Detalle
completo en `~/Projects/caja/docs/API.md`, sección "Apps (OAuth)". Se usa en el PR 20.

## Credenciales (CREDENCIALES.md §13)
- `client_id`: `app_mostrador` (se activa cuando Nahuel carga la fila del cliente en el D1 de Caja).
- `CAJA_CLIENT_SECRET`, `CAJA_PROMO_SECRET`, `CAJA_WEBHOOK_SECRET`: generados del lado de Caja; los pasa Nahuel.
- Redirects registrados: `https://tulocalweb.com.ar/panel/integraciones/caja/callback` y
  `http://localhost:5173/panel/integraciones/caja/callback`.
- Webhook de desconexión (Caja → Mostrador): `https://tulocalweb.com.ar/api/webhooks/caja`.

## OAuth 2.1 (código + PKCE S256, cliente confidencial)
- Autorizar: `GET https://caja.iamnaw.dev/oauth/autorizar?response_type=code&client_id=app_mostrador&redirect_uri=…&code_challenge=…&code_challenge_method=S256&state=…&scope=lectura%20movimientos:escribir%20esperados:escribir`.
  La persona elige el espacio; vuelve con `?code&state` o `error=access_denied`.
- Token: `POST https://caja.iamnaw.dev/api/oauth/token`, Basic `client_id:client_secret` (o el secret
  en el body), form o JSON. `grant_type=authorization_code&code&redirect_uri&code_verifier` →
  `{access_token:"caja_app_…", token_type, expires_in:2592000, refresh_token, scope, space_id}`.
- Refresh: `grant_type=refresh_token&refresh_token` con el mismo Basic. Rota (el par anterior muere);
  vence a los 90 días sin uso.
- Revocar (RFC 7009): `POST /api/oauth/revoke token=…`.
- Los tokens `caja_app_` sólo sirven en `/api/apps/v1`.

## API: `https://caja.iamnaw.dev/api/apps/v1`, `Authorization: Bearer caja_app_…`
Montos en centavos ARS. Todo idempotente por `externalId` (`mostrador:<tenantId>:<entidad>:<id>`).
Ítem suelto: 201 created · 200 existing · 400 error. Lote `{items:[…≤50]}`: siempre 200 con
`{items:[{externalId,id,status:"created|existing|error",error?}]}` en el mismo orden.

- `POST /movimientos` (`movimientos:escribir`):
  `{externalId,date,amountMinor,kind:"income|expense",description(≤120),category?,channel?,paymentMethod?,relatedExternalId?,invoice?:{tipo:"A|B|C",ptoVta,numero,cae?,caeVto?},meta?}`
  - Fecha futura = error. `description` queda como comercio; canal y medio, en la nota. Sin cuenta.
  - Categoría inexistente: Caja la crea (no gasta límite). Usar "Ventas Mostrador", "Comisiones", "Envíos".
  - Reintegro: egreso con `relatedExternalId` = la venta.
  - Factura posterior al pago: reenviar el mismo movimiento con `invoice` (devuelve existing y la registra).
- `POST /esperados` (`esperados:escribir`): `{externalId,expectedDate,amountMinor,description,direction:"in|out"}`.
- `POST /esperados/acreditar`: `{externalId,date,amountMinor}` → `{externalId,id,status,transactionId}`;
  201 acreditado, 200 ya estaba, 404 no existe, 409 cancelado.
- `GET /resumen` (`lectura`): `{lastSyncAt,gananciaMesMinor,mep:{compra,venta,at},monotributo:{categoria,usoPct}|null,plan:"free|pro|promax",promoHasta,espacio:{id,nombre}}`.
- `POST /desconectar`: revoca el par desde Mostrador.
- `POST /promo` (server a server, sin Bearer): header
  `X-Caja-Signature: t=<unix>,v1=<hex(HMAC-SHA256(PROMO_SECRET, "<t>.<body crudo>"))>`, ventana ±300 s.
  Body `{tenantId, accessToken}` → `{granted, promoHasta, motivo?}`. Pro Max 3 meses para el dueño
  del espacio; idempotente por tenant y una vez por persona; si ya tiene plan, `granted:false` + motivo.
- Webhook Caja → Mostrador: `POST {event:"connection.revoked",spaceId,at}` con `X-Caja-Signature`
  (misma fórmula, `WEBHOOK_SECRET`). Sin reintentos: si se pierde, nos enteramos por el 401.

Lo que entra por la app no se corta por plan: la sincronización de 14 meses entra entera.

## Límites y errores
- 120 pedidos/min por conexión · 5.000 ítems/día por espacio · promo: 30/h por IP → 429 `code:"rate_limit"`.
- Errores `{message, issues:[{path,message}], code?}`. 401 sin token o revocado; 403 `code:"scope"`.

## Mascota
`apps/web/static/caja/01-saludo.png` y `05-exito.png` (copiados de `~/Projects/caja/static/mascota/`).

## Estado
Probado en producción del lado de Caja: el token exige el secret, sin token da 401, la firma de la
promo valida. Falta el flujo completo con una conexión real (lo hace el PR 20).
