package torbit

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"math/big"
	"time"
)

func makeSelfSignedCert() (*tls.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{Organization: []string{"Chat Self-Signed Dev Cert"}},
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

// DefaultTLSConfig is the default config used for serving content with TLS,
// such as in the HTTPS server.
func DefaultTLSConfig() *tls.Config {
	cert, err := makeSelfSignedCert()
	if err != nil {
		// You'll and obvious error if the nil config is returned, so for simplicity
		// sake, just return nil here. In a real app, this would be a horrible idea.
		log.Printf("Unable to generate a self signed cert: %s\n", err.Error())
		return nil
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
