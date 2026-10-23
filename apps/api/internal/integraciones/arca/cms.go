// Package arca: facturación electrónica con ARCA (ex AFIP). WSAA para el ticket de acceso (con el
// certificado de la tienda), WSFEv1 para pedir el CAE, y la clave y el pedido de certificado (CSR)
// que la tienda sube a ARCA. Todo por SOAP con el cliente HTTP de la biblioteca estándar.
package arca

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"math/big"
	"sort"
	"time"
)

// El WSAA pide el pedido de acceso firmado como CMS (PKCS#7 SignedData) con el contenido adentro.
// Go no trae PKCS#7: se arma a mano con encoding/asn1. Lo verifica `openssl cms -verify` en los tests.
// Tres detalles que no dan ningún error legible si están mal:
//   - la firma va sobre los atributos codificados como SET OF (tag 0x31), no sobre el [0] IMPLICIT
//     con el que viajan;
//   - los atributos del SET OF van ordenados por su codificación DER;
//   - el emisor y el serial del certificado viajan crudos (RawIssuer), sin reconstruirlos.

var (
	oidData          = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
	oidSignedData    = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	oidContentType   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}
	oidMessageDigest = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}
	oidSigningTime   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}
	oidSHA256        = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
	oidRSA           = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
)

type algoritmo struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

var (
	sha256Alg = algoritmo{Algorithm: oidSHA256, Parameters: asn1.NullRawValue}
	rsaAlg    = algoritmo{Algorithm: oidRSA, Parameters: asn1.NullRawValue}
)

func mustDER(v any) []byte {
	b, err := asn1.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// atributo codifica un Attribute ::= SEQUENCE { type OID, values SET OF ANY } con un solo valor.
func atributo(oid asn1.ObjectIdentifier, valor []byte) []byte {
	set := asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: valor}
	return mustDER(struct {
		Type   asn1.ObjectIdentifier
		Values asn1.RawValue
	}{oid, set})
}

// setOrdenado arma el contenido de un SET OF con sus elementos ordenados por codificación (DER).
func setOrdenado(elems [][]byte) []byte {
	ord := append([][]byte(nil), elems...)
	sort.Slice(ord, func(i, j int) bool { return bytes.Compare(ord[i], ord[j]) < 0 })
	return bytes.Join(ord, nil)
}

// FirmarCMS firma el contenido con la clave y el certificado de la tienda y devuelve el CMS en DER.
func FirmarCMS(contenido []byte, cert *x509.Certificate, clave *rsa.PrivateKey, ahora time.Time) ([]byte, error) {
	if cert == nil || clave == nil {
		return nil, errors.New("arca: falta el certificado o la clave")
	}
	resumen := sha256.Sum256(contenido)
	attrs := setOrdenado([][]byte{
		atributo(oidContentType, mustDER(oidData)),
		atributo(oidSigningTime, mustDER(ahora.UTC())),
		atributo(oidMessageDigest, mustDER(resumen[:])),
	})
	// Lo que se firma: los atributos como SET OF.
	firmado := mustDER(asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: attrs})
	h := sha256.Sum256(firmado)
	firma, err := rsa.SignPKCS1v15(rand.Reader, clave, crypto.SHA256, h[:])
	if err != nil {
		return nil, err
	}
	signerInfo := mustDER(struct {
		Version   int
		SID       asn1.RawValue
		Digest    algoritmo
		Attrs     asn1.RawValue
		SigAlg    algoritmo
		Signature []byte
	}{
		Version: 1,
		SID: asn1.RawValue{FullBytes: mustDER(struct {
			Issuer asn1.RawValue
			Serial *big.Int
		}{asn1.RawValue{FullBytes: cert.RawIssuer}, cert.SerialNumber})},
		Digest: sha256Alg,
		// Viajan como [0] IMPLICIT.
		Attrs:     asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: attrs},
		SigAlg:    rsaAlg,
		Signature: firma,
	})
	signedData := mustDER(struct {
		Version          int
		DigestAlgorithms asn1.RawValue
		Encap            asn1.RawValue
		Certificates     asn1.RawValue
		SignerInfos      asn1.RawValue
	}{
		Version:          1,
		DigestAlgorithms: asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: mustDER(sha256Alg)},
		Encap: asn1.RawValue{FullBytes: mustDER(struct {
			Type    asn1.ObjectIdentifier
			Content asn1.RawValue
		}{oidData, asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: mustDER(contenido)}})},
		Certificates: asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: cert.Raw},
		SignerInfos:  asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSet, IsCompound: true, Bytes: signerInfo},
	})
	return asn1.Marshal(struct {
		Type    asn1.ObjectIdentifier
		Content asn1.RawValue
	}{oidSignedData, asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: signedData}})
}
