// Package partners es el programa de partners: la ficha del partner (nivel, link, cupón, perfil
// público y datos de cobro), las tiendas que trajo (por el ref o el cupón del origen de la tienda),
// la comisión de cada una y los pagos por mes. Todo es de una persona, no de una tienda: se accede
// con db.EnUsuario y RLS compara contra la persona de la transacción.
package partners

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
)

var (
	// ErrNoEsPartner: la persona no tiene ficha de partner (no se postuló o no la aprobaron).
	ErrNoEsPartner = errors.New("partners: no es partner")
	// ErrSinCuenta: no hay una cuenta con ese email para aprobar.
	ErrSinCuenta = errors.New("partners: sin cuenta")
	// ErrSinPago: no hay liquidación abierta de ese mes.
	ErrSinPago = errors.New("partners: sin pago")
)

// ErrInvalido es un dato que no sirve, con el campo para marcarlo en el formulario.
type ErrInvalido struct{ Campo, Mensaje string }

func (e ErrInvalido) Error() string { return e.Campo + ": " + e.Mensaje }

// Caso es un trabajo del partner que se muestra en el perfil público.
type Caso struct {
	Nombre string `json:"nombre"`
	URL    string `json:"url"`
}

// Perfil es lo que se ve en "Buscar un partner".
type Perfil struct {
	Nombre         string   `json:"nombre"`
	Ciudad         string   `json:"ciudad"`
	Descripcion    string   `json:"descripcion"`
	Especialidades []string `json:"especialidades"`
	Casos          []Caso   `json:"casos"`
	Email          string   `json:"email"`
	Whatsapp       string   `json:"whatsapp"`
	Web            string   `json:"web"`
	Logo           string   `json:"logo"`
}

// Cobro son los datos para pagarle al partner.
type Cobro struct {
	// cbu (CBU o CVU) o alias (alias de Mercado Pago o del banco).
	Tipo         string `json:"tipo"`
	Valor        string `json:"valor"`
	Titular      string `json:"titular"`
	Cuit         string `json:"cuit"`
	CondicionIva string `json:"condicionIva"`
}

// Ficha del partner.
type Ficha struct {
	Nombre        string `json:"nombre"`
	Email         string `json:"email"`
	Nivel         string `json:"nivel"`
	NivelEfectivo string `json:"nivelEfectivo"`
	Porcentaje    int64  `json:"porcentaje"`
	Codigo        string `json:"codigo"`
	Cupon         string `json:"cupon"`
	AprobadoEn    string `json:"aprobadoEn"`
	Perfil        Perfil `json:"perfil"`
	Cobro         *Cobro `json:"cobro"`
}

// Referida es una tienda que trajo el partner.
type Referida struct {
	ID          string `json:"id"`
	Nombre      string `json:"nombre"`
	Slug        string `json:"slug"`
	Plan        string `json:"plan"`
	Ciclo       string `json:"ciclo"`
	Estado      string `json:"estado"`
	Alta        string `json:"alta"`
	FinComision string `json:"finComision"`
	// Lo que genera este mes (0 si no está activa o pasaron los 12 meses).
	ComisionMes int64 `json:"comisionMes"`
}

// Linea de una liquidación.
type Linea struct {
	TiendaID string `json:"tiendaId"`
	Tienda   string `json:"tienda"`
	Plan     string `json:"plan"`
	Comision int64  `json:"comision"`
}

// Pago es la liquidación de un mes.
type Pago struct {
	Mes        string  `json:"mes"`
	Estado     string  `json:"estado"`
	Monto      int64   `json:"monto"`
	Detalle    []Linea `json:"detalle"`
	FacturaURL *string `json:"facturaUrl"`
	FechaPago  string  `json:"fechaPago"`
	PagadoEn   *string `json:"pagadoEn"`
}

// Clic es el contador de un mes.
type Clic struct {
	Mes string `json:"mes"`
	N   int    `json:"n"`
}

