# Handoff: Tu Local Web — app para iPhone

## Overview
App nativa de iOS para comerciantes que usan Tu Local Web (tienda online + punto de venta). Es "el panel en el bolsillo": ver cómo viene el día, preparar y cobrar pedidos, vender en el local (escáner, efectivo, QR, transferencia, link), gestionar productos/stock/clientes, hablar con **Tincho** (asistente IA) y recibir avisos. Idioma: español rioplatense (voseo). Moneda: ARS, formato `$ 184.350` (espacio fino después de `$`, punto de miles, coma decimal).

## About the Design Files
Los archivos de este paquete son **referencias de diseño hechas en HTML**: muestran el aspecto y comportamiento esperado, **no son código de producción**. La tarea es **recrearlos en SwiftUI** (iOS 17+, recomendado) con los patrones nativos: `NavigationStack`, `TabView`, `.sheet` con detents, `List`/`ScrollView`, `.swipeActions`, `.refreshable`, WidgetKit, ActivityKit, App Intents. Si ya existe una base de código iOS, seguir sus convenciones.

Abrí `Tu Local Web iPhone.dc.html` en un navegador: es un lienzo con zoom; cada pantalla tiene un ID visible (`1a`, `3b`, `12g`…) que se usa como referencia en este documento.

## Fidelity
**Alta fidelidad.** Colores, tipografía, espaciados, radios y copy son finales. Recrear pixel-perfect. Excepciones (placeholders): fotos de productos (rayado diagonal), logos de Mercado Pago / Mercado Libre / ARCA / Telegram, íconos de apps de terceros en la pantalla de inicio de iOS. Los íconos de trazo son referencia → usar **SF Symbols** equivalentes (ver tabla en "Assets").

Frame de referencia: iPhone 15, **393 × 852 pt**. También hay SE (375 × 667, `11i`), Pro Max (430 × 932, `11j`) y Dynamic Type AX (`11k`).

---

## Design Tokens

### Color — claro
| Token | Hex | Uso |
|---|---|---|
| `bg` | `#F7F1EA` | Fondo de pantalla |
| `bg-elevated` | `#FBF7F2` | Hojas, barras con blur, tab bar (al 92 %) |
| `surface` | `#FFFFFF` | Tarjetas, filas, campos |
| `text` | `#2A140C` | Texto principal, montos, chip de filtro activo |
| `text-2` | `#6B5244` | Texto secundario, labels de sección |
| `text-3` | `#7D6656` | Placeholders, íconos inactivos de tab bar |
| `chevron` | `#B5A193` | Chevrons de fila |
| `accent` | `#D9652F` | Línea de gráfico, foco de campo, cursor, punto de no leído |
| `primary` | gradiente `#C24E25 → #B0441D` (180°) | Botón principal, burbuja del usuario, FAB |
| `primary-text` | `#B0441D` | Links, botones de texto, back, tab activo |
| `primary-ink` | `#8E3A16` | Texto sobre `cream` |
| `cream` | `#F8E6D4` | Botón tonal, avatar de tienda, fondo de Tincho |
| `cream-2` | `#FBF3EA` | Tarjetas de Tincho, acciones rápidas, notas |
| `track` | `#F3E9DF` | Stepper, fondo de barra de meta (`#F3E6DA`) |
| `ok` | `#2F8F63` | Toggle on, checks, swipe "Cobrado" |
| `ok-bg` / `ok-ink` | `#E1F1E8` / `#1F6B48` | Chip Entregado/Retirado/Acreditado, deltas positivos |
| `warn` | `#C2861B` | Punto de aviso |
| `warn-bg` / `warn-ink` | `#FBEFD8` / `#7A4F08` (banner: `#6E4707`) | Chip Para preparar, banners de aviso |
| `error` | `#D2582F` | Borde de campo con error, insignias numéricas |
| `error-bg` / `error-ink` | `#FBE4DA` / `#A23F1C` | Chip Sin pagar, botón destructivo, deltas negativos |
| `neutral-chip` | `#EFE6DF` / `#5A4033` | Chip En camino |
| `muted-chip` | `#EEE8E3` / `#6B5244` | Chip Cancelado / En pausa |
| `line` | `rgba(58,29,18,.09)` | Separadores de fila |
| `line-2` | `rgba(58,29,18,.16)` | Bordes de campo, botón secundario, chip de filtro |
| `card-ring` | `rgba(58,29,18,.06)` | Borde 1 pt de tarjetas (sin sombra) |
| `scrim` | `rgba(42,20,12,.38–.45)` | Fondo detrás de hojas y alertas |
| `segmented-bg` | `rgba(58,29,18,.07)` | Fondo de control segmentado y buscador |

