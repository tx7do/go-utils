package geolite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoLite(t *testing.T) {
	g, err := NewClient()
	assert.Nil(t, err)

	ret, err := g.Query("47.108.149.89")
	assert.Nil(t, err)
	assert.Equal(t, ret.Country, "中国")
	assert.Equal(t, ret.Province, "四川省")
	assert.Equal(t, ret.City, "成都")

	ret, err = g.Query("::1")
	assert.Nil(t, err)

	ret, err = g.Query("127.0.0.1")
	assert.Nil(t, err)
}
