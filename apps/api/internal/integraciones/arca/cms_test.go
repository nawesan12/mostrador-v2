package arca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// certPrueba arma un certificado autofirmado con un sujeto como el de ARCA (serialNumber "CUIT …").
func certPrueba(t testing.TB) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	sujeto := pkix.Name{Country: []string{"AR"}, Organization: []string{"Tienda de Prueba"}, CommonName: "mostrador",
		ExtraNames: []pkix.AttributeTypeAndValue{{Type: asn1.ObjectIdentifier{2, 5, 4, 5}, Value: "CUIT 20448592341"}}}
	tpl := &x509.Certificate{SerialNumber: big.NewInt(0x1234567890), Subject: sujeto, Issuer: sujeto,
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &k.PublicKey, k)
	if err != nil {
		t.Fatal(err)
	}
	c, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return c, k
}

// openssl es un verificador que no escribí yo: si acepta la firma, el DER está bien armado y la
// firma corresponde al contenido (no prueba que ARCA la acepte, pero cubre lo que más se rompe).
func TestOpensslVerificaLaFirma(t *testing.T) {
	if _, err := exec.LookPath("openssl"); err != nil {
		t.Skip("sin openssl")
	}
	cert, clave := certPrueba(t)
	contenido := []byte(`<?xml version="1.0" encoding="UTF-8"?><loginTicketRequest version="1.0"><header><uniqueId>1</uniqueId></header><service>wsfe</service></loginTicketRequest>`)
	cms, err := FirmarCMS(contenido, cert, clave, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "firma.der"), cms, 0o600)
	_ = os.WriteFile(filepath.Join(dir, "cert.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}), 0o600)
	out, err := exec.Command("openssl", "cms", "-verify", "-inform", "DER", "-in", filepath.Join(dir, "firma.der"),
		"-CAfile", filepath.Join(dir, "cert.pem"), "-purpose", "any").CombinedOutput()
	if err != nil {
		t.Fatalf("openssl no la acepta: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), string(contenido)) {
		t.Fatalf("el contenido no viaja adentro: %q", out)
	}

	// Y si se toca un byte del contenido firmado, openssl la rechaza (el test no es de adorno).
	roto := append([]byte(nil), cms...)
	i := indice(roto, []byte("wsfe"))
	roto[i] = 'X'
	_ = os.WriteFile(filepath.Join(dir, "rota.der"), roto, 0o600)
	if err := exec.Command("openssl", "cms", "-verify", "-inform", "DER", "-in", filepath.Join(dir, "rota.der"),
		"-CAfile", filepath.Join(dir, "cert.pem"), "-purpose", "any").Run(); err == nil {
		t.Fatal("openssl aceptó un contenido modificado")
	}
}

func indice(b, sub []byte) int {
	for i := 0; i+len(sub) <= len(b); i++ {
		if string(b[i:i+len(sub)]) == string(sub) {
			return i
		}
	}
	return -1
}
