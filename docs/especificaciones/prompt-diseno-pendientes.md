Seguimos con Mostrador (mismo sistema de diseño de este proyecto: tokens de color, Instrument Sans/Serif, radios, arco como motivo, grano, Hornero y sus sprites, íconos Phosphor). Necesito el diseño de tres cosas que todavía no existen, en el mismo formato y con el mismo nivel de detalle que PanelOperacion / PanelCuenta: escritorio (1440) y celular (375) para cada pantalla, con sus estados (vacío, cargando, error, éxito) y los modales que abran. Textos reales en español rioplatense, datos de ejemplo verosímiles, precios vigentes: Gratis / Emprende $ 19.900 / Pro $ 49.900 / Negocio $ 149.900 por mes, anual −25 %, prueba de 14 días, 0 % de comisión. Dominio: mostradorapp.com.ar.

────────────────────────────────

1. BACKOFFICE DEL EQUIPO DE MOSTRADOR (uso interno, /staff)

Quién lo usa: el equipo de Mostrador (soporte, ventas, administración). No lo ve ninguna tienda. Tiene que verse "Mostrador" pero claramente interno (por ejemplo, una barra superior oscura con "Equipo" y el nombre de quien está logueado).

Roles: Soporte (lee tiendas y responde), Administración (además ve facturación y aprueba partners), Dueño (todo, incluido dar de baja). Mostrá cómo cambia la navegación según el rol.

Pantallas:

- Inicio: tiendas activas, nuevas en la semana, en prueba que vencen en 3 días, ingresos del mes (MRR), bajas del mes con motivo, consultas sin responder, postulaciones de partners nuevas, estado de los servicios (API, pagos, ARCA, envíos), y alertas (tiendas con pagos rechazados, certificados ARCA por vencer, picos de errores).
- Tiendas: lista con búsqueda y filtros (plan, estado de la suscripción, rubro, plantilla, partner que la trajo, fecha de alta), y la ficha de una tienda: datos del negocio y del titular, plan y suscripción de Mercado Pago (próximo cobro, cobros fallidos), uso (productos, pedidos del mes, respuestas de Hornero contra el tope), integraciones conectadas y su salud, dominio propio, actividad reciente, notas internas del equipo, y acciones: extender prueba, cambiar plan, aplicar crédito, regalar pack de IA, ver la tienda como la ve la dueña (modo lectura, con un aviso fijo en pantalla), pausar o cerrar la tienda (con confirmación fuerte y motivo).
- Consultas: bandeja de lo que llega por el formulario de contacto, el de arrepentimiento (con su código ARR y el plazo legal visible) y reclamos de consumidor: estados nueva / en curso / resuelta, asignar a alguien, responder por mail desde ahí.
- Partners: postulaciones (nombre, agencia, web, especialidades: diseño, desarrollo, marketing, contabilidad, fotografía) con aprobar / rechazar y motivo; partners activos con su nivel, tiendas referidas y comisiones a pagar.
- Facturación: cobros del mes, comprobantes emitidos por ARCA (factura de Mostrador a cada tienda), reintentos fallidos, reembolsos.
- Auditoría: quién del equipo hizo qué y cuándo (incluido "entró a ver la tienda X"), con filtros.

──────────────────────────────── 2) PANEL DE PARTNERS (/partners/panel, para agencias y freelancers aprobados)

Programa (ya publicado en la landing de Partners): niveles Partner (20 % de comisión recurrente por 12 meses, cuenta de prueba ilimitada, materiales de venta, sin requisitos), Plata (25 % recurrente) y Oro (30 % recurrente); el ejemplo de la landing es "12 tiendas en plan Pro, nivel Plata, 25 % sobre $ 49.900 = $ 149.700 por mes, mientras sigan activas". Los partners Oro aparecen en "Buscar un partner".

Pantallas:

- Inicio: comisión estimada del mes, próximo pago y fecha, tiendas activas referidas, en prueba, perdidas; barra de progreso al siguiente nivel con el requisito que falta; tareas (ej. "3 tiendas tuyas vencen la prueba esta semana").
- Mis tiendas: lista de las tiendas que trajo (plan, estado, desde cuándo, comisión que genera), alta de una tienda nueva para un cliente (invitación por mail a la dueña, o la crea el partner y después la transfiere), y acceso a las tiendas donde la dueña le dio permiso de colaborador.
- Link y cupón de referido: su link único, un QR para descargar, código de cupón propio, cuántos clics, altas y conversiones.
- Comisiones y pagos: detalle mes a mes por tienda, estado (pendiente, aprobada, pagada), datos de cobro (CBU/CVU o Mercado Pago, CUIT, condición frente al IVA) y las facturas que el partner sube por cada pago.
- Perfil público: lo que se ve en "Buscar un partner" (logo, nombre, ciudad, especialidades, casos con link a tiendas, contacto), con vista previa.
- Materiales: piezas de marca listas para usar (las de Redes e Imprimibles), con su link de referido ya puesto.

──────────────────────────────── 3) GENERADOR DE LA TARJETA A6 PARA PAQUETES (dentro del panel de la tienda)

Base: la pieza 7a de Imprimibles ("Tarjeta para paquetes · A6 105 × 148 mm · la imprimen las tiendas desde el panel"; frente y dorso, con la marca de la tienda, "¡Gracias, Lucía! Lo preparamos a mano el 20 de octubre. Que lo disfrutes." y "¿Te gustó?").

Qué tiene que permitir:

- Elegir modelo (2 o 3 estilos de frente) y ver frente y dorso en vivo, a tamaño real escalado.
- Usar automáticamente el logo, los colores y la tipografía de la plantilla de la tienda.
- Texto con variables que se completan por pedido: nombre del cliente, fecha, productos, número de pedido; con ejemplos precargados y un límite de caracteres visible.
- Dorso con un QR a elegir: dejar una reseña, volver a comprar con un cupón (elegir uno de los cupones de la tienda), seguir en Instagram, o WhatsApp de la tienda.
- Dónde se usa: desde un pedido ("Imprimir tarjeta"), en lote desde la lista de pedidos seleccionados (varias por hoja A4, con marcas de corte), y una opción "incluir siempre al imprimir la etiqueta de envío".
- Salida: PDF A6 para imprimir en casa o 4 por A4 con marcas de corte; aviso de márgenes seguros y sangrado.
- Estado vacío cuando la tienda todavía no subió logo, y cómo se ve en el celular (vista previa arriba, ajustes abajo).

────────────────────────────────
Para las tres: incluí las variantes de celular completas (navegación, tablas que pasan a tarjetas, acciones en hoja inferior), estados vacíos con el sprite de Hornero que corresponda, y nada de textos con palabras sueltas en el último renglón.
