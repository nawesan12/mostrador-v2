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

## 14. Legales y botón de arrepentimiento (PR #12)

- Los 8 legales salen de `design/Legal *.dc.html` con `scripts/legales.ts` (genera `src/lib/legales.ts`):
  el texto no se copia a mano, y el HTML queda sólo con `<b>` y `<a>` (links reescritos a rutas).
- Las páginas públicas comparten el layout `(sitio)`: header y footer del handoff y el banner de cookies.
  En el celular la nav del header baja a una fila deslizable y el índice lateral de los legales
  (sticky en escritorio) pasa a chips + un desplegable: con poco ancho, el sticky tapaba el texto.
- Botón de arrepentimiento (Res. 424/2020): con la API, `POST /v1/arrepentimientos` lo registra en
  `soporte.arrepentimientos` con código `ARR-AAAA-NNNNNN`, límite por IP y por email, y mail de
  confirmación. La app sólo puede insertar y leer el código; los datos personales los ve staff.
  En la demo no hay a quién revocarle nada: se valida igual y se da un código correlativo (tabla
  `contadores` de D1), sin guardar ningún dato del pedido.

## 15. Formularios públicos (PR #14)

- Contacto, postulación de partners y novedades: dominio Go `formularios` (tablas en `soporte`),
  con límite por IP y por email, mail de confirmación a quien escribe y aviso a la bandeja del
  equipo (`MOSTRADOR_CORREO_EQUIPO`, opcional; también recibe los arrepentimientos).
- La app sólo inserta y lee el id: las consultas y postulaciones las lee staff. Por eso la
  postulación repetida se detecta por el error de unicidad y no con `ON CONFLICT` (que exigiría leer
  los emails).
- La baja de novedades es un POST con el token del mail: abrir el link sólo muestra el botón, así los
  clientes de correo que previsualizan links no desuscriben a nadie.
- En la demo los formularios validan igual que Go y muestran el éxito, sin guardar ni mandar nada.
- Los montos usan espacio duro entre "$" y el número (nunca se parten en dos renglones).

## 16. /estado con monitor propio (PR #16)

- El monitor corre en `mostrador-jobs` (el cron horario que ya existía): un chequeo por componente,
  unos 170 pedidos por día. Vive fuera del sistema que mide y no necesita UptimeRobot ni el plan
  Pro de Cloudflare.
- Guarda en su propia base D1 (`mostrador-estado`): resumen por día, último resultado e incidentes.
  El sitio sólo la lee, también en modo API, porque el estado tiene que verse aunque lo demás falle.
- Al propio sitio lo chequea por service binding (un worker no puede pedirle por URL pública a otro
  de la misma zona); a los proveedores, que respondan (sin credenciales, un 401 también dice "arriba").
  La API Go se chequea sólo si `API_URL` está configurada.
- Un incidente se abre con dos fallas seguidas y se cierra solo cuando vuelve; los días sin chequeos
  se ven grises ("Sin datos"), nunca verdes inventados.
- Avisos a suscriptos por mail con Resend si `mostrador-jobs` tiene `RESEND_API_KEY`; por WhatsApp,
  cuando esté WhatsApp Cloud. La baja es un POST con token, como en novedades.

## 17. Onboarding (PR #17)

- `/onboarding` es el registro (decisión §0): todos los "Crear mi tienda" de la landing y del sitio
  llevan ahí, y "Usar esta plantilla" lo arranca con esa plantilla (`ob-tpl`, como en el handoff).
- Lo elegido queda en el navegador (`ob-estado`) para que recargar no pierda nada. En la demo, al
  terminar se crea el sandbox con la plantilla elegida y lo demás se guarda como cambio del sandbox.
  Con la API, el paso 1 pasa por `/ingresar` (que al registrar vuelve a `?paso=1`) y cada paso se
  guarda con `PUT /v1/panel/onboarding` (columna `tiendas.onboarding`, jsonb hasta 16 KB).
- Los datos que muestran los pasos son los del negocio de ejemplo del handoff (Café La Esquina);
  traerlos de verdad de cada integración llega con el PR de esa integración.

