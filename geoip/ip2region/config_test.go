package ip2region

import (
	"fmt"
	"testing"

	"github.com/tx7do/go-utils/geoip/ip2region/assets"
)

func TestV4Config(t *testing.T) {
	v4Config, err := NewV4Config(VIndexCache, assets.Ip2RegionV4, 10)
	if err != nil {
		t.Errorf("failed to new v4 config: %s", err)
		return
	}

	v4BufferConfig, err := NewV4Config(BufferCache, assets.Ip2RegionV4, 10)
	if err != nil {
		t.Errorf("failed to new v4 config: %s", err)
		return
	}

	fmt.Printf("v4Config: %s\n", v4Config)
	fmt.Printf("v4BufferConfig: %s\n", v4BufferConfig)
}

func TestV6Config(t *testing.T) {
	v6Config, err := NewV6Config(NoCache, assets.Ip2RegionV6, 10)
	if err != nil {
		t.Errorf("failed to new v6 config: %s", err)
		return
	}

	fmt.Printf("v6Config: %s\n", v6Config)
}
