package ip2region

import (
	"fmt"
	"testing"

	"github.com/tx7do/go-utils/geoip/ip2region/assets"
)

func TestV4SearcherPool(t *testing.T) {
	v4Config, err := NewV4Config(VIndexCache, assets.Ip2RegionV4, 5)
	if err != nil {
		t.Fatalf("failed to new v4 config: %s", err)
	}

	searcherPool, err := NewSearcherPool(v4Config)
	if err != nil {
		t.Fatalf("failed to create searcher pool: %s", err)
	}

	ipString := "219.133.110.197"
	for i := 0; i < 20; i++ {
		searcher := searcherPool.BorrowSearcher()
		region, err := searcher.SearchByStr(ipString)
		if err != nil {
			t.Fatalf("failed to search(%s): %s", ipString, err)
		}

		fmt.Printf("%2d->search(%s)=%s\n", i, ipString, region)
		searcherPool.ReturnSearcher(searcher)
	}

	// borrow one at last for Close timeout wait testing ONLY
	// searcherPool.BorrowSearcher()

	// close the searcher pool
	searcherPool.Close()
}

func TestV6SearcherPool(t *testing.T) {
	v6Config, err := NewV6Config(VIndexCache, assets.Ip2RegionV6, 5)
	if err != nil {
		t.Fatalf("failed to new v6 config: %s", err)
	}

	searcherPool, err := NewSearcherPool(v6Config)
	if err != nil {
		t.Fatalf("failed to create searcher pool: %s", err)
	}

	ipString := "240e:3b7:3275:f090:d2a3:7d1a:dd90:c3b6"
	for i := 0; i < 20; i++ {
		searcher := searcherPool.BorrowSearcher()
		region, err := searcher.SearchByStr(ipString)
		if err != nil {
			t.Fatalf("failed to search(%s): %s", ipString, err)
		}

		fmt.Printf("%2d->search(%s)=%s\n", i, ipString, region)
		searcherPool.ReturnSearcher(searcher)
	}

	// borrow one at last for Close timeout wait testing ONLY
	// searcherPool.BorrowSearcher()

	// close the searcher pool
	searcherPool.Close()
}
