package correoargentino

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Estado es lo que ve el panel (components/schemas/ConexionCorreo).
type Estado struct {
	// Disponible: se puede conectar (hay clave de cifrado).
	Disponible bool `json:"disponible"`
	// PideCredenciales: Tu Local Web no tiene credenciales de API, la tienda tiene que traer las suyas.
	PideCredenciales bool       `json:"pideCredenciales"`
	Pruebas          bool       `json:"pruebas"`
	Conectada        bool       `json:"conectada"`
	Email            string     `json:"email,omitempty"`
	Cliente          string     `json:"cliente,omitempty"`
	Propias          bool       `json:"propias"`
	Desde            *time.Time `json:"desde,omitempty"`
	Origen           *Origen    `json:"origen,omitempty"`
	Paquete          *Paquete   `json:"paquete,omitempty"`
}

// EstadoDe la conexión de la tienda (sin descifrar nada).
func (s *Servicio) EstadoDe(ctx context.Context, tiendaID string) (Estado, error) {
	e := Estado{Disponible: s.Activo(), PideCredenciales: s.Plataforma.vacias(), Pruebas: s.cliente().base() == BasePruebas}
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var o Origen
		var p Paquete
		var desde time.Time
		var cred []byte
		err := t.QueryRow(ctx, `SELECT cliente, email, credencial, origen, paquete, conectada_en FROM ca_conexiones WHERE tienda_id = $1`,
			tiendaID).Scan(&e.Cliente, &e.Email, &cred, &o, &p, &desde)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		e.Conectada, e.Propias, e.Desde, e.Origen, e.Paquete = true, cred != nil, &desde, &o, &p
		return nil
	})
	return e, err
}

// Alta: lo que manda el modal de Conectar.
type Alta struct {
	// Cuenta de MiCorreo (la contraseña sólo se usa para /users/validate; no se guarda).
	Email string `json:"email"`
	Clave string `json:"clave"`
	// Credenciales de API propias (opcionales si Tu Local Web tiene las suyas).
	API     *Credenciales `json:"api"`
	Origen  Origen        `json:"origen"`
	Paquete Paquete       `json:"paquete"`
}

// Conectar valida la cuenta contra MiCorreo y guarda la conexión.
func (s *Servicio) Conectar(ctx context.Context, tiendaID string, a Alta) (Estado, error) {
	if !s.Activo() {
		return Estado{}, ErrApagado
	}
	a.Email = strings.TrimSpace(a.Email)
	if !emailValido.MatchString(a.Email) || len(a.Email) > 120 || a.Clave == "" || len(a.Clave) > 200 {
		return Estado{}, ErrInvalido{"Poné el email y la contraseña con los que entrás a MiCorreo."}
	}
	if err := a.Origen.limpiar(); err != nil {
		return Estado{}, err
	}
	if err := a.Paquete.valido(); err != nil {
		return Estado{}, err
	}
	cred := s.Plataforma
	var cifrada []byte
	if a.API != nil && (strings.TrimSpace(a.API.Usuario) != "" || a.API.Clave != "") {
		cred = Credenciales{strings.TrimSpace(a.API.Usuario), a.API.Clave}
		if cred.vacias() || len(cred.Usuario) > 120 || len(cred.Clave) > 200 {
			return Estado{}, ErrInvalido{"Completá el usuario y la contraseña de la API."}
		}
		b, _ := json.Marshal(cred)
		var err error
		if cifrada, err = s.Cifra.Cifrar(contexto(tiendaID), b); err != nil {
			return Estado{}, err
		}
	}
	if cred.vacias() {
		return Estado{}, ErrSinCredenciales
	}
	cliente, err := s.cliente().Validar(ctx, cred, a.Email, a.Clave)
	var e *ErrorCA
	if errors.As(err, &e) {
		switch {
		case e.Estado == http.StatusUnauthorized && cifrada != nil:
			return Estado{}, ErrInvalido{"Correo Argentino no aceptó el usuario y la contraseña de la API."}
		case e.Estado == http.StatusNotFound || e.Estado == http.StatusPaymentRequired || e.Estado == http.StatusBadRequest:
			return Estado{}, ErrCuenta
		}
	}
	if err != nil {
		return Estado{}, err
	}
	if !clienteValid.MatchString(cliente) {
		return Estado{}, errors.New("correoargentino: customerId inesperado")
	}
	err = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `INSERT INTO ca_conexiones (tienda_id, cliente, email, credencial, origen, paquete)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (tienda_id) DO UPDATE SET cliente = $2, email = $3, credencial = $4, origen = $5, paquete = $6,
				conectada_en = now()`, tiendaID, cliente, a.Email, cifrada, a.Origen, a.Paquete)
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "integracion.correo_conectar",
			Entidad: "integracion", EntidadID: "correo", Texto: "conectó Correo Argentino (MiCorreo de " + a.Email + ")"})
	})
	if err != nil {
		return Estado{}, err
	}
	s.olvidarTienda(tiendaID)
	return s.EstadoDe(ctx, tiendaID)
}

// Ajustar cambia el origen y el paquete sin volver a pedir la contraseña.
func (s *Servicio) Ajustar(ctx context.Context, tiendaID string, o Origen, p Paquete) (Estado, error) {
	if err := o.limpiar(); err != nil {
		return Estado{}, err
	}
	if err := p.valido(); err != nil {
		return Estado{}, err
	}
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE ca_conexiones SET origen = $2, paquete = $3 WHERE tienda_id = $1`, tiendaID, o, p)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrSinConexion
		}
		return err
	})
	if err != nil {
		return Estado{}, err
	}
	s.olvidarTienda(tiendaID)
	return s.EstadoDe(ctx, tiendaID)
}

// Desconectar borra la conexión (y las credenciales de API propias, si había). Los envíos ya
// importados quedan en MiCorreo y en ca_envios.
func (s *Servicio) Desconectar(ctx context.Context, tiendaID string) error {
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM ca_conexiones WHERE tienda_id = $1`, tiendaID)
		if err != nil || tag.RowsAffected() == 0 {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "integracion.correo_desconectar",
			Entidad: "integracion", EntidadID: "correo", Texto: "desconectó Correo Argentino"})
	})
	s.olvidarTienda(tiendaID)
	return err
}

func (s *Servicio) descifrar(tiendaID string, dato []byte) (Credenciales, error) {
	b, err := s.Cifra.Descifrar(contexto(tiendaID), dato)
	if err != nil {
		return Credenciales{}, err
	}
	var c Credenciales
	err = json.Unmarshal(b, &c)
	return c, err
}

// olvidarTienda borra lo cotizado de la tienda (cambió el origen, el paquete o la cuenta).
func (s *Servicio) olvidarTienda(tiendaID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.cotizados {
		if strings.HasPrefix(k, tiendaID+"|") {
			delete(s.cotizados, k)
		}
	}
	for k := range s.sucursales {
		if strings.HasPrefix(k, tiendaID+"|") {
			delete(s.sucursales, k)
		}
	}
}