// Colaboracion es una tienda donde la dueña sumó al partner a su equipo.
type Colaboracion struct {
	TiendaID string `json:"tiendaId"`
	Nombre   string `json:"nombre"`
	Slug     string `json:"slug"`
	Rol      string `json:"rol"`
}

// Panel es todo lo que muestra /partners/panel.
type Panel struct {
	Hoy            string         `json:"hoy"`
	Partner        Ficha          `json:"partner"`
	Tiendas        []Referida     `json:"tiendas"`
	Clics          []Clic         `json:"clics"`
	Pagos          []Pago         `json:"pagos"`
	Colaboraciones []Colaboracion `json:"colaboraciones"`
}

type fila struct {
	id, nombre, slug, plan, ciclo string
	creada                        time.Time
	baja                          bool
}

// referidas lee las tiendas del partner de la transacción.
func referidas(ctx context.Context, tx pgx.Tx) ([]fila, error) {
	rows, err := tx.Query(ctx, `SELECT id::text, nombre, slug, plan, plan_ciclo, creada_en,
		cancelada_en IS NOT NULL OR cerrada_en IS NOT NULL FROM partners.tiendas_referidas()`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (fila, error) {
		var f fila
		err := r.Scan(&f.id, &f.nombre, &f.slug, &f.plan, &f.ciclo, &f.creada, &f.baja)
		return f, err
	})
}

// calcular arma las tiendas con su estado, el nivel efectivo y la comisión de `mes`.
func calcular(asignado string, fs []fila, ahora, mes time.Time) (string, []Referida) {
	activas := 0
	estados := make([]string, len(fs))
	for i, f := range fs {
		estados[i] = Estado(f.plan, f.creada, f.baja, ahora)
		if estados[i] == EstadoActiva {
			activas++
		}
	}
	nivel := NivelEfectivo(asignado, activas)
	out := make([]Referida, len(fs))
	for i, f := range fs {
		var c int64
		if estados[i] == EstadoActiva {
			c = Comision(nivel, f.plan, f.ciclo, f.creada, mes)
		}
		out[i] = Referida{ID: f.id, Nombre: f.nombre, Slug: f.slug, Plan: f.plan, Ciclo: f.ciclo, Estado: estados[i],
			Alta: f.creada.In(zona).Format(time.DateOnly), FinComision: FinComision(f.creada).Format(time.DateOnly), ComisionMes: c}
	}
	return nivel, out
}

func primeroDeMes(t time.Time) time.Time {
	a := t.In(zona)
	return time.Date(a.Year(), a.Month(), 1, 0, 0, 0, 0, zona)
}

