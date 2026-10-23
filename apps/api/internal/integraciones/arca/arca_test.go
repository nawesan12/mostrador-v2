package arca

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCUIT(t *testing.T) {
	for cuit, ok := range map[string]bool{"20448592341": true, "20448592342": false, "30500010912": true, "2044859234": false, "2044859234a": false} {
		if CUITValido(cuit) != ok {
			t.Errorf("%s: %v", cuit, !ok)
		}
	}
}

func TestPedidoDeCertificadoYCertificado(t *testing.T) {
	clave, err := NuevaClave()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PedidoCertificado(clave, "20448592342", "Tienda", "mostrador"); err == nil {
		t.Fatal("CUIT con dígito verificador mal")
	}
	csrPEM, err := PedidoCertificado(clave, "20448592341", "Nahuel Santillán", "mostrador")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath("openssl"); err == nil {
		dir := t.TempDir()
		_ = os.WriteFile(filepath.Join(dir, "p.csr"), csrPEM, 0o600)
		out, err := exec.Command("openssl", "req", "-in", filepath.Join(dir, "p.csr"), "-noout", "-verify", "-subject").CombinedOutput()
		if err != nil || !strings.Contains(string(out), "serialNumber") || !strings.Contains(string(out), "CUIT 20448592341") {
			t.Fatalf("openssl req: %v %s", err, out)
		}
	}
	// ARCA firma el CSR: lo simulamos con una CA de prueba ("Computadores Test").
	b, _ := pem.Decode(csrPEM)
	csr, _ := x509.ParseCertificateRequest(b.Bytes)
	caK, _ := leerClave(must(NuevaClave()))
	ca := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Computadores Test"}, IsCA: true,
		BasicConstraintsValid: true, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(48 * time.Hour)}
	tpl := &x509.Certificate{SerialNumber: big.NewInt(77), Subject: csr.Subject, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tpl, ca, csr.PublicKey, caK)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	d, err := RevisarCertificado(certPEM, clave, "20448592341", time.Now())
	if err != nil || d.CUIT != "20448592341" || !d.Homolog {
		t.Fatalf("revisar: %+v %v", d, err)
	}
	if _, err := RevisarCertificado(certPEM, must(NuevaClave()), "20448592341", time.Now()); err != ErrCertificadoAjeno {
		t.Fatalf("con otra clave: %v", err)
	}
	if _, err := RevisarCertificado(certPEM, clave, "30500010912", time.Now()); err != ErrCertificadoAjeno {
		t.Fatalf("de otro CUIT: %v", err)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// arcaFalso valida como ARCA: el WSAA exige un CMS cuyo certificado corresponda a la firma, y el
// WSFE controla importes (10048), tipo de comprobante, condición del receptor y numeración.
type arcaFalso struct {
	t         *testing.T
	ultimo    map[int]int
	ultimoReq string
	tickets   int
}

func (f *arcaFalso) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/wsaa", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		m := regexp.MustCompile(`<wsaa:in0>([^<]+)</wsaa:in0>`).FindSubmatch(b)
		if m == nil {
			http.Error(w, "sin in0", 500)
			return
		}
		cms, err := base64.StdEncoding.DecodeString(string(m[1]))
		var ci struct {
			Type asn1.ObjectIdentifier
		}
		if err != nil || func() error { _, e := asn1.Unmarshal(cms, &ci); return e }() != nil || !ci.Type.Equal(oidSignedData) {
			w.WriteHeader(500)
			fmt.Fprint(w, `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"><soapenv:Body><soapenv:Fault><faultcode>ns1:cms.bad</faultcode><faultstring>CMS no es valido</faultstring></soapenv:Fault></soapenv:Body></soapenv:Envelope>`)
			return
		}
		if f.tickets > 0 {
			w.WriteHeader(500)
			fmt.Fprint(w, `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"><soapenv:Body><soapenv:Fault><faultcode>ns1:coe.alreadyAuthenticated</faultcode><faultstring>El CEE ya posee un TA valido para el acceso al WSN solicitado</faultstring></soapenv:Fault></soapenv:Body></soapenv:Envelope>`)
			return
		}
		f.tickets++
		ta := `<?xml version="1.0" encoding="UTF-8"?><loginTicketResponse version="1.0"><header><expirationTime>2026-10-24T03:00:00.000-03:00</expirationTime></header><credentials><token>TOKEN&amp;1</token><sign>FIRMA</sign></credentials></loginTicketResponse>`
		fmt.Fprintf(w, `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"><soapenv:Body><loginCmsResponse xmlns="http://wsaa.view.sua.dvadac.desein.afip.gov"><loginCmsReturn>%s</loginCmsReturn></loginCmsResponse></soapenv:Body></soapenv:Envelope>`,
			escapar(ta))
	})
	mux.HandleFunc("/wsfe", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		s := string(b)
		f.ultimoReq = s
		v := func(tag string) string {
			m := regexp.MustCompile(`<ar:` + tag + `>([^<]*)</ar:` + tag + `>`).FindStringSubmatch(s)
			if m == nil {
				return ""
			}
			return m[1]
		}
		if v("Token") != "TOKEN&amp;1" || v("Sign") != "FIRMA" {
			fmt.Fprint(w, wsfeErr("FECompUltimoAutorizado", 600, "ValidacionDeToken: No validaron las firmas digitales"))
			return
		}
		tipo, _ := strconv.Atoi(v("CbteTipo"))
		switch r.Header.Get("SOAPAction") {
		case "http://ar.gov.afip.dif.FEV1/FECompUltimoAutorizado":
			fmt.Fprintf(w, `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><FECompUltimoAutorizadoResponse xmlns="http://ar.gov.afip.dif.FEV1/"><FECompUltimoAutorizadoResult><PtoVta>%s</PtoVta><CbteTipo>%d</CbteTipo><CbteNro>%d</CbteNro></FECompUltimoAutorizadoResult></FECompUltimoAutorizadoResponse></soap:Body></soap:Envelope>`,
				v("PtoVta"), tipo, f.ultimo[tipo])
		case "http://ar.gov.afip.dif.FEV1/FECAESolicitar":
			num := func(tag string) float64 { x, _ := strconv.ParseFloat(v(tag), 64); return x }
			total, neto, iva := num("ImpTotal"), num("ImpNeto"), num("ImpIVA")
			desde, _ := strconv.Atoi(v("CbteDesde"))
			rechazo := func(code int, msg string) {
				fmt.Fprintf(w, `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><FECAESolicitarResponse xmlns="http://ar.gov.afip.dif.FEV1/"><FECAESolicitarResult><FeCabResp><Resultado>R</Resultado></FeCabResp><FeDetResp><FECAEDetResponse><Resultado>R</Resultado><CbteDesde>%d</CbteDesde><Observaciones><Obs><Code>%d</Code><Msg>%s</Msg></Obs></Observaciones></FECAEDetResponse></FeDetResp></FECAESolicitarResult></FECAESolicitarResponse></soap:Body></soap:Envelope>`,
					desde, code, msg)
			}
			switch {
			case math.Abs(neto+iva-total) > 0.001:
				rechazo(10048, "ImpTotal debe ser igual a la suma de ImpNeto, ImpIVA y demás")
			case (tipo == FacturaA || tipo == FacturaB) && math.Abs(math.Round(neto*21)/100-iva) > 0.011:
				rechazo(10051, "El IVA no corresponde a la alícuota")
			case tipo == FacturaC && iva != 0:
				rechazo(10071, "Factura C no lleva IVA")
			case v("CondicionIVAReceptorId") == "":
				rechazo(10242, "Falta CondicionIVAReceptorId")
			case desde != f.ultimo[tipo]+1:
				rechazo(10016, "El número de comprobante no es el siguiente")
			default:
				f.ultimo[tipo] = desde
				fmt.Fprintf(w, `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><FECAESolicitarResponse xmlns="http://ar.gov.afip.dif.FEV1/"><FECAESolicitarResult><FeCabResp><Resultado>A</Resultado></FeCabResp><FeDetResp><FECAEDetResponse><Resultado>A</Resultado><CbteDesde>%d</CbteDesde><CAE>7643000000%04d</CAE><CAEFchVto>20261102</CAEFchVto></FECAEDetResponse></FeDetResp></FECAESolicitarResult></FECAESolicitarResponse></soap:Body></soap:Envelope>`,
					desde, desde)
			}
		}
	})
	return mux
}

