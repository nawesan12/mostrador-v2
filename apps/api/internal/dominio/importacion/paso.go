package importacion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Paso es el manejador de `importacion.paso`: lee una página de la fuente (fuera de la transacción,
// porque es una llamada de red), y en una transacción la guarda, avanza el cursor y encola el paso
// siguiente. Es idempotente: si el paso ya se hizo (`pasos` avanzó) o la importación terminó o se
// canceló, no hace nada. Un error de la fuente se reintenta con la espera del outbox; en el último
// intento, o si es definitivo, la importación queda fallida con ese error.
func (s *Servicio) Paso(ctx context.Context, ev outbox.Evento) error {
	var e evento
	if json.Unmarshal(ev.Datos, &e) != nil || e.ID == "" {
		return nil
	}
	tienda := ev.TiendaID
	var plataforma, urlOrigen, etapa, estado string
	var cursor json.RawMessage
	var cred []byte
	var pasos int
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT plataforma, url_origen, etapa, estado, cursor, credencial, pasos FROM importaciones
			WHERE tienda_id = $1 AND id::text = $2`, tienda, e.ID).Scan(&plataforma, &urlOrigen, &etapa, &estado, &cursor, &cred, &pasos)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if pasos != e.Paso || (estado != "pendiente" && estado != "corriendo") {
		return nil
	}
	if pasos >= MaxPasos {
		return s.fallar(ctx, tienda, e, etapa, "La tienda de origen no terminaba de mandar datos: frenamos la importación.")
	}
	abrir := s.Fuentes[plataforma]
	if abrir == nil {
		return s.fallar(ctx, tienda, e, etapa, "Todavía no podemos importar desde esa plataforma.")
	}
	o := Origen{Plataforma: plataforma, URL: urlOrigen}
	if len(cred) > 0 {
		if s.Cifra == nil {
			return s.fallar(ctx, tienda, e, etapa, "No pudimos leer la credencial guardada. Empezá la importación de nuevo.")
		}
		if o.Credencial, err = s.Cifra.Descifrar(contextoCifrado(tienda), cred); err != nil {
			return s.fallar(ctx, tienda, e, etapa, "No pudimos leer la credencial guardada. Empezá la importación de nuevo.")
		}
	}
	pag, err := leerPagina(ctx, abrir, o, etapa, cursor)
	if err != nil {
		if errors.Is(err, ErrDefinitivo) || ev.Intentos+1 >= outbox.MaxIntentos {
			return s.fallar(ctx, tienda, e, etapa, detalleDe(err))
		}
		return err
	}
	if len(pag.Siguiente) > 2048 {
		return s.fallar(ctx, tienda, e, etapa, "La tienda de origen mandó un cursor demasiado largo.")
	}
	return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var prog Progreso
		var errs []Error
		if err := t.QueryRow(ctx, `SELECT pasos, estado, progreso, errores FROM importaciones
			WHERE tienda_id = $1 AND id::text = $2 FOR UPDATE`, tienda, e.ID).Scan(&pasos, &estado, &prog, &errs); err != nil {
			return err
		}
		// Otro worker lo hizo o la cancelaron mientras se leía la página.
		if pasos != e.Paso || (estado != "pendiente" && estado != "corriendo") {
			return nil
		}
		p, nuevos, err := GuardarPagina(ctx, t, tienda, plataforma, etapa, pag)
		if err != nil {
			return err
		}
		prog.sumar(p)
		errs = append(errs, nuevos...)
		if len(errs) > MaxErrores {
			errs = errs[len(errs)-MaxErrores:]
		}
		sig, cur := etapa, json.RawMessage(pag.Siguiente)
		if len(cur) == 0 || string(cur) == "null" {
			cur = nil
			i := slices.Index(Etapas, etapa)
			sig = ""
			if i+1 < len(Etapas) {
				sig = Etapas[i+1]
			}
		}
		if sig == "" {
			// Terminó: la credencial se borra.
			_, err = t.Exec(ctx, `UPDATE importaciones SET estado = 'lista', cursor = NULL, pasos = pasos + 1, progreso = $3,
				errores = $4, credencial = NULL, terminada = now() WHERE tienda_id = $1 AND id::text = $2`, tienda, e.ID, prog, errs)
			return err
		}
		if _, err := t.Exec(ctx, `UPDATE importaciones SET estado = 'corriendo', etapa = $3, cursor = $4, pasos = pasos + 1,
			progreso = $5, errores = $6 WHERE tienda_id = $1 AND id::text = $2`, tienda, e.ID, sig, crudo(cur), prog, errs); err != nil {
			return err
		}
		return outbox.Publicar(ctx, t, tienda, TipoPaso, evento{ID: e.ID, Paso: e.Paso + 1})
	})
}

// leerPagina abre la fuente y pide la página; un panic de una fuente es un error, no un worker caído.
func leerPagina(ctx context.Context, abrir Abrir, o Origen, etapa string, cursor json.RawMessage) (p Pagina, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: la fuente falló (%v)", ErrDefinitivo, r)
		}
	}()
	f, err := abrir(ctx, o)
	if err != nil {
		return Pagina{}, err
	}
	return f.Pagina(ctx, etapa, cursor)
}

func detalleDe(err error) string {
	d := strings.TrimPrefix(err.Error(), ErrDefinitivo.Error()+": ")
	if d == ErrDefinitivo.Error() || d == "" {
		d = "La tienda de origen no respondió."
	}
	r := []rune(d)
	if len(r) > 300 {
		d = string(r[:300])
	}
	return d
}

// fallar deja la importación fallida con el error al final de la lista (si el paso sigue siendo
// el de ahora: un evento viejo no pisa nada).
func (s *Servicio) fallar(ctx context.Context, tienda string, e evento, etapa, detalle string) error {
	slog.Warn("importacion: fallida", "tienda", tienda, "importacion", e.ID, "detalle", detalle)
	return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var errs []Error
		err := t.QueryRow(ctx, `SELECT errores FROM importaciones WHERE tienda_id = $1 AND id::text = $2 AND pasos = $3
			AND estado IN ('pendiente', 'corriendo') FOR UPDATE`, tienda, e.ID, e.Paso).Scan(&errs)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		errs = append(errs, Error{Etapa: etapa, Detalle: detalle})
		if len(errs) > MaxErrores {
			errs = errs[len(errs)-MaxErrores:]
		}
		_, err = t.Exec(ctx, `UPDATE importaciones SET estado = 'fallida', errores = $3, credencial = NULL, terminada = now()
			WHERE tienda_id = $1 AND id::text = $2`, tienda, e.ID, errs)
		return err
	})
}

// cortar recorta un texto a n letras y avisa si cortó.
func cortar(s string, n int) (string, bool) {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s, false
	}
	return strings.TrimSpace(string(r[:n])), true
}

func cortarPtr(p *string, n int) bool {
	if p == nil {
		return false
	}
	var c bool
	*p, c = cortar(*p, n)
	return c
}

// GuardarPagina guarda lo de una etapa en la transacción (con la tienda fijada). Cada ítem va en
// su propio savepoint: uno que no pasa las validaciones queda en los errores y el resto entra.
func GuardarPagina(ctx context.Context, t pgx.Tx, tienda, plataforma, etapa string, p Pagina) (Progreso, []Error, error) {
	var prog Progreso
	errs := []Error{}
	origen := "importado:" + plataforma
	item := func(externo string, fn func(pgx.Tx) error) error {
		sp, err := t.Begin(ctx)
		if err != nil {
			return err
		}
		if err := fn(sp); err != nil {
			_ = sp.Rollback(ctx)
			// Lo que corta la transacción entera (la conexión, el contexto) no es culpa del ítem.
			if ctx.Err() != nil || t.Conn().IsClosed() {
				return err
			}
			var d string
			if v, ok := errors.AsType[validacion](err); ok {
				d = string(v)
			} else {
				d = "No se pudo guardar: " + err.Error()
			}
			errs = append(errs, Error{Etapa: etapa, Externo: externo, Detalle: d})
			return nil
		}
		return sp.Commit(ctx)
	}
	var err error
	switch etapa {
	case "productos":
		for _, x := range p.Productos {
			if err = item(x.Externo, func(t pgx.Tx) error { return guardarProducto(ctx, t, tienda, origen, x, &prog) }); err != nil {
				break
			}
		}
	case "clientes":
		for _, x := range p.Clientes {
			if err = item(x.Externo, func(t pgx.Tx) error { return guardarCliente(ctx, t, tienda, origen, x, &prog) }); err != nil {
				break
			}
		}
	case "pedidos":
		for _, x := range p.Pedidos {
			if err = item(x.Externo, func(t pgx.Tx) error { return guardarPedido(ctx, t, tienda, plataforma, x, &prog) }); err != nil {
				break
			}
		}
	case "redirecciones":
		for _, x := range p.Redirecciones {
			if err = item(x.Desde, func(t pgx.Tx) error {
				desde := vidriera.NormalizarRuta(x.Desde)
				if desde == "" || !vidriera.HaciaValido(x.Hacia) {
					return validacion("La dirección vieja o la nueva no es un path válido.")
				}
				n, err := redirigir(ctx, t, tienda, origen, desde, x.Hacia)
				prog.Redirecciones += n
				return err
			}); err != nil {
				break
			}
		}
	}
	return prog, errs, err
}

// validacion es un error de un ítem para mostrar tal cual.
type validacion string

func (v validacion) Error() string { return string(v) }

func redirigir(ctx context.Context, t pgx.Tx, tienda, origen, desde, hacia string) (int, error) {
	tag, err := t.Exec(ctx, `INSERT INTO redirecciones (tienda_id, desde, hacia, origen) VALUES ($1, $2, $3, $4)
		ON CONFLICT (tienda_id, desde) DO UPDATE SET hacia = EXCLUDED.hacia, origen = EXCLUDED.origen`, tienda, desde, hacia, origen)
	return int(tag.RowsAffected()), err
}

// guardarProducto: con listas.GuardarEnTx (los mismos topes y validaciones que el panel). Si la URL
// vieja ya está redirigida o el SKU ya existe en la tienda, es un repetido y no se vuelve a crear.
func guardarProducto(ctx context.Context, t pgx.Tx, tienda, origen string, x ProductoViejo, prog *Progreso) error {
	e := x.Producto
	desde := vidriera.NormalizarRuta(x.URL)
	var ya bool
	if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM redirecciones WHERE tienda_id = $1 AND desde = $2 AND $2 <> '')
		OR EXISTS (SELECT 1 FROM productos WHERE tienda_id = $1 AND sku = $3 AND $3 <> '')`,
		tienda, desde, strings.TrimSpace(valor(e.SKU))).Scan(&ya); err != nil {
		return err
	}
	if ya {
		prog.Repetidos++
		return nil
	}
	var corto bool
	e.Nombre, corto = cortar(e.Nombre, 120)
	c := []bool{corto, cortarPtr(e.Descripcion, 4000), cortarPtr(e.Categoria, 60), cortarPtr(e.SKU, 40)}
	if len(e.Fotos) > 8 {
		e.Fotos, c = e.Fotos[:8], append(c, true)
	}
	id, err := listas.GuardarEnTx(ctx, t, tienda, "", e)
	if err != nil {
		return validacion(primeraMayuscula(err.Error()) + ".")
	}
	prog.Productos++
	if slices.Contains(c, true) {
		prog.Recortados++
	}
	if desde != "" {
		n, err := redirigir(ctx, t, tienda, origen, desde, "/producto/"+id)
		prog.Redirecciones += n
		return err
	}
	return nil
}

