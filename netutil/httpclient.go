package netutil

import (
	"net/http"
	"time"
)

// TunedHTTPTransport returns a cloned http.Transport tuned for high-QPS
// traffic against the same remote host (e.g. an upstream gateway called on
// every request): the default Transport keeps only MaxIdleConnsPerHost=2,
// so idle connections get torn down in bulk between bursts and the next
// wave of requests pays TCP+TLS handshakes again (1-2 extra RTT each).
func TunedHTTPTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxIdleConnsPerHost = 16
	t.IdleConnTimeout = 90 * time.Second
	return t
}
