package cuentas_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/clerk"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/clerk/clerkfalso"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/whatsapp"
)

// proveedorFalso es un servidor OAuth: valida el PKCE contra el desafío del inicio y devuelve el
// perfil que se le cargue.
type proveedorFalso struct {
	mu      sync.Mutex
	desafio string
	perfil  map[string]any
}

func (f *proveedorFalso) servidor(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		switch r.URL.Path {
		case "/token":
			_ = r.ParseForm()
			h := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
			if base64.RawURLEncoding.EncodeToString(h[:]) != f.desafio || r.PostForm.Get("code") != "cod" {
				http.Error(w, "pkce", 400)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "at"})
		case "/perfil":
			if r.Header.Get("Authorization") != "Bearer at" {
				http.Error(w, "token", 401)
				return
			}
			_ = json.NewEncoder(w).Encode(f.perfil)
		}
	}))
}

func conProveedores(t *testing.T) (*cuentas.Servicio, *proveedorFalso, *whatsapp.Memoria) {
	s, _, _ := nuevo(t)
	f := &proveedorFalso{}
	srv := f.servidor(t)
	t.Cleanup(srv.Close)
	ps := cuentas.Proveedores(map[string][2]string{"mercadopago": {"mp", "sec"}, "mercadolibre": {"ml", "sec"}})
	for _, p := range ps {
		p.AuthURL, p.TokenURL, p.PerfilURL = srv.URL+"/auth", srv.URL+"/token", srv.URL+"/perfil"
	}
	s.OAuth = ps
	wa := &whatsapp.Memoria{}
	s.WhatsApp = wa
	return s, f, wa
}

// iniciar devuelve el state y deja el desafío PKCE en el proveedor falso.
func iniciar(t *testing.T, s *cuentas.Servicio, f *proveedorFalso, prov string) string {
	t.Helper()
	u, err := s.IniciarOAuth(ctx, prov, cl)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.Parse(u)
	if q.Query().Get("redirect_uri") != "https://tulocalweb.com.ar/ingresar/"+prov+"/vuelta" || q.Query().Get("code_challenge_method") != "S256" {
		t.Fatalf("url %s", u)
	}
	f.mu.Lock()
	f.desafio = q.Query().Get("code_challenge")
	f.mu.Unlock()
	return q.Query().Get("state")
}

func TestOAuthCuentaNuevaYDespuesLaMisma(t *testing.T) {
	s, f, _ := conProveedores(t)
	f.perfil = map[string]any{"id": 501, "email": "Lucia@Gmail.com", "first_name": "Lucía"}
	r, err := s.VolverOAuth(ctx, "mercadopago", "cod", iniciar(t, s, f, "mercadopago"), cl)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Nueva || r.Token == "" || r.TiendaID == "" {
		t.Fatalf("%+v", r)
	}
	d, _ := s.Sesion(ctx, r.Token)
	if d.Usuario.Email != "lucia@gmail.com" || d.Usuario.Nombre != "Lucía" || d.Rol != "duenio" {
		t.Fatalf("%+v", d)
	}
	// La segunda vez entra a la misma cuenta.
	r2, err := s.VolverOAuth(ctx, "mercadopago", "cod", iniciar(t, s, f, "mercadopago"), cl)
	if err != nil || r2.Nueva || r2.TiendaID != r.TiendaID {
		t.Fatalf("%+v %v", r2, err)
	}
}

func TestOAuthStateDeUnSoloUsoYProveedorApagado(t *testing.T) {
	s, f, _ := conProveedores(t)
	f.perfil = map[string]any{"id": 502, "email": "ana@gmail.com"}
	estado := iniciar(t, s, f, "mercadopago")
	if _, err := s.VolverOAuth(ctx, "mercadopago", "cod", estado, cl); err != nil {
		t.Fatal(err)
	}
	if _, err := s.VolverOAuth(ctx, "mercadopago", "cod", estado, cl); !errors.Is(err, cuentas.ErrEstadoInvalido) {
		t.Fatalf("reusar el state: %v", err)
	}
	if _, err := s.VolverOAuth(ctx, "mercadolibre", "cod", iniciar(t, s, f, "mercadopago"), cl); !errors.Is(err, cuentas.ErrEstadoInvalido) {
		t.Fatalf("state de otro proveedor: %v", err)
	}
	// Google ya no es un OAuth propio: va por Clerk.
	if _, err := s.IniciarOAuth(ctx, "google", cl); !errors.Is(err, cuentas.ErrProveedorApagado) {
		t.Fatalf("google: %v", err)
	}
	if got := s.Disponibles(); strings.Join(got, ",") != "mercadopago,mercadolibre" {
		t.Fatalf("disponibles %v", got)
	}
}

