package configuracion_test

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cobros"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Cargar el CBU por primera vez no pide nada; cambiarlo pide volver a verificarse (con la
// contraseña o con el código del mail, para las cuentas sin contraseña). Cada cambio avisa a la
// dueña, y el link del aviso lo vuelve atrás y cierra todas las sesiones.
func TestCambiarCBUPideVerificacion(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	mails := &correo.Memoria{}
	s := &cuentas.Servicio{Pool: b.App, Correo: mails, Sitio: sitio}
	cl := cuentas.Cliente{IP: "190.210.32.1", Agente: "Mozilla/5.0 (Macintosh) Chrome/130"}
	ana, err := s.Registrar(ctx, cuentas.Alta{Nombre: "Ana", Tienda: "Sol", Email: "ana@sol.com", Clave: "clave-larga-1"}, cl)
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.Sesion(ctx, ana.Token)
	if err != nil {
		t.Fatal(err)
	}
	v := configuracion.Verificacion{Usuario: d.Usuario.ID, Sesion: d.SesionHash}
	n := configuracion.Negocio{Nombre: "Sol", Alias: "sol.deco.mp", CBU: "0170099220000067797370", Titular: "Ana Sol"}
	if err := configuracion.GuardarNegocioDe(ctx, b.App, ana.TiendaID, n, v); err != nil {
		t.Fatalf("la primera carga pidió verificación: %v", err)
	}
	n.Alias = "otro.alias.mp"
	if err := configuracion.GuardarNegocioDe(ctx, b.App, ana.TiendaID, n, v); !errors.Is(err, configuracion.ErrReverificar) {
		t.Fatalf("cambiar el alias sin verificarse: %v", err)
	}
	if _, _, err := s.Reverificar(ctx, d, "mal", ""); !errors.Is(err, cuentas.ErrCredenciales) {
		t.Fatalf("clave mal: %v", err)
	}
	// Con el código del mail (el camino de quien entra con Google o WhatsApp).
	if err := s.PedirCodigoReverificar(ctx, d); err != nil {
		t.Fatal(err)
	}
	codigo := regexp.MustCompile(`\d{6}`).FindString(mails.Enviados[len(mails.Enviados)-1].Asunto)
	tok, _, err := s.Reverificar(ctx, d, "", codigo)
	if err != nil {
		t.Fatalf("código del mail: %v", err)
	}
	v.Token = tok
	if err := configuracion.GuardarNegocioDe(ctx, b.App, ana.TiendaID, n, v); err != nil {
		t.Fatalf("con el permiso: %v", err)
	}
	// El permiso es de esa sesión: otra no lo puede usar.
	otra := v
	otra.Sesion = "otra"
	n.Alias = "tercero.mp"
	if err := configuracion.GuardarNegocioDe(ctx, b.App, ana.TiendaID, n, otra); !errors.Is(err, configuracion.ErrReverificar) {
		t.Fatalf("el permiso sirvió en otra sesión: %v", err)
	}
	// Dos avisos (la carga y el cambio), con el link para volver atrás.
	filas, err := b.Worker.Query(ctx, `SELECT datos FROM outbox WHERE tipo = $1 ORDER BY id`, cobros.TipoCambio)
	if err != nil {
		t.Fatal(err)
	}
	var avisos []cobros.AvisoCambio
	for filas.Next() {
		var raw []byte
		_ = filas.Scan(&raw)
		var a cobros.AvisoCambio
		_ = json.Unmarshal(raw, &a)
		avisos = append(avisos, a)
	}
	filas.Close()
	if len(avisos) != 2 || avisos[1].Antes["alias"] != "sol.deco.mp" || avisos[1].Despues["alias"] != "otro.alias.mp" || avisos[1].Token == "" {
		t.Fatalf("avisos: %+v", avisos)
	}
	if avisos[0].Despues["cbu"] == n.CBU {
		t.Fatal("el aviso trae el CBU entero")
	}
	// «No fui yo»: vuelve el alias de antes y se cierran todas las sesiones.
	if _, err := configuracion.RevertirCambio(ctx, b.App, avisos[1].Token); err != nil {
		t.Fatal(err)
	}
	var alias string
	_ = b.Staff.QueryRow(ctx, `SELECT negocio->>'alias' FROM tiendas WHERE id = $1`, ana.TiendaID).Scan(&alias)
	if alias != "sol.deco.mp" {
		t.Fatalf("no volvió el alias: %q", alias)
	}
	if _, err := s.Sesion(ctx, ana.Token); !errors.Is(err, cuentas.ErrSinSesion) {
		t.Fatalf("la sesión sigue abierta: %v", err)
	}
	if _, err := configuracion.RevertirCambio(ctx, b.App, avisos[1].Token); !errors.Is(err, cobros.ErrCambioVencido) {
		t.Fatalf("el link sirvió dos veces: %v", err)
	}
}
