package importacion

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Importar desde Tiendanube con OAuth: el panel pide adónde mandar a la dueña (Conectar), Tiendanube
// vuelve a la web con el código y la web se lo pasa a Go (VolverTiendanube), que canjea el token y
// crea la importación con la credencial cifrada. El token nunca pasa por el navegador.

var (
	// ErrTiendanubeApagado: el servidor no tiene las credenciales de la app de socio.
	ErrTiendanubeApagado = errors.New("importacion: Tiendanube no está disponible en este servidor")
	// ErrEstado: el state no existe, venció o ya se usó.
	ErrEstado = errors.New("importacion: la conexión venció")
	// ErrCanje: Tiendanube no cambió el código por un token.
	ErrCanje = errors.New("importacion: Tiendanube no aceptó la conexión")
)

// TiendanubeDisponible: hay credenciales de la app y con qué cifrar el token.
func (s *Servicio) TiendanubeDisponible() bool {
	return s.Tiendanube.ID != "" && s.Tiendanube.Secreto != "" && s.Cifra != nil
}

// ConectarTiendanube guarda un state de un solo uso (10 minutos, sólo su hash) y devuelve la URL de
// autorización de la app. La vuelta la registra la app en Tiendanube, no va en la URL.
func (s *Servicio) ConectarTiendanube(ctx context.Context, tienda, rol, usuario string) (string, error) {
	if !esAdmin(rol) {
		return "", ErrSinPermiso
	}
	if !s.TiendanubeDisponible() {
		return "", ErrTiendanubeApagado
	}
	estado := azar()
	if err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		// Los vencidos de la tienda se van con cada conexión nueva (no hace falta un cron).
		if _, err := t.Exec(ctx, `DELETE FROM importacion_estados WHERE tienda_id = $1 AND vence_en < now()`, tienda); err != nil {
			return err
		}
		_, err := t.Exec(ctx, `INSERT INTO importacion_estados (estado_hash, tienda_id, usuario_id, plataforma, vence_en)
			VALUES ($1, $2, $3, 'tiendanube', now() + interval '10 minutes')`, hashEstado(estado), tienda, usuario)
		return err
	}); err != nil {
		return "", err
	}
	return s.Tiendanube.URLAutorizar(estado)
}

// VolverTiendanube consume el state (de la misma tienda y la misma persona), canjea el código y crea
// la importación con la credencial. Un state usado o vencido no canjea nada.
func (s *Servicio) VolverTiendanube(ctx context.Context, tienda, rol, usuario, codigo, estado string) (Importacion, error) {
	if !esAdmin(rol) {
		return Importacion{}, ErrSinPermiso
	}
	if !s.TiendanubeDisponible() {
		return Importacion{}, ErrTiendanubeApagado
	}
	if codigo == "" || estado == "" || len(codigo) > 512 || len(estado) > 512 {
		return Importacion{}, ErrEstado
	}
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE importacion_estados SET usado_en = now()
			WHERE estado_hash = $1 AND tienda_id = $2 AND usuario_id = $3 AND plataforma = 'tiendanube'
				AND usado_en IS NULL AND vence_en > now()`, hashEstado(estado), tienda, usuario)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrEstado
		}
		return err
	})
	if err != nil {
		return Importacion{}, err
	}
	tk, err := s.Tiendanube.Canjear(ctx, codigo)
	if err != nil {
		return Importacion{}, fmt.Errorf("%w: %v", ErrCanje, err)
	}
	return s.Crear(ctx, tienda, rol, Nueva{Plataforma: "tiendanube", Credencial: string(CredencialTiendanube(tk))})
}

func azar() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashEstado(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