### Color — oscuro
| Token | Hex |
|---|---|
| `bg` | `#22140D` |
| `surface` | `#2E1B12` |
| `segmented-selected` | `#4A3125` |
| `text` | `#FBF7F2` |
| `text-2` | `#CDB8A8` |
| `text-3` / tab inactivo | `#A8907F` |
| `chevron` | `#7D6656` |
| `primary-text` / tab activo | `#E07A48` |
| `primary` gradiente | `#D9652F → #B0441D` |
| `ok` / texto ok | `#3FA877` / `#6FD0A0` sobre `rgba(63,168,119,.18)` |
| `warn` / texto warn | `#E3A93B` / `#F0C267` sobre `rgba(227,169,59,.18)` |
| texto error | `#F08A63` |
| `card-ring` | `rgba(251,247,242,.08)` |
| tab bar | `rgba(34,20,13,.92)` + blur |

Regla: los montos van siempre en `text`, nunca en acento. Texto sobre chips usa la variante `-ink` para pasar AA.

### Tipografía
Familia: **Instrument Sans** (Google Fonts, variable 400–700; incluir en el bundle). Montos con `.monospacedDigit()` (tabular-nums). Mapear a Dynamic Type con `relativeTo:`:

| Estilo | Tamaño/peso | Tracking | Dynamic Type |
|---|---|---|---|
| Monto héroe | 46–52 / 700 | −0,035 em | `.largeTitle` |
| Monto pantalla | 40 / 700 | −0,03 em | `.largeTitle` |
| Large Title | 34 / 700 | −0,025 em | `.largeTitle` |
| Saludo / título onboarding | 30–32 / 700 | −0,025 em | `.title` |
| Título hoja | 22–24 / 700 | −0,02 em | `.title2` |
| Título sección | 21 / 700 | −0,02 em | `.title3` |
| Indicador | 22 / 700 | −0,02 em | `.title3` |
| Headline / botón | 17 / 600 | 0 | `.headline` |
| Body | 16–17 / 400 | 0 | `.body` |
| Subhead | 15 / 400–600 | 0 | `.subheadline` |
| Label sección | 14 / 600, color `text-2` | 0 | `.footnote` |
| Footnote | 13 / 400 | 0 | `.footnote` |
| Caption / chip | 12 / 600 | 0 | `.caption` |
| Tab bar | 10,5 / 600 | 0 | `.caption2` |

Status bar y hora de pantalla bloqueada: SF Pro del sistema.

### Espaciado
Escala 4 · 8 · 12 · 16 · 20 · 24 · 32. Margen horizontal de pantalla **20 pt** (24 en Pro Max, 16 en SE). Gap entre bloques **14 pt**; entre tarjetas apiladas **10 pt**. Padding de tarjeta 16–18. Padding de fila 12–14 vertical × 14–16 horizontal. Separadores con inset que arranca donde arranca el texto (después del thumbnail/ícono).