func valor(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func primeraMayuscula(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	return strings.ToUpper(string(r[0])) + string(r[1:])
}

// email en minúsculas, o "" si no parece un email (el cliente entra igual, sin email).
func emailDe(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", false
	}
	at := strings.IndexByte(s, '@')
	if len(s) < 3 || len(s) > 120 || at < 1 || at == len(s)-1 || strings.ContainsAny(s, " \t\r\n") {
		return "", true
	}
	return s, false
}

// guardarCliente: con email, un upsert que sólo completa lo que falta (nunca pisa nombre ni email);
// sin email, se busca por el id de la plataforma vieja para no repetirlo.
func guardarCliente(ctx context.Context, t pgx.Tx, tienda, origen string, x ClienteViejo, prog *Progreso) error {
	email, malo := emailDe(x.Email)
	nombre, c1 := cortar(x.Nombre, 120)
	if nombre == "" && email != "" {
		nombre = email[:strings.IndexByte(email, '@')]
	}
	if nombre == "" {
		return validacion("Un cliente sin nombre ni email no se puede guardar.")
	}
	tel, c2 := cortar(x.Telefono, 30)
	ciudad, c3 := cortar(x.Ciudad, 80)
	prov, c4 := cortar(x.Provincia, 60)
	notas, c5 := cortar(x.Notas, 2000)
	externo, c6 := cortar(x.Externo, 80)
	nulo := func(s string) any {
		if s == "" {
			return nil
		}
		return s
	}
	var nuevo bool
	var err error
	if email != "" {
		err = t.QueryRow(ctx, `INSERT INTO clientes (tienda_id, nombre, email, telefono, ciudad, provincia, notas, origen, externo)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (tienda_id, lower(email)) WHERE email IS NOT NULL DO UPDATE SET
				telefono = coalesce(clientes.telefono, EXCLUDED.telefono),
				ciudad = CASE WHEN clientes.ciudad = '' THEN EXCLUDED.ciudad ELSE clientes.ciudad END,
				provincia = CASE WHEN clientes.provincia = '' THEN EXCLUDED.provincia ELSE clientes.provincia END,
				notas = CASE WHEN clientes.notas = '' THEN EXCLUDED.notas ELSE clientes.notas END
			RETURNING xmax = 0`, tienda, nombre, email, nulo(tel), ciudad, prov, notas, origen, nulo(externo)).Scan(&nuevo)
	} else {
		var tag pgconn.CommandTag
		tag, err = t.Exec(ctx, `INSERT INTO clientes (tienda_id, nombre, telefono, ciudad, provincia, notas, origen, externo)
			SELECT $1, $2, $3, $4, $5, $6, $7, $8
			WHERE $8::text IS NULL OR NOT EXISTS (SELECT 1 FROM clientes WHERE tienda_id = $1 AND origen = $7 AND externo = $8)`,
			tienda, nombre, nulo(tel), ciudad, prov, notas, origen, nulo(externo))
		nuevo = err == nil && tag.RowsAffected() == 1
	}
	if err != nil {
		return err
	}
	if nuevo {
		prog.Clientes++
	} else {
		prog.Repetidos++
	}
	if malo || c1 || c2 || c3 || c4 || c5 || c6 {
		prog.Recortados++
	}
	return nil
}

