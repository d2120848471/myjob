package ipx

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

type RegionResolver interface {
	Resolve(ip string) string
}

type emptyResolver struct{}

func (emptyResolver) Resolve(string) string {
	return ""
}

type xdbResolver struct {
	mu       sync.Mutex
	searcher *xdb.Searcher
}

func DefaultDBPaths() []string {
	return []string{
		"resource/ipdb/ip_region.xdb",
		filepath.Join("..", "..", "resource", "ipdb", "ip_region.xdb"),
	}
}

func NewRegionResolver(dbPaths ...string) RegionResolver {
	for _, path := range dbPaths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		if _, err := os.Stat(path); err != nil {
			continue
		}
		content, err := xdb.LoadContentFromFile(path)
		if err != nil {
			continue
		}
		header, err := xdb.LoadHeaderFromBuff(content)
		if err != nil {
			continue
		}
		version, err := xdb.VersionFromHeader(header)
		if err != nil {
			continue
		}
		searcher, err := xdb.NewWithBuffer(version, content)
		if err != nil {
			continue
		}
		return &xdbResolver{searcher: searcher}
	}
	return emptyResolver{}
}

func (r *xdbResolver) Resolve(ip string) string {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	region, err := r.searcher.Search(parsed.String())
	if err != nil {
		return ""
	}
	return normalizeRegion(region)
}

func normalizeRegion(region string) string {
	parts := strings.Split(strings.TrimSpace(region), "|")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "0" {
			continue
		}
		cleaned = append(cleaned, part)
	}
	return strings.Join(cleaned, " ")
}