### Radios
| Elemento | Radio |
|---|---|
| Esquina de pantalla (frame) | 56 |
| Hoja inferior | 30 superiores (hoja flotante del «+»: 38 todas) |
| Tarjeta héroe | 22 |
| Tarjeta / grupo de filas | 20 |
| Indicador chico | 18 |
| Botón grande / campo | 14 (botón 56 pt alto: 15) |
| Ícono de fila / thumbnail | 11–14 |
| Botón chico | 10–11 |
| Segmentado | 11 exterior / 8 interior |
| Chip / píldora | 999 |

### Alturas y tamaños
Botón principal 52 (o 56 en barras de acción). Botón chico 34–38. Campo 52. Buscador 40. Chip de filtro 34. Segmentado 36. Fila de lista 64–74. FAB 58 Ø a 20 pt del borde derecho y 16 pt sobre la tab bar. Tab bar 83 (49 + safe area). Navbar 44. Toggle 51 × 31. Área táctil mínima 44.

### Sombras
- Tarjetas: **sin sombra**, solo `card-ring` 1 pt.
- Botón principal: `0 6 16 −8 rgba(176,68,29,.7)`.
- FAB: `0 10 24 −8 rgba(176,68,29,.7)`.
- Toast: `0 12 30 −10 rgba(42,20,12,.55)`.
- QR sobre tarjeta: `0 20 50 −20 rgba(42,20,12,.35)`.
- Segmento seleccionado: `0 1 3 rgba(42,20,12,.12)`.

---

## Navegación global
- **TabView** con 5 tabs: Inicio (casa), Pedidos (recibo; insignia numérica = pedidos por preparar, fondo `error`, anillo 2 pt del color de fondo), Productos (caja), Tincho (avatar circular 27 pt; anillo 2 pt `primary-text` cuando está activo), Más (tres puntos). Ícono 27 pt, label 10,5/600. Activo `#B0441D`, inactivo `#7D6656`. Fondo `bg-elevated` 92 % + blur 20, borde superior 0,5 pt `line-2`.
- **FAB «+»** en Inicio y Productos → hoja (`2e`) con 3 opciones: Venta en el local (ícono sobre gradiente primary), Producto con la cámara, Pedido manual (íconos sobre `cream`).
- Navbar: back con chevron + label del padre en `primary-text` 17 pt; título centrado 17/600; acción derecha en `primary-text` (700 si es la confirmatoria, ej. "Guardar", "Listo").
- Hojas modales: grabber 36 × 5, `rgba(58,29,18,.2)`; scrim `rgba(42,20,12,.38)`.
- Deep links: `tulocalweb://pedido/{id}`, `tulocalweb://producto/{id}`; push, widgets, Live Activity y Siri abren la pantalla exacta.
- Lo que vive solo en la web (editor visual, dominio, invitar usuarios) se abre en `SFSafariViewController`.

---

## Screens / Views

