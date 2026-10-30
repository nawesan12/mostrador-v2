# Brief para Claude Design: app de Tu Local Web para iPhone

Diseñá la app nativa de **Tu Local Web** para iPhone: el panel del comerciante en el bolsillo. Tu Local Web es una plataforma argentina de tiendas online para comercios chicos (almacenes, ropa, dietéticas, ferreterías, cafés): tienda propia, cobros con Mercado Pago, facturas de ARCA, pedidos, stock y un asistente con IA que se llama **Tincho**. La web y el panel ya existen; esta app no reemplaza el panel, resuelve lo que el comerciante hace parado detrás del mostrador o en la calle, con una mano.

Entregá un handoff completo: pantallas en alta fidelidad, flujos conectados, todos los estados, componentes reutilizables y tokens. Se implementa en **SwiftUI** (iOS 17+), así que respetá las convenciones de iOS (Human Interface Guidelines): navegación con tab bar, sheets, pull to refresh, swipe actions, haptics, Dynamic Type y modo oscuro.

## Marca (respetala tal cual)

- **Nombre:** «Tu Local Web», siempre con esas mayúsculas y **nunca partido en dos renglones**. No mezclar en un mismo renglón texto blanco y naranja.
- **Isotipo:** una casita naranja con techo redondeado y, adentro, un bolso crema con manija marrón (`/static/mark.svg` del repo). Hay wordmark marrón y crema.
- **Tipografía:** Instrument Sans (variable). Números tabulares en montos.
- **Colores, tema claro:** fondo `#F7F1EA`, superficie `#FFFFFF`, texto `#2A140C`, texto secundario `#6B5244`, líneas `rgba(58,29,18,.09)` y `rgba(58,29,18,.16)`, acento naranja `#D9652F` / terracota `#B0441D` (botón principal en degradé `#C24E25 → #B0441D`), crema `#F8E6D4` / `#FBF3EA`.
- **Colores, tema oscuro:** fondo `#22140D`, superficie `#2E1B12`, texto `#FBF7F2`, secundario `#CDB8A8`, líneas `rgba(251,247,242,.08)` y `.15`. El acento se mantiene.
- **Estados:** verde `#2F8F63` / `#3FA877` (ok, cobrado), ámbar `#C2861B` / `#E3A93B` (atención), terracota `#D2582F` (error o caída).
- **Tincho:** la mascota es un hornero (pájaro argentino) ilustrado, con poses: saludo, hablando, pensando, celebrando, exito, preocupado, durmiendo, desenchufado, lupa, portapapeles, ticket, caja, envio, foto, grafico, whatsapp, seguridad, asomado, horno, horno-vacio, avatar. Usalas en vacíos, errores, onboarding y en el chat. Su avatar redondo es el ícono de Tincho.
- **Tono:** español rioplatense, cálido y directo, con voseo («Cargá», «Tocá», «Te avisamos»). Frases cortas. **Nada de renglones viudos** (ninguna palabra sola en el último renglón) ni cortes raros. Montos como `$ 15.215`. Fechas como «hoy, 14:32» o «jue 30 oct».

## Estructura

**Tab bar de 5:** Inicio · Pedidos · Productos · Tincho · Más.
- En «Pedidos», una insignia con los pedidos por preparar.
- El botón central o flotante «+» crea algo rápido: un producto con la cámara, un pedido o una venta en el local.

## Pantallas y flujos (no saltees ninguna)

### 1. Primera vez
- Splash con el isotipo.
- Ingreso: email y contraseña, «Seguir con Apple», «Seguir con Google», «Olvidé mi contraseña». Errores: clave incorrecta, sin conexión, cuenta bloqueada por intentos.
- Si la persona maneja varias tiendas, un selector de tienda.
- Permisos explicados antes del pedido del sistema, con Tincho: notificaciones («Te aviso cada venta al instante»), cámara (para fotos de productos y escanear códigos) y Face ID (para entrar rápido).
- Recorrido corto de tres pantallas: qué podés hacer desde el celular.

