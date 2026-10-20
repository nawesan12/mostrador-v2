package correo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResendMandaLoQueCorresponde(t *testing.T) {
	var auth string
	var cuerpo map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &cuerpo)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	r := &Resend{Clave: "re_x", De: "Mostrador <hola@mostrador.iamnaw.dev>", URL: srv.URL}
	if err := r.Enviar(context.Background(), Mensaje{Para: "a@b.co", Asunto: "Hola", HTML: "<b>h</b>", Texto: "h"}); err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer re_x" || cuerpo["subject"] != "Hola" || cuerpo["to"].([]any)[0] != "a@b.co" {
		t.Fatalf("auth=%q cuerpo=%v", auth, cuerpo)
	}
}

func TestResendDevuelveElError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		_, _ = w.Write([]byte(`{"message":"dominio sin verificar"}`))
	}))
	defer srv.Close()
	err := (&Resend{Clave: "x", URL: srv.URL}).Enviar(context.Background(), Mensaje{Para: "a@b.co"})
	if err == nil || !strings.Contains(err.Error(), "dominio sin verificar") {
		t.Fatalf("err = %v", err)
	}
}

func TestDireccionDe(t *testing.T) {
	if got := direccionDe("Mostrador <hola@x.dev>"); got != "hola@x.dev" {
		t.Fatal(got)
	}
}