## 18. Ingreso con Google, Mercado Pago, Mercado Libre y WhatsApp (PR #18)

- OAuth 2 con PKCE (S256) en la API: Go guarda el state (hasheado, un uso, 10 min) y el
  verificador; el BFF sólo redirige y, en `/ingresar/<proveedor>/vuelta`, canjea y guarda la
  sesión. Una cuenta nueva sigue al onboarding (tienda provisoria "La tienda de <nombre>").
- Vinculación por email sólo si el proveedor lo verifica (Google). Mercado Pago y Mercado Libre no
  garantizan el email: si ya hay una cuenta con ese email no se vinculan solas (error que explica
  cómo hacerlo desde Mis datos). Así nadie entra a una cuenta ajena creando otra en un proveedor.
- WhatsApp: link para entrar sin contraseña (15 min, un uso) y link de contraseña nueva, al número
  de la cuenta. Sin pantalla de código: el handoff no la tiene y sí tiene "Revisá tu WhatsApp". El
  link pide un clic para entrar porque WhatsApp abre los links para armar la vista previa. Responde
  igual haya o no cuenta con ese número.
- Cada proveedor se prende al cargar sus credenciales; apagado, el botón lo explica. En desarrollo
  los mensajes de WhatsApp van al log.

## 19. Los datos de cada demo (PR #19)

- Cada demo es la tienda de su plantilla, con su catálogo del handoff: `scripts/catalogos.ts` genera
  `src/lib/catalogos.ts` desde `Plantillas Tienda` (12 tiendas) y las plantillas trabajadas (Ceibo,
  Mostaza, Yerbal, Tijera). Las categorías de cada producto se asignan en el generador (el handoff
  da las de la tienda, no en cuál va cada uno).
- Clientes y pedidos de 30 días se generan con semilla por plantilla (`sandbox/demo.ts`): siempre
  los mismos, fechados desde que se creó el sandbox. No se guardan: D1 sólo guarda lo que cambia el
  visitante. Canales, pagos, envíos y estados son los del panel del handoff.
- `GET /v1/panel/resumen` (insignias, notificaciones, canales y prueba) está en el contrato. En la
  demo sale de esos datos; en Go, la prueba es real y lo demás va en cero hasta que existan los
  dominios (pedidos, carritos, WhatsApp, tareas).

## 20. Dominio: mostradorapp.com.ar (PR #52)

- mostrador.com.ar está registrado por un tercero (vence a mediados de diciembre). La marca sigue
  siendo Mostrador y el dominio será **mostradorapp.com.ar**: los mails públicos (ayuda@,
  legales@, seguridad@…) ya lo usan. Hasta que esté comprado y con DNS, el sitio sigue en
  mostrador.iamnaw.dev (`PUBLIC_SITE_URL` / `MOSTRADOR_SITIO`).
- Al pasarlo: cambiar esas dos variables, la ruta de `wrangler.jsonc`, y en los portales las vueltas
  de OAuth (`/integraciones/mercadopago/callback`, `/integraciones/mercadolibre/callback`,
  `/ingresar/*/vuelta`) y las notificaciones (`api.mostradorapp.com.ar/v1/webhooks/*`). Mercado
  Libre pide validar con QR desde el celular de Nahuel para editar la app.

## 21. Manejá Mostrador desde tu IA: servidor MCP con OAuth propio (PR #54)

- Mismo patrón que Caja: Mostrador es su propio servidor de autorización (OAuth 2.1 para clientes
  MCP) y expone `/api/mcp` sin estado. Registro dinámico (clientes públicos, vueltas https o
  localhost, 20 por hora por IP), PKCE S256 obligatorio, código de un solo uso (5 min), access token
  de 1 h y refresh token que rota en cada uso; si llega uno ya rotado, se corta la conexión entera.
  Códigos y tokens se guardan hasheados (sha256).
- En la demo todo vive en la D1 del sandbox (migración 0005): la conexión es del sandbox del
  visitante (el de su cookie al dar permiso), vence con él y se borra con él (`ON DELETE CASCADE`).
  El cron limpia códigos vencidos, clientes de más de 30 días sin conexiones y los límites.