### 2. Inicio
- Saludo con el nombre de pila y la fecha.
- Ventas de hoy, comparadas con ayer: monto grande y variación en %. Selector Hoy / 7 días / 30 días.
- Gráfico simple de la serie acumulada contra el período anterior.
- Indicadores: pedidos, ticket promedio y clientes nuevos.
- «Para hacer hoy»: pedidos por preparar, transferencias por confirmar y productos sin stock. Cada uno lleva a su lista filtrada.
- Meta del mes: barra de progreso, lo vendido y la proyección al ritmo actual. Se puede editar.
- Los últimos 5 pedidos.
- Los más vendidos de la semana: foto, nombre y unidades.
- Canales: la parte de cada canal en las ventas (web, Mercado Libre, local).
- Ideas de Tincho: tarjetas con una sugerencia accionable, por ejemplo «3 carritos abandonados por $ 42.000: ¿les escribo?».
- Estado vacío de una tienda sin ventas, con Tincho y los primeros pasos.

### 3. Pedidos
- Lista con búsqueda y filtros por estado (Sin pagar, Para preparar, Listo para despachar, Listo para retirar, En camino, Entregado, Retirado, Cancelado), canal, pago, envío y fecha.
- Cada fila: número, cliente, total, estado en un chip, canal, la miniatura del primer producto y la hora.
- Swipe actions: avanzar de estado y marcar como cobrado.
- **Detalle:**
  - Cliente, con botones para llamar, escribir por WhatsApp y copiar el mail.
  - Productos, con variante, cantidad y precio.
  - Subtotal, cada descuento en su propia línea (cupón, transferencia), envío y total. La cuenta tiene que cerrar.
  - Pago (Mercado Pago o transferencia) y su estado.
  - Entrega (retiro en el local o envío con dirección) y la línea de tiempo de estados.
  - Acciones según el tipo:
    - Con envío: Preparar → Listo para despachar → En camino → Entregado.
    - Con retiro: Listo para retirar → Retirado.
    - Para los dos: cancelar, con motivo y confirmación.
  - Imprimir o compartir el comprobante y la etiqueta, desde la hoja de compartir.
  - Factura de ARCA: emitir, y ver o compartir el PDF.
  - Notas internas.
- **Crear pedido manual:** buscar productos, elegir variante, cantidad, cliente (nuevo o existente), pago, entrega y confirmar.
- **Devoluciones / arrepentimiento:** lista y detalle, con su código (ARR-2026-1) y las acciones.

### 4. Ventas en el local (punto de venta)
- Venta rápida: buscar o escanear un producto con la cámara (código de barras), armar el ticket, aplicar un descuento y cobrar.
- Cobro: efectivo (con el vuelto calculado), transferencia (mostrar alias y CBU para dictar o compartir), **QR de Mercado Pago** en pantalla grande para que el cliente escanee, y **link de pago** para mandar por WhatsApp.
- Confirmación de la venta, con la opción de emitir factura y compartir el comprobante.
- Caja del día: apertura, lo vendido por medio de pago y el cierre.

### 5. Productos
- Lista con búsqueda, filtros (publicados, pausados, sin stock, por categoría) y orden. Cada fila: foto, nombre, precio, stock y canales (web, Mercado Libre).
- **Alta rápida con la cámara:** sacar o elegir hasta 8 fotos (la primera es la portada, se reordenan arrastrando), nombre, precio, costo, stock y categoría. Guardar como publicado o en pausa.
- **Edición completa:** descripción, con «Mejorar con Tincho»; opciones y variantes (hasta 3 opciones, por ejemplo Talle, Color y Presentación), con una grilla de combinaciones para el stock, el precio y el SKU de cada una; dónde se publica (tienda web, Mercado Libre); cómo aparece en Google.
- Acciones: duplicar, pausar o publicar, y borrar con confirmación. Si el producto está en Mercado Libre, avisar que primero hay que sacarlo de ahí.
- **Stock:** ajuste rápido con + y −, ingreso de mercadería y conteo, y alerta de stock bajo.
- **Escanear un código** para encontrar el producto.

### 6. Clientes
- Lista con búsqueda.
- Ficha: datos, total comprado, pedidos, última compra, etiquetas y los botones de WhatsApp, llamar y mail.

