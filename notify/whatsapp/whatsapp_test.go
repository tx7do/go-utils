package whatsapp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tx7do/go-utils/notify"
)

func TestSendDeliversCloudAPIShape(t *testing.T) {
	var (
		auth    string
		path    string
		payload map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		auth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		_, _ = w.Write([]byte(`{"messaging_product":"whatsapp","contacts":[],"messages":[{"id":"wamid.XXX"}]}`))
	}))
	t.Cleanup(srv.Close)

	svc := New(srv.Client(), Config{Ref: "7", PhoneNumberID: "1065403", AccessToken: "eaag",
		APIVersion: "v21.0", APIBase: srv.URL})
	receipt, err := svc.Send(context.Background(), &notify.Message{
		Target:  "8613800138000",
		Title:   "重置验证码",
		Content: "您的验证码是 123456",
	})
	require.NoError(t, err)
	require.Equal(t, "wamid.XXX", receipt.ProviderID, "wamid 要进回执")
	require.Equal(t, "/v21.0/1065403/messages", path, "版本段与号码 ID 进路径")
	require.Equal(t, "Bearer eaag", auth)

	require.Equal(t, "whatsapp", payload["messaging_product"])
	require.Equal(t, "individual", payload["recipient_type"])
	require.Equal(t, "8613800138000", payload["to"])
	text := payload["text"].(map[string]any)
	require.Equal(t, "重置验证码\n您的验证码是 123456", text["body"])
}

func TestSendConfigIssue(t *testing.T) {
	svc := New(nil, Config{PhoneNumberID: "1"})
	_, err := svc.Send(context.Background(), &notify.Message{Target: "8613800138000", Content: "c"})
	require.ErrorIs(t, err, notify.ErrNotConfigured)
}