// Leer arma el panel de la persona. ErrNoEsPartner si no tiene ficha.
func Leer(ctx context.Context, pool *db.Pool, usuarioID string, ahora time.Time) (Panel, error) {
	var p Panel
	err := pool.EnUsuario(ctx, usuarioID, func(tx pgx.Tx) error {
		var perfil, cobro []byte
		var aprobado time.Time
		err := tx.QueryRow(ctx, `SELECT u.nombre, u.email, p.nivel, p.codigo, p.cupon, p.perfil, p.cobro, p.aprobado_en
			FROM partners.partners p JOIN auth.usuarios u ON u.id = p.usuario_id`).
			Scan(&p.Partner.Nombre, &p.Partner.Email, &p.Partner.Nivel, &p.Partner.Codigo, &p.Partner.Cupon, &perfil, &cobro, &aprobado)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoEsPartner
		}
		if err != nil {
			return err
		}
		p.Partner.AprobadoEn = aprobado.In(zona).Format(time.DateOnly)
		if err := json.Unmarshal(perfil, &p.Partner.Perfil); err != nil {
			return err
		}
		if cobro != nil {
			p.Partner.Cobro = &Cobro{}
			if err := json.Unmarshal(cobro, p.Partner.Cobro); err != nil {
				return err
			}
		}
		fs, err := referidas(ctx, tx)
		if err != nil {
			return err
		}
		p.Partner.NivelEfectivo, p.Tiendas = calcular(p.Partner.Nivel, fs, ahora, ahora)
		p.Partner.Porcentaje = Porcentaje(p.Partner.NivelEfectivo)

		rows, err := tx.Query(ctx, `SELECT to_char(mes, 'YYYY-MM'), n FROM partners.clics ORDER BY mes DESC LIMIT 12`)
		if err != nil {
			return err
		}
		if p.Clics, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (Clic, error) {
			var c Clic
			return c, r.Scan(&c.Mes, &c.N)
		}); err != nil {
			return err
		}

		rows, err = tx.Query(ctx, `SELECT mes, estado, monto, detalle, factura_url, pagado_en
			FROM partners.pagos ORDER BY mes DESC LIMIT 24`)
		if err != nil {
			return err
		}
		if p.Pagos, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (Pago, error) {
			var g Pago
			var mes time.Time
			var detalle []byte
			var pagado *time.Time
			if err := r.Scan(&mes, &g.Estado, &g.Monto, &detalle, &g.FacturaURL, &pagado); err != nil {
				return g, err
			}
			m := time.Date(mes.Year(), mes.Month(), 1, 0, 0, 0, 0, zona)
			g.Mes = m.Format("2006-01")
			g.FechaPago = FechaDePago(m).Format(time.DateOnly)
			if pagado != nil {
				s := pagado.In(zona).Format(time.DateOnly)
				g.PagadoEn = &s
			}
			return g, json.Unmarshal(detalle, &g.Detalle)
		}); err != nil {
			return err
		}

		rows, err = tx.Query(ctx, `SELECT id::text, nombre, slug, rol FROM partners.colaboraciones()`)
		if err != nil {
			return err
		}
		p.Colaboraciones, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (Colaboracion, error) {
			var c Colaboracion
			return c, r.Scan(&c.TiendaID, &c.Nombre, &c.Slug, &c.Rol)
		})
		return err
	})
	p.Hoy = ahora.In(zona).Format(time.DateOnly)
	if p.Tiendas == nil {
		p.Tiendas = []Referida{}
	}
	if p.Partner.Perfil.Especialidades == nil {
		p.Partner.Perfil.Especialidades = []string{}
	}
	if p.Partner.Perfil.Casos == nil {
		p.Partner.Perfil.Casos = []Caso{}
	}
	return p, err
}

// ---- Perfil y cobro ----

// Especialidades del programa (las mismas que la postulación).
var Especialidades = []string{"diseno", "desarrollo", "marketing", "contabilidad", "fotografia"}

func limpiar(v string) string { return strings.Join(strings.Fields(v), " ") }

func https(campo, v string, max int) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", nil
	}
	u, err := url.Parse(v)
	if err != nil || u.Scheme != "https" || u.Host == "" || len(v) > max {
		return "", ErrInvalido{campo, "Poné un link que empiece con https://."}
	}
	return v, nil
}

var soloDigitos = regexp.MustCompile(`\D`)

