package tiendanube

import (
	"strconv"
	"strings"
	"time"
)

// ClienteTN es lo que se usa del cliente de la API (GET /customers).
type ClienteTN struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	Note           *string `json:"note"`
	DefaultAddress *struct {
		City     *string `json:"city"`
		Locality *string `json:"locality"`
		Province *string `json:"province"`
	} `json:"default_address"`
	BillingCity     *string `json:"billing_city"`
	BillingProvince *string `json:"billing_province"`
}

// Cliente es la ficha que se guarda en Mostrador.
type Cliente struct {
	Externo   string `json:"externo"`
	Nombre    string `json:"nombre"`
	Email     string `json:"email"`
	Telefono  string `json:"telefono"`
	Ciudad    string `json:"ciudad"`
	Provincia string `json:"provincia"`
	Notas     string `json:"notas"`
}

// ClienteDe convierte un cliente. Sin nombre ni email no hay a quién guardar.
func ClienteDe(c ClienteTN) (Cliente, bool) {
	r := Cliente{
		Externo:  strconv.FormatInt(c.ID, 10),
		Nombre:   corto(c.Name, 80),
		Email:    strings.ToLower(corto(valor(c.Email), 120)),
		Telefono: corto(valor(c.Phone), 30),
		Notas:    corto(valor(c.Note), 500),
	}
	if d := c.DefaultAddress; d != nil {
		r.Ciudad = corto(primero(valor(d.City), valor(d.Locality)), 60)
		r.Provincia = corto(valor(d.Province), 60)
	}
	if r.Ciudad == "" {
		r.Ciudad = corto(valor(c.BillingCity), 60)
	}
	if r.Provincia == "" {
		r.Provincia = corto(valor(c.BillingProvince), 60)
	}
	if r.Nombre == "" {
		r.Nombre = r.Email
	}
	if r.Nombre == "" || (r.Email != "" && !strings.Contains(r.Email, "@")) {
		return Cliente{}, false
	}
	return r, true
}

// PedidoTN es lo que se usa del pedido de la API (GET /orders).
type PedidoTN struct {
	ID            int64   `json:"id"`
	Number        int64   `json:"number"`
	CreatedAt     string  `json:"created_at"`
	Total         *string `json:"total"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	PaymentStatus string  `json:"payment_status"`
	ContactEmail  *string `json:"contact_email"`
	ContactName   *string `json:"contact_name"`
	Customer      *struct {
		Name  string  `json:"name"`
		Email *string `json:"email"`
	} `json:"customer"`
	Products []struct {
		Name          string  `json:"name"`
		Price         *string `json:"price"`
		Quantity      any     `json:"quantity"`
		VariantValues []any   `json:"variant_values"`
	} `json:"products"`
}

// Item es un renglón del pedido viejo.
type Item struct {
	Nombre   string `json:"nombre"`
	Variante string `json:"variante"`
	Cantidad int    `json:"cantidad"`
	Precio   int64  `json:"precio"`
}

// Pedido es un pedido viejo para el historial del cliente. No entra en Pedidos ni toca el stock.
type Pedido struct {
	Externo       string    `json:"externo"`
	NumeroExterno string    `json:"numeroExterno"`
	Fecha         time.Time `json:"fecha"`
	ClienteEmail  string    `json:"clienteEmail"`
	ClienteNombre string    `json:"clienteNombre"`
	Total         int64     `json:"total"`
	Estado        string    `json:"estado"`
	Items         []Item    `json:"items"`
}

// PedidoDe convierte un pedido. Sólo los pagos en pesos: los cancelados y los que no se cobraron no
// cuentan para la historia del cliente.
func PedidoDe(p PedidoTN) (Pedido, bool) {
	if p.Currency != "" && p.Currency != "ARS" {
		return Pedido{}, false
	}
	switch p.PaymentStatus {
	case "paid", "partially_paid", "partially_refunded", "authorized":
	default:
		return Pedido{}, false
	}
	fecha, ok := fechaDe(p.CreatedAt)
	if p.Status == "cancelled" || !ok {
		return Pedido{}, false
	}
	r := Pedido{
		Externo:       strconv.FormatInt(p.ID, 10),
		NumeroExterno: strconv.FormatInt(p.Number, 10),
		Fecha:         fecha,
		ClienteEmail:  strings.ToLower(valor(p.ContactEmail)),
		ClienteNombre: corto(valor(p.ContactName), 80),
		Total:         pesos(p.Total),
		Estado:        p.Status,
	}
	if c := p.Customer; c != nil {
		if r.ClienteEmail == "" {
			r.ClienteEmail = strings.ToLower(valor(c.Email))
		}
		if r.ClienteNombre == "" {
			r.ClienteNombre = corto(c.Name, 80)
		}
	}
	for _, it := range p.Products {
		var variante []string
		for _, v := range it.VariantValues {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				variante = append(variante, strings.TrimSpace(s))
			}
		}
		r.Items = append(r.Items, Item{
			Nombre:   corto(it.Name, 120),
			Variante: corto(strings.Join(variante, " / "), 60),
			Cantidad: max(1, entero(it.Quantity)),
			Precio:   pesos(it.Price),
		})
	}
	if r.ClienteEmail == "" && r.ClienteNombre == "" {
		return Pedido{}, false
	}
	return r, true
}

// fechaDe lee las fechas de la API, que vienen como RFC 3339 o con la zona sin dos puntos
// ("2022-11-15T19:36:59+0000").
func fechaDe(s string) (time.Time, bool) {
	for _, f := range []string{time.RFC3339, "2006-01-02T15:04:05-0700"} {
		if t, err := time.Parse(f, strings.TrimSpace(s)); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

// entero lee la cantidad, que la API manda como número o como texto ("1").
func entero(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(x))
		return n
	}
	return 0
}

func valor(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func primero(xs ...string) string {
	for _, x := range xs {
		if x != "" {
			return x
		}
	}
	return ""
}

func corto(s string, n int) string {
	s = espacios(s)
	if r := []rune(s); len(r) > n {
		return strings.TrimSpace(string(r[:n]))
	}
	return s
}
