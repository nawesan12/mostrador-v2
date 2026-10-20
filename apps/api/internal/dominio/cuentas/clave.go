package cuentas

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Parámetros de argon2id recomendados por OWASP (19 MiB, 2 pasadas, 1 hilo): unos 30 ms en el VPS.
const (
	argonMemoria = 19 * 1024
	argonTiempo  = 2
	argonHilos   = 1
	argonLargo   = 32
)

// HashClave devuelve el hash en formato PHC: $argon2id$v=19$m=…,t=…,p=…$sal$hash.
func HashClave(clave string) (string, error) {
	sal := make([]byte, 16)
	if _, err := rand.Read(sal); err != nil {
		return "", err
	}
	h := argon2.IDKey([]byte(clave), sal, argonTiempo, argonMemoria, argonHilos, argonLargo)
	b := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argonMemoria, argonTiempo, argonHilos,
		b.EncodeToString(sal), b.EncodeToString(h)), nil
}

// VerificarClave compara en tiempo constante. Lee los parámetros del propio hash, así subirlos
// después no invalida las contraseñas viejas.
func VerificarClave(clave, hash string) (bool, error) {
	p := strings.Split(hash, "$")
	if len(p) != 6 || p[1] != "argon2id" {
		return false, errors.New("hash con formato desconocido")
	}
	var m, t uint32
	var hilos uint8
	if _, err := fmt.Sscanf(p[3], "m=%d,t=%d,p=%d", &m, &t, &hilos); err != nil {
		return false, err
	}
	b := base64.RawStdEncoding
	sal, err := b.DecodeString(p[4])
	if err != nil {
		return false, err
	}
	esperado, err := b.DecodeString(p[5])
	if err != nil {
		return false, err
	}
	h := argon2.IDKey([]byte(clave), sal, t, m, hilos, uint32(len(esperado)))
	return subtle.ConstantTimeCompare(h, esperado) == 1, nil
}

// hashFalso sirve para gastar el mismo tiempo cuando el email no existe: así la respuesta no
// delata quién tiene cuenta.
var hashFalso, _ = HashClave("una clave que nadie tiene")