### 01 · Primera vez
- **1a Splash** — fondo `bg`, isotipo 132 pt centrado, wordmark marrón 140 pt de ancho a 64 pt del borde inferior.
- **1b Ingreso con error** — isotipo 52, título "Entrá a tu tienda" (32/700), sub "Ventas, pedidos y stock, desde el mostrador." Campos Email y Contraseña (label 13/600 `text-2` encima, gap 6). Campo con error: borde 1,5 `error`, mensaje con círculo "!" 17 pt: "La contraseña no coincide. Te quedan 2 intentos antes del bloqueo." "Mostrar" dentro del campo. Link "Olvidé mi contraseña" alineado a la derecha. Botón "Entrar". Divisor "o". "Seguir con Apple" (negro) y "Seguir con Google" (blanco con borde). Pie: "¿Todavía no tenés tienda? **Creala en la web**". Haptic `.error` al fallar.
- **1c Bloqueado** — Tincho preocupado 200 pt, "Pausamos el ingreso por 15 minutos", explicación, píldora `warn-bg` con cuenta regresiva "Podés volver a probar en 14:32" (actualiza cada segundo), botones "Cambiar contraseña" / "Volver".
- **1d Selector de tienda** — hoja ~520 pt: "¿A qué tienda entrás?" + "Tenés 3 tiendas con esta cuenta." Lista con avatar de iniciales 46 pt radio 13, nombre, dominio · rol; radio de selección 24 pt (seleccionado: círculo `primary-text` con check blanco). Nota: "Después cambiás de tienda desde Más, tocando el nombre." Botón "Entrar a {tienda}".
- **1e–1g Permisos (pre-prompt)** — indicador de pasos (activo 22 × 5 `primary-text`, resto 8 × 5) + "Saltear". Tincho 210–240 pt, título 30/700, explicación, botón principal que dispara el diálogo del sistema + "Ahora no". 1e Notificaciones ("Te aviso cada venta al instante", con mock de push), 1f Cámara ("Fotos y códigos, con la cámara"), 1g Face ID ("Entrá con Face ID").
- **1h / 11n / 11o Recorrido (3 pasos)** — mitad superior terracota `#B0441D` con radial `#C24E25`, Tincho 300 pt apoyado en el borde; hoja inferior `bg-elevated` radio 30 con indicador, título, texto y botones "Saltear" + "Siguiente" (último: "Empezar"). Textos: "Mirá cómo venís, donde estés" · "Prepará y cobrá desde el mostrador" · "Preguntale a Tincho".

### 02 · Inicio
- **2a Inicio (scroll)** — Header: fecha 15 `text-2` + "Hola, {nombre}" 32/700; a la derecha campana 40 pt (punto `error` de no leído) y avatar de tienda 40 pt. Segmentado Hoy / 7 días / 30 días. **Tarjeta de ventas**: label, monto 46/700, chip delta (`ok-bg` "▲ 12,4 %") + "vs. ayer a esta hora", gráfico de línea 130 pt alto (hoy: `accent` 3 pt con relleno degradé 22 %→0 y punto final 6 pt con borde blanco 3; ayer: punteado `#8A7364` 2 pt, 4/5), eje 8 h–20 h, leyenda. **3 indicadores** en grilla (Pedidos, Ticket prom., Clientes nuevos) con delta coloreado. **Para hacer hoy**: filas con cuadro numérico 40 pt (warn/cream/error-bg según tipo) → pedidos por preparar, transferencias por confirmar, productos sin stock; cada una abre la lista filtrada. **Meta del mes**: monto 28/700 "de $ 3.000.000", barra 12 pt con gradiente `accent→primary` y proyección rayada detrás, "73 % vendido · Queda 1 día", proyección en texto; "Editar" → `12e`. **Ideas de Tincho**: carrusel horizontal de tarjetas 300 pt `cream-2` con borde `rgba(176,68,29,.16)` y botones. **Últimos pedidos** (5 filas + "Ver todos"). **Lo más vendido** (carrusel de 136 pt). **Canales**: barra apilada 14 pt (Web `#D9652F`, ML `#E3A93B`, Local `#6B5244`) + leyenda con monto y %.
- **2b Modo oscuro** — misma estructura con tokens oscuros.
- **2c Cargando** — skeletons con los mismos tamaños (`#EFE6DD` / `#F3ECE5`, shimmer horizontal 1,2 s).
- **2d Sin ventas** — tarjeta `cream-2` con Tincho saludando, "Tu tienda ya está online" + checklist (paso hecho tachado con check verde 28 pt; pasos pendientes con número en círculo; el siguiente con botón "Cargar").
- **2e Hoja «+»** — ver Navegación.

