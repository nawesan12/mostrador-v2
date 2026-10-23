package mercadolibre

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
)

type producto struct {
	nombre, descripcion string
	foto                *string
	precio              int64
	stock               int
	publicado           bool
}

func (s *Servicio) leerProducto(ctx context.Context, tiendaID, productoID string) (producto, int, bool, error) {
	var p producto
	var recargo int
	var pausar bool
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT nombre, descripcion, foto, precio, stock, publicado FROM productos
			WHERE tienda_id = $1 AND id::text = $2`, tiendaID, productoID).
			Scan(&p.nombre, &p.descripcion, &p.foto, &p.precio, &p.stock, &p.publicado); err != nil {
			return err
		}
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
	if p.foto != nil && strings.HasPrefix(*p.foto, "https://") {
		cuerpo["pictures"] = []map[string]string{{"source": *p.foto}}
	}
	var item struct {
		ID        string `json:"id"`
		Permalink string `json:"permalink"`
		Status    string `json:"status"`
	}
	if err := s.pedir(ctx, http.MethodPost, "/items", token, cuerpo, &item); err != nil {
		return Publicacion{}, err
	}
	if strings.TrimSpace(p.descripcion) != "" {
		// La descripción va aparte; si falla, la publicación igual queda.
		if err := s.pedir(ctx, http.MethodPost, "/items/"+item.ID+"/description", token,
			map[string]string{"plain_text": recortar(p.descripcion, 4000)}, nil); err != nil {
			s.log().Warn("mercadolibre: no se pudo cargar la descripción", "item", item.ID, "error", err)
		}
	}
	err = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `INSERT INTO ml_publicaciones (tienda_id, producto_id, item_id, permalink, estado, sincronizado_en)
			VALUES ($1, $2::uuid, $3, $4, $5, now())`, tiendaID, productoID, item.ID, item.Permalink, item.Status)
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
	} else {
		cuerpo["available_quantity"] = p.stock
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
