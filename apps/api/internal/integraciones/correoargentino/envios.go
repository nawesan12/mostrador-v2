package correoargentino

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Opcion cotizada por Correo: el precio en pesos enteros (redondeado para arriba) y el plazo en días
// hábiles si MiCorreo lo informa.
type Opcion struct {
	Precio   int64  `json:"precio"`
	Producto string `json:"producto"`
	DiasMin  string `json:"diasMin,omitempty"`
	DiasMax  string `json:"diasMax,omitempty"`
}

// Cotizacion de un código postal: a domicilio y a sucursal (nil: Correo no la ofreció).
type Cotizacion struct {
	Domicilio *Opcion `json:"domicilio"`
	Sucursal  *Opcion `json:"sucursal"`
}

// Lo cotizado se guarda 30 minutos (o hasta validTo): el checkout cotiza al tipear el CP y otra vez
// al confirmar, y no hace falta pedirle a Correo lo mismo dos veces.
const duracionCotizacion = 30 * time.Minute

// Cotizar el envío de una tienda a un código postal, para tantas unidades.
func (s *Servicio) Cotizar(ctx context.Context, tiendaID, cp string, unidades int) (Cotizacion, error) {
	cp = strings.TrimSpace(cp)
	if !cpValido.MatchString(cp) {
		return Cotizacion{}, ErrInvalido{"El código postal tiene que tener 4 números."}
	}
	c, err := s.conexionDe(ctx, tiendaID)
	if err != nil {
		return Cotizacion{}, err
	}
	m := c.paquete.medidas(unidades)
	k := fmt.Sprintf("%s|%s|%s|%d|%d|%d|%d", tiendaID, c.origen.CP, cp, m.Weight, m.Height, m.Width, m.Length)
	s.mu.Lock()
	g, ok := s.cotizados[k]
	s.mu.Unlock()
	if ok && g.vence.After(s.ahora()) {
		return g.c, nil
	}
	r, err := s.cliente().Tarifas(ctx, c.cred, PedidoTarifa{CustomerID: c.cliente, PostalCodeOrigin: c.origen.CP,
		PostalCodeDestination: cp, Dimensions: m})
	if err != nil {
		return Cotizacion{}, err
	}
	var out Cotizacion
	for _, t := range r.Rates {
		if t.Price <= 0 {
			continue
		}
		o := &Opcion{Precio: int64(math.Ceil(t.Price)), Producto: t.ProductName, DiasMin: string(t.DeliveryTimeMin), DiasMax: string(t.DeliveryTimeMax)}
		// Si vienen varios productos para la misma entrega, el más barato.
		switch t.DeliveredType {
		case "D":
			if out.Domicilio == nil || o.Precio < out.Domicilio.Precio {
				out.Domicilio = o
			}
		case "S":
			if out.Sucursal == nil || o.Precio < out.Sucursal.Precio {
				out.Sucursal = o
			}
		}
	}
	vence := s.ahora().Add(duracionCotizacion)
	if v, err := time.Parse(time.RFC3339, r.ValidTo); err == nil && v.Before(vence) {
		vence = v
	}
	s.mu.Lock()
	if s.cotizados == nil || len(s.cotizados) > 5000 {
		s.cotizados = map[string]cotizado{}
	}
	s.cotizados[k] = cotizado{out, vence}
	s.mu.Unlock()
	return out, nil
}

// CotizarPublico: la cotización del checkout de /t/{slug}.
func (s *Servicio) CotizarPublico(ctx context.Context, slug, cp string, unidades int) (Cotizacion, error) {
	id, err := s.tiendaDeSlug(ctx, slug)
	if err != nil {
		return Cotizacion{}, err
	}
	return s.Cotizar(ctx, id, cp, unidades)
}

// Sucursal de Correo donde se puede retirar un paquete.
type Sucursal struct {
	Codigo    string `json:"codigo"`
	Nombre    string `json:"nombre"`
	Direccion string `json:"direccion"`
	Localidad string `json:"localidad"`
	CP        string `json:"cp"`
	Horario   string `json:"horario"`
}

var digitos = regexp.MustCompile(`\d{4}`)

// cpNumero saca los 4 números de un CP (también de un CPA como B1842ZAB).
func cpNumero(cp string) int {
	n, err := strconv.Atoi(digitos.FindString(cp))
	if err != nil {
		return -1
	}
	return n
}

var dias = []struct{ k, t string }{{"monday", "Lun"}, {"tuesday", "Mar"}, {"wednesday", "Mié"}, {"thursday", "Jue"},
	{"friday", "Vie"}, {"saturday", "Sáb"}, {"sunday", "Dom"}}

