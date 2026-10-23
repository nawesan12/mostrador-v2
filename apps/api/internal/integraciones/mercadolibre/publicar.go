package mercadolibre

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
)

type producto struct {
	nombre, descripcion string
	foto                *string
	fotos               []string
	precio              int64
	stock               int
	publicado           bool
	opciones            []listas.Opcion
	// Todas las variantes (activas o no), en el orden de las combinaciones.
	variantes []listas.Variante
	// La variación de Mercado Libre de cada variante publicada.
	variaciones map[string]int64
}

func (s *Servicio) leerProducto(ctx context.Context, tiendaID, productoID string) (producto, int, bool, error) {
	var p producto
	var recargo int
	var pausar bool
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT nombre, descripcion, foto, fotos, precio, stock, publicado, opciones,
				coalesce((SELECT variaciones FROM ml_publicaciones WHERE tienda_id = $1 AND producto_id = productos.id), '{}')
			FROM productos WHERE tienda_id = $1 AND id::text = $2`, tiendaID, productoID).
			Scan(&p.nombre, &p.descripcion, &p.foto, &p.fotos, &p.precio, &p.stock, &p.publicado, &p.opciones, &p.variaciones); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT id::text, valores, sku, precio, stock, foto, activa FROM variantes
			WHERE tienda_id = $1 AND producto_id::text = $2`, tiendaID, productoID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var v listas.Variante
			if err := filas.Scan(&v.ID, &v.Valores, &v.SKU, &v.Precio, &v.Stock, &v.Foto, &v.Activa); err != nil {
				filas.Close()
				return err
			}
			p.variantes = append(p.variantes, v)
		}
		if err := filas.Err(); err != nil {
			return err
		}
		listas.OrdenarVariantes(p.opciones, p.variantes)
		return t.QueryRow(ctx, `SELECT recargo, pausar_sin_stock FROM ml_conexiones WHERE tienda_id = $1`, tiendaID).
			Scan(&recargo, &pausar)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return p, 0, false, ErrNoExiste
	}
	return p, recargo, pausar, err
}

// precioML: el precio de la tienda con el recargo (para cubrir la comisión), en pesos enteros.
func precioML(precio int64, recargo int) int64 { return (precio*int64(100+recargo) + 50) / 100 }

func recortar(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) > n {
		return strings.TrimSpace(string(r[:n]))
	}
	return string(r)
}

// Publicacion es lo que devuelve publicar.
type Publicacion struct {
	ItemID    string `json:"itemId"`
	Permalink string `json:"permalink"`
}