- Una fila por conexión con el access token, el refresh token y la ventana del límite (60 pedidos
  por minuto): autenticar y contar es una sola escritura (`UPDATE … RETURNING`).
- Permisos: `lectura` (siempre) y `acciones` (si la IA lo pide y la persona lo deja marcado). Las
  acciones además respetan los interruptores de Integraciones: ajustar productos pide "Cargar y
  editar productos" y crear cupones, "Reembolsos y descuentos".
- Las herramientas trabajan sobre `Datos`, sin métodos nuevos: cuando exista la API Go, la misma
  lista sirve contra `DatosApi`. Montos en pesos enteros, con el texto formateado al lado.
- Con `DATA_SOURCE=api` los endpoints responden 501 con un mensaje claro: la versión real (atada a la
  cuenta y la tienda, no al sandbox) la va a atender Go.
- `/api/mcp` queda fuera del chequeo de CSRF (no usa cookies, sólo el token Bearer), como los
  endpoints de OAuth. La pantalla de consentimiento sí lo pasa.

## 22. El MCP en producción: OAuth en Go y el token de la IA como sesión limitada (PR #58)

- Con `DATA_SOURCE=api`, clientes, códigos y conexiones viven en Postgres (`auth.ia_*`, sólo hashes).
  La conexión es de una persona en una tienda: si le sacan la membresía, el token deja de servir.
