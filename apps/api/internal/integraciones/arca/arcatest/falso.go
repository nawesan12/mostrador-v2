// Package arcatest: un ARCA falso para los tests, difícil de contentar. El WSAA exige un CMS
// SignedData y no da otro ticket mientras haya uno vigente; el WSFE valida importes (10048), IVA,
// factura C sin IVA, condición del receptor y numeración correlativa (10016), y guarda lo emitido
// para FECompConsultar.
package arcatest

import (
	"encoding/asn1"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type emitido struct{ cae, total string }

type Falso struct {
	*httptest.Server
	mu sync.Mutex
	// Último número autorizado por tipo.
	Ultimo map[int]int
	// Lo último que llegó al WSFE (para mirar el cable).
	UltimoPedido string
	// Tickets entregados.
	Tickets int
	// Si es true, la próxima autorización se hace pero la respuesta "se corta" (500 sin cuerpo).
	CortarDespuesDeAutorizar bool
	emitidos                 map[string]emitido
}

var oidSignedData = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}

func Nuevo() *Falso {
	f := &Falso{Ultimo: map[int]int{}, emitidos: map[string]emitido{}}
	f.Server = httptest.NewServer(http.HandlerFunc(f.servir))
	return f
}

// WSAA y WSFE son las direcciones para arca.Direcciones.
func (f *Falso) WSAA() string { return f.URL + "/wsaa" }
func (f *Falso) WSFE() string { return f.URL + "/wsfe" }