// Publicar crea la publicación del producto: categoría sugerida por Mercado Libre para el título,
// precio con el recargo, stock, foto y descripción.
func (s *Servicio) Publicar(ctx context.Context, tiendaID, productoID string) (Publicacion, error) {
	var ya bool
	_ = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT true FROM ml_publicaciones WHERE tienda_id = $1 AND producto_id::text = $2`, tiendaID, productoID).Scan(&ya)
	})
	if ya {
		return Publicacion{}, ErrYaPublicado
	}
	p, recargo, _, err := s.leerProducto(ctx, tiendaID, productoID)
	if err != nil {
		return Publicacion{}, err
	}
	if p.stock < 1 || p.precio < 1 {
		return Publicacion{}, ErrInvalido
	}
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return Publicacion{}, err
	}
	titulo := recortar(p.nombre, 60)
	var sugerencias []struct {
		CategoryID string `json:"category_id"`
		Attributes []struct {
			ID        string `json:"id"`
			ValueID   string `json:"value_id"`
			ValueName string `json:"value_name"`
		} `json:"attributes"`
	}
	if err := s.pedir(ctx, http.MethodGet, "/sites/MLA/domain_discovery/search?limit=1&q="+url.QueryEscape(titulo), token, nil, &sugerencias); err != nil {
		return Publicacion{}, err
	}
	if len(sugerencias) == 0 || sugerencias[0].CategoryID == "" {
		return Publicacion{}, &ErrorML{Estado: http.StatusBadRequest, Mensaje: "Mercado Libre no encontró una categoría para ese título"}
	}
	atributos := []map[string]string{}
	marca := false
	for _, a := range sugerencias[0].Attributes {
		v := map[string]string{"id": a.ID}
		if a.ValueID != "" {
			v["value_id"] = a.ValueID
		} else if a.ValueName != "" {
			v["value_name"] = a.ValueName
		} else {
			continue
		}
		marca = marca || a.ID == "BRAND"
		atributos = append(atributos, v)
	}
	if !marca {
		atributos = append(atributos, map[string]string{"id": "BRAND", "value_name": "Genérica"})
	}
	cuerpo := map[string]any{
		"title":              titulo,
		"category_id":        sugerencias[0].CategoryID,
		"price":              precioML(p.precio, recargo),
		"currency_id":        "ARS",
		"available_quantity": p.stock,
		"buying_mode":        "buy_it_now",
		"listing_type_id":    "gold_special",
		"condition":          "new",
		"attributes":         atributos,
	}
	// La galería: sólo las fotos con URL pública (Mercado Libre las descarga).
	fotos := []map[string]string{}
	for _, f := range p.fotos {
		if strings.HasPrefix(f, "https://") {
			fotos = append(fotos, map[string]string{"source": f})
		}
	}
	if len(fotos) > 0 {
		cuerpo["pictures"] = fotos
	}
	var publicadas []listas.Variante
	if len(p.opciones) > 0 {
		var variaciones []map[string]any
		var precio int64
		variaciones, publicadas, precio, err = s.armarVariaciones(ctx, token, sugerencias[0].CategoryID, p, recargo)
		if err != nil {
			return Publicacion{}, err
		}
		cuerpo["variations"] = variaciones
		cuerpo["price"] = precio
		delete(cuerpo, "available_quantity") // la suma la hace Mercado Libre
	}
	var item struct {
		ID         string `json:"id"`
		Permalink  string `json:"permalink"`
		Status     string `json:"status"`
		Variations []struct {
			ID int64 `json:"id"`
		} `json:"variations"`
	}
	if err := s.pedir(ctx, http.MethodPost, "/items", token, cuerpo, &item); err != nil {
		var em *ErrorML
		if len(publicadas) > 0 && errors.As(err, &em) && em.Estado < 500 {
			return Publicacion{}, &ErrVariaciones{Mensaje: "Mercado Libre no aceptó las variantes de este producto: " + em.Mensaje}
		}
		return Publicacion{}, err
	}
	// Mercado Libre devuelve las variaciones en el orden en que se mandaron.
	variaciones := map[string]int64{}
	for i, v := range item.Variations {
		if i < len(publicadas) {
			variaciones[publicadas[i].ID] = v.ID
		}
	}
	if strings.TrimSpace(p.descripcion) != "" {
		// La descripción va aparte; si falla, la publicación igual queda.
		if err := s.pedir(ctx, http.MethodPost, "/items/"+item.ID+"/description", token,
			map[string]string{"plain_text": recortar(p.descripcion, 4000)}, nil); err != nil {
			s.log().Warn("mercadolibre: no se pudo cargar la descripción", "item", item.ID, "error", err)
		}
	}
	err = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `INSERT INTO ml_publicaciones (tienda_id, producto_id, item_id, permalink, estado, sincronizado_en, variaciones)
			VALUES ($1, $2::uuid, $3, $4, $5, now(), $6)`, tiendaID, productoID, item.ID, item.Permalink, item.Status, variaciones)
		return err
	})
	return Publicacion{ItemID: item.ID, Permalink: item.Permalink}, err
}

// Sincronizar manda stock, precio y estado de un producto publicado. Sin stock (o despublicado en la
// tienda) la publicación se pausa; con stock de nuevo, se reactiva.
func (s *Servicio) Sincronizar(ctx context.Context, tiendaID, productoID string) error {
	var itemID, estado string
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT item_id, estado FROM ml_publicaciones WHERE tienda_id = $1 AND producto_id::text = $2`,
			tiendaID, productoID).Scan(&itemID, &estado)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // se despublicó
	}
	if err != nil {
		return err
	}
	if estado == "closed" {
		return nil
	}
	p, recargo, pausar, err := s.leerProducto(ctx, tiendaID, productoID)
	if errors.Is(err, ErrNoExiste) {
		return nil
	}
	if err != nil {
		return err
	}
	token, _, err := s.tokenDe(ctx, tiendaID)
	if errors.Is(err, ErrSinConexion) || errors.Is(err, ErrReautorizar) {
		return nil // sin cuenta no hay nada que sincronizar; al reconectar se vuelve a mandar
	}
	if err != nil {
		return err
	}
	cuerpo := map[string]any{"price": precioML(p.precio, recargo)}
	nuevo := "active"
	if !p.publicado || (p.stock == 0 && pausar) {
		nuevo = "paused"
	} else if len(p.variaciones) == 0 {
		cuerpo["available_quantity"] = p.stock
	}
	if len(p.variaciones) > 0 {
		// Stock por variación. Una variante borrada o desactivada queda en 0 (la variación sigue en
		// Mercado Libre); una variante nueva no tiene variación hasta volver a publicar.
		porVariante := map[string]listas.Variante{}
		for _, v := range p.variantes {
			porVariante[v.ID] = v
		}
		vars := []map[string]any{}
		for _, id := range ordenVariaciones(p) {
			v, ok := porVariante[id]
			stock, precio := 0, p.precio
			if ok && v.Activa {
				stock = v.Stock
			}
			if ok && v.Precio != nil {
				precio = *v.Precio
			}
			vars = append(vars, map[string]any{"id": p.variaciones[id], "available_quantity": stock, "price": precioML(precio, recargo)})
		}
		cuerpo["variations"] = vars
		delete(cuerpo, "price")
	}
	if nuevo != estado {
		cuerpo["status"] = nuevo
	}
	errML := s.pedir(ctx, http.MethodPut, "/items/"+itemID, token, cuerpo, nil)
	var em *ErrorML
	if errors.As(errML, &em) && em.Estado < 500 {
		// Un rechazo de Mercado Libre no se arregla reintentando: queda anotado para mostrarlo.
		_ = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
			_, err := t.Exec(ctx, `UPDATE ml_publicaciones SET error = $3 WHERE tienda_id = $1 AND producto_id::text = $2`,
				tiendaID, productoID, recortar(em.Mensaje, 300))
			return err
		})
		return nil
	}
	if errML != nil {
		return errML
	}
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE ml_publicaciones SET estado = $3, sincronizado_en = now(), error = NULL
			WHERE tienda_id = $1 AND producto_id::text = $2`, tiendaID, productoID, nuevo)
		return err
	})
}

// Despublicar pausa la publicación y deja de sincronizarla.
func (s *Servicio) Despublicar(ctx context.Context, tiendaID, productoID string) error {
	var itemID string
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT item_id FROM ml_publicaciones WHERE tienda_id = $1 AND producto_id::text = $2`,
			tiendaID, productoID).Scan(&itemID)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoExiste
	}
	if err != nil {
		return err
	}
	if token, _, err := s.tokenDe(ctx, tiendaID); err == nil {
		if err := s.pedir(ctx, http.MethodPut, "/items/"+itemID, token, map[string]string{"status": "paused"}, nil); err != nil {
			return err
		}
	}
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `DELETE FROM ml_publicaciones WHERE tienda_id = $1 AND producto_id::text = $2`, tiendaID, productoID)
		return err
	})
}

