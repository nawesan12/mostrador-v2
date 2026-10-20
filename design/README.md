# Handoff: Mostrador — plataforma de e-commerce para tiendas argentinas

## Overview
Mostrador es una plataforma para crear y operar tiendas online (competencia directa de Tiendanube y Empretienda). Este paquete cubre todo el diseño producido:

1. **Landing / sitio corporativo** (home + 19 páginas públicas: legales, empresa, ayuda, login, 404).
2. **Onboarding** de alta de cuenta que precarga la tienda (productos, marca, pagos, envíos, canales) para que el cliente no arranque en cero.
3. **Panel de administración** (sidebar + dashboard + ~30 pantallas: pedidos, catálogo, cupones, marketing, finanzas, facturación, envíos, cuenta, etc.), con tema claro/oscuro y modo Simple/Completo.
4. **Plantillas de tienda** (storefronts de ejemplo por rubro).
5. **Marca**: papelería, piezas de redes, imprimibles, imágenes OG.
6. **Hornero**: mascota/asistente de IA con sprites por estado.
7. **Integración con Caja** (app hermana de finanzas: https://caja.iamnaw.dev).

Repo de destino: `nawesan12/mostrador` (ver `github.md` en la raíz del proyecto original).

## About the Design Files
Los archivos de este paquete son **referencias de diseño hechas en HTML**: prototipos que muestran el aspecto y el comportamiento buscados, **no código de producción para copiar**. La tarea es **recrear estos diseños en el entorno del codebase** (React/Next, Vue, etc.) usando sus patrones y librerías. Si todavía no hay entorno, se recomienda **Next.js (App Router) + TypeScript + Tailwind o CSS Modules + Radix UI** para primitivas accesibles (Dialog, Sheet, Popover, Switch, Slider, Tabs), **TanStack Virtual** para listas virtuales y **Lenis** para el smooth scroll de la landing.

Formato de los archivos: cada `*.dc.html` es un "Design Component": HTML con estilos inline y huecos `{{ valor }}`, más una clase JS (`class Component extends DCLogic`) dentro de `<script data-dc-script>` al final del archivo con el estado y los handlers (equivale a un componente de clase React: `state`, `setState`, `renderVals()` devuelve lo que consume la plantilla). `<sc-for>` = map, `<sc-if>` = render condicional, `<dc-import name="X">` = componente hijo `X.dc.html`. Se abren directo en el navegador (necesitan `support.js` al lado).

## Fidelity
**Alta fidelidad (hi-fi).** Colores, tipografía, espaciados, radios, sombras, copy, estados y animaciones son finales. Recrear pixel-perfect. Los **datos** (números, clientes, pedidos, precios de la competencia, testimonios) son **de ejemplo** y deben venir de la API.

---

## Design Tokens

### Color — marca
| Token | Hex | Uso |
|---|---|---|
| terracota | `#D2582F` | Acento principal, CTAs, toggles activos, barras |
| terracota-oscuro | `#C24E25` | Texto de acento sobre claro, hover |
| terracota-claro | `#E8875F` | Acento sobre fondos oscuros |
| tostado (ink) | `#2A140C` | Texto principal, botones oscuros, fondos de bloques premium |
| crema | `#FBF7F2` | Fondo landing, papel |
| crema-2 | `#F7F1EA` | Fondo panel (light) |
| arena | `#EFE6DA` / `#E7DED3` | Superficies secundarias, lienzo |
| durazno | `#FCEDE5` / `#F7E2D3` / `#FCEADF` | Bloques de Hornero, tips |
| marrón medio | `#6B5244` (muted) · `#7A6252` · `#7E6656` (faint) · `#9C8676` | Textos secundarios |
| Caja negro | `#131211` · texto `#F4EEE6` | Bloques de la integración con Caja |

### Color — semánticos
| | Hex | Fondo suave |
|---|---|---|
| éxito | `#2F8F63` (dark-mode `#7FD3A6`) | `rgba(47,143,99,.13)` |
| alerta | `#E3A93B` / texto `#9A6A10` | `rgba(227,169,59,.18)` |
| error | `#B83A2E` | `rgba(184,58,46,.1)` |
| info | `#3E6BFD` | `rgba(62,107,253,.12)` |
| VIP | `#C98A12` | — |
| WhatsApp suave | bg `#E9F8EE` / fg `#1F6F4A` | — |

### Panel — variables de tema (CSS custom properties)
```css
/* light */
--bg:#F7F1EA; --bga:rgba(247,241,234,.82); --sf:#FFFFFF; --sf2:#F1E9DF; --sf3:#F7F1EA;
--ln:rgba(58,29,18,.09); --ln2:rgba(58,29,18,.16); --tx:#2A140C; --mu:#6B5244; --fa:#7E6656;
--chip:rgba(58,29,18,.055); --sh:0 1px 2px rgba(58,29,18,.08);
--shd:inset 0 1px 0 #fff, 0 1px 2px rgba(58,29,18,.05), 0 10px 24px -14px rgba(58,29,18,.22);
/* dark */
--bg:#22140D; --bga:rgba(34,20,13,.82); --sf:#2E1B12; --sf2:#28170F; --sf3:#3A2418;
--ln:rgba(251,247,242,.08); --ln2:rgba(251,247,242,.15); --tx:#FBF7F2; --mu:#CDB8A8; --fa:#B09A8A;
--chip:rgba(251,247,242,.07); --sh:0 1px 2px rgba(0,0,0,.3);
--shd:inset 0 1px 0 rgba(255,255,255,.04), 0 14px 30px -16px rgba(0,0,0,.6);
```
`sf` = superficie (cards), `sf3` = superficie hundida (inputs, filas internas), `ln` = líneas, `mu` = texto secundario, `fa` = texto terciario, `chip` = fondo de chips/segmented, `shd` = sombra de card.

### Tipografía
- **Instrument Sans** (Google Fonts, 400/500/600/700) en todo el producto. **No usar monoespaciada.**
- Las plantillas de tienda usan sus propias fuentes (ver sección Plantillas).
- Escala del panel: H1 30px/600, tracking −0.035em · título de card 15–19px/600 · cuerpo 13.5–15px/500 · meta 12–12.5px · KPI 22–30px/600 tracking −0.03em.
- Escala de la landing: hero `clamp(44px, 7vw, 96px)` / 600 / line-height .92–1 / tracking −0.045 a −0.055em · H2 `clamp(34px, 4.6vw, 58px)` · cuerpo 17–18px line-height 1.5 · `text-wrap: balance` en títulos y `pretty` en párrafos.
- Labels en mayúsculas: 11–13px / 700 / tracking .08–.14em / color terracota.

### Radios
Chips/botones `999px` · inputs 14px · cards chicas 16–18px · cards 20–24px · bloques grandes 26–36px · **arco** (motivo de marca): `border-radius: R R 0 0` o `R R 6px 6px` con R = mitad del ancho, para fotos y contenedores.

### Sombras
- Card: `--shd` (ver arriba).
- Flotante (toast, barra masiva, dropdown): `0 24px 50px -16px rgba(42,20,12,.7)`.
- CTA terracota: `0 10px 24px -10px rgba(210,88,47,.7)`.
- Mockups de landing: `0 40px 80px -40px rgba(42,20,12,.4), 0 0 0 1px rgba(42,20,12,.06)`.

### Espaciado
Base 4px. Gaps más usados: 6, 8, 10, 12, 14, 16, 18, 24. Padding de card 14–20px. Secciones de landing: `padding: 120px 28px`, `max-width: 1240px`.

### Texturas
- **Grano**: `assets/grain.png` como overlay `opacity .15–.25`, `mix-blend-mode: multiply` sobre terracota.
- **Patrón de arcos**: `assets/arch-pattern-{light,dark,accent}.svg` repetido (`background-size` 22–160px), a veces con máscara degradé.
- **Líneas blueprint** finas en algunos bloques de landing.

---

## Motion
- Easing estándar: `cubic-bezier(.2,.7,.1,1)`.
- Entrada de bloques (`data-r` en el panel, `data-anim="reveal"` en la landing): `opacity 0→1`, `translateY(14px)→0`, 560–600ms, stagger 45–50ms.
- Toast: entra `translateX(30px) scale(.97)` → none, 440ms; barra de vida `scaleX(1→0)` lineal 5,5s.
- Mensajes de chat: `translateY(12px) scale(.97)`, 420ms.
- Sheets/drawers: `transform` .4s mismo easing. Modales: backdrop `rgba(32,15,8,.55)` + `backdrop-filter: blur(6px)`.
- Toggles: knob `left` .25s, fondo .25s.
- Landing: **Lenis** (`duration 1.15`, easing `t => Math.min(1, 1.001 - 2^(-10t))`), anclas `#id` con `offset -80`. Lenis se detiene cuando hay menú/modal abierto (`data-lenis-prevent` en contenedores con scroll propio). Parallax suave en elementos `data-px` según scroll.
- Contadores en vivo (ofertas relámpago, cronómetros) actualizan cada 1s.
- Respetar `prefers-reduced-motion` (desactivar parallax, reveals y Lenis).

---

## Screens / Views

### 1. Landing (`Mostrador v4.dc.html`, vista "Landing")
Mensaje central: **"Todo lo que tienen las otras. Por menos plata y más seguro."**
- **Header fijo**: wordmark + nav (Producto, Plantillas, Hornero, Precios, Caja, Integraciones) + "Ingresar" + CTA terracota "Crear mi tienda". Fondo crema translúcido con blur al scrollear. En tablet/mobile: botón hamburguesa que abre un **panel desplegable** desde debajo del header (no full-screen), con backdrop, grilla de accesos en tarjetas y CTAs; el ícono rota al abrir; **no debe mover el layout** (scrollbar-gutter estable y Lenis detenido).
- **Hero**: titular grande, subtítulo, 2 CTAs, prueba social; en desktop composición con mockup del panel + foto en arco + notificación de venta en vivo animada (feed que agrega ventas). Mobile/tablet: versión apilada.
- **Mockups del panel real** (1–2) mostrando el dashboard tal como se ve.
- **Hornero (IA)**: sección con gradiente durazno, chat demo animado, controles "Vos tenés el control" (aprobar antes de enviar, horarios, derivar a humano), proveedores de IA (OpenAI, Claude, Gemini… vía OAuth).
- **Comparativa vs. Tiendanube/Empretienda**: calculadora con slider de ventas y barra de costo; tabla de features (✓ / "No figura en el plan").
- **Plantillas**: énfasis reducido; carrusel/galería; abre modal con preview desktop/mobile y "Ver la tienda demo completa".
- **Caja**: bloque negro `#131211` — "Mostrador vende. Caja ordena la plata." flujo de movimientos → tarjeta de Caja (ganancia, MEP, monotributo). Caja es conector de IAs (MCP/conectores) — se maneja desde la IA del usuario. Promo: **3 meses de Caja Pro Max con cualquier plan**.
- **Integraciones**: muro de logos (WhatsApp, Telegram, Instagram, Mercado Libre, Mercado Pago, MODO, Payway, Ualá, Andreani, OCA, Correo Argentino, ARCA, Google, Meta, TikTok, etc.).
- **Precios**, **FAQ**, **CTA final**, **footer**.
- **Widget de Hornero** flotante abajo a la derecha: aparece a los 2,6s; al cerrarlo **no desaparece**, queda el avatar con globito.
- **Banner de cookies** (`CookieBanner.dc.html`), persiste en `localStorage['mostrador-cookies']`.

### 2. Onboarding (`Onboarding Mostrador.dc.html`)
Wizard de varios pasos con barra de progreso ("Paso N de M"), Hornero acompañando con tips, botones Atrás/Continuar fijos. Objetivo: **precargar todo**. Pasos: datos del negocio/rubro → importar desde (Tiendanube, Empretienda, Mercado Shops, Instagram, Excel) → marca (logo, colores, plantilla) → productos detectados → pagos → envíos → canales (WhatsApp, IG, ML) → facturación ARCA → **banner promocional de Caja** y pantalla final con beneficios de Caja → tienda lista.

### 3. Panel (`Mostrador v4.dc.html`, vista "Panel")
**Layout**: sidebar 260–280px (desktop fija; tablet/mobile como drawer con backdrop) + topbar sticky (`--bga` + blur) + contenido `max-width ~1280px`, padding 24px (16px mobile). Mobile: barra inferior con accesos.

**Sidebar**: selector de tienda arriba (avatar + nombre + plan), buscador/⌘K, áreas agrupadas con ícono (Phosphor) y badges numéricos; modo **Simple** (lista corta de 10) vs **Completo** (áreas):
- Inicio: Resumen · Para hacer hoy (4) · Actividad
- Ventas: Pedidos (7) · Devoluciones · Ventas en local · Carritos abandonados (12) · WhatsApp (3) · Clientes · Reseñas y preguntas
- Catálogo: Productos · Inventario · Listas de precio
- Crecer: Marketing · **Cupones** · Canales · Informes
- Operación: Finanzas · Facturación · Envíos
- Tienda: Mi tienda · Configuración · Integraciones
- Cuenta: Plan y facturación · Usuarios y permisos · Dominio · Notificaciones · Mis datos
Abajo: tarjeta de Hornero, toggle claro/oscuro.

**Mapa pantalla → archivo**
| Pantallas | Archivo |
|---|---|
| Resumen / Inicio (dashboard) | `Mostrador v4.dc.html` |
| Pedidos | `PanelPedidos.dc.html` |
| Clientes, Productos, Inventario, Facturación, Envíos | `PanelListas.dc.html` |
| WhatsApp, Carritos abandonados, Devoluciones, Ventas en local, Reseñas y preguntas | `PanelVentas.dc.html` |
| Listas de precio | `PanelCatalogo.dc.html` |
| Cupones (cupones, automáticas, relámpago) | `PanelCupones.dc.html` |
| Marketing, Canales, Informes, Actividad, Para hacer hoy | `PanelCrecer.dc.html` |
| Finanzas (con tarjeta "Conectado con Caja"), Cobros, Mi tienda, Configuración, Integraciones, Ayuda | `PanelOperacion.dc.html` |
| Plan y facturación, Usuarios y permisos, Dominio, Notificaciones, Mis datos | `PanelCuenta.dc.html` |

**Principio clave de las listas: nada de "tablas muertas".** Cada fila es una **tarjeta operable**: avatar/miniatura, canal como badge sobre el avatar, estado visual (barra de etapas), urgencia ("Despachá hoy", "Atrasado"), y **la próxima acción en un botón** (reclamar pago por WhatsApp, preparar, imprimir etiqueta, seguimiento). Todas las listas largas usan **scroll virtual** y **todos los filtros posibles**.

**Pedidos (referencia del patrón)**:
- Header con conteo, suma y ticket promedio filtrados + Exportar (respeta filtros).
- Vistas guardadas: Todos · Para despachar hoy · Sin pagar · VIP · Mayoristas · Vendidos por Hornero.
- Embudo de etapas clickeable (Sin pagar → Para preparar → Listo para despachar → En camino → Entregado) con contadores que funcionan como filtro.
- Barra de filtros: búsqueda (cliente, N.º, producto, ciudad), segmented de fecha (Hoy/7d/30d/Todo), panel "Filtros" con facetas Canal/Pago/Envío **con conteos que se recalculan excluyendo su propia faceta**, slider de monto mínimo, toggles VIP/Hornero/Mayoristas; chips de filtros activos removibles + "Limpiar todo"; orden cíclico (Más nuevos, Urgentes primero, Mayor monto, Más viejos).
- Lista virtual: fila de 76px, ventana visible + overscan 4/6, contenedor de 600px (520 mobile), botón "volver arriba" tras 1200px.
- Selección múltiple → **barra flotante** oscura con N seleccionados + total + acciones (Preparar, Etiquetas, Facturar, Avisar).
- Click en fila abre sheet de detalle. Acciones cambian estado con toast + "Deshacer".

**Cupones** (`PanelCupones.dc.html`):
- KPIs: activos, ventas con cupón, descontado, rendimiento (ventas / descuento).
- Tip de Hornero con acción aplicable.
- Tabs: Cupones · Automáticas · Relámpago.
- Cupones como **tickets** (cabecera oscura con código, línea punteada, valor grande), estado, vencimiento, canales, barra de usos/límite, Vendió/Descontó/Rinde (verde ≥ ×6, rojo < ×4), sugerencia de Hornero, acciones (copiar, mandar por WhatsApp, historia IG, duplicar, toggle activo).
- **Armador**: objetivo (primera compra, que vuelvan, liquidar stock, influencer) que precarga valores; código con "Sugerir"; tipo (%/monto/envío gratis/2x1); valor; aplica a (tienda/categoría/productos); quién (todos/primera compra/VIP/en riesgo); compra mínima; límite total; usos por cliente; vigencia; canales; opciones (Hornero lo ofrece, combinable, códigos únicos en lote, link auto-aplicable `mostrador.com.ar/c/<codigo>`). Columna derecha sticky: preview del ticket + estimación de Hornero (ventas extra, costo, **margen** con semáforo y alerta < 15 %). Crear / Programar.
- Automáticas: transferencia, envío gratis desde, por cantidad, 3x2, cuotas sin interés, mayorista, rescate de carrito con Hornero, cumpleaños — cada una con toggle, ejemplo en carrito y resultado.
- Relámpago: oferta en vivo con cuenta regresiva (h/min/seg), stock vendido/restante, acciones (+1 h, liberar stock, terminar) y próximas programadas.

**Dashboard**: saludo con Hornero, KPIs, gráfico de ventas por período, meta del mes, "Para hacer hoy", feed en vivo, accesos rápidos; grilla que rellena huecos (`grid-auto-flow: dense`, la última card ocupa el ancho en tablet).

### 4. Feedback: toasts, alertas, modales, sheets
- **Toast**: card oscura/clara con ícono por tipo (ok, warn, info, wa, andreani…), título + detalle, acción opcional ("Deshacer", "Compartir"), barra de vida 5,5s; apilados abajo a la derecha (arriba de la barra mobile).
- **Modal**: centrado, radio 24–28px, backdrop blur; acciones a la derecha.
- **Sheet**: lateral derecho en desktop, bottom-sheet en mobile, con handle.
- **Alert inline**: bloque con fondo semántico suave + ícono + acción.
- Vista **Componentes** en `Mostrador v4.dc.html` muestra todo el catálogo.

### 5. Páginas públicas
`Legal *.dc.html` (Términos, Privacidad, Cookies, Arrepentimiento con formulario, Defensa del consumidor, Tratamiento de datos, Uso aceptable, Reembolsos) y `Mostrador *.dc.html` (Seguridad, Estado, Contacto, Ayuda, Migrar `?desde=tiendanube|empretienda|mercadoshops`, Nosotros, Partners, Novedades, Ingresar `?m=registro`, 404). Mismo header/footer/banner de cookies. Dominio: **https://mostrador.com.ar** — cada página tiene `og:url`, `canonical`, `og:image` absolutos (`/assets/og/*.png`). Rutas: `/`, `/ingresar`, `/contacto`, `/ayuda`, `/estado`, `/seguridad`, `/migrar`, `/nosotros`, `/novedades`, `/partners`, `/legales/{terminos,privacidad,cookies,arrepentimiento,defensa-del-consumidor,tratamiento-de-datos,uso-aceptable,reembolsos}`. Botón de arrepentimiento obligatorio (normativa AR) visible.

### 6. Plantillas de tienda
Storefronts por rubro, cada uno con Inicio, Catálogo con filtros, Ficha, Carrito, Checkout, Nosotros/contacto (+ páginas extra por rubro). Secciones de conversión: comprar por WhatsApp, cuotas + precio por transferencia, reseñas con fotos, envío por CP, Instagram, más vendidos/combos, promo con cuenta regresiva. Responsive con `container-type` + `cqw`.
| Archivo | Rubro | Fuente / acento |
|---|---|---|
| Plantilla Ceibo | Moda editorial | Instrument Serif |
| Plantilla Mostaza | Gastronomía (Café Bruma), menú y pedido | — |
| Plantilla Yerbal | Almacén, catálogo denso | — |
| Plantilla Tijera | Servicios (barbería), turnos | Syne · `#5B3FD6` |
| Plantilla Pampa | Streetwear brutalista | Archivo Black · `#E6FF4D` |
| Plantilla Malbec | Vinoteca +18, club de vinos | Cormorant Garamond · bordó/dorado |
| Plantilla Miga / Miga Clasica | Panadería, cronograma del horno | Schibsted Grotesk / Young Serif · `#B5651D` |
| Plantilla Rosedal | Belleza, rutina por tipo de piel | Gloock + Hanken Grotesk · `#B4475B` |
| Plantilla Andes | Tecno, comparar, factura A | Sora · `#2F6BFF` |
| Plantilla Quebracho | Hogar, "shop the room", a medida | Newsreader + Familjen Grotesk · `#4F5B3A` |
| Plantillas Tienda | 12 temas data-driven (`?t=<id>`) | ver `templates-data.js` |

### 7. Marca
- `Papeleria Mostrador.dc.html`: tarjetas 85×55 mm (1a/1b), membrete A4 y Carta, sobre DL, firmas de email (600px), factura B y remito R (A4, con QR ARCA real), carpeta 230×310 mm, credenciales 54×86 mm, plancha de stickers.
- `Redes Mostrador.dc.html`: avatares 400², destacadas IG, portadas LinkedIn 1584×396, X 1500×500, YouTube 2560×1440 (zona segura 1546×423), 11 ads Meta (1080×1350 / 1080×1920 / 1080²), 3 miniaturas YT 1280×720.
- `Imprimibles Mostrador.dc.html`: flyers A5, afiche A3, tríptico A4, one-pager A4, roll-up 85×200 cm, backing 300×225 cm, tarjeta para paquetes A6.
- `OG Mostrador.dc.html`: 9 imágenes 1200×630 → `assets/og/`.
- Medidas reales en mm; imprenta con 3 mm de sangrado. Pantones de referencia (validar con prueba de color): terracota ≈ 7592 C, tostado ≈ 4625 C.

### 8. Hornero (mascota IA)
Hornero (ave argentina). Sprites en `assets/hornero/`: saludo, hablando, pensando, celebrando, exito, preocupado, durmiendo, desenchufado, lupa, foto, grafico, caja, envio, ticket, portapapeles, seguridad, whatsapp, horno, horno-vacio, asomado, avatar. Uso: estados vacíos (lupa), éxito (celebrando/exito), error/offline (preocupado/desenchufado), carga (pensando), onboarding (saludo), widgets y avatar del chat. Símbolo `✦` = "hecho por Hornero".

---

## Interactions & Behavior (resumen)
- Navegación del panel por `sub` (pantalla activa) + área; en mobile el drawer se cierra al elegir.
- Tema claro/oscuro: toggle en sidebar; aplica variables de tema en el root del panel.
- Modo Simple/Completo cambia el sidebar.
- Toda acción destructiva o de cambio de estado → toast con "Deshacer".
- Filtros: estado en URL (querystring) recomendado para compartir vistas.
- Responsive: breakpoints usados ~760px (mobile), ~1000px (tablet: colapsan columnas laterales), 1280px+ desktop. Todo debe reflujar (`minmax(0,1fr)`, `auto-fit/auto-fill`), sin textos partidos en chips/botones (`white-space: nowrap` + ellipsis en nombres).
- Hit targets mobile ≥ 44px.

## State Management (mínimo)
- Global: `theme`, `mode (simple|completo)`, `sub`, `sidebarOpen`, `toasts[]`, `sheet/modal` activo.
- Pedidos: `q, stages{}, channels{}, pay{}, ship{}, date, minAmount, vip, ai, mayor, sort, view, selection{}, scrollTop`; datos paginados/infinite desde API, virtualizados.
- Cupones: `tab, filter, q, sort, coupons[]`, borrador del armador (`goal, code, type, value, scope, products, who, minPurchase, limit, perCustomer, duration, channels, options`), estimación calculada en backend (Hornero).
- Onboarding: paso actual + datos importados (persistir en servidor por paso).
- Landing: `menuOpen`, `widgetOpen`, `tplModal`, calculadora de precios.

## Integraciones (referenciadas en UI)
Pagos: Mercado Pago, MODO, Payway, Ualá, tarjetas, transferencia. Canales: WhatsApp Cloud API, Instagram, Telegram, Mercado Libre, TikTok, Google. Envíos: Andreani, OCA, Correo Argentino, motos. Fiscal: ARCA (factura A/B/C, CAE, QR). IA: OAuth con OpenAI, Anthropic, Google (Gemini). Finanzas: **Caja** (https://caja.iamnaw.dev) — sincroniza ventas, liquidaciones, comisiones, envíos y facturas; conector de IA vía MCP.

## Assets
- `assets/mark.svg` — isotipo (arco), `assets/wordmark-{brown,cream}.png` — logotipo. Logo original subido: `brand/logo-original.png`.
- `assets/grain.png`, `assets/arch-pattern-*.svg` — texturas.
- `assets/hornero/*.png` — sprites de la mascota.
- `assets/og/*.png` — imágenes OG 1200×630.
- `assets/qr/*.svg` — QR reales (vCard, ARCA, /empezar con utm, /live, /contacto, referidos).
- Íconos: **Phosphor** (`ph:*`) y logos de marcas vía Iconify (`logos:*`, `simple-icons:*`).
- Fotos: Unsplash (URLs en el código) — reemplazar por fotos propias/licenciadas.
- Mascota de Caja: `https://caja.iamnaw.dev/mascota/*.png` (pedir PNG al equipo de Caja para uso offline).

## Files
Todos en la raíz de este paquete:
- `Mostrador v4.dc.html` — landing + panel + componentes (archivo principal; píldora inferior cambia de vista).
- `Onboarding Mostrador.dc.html`
- `Panel*.dc.html` — pantallas del panel (ver mapa).
- `CookieBanner.dc.html`, `Legal *.dc.html`, `Mostrador *.dc.html` — sitio público.
- `Plantilla *.dc.html`, `Plantillas Tienda.dc.html`, `templates-data.js` — storefronts.
- `Papeleria / Redes / Imprimibles / OG Mostrador.dc.html` — marca.
- `support.js` (runtime para abrir los .dc.html en el navegador), `image-slot.js`.
- `assets/`, `brand/`.

## Pendientes / a validar
- Cifras de marketing de ejemplo (12.000 tiendas, −38 %, ahorro $ 181.188/año, 99,99 % uptime, precios de la competencia oct-2026).
- Crear las rutas `/empezar`, `/live`, `/live/acceso/:id`, `/c/:codigo` que usan los QR y links.
- Cupones/códigos de ejemplo: ARRANCO50, LUCIA15, BRUMA-REF, FINDE10, etc.