func esc(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

func fault(w http.ResponseWriter, code, msg string) {
	w.WriteHeader(500)
	fmt.Fprintf(w, `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"><soapenv:Body><soapenv:Fault><faultcode>%s</faultcode><faultstring>%s</faultstring></soapenv:Fault></soapenv:Body></soapenv:Envelope>`, code, msg)
}

func (f *Falso) servir(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, _ := io.ReadAll(r.Body)
	s := string(b)
	if strings.HasSuffix(r.URL.Path, "/wsaa") {
		m := regexp.MustCompile(`<wsaa:in0>([^<]+)</wsaa:in0>`).FindStringSubmatch(s)
		var ci struct{ Type asn1.ObjectIdentifier }
		if m == nil {
			fault(w, "ns1:cms.bad", "CMS no es valido")
			return
		}
		der, err := base64.StdEncoding.DecodeString(m[1])
		if _, e := asn1.Unmarshal(der, &ci); err != nil || e != nil || !ci.Type.Equal(oidSignedData) {
			fault(w, "ns1:cms.bad", "CMS no es valido")
			return
		}
		if f.Tickets > 0 {
			fault(w, "ns1:coe.alreadyAuthenticated", "El CEE ya posee un TA valido para el acceso al WSN solicitado")
			return
		}
		f.Tickets++
		ta := `<?xml version="1.0" encoding="UTF-8"?><loginTicketResponse version="1.0"><header><expirationTime>2099-01-01T00:00:00.000-03:00</expirationTime></header><credentials><token>TOKEN&amp;1</token><sign>FIRMA</sign></credentials></loginTicketResponse>`
		fmt.Fprintf(w, `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"><soapenv:Body><loginCmsResponse xmlns="http://wsaa.view.sua.dvadac.desein.afip.gov"><loginCmsReturn>%s</loginCmsReturn></loginCmsResponse></soapenv:Body></soapenv:Envelope>`, esc(ta))
		return
	}
	f.UltimoPedido = s
	v := func(tag string) string {
		m := regexp.MustCompile(`<ar:` + tag + `>([^<]*)</ar:` + tag + `>`).FindStringSubmatch(s)
		if m == nil {
			return ""
		}
		return m[1]
	}
	op := r.Header.Get("SOAPAction")
	op = op[strings.LastIndex(op, "/")+1:]
	if v("Token") != "TOKEN&amp;1" || v("Sign") != "FIRMA" {
		fmt.Fprintf(w, `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><%[1]sResponse xmlns="http://ar.gov.afip.dif.FEV1/"><%[1]sResult><Errors><Err><Code>600</Code><Msg>ValidacionDeToken: No validaron las firmas</Msg></Err></Errors></%[1]sResult></%[1]sResponse></soap:Body></soap:Envelope>`, op)
		return
	}
	tipo, _ := strconv.Atoi(v("CbteTipo"))
	pv := v("PtoVta")
	switch op {
	case "FECompUltimoAutorizado":
		fmt.Fprintf(w, `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><FECompUltimoAutorizadoResponse xmlns="http://ar.gov.afip.dif.FEV1/"><FECompUltimoAutorizadoResult><CbteNro>%d</CbteNro></FECompUltimoAutorizadoResult></FECompUltimoAutorizadoResponse></soap:Body></soap:Envelope>`, f.Ultimo[tipo])
	case "FECompConsultar":
		nro := v("CbteNro")
		e, ok := f.emitidos[fmt.Sprintf("%d-%s-%s", tipo, pv, nro)]
		if !ok {
			fmt.Fprint(w, `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><FECompConsultarResponse xmlns="http://ar.gov.afip.dif.FEV1/"><FECompConsultarResult><Errors><Err><Code>602</Code><Msg>No existen datos</Msg></Err></Errors></FECompConsultarResult></FECompConsultarResponse></soap:Body></soap:Envelope>`)
			return
		}
		fmt.Fprintf(w, `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><FECompConsultarResponse xmlns="http://ar.gov.afip.dif.FEV1/"><FECompConsultarResult><ResultGet><CodAutorizacion>%s</CodAutorizacion><FchVto>20991231</FchVto><ImpTotal>%s</ImpTotal></ResultGet></FECompConsultarResult></FECompConsultarResponse></soap:Body></soap:Envelope>`, e.cae, e.total)
	case "FECAESolicitar":
		num := func(tag string) float64 { x, _ := strconv.ParseFloat(v(tag), 64); return x }
		total, neto, iva := num("ImpTotal"), num("ImpNeto"), num("ImpIVA")
		desde, _ := strconv.Atoi(v("CbteDesde"))
		rech := func(code int, msg string) {
			fmt.Fprintf(w, `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><FECAESolicitarResponse xmlns="http://ar.gov.afip.dif.FEV1/"><FECAESolicitarResult><FeCabResp><Resultado>R</Resultado></FeCabResp><FeDetResp><FECAEDetResponse><Resultado>R</Resultado><CbteDesde>%d</CbteDesde><Observaciones><Obs><Code>%d</Code><Msg>%s</Msg></Obs></Observaciones></FECAEDetResponse></FeDetResp></FECAESolicitarResult></FECAESolicitarResponse></soap:Body></soap:Envelope>`, desde, code, msg)
		}
		switch {
		case math.Abs(neto+iva-total) > 0.001:
			rech(10048, "ImpTotal debe ser igual a la suma de ImpNeto, ImpIVA y demas")
		case (tipo == 1 || tipo == 6 || tipo == 3 || tipo == 8) && math.Abs(math.Round(neto*21)/100-iva) > 0.011:
			rech(10051, "El IVA no corresponde a la alicuota")
		case (tipo == 11 || tipo == 13) && iva != 0:
			rech(10071, "Factura C no lleva IVA")
		case v("CondicionIVAReceptorId") == "":
			rech(10242, "Falta CondicionIVAReceptorId")
		case desde != f.Ultimo[tipo]+1:
			rech(10016, "El numero de comprobante no es el siguiente")
		default:
			f.Ultimo[tipo] = desde
			cae := fmt.Sprintf("764300%08d", desde)
			f.emitidos[fmt.Sprintf("%d-%s-%d", tipo, pv, desde)] = emitido{cae, v("ImpTotal")}
			if f.CortarDespuesDeAutorizar {
				f.CortarDespuesDeAutorizar = false
				w.WriteHeader(502)
				return
			}
			fmt.Fprintf(w, `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><FECAESolicitarResponse xmlns="http://ar.gov.afip.dif.FEV1/"><FECAESolicitarResult><FeCabResp><Resultado>A</Resultado></FeCabResp><FeDetResp><FECAEDetResponse><Resultado>A</Resultado><CbteDesde>%d</CbteDesde><CAE>%s</CAE><CAEFchVto>20991231</CAEFchVto></FECAEDetResponse></FeDetResp></FECAESolicitarResult></FECAESolicitarResponse></soap:Body></soap:Envelope>`, desde, cae)
		}
	}
}