// conClerk suma al servicio una instancia de Clerk falsa (JWKS y Backend API en httptest).
func conClerk(t *testing.T, s *cuentas.Servicio) *clerkfalso.Servidor {
	f := clerkfalso.Nuevo(t)
	s.Clerk = &clerk.Cliente{Secreto: clerkfalso.Secreto, Emisor: clerkfalso.Emisor,
		Origenes: []string{"https://tulocalweb.com.ar"}, API: f.URL, HTTP: f.Client()}
	return f
}

func TestClerkCuentaNuevaYDespuesLaMisma(t *testing.T) {
	s, m, _ := nuevo(t)
	f := conClerk(t, s)
	f.Usuarios["user_lu"] = clerkfalso.Usuario{Nombre: "Lucía", Apellido: "Fernández", Email: "Lucia@Gmail.com",
		Verificado: true, Foto: "https://img.clerk.com/lu"}
	tok := f.Token(t, clerkfalso.Reclamos{Sub: "user_lu", Azp: "https://tulocalweb.com.ar"})
	r, err := s.IngresarConClerk(ctx, tok, cl)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Nueva || r.Token == "" || r.TiendaID == "" {
		t.Fatalf("%+v", r)
	}
	d, _ := s.Sesion(ctx, r.Token)
	if d.Usuario.Email != "lucia@gmail.com" || d.Usuario.Nombre != "Lucía" || d.Usuario.Foto != "https://img.clerk.com/lu" || d.Rol != "duenio" {
		t.Fatalf("%+v", d)
	}
	// Mismo camino que el registro: mail de bienvenida.
	if len(m.Enviados) != 1 || m.Enviados[0].Para != "lucia@gmail.com" {
		t.Fatalf("mails %+v", m.Enviados)
	}
	// La app de iPhone manda un token sin origen (azp): entra a la misma cuenta.
	r2, err := s.IngresarConClerk(ctx, f.Token(t, clerkfalso.Reclamos{Sub: "user_lu"}), cl)
	if err != nil || r2.Nueva || r2.TiendaID != r.TiendaID {
		t.Fatalf("%+v %v", r2, err)
	}
}

func TestClerkVinculaLaCuentaDelMismoEmail(t *testing.T) {
	s, f, _ := conProveedores(t)
	ses, err := s.Registrar(ctx, alta("vale@laesquina.com.ar"), cl)
	if err != nil {
		t.Fatal(err)
	}
	// Mercado Pago no verifica el email: no se vincula sola a la cuenta con contraseña.
	f.perfil = map[string]any{"id": 77, "email": "vale@laesquina.com.ar", "first_name": "Vale"}
	if _, err := s.VolverOAuth(ctx, "mercadopago", "cod", iniciar(t, s, f, "mercadopago"), cl); !errors.Is(err, cuentas.ErrVincular) {
		t.Fatalf("err = %v", err)
	}
	// Clerk sí (sólo da emails verificados): queda vinculada y entra a la misma tienda.
	c := conClerk(t, s)
	c.Usuarios["user_vale"] = clerkfalso.Usuario{Nombre: "Vale", Email: "vale@laesquina.com.ar", Verificado: true}
	r, err := s.IngresarConClerk(ctx, c.Token(t, clerkfalso.Reclamos{Sub: "user_vale"}), cl)
	if err != nil || r.Nueva || r.TiendaID != ses.TiendaID {
		t.Fatalf("%+v %v", r, err)
	}
	var prov string
	var verificado bool
	if err := s.Pool.QueryRow(ctx, `SELECT i.proveedor, u.email_verificado_en IS NOT NULL FROM auth.identidades i
		JOIN auth.usuarios u ON u.id = i.usuario_id WHERE i.sujeto = 'user_vale'`).Scan(&prov, &verificado); err != nil || prov != "clerk" || !verificado {
		t.Fatalf("%s %v %v", prov, verificado, err)
	}
}

