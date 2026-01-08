package ip2region

import (
	"fmt"
	"strings"

	"github.com/tx7do/go-utils/geoip"
	"github.com/tx7do/go-utils/geoip/ip2region/assets"
)

type Client struct {
	ip2region *Ip2Region
}

func NewClient() (*Client, error) {
	v4Config, err := NewV4Config(VIndexCache, assets.Ip2RegionV4, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to create v4 config: %s", err)
	}

	v6Config, err := NewV6Config(VIndexCache, assets.Ip2RegionV6, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to create v6 config: %s", err)
	}

	ip2region, err := NewIp2Region(v4Config, v6Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create ip2region service: %s", err)
	}

	return &Client{
		ip2region: ip2region,
	}, nil
}

func (g *Client) Close() {
	g.ip2region.Close()
}

func (g *Client) Query(rawIP string) (ret geoip.Result, err error) {
	regionData, err := g.ip2region.SearchByStr(rawIP)
	if err != nil {
		return ret, err
	}

	parts := strings.Split(regionData, "|")
	if len(parts) != 4 {
		return ret, fmt.Errorf("invalid region data: %s", regionData)
	}

	ret.Country = parts[0]
	ret.Province = parts[1]
	ret.City = parts[2]
	ret.ISP = parts[3]

	return ret, nil
}
