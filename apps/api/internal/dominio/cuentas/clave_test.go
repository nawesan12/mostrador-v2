package cuentas

import (
	"strings"
	"testing"
)

func TestHashYVerificacion(t *testing.T) {
	h, err := HashClave("Abcdefghijk1")
	if err != nil || !strings.HasPrefix(h, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("hash %q err %v", h, err)
	}
	if ok, _ := VerificarClave("Abcdefghijk1", h); !ok {
		t.Fatal("la clave correcta no verificó")
	}
	if ok, _ := VerificarClave("abcdefghijk1", h); ok {
		t.Fatal("una clave distinta verificó")
	}
	otro, _ := HashClave("Abcdefghijk1")
	if otro == h {
		t.Fatal("dos hashes de la misma clave tienen que diferir (sal)")
	}
	if _, err := VerificarClave("x", "$bcrypt$..."); err == nil {
		t.Fatal("un formato desconocido tendría que dar error")
	}
}