### 03 · Pedidos
- **3a Lista** — Large title "Pedidos" + botón de orden/filtro 40 pt. Buscador "Número, cliente o producto". Chips horizontales con contador (activo: fondo `text`, texto `bg-elevated`). Secciones por día ("Hoy", "Ayer"). Fila: thumbnail 48 pt radio 12, nombre 16/600, monto 16/600 tabular a la derecha, meta 13 `text-2` ("#1048 · Web · 14:32"), chip de estado. **Swipe a la izquierda**: "Más" (`#7D6656`) + acción principal del siguiente estado ("Listo", `#B0441D`), 84 pt c/u. **Swipe a la derecha**: "Cobrado" (`ok`) en pedidos sin pagar. Al completar: toast inferior oscuro con check verde y "Deshacer" (`#F3A57E`), 4 s. Haptic `.impact(.medium)`.
- **3b Detalle** — Navbar "#1048" con compartir y "…". Chip de estado, monto 40/700, "hoy, 14:32 · Tienda web · 5 productos". **Cliente**: avatar, nombre, "3.er pedido · clienta desde marzo", teléfono, mail, 3 acciones (Llamar, WhatsApp, Copiar mail) en `cream-2`. **Productos** con desglose: Subtotal, Cupón, Descuento por transferencia, Envío, **Total** 18/700 (la suma debe cerrar). **Pago** con chip y "Ver comprobante de la clienta". **Entrega**: dirección + timeline vertical (hecho: punto 14 pt `ok` con línea `ok`; actual: anillo 4 pt `accent`, texto 700; futuro: anillo 2 pt `line-2`, texto `text-3`). **Factura y comprobantes** ("Emitir factura B" tonal, descargas). **Notas internas** (`cream-2`, autor y hora). "Cancelar pedido" en `error-ink`. **Barra de acción fija** abajo con el siguiente estado; las acciones dependen del tipo de entrega (envío: despachar/en camino/entregado; retiro: listo para retirar/retirado).
- **3c Cancelar** — hoja grande: "¿Cancelás el pedido #1048?", aviso de reintegro manual si fue transferencia, motivos (single-select con check), toggle "Avisarle a {cliente} por mail", botón destructivo con ícono Face ID → autenticación biométrica antes de ejecutar.
- **3d Vacío con filtro** — Tincho con caja, "No hay pedidos en camino", "Ver todos los pedidos".
- **12f Error** — Tincho preocupado, "No pudimos traer los pedidos", "Reintentar" + "Ver estado del servicio".
- **12g Detalle oscuro**.

### 04 · Venta en el local
- **4a Venta rápida** — Cámara arriba (300 pt) con visor de esquinas blancas 4 pt y línea de escaneo `#E07A48` con glow; botones circulares cerrar / linterna. Hoja inferior que se superpone 20 pt: buscador, "Ticket · N productos" + "Vaciar", filas con stepper (34 pt, `track`), fila recién escaneada resaltada `#FFFBF6`. Descuento por medio de pago como chip + "Cambiar". Barra inferior: botón 56 pt "Cobrar" + total a la derecha. Escaneo exitoso: haptic `.selection` + sonido corto.
- **4b Efectivo** — monto 48/700 centrado; grilla 2 × 2 de medios (seleccionado: borde 2 pt `primary-text`); "¿Con cuánto paga?" con campo grande y atajos (Justo, billetes redondos); tarjeta `ok-bg` "Vuelto" 32/700. "Confirmar cobro".
- **4c QR Mercado Pago** — pensado para girar el teléfono al cliente: nombre de tienda, monto 52/700, QR 290 pt en tarjeta blanca radio 28 con isotipo al centro, "Escaneá con la app de Mercado Pago o tu banco", estado "Esperando el pago…" con punto `warn` pulsante. Al acreditarse (webhook/polling) pasa solo a `4e`. Subir brillo al máximo mientras se muestra.
- **4d Transferencia** — alias (26/700) y CVU con "Copiar", titular; "Compartir datos" + "Ya me transfirió".
- **12d Link de pago** — link con "Copiar", vencimiento 24 h, preview de burbuja de WhatsApp (`#DCF2D6`), "Mandar por WhatsApp" / "Otras apps".
- **4e Confirmada** — Tincho celebrando, "¡Venta cobrada!" (`ok-ink`), monto 52/700, detalle; accesos "Emitir factura" / "Compartir ticket"; "Nueva venta" + "Listo". Haptic `.success`.
- **4f Caja del día** — estado abierta, "Tendría que haber en efectivo" con fórmula, desglose por medio de pago, movimientos (retiros en `error-ink`), "Cerrar caja" + "Registrar retiro o ingreso".

