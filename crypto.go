package torbit

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

func makeSelfSignedCert() (*tls.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err // might be good to add a message to the error
		// mentioning being unable to gen a serial no
	}
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{Organization: []string{"Torbit Go Programming Challenge Self-Signed"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour * 31), // a month-ish
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// ecdsa256 has a constant-time assembly implementation to prevent timing attacks
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	return &tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privateKey,
		Leaf:        cert,
	}, err
}

// DefaultTLSConfig is the default blah
func DefaultTLSConfig() *tls.Config {
	cert, err := makeSelfSignedCert()
	if err != nil {
		println(err.Error())
		return nil // DONT DO THIS
	}

	pool := x509.NewCertPool()
	pool.AddCert(cert.Leaf)

	return &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
		},
		Certificates: []tls.Certificate{*cert},
		ClientCAs:    pool,
	}
}
