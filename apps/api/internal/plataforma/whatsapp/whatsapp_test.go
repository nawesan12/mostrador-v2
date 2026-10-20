package whatsapp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/whatsapp"
)

func TestNormalizar(t *testing.T) {
	casos := map[string]string{
		"11 5555-0000":       "5491155550000",
		"+54 9 11 5555 0000": "5491155550000",
		"54 11 5555 0000":    "5491155550000",
		"0351 15 555 0000":   "5493515550000",
		"351 555-0000":       "5493515550000",
		"123":                "",
	}
	for in, want := range casos {
		if got := whatsapp.Normalizar(in); got != want {
			t.Errorf("%q: %q, quería %q", in, got, want)
		}
	}
}

func TestCloudMandaLaPlantillaConElLink(t *testing.T) {
	var recibido map[string]any
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		if r.URL.Path != "/123/messages" {
			t.Errorf("ruta %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&recibido)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	c := &whatsapp.Cloud{Token: "tk", NumeroID: "123", Base: srv.URL, Plantillas: map[string]string{"ingreso": "mostrador_ingreso"}}
	if err := c.Enviar(context.Background(), whatsapp.Mensaje{Para: "5491155550000", Tipo: "ingreso", Sufijo: "m=enlace&token=abc"}); err != nil {
		t.Fatal(err)
	}
	tpl := recibido["template"].(map[string]any)
	if auth != "Bearer tk" || recibido["to"] != "5491155550000" || tpl["name"] != "mostrador_ingreso" {
		t.Fatalf("%v %v", auth, recibido)
	}
	if err := c.Enviar(context.Background(), whatsapp.Mensaje{Tipo: "otro"}); err == nil {
		t.Fatal("sin plantilla tendría que fallar")
	}
}
