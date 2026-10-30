// Package correoargentino: envíos con Correo Argentino por la API MiCorreo
// (docs/integraciones/correo-argentino.md).
//
//   - /token con HTTP Basic (credenciales de API que da Correo, por ambiente) devuelve un JWT; el
//     resto va con Bearer. Las de Tu Local Web sirven para todas las tiendas; una tienda puede usar
//     las suyas.
//   - Cada tienda conecta su cuenta de MiCorreo con email y contraseña: /users/validate devuelve el
//     customerId, que es lo único que se guarda de la cuenta.
//   - Cotiza (/rates, a domicilio y a sucursal), lista sucursales (/agencies) y da de alta envíos
//     (/shipping/import). El seguimiento: MiCorreo no lo expone por API; el número lo carga la tienda.
package correoargentino

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Bases de la API MiCorreo.
const (
	BaseProduccion = "https://api.correoargentino.com.ar/micorreo/v1"
	BasePruebas    = "https://apitest.correoargentino.com.ar/micorreo/v1"
)

// Credenciales de API para /token (las pide quien integra a Correo, una por ambiente).
type Credenciales struct {
	Usuario string `json:"usuario"`
	Clave   string `json:"clave"`
}

func (c Credenciales) vacias() bool { return c.Usuario == "" || c.Clave == "" }

// ErrorCA es una respuesta no exitosa de MiCorreo ({"code": "402", "message": "..."}). El mensaje
// es de Correo: sirve para el panel (peso, sucursal, provincia) pero nunca se le muestra a quien compra.
type ErrorCA struct {
	Estado  int
	Mensaje string
}

func (e *ErrorCA) Error() string {
	return fmt.Sprintf("correoargentino: HTTP %d: %s", e.Estado, e.Mensaje)
}

// Cliente habla con la API MiCorreo y guarda el JWT de cada credencial hasta que vence.
type Cliente struct {
	Base string
	HTTP *http.Client
	// Ahora: para los tests del vencimiento del token.
	Ahora func() time.Time

	mu     sync.Mutex
	tokens map[string]token
}

type token struct {
	valor string
	vence time.Time
}

func (c *Cliente) base() string {
	if c.Base != "" {
		return strings.TrimRight(c.Base, "/")
	}
	return BaseProduccion
}

func (c *Cliente) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func (c *Cliente) ahora() time.Time {
	if c.Ahora != nil {
		return c.Ahora()
	}
	return time.Now()
}

var argentina = time.FixedZone("ART", -3*3600)

func clave(cr Credenciales) string {
	h := sha256.Sum256([]byte(cr.Usuario + "\x00" + cr.Clave))
	return hex.EncodeToString(h[:])
}

// tokenDe devuelve el JWT de la credencial, pidiendo uno nuevo si falta o vence en menos de 5 min.
func (c *Cliente) tokenDe(ctx context.Context, cr Credenciales) (string, error) {
	if cr.vacias() {
		return "", ErrSinCredenciales
	}
	k := clave(cr)
	c.mu.Lock()
	t, ok := c.tokens[k]
	c.mu.Unlock()
	if ok && t.vence.After(c.ahora().Add(5*time.Minute)) {
		return t.valor, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base()+"/token", nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(cr.Usuario, cr.Clave)
	req.Header.Set("Accept", "application/json")
	var r struct {
		Token   string `json:"token"`
		Expires string `json:"expires"`
	}
	if err := c.hacer(req, &r); err != nil {
		return "", err
	}
	if r.Token == "" {
		return "", &ErrorCA{Estado: http.StatusBadGateway, Mensaje: "el token vino vacío"}
	}
	// "expires": "2022-04-26 21:16:20", en hora de Argentina. Si no se entiende, se usa 30 min.
	vence, err := time.ParseInLocation("2006-01-02 15:04:05", r.Expires, argentina)
	if err != nil {
		vence = c.ahora().Add(30 * time.Minute)
	}
	c.mu.Lock()
	if c.tokens == nil {
		c.tokens = map[string]token{}
	}
	c.tokens[k] = token{r.Token, vence}
	c.mu.Unlock()
	return r.Token, nil
}

func (c *Cliente) olvidar(cr Credenciales) {
	c.mu.Lock()
	delete(c.tokens, clave(cr))
	c.mu.Unlock()
}

// pedir hace un pedido con Bearer. Un 401 con un token guardado se reintenta una vez con uno nuevo
// (Correo puede invalidarlo antes de "expires").
func (c *Cliente) pedir(ctx context.Context, cr Credenciales, metodo, ruta string, q url.Values, cuerpo, salida any) error {
	for intento := 0; ; intento++ {
		tk, err := c.tokenDe(ctx, cr)
		if err != nil {
			return err
		}
		var r io.Reader
		if cuerpo != nil {
			b, err := json.Marshal(cuerpo)
			if err != nil {
				return err
			}
			r = bytes.NewReader(b)
		}
		u := c.base() + ruta
		if len(q) > 0 {
			u += "?" + q.Encode()
		}
		req, err := http.NewRequestWithContext(ctx, metodo, u, r)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+tk)
		req.Header.Set("Accept", "application/json")
		if cuerpo != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		err = c.hacer(req, salida)
		var e *ErrorCA
		if intento == 0 && errors.As(err, &e) && e.Estado == http.StatusUnauthorized {
			c.olvidar(cr)
			continue
		}
		return err
	}
}

func (c *Cliente) hacer(req *http.Request, salida any) error {
	res, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if res.StatusCode >= 300 {
		// {"code": "402" | 402, "message": "..."}
		var p struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(b, &p)
		return &ErrorCA{Estado: res.StatusCode, Mensaje: recortar(p.Message, 200)}
	}
	if salida == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, salida)
}

