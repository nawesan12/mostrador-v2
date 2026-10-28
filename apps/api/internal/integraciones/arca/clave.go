package arca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"regexp"
	"strings"
	"time"
)

// La tienda no maneja claves: Mostrador genera la clave privada (queda cifrada en la base y nunca
// sale), arma el pedido de certificado (CSR) para que la tienda lo suba en ARCA ("Administración de
// certificados digitales", o WSASS en homologación), y después la tienda pega el certificado que
// ARCA le devuelve.

var cuitValido = regexp.MustCompile(`^\d{11}$`)

// CUITValido verifica los 11 dígitos y el dígito verificador.
func CUITValido(cuit string) bool {
	if !cuitValido.MatchString(cuit) {
		return false
	}
	pesos := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	s := 0
	for i, p := range pesos {
		s += int(cuit[i]-'0') * p
	}
	v := 11 - s%11
	switch v {
	case 11:
		v = 0
	case 10:
		v = 9
	}
	return int(cuit[10]-'0') == v
}

// NuevaClave genera la clave RSA de 2048 bits (lo que pide ARCA) en PEM PKCS#8.
func NuevaClave() ([]byte, error) {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	der, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}

func leerClave(p []byte) (*rsa.PrivateKey, error) {
	b, _ := pem.Decode(p)
	if b == nil {
		return nil, errors.New("arca: la clave no es PEM")
	}
	k, err := x509.ParsePKCS8PrivateKey(b.Bytes)
	if err != nil {
		if k1, err1 := x509.ParsePKCS1PrivateKey(b.Bytes); err1 == nil {
			return k1, nil
		}
		return nil, err
	}
	r, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("arca: la clave no es RSA")
	}
	return r, nil
}

// PedidoCertificado arma el CSR con el sujeto que pide ARCA: C=AR, O=razón social, CN=alias,
// serialNumber="CUIT <cuit>".
func PedidoCertificado(clavePEM []byte, cuit, razonSocial, alias string) ([]byte, error) {
	if !CUITValido(cuit) {
		return nil, errors.New("arca: CUIT inválido")
	}
	k, err := leerClave(clavePEM)
	if err != nil {
		return nil, err
	}
	razonSocial = strings.TrimSpace(razonSocial)
	if razonSocial == "" || len(razonSocial) > 64 {
		return nil, errors.New("arca: razón social inválida")
	}
	sujeto := pkix.Name{Country: []string{"AR"}, Organization: []string{razonSocial}, CommonName: alias,
		ExtraNames: []pkix.AttributeTypeAndValue{{Type: asn1.ObjectIdentifier{2, 5, 4, 5}, Value: "CUIT " + cuit}}}
	der, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{Subject: sujeto,
		SignatureAlgorithm: x509.SHA256WithRSA}, k)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der}), nil
}

// DatosCertificado es lo que se muestra del certificado que pegó la tienda.
type DatosCertificado struct {
	CUIT    string    `json:"cuit"`
	Alias   string    `json:"alias"`
	Desde   time.Time `json:"desde"`
	Vence   time.Time `json:"vence"`
	Emisor  string    `json:"emisor"`
	Homolog bool      `json:"homologacion"`
}

var (
	ErrCertificado        = errors.New("arca: el certificado no se puede leer")
	ErrCertificadoAjeno   = errors.New("arca: el certificado no corresponde a la clave")
	ErrCertificadoVencido = errors.New("arca: el certificado venció")
)

// RevisarCertificado lee el PEM que pegó la tienda y comprueba que sea de su clave y de su CUIT.
func RevisarCertificado(certPEM, clavePEM []byte, cuit string, ahora time.Time) (DatosCertificado, error) {
	var d DatosCertificado
	b, _ := pem.Decode(certPEM)
	if b == nil || b.Type != "CERTIFICATE" {
		return d, ErrCertificado
	}
	c, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		return d, ErrCertificado
	}
	k, err := leerClave(clavePEM)
	if err != nil {
		return d, err
	}
	pub, ok := c.PublicKey.(*rsa.PublicKey)
	if !ok || !pub.Equal(&k.PublicKey) {
		return d, ErrCertificadoAjeno
	}
	d = DatosCertificado{Alias: c.Subject.CommonName, Desde: c.NotBefore, Vence: c.NotAfter, Emisor: c.Issuer.CommonName,
		Homolog: strings.Contains(strings.ToLower(c.Issuer.CommonName), "test") || strings.Contains(strings.ToLower(c.Issuer.CommonName), "homo")}
	for _, n := range c.Subject.Names {
		if n.Type.Equal(asn1.ObjectIdentifier{2, 5, 4, 5}) {
			if v, ok := n.Value.(string); ok {
				d.CUIT = strings.TrimSpace(strings.TrimPrefix(v, "CUIT"))
			}
		}
	}
	if d.CUIT != cuit {
		return d, ErrCertificadoAjeno
	}
	if ahora.After(c.NotAfter) {
		return d, ErrCertificadoVencido
	}
	return d, nil
}

func leerCert(certPEM []byte) (*x509.Certificate, error) {
	b, _ := pem.Decode(certPEM)
	if b == nil {
		return nil, ErrCertificado
	}
	return x509.ParseCertificate(b.Bytes)
}
