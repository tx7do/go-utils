package netutil

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// TunedHTTPTransport
// ============================================================

func TestTunedHTTPTransport_ConnectPoolTuning(t *testing.T) {
	tr := TunedHTTPTransport()

	assert.Equal(t, 100, tr.MaxIdleConns, "total idle-connection cap")
	assert.Equal(t, 16, tr.MaxIdleConnsPerHost, "per-host idle cap (default 2 is what forces TCP rebuilds)")
	assert.Equal(t, 90*time.Second, tr.IdleConnTimeout)
}

// ============================================================
// TuneRestServerTimeouts
// ============================================================

func TestTuneRestServerTimeouts(t *testing.T) {
	srv := khttp.NewServer()
	TuneRestServerTimeouts(srv)

	assert.Equal(t, 10*time.Second, srv.ReadHeaderTimeout, "slowloris line")
	assert.Equal(t, 120*time.Second, srv.IdleTimeout, "keep-alive reclamation")

	// Read/Write timeouts are deliberately left unset: REST frontends carry
	// SSE long-lived event streams, and deadlines would cut the streams off.
	assert.Zero(t, srv.ReadTimeout, "ReadTimeout must stay unset (SSE)")
	assert.Zero(t, srv.WriteTimeout, "WriteTimeout must stay unset (SSE)")
}

func TestTuneRestServerTimeouts_NilSafe(t *testing.T) {
	assert.NotPanics(t, func() { TuneRestServerTimeouts(nil) })
}

// ============================================================
// HeaderFromContext (kratos HTTP transport)
// ============================================================

func TestHeaderFromContext_HTTPTransport(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://test/captcha", nil)
	require.NoError(t, err)
	req.Header.Set("X-Captcha-Id", "cap-7")

	ctx := transport.NewServerContext(context.Background(), &stubHTTPTransporter{req: req})

	h := HeaderFromContext(ctx)
	require.NotNil(t, h)
	assert.Equal(t, "cap-7", h.Get("X-Captcha-Id"))
}

func TestHeaderFromContext_NoTransport(t *testing.T) {
	assert.Nil(t, HeaderFromContext(context.Background()), "bare context returns nil")
}

func TestHeaderFromContext_RequestNil(t *testing.T) {
	// transporter present but request not ready → nil
	ctx := transport.NewServerContext(context.Background(), &stubHTTPTransporter{req: nil})
	assert.Nil(t, HeaderFromContext(ctx))
}