func recortar(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// Validar devuelve el customerId de una cuenta de MiCorreo.
func (c *Cliente) Validar(ctx context.Context, cr Credenciales, email, clave string) (string, error) {
	var r struct {
		CustomerID string `json:"customerId"`
	}
	if err := c.pedir(ctx, cr, http.MethodPost, "/users/validate", nil,
		map[string]string{"email": email, "password": clave}, &r); err != nil {
		return "", err
	}
	if r.CustomerID == "" {
		return "", &ErrorCA{Estado: http.StatusNotFound, Mensaje: "Usuario no valido o inexistente"}
	}
	return r.CustomerID, nil
}

// Medidas del paquete: gramos y centímetros enteros (peso de 1 a 25000 g, lados hasta 150 cm).
type Medidas struct {
	Weight int `json:"weight"`
	Height int `json:"height"`
	Width  int `json:"width"`
	Length int `json:"length"`
}

// PedidoTarifa es el cuerpo de /rates. Sin DeliveredType trae las dos (domicilio y sucursal).
type PedidoTarifa struct {
	CustomerID            string  `json:"customerId"`
	PostalCodeOrigin      string  `json:"postalCodeOrigin"`
	PostalCodeDestination string  `json:"postalCodeDestination"`
	DeliveredType         string  `json:"deliveredType,omitempty"`
	Dimensions            Medidas `json:"dimensions"`
}

// Tarifa de /rates. Price en pesos con centavos; los plazos, en días hábiles, cuando vienen.
type Tarifa struct {
	DeliveredType   string  `json:"deliveredType"`
	ProductType     string  `json:"productType"`
	ProductName     string  `json:"productName"`
	Price           float64 `json:"price"`
	DeliveryTimeMin Texto   `json:"deliveryTimeMin"`
	DeliveryTimeMax Texto   `json:"deliveryTimeMax"`
}

// Texto acepta un string o un número (los plazos vienen de las dos formas según la versión).
type Texto string

func (t *Texto) UnmarshalJSON(b []byte) error {
	var s string
	if json.Unmarshal(b, &s) == nil {
		*t = Texto(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		*t = ""
		return nil
	}
	*t = Texto(n.String())
	return nil
}

// RespuestaTarifas de /rates.
type RespuestaTarifas struct {
	CustomerID string   `json:"customerId"`
	ValidTo    string   `json:"validTo"`
	Rates      []Tarifa `json:"rates"`
}

// Tarifas cotiza un envío.
func (c *Cliente) Tarifas(ctx context.Context, cr Credenciales, p PedidoTarifa) (RespuestaTarifas, error) {
	var r RespuestaTarifas
	err := c.pedir(ctx, cr, http.MethodPost, "/rates", nil, p, &r)
	return r, err
}

// Direccion de MiCorreo (sucursales, remitente y destino).
type Direccion struct {
	StreetName   string  `json:"streetName"`
	StreetNumber string  `json:"streetNumber"`
	Floor        *string `json:"floor"`
	Apartment    *string `json:"apartment"`
	Locality     string  `json:"locality,omitempty"`
	City         string  `json:"city"`
	Province     string  `json:"province,omitempty"`
	ProvinceCode string  `json:"provinceCode"`
	PostalCode   string  `json:"postalCode"`
}

// Agencia es una sucursal de /agencies.
type Agencia struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Services struct {
		PackageReception   bool `json:"packageReception"`
		PickupAvailability bool `json:"pickupAvailability"`
	} `json:"services"`
	Location struct {
		Address   Direccion `json:"address"`
		Latitude  string    `json:"latitude"`
		Longitude string    `json:"longitude"`
	} `json:"location"`
	Hours  map[string]*struct{ Start, End string } `json:"hours"`
	Status string                                  `json:"status"`
}

// Agencias de una provincia que entregan paquetes (pickup_availability).
func (c *Cliente) Agencias(ctx context.Context, cr Credenciales, cliente, provincia string) ([]Agencia, error) {
	var r []Agencia
	err := c.pedir(ctx, cr, http.MethodGet, "/agencies", url.Values{
		"customerId": {cliente}, "provinceCode": {provincia}, "services": {"pickup_availability"},
	}, nil, &r)
	return r, err
}

// Persona del envío (remitente o destinatario).
type Persona struct {
	Name          string     `json:"name"`
	Phone         string     `json:"phone"`
	CellPhone     string     `json:"cellPhone"`
	Email         string     `json:"email"`
	OriginAddress *Direccion `json:"originAddress,omitempty"`
}

// DatosEnvio es shipping de /shipping/import.
type DatosEnvio struct {
	DeliveryType  string    `json:"deliveryType"`
	Agency        *string   `json:"agency"`
	Address       Direccion `json:"address"`
	Weight        int       `json:"weight"`
	DeclaredValue float64   `json:"declaredValue"`
	Height        int       `json:"height"`
	Length        int       `json:"length"`
	Width         int       `json:"width"`
}

// Importacion es el cuerpo de /shipping/import.
type Importacion struct {
	CustomerID  string     `json:"customerId"`
	ExtOrderID  string     `json:"extOrderId"`
	OrderNumber string     `json:"orderNumber"`
	Sender      Persona    `json:"sender"`
	Recipient   Persona    `json:"recipient"`
	Shipping    DatosEnvio `json:"shipping"`
}

// Importar da de alta el envío en MiCorreo (queda para pagar e imprimir el rótulo allá).
func (c *Cliente) Importar(ctx context.Context, cr Credenciales, a Importacion) error {
	return c.pedir(ctx, cr, http.MethodPost, "/shipping/import", nil, a, nil)
}
