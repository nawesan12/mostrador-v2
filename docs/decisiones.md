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

## 9. El sandbox de la demo es "copia al escribir" (PR #5)
El cupo gratis de D1 son 100.000 escrituras por día. Si cada visitante copiara su tienda de ejemplo
(el handoff muestra 2.400 pedidos), 40 visitas lo agotarían. Por eso los datos de ejemplo de cada
rubro viven en el código, iguales para todos, y D1 guarda sólo el sandbox (una fila) y lo que el
visitante cambia (`cambios`, por entidad y clave). Leer es: ejemplo + cambios del sandbox.
El sandbox vence a las 24 h; `mostrador-jobs` llama a `/api/internal/cron` una vez por hora y la
purga va de a tandas de 50. En desarrollo y en los e2e se usa la D1 local, nunca la remota.

## 10. Contrato compartido entre la demo y la API Go
`packages/contrato` tiene el OpenAPI (de donde salen los tipos del front), los rubros y los casos
(`casos/*.json`). El mismo caso corre en vitest contra el sandbox y en Go contra Postgres, así la
demo no puede mostrar algo que la API real no devuelve. Go no puede embeber archivos fuera de su
módulo: sus rubros son una tabla propia que un test compara contra `rubros.json`.

## 11. Las páginas no saben de dónde salen los datos
`datosDe(locals)` devuelve la implementación de `DATA_SOURCE`: el sandbox (D1) o la API Go (BFF:
el navegador nunca habla con Go). Los hooks sólo leen el sandbox en `/demo` y `/panel`, para que
la landing no gaste lecturas de D1.

## 12. La demo no tiene cuentas (PR #7)
En `DATA_SOURCE=sandbox`, `/ingresar` funciona entero (validación, errores, modos, medidor de fuerza,
recuperar) pero no crea cuentas: ingresar, registrarse o entrar con Google/WhatsApp llevan a la
demo, y "recuperar" muestra "Revisá tu email" sin mandar nada. Las cuentas reales viven en la API Go
(Postgres): hashear contraseñas en Workers Free pasaría el tope de 10 ms de CPU, y una demo pública
no necesita guardar datos personales. El destino de "recuperar" vuelve en la respuesta, no en la
URL, para que no quede en el historial ni en logs.

## 13. Cuentas y sesiones en la API Go (PR #8)
- Contraseñas con argon2id (parámetros de OWASP); el hash guarda sus parámetros para poder subirlos.
- Un email inexistente tarda lo mismo y responde igual que una contraseña incorrecta, y "recuperar"
  responde 202 siempre: la API no delata quién tiene cuenta.
- Las sesiones guardan el sha256 del token. El token se entrega una sola vez al BFF, que lo guarda en
  la cookie `__Host-mostrador_sesion` (httpOnly, Secure, SameSite=Lax); el navegador nunca habla con
  Go. "Visto" se actualiza como mucho una vez por hora.
- Cambiar la contraseña con el link (un solo uso, 30 min) cierra todas las sesiones de la cuenta.
- Límites de intentos con un upsert atómico (`auth.intentos`), como Caja: 5 ingresos por email y 30
  por IP cada 15 min; 3 links de recuperación por email por hora.
- La IP del visitante llega del BFF en `X-Cliente-IP` y Go sólo le cree si viene con la clave del
  BFF; si no, usa la IP de la conexión.
- Cuentas, membresías y sesiones viven en el esquema `auth` (una persona puede tener varias
  tiendas), fuera de RLS. Los slugs tomados se registran en `app.slugs`: con RLS forzada ni el dueño
  ve las tiendas ajenas para chequear si un slug está libre.
- En Workers, `fetch` guardado como propiedad y llamado como método da "Illegal invocation": los
  clientes HTTP lo envuelven en una función. Lo detectó el e2e contra la API real, no los unitarios.