// validarPerfil limpia y valida el perfil público.
func validarPerfil(p Perfil) (Perfil, error) {
	var err error
	p.Nombre = limpiar(p.Nombre)
	if p.Nombre == "" || len([]rune(p.Nombre)) > 80 {
		return p, ErrInvalido{"nombre", "Poné el nombre con el que te conocen (hasta 80 letras)."}
	}
	p.Ciudad = limpiar(p.Ciudad)
	if len([]rune(p.Ciudad)) > 60 {
		return p, ErrInvalido{"ciudad", "La ciudad es muy larga."}
	}
	p.Descripcion = limpiar(p.Descripcion)
	if len([]rune(p.Descripcion)) > 280 {
		return p, ErrInvalido{"descripcion", "Contalo en 280 letras o menos."}
	}
	esp := []string{}
	for _, e := range p.Especialidades {
		if !slices.Contains(Especialidades, e) {
			return p, ErrInvalido{"especialidades", "Elegí especialidades de la lista."}
		}
		if !slices.Contains(esp, e) {
			esp = append(esp, e)
		}
	}
	p.Especialidades = esp
	if len(p.Casos) > 6 {
		return p, ErrInvalido{"casos", "Hasta 6 casos."}
	}
	casos := []Caso{}
	for _, c := range p.Casos {
		c.Nombre = limpiar(c.Nombre)
		if c.Nombre == "" && strings.TrimSpace(c.URL) == "" {
			continue
		}
		if c.Nombre == "" || len([]rune(c.Nombre)) > 60 {
			return p, ErrInvalido{"casos", "Cada caso lleva el nombre de la tienda (hasta 60 letras)."}
		}
		if c.URL, err = https("casos", c.URL, 200); err != nil || c.URL == "" {
			return p, ErrInvalido{"casos", "Cada caso lleva el link de la tienda, con https://."}
		}
		casos = append(casos, c)
	}
	p.Casos = casos
	p.Email = strings.TrimSpace(p.Email)
	if p.Email != "" {
		if a, err := mail.ParseAddress(p.Email); err != nil || a.Address != p.Email || len(p.Email) > 254 {
			return p, ErrInvalido{"email", "Revisá el email."}
		}
	}
	p.Whatsapp = soloDigitos.ReplaceAllString(p.Whatsapp, "")
	if p.Whatsapp != "" && (len(p.Whatsapp) < 8 || len(p.Whatsapp) > 15) {
		return p, ErrInvalido{"whatsapp", "El WhatsApp va con la característica: 11 5555 1234."}
	}
	if p.Web, err = https("web", p.Web, 200); err != nil {
		return p, err
	}
	if p.Logo, err = https("logo", p.Logo, 300); err != nil {
		return p, err
	}
	return p, nil
}

var aliasValido = regexp.MustCompile(`^[a-zA-Z0-9.-]{6,20}$`)

// CuitValido comprueba el dígito verificador (módulo 11) de un CUIT o CUIL de 11 dígitos.
func CuitValido(c string) bool {
	if len(c) != 11 || soloDigitos.MatchString(c) {
		return false
	}
	pesos := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	s := 0
	for i, p := range pesos {
		s += int(c[i]-'0') * p
	}
	v := 11 - s%11
	switch v {
	case 11:
		v = 0
	case 10:
		v = 9
	}
	return v == int(c[10]-'0')
}

// CondicionesIva frente al IVA.
var CondicionesIva = []string{"monotributo", "responsable_inscripto", "exento"}

func validarCobro(c Cobro) (Cobro, error) {
	switch c.Tipo {
	case "cbu":
		c.Valor = soloDigitos.ReplaceAllString(c.Valor, "")
		if !configuracion.CBUValido(c.Valor) {
			return c, ErrInvalido{"valor", "Revisá el CBU o CVU: son 22 números."}
		}
	case "alias":
		c.Valor = strings.TrimSpace(c.Valor)
		if !aliasValido.MatchString(c.Valor) {
			return c, ErrInvalido{"valor", "El alias tiene entre 6 y 20 letras, números, puntos o guiones."}
		}
	default:
		return c, ErrInvalido{"tipo", "Elegí CBU/CVU o alias."}
	}
	c.Titular = limpiar(c.Titular)
	if c.Titular == "" || len([]rune(c.Titular)) > 80 {
		return c, ErrInvalido{"titular", "Poné el titular de la cuenta."}
	}
	c.Cuit = soloDigitos.ReplaceAllString(c.Cuit, "")
	if !CuitValido(c.Cuit) {
		return c, ErrInvalido{"cuit", "Revisá el CUIT: son 11 números."}
	}
	if !slices.Contains(CondicionesIva, c.CondicionIva) {
		return c, ErrInvalido{"condicionIva", "Elegí tu condición frente al IVA."}
	}
	return c, nil
}

func actualizar(ctx context.Context, pool *db.Pool, usuarioID, columna string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return pool.EnUsuario(ctx, usuarioID, func(tx pgx.Tx) error {
		// La columna sale de este paquete (perfil o cobro), nunca del pedido.
		tag, err := tx.Exec(ctx, `UPDATE partners.partners SET `+columna+` = $1`, b)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoEsPartner
		}
		return err
	})
}

