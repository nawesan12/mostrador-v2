package config

import "testing"

func TestDesarrolloTieneDefectos(t *testing.T) {
	t.Setenv("MOSTRADOR_ENTORNO", "")
	t.Setenv("MOSTRADOR_BASE_DE_DATOS", "")
	c, err := Leer()
	if err != nil || !c.Desarrollo() || c.BaseApp == "" || c.Direccion != ":4000" {
		t.Fatalf("c=%+v err=%v", c, err)
	}
}

func TestProduccionExigeLaBase(t *testing.T) {
	t.Setenv("MOSTRADOR_ENTORNO", "produccion")
	t.Setenv("MOSTRADOR_BASE_DE_DATOS", "")
	c, err := Leer()
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Exigir("BaseApp"); err == nil {
		t.Fatal("en producción sin base tendría que fallar")
	}
}

func TestEntornoInvalido(t *testing.T) {
	t.Setenv("MOSTRADOR_ENTORNO", "prod")
	if _, err := Leer(); err == nil {
		t.Fatal("tendría que rechazar un entorno desconocido")
	}
}
