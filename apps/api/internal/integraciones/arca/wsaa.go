package arca

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"
)

// Ambiente de ARCA: homologación (pruebas, certificado de WSASS) o producción.
type Ambiente string

const (
	Homologacion Ambiente = "homologacion"
	Produccion   Ambiente = "produccion"
)

// Direcciones de los servicios (se pueden cambiar en los tests).
type Direcciones struct{ WSAA, WSFE string }

func DireccionesDe(a Ambiente) Direcciones {
	if a == Produccion {
		return Direcciones{WSAA: "https://wsaa.afip.gov.ar/ws/services/LoginCms",
			WSFE: "https://servicios1.afip.gov.ar/wsfev1/service.asmx"}
	}
	return Direcciones{WSAA: "https://wsaahomo.afip.gov.ar/ws/services/LoginCms",
		WSFE: "https://wswhomo.afip.gov.ar/wsfev1/service.asmx"}
}

// Ticket de acceso (TA): dura 12 h y ARCA no da otro mientras haya uno vigente ("El CEE ya posee
// un TA valido"), así que se guarda y se reusa.
type Ticket struct {
	Token string    `json:"token"`
	Sign  string    `json:"sign"`
	Vence time.Time `json:"vence"`
}

func (t Ticket) Vigente(ahora time.Time) bool {
	return t.Token != "" && ahora.Add(10*time.Minute).Before(t.Vence)
}

var arg = time.FixedZone("ART", -3*3600)

// pedidoAcceso es el loginTicketRequest (TRA).
func pedidoAcceso(servicio string, ahora time.Time) []byte {
	a := ahora.In(arg)
	return fmt.Appendf(nil, `<?xml version="1.0" encoding="UTF-8"?><loginTicketRequest version="1.0"><header><uniqueId>%d</uniqueId><generationTime>%s</generationTime><expirationTime>%s</expirationTime></header><service>%s</service></loginTicketRequest>`,
		a.Unix(), a.Add(-5*time.Minute).Format(time.RFC3339), a.Add(10*time.Minute).Format(time.RFC3339), servicio)
}

// ErrARCA es un error que contestó ARCA (con su código, lo único accionable).
type ErrARCA struct {
	Codigo  string
	Mensaje string
}

func (e ErrARCA) Error() string { return "arca: " + e.Codigo + " " + e.Mensaje }

func soap(ctx context.Context, cli *http.Client, url, accion string, cuerpo []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(cuerpo))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", accion)
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	// Un fault SOAP viene con 500 y el detalle adentro.
	var f struct {
		Code   string `xml:"Body>Fault>faultcode"`
		String string `xml:"Body>Fault>faultstring"`
	}
	if xml.Unmarshal(b, &f) == nil && f.String != "" {
		return nil, ErrARCA{Codigo: f.Code, Mensaje: f.String}
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arca: HTTP %d", res.StatusCode)
	}
	return b, nil
}

// PedirTicket firma el pedido de acceso y lo canjea en el WSAA por un ticket para el servicio.
func PedirTicket(ctx context.Context, cli *http.Client, dir Direcciones, certPEM, clavePEM []byte, servicio string, ahora time.Time) (Ticket, error) {
	cert, err := leerCert(certPEM)
	if err != nil {
		return Ticket{}, err
	}
	k, err := leerClave(clavePEM)
	if err != nil {
		return Ticket{}, err
	}
	cms, err := FirmarCMS(pedidoAcceso(servicio, ahora), cert, k, ahora)
	if err != nil {
		return Ticket{}, err
	}
	cuerpo := fmt.Appendf(nil, `<?xml version="1.0" encoding="UTF-8"?><soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:wsaa="http://wsaa.view.sua.dvadac.desein.afip.gov"><soapenv:Header/><soapenv:Body><wsaa:loginCms><wsaa:in0>%s</wsaa:in0></wsaa:loginCms></soapenv:Body></soapenv:Envelope>`,
		base64.StdEncoding.EncodeToString(cms))
	b, err := soap(ctx, cli, dir.WSAA, `""`, cuerpo)
	if err != nil {
		return Ticket{}, err
	}
	var env struct {
		Ret string `xml:"Body>loginCmsResponse>loginCmsReturn"`
	}
	if err := xml.Unmarshal(b, &env); err != nil || env.Ret == "" {
		return Ticket{}, errors.New("arca: respuesta del WSAA sin ticket")
	}
	// El ticket viaja como XML escapado dentro del XML.
	var lt struct {
		Vence string `xml:"header>expirationTime"`
		Token string `xml:"credentials>token"`
		Sign  string `xml:"credentials>sign"`
	}
	ret := env.Ret
	if !strings.HasPrefix(strings.TrimSpace(ret), "<") {
		ret = html.UnescapeString(ret)
	}
	if err := xml.Unmarshal([]byte(ret), &lt); err != nil || lt.Token == "" || lt.Sign == "" {
		return Ticket{}, errors.New("arca: ticket ilegible")
	}
	vence, err := time.Parse(time.RFC3339Nano, lt.Vence)
	if err != nil {
		vence = ahora.Add(11 * time.Hour)
	}
	return Ticket{Token: lt.Token, Sign: lt.Sign, Vence: vence}, nil
}
