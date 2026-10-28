package ventas

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Tarjeta es la tarjeta A6 para los paquetes (la misma regla que lib/panel/tarjeta.ts de la web).
type Tarjeta struct {
	Modelo      string `json:"modelo"`
	Titulo      string `json:"titulo"`
	Mensaje     string `json:"mensaje"`
	DorsoTitulo string `json:"dorsoTitulo"`
	DorsoTexto  string `json:"dorsoTexto"`
	QR          string `json:"qr"`
	Cupon       string `json:"cupon"`
	Instagram   string `json:"instagram"`
}

// TarjetaInicial es la de fábrica: la del pedido de Lucía de Imprimibles.
var TarjetaInicial = Tarjeta{
	Modelo:      "arco",
	Titulo:      "¡Gracias, {nombre}!",
	Mensaje:     "Lo preparamos a mano el {fecha}. Que lo disfrutes.",
	DorsoTitulo: "¿Te gustó?",
	DorsoTexto:  "Dejanos una reseña con foto: nos ayuda un montón a seguir creciendo.",
	QR:          "resena",
}

var (
	cuponTarjeta     = regexp.MustCompile(`^[A-Z0-9-]{0,20}$`)
	instagramTarjeta = regexp.MustCompile(`^[a-zA-Z0-9._]{0,30}$`)
)

// Valida devuelve el error para mostrar, o "".
func (t Tarjeta) Valida() string {
	switch {
	case t.Modelo != "arco" && t.Modelo != "sello" && t.Modelo != "carta":
		return "Elegí un modelo."
	case t.QR != "resena" && t.QR != "cupon" && t.QR != "instagram" && t.QR != "whatsapp":
		return "Elegí adónde lleva el QR."
	case strings.TrimSpace(t.Titulo) == "":
		return "Poné un título."
	case utf8.RuneCountInString(t.Titulo) > 40 || utf8.RuneCountInString(t.DorsoTitulo) > 40:
		return "El texto es más largo que 40 caracteres."
	case utf8.RuneCountInString(t.Mensaje) > 140 || utf8.RuneCountInString(t.DorsoTexto) > 140:
		return "El texto es más largo que 140 caracteres."
	case !cuponTarjeta.MatchString(t.Cupon):
		return "Ese cupón no existe."
	case !instagramTarjeta.MatchString(t.Instagram):
		return "Revisá la cuenta de Instagram."
	}
	return ""
}

// LeerTarjeta devuelve la de la tienda, o la de fábrica.
func LeerTarjeta(ctx context.Context, pool tx, tiendaID string) (Tarjeta, error) {
	t := TarjetaInicial
	err := pool.EnTienda(ctx, tiendaID, func(x pgx.Tx) error {
		var raw []byte
		if err := x.QueryRow(ctx, `SELECT tarjeta FROM tiendas WHERE id = $1`, tiendaID).Scan(&raw); err != nil {
			return err
		}
		if raw != nil {
			return json.Unmarshal(raw, &t)
		}
		return nil
	})
	return t, err
}

// GuardarTarjeta la guarda (ya validada).
func GuardarTarjeta(ctx context.Context, pool tx, tiendaID string, t Tarjeta) error {
	return pool.EnTienda(ctx, tiendaID, func(x pgx.Tx) error {
		_, err := x.Exec(ctx, `UPDATE tiendas SET tarjeta = $2 WHERE id = $1`, tiendaID, t)
		return err
	})
}

func rutasTarjeta(mux *http.ServeMux, pool tx, con func(func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc) {
	mux.HandleFunc("GET /v1/panel/tarjeta", con(func(w http.ResponseWriter, r *http.Request, id string) {
		t, err := LeerTarjeta(r.Context(), pool, id)
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, t)
	}))
	mux.HandleFunc("PUT /v1/panel/tarjeta", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var t Tarjeta
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&t); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá la tarjeta."))
			return
		}
		if msg := t.Valida(); msg != "" {
			problema.Escribir(w, problema.Invalido(msg))
			return
		}
		if err := GuardarTarjeta(r.Context(), pool, id, t); err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}
