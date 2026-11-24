package netutil

import (
	"time"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const (
	// restReadHeaderTimeout caps request-header reading — the slowloris line.
	restReadHeaderTimeout = 10 * time.Second
	// restIdleTimeout reclaims idle keep-alive connections so they cannot
	// pile up indefinitely.
	restIdleTimeout = 120 * time.Second
)

// TuneRestServerTimeouts adds network-layer timeouts to a Kratos REST
// server. Servers created by kratos (and kratos-bootstrap) carry zero
// timeouts by default.
//
// ReadTimeout/WriteTimeout are deliberately NOT set: REST frontends also
// carry SSE long-lived event streams, and read/write deadlines would cut
// those streams off mid-flight. Only the slowloris surface (header read)
// and idle-connection pile-up are closed here.
func TuneRestServerTimeouts(srv *khttp.Server) {
	if srv == nil {
		return
	}
	srv.ReadHeaderTimeout = restReadHeaderTimeout
	srv.IdleTimeout = restIdleTimeout
}
