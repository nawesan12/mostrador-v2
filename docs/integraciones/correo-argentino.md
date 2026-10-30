# Correo Argentino (API MiCorreo)

Lo que usa Tu Local Web para que cada tienda envíe con Correo Argentino desde su cuenta de MiCorreo.
Código: `apps/api/internal/integraciones/correoargentino` (Go), migración `00047_correo_argentino.sql`.

## Fuentes

- Documentación oficial de la API MiCorreo (PDF de Correo, 8/8/2022):
  <https://www.correoargentino.com.ar/MiCorreo/public/img/pag/apiMiCorreo.pdf>
- Manual de la API Paq.ar 2.0 (otra API, para clientes con acuerdo comercial; abril 2023):
  <https://www.correoargentino.com.ar/MiCorreo/public/img/pag/apiPaqAr-v2.pdf>
- El plugin oficial "Correo Argentino for WooCommerce" (wordpress.org) usa las dos: MiCorreo con
  usuario y contraseña de la cuenta, o Paq.ar con "Acuerdo" y "Apikey" que da un ejecutivo de cuentas.
- Wrapper comunitario en TypeScript (no oficial, sirve de referencia de forma):
  <https://github.com/YamilEzequiel/correo-argentino>

Probado el 30/10/2026: `POST /token` sin credenciales válidas responde
`401 {"code":401,"message":"Unauthorized"}` tanto en `api` como en `apitest` (los dos ambientes
existen). No tenemos credenciales de API todavía, así que **todo lo demás se probó sólo contra el
MiCorreo falso de los tests**, armado con las respuestas del PDF.

## Ambientes

| Ambiente   | Base                                                  |
| ---------- | ----------------------------------------------------- |
| Pruebas    | `https://apitest.correoargentino.com.ar/micorreo/v1`  |
| Producción | `https://api.correoargentino.com.ar/micorreo/v1`      |

Sólo HTTPS. Las credenciales de API son **distintas por ambiente**. En Go:
`MOSTRADOR_CA_AMBIENTE=pruebas` usa apitest; vacío, producción.

## Autenticación (dos niveles)

1. **Credenciales de API** (usuario y contraseña para HTTP Basic). Las da Correo a quien integra
   ("deberán ser solicitadas a Correo para cada uno de los ambientes"). `POST /token -u user:pass`
   devuelve `{"token": "<JWT>", "expires": "2022-04-26 21:16:20"}` (hora de Argentina). El resto de
   los endpoints van con `Authorization: Bearer <JWT>`. Go guarda el JWT por credencial hasta 5 min
   antes de `expires` y, ante un 401, pide otro y reintenta una vez.
   - Las de Tu Local Web: `MOSTRADOR_CA_USUARIO` y `MOSTRADOR_CA_CLAVE` (sirven para todas las
     tiendas, como hace el plugin de WooCommerce).
   - Si Tu Local Web no las tiene, o la tienda prefiere usar las suyas, el modal de conexión las
     pide; van cifradas con `plataforma/cifrado` (contexto `correo:<tienda>`).
2. **Cuenta de MiCorreo de la tienda** (email y contraseña de micorreo.correoargentino.com.ar).
   `POST /users/validate {"email","password"}` devuelve `{"customerId": "0090000025", "createdAt"}`
   o `404 "Usuario no valido o inexistente"`. **Sólo se guarda el customerId**: la contraseña se usa
   para validar y se descarta. Todos los demás endpoints piden ese `customerId`.

## Endpoints

| Endpoint                | Para qué                                     | Uso en Tu Local Web                         |
| ----------------------- | -------------------------------------------- | ------------------------------------------- |
| `POST /token`           | JWT (Basic)                                  | Cliente, con caché                          |
| `POST /users/validate`  | customerId de una cuenta                     | Conectar                                    |
| `POST /register`        | Alta de cuenta nueva en MiCorreo             | No se usa (la tienda se registra en la web) |
| `GET /agencies`         | Sucursales de una provincia                  | Selector de sucursal del checkout           |
| `POST /rates`           | Cotización a domicilio (`D`) y sucursal (`S`)| Checkout y confirmación del pedido          |
| `POST /shipping/import` | Alta del envío en MiCorreo                   | "Generar envío" del pedido en el panel      |

### /rates

```json
{ "customerId": "0090000025", "postalCodeOrigin": "1757", "postalCodeDestination": "1704",
  "deliveredType": "D", "dimensions": { "weight": 2500, "height": 10, "width": 20, "length": 30 } }
```

- Sin `deliveredType` devuelve las dos (así lo pide Go: un solo pedido por CP).
- Peso en gramos (1 a 25.000), lados en cm (hasta 150), todos enteros.
- Respuesta: `{customerId, validTo, rates: [{deliveredType, productType, productName, price}]}`;
  `price` en pesos con centavos. El wrapper comunitario además tipa `deliveryTimeMin/Max` (plazo en
  días): Go los usa si vienen, como texto o número.
- Error: `402 "Cliente FAP no identificado {customerId}"`.

Tu Local Web redondea el precio **para arriba** al peso entero (los montos son pesos enteros) y guarda
lo cotizado 30 min (o hasta `validTo`). El precio que se cobra se vuelve a pedir en Go al confirmar la
compra: nunca sale del navegador. Si Correo no cotiza, la opción no aparece; al confirmar, la compra
falla con "No pudimos cotizar el envío" (409 `sin_cotizacion`). Nunca se muestra un precio
inventado: la tabla por región que había antes se borró.