// horario resumido: "Lun a Vie 10:00 a 18:00" (el del lunes, o el primero que haya).
func horario(h map[string]*struct{ Start, End string }) string {
	hora := func(s string) string {
		if len(s) == 4 {
			return s[:2] + ":" + s[2:]
		}
		return s
	}
	var abiertos []string
	var ref *struct{ Start, End string }
	for _, d := range dias {
		if x := h[d.k]; x != nil && x.Start != "" {
			abiertos = append(abiertos, d.t)
			if ref == nil {
				ref = x
			}
		}
	}
	if ref == nil {
		return ""
	}
	rango := abiertos[0]
	if len(abiertos) > 1 {
		rango += " a " + abiertos[len(abiertos)-1]
	}
	return fmt.Sprintf("%s %s a %s", rango, hora(ref.Start), hora(ref.End))
}

// Sucursales de una provincia que entregan paquetes, las más cercanas al CP primero. Se guardan 6 h.
func (s *Servicio) Sucursales(ctx context.Context, tiendaID, provincia, cp string) ([]Sucursal, error) {
	provincia = strings.ToUpper(strings.TrimSpace(provincia))
	if _, ok := Provincias[provincia]; !ok {
		return nil, ErrInvalido{"Elegí la provincia."}
	}
	c, err := s.conexionDe(ctx, tiendaID)
	if err != nil {
		return nil, err
	}
	k := tiendaID + "|" + c.cliente + "|" + provincia
	s.mu.Lock()
	g, ok := s.sucursales[k]
	s.mu.Unlock()
	var lista []Sucursal
	if ok && g.vence.After(s.ahora()) {
		lista = g.s
	} else {
		as, err := s.cliente().Agencias(ctx, c.cred, c.cliente, provincia)
		if err != nil {
			return nil, err
		}
		lista = []Sucursal{}
		for _, a := range as {
			if (a.Status != "" && a.Status != "ACTIVE") || !a.Services.PickupAvailability || a.Code == "" {
				continue
			}
			d := a.Location.Address
			dir := strings.TrimSpace(d.StreetName + " " + d.StreetNumber)
			loc := d.Locality
			if loc == "" {
				loc = d.City
			}
			lista = append(lista, Sucursal{Codigo: a.Code, Nombre: a.Name, Direccion: dir, Localidad: loc, CP: d.PostalCode,
				Horario: horario(a.Hours)})
		}
		s.mu.Lock()
		if s.sucursales == nil || len(s.sucursales) > 2000 {
			s.sucursales = map[string]sucursalesGuardadas{}
		}
		s.sucursales[k] = sucursalesGuardadas{lista, s.ahora().Add(6 * time.Hour)}
		s.mu.Unlock()
	}
	out := append([]Sucursal(nil), lista...)
	if n := cpNumero(cp); n > 0 {
		dist := func(x Sucursal) int {
			m := cpNumero(x.CP)
			if m < 0 {
				return math.MaxInt32
			}
			if m > n {
				return m - n
			}
			return n - m
		}
		sort.SliceStable(out, func(i, j int) bool { return dist(out[i]) < dist(out[j]) })
	}
	return out, nil
}

// SucursalesPublico: el selector de sucursal del checkout.
func (s *Servicio) SucursalesPublico(ctx context.Context, slug, provincia, cp string) ([]Sucursal, error) {
	id, err := s.tiendaDeSlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return s.Sucursales(ctx, id, provincia, cp)
}

// Sucursal busca una sucursal por código (para validar la que eligió quien compra).
func (s *Servicio) Sucursal(ctx context.Context, tiendaID, provincia, codigo string) (Sucursal, error) {
	l, err := s.Sucursales(ctx, tiendaID, provincia, "")
	if err != nil {
		return Sucursal{}, err
	}
	for _, x := range l {
		if x.Codigo == codigo {
			return x, nil
		}
	}
	return Sucursal{}, ErrInvalido{"Esa sucursal no está disponible. Elegí otra."}
}

// Entrega del pedido (pedidos.entrega) con lo que agrega Correo Argentino.
type Entrega struct {
	Direccion      string `json:"direccion"`
	CP             string `json:"cp"`
	Tipo           string `json:"tipo"`
	Provincia      string `json:"provincia"`
	Localidad      string `json:"localidad"`
	Sucursal       string `json:"sucursal"`
	SucursalNombre string `json:"sucursalNombre"`
}

