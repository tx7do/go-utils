package notify

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubService struct {
	name string
}

func (s *stubService) Name() string { return s.name }
func (s *stubService) Send(_ context.Context, _ *Message) (*Receipt, error) {
	return &Receipt{}, nil
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	_, ok := r.Service("email")
	require.False(t, ok, "未注册的渠道取不到")

	r.Register(&stubService{name: "email"})
	r.Register(&stubService{name: "webhook"})

	got, ok := r.Service("email")
	require.True(t, ok)
	require.Equal(t, "email", got.Name())

	// Register 覆盖同渠道旧实现：装配期一次性写入，运行期只读。
	r.Register(&stubService{name: "email"})
	got, ok = r.Service("email")
	require.True(t, ok)
	require.Equal(t, "email", got.Name())
}
