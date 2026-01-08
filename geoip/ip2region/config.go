package ip2region

import (
	"fmt"
	"strings"

	"github.com/tx7do/go-utils/geoip/ip2region/xdb"
)

const (
	NoCache     = 0
	VIndexCache = 1
	BufferCache = 2
)

func CachePolicyFromName(name string) (int, error) {
	switch strings.ToLower(name) {
	case "file", "nocache":
		return NoCache, nil
	case "vectorindex", "vindex", "vindexcache":
		return VIndexCache, nil
	case "content", "buffercache":
		return BufferCache, nil
	default:
		return NoCache, fmt.Errorf("invalid cache policy name `%s`", name)
	}
}

type Config struct {
	cachePolicy int
	ipVersion   *xdb.Version

	header *xdb.Header

	// buffers
	vIndex  []byte
	cBuffer []byte

	searchers int
}

func NewV4Config(cachePolicy int, xdbContent []byte, searchers int) (*Config, error) {
	return newConfig(cachePolicy, xdb.IPv4, xdbContent, searchers)
}

func NewV6Config(cachePolicy int, xdbContent []byte, searchers int) (*Config, error) {
	return newConfig(cachePolicy, xdb.IPv6, xdbContent, searchers)
}

func newConfig(cachePolicy int, ipVersion *xdb.Version, xdbContent []byte, searchers int) (*Config, error) {
	if searchers < 1 {
		return nil, fmt.Errorf("searchers=%d, > 0 expected", searchers)
	}

	header, err := xdb.LoadHeaderFromBuff(xdbContent)
	if err != nil {
		return nil, err
	}

	// verify the ip version
	xIpVersion, err := xdb.VersionFromHeader(header)
	if err != nil {
		return nil, err
	}

	if xIpVersion.Id != ipVersion.Id {
		return nil, fmt.Errorf("ip version mismatch, xdb version=%s, expected=%s", xIpVersion.String(), ipVersion.String())
	}

	// 3, check and load the vector index buffer
	var vIndex []byte = nil
	if cachePolicy == VIndexCache {
		vIndex, err = xdb.LoadVectorIndexFromBuff(xdbContent)
		if err != nil {
			return nil, err
		}
	}

	return &Config{
		cachePolicy: cachePolicy,
		ipVersion:   ipVersion,

		header: header,

		vIndex:  vIndex,
		cBuffer: xdbContent,

		searchers: searchers,
	}, nil
}

func (c *Config) String() string {
	vIndex := "null"
	if c.vIndex != nil {
		vIndex = fmt.Sprintf("{bytes:%d}", len(c.vIndex))
	}

	cBuffer := "null"
	if c.cBuffer != nil {
		cBuffer = fmt.Sprintf("{bytes:%d}", len(c.cBuffer))
	}

	return fmt.Sprintf(
		"{cache_policy:%d, version:%s, header:%s, v_index:%s, c_buffer:%s}",
		c.cachePolicy, c.ipVersion.String(), c.header.String(), vIndex, cBuffer,
	)
}

func (c *Config) CachePolicy() int {
	return c.cachePolicy
}

func (c *Config) IPVersion() *xdb.Version {
	return c.ipVersion
}

func (c *Config) Header() *xdb.Header {
	return c.header
}

func (c *Config) VIndex() []byte {
	return c.vIndex
}

func (c *Config) CBuffer() []byte {
	return c.cBuffer
}

func (c *Config) Searchers() int {
	return c.searchers
}