### 05 · Productos y clientes
- **5a Lista (sin conexión)** — banner `warn-bg` "Sin conexión. Ves lo guardado hoy, 14:20." Large title + botón escanear. Buscador, chips (Todos 86, Publicados, Pausados, Sin stock 3). Fila: thumbnail 56, nombre, precio en `text` 700 · stock (bajo: `warn-ink` "Quedan 2"; cero: `error-ink` "Sin stock"), etiquetas de canal (Web / ML 11/700). Pausados al 75 % de opacidad.
- **5b Alta rápida** — grilla de fotos (portada 2 × 2 con etiqueta "Portada", reordenable por long-press con elevación y rotación −4°), slots "Sacar" (borde punteado `primary`) y "Elegir". Campos: Nombre, Precio | Costo, margen calculado en `ok-ink` ("Ganás $ 1.700 por unidad (35 %)"), Stock con stepper, Categoría. Segmentado Publicado / En pausa. "Guardar y publicar".
- **5c Variantes** — hasta 3 opciones con valores en píldoras + "+ Valor"; grilla de combinaciones (Variante · Stock · Precio · SKU), stock editable inline (0 en `error-bg`, bajo en `warn-bg`, foco con anillo `accent`).
- **5d Borrar bloqueado** — alerta: "Primero sacalo de Mercado Libre" con acciones "Abrir en Mercado Libre", "Pausar en la tienda", "Cancelar".
- **11c Edición completa** — descripción + "Mejorar con Tincho"; toggles por canal; vista previa de Google (URL verde, título azul `#1A3E9A`); Duplicar / Pausar / Borrar.
- **11l Escanear para encontrar** — visor verde al reconocer, hoja con el producto y acciones Ajustar stock / Editar / Vender.
- **11m Stock** — segmentado Ajuste / Ingreso / Conteo; stepper grande (52 pt) con valor relativo "+24", "Quedan 2 → van a quedar 26"; costo y proveedor.
- **5e Ficha de cliente** — avatar 76, nombre, antigüedad, etiquetas; acciones WhatsApp / Llamar / Mail; 4 indicadores; historial de pedidos.
- **12a Lista de clientes** — buscador, orden, filas con total comprado.

### 06 · Tincho (asistente IA)
- **6a Chat** — header con avatar 38, "Tincho" / "Conoce tu tienda al día". Burbuja usuario: gradiente primary, radio 20/20/6/20, máx. 260 pt. Burbuja Tincho: blanca, radio 20/20/20/6, avatar 28 a la izquierda. **Tarjetas accionables** dentro de la respuesta (lista de productos + pie `cream-2` con propuesta y botón "Hacelo" que ejecuta la acción tras confirmar). Respuesta en streaming con cursor `accent`. Sugerencias en chips arriba del input. Input 44 pt píldora + botón de micrófono (dictado).
- **6b Pensando / tope** — banner `warn-bg` "Te quedan 4 mensajes con IA este mes…" + "Ver planes" (aparece al quedar ≤10 %). Indicador de 3 puntos `accent` con opacidades 1/.6/.3 animadas. Botón enviar pasa a "detener" (cuadrado).
- **6c Sin conexión** — avatar en gris, "Sin conexión" en `error-ink`, Tincho desenchufado, "Reintentar".

