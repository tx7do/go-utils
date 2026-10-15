package postmark

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversTokenHeaderAndShape(t *testing.T) {
	var (
		token   string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/email", r.URL.Path)
		token = r.Header.Get("X-Postmark-Server-Token")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"To":"alice@example.com","SubmittedAt":"2026-10-15","MessageID":"1234567","ErrorCode":0,"Message":"OK"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", ServerToken: "st-1",
		From: "noreply@example.com", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "alice@example.com",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "1234567", receipt.ProviderID, "Postmark 的 MessageID 要进回执")
	require.Equal(t, "st-1", token)

	require.Equal(t, "noreply@example.com", payload["From"])
	require.Equal(t, "alice@example.com", payload["To"])
	require.Equal(t, "重置验证码", payload["Subject"])
	require.Equal(t, "您的验证码是 123456", payload["TextBody"])
}

func TestSendPostmarkErrorInside200(t *testing.T) {
	// Postmark 的业务错误也是 HTTP 200 + ErrorCode!=0，只看状态码会假成功。
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ErrorCode":401,"Message":"Invalid API Request"}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{ServerToken: "bad", From: "f@e.c", APIBase: srv.URL})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "a@b.c", Content: "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "error_code=401")
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "a@b.c", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