// guardarPedido: al historial, sin duplicar por (plataforma, externo). No toca pedidos, stock,
// facturas ni numeración.
func guardarPedido(ctx context.Context, t pgx.Tx, tienda, plataforma string, x PedidoViejo, prog *Progreso) error {
	externo, c := cortar(x.Externo, 80)
	if externo == "" || x.Fecha.IsZero() || x.Total < 0 {
		return validacion("Un pedido sin número, sin fecha o con total negativo no se puede guardar.")
	}
	recorte := c
	numero, c := cortar(x.Numero, 40)
	recorte = recorte || c
	nombre, c := cortar(x.Nombre, 120)
	recorte = recorte || c
	email, malo := emailDe(x.Email)
	estado, c := cortar(x.Estado, 40)
	recorte = recorte || c || malo
	items := x.Items
	if len(items) > 100 {
		items, recorte = items[:100], true
	}
	limpios := make([]ItemViejo, 0, len(items))
	for _, it := range items {
		var c1, c2 bool
		it.Nombre, c1 = cortar(it.Nombre, 120)
		it.Variante, c2 = cortar(it.Variante, 120)
		recorte = recorte || c1 || c2
		if it.Cantidad < 0 || it.Precio < 0 {
			return validacion("Un pedido con cantidades o precios negativos no se puede guardar.")
		}
		limpios = append(limpios, it)
	}
	var correo any
	if email != "" {
		correo = email
	}
	tag, err := t.Exec(ctx, `INSERT INTO historial_pedidos (tienda_id, plataforma, externo, numero_externo, fecha, cliente_email,
			cliente_nombre, total, items, estado)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (tienda_id, plataforma, externo) DO NOTHING`,
		tienda, plataforma, externo, numero, x.Fecha, correo, nombre, x.Total, limpios, estado)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		prog.Repetidos++
		return nil
	}
	prog.Pedidos++
	if recorte {
		prog.Recortados++
	}
	return nil
}
