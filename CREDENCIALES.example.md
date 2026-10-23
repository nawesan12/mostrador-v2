# Credenciales de Mostrador v2

Este archivo NO se sube a git (está en `.gitignore`).
Completá cada valor a la derecha del `=` y cambiá el estado a ✅.
Lo que diga "lo genero yo" no hace falta que lo consigas.

Leyenda: ⬜ falta · ✅ cargado · 🧪 cargado y probado

---

## 1. Cloudflare (PR 1: lo necesito primero)

Dónde: dash.cloudflare.com → My Profile → API Tokens → Create Token → "Edit Cloudflare Workers" + permisos extra.
Permisos del token: Workers Scripts:Edit, Workers Routes:Edit, D1:Edit, R2:Edit, Workers AI:Read, Turnstile:Edit, Zone DNS:Edit (zona iamnaw.dev), Account Analytics:Read.

- ⬜ `CLOUDFLARE_API_TOKEN=`
- ⬜ `CLOUDFLARE_ACCOUNT_ID=` (dash → Workers & Pages → panel derecho)
- ⬜ Confirmar que la zona `iamnaw.dev` está en esta misma cuenta: sí / no
- ⬜ Turnstile (dash → Turnstile → Add site, dominio `mostrador.iamnaw.dev` + `localhost`)
  - `PUBLIC_TURNSTILE_SITE_KEY=`
  - `TURNSTILE_SECRET=`
- ⬜ Web Analytics (dash → Analytics → Web Analytics → Add site) `PUBLIC_CF_ANALYTICS_TOKEN=`
- D1 `mostrador-sandbox`, R2 `mostrador-files` y el binding de Workers AI: los creo yo con el token.

## 2. GitHub (PR 1)

- ✅ Cuenta `nawesan12` logueada en `gh` (verificado).
- El repo `nawesan12/mostrador-v2` (privado) lo creo yo. Confirmá: ¿privado? sí / no
- Secrets de Actions: los cargo yo con los valores de este archivo.

## 3. Resend: mails (PR 5)

Dónde: resend.com → API Keys. Dominio: Domains → Add `mostrador.iamnaw.dev` y cargar los DNS que pida (SPF/DKIM). OJO: sin el MX de Resend en `iamnaw.dev`, para no pisar el Email Routing de Caja.
Ojo: la cuenta compartida con Caja tiene tope de 100 mails por día. ¿Usamos la misma o una nueva?

- ⬜ `RESEND_API_KEY=`
- ⬜ `MAIL_FROM=` (sugerido: `Mostrador <hola@mostrador.iamnaw.dev>`)

## 4. Google (PR 5 login, PR 8 onboarding, PR 20 integraciones)

Dónde: console.cloud.google.com → proyecto "Mostrador".

- ⬜ OAuth client (Web), con pantalla de consentimiento en producción. Redirects:
  `https://mostrador.iamnaw.dev/ingresar/google/vuelta`, `http://localhost:5173/ingresar/google/vuelta`
  - `GOOGLE_CLIENT_ID=`
  - `GOOGLE_CLIENT_SECRET=`
- ⬜ Places API (New) habilitada + API key restringida: `GOOGLE_PLACES_API_KEY=`
- ⬜ Merchant Center (para Google Shopping): `GOOGLE_MERCHANT_ID=`, más una service account JSON con acceso a Content API: `GOOGLE_SERVICE_ACCOUNT_JSON=`
- ⬜ GA4 del sitio de Mostrador: `PUBLIC_GA4_ID=` (G-XXXX) y `GA4_API_SECRET=` (Measurement Protocol)

## 5. Mercado Pago (PR 5 login, PR 17 abono, PR 20 cobros)

Dónde: mercadopago.com.ar/developers → Tus integraciones → Crear aplicación "Mostrador" (pagos online + suscripciones + OAuth/marketplace).
Redirects OAuth: ingreso `https://mostrador.iamnaw.dev/ingresar/mercadopago/vuelta` y conexión de cobros `https://mostrador.iamnaw.dev/integraciones/mercadopago/callback` (+ localhost).

- ⬜ `MP_CLIENT_ID=`
- ⬜ `MP_CLIENT_SECRET=`
- ⬜ `MP_PUBLIC_KEY_TEST=` / `MP_ACCESS_TOKEN_TEST=`
- ⬜ `MP_PUBLIC_KEY=` / `MP_ACCESS_TOKEN=` (producción, puede esperar)
- ⬜ `MP_WEBHOOK_SECRET=` (Webhooks → configurar → clave secreta)
- ⬜ Usuarios de prueba (vendedor y comprador): usuario/contraseña de cada uno
- Planes de suscripción (preapproval_plan) de Emprende, Pro y Negocio, mensual y anual: los creo yo por API.

## 6. Mercado Libre (PR 5 login, PR 20 canal)

Dónde: developers.mercadolibre.com.ar → Mis aplicaciones → Crear. Redirects: ingreso `https://mostrador.iamnaw.dev/ingresar/mercadolibre/vuelta` y canal `https://mostrador.iamnaw.dev/integraciones/mercadolibre/callback`. Scopes: read, write, offline_access. Notificaciones: orders_v2, items, questions, messages.

- ⬜ `ML_CLIENT_ID=`
- ⬜ `ML_CLIENT_SECRET=`
- ⬜ Usuario de test vendedor (lo podemos crear por API una vez que esté la app)

## 7. Meta: WhatsApp, Instagram, Messenger, Pixel (PR 5 códigos por WA, PR 12, PR 20)

Dónde: developers.facebook.com → Crear app tipo "Business" → productos WhatsApp, Instagram, Messenger, Facebook Login for Business. Business Manager verificado (para producción).