Como los productos no tienen peso, la tienda carga un **paquete típico**: peso por unidad y medidas
de la caja. Peso del envío = peso por unidad × unidades (tope 25 kg).

### /agencies

`GET /agencies?customerId=…&provinceCode=B&services=pickup_availability`. Cada sucursal trae `code`,
`name`, `services.{packageReception, pickupAvailability}`, `location.address`, `hours` por día
(`{"start":"0930","end":"1800"}` o `null`) y `status` (`ACTIVE`). Go filtra activas que entregan,
ordena por cercanía de CP al de quien compra y guarda la lista 6 h. Error `402 "Customer ID no valido"`.

### /shipping/import

Campos obligatorios: `customerId`, `extOrderId` (único: MiCorreo rechaza importar dos veces el
mismo, "La orden ya fue importada con anterioridad"), `recipient.name`, `recipient.email`,
`shipping.deliveryType` (`D`/`S`), `shipping.agency` (sólo para `S`), `shipping.address.*` (sólo
para `D`: calle, altura, ciudad, provincia, CP; piso y depto se cortan a 3 letras),
`shipping.weight`, `shipping.declaredValue`, `shipping.height/length/width`. `sender` es opcional
(Tu Local Web manda el nombre de la tienda, su WhatsApp, el email de MiCorreo y el origen).
Respuesta: `{"createdAt": "…"}`, **sin número de seguimiento**.

Tu Local Web usa `extOrderId = TLW-<8 primeros de la tienda>-<número de pedido>`; si la respuesta se
corta y al reintentar MiCorreo dice que ya fue importada, se toma como hecho (no se duplica).

Errores 402 documentados: "Peso no valido", "Tipo de entrega invalido", "Verifique la sucursal de
destino", "El peso excede el maximo permitido para el producto", "no se encontro datos de remitente",
"El codigo Postal del emisor debe tener valor", "La provincia es invalida", "El alto/ancho/largo
debe estar entre 0 y 255". El panel muestra el mensaje de Correo tal cual ("Correo Argentino
respondió: …"); a quien compra nunca le llega.

### Provincias

Una letra (la del CPA): A Salta, B Buenos Aires, C CABA, D San Luis, E Entre Ríos, F La Rioja,
G Santiago del Estero, H Chaco, J San Juan, K Catamarca, L La Pampa, M Mendoza, N Misiones,
P Formosa, Q Neuquén, R Río Negro, S Santa Fe, T Tucumán, U Chubut, V Tierra del Fuego,
W Corrientes, X Córdoba, Y Jujuy, Z Santa Cruz.

### Límites

Correo documenta `429 Too Many Requests` con retroceso exponencial, sin número. Tu Local Web: las
rutas públicas de cotización y sucursales tienen un tope de 120 por hora por IP (`limite`), más la
caché de cotizaciones y sucursales.

## Seguimiento

**La API MiCorreo no tiene endpoint de seguimiento** (ni el PDF ni el wrapper lo documentan). El
envío importado queda en MiCorreo "para pagar": la tienda entra a MiCorreo, paga e imprime el rótulo,
y ahí recibe el número. En Tu Local Web lo carga en el pedido ("Número de seguimiento"), queda en
`ca_envios` y en `pedidos.seguimiento` (lo ve quien compra en la página de su pedido) y se enlaza al
seguimiento público de Correo (<https://www.correoargentino.com.ar/formularios/e-commerce>).

La consulta de historial por API (`GET /v1/tracking`) y el rótulo (`/v1/labels`) existen sólo en
**Paq.ar 2.0** (`https://api.correoargentino.com.ar/paqar/v1`, headers `Authorization: Apikey <key>`
y `agreement: <acuerdo>`), que pide un acuerdo comercial con Correo. Queda como paso siguiente si una
tienda trae acuerdo.

## Qué necesita el comerciante

1. Una cuenta de MiCorreo (<https://micorreo.correoargentino.com.ar>), dada de alta con el formulario
   de 2 pasos (DNI o CUIT), activa y con los datos de remitente completos.
2. Su email y contraseña de MiCorreo para conectar desde Integraciones › Correo Argentino.
3. Si Tu Local Web todavía no tiene credenciales de API (o la tienda quiere usar las suyas): pedirle a
   Correo Argentino el usuario y la contraseña de la API MiCorreo para producción (comercial /
   formulario de contacto de Correo) y cargarlos en el mismo modal.
4. El CP y la dirección desde donde despacha, y el peso y las medidas de un paquete típico.

## Qué necesita Tu Local Web

- Pedirle a Correo Argentino credenciales de la API MiCorreo para pruebas (apitest) y producción, y
  cargarlas en el VPS como `MOSTRADOR_CA_USUARIO` / `MOSTRADOR_CA_CLAVE` (nunca en el repo).
- `MOSTRADOR_CLAVE_CIFRADO` (ya está: la usan Mercado Pago y ARCA). Sin ella, Correo no se puede conectar.
- Con las credenciales de pruebas: conectar una cuenta de MiCorreo de prueba y verificar `/rates`,
  `/agencies` y `/shipping/import` de punta a punta (los tests sólo cubren el falso).