### 07 · Notificaciones
- **7a Pantalla bloqueada** — widget de ventas en lock screen; push: título "Venta nueva · $ 24.020", cuerpo "Martina compró Yerba Lumbre suave 1 kg y 2 productos más." Agrupar por hilo (`threadIdentifier` = tipo).
- **7b Ampliada** — Notification Content Extension con lista de productos + acciones `UNNotificationAction`: "Ver pedido", "Marcar como preparado" (en background, requiere desbloqueo).
- **7c Centro de avisos** — agrupado por día, punto de color por tipo, punto `accent` de no leído, "Leí todo".
- **7d Preferencias** — toggles por tipo (venta, pago acreditado, transferencia por confirmar, preguntas ML, stock bajo, arrepentimientos, certificado ARCA) y horario de silencio (por defecto 22:00–8:00, avisos sin sonido).

### 08 · Más
- **8a Menú** — tarjeta de tienda (cambiar tienda, Ver tienda, Compartir, QR) y grupos: Ventas y clientes (Clientes, Caja del día, Carritos abandonados [insignia], Reseñas y preguntas [insignia], Devoluciones, Cupones), Plata (Finanzas, Facturación de ARCA), Tu cuenta (Plan, Integraciones, Usuarios y permisos, Perfil/seguridad/tema, Ayuda), "Cerrar sesión".
- **8b Finanzas** — cobrado del mes neto de comisiones, calendario de acreditaciones, ganancia con aviso "sin dato" (productos sin costo excluidos) + "Cargar costos".
- **8c Plan e integraciones** — tarjeta oscura de prueba gratis con barra y "Elegir plan" (pago en Mercado Pago, fuera de IAP — validar con App Review 3.1.3), estado de MP / ML / ARCA / Telegram.
- **11a Pedido manual**, **11b Devolución por arrepentimiento** (plazo legal 10 días visible), **11d Cupones**, **11e Carritos abandonados** ("Recordarle"), **11f Reseñas y preguntas** (respuesta con "Sugerir respuesta" de Tincho), **11g Usuarios** (solo lectura; invitar desde la web), **11h Cuenta** (Face ID, contraseña, sesiones, tema Claro/Oscuro/Sistema), **12b Facturación ARCA** (comprobantes C y notas de crédito con PDF), **12c Devoluciones**, **12e Editar meta** (hoja con atajos "+10 %" y promedio diario requerido).

### 09 · Funciones nativas de iOS
- **9a Widgets** — chico (ventas de hoy, delta, sparkline) y mediano (hoy + por preparar | último pedido). Fondo `bg-elevated`, radio del sistema. Refresh por timeline cada 15 min + push de WidgetKit al entrar venta.
- **9a/9b Live Activity** — activa mientras haya pedidos por preparar. Compacta: isotipo + número en `#E07A48`. Ampliada: contador 34/700, último pedido, botones "Ver pedido" / "Preparado" (App Intents). Lock screen: barra segmentada de progreso del día.
- **9c Siri y Atajos** — App Intents "¿Cuánto vendí hoy?" (respuesta con snippet view) y "Nueva venta" (abre `4a`).

### 11i–11k · Tamaños
- **SE**: margen 16, saludo 28, monto 38, indicadores 19; se ocultan gráfico y secciones secundarias del primer viewport.
- **Pro Max**: margen 24, filas 52 pt de thumbnail, texto 17.
- **AX**: filas de pedido apiladas (nombre / monto / chip / meta), sin truncar; tab bar crece a 100 pt; indicadores pasan a una columna.

---

