package mercadopago_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// La primera conexión no pide volver a verificarse; cambiar la cuenta sí, también después de
// «Desconectar» (si no, desconectar y conectar otra salteaba la verificación).
func TestConectarOtraCuentaPideVerificarse(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	var ana string
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana@luna.com', 'Ana') RETURNING id::text`).Scan(&ana)
	}); err != nil {
		t.Fatal(err)
	}
	f := &mpFalso{cuerpos: map[string]map[string]any{}, pagos: map[string]map[string]any{}, susc: map[string]map[string]any{}, usuario: 779, refreshOK: true}
	srv := httptest.NewServer(f)
	defer srv.Close()
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	cif, _ := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	s := &mercadopago.Servicio{Pool: b.App, Cifra: cif, Cliente: &mercadopago.Cliente{Base: srv.URL}, Ahora: time.Now,
		Config: mercadopago.Config{AppID: "123", Secreto: "sec", SecretoWebhook: "whsec", Sitio: "https://tulocalweb.com.ar"}}
	d := cuentas.Datos{TiendaID: luna, SesionHash: "sesion-sin-permiso"}
	pasa := func() bool {
		t.Helper()
		ok, err := mercadopago.ReverificadoSiConectada(s, ctx, d, "")
		if err != nil {
			t.Fatal(err)
		}
		return ok
	}
	if !pasa() {
		t.Fatal("la primera conexión pidió verificarse")
	}
	u, err := s.Conectar(ctx, luna, ana)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.Parse(u)
	if err := s.Volver(ctx, luna, ana, "codigo", q.Query().Get("state")); err != nil {
		t.Fatal(err)
	}
	if pasa() {
		t.Fatal("cambiar la cuenta conectada no pidió verificarse")
	}
	if err := s.Desconectar(ctx, luna); err != nil {
		t.Fatal(err)
	}
	if pasa() {
		t.Fatal("desconectar y conectar otra salteó la verificación")
	}
}