func wsfeErr(op string, code int, msg string) string {
	return fmt.Sprintf(`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><%[1]sResponse xmlns="http://ar.gov.afip.dif.FEV1/"><%[1]sResult><Errors><Err><Code>%d</Code><Msg>%s</Msg></Err></Errors></%[1]sResult></%[1]sResponse></soap:Body></soap:Envelope>`,
		op, code, msg)
}

func TestTicketYCAE(t *testing.T) {
	ctx := context.Background()
	f := &arcaFalso{t: t, ultimo: map[int]int{FacturaB: 41}}
	srv := httptest.NewServer(f.handler())
	defer srv.Close()
	dir := Direcciones{WSAA: srv.URL + "/wsaa", WSFE: srv.URL + "/wsfe"}
	cert, clave := certPrueba(t)
	der, _ := x509.MarshalPKCS8PrivateKey(clave)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	clavePEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	ahora := time.Date(2026, 10, 23, 15, 0, 0, 0, arg)

	tk, err := PedirTicket(ctx, srv.Client(), dir, certPEM, clavePEM, "wsfe", ahora)
	if err != nil || tk.Token != "TOKEN&1" || tk.Sign != "FIRMA" || !tk.Vigente(ahora) {
		t.Fatalf("ticket: %+v %v", tk, err)
	}
	// Pedir otro con uno vigente: ARCA contesta con un fault, que llega con su código.
	_, err = PedirTicket(ctx, srv.Client(), dir, certPEM, clavePEM, "wsfe", ahora)
	if e, ok := err.(ErrARCA); !ok || !strings.Contains(e.Codigo, "alreadyAuthenticated") {
		t.Fatalf("segundo ticket: %v", err)
	}

	ult, err := UltimoAutorizado(ctx, srv.Client(), dir, tk, "20448592341", 3, FacturaB)
	if err != nil || ult != 41 {
		t.Fatalf("último: %d %v", ult, err)
	}
	c := Comprobante{Tipo: FacturaB, PuntoVenta: 3, DocTipo: DocConsumidorFin, CondReceptor: CondConsumidorFinal, Fecha: ahora, Total: 12100}
	a, err := SolicitarCAE(ctx, srv.Client(), dir, tk, "20448592341", c, ult+1)
	if err != nil || a.Numero != 42 || a.CAE == "" || a.VenceCAE.Day() != 2 {
		t.Fatalf("CAE: %+v %v", a, err)
	}
	// Lo que salió por el cable: pesos con dos decimales, neto + IVA 21 % que cierran, tipo 6.
	for _, want := range []string{"<ar:ImpTotal>12100.00</ar:ImpTotal>", "<ar:ImpNeto>10000.00</ar:ImpNeto>", "<ar:ImpIVA>2100.00</ar:ImpIVA>",
		"<ar:CbteTipo>6</ar:CbteTipo>", "<ar:CbteFch>20261023</ar:CbteFch>", "<ar:CondicionIVAReceptorId>5</ar:CondicionIVAReceptorId>"} {
		if !strings.Contains(f.ultimoReq, want) {
			t.Errorf("falta %s en el pedido", want)
		}
	}

	t.Run("redondeo al IVA: la suma cierra", func(t *testing.T) {
		for _, total := range []int64{1, 99, 999, 12345, 1999999} {
			n, i, tot := Comprobante{Tipo: FacturaB, Total: total}.importes()
			if n+i != tot {
				t.Fatalf("%d: %d + %d != %d", total, n, i, tot)
			}
		}
	})
	t.Run("número salteado: rechazo con el código de ARCA", func(t *testing.T) {
		_, err := SolicitarCAE(ctx, srv.Client(), dir, tk, "20448592341", c, 99)
		if r, ok := err.(Rechazado); !ok || r.Codigo != "10016" {
			t.Fatalf("salteado: %v", err)
		}
	})
	t.Run("factura C sin IVA", func(t *testing.T) {
		cc := c
		cc.Tipo = FacturaC
		f.ultimo[FacturaC] = 0
		if _, err := SolicitarCAE(ctx, srv.Client(), dir, tk, "20448592341", cc, 1); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(f.ultimoReq, "<ar:ImpIVA>0.00</ar:ImpIVA>") || strings.Contains(f.ultimoReq, "<ar:Iva>") {
			t.Fatal("la C mandó IVA")
		}
	})
	t.Run("validaciones antes de llamar", func(t *testing.T) {
		a := c
		a.Tipo, a.DocTipo, a.DocNro = FacturaA, DocCUIT, 20448592342
		if _, err := SolicitarCAE(ctx, srv.Client(), dir, tk, "20448592341", a, 1); err == nil || strings.Contains(err.Error(), "rechazado") {
			t.Fatalf("A con CUIT inválido: %v", err)
		}
		nc := c
		nc.Tipo = NotaCreditoB
		if err := nc.Validar(); err == nil {
			t.Fatal("NC sin asociado")
		}
	})
	t.Run("token inválido: error de ARCA", func(t *testing.T) {
		_, err := UltimoAutorizado(ctx, srv.Client(), dir, Ticket{Token: "x", Sign: "y"}, "20448592341", 3, FacturaB)
		if e, ok := err.(ErrARCA); !ok || e.Codigo != "600" {
			t.Fatalf("token: %v", err)
		}
	})
}
