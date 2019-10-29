package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
)

func listenAndServeMutualTLS(addr, clientCert, myCert, myKey string) error {
	// Create a CA certificate pool and add cert.pem to it
	caCert, err := ioutil.ReadFile(clientCert)
	if err != nil {
		return err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create the TLS Config with the CA pool and enable Client certificate validation
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	// Create a Server instance with the TLS config
	return (&http.Server{
		Addr:      addr,
		TLSConfig: tlsConfig,
	}).ListenAndServeTLS(myCert, myKey)
}