// ErrVariaciones: el producto tiene variantes y no se pueden publicar como variaciones de Mercado
// Libre. El mensaje dice por qué (va tal cual al panel); no se publica nada.
type ErrVariaciones struct{ Mensaje string }

func (e *ErrVariaciones) Error() string { return "mercadolibre: " + e.Mensaje }

// alias: nombres de opción comunes y el atributo de Mercado Libre que suelen ser.
var alias = map[string][]string{
	"talle": {"SIZE"}, "talla": {"SIZE"}, "numero": {"SIZE", "SHOE_SIZE"}, "tamano": {"SIZE"},
	"color": {"COLOR", "MAIN_COLOR"}, "sabor": {"FLAVOR"}, "aroma": {"FRAGRANCE", "SCENT"},
}

// armarVariaciones mapea cada opción a un atributo de la categoría que admita variaciones (por nombre
// o por alias); la que no mapea va con su nombre tal cual (atributo personalizado) y, si Mercado
// Libre no lo acepta, el POST falla y se avisa. Mercado Libre pide un solo precio para todas las
// variaciones: si las variantes tienen precios distintos, no se publica.
func (s *Servicio) armarVariaciones(ctx context.Context, token, categoria string, p producto, recargo int) ([]map[string]any, []listas.Variante, int64, error) {
	var atributos []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Tags struct {
			AllowVariations    bool `json:"allow_variations"`
			VariationAttribute bool `json:"variation_attribute"`
		} `json:"tags"`
		Values []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"values"`
	}
	if err := s.pedir(ctx, http.MethodGet, "/categories/"+url.PathEscape(categoria)+"/attributes", token, nil, &atributos); err != nil {
		return nil, nil, 0, err
	}
	type atributo struct {
		id      string
		valores map[string]string // nombre normalizado → value_id
	}
	admite := map[string]atributo{} // por id
	porNombre := map[string]string{}
	for _, a := range atributos {
		if !a.Tags.AllowVariations && !a.Tags.VariationAttribute {
			continue
		}
		at := atributo{id: a.ID, valores: map[string]string{}}
		for _, v := range a.Values {
			at.valores[normal(v.Name)] = v.ID
		}
		admite[a.ID] = at
		porNombre[normal(a.Name)] = a.ID
	}
	if len(admite) == 0 {
		return nil, nil, 0, &ErrVariaciones{Mensaje: "La categoría que sugiere Mercado Libre para este producto no admite variantes. Publicalo sin opciones o como productos separados."}
	}
	mapa := map[string]string{} // opción → atributo ("" = personalizado)
	usados := map[string]bool{}
	for _, o := range p.opciones {
		id := porNombre[normal(o.Nombre)]
		if id == "" {
			for _, a := range alias[normal(o.Nombre)] {
				if _, ok := admite[a]; ok {
					id = a
					break
				}
			}
		}
		if id != "" && usados[id] {
			return nil, nil, 0, &ErrVariaciones{Mensaje: fmt.Sprintf("Dos opciones (%s y otra) son el mismo atributo en Mercado Libre.", o.Nombre)}
		}
		usados[id] = id != ""
		mapa[o.Nombre] = id
	}
	var activas []listas.Variante
	var precio *int64
	for _, v := range p.variantes {
		if !v.Activa {
			continue
		}
		pv := p.precio
		if v.Precio != nil {
			pv = *v.Precio
		}
		if precio != nil && *precio != pv {
			return nil, nil, 0, &ErrVariaciones{Mensaje: "Mercado Libre pide el mismo precio para todas las variantes. Igualá los precios o publicalas por separado."}
		}
		precio = &pv
		activas = append(activas, v)
	}
	if len(activas) == 0 || precio == nil || *precio < 1 {
		return nil, nil, 0, ErrInvalido
	}
	if len(activas) > 100 {
		return nil, nil, 0, &ErrVariaciones{Mensaje: "Mercado Libre admite hasta 100 variantes por publicación."}
	}
	out := make([]map[string]any, 0, len(activas))
	for _, v := range activas {
		combinacion := []map[string]string{}
		for _, o := range p.opciones {
			valor := v.Valores[o.Nombre]
			c := map[string]string{"value_name": valor}
			if id := mapa[o.Nombre]; id != "" {
				c["id"] = id
				if vid := admite[id].valores[normal(valor)]; vid != "" {
					c["value_id"] = vid
				}
			} else {
				c["name"] = o.Nombre
			}
			combinacion = append(combinacion, c)
		}
		x := map[string]any{"attribute_combinations": combinacion, "price": precioML(*precio, recargo), "available_quantity": v.Stock}
		if v.SKU != "" {
			x["seller_custom_field"] = v.SKU
		}
		if v.Foto != nil && strings.HasPrefix(*v.Foto, "https://") {
			x["picture_ids"] = []string{*v.Foto}
		} else if len(p.fotos) > 0 && strings.HasPrefix(p.fotos[0], "https://") {
			x["picture_ids"] = []string{p.fotos[0]}
		}
		out = append(out, x)
	}
	return out, activas, precioML(*precio, recargo), nil
}

// ordenVariaciones: las variantes publicadas, en el orden del producto (las borradas al final).
func ordenVariaciones(p producto) []string {
	ids := []string{}
	for _, v := range p.variantes {
		if _, ok := p.variaciones[v.ID]; ok {
			ids = append(ids, v.ID)
		}
	}
	resto := []string{}
	for id := range p.variaciones {
		if !slices.Contains(ids, id) {
			resto = append(resto, id)
		}
	}
	slices.Sort(resto)
	return append(ids, resto...)
}

// normal: minúsculas y sin tildes, para comparar nombres de opciones y atributos.
func normal(s string) string {
	r := strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")
	return r.Replace(strings.ToLower(strings.TrimSpace(s)))
}
