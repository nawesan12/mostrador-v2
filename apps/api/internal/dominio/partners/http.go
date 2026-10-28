package partners

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Rutas del programa. Las de /v1/partners/yo son de la persona de la sesión; el clic y las de
// /v1/interno/partners sólo se aceptan con la clave del BFF (las internas las va a usar el
// backoffice, que todavía no existe).
func Rutas(mux *http.ServeMux, pool *db.Pool, claveBFF string, ahora func() time.Time) {
	if ahora == nil {
		ahora = time.Now
	}
	persona := func(w http.ResponseWriter, r *http.Request) (string, bool) {
		s, ok := cuentas.SesionDe(r)
		// Una IA conectada por MCP no entra al programa: es de la persona, no de una tienda.
		if !ok || s.Usuario.ID == "" || s.IA != "" {
			problema.Escribir(w, problema.SinSesion())
			return "", false
		}
		return s.Usuario.ID, true
	}
	delBFF := func(w http.ResponseWriter, r *http.Request) bool {
		if claveBFF == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Mostrador-BFF")), []byte(claveBFF)) != 1 {
			problema.Escribir(w, problema.SinPermiso("Sólo el equipo de Mostrador."))
			return false
		}
		return true
	}

	mux.HandleFunc("GET /v1/partners/yo", func(w http.ResponseWriter, r *http.Request) {
		if u, ok := persona(w, r); ok {
			p, err := Leer(r.Context(), pool, u, ahora())
			responder(w, p, err)
		}
	})
	mux.HandleFunc("PUT /v1/partners/yo/perfil", func(w http.ResponseWriter, r *http.Request) {
		var p Perfil
		if u, ok := persona(w, r); ok && leer(w, r, &p) {
			p, err := GuardarPerfil(r.Context(), pool, u, p)
			responder(w, p, err)
		}
	})
	mux.HandleFunc("PUT /v1/partners/yo/cobro", func(w http.ResponseWriter, r *http.Request) {
		var c Cobro
		if u, ok := persona(w, r); ok && leer(w, r, &c) {
			c, err := GuardarCobro(r.Context(), pool, u, c)
			responder(w, c, err)
		}
	})
	mux.HandleFunc("PUT /v1/partners/yo/pagos/{mes}/factura", func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			URL string `json:"url"`
		}
		if u, ok := persona(w, r); ok && leer(w, r, &a) {
			responder(w, nil, SubirFactura(r.Context(), pool, u, r.PathValue("mes"), a.URL))
		}
	})
	mux.HandleFunc("POST /v1/partners/clics", func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			Ref string `json:"ref"`
		}
		if delBFF(w, r) && leer(w, r, &a) {
			_, err := SumarClic(r.Context(), pool, a.Ref)
			responder(w, nil, err)
		}
	})

	mux.HandleFunc("POST /v1/interno/partners", func(w http.ResponseWriter, r *http.Request) {
		var a Aprobacion
		if delBFF(w, r) && leer(w, r, &a) {
			c, err := Aprobar(r.Context(), pool, a)
			responder(w, map[string]string{"codigo": c}, err)
		}
	})
	mux.HandleFunc("POST /v1/interno/partners/liquidaciones", func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			Email string `json:"email"`
			Mes   string `json:"mes"`
		}
		if delBFF(w, r) && leer(w, r, &a) {
			g, err := Liquidar(r.Context(), pool, a.Email, a.Mes, ahora())
			responder(w, g, err)
		}
	})
	mux.HandleFunc("PUT /v1/interno/partners/pagos", func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			Email  string `json:"email"`
			Mes    string `json:"mes"`
			Estado string `json:"estado"`
		}
		if delBFF(w, r) && leer(w, r, &a) {
			responder(w, nil, CambiarEstado(r.Context(), pool, a.Email, a.Mes, a.Estado))
		}
	})
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(v); err != nil {
		problema.Escribir(w, problema.Invalido("El pedido no es un JSON válido."))
		return false
	}
	return true
}

func responder(w http.ResponseWriter, v any, err error) {
	var inv ErrInvalido
	switch {
	case err == nil && v == nil:
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusNoContent)
	case err == nil:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(v)
	case errors.As(err, &inv):
		problema.Escribir(w, problema.Invalido(inv.Mensaje, problema.Campo{Ruta: inv.Campo, Mensaje: inv.Mensaje}))
	case errors.Is(err, ErrNoEsPartner):
		problema.Escribir(w, problema.NoEncontrado("Tu cuenta todavía no es partner. Postulate en /partners."))
	case errors.Is(err, ErrSinCuenta):
		problema.Escribir(w, problema.NoEncontrado("No hay una cuenta con ese email: primero tiene que crearla."))
	case errors.Is(err, ErrSinPago):
		problema.Escribir(w, problema.NoEncontrado("No hay un pago abierto de ese mes."))
	default:
		slog.Error("partners", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}