### 7. Tincho (chat)
- Chat a pantalla completa con el avatar, burbujas, respuestas que se escriben en vivo e indicador de «Tincho está pensando».
- Sugerencias de preguntas en chips: «¿Cuánto vendí hoy?», «¿Qué se está por quedar sin stock?», «¿Quién no volvió a comprar?».
- Respuestas con tarjetas accionables (un pedido, un producto, un botón «Hacelo»).
- Entrada por voz con el dictado del sistema.
- Estado sin conexión, y el tope de mensajes con IA del plan.

### 8. Notificaciones
- **Push:** venta nueva (con monto y producto), pago acreditado, transferencia por confirmar, pregunta nueva en Mercado Libre, stock bajo, arrepentimiento recibido y vencimiento del certificado de ARCA.
  - Diseñá cómo se ven en la pantalla bloqueada, con la vista ampliada y acciones rápidas: «Ver pedido» y «Marcar como preparado».
- **Centro de notificaciones de la app:** lista agrupada por día, leídas y sin leer.
- **Preferencias:** qué avisos llegan al celular, por tipo, y un horario de silencio.

### 9. Más
- Mi tienda: ver la tienda (abre la web), compartir el link y el QR de la tienda.
- Finanzas resumidas: lo cobrado, lo que se acredita y la ganancia (con «sin dato» cuando falta el costo de algún producto).
- Facturación de ARCA: lista de comprobantes.
- Cupones: lista, crear y pausar.
- Carritos abandonados, con la acción de recordar.
- Reseñas y preguntas: responder.
- Integraciones: el estado de Mercado Pago, Mercado Libre, ARCA y Telegram. Lo que todavía no conecta dice «Próximamente».
- Plan y facturación: el plan actual o la prueba gratis con los días que quedan, y el cambio de plan, que abre Mercado Pago.
- Usuarios y permisos, en solo lectura.
- Cuenta: perfil, seguridad (Face ID, cambiar clave, cerrar otras sesiones), tema claro/oscuro/sistema, ayuda (artículos y contacto con soporte) y **cerrar sesión**.

### 10. Funciones nativas de iOS
- **Widgets:**
  - Pantalla de inicio en tamaño chico (ventas de hoy) y mediano (ventas de hoy, pedidos por preparar y el último pedido).
  - Pantalla bloqueada: ventas de hoy.
- **Live Activity / Dynamic Island:** mientras hay pedidos por preparar, el contador y el último que entró.
- **Atajos y Siri:** «¿Cuánto vendí hoy?» y «Nueva venta».
- **Face ID** para entrar, y para confirmar acciones sensibles como cancelar un pedido.
- **Haptics:** venta confirmada, cambio de estado y error.
- **Hoja de compartir:** link de la tienda, link de pago, comprobante y producto.
- **Enlaces profundos:** tocar un push abre ese pedido o producto.

## Estados obligatorios en cada pantalla
- Cargando, con esqueletos.
- Vacío, con Tincho y una acción clara.
- Error, con reintentar.
- Sin conexión: un banner, y se ve lo último guardado.
- Éxito, con un toast o haptic.
- Confirmación destructiva.
- Además: modo claro y oscuro, Dynamic Type (de chico a AX grande, sin cortes), VoiceOver (etiquetas de los íconos) y contraste AA.

## Tamaños
- iPhone SE / 13 mini (375 × 667 y 375 × 812), iPhone 15 (393 × 852) y iPhone 15 Pro Max (430 × 932).
- Respetá el notch, la Dynamic Island y el home indicator (safe areas).

## Componentes y tokens a entregar
Botones (principal, secundario, destructivo, de texto), chips de estado, celdas de lista con swipe, tarjetas, campos de texto, selectores, stepper de stock, montos, gráfico, barra de progreso, tab bar, navigation bar, sheets, toasts, banners, avatares de cliente, burbuja de Tincho, insignias y vacíos con ilustración. Tokens de color, tipografía (escala con Dynamic Type), espaciado, radios y sombras, en claro y oscuro.

## Datos de ejemplo
Usá un comercio argentino verosímil para los mocks, por ejemplo «Almacén Lumbre»: yerba, alfajores, dulce de leche, pedidos de $ 8.000 a $ 60.000 y clientes con nombres comunes. Son sólo ilustración del diseño; la app muestra siempre los datos reales de la tienda.

## Fuera de alcance
El editor visual de la tienda (plantillas, colores, dominio) y la migración desde otras plataformas quedan en la web. La app lleva a la web cuando hace falta.
