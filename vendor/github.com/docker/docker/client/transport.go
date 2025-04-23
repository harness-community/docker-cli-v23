package client // import "github.com/DevanshMathur19/docker-v23/client"

import (
	"crypto/tls"
	"net/http"
)

// resolveTLSConfig attempts to resolve the TLS configuration from the
// RoundTripper.
func resolveTLSConfig(transport http.RoundTripper) *tls.Config {
	switch tr := transport.(type) {
	case *http.Transport:
		return tr.TLSClientConfig
	default:
		return nil
	}
}