// EnvioPedido es lo que ve el panel en el pedido (components/schemas/EnvioCorreo).
type EnvioPedido struct {
	Conectada   bool       `json:"conectada"`
	Generado    bool       `json:"generado"`
	Tipo        string     `json:"tipo,omitempty"`
	Sucursal    *string    `json:"sucursal,omitempty"`
	Importado   *time.Time `json:"importado,omitempty"`
	Seguimiento *string    `json:"seguimiento,omitempty"`
}

func (s *Servicio) envioDe(ctx context.Context, t pgx.Tx, tiendaID string, numero int) (EnvioPedido, error) {
	var e EnvioPedido
	var imp time.Time
	err := t.QueryRow(ctx, `SELECT tipo, sucursal, importado_en, seguimiento FROM ca_envios WHERE tienda_id = $1 AND pedido = $2`,
		tiendaID, numero).Scan(&e.Tipo, &e.Sucursal, &imp, &e.Seguimiento)
	if errors.Is(err, pgx.ErrNoRows) {
		return e, nil
	}
	if err != nil {
		return e, err
	}
	e.Generado, e.Importado = true, &imp
	return e, nil
}

// EnvioDe un pedido: si la tienda está conectada y si ya se dio de alta en MiCorreo.
func (s *Servicio) EnvioDe(ctx context.Context, tiendaID string, numero int) (EnvioPedido, error) {
	var e EnvioPedido
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var envio string
		if err := t.QueryRow(ctx, `SELECT envio FROM pedidos WHERE tienda_id = $1 AND numero = $2`, tiendaID, numero).Scan(&envio); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNoExiste
			}
			return err
		}
		if envio != "correo" {
			return ErrNoEsCorreo
		}
		var err error
		if e, err = s.envioDe(ctx, t, tiendaID, numero); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ca_conexiones WHERE tienda_id = $1)`, tiendaID).Scan(&e.Conectada)
	})
	return e, err
}

var calleAltura = regexp.MustCompile(`^(.*?\D)\s*(\d{1,6})\b[\s,]*(.*)$`)

// partirDireccion separa "Av. Belgrano 742 3B" en calle, altura y el resto (piso/depto, hasta 3 letras).
func partirDireccion(d string) (calle, altura, resto string) {
	d = strings.Join(strings.Fields(d), " ")
	m := calleAltura.FindStringSubmatch(d)
	if m == nil {
		return d, "S/N", ""
	}
	resto = strings.NewReplacer("piso", "", "Piso", "", "dto", "", "Dto", "", "depto", "", "Depto", "", " ", "", ",", "", ".", "").Replace(m[3])
	return strings.TrimSpace(m[1]), m[2], recortar(resto, 3)
}

func texto(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// Generar da de alta en MiCorreo el envío de un pedido de la tienda online. Si ya estaba dado de
// alta, devuelve lo que hay (un segundo toque no duplica: MiCorreo tampoco deja importar dos veces la
// misma referencia, y ese error se toma como hecho).
func (s *Servicio) Generar(ctx context.Context, tiendaID string, numero int) (EnvioPedido, error) {
	c, err := s.conexionDe(ctx, tiendaID)
	if err != nil {
		return EnvioPedido{}, err
	}
	var (
		envio, estado, nombreTienda string
		entrega                     *Entrega
		comprador                   struct {
			Nombre   string  `json:"nombre"`
			Email    *string `json:"email"`
			Telefono *string `json:"telefono"`
		}
		items []struct {
			Cantidad int `json:"cantidad"`
		}
		total   int64
		ya      EnvioPedido
		negoc   []byte
		negocio struct{ Whatsapp string }
	)
	err = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		err := t.QueryRow(ctx, `SELECT p.envio, p.estado, p.entrega, p.comprador, p.items, p.total, ti.nombre, ti.negocio
			FROM pedidos p JOIN tiendas ti ON ti.id = p.tienda_id WHERE p.tienda_id = $1 AND p.numero = $2`, tiendaID, numero).
			Scan(&envio, &estado, &entrega, &comprador, &items, &total, &nombreTienda, &negoc)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		ya, err = s.envioDe(ctx, t, tiendaID, numero)
		return err
	})
	if err != nil {
		return EnvioPedido{}, err
	}
	ya.Conectada = true
	if ya.Generado {
		return ya, nil
	}
	if envio != "correo" {
		return EnvioPedido{}, ErrNoEsCorreo
	}
	if estado == "Cancelado" {
		return EnvioPedido{}, ErrInvalido{"El pedido está cancelado."}
	}
	if entrega == nil || (entrega.Tipo != "D" && entrega.Tipo != "S") {
		return EnvioPedido{}, ErrInvalido{"El pedido no tiene los datos de entrega de Correo Argentino."}
	}
	if comprador.Email == nil || !emailValido.MatchString(*comprador.Email) {
		return EnvioPedido{}, ErrInvalido{"Correo Argentino pide el email de quien recibe y el pedido no lo tiene."}
	}
	_ = json.Unmarshal(negoc, &negocio)
	unidades := 0
	for _, it := range items {
		unidades += it.Cantidad
	}
	m := c.paquete.medidas(unidades)
	calle, altura, resto := partirDireccion(entrega.Direccion)
	piso := resto
	ref := fmt.Sprintf("TLW-%s-%d", strings.ReplaceAll(tiendaID, "-", "")[:8], numero)
	o := c.origen
	a := Importacion{
		CustomerID: c.cliente, ExtOrderID: ref, OrderNumber: strconv.Itoa(numero),
		Sender: Persona{Name: recortar(nombreTienda, 80), CellPhone: recortar(negocio.Whatsapp, 20), Email: c.email,
			OriginAddress: &Direccion{StreetName: o.Calle, StreetNumber: o.Altura, Floor: &o.Piso, Apartment: &o.Depto,
				City: o.Localidad, ProvinceCode: o.Provincia, PostalCode: o.CP}},
		Recipient: Persona{Name: recortar(comprador.Nombre, 80), CellPhone: recortar(texto(comprador.Telefono), 20),
			Email: *comprador.Email},
		Shipping: DatosEnvio{DeliveryType: entrega.Tipo,
			Address: Direccion{StreetName: calle, StreetNumber: altura, Floor: &piso, Apartment: new(string),
				City: entrega.Localidad, ProvinceCode: entrega.Provincia, PostalCode: entrega.CP},
			Weight: m.Weight, DeclaredValue: float64(total), Height: m.Height, Length: m.Length, Width: m.Width},
	}
	var suc *string
	if entrega.Tipo == "S" {
		if entrega.Sucursal == "" {
			return EnvioPedido{}, ErrInvalido{"El pedido no tiene la sucursal elegida."}
		}
		suc = &entrega.Sucursal
		a.Shipping.Agency = suc
	} else if entrega.Localidad == "" || entrega.Provincia == "" {
		return EnvioPedido{}, ErrInvalido{"Falta la localidad o la provincia de la entrega."}
	}
	err = s.cliente().Importar(ctx, c.cred, a)
	var e *ErrorCA
	if errors.As(err, &e) && strings.Contains(strings.ToLower(e.Mensaje), "importada con anterioridad") {
		// La respuesta anterior se cortó, pero MiCorreo ya la tenía: queda como hecha.
		err = nil
	}
	if err != nil {
		return EnvioPedido{}, err
	}
	err = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, `INSERT INTO ca_envios (tienda_id, pedido, tipo, sucursal, referencia) VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING`, tiendaID, numero, entrega.Tipo, suc, ref); err != nil {
			return err
		}
		ya, err = s.envioDe(ctx, t, tiendaID, numero)
		return err
	})
	ya.Conectada = true
	return ya, err
}

var seguimientoValido = regexp.MustCompile(`^[A-Z0-9]{6,40}$`)

// Seguimiento guarda el número que asignó MiCorreo al pagar el rótulo (también en el pedido, que es
// lo que ve quien compra en su página del pedido).
func (s *Servicio) Seguimiento(ctx context.Context, tiendaID string, numero int, codigo string) (EnvioPedido, error) {
	codigo = strings.ToUpper(strings.Join(strings.Fields(codigo), ""))
	if !seguimientoValido.MatchString(codigo) {
		return EnvioPedido{}, ErrInvalido{"El número de seguimiento son de 6 a 40 letras y números."}
	}
	var e EnvioPedido
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE ca_envios SET seguimiento = $3 WHERE tienda_id = $1 AND pedido = $2`, tiendaID, numero, codigo)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrInvalido{"Primero generá el envío en Correo Argentino."}
		}
		if _, err := t.Exec(ctx, `UPDATE pedidos SET seguimiento = $3 WHERE tienda_id = $1 AND numero = $2`, tiendaID, numero, codigo); err != nil {
			return err
		}
		e, err = s.envioDe(ctx, t, tiendaID, numero)
		return err
	})
	e.Conectada = true
	return e, err
}