- El token de acceso entra por el mismo `Autenticar` que las sesiones (no hay otra puerta), pero
  `ia.Guardia` lo deja pasar sólo a las rutas que usan las herramientas (lista explícita), con GET
  si la conexión es de lectura, y controla en el servidor los interruptores del panel ("Cargar y
  editar productos", "Reembolsos y descuentos"): el BFF también los mira, pero no es el que manda.
  Cuenta, usuarios, plan, sesiones y cobros quedan afuera aunque quien autorizó sea la dueña.
- Sólo dueña o admin autorizan. Refresco rotativo: si llega uno ya usado, se borra la conexión.
- El BFF sigue siendo el que habla con las IA (nombres de OAuth, `/api/mcp` JSON-RPC); Go sólo ve
  su propio contrato (`/v1/ia/*`, `/v1/panel/ia*`).
- Goose corre con migraciones fuera de orden: varias ramas suman migraciones a la vez.

## 23. Turnos con seña (PR #59)

- Modelo: `turnos_ajustes` (una fila por tienda; sin fila, apagado con seña 30 %, 60 min de
  anticipación y 14 días de vista), `turnos_servicios`, `turnos_profesionales` y `turnos`, todas con
  `app.aplicar_rls`. Cada turno copia el nombre del servicio y del profesional (no son FK): borrar
  uno no deja la agenda en blanco. El turno tiene un `codigo_publico` aleatorio para su página
  (`/t/{slug}/turno/{codigo}`), como los pedidos de la tienda online.
- Disponibilidad (`turnos.Libres` en Go, `lib/turnos/libres.ts` en la web): grilla de 15 min desde
  el comienzo de cada franja, el servicio tiene que entrar entero en una franja, el profesional lo
  tiene que hacer (lista vacía = todos) y no tener otro turno que se superponga. Los días que se
  ofrecen son hoy y los `diasVista − 1` siguientes, y nada antes de `ahora + anticipacionMin`.
  "Cualquiera" sirve si algún profesional está libre; al reservar se asigna el primero libre por
  `orden`. Las dos implementaciones corren la misma tabla de casos
  (`apps/api/internal/dominio/turnos/testdata/libres.json`).
- Sin cron para liberar horarios: un turno `reservado` sin seña pagada deja de bloquear cuando pasa
  su `venceEn` (15 min). La regla está en la disponibilidad, no en un proceso que tenga que correr;
  la fila queda `reservado` y la API la muestra con `vencido: true`. La preferencia de Mercado Pago
  vence a la misma hora, así nadie paga una seña de un horario que ya se liberó.
- Dos reservas al mismo horario: la reserva bloquea los profesionales activos de la tienda
  (`SELECT … FOR UPDATE`, siempre por id para no trabarse) y recién ahí mira los turnos e inserta,
  todo en la misma transacción. Una restricción de exclusión no alcanza porque "vencido" depende de
  la hora actual. El test concurrente lo prueba (10 a la vez: gana uno; sin el bloqueo, fallaba).
- Seña: porcentaje del precio del servicio, redondeado al peso. Precio, duración y seña salen del
  servidor. Con seña > 0 y Mercado Pago conectado, el turno queda `reservado` y se devuelve el
  `init_point` (Checkout Pro con la cuenta de la tienda, `mp_cobros.tipo = 'turno'`, sin cuotas).
  El webhook aprobado lo pasa a `confirmado` con `senaPagada`. Sin Mercado Pago o con seña 0, queda
  `confirmado` y sin seña. Si Mercado Pago no responde al armar el pago, el turno se cancela en el
  momento para no tomar el horario.
- Cancelar (la persona desde su página, hasta `anticipacionMin` antes; la tienda desde el panel)
  nunca devuelve la seña sola: la tienda la devuelve con "Devolver la seña" (refund del pago en
  Mercado Pago, sólo dueña o admin), y el turno queda con `senaDevuelta`.
- El turno a mano del panel no tiene seña ni mira la grilla; si se superpone con otro, pide
  confirmar (`forzar`). Turno nuevo y cancelado van al outbox (`turno.nuevo`, `turno.cancelado`)
  para los avisos.
## 24. Cupones en la tienda online (PR #65)

- Una sola regla en TS (`lib/tienda/cupon.ts`) y en Go (`vidriera/cupon.go`), con los mismos casos
  en `vidriera/testdata/cupones.json`. El navegador la usa para mostrar el descuento; al comprar,
  el servidor la vuelve a correr y nunca se cree el total del navegador.
- El uso se cuenta en la misma transacción que crea el pedido (el pedido guarda `cupon` y el cupón
  suma usos, vendido y descontado, migración 00028), así el límite de usos no se pasa.
- "VIP" = quien tiene algún pedido VIP con el mismo email; "en riesgo" = su última compra fue hace
  más de 60 días. Los dos se calculan en el servidor a partir de los pedidos.
- El cupón por categoría guarda la categoría (antes se perdía); si no quedó guardada, vale para
  toda la tienda.
- La vista previa del cupón (`/t/[tienda]/cupon`) tiene un tope de 30 pruebas por IP cada 10 min,
  con el limitador del #54, para que no se adivinen códigos probando.
## 25. Medición de la tienda: consentimiento primero y sin API de conversiones todavía (PR #67)

- Google Shopping, Google Analytics y Meta Ads usan las cuentas de la tienda, no credenciales de
  Mostrador: la tienda pega sus IDs (GA4 y píxel) y carga el feed `/t/{slug}/feed.xml` en Merchant
  Center y en Commerce Manager. "Conectada" es el ID guardado o el "Ya lo cargué" del feed: no hay
  API que confirme que Google o Meta lo leyeron.
- gtag.js y fbevents.js se cargan sólo si la tienda tiene IDs y quien compra aceptó las cookies de
  medición en el aviso de esa tienda (se recuerda por tienda en el navegador). Sin elección no se
  pide nada a Google ni a Meta; la vitrina de una plantilla no mide nunca. Un solo módulo
  (`lib/tienda/medicion.ts`) manda los eventos desde el layout, las rutas y el carrito compartido:
  las plantillas no se tocan.
- La compra se mide en la página del pedido con el número de pedido como `transaction_id` (GA4) y
  `eventID` (Meta), para que deduplique cuando llegue la API de conversiones.
- La API de conversiones de Meta queda para otro PR: necesita un token de acceso de la tienda
  guardado cifrado (como Mercado Pago) y mandar las compras desde Go con los datos de quien compra
  hasheados. Hasta entonces, Configurar la nombra como "Próximamente" sin prometer fecha.
- La CSP se suma sólo en `/t/`: scripts y conexiones al propio sitio, a Google y a Meta. El resto
  del sitio sigue sin CSP hasta el PR de endurecimiento.
