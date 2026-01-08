package assets

import _ "embed"

//go:embed ip2region_v4.xdb
var Ip2RegionV4 []byte

//go:embed ip2region_v6.xdb
var Ip2RegionV6 []byte