// GuardarPerfil valida y guarda el perfil público.
func GuardarPerfil(ctx context.Context, pool *db.Pool, usuarioID string, p Perfil) (Perfil, error) {
	p, err := validarPerfil(p)
	if err != nil {
		return p, err
	}
	return p, actualizar(ctx, pool, usuarioID, "perfil", p)
}

// GuardarCobro valida y guarda los datos de cobro.
func GuardarCobro(ctx context.Context, pool *db.Pool, usuarioID string, c Cobro) (Cobro, error) {
	c, err := validarCobro(c)
	if err != nil {
		return c, err
	}
	return c, actualizar(ctx, pool, usuarioID, "cobro", c)
}

var mesValido = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

func parsearMes(mes string) (time.Time, error) {
	if !mesValido.MatchString(mes) {
		return time.Time{}, ErrInvalido{"mes", "El mes va como 2026-10."}
	}
	return time.ParseInLocation("2006-01", mes, zona)
}

// SubirFactura guarda el link de la factura del partner para un pago (una vez pagado, no cambia).
func SubirFactura(ctx context.Context, pool *db.Pool, usuarioID, mes, link string) error {
	m, err := parsearMes(mes)
	if err != nil {
		return err
	}
	link, err = https("url", link, 500)
	if err != nil {
		return err
	}
	if link == "" {
		return ErrInvalido{"url", "Pegá el link de la factura (PDF en Drive, Dropbox o ARCA)."}
	}
	return pool.EnUsuario(ctx, usuarioID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE partners.pagos SET factura_url = $2, factura_en = now()
			WHERE mes = $1 AND estado <> 'pagada'`, m, link)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrSinPago
		}
		return err
	})
}

var codigoValido = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$`)

// SumarClic cuenta una visita al link del partner (sin sesión). Un ref que no es de nadie no hace
// nada.
func SumarClic(ctx context.Context, pool *db.Pool, ref string) (bool, error) {
	ref = strings.ToLower(strings.TrimSpace(ref))
	if !codigoValido.MatchString(ref) {
		return false, nil
	}
	var ok bool
	err := pool.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT partners.sumar_clic($1)`, ref).Scan(&ok)
	})
	return ok, err
}

// ---- Uso interno (backoffice, con la clave del BFF) ----

// codigoDe arma el ref a partir de un nombre, como la dirección de una tienda: "Pampa Code" →
// "pampa-code".
func codigoDe(nombre string) string {
	c := tiendas.Slug(nombre)
	if len(c) > 30 {
		c = strings.TrimRight(c[:30], "-")
	}
	for len(c) < 3 {
		c += "0"
	}
	return c
}

// Aprobacion es el pedido del equipo para aprobar a alguien.
type Aprobacion struct {
	Email  string `json:"email"`
	Nivel  string `json:"nivel"`
	Nombre string `json:"nombre"`
}

// Aprobar convierte la cuenta del email en partner (o le cambia el nivel si ya lo era). El link y el
// cupón salen del nombre (de la agencia, si viene) y no cambian después. Devuelve el código.
func Aprobar(ctx context.Context, pool *db.Pool, a Aprobacion) (string, error) {
	if a.Nivel == "" {
		a.Nivel = "partner"
	}
	if !slices.Contains(Niveles, a.Nivel) {
		return "", ErrInvalido{"nivel", "El nivel es partner, plata u oro."}
	}
	var usuario, nombre string
	err := pool.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT id::text, nombre FROM auth.usuarios WHERE email = $1`, strings.TrimSpace(a.Email)).Scan(&usuario, &nombre)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrSinCuenta
	}
	if err != nil {
		return "", err
	}
	if limpiar(a.Nombre) != "" {
		nombre = limpiar(a.Nombre)
	}
	base := codigoDe(nombre)
	perfil, _ := json.Marshal(Perfil{Nombre: nombre, Especialidades: []string{}, Casos: []Caso{}})
	// Si el código ya es de otro partner (RLS no deja verlo: avisa el índice único), se le suma un
	// número.
	for i := 1; i <= 20; i++ {
		codigo := base
		if i > 1 {
			codigo = fmt.Sprintf("%s-%d", base, i)
		}
		cupon := strings.ToUpper(strings.ReplaceAll(codigo, "-", ""))
		if len(cupon) > 16 {
			cupon = cupon[:16]
		}
		cupon += "20"
		var final string
		err = pool.EnUsuario(ctx, usuario, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `INSERT INTO partners.partners (usuario_id, nivel, codigo, cupon, perfil)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (usuario_id) DO UPDATE SET nivel = excluded.nivel
				RETURNING codigo`, usuario, a.Nivel, codigo, cupon, perfil).Scan(&final)
		})
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			continue
		}
		return final, err
	}
	return "", errors.New("partners: no hay código libre")
}