- ⬜ `META_APP_ID=`
- ⬜ `META_APP_SECRET=`
- ⬜ `META_CONFIG_ID=` (Embedded Signup para que cada tienda conecte su número)
- ⬜ Número de WhatsApp de la PLATAFORMA (para mandar los links de ingreso y de contraseña nueva):
  - `WHATSAPP_TOKEN=` (System User token permanente)
  - `WHATSAPP_PHONE_NUMBER_ID=`
  - `WHATSAPP_BUSINESS_ACCOUNT_ID=`
  - Dos plantillas de utilidad aprobadas, en español (es_AR), con un botón de link `https://mostrador.iamnaw.dev/ingresar?{{1}}`: `mostrador_ingreso` ("Tocá para entrar a tu cuenta de Mostrador") y `mostrador_clave` ("Tocá para crear tu contraseña nueva"). Ya no hace falta una de autenticación con código: `WHATSAPP_OTP_TEMPLATE=`
- `WHATSAPP_VERIFY_TOKEN`: lo genero yo.
- ⬜ Pixel del sitio de Mostrador: `PUBLIC_META_PIXEL_ID=` y `META_CAPI_TOKEN=`

## 8. TikTok Shop (PR 20)

Dónde: partner.tiktokshop.com → App. (Ojo: TikTok Shop no opera oficialmente en Argentina. ¿Lo hacemos igual, contra la región que esté disponible? sí / no)

- ⬜ `TIKTOK_APP_KEY=` / `TIKTOK_APP_SECRET=`

## 9. Telegram (PR 20)

Dónde: @BotFather → /newbot.

- ⬜ `TELEGRAM_BOT_TOKEN=`
- ⬜ `TELEGRAM_BOT_USERNAME=`
- `TELEGRAM_WEBHOOK_SECRET`: lo genero yo.

## 10. Medios de pago adicionales (PR 20)

- ⬜ MODO: alta como comercio/partner (modo.com.ar/comercios) → `MODO_CLIENT_ID=`, `MODO_CLIENT_SECRET=`, store id de sandbox
- ⬜ Payway (ex Decidir): developers.payway.com.ar → `PAYWAY_PUBLIC_KEY=`, `PAYWAY_PRIVATE_KEY=`, `PAYWAY_SITE_ID=` (sandbox)
- ⬜ Ualá Bis: developers.ualabis.com.ar → `UALA_USERNAME=`, `UALA_CLIENT_ID=`, `UALA_CLIENT_SECRET=` (stage)

## 11. Envíos (PR 20)

- ⬜ Andreani: `ANDREANI_USER=`, `ANDREANI_PASSWORD=`, `ANDREANI_CLIENTE=`, contratos `ANDREANI_CONTRATO_DOMICILIO=`, `ANDREANI_CONTRATO_SUCURSAL=` (de QA)
- ⬜ OCA e-Pak: `OCA_USER=`, `OCA_PASSWORD=`, `OCA_CUIT=`, `OCA_OPERATIVA_DOMICILIO=`, `OCA_OPERATIVA_SUCURSAL=`
- ⬜ Correo Argentino (MiCorreo / Paq.ar): `CORREO_USER=`, `CORREO_PASSWORD=`, `CORREO_CUSTOMER_ID=`
- Moto propia: no necesita credenciales.

## 12. ARCA (PR 17 facturas de Mostrador, PR 20 facturación de tiendas)

Dónde: con clave fiscal → "Administración de certificados digitales" (homologación: WSASS).

- ⬜ CUIT de la plataforma (quien factura el abono de Mostrador): `ARCA_CUIT=`
- ⬜ Certificado de homologación (.crt) y clave privada (.key) asociados a `wsfe` y `ws_sr_padron_a13`: los pasás como archivo, no acá
- ⬜ Punto de venta de facturación electrónica: `ARCA_PUNTO_VENTA=`
- ⬜ Condición frente al IVA de la plataforma (RI / monotributo): ______
- Producción: más adelante, mismos datos con certificado de producción.

## 13. Caja (PR 20). Lo arma la sesión de Caja

- ⬜ `CAJA_CLIENT_ID=`
- ⬜ `CAJA_CLIENT_SECRET=`
- ⬜ `CAJA_PROMO_SECRET=`
- ⬜ `CAJA_WEBHOOK_SECRET=`
- ⬜ URL base de la API de Caja: `CAJA_API_URL=https://caja.iamnaw.dev`

## 14. Monitoreo para /estado

- ✅ No hace falta: el monitor es propio (mostrador-jobs, una vez por hora, base D1 `mostrador-estado`).
- ⬜ Opcional: `RESEND_API_KEY` también en el worker `mostrador-jobs`, para mandar los avisos de incidentes por mail.

## 15. Datos que no son credenciales pero necesito

- ⬜ Razón social, CUIT y domicilio legal de Mostrador para los legales, el pie y la Data Fiscal: ______
- ⬜ Mails públicos: hola@, legales@, seguridad@, ventas@. ¿Existen o los creamos con Email Routing de Cloudflare?
- ⬜ WhatsApp humano de soporte (número): ______
- ⬜ Redes reales (IG, X, LinkedIn, YouTube) o se dejan las del diseño: ______
- ⬜ Fecha y lugar del evento /live: ______

## 16. Lo genero yo (no hace falta que hagas nada)

`SESSION_SECRET`, `CODES_HMAC_KEY`, `CREDENTIALS_MASTER_KEY` (AES-GCM + HKDF por tienda), `INTERNAL_SECRET`, `OAUTH_SIGNING_KEY` (MCP), `WHATSAPP_VERIFY_TOKEN`, `TELEGRAM_WEBHOOK_SECRET`, `REVALIDATION_KEY`, contraseñas de Postgres local.