## Interactions & Behavior
- **Pull-to-refresh** en Inicio, Pedidos, Productos, Clientes, Avisos.
- **Swipe** en filas de pedido (ver 3a) con acciones personalizadas de VoiceOver equivalentes.
- **Toasts**: abajo, sobre la tab bar, 52 pt, fondo `text`, 4 s, con "Deshacer" cuando aplica.
- **Hojas**: `.presentationDetents([.medium, .large])` salvo selector de tienda y meta (altura fija); cerrar con swipe.
- **Transiciones**: push estándar; `matchedGeometryEffect` del monto entre Cobrar (4b) y Confirmada (4e) recomendado. Respetar Reduce Motion.
- **Haptics**: venta confirmada `.success`; cambio de estado/swipe `.impact(.medium)`; error `.error`; stepper/selección/escaneo `.selection`.
- **Acciones sensibles** (cancelar pedido, cerrar caja, cambiar datos bancarios): Face ID / código.
- **Validación**: mensajes que dicen qué hacer ("Te quedan 2 intentos…"), no códigos. Precio > 0; costo opcional; stock entero ≥ 0.
- **Offline**: cache local (SwiftData) de pedidos/productos; banner con hora del último guardado; acciones de escritura en cola o deshabilitadas con explicación.
- **Estados** por pantalla: cargando (skeleton con forma real, `2c`), vacío (Tincho + acción, `3d`), error (Tincho + Reintentar, `12f`), sin conexión (`5a`, `6c`). Aplicar el mismo patrón a todas las listas.

## State Management
- Sesión: usuario, token (Keychain), tienda activa, lista de tiendas, rol/permisos (oculta acciones no permitidas).
- Inicio: rango (hoy/7/30), métricas, serie horaria hoy vs. ayer, pendientes, meta, ideas de Tincho, últimos pedidos, top productos, canales.
- Pedidos: filtro activo + contadores, búsqueda, lista paginada por fecha, detalle con timeline; mutaciones optimistas de estado con rollback y "Deshacer".
- Venta en local: ticket (items, cantidades), medio de pago, descuento asociado al medio, monto recibido → vuelto; QR/link con polling o socket de acreditación; caja (apertura, movimientos).
- Productos: filtros, búsqueda, borrador de alta (fotos locales hasta subir), variantes (matriz opciones × valores).
- Tincho: hilo de mensajes, streaming, cuota restante del plan, acciones propuestas pendientes de confirmación.
- Avisos: token APNs, preferencias por tipo, horario de silencio, no leídos.
- Live Activity: cantidad por preparar + último pedido (push token de ActivityKit).

## Assets
En `assets/`:
- `marca/mark.svg` (isotipo vectorial), `marca/isotipo-512.png` (ícono de app / avatar de notificación), `marca/wordmark-brown.png`, `marca/wordmark-cream.png`.
- `tincho/*.png` — ilustraciones del personaje (fondo transparente): `avatar`, `saludo`, `hablando`, `pensando`, `preocupado`, `celebrando`, `desenchufado`, `foto`, `seguridad`, `grafico`, `caja`, `envio`, `asomado`, `durmiendo`, etc. Uso por pantalla indicado arriba.

Íconos → SF Symbols sugeridos: Inicio `house`, Pedidos `doc.plaintext` / `receipt`, Productos `shippingbox`, Más `ellipsis`, campana `bell`, buscar `magnifyingglass`, escanear `barcode.viewfinder`, cámara `camera`, compartir `square.and.arrow.up`, llamar `phone`, WhatsApp/mensaje `message`, mail `envelope`, copiar `doc.on.doc`, efectivo `banknote`, transferencia `arrow.left.arrow.right`, QR `qrcode`, link `link`, check `checkmark`, Face ID `faceid`, linterna `bolt`, micrófono `mic`, sin conexión `wifi.slash`, ver tienda `arrow.up.right.square`. Logos de terceros: usar los oficiales según sus guías de marca.

## Files
- `Tu Local Web iPhone.dc.html` — todas las pantallas (secciones 01–12, IDs en badges). Abrir en navegador (necesita `support.js` y `assets/` al lado).
- `StatusBar.dc.html` — status bar de referencia (claro/oscuro, con o sin Dynamic Island).
- `TabBar.dc.html` — tab bar con estados activo, insignia y oscuro.
- `support.js` — runtime para visualizar los archivos HTML (no es parte de la app).
- `assets/` — marca e ilustraciones de Tincho.