func usuarioDe(ctx context.Context, pool *db.Pool, email string) (string, error) {
	var id string
	err := pool.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT id::text FROM auth.usuarios WHERE email = $1`, strings.TrimSpace(email)).Scan(&id)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrSinCuenta
	}
	return id, err
}

// Liquidar arma (o rehace, si sigue pendiente) la liquidación de un mes: la comisión de cada
// tienda activa con el nivel de hoy, congelada en el detalle.
func Liquidar(ctx context.Context, pool *db.Pool, email, mes string, ahora time.Time) (Pago, error) {
	m, err := parsearMes(mes)
	if err != nil {
		return Pago{}, err
	}
	usuario, err := usuarioDe(ctx, pool, email)
	if err != nil {
		return Pago{}, err
	}
	g := Pago{Mes: mes, Estado: "pendiente", Detalle: []Linea{}, FechaPago: FechaDePago(m).Format(time.DateOnly)}
	err = pool.EnUsuario(ctx, usuario, func(tx pgx.Tx) error {
		var asignado, id string
		err := tx.QueryRow(ctx, `SELECT id::text, nivel FROM partners.partners`).Scan(&id, &asignado)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoEsPartner
		}
		if err != nil {
			return err
		}
		fs, err := referidas(ctx, tx)
		if err != nil {
			return err
		}
		_, ts := calcular(asignado, fs, ahora, m)
		for _, t := range ts {
			if t.ComisionMes > 0 {
				g.Detalle = append(g.Detalle, Linea{TiendaID: t.ID, Tienda: t.Nombre, Plan: t.Plan, Comision: t.ComisionMes})
				g.Monto += t.ComisionMes
			}
		}
		d, _ := json.Marshal(g.Detalle)
		tag, err := tx.Exec(ctx, `INSERT INTO partners.pagos (partner_id, mes, monto, detalle) VALUES ($1, $2, $3, $4)
			ON CONFLICT (partner_id, mes) DO UPDATE SET monto = excluded.monto, detalle = excluded.detalle
			WHERE partners.pagos.estado = 'pendiente'`, id, m, g.Monto, d)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrInvalido{"mes", "Ese mes ya está aprobado o pagado."}
		}
		return err
	})
	return g, err
}

// CambiarEstado pasa un pago a aprobada o pagada.
func CambiarEstado(ctx context.Context, pool *db.Pool, email, mes, estado string) error {
	m, err := parsearMes(mes)
	if err != nil {
		return err
	}
	if estado != "aprobada" && estado != "pagada" {
		return ErrInvalido{"estado", "El estado es aprobada o pagada."}
	}
	usuario, err := usuarioDe(ctx, pool, email)
	if err != nil {
		return err
	}
	return pool.EnUsuario(ctx, usuario, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE partners.pagos SET estado = $2,
			aprobado_en = coalesce(aprobado_en, now()),
			pagado_en = CASE WHEN $2 = 'pagada' THEN now() ELSE pagado_en END
			WHERE mes = $1`, m, estado)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrSinPago
		}
		return err
	})
}
