package tiendanube

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// El caso es compartido con la web (packages/contrato/casos/importar): los mismos datos de la API
// tienen que dar lo mismo en Go y en TS. `go test -run TestCaso -actualizar` regenera lo esperado.
var actualizar = flag.Bool("actualizar", false, "reescribe tiendanube.esperado.json")

const caso = "../../../../../packages/contrato/casos/importar/"

type esperado struct {
	Productos []Importado `json:"productos"`
	Clientes  []Cliente   `json:"clientes"`
	Pedidos   []Pedido    `json:"pedidos"`
}

func TestCaso(t *testing.T) {
	crudo, err := os.ReadFile(filepath.Join(caso, "tiendanube.json"))
	if err != nil {
		t.Fatal(err)
	}
	var entrada struct {
		Productos []ProductoTN `json:"productos"`
		Clientes  []ClienteTN  `json:"clientes"`
		Pedidos   []PedidoTN   `json:"pedidos"`
	}
	if err := json.Unmarshal(crudo, &entrada); err != nil {
		t.Fatal(err)
	}
	r := esperado{Productos: []Importado{}, Clientes: []Cliente{}, Pedidos: []Pedido{}}
	for _, p := range entrada.Productos {
		if i, ok := Producto(p, "es"); ok {
			r.Productos = append(r.Productos, i)
		}
	}
	for _, c := range entrada.Clientes {
		if x, ok := ClienteDe(c); ok {
			r.Clientes = append(r.Clientes, x)
		}
	}
	for _, p := range entrada.Pedidos {
		if x, ok := PedidoDe(p); ok {
			r.Pedidos = append(r.Pedidos, x)
		}
	}
	obtenido, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	obtenido = append(obtenido, '\n')
	ruta := filepath.Join(caso, "tiendanube.esperado.json")
	if *actualizar {
		if err := os.WriteFile(ruta, obtenido, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	quiero, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(normalizar(t, quiero), normalizar(t, obtenido)) {
		t.Fatalf("no coincide con %s:\n%s", ruta, obtenido)
	}
}

// normalizar compara el JSON sin importar el formato (la web lo escribe con prettier).
func normalizar(t *testing.T, b []byte) []byte {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(v)
	return out
}

func TestTextoDeHTML(t *testing.T) {
	for entrada, quiero := range map[string]string{
		"<p>Hola <b>mundo</b></p><p>Chau&nbsp;&amp; listo</p>": "Hola mundo\nChau & listo",
		"Uno<br>Dos<br/><br/><br/><br/>Tres":                   "Uno\nDos\n\nTres",
		"":                                                     "",
	} {
		if r := TextoDeHTML(entrada); r != quiero {
			t.Errorf("TextoDeHTML(%q) = %q, quiero %q", entrada, r, quiero)
		}
	}
}