func TestClerkRechazaLoQueNoVerifica(t *testing.T) {
	s, _, _ := nuevo(t)
	if _, err := s.IngresarConClerk(ctx, "x.y.z", cl); !errors.Is(err, cuentas.ErrProveedorApagado) {
		t.Fatalf("sin Clerk: %v", err)
	}
	f := conClerk(t, s)
	f.Usuarios["user_sin"] = clerkfalso.Usuario{Nombre: "Sin", Email: "sin@gmail.com"}
	for nombre, tok := range map[string]string{
		"vencido":      f.Token(t, clerkfalso.Reclamos{Sub: "user_sin", Exp: time.Now().Add(-time.Hour)}),
		"origen ajeno": f.Token(t, clerkfalso.Reclamos{Sub: "user_sin", Azp: "https://otro.example"}),
		"otra firma":   f.TokenAjeno(t, clerkfalso.Reclamos{Sub: "user_sin"}),
		"vacío":        "",
	} {
		if _, err := s.IngresarConClerk(ctx, tok, cl); !errors.Is(err, cuentas.ErrTokenClerk) {
			t.Errorf("%s: %v", nombre, err)
		}
	}
	// Un email sin verificar en Clerk no entra ni crea nada.
	if _, err := s.IngresarConClerk(ctx, f.Token(t, clerkfalso.Reclamos{Sub: "user_sin"}), cl); !errors.Is(err, cuentas.ErrSinEmail) {
		t.Fatalf("sin verificar: %v", err)
	}
}

func TestEnlacePorWhatsApp(t *testing.T) {
	s, _, wa := conProveedores(t)
	ses, err := s.Registrar(ctx, alta("tel@laesquina.com.ar"), cl)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := s.Sesion(ctx, ses.Token)
	if _, err := s.Pool.Exec(ctx, `UPDATE auth.usuarios SET telefono = '5491155550000' WHERE id = $1`, d.Usuario.ID); err != nil {
		t.Fatal(err)
	}
	// Un número sin cuenta no da error ni manda nada.
	if err := s.PedirEnlace(ctx, "11 4444-0000", cl); err != nil || len(wa.Enviados) != 0 {
		t.Fatalf("%v %d", err, len(wa.Enviados))
	}
	if err := s.PedirEnlace(ctx, "11 5555-0000", cl); err != nil {
		t.Fatal(err)
	}
	m, ok := wa.Ultimo("5491155550000")
	if !ok || m.Tipo != "ingreso" || !strings.HasPrefix(m.Sufijo, "m=enlace&token=") || !strings.HasSuffix(m.Link, m.Sufijo) {
		t.Fatalf("%+v", m)
	}
	token := strings.TrimPrefix(m.Sufijo, "m=enlace&token=")
	nueva, err := s.UsarEnlace(ctx, token, cl)
	if err != nil || nueva.TiendaID != ses.TiendaID {
		t.Fatalf("%+v %v", nueva, err)
	}
	if _, err := s.UsarEnlace(ctx, token, cl); !errors.Is(err, cuentas.ErrTokenInvalido) {
		t.Fatalf("segundo uso: %v", err)
	}

	t.Run("recuperar por WhatsApp", func(t *testing.T) {
		if err := s.RecuperarPorWhatsApp(ctx, "+54 9 11 5555 0000", cl); err != nil {
			t.Fatal(err)
		}
		m, _ := wa.Ultimo("5491155550000")
		tok := strings.TrimPrefix(m.Sufijo, "m=restablecer&token=")
		if m.Tipo != "restablecer" || tok == m.Sufijo {
			t.Fatalf("%+v", m)
		}
		if err := s.Restablecer(ctx, tok, "Otraclave123!"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("sin WhatsApp configurado", func(t *testing.T) {
		s.WhatsApp = nil
		if err := s.PedirEnlace(ctx, "11 5555-0000", cl); !errors.Is(err, cuentas.ErrSinWhatsApp) {
			t.Fatalf("err = %v", err)
		}
	})
}
