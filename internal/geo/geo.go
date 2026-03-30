package geo

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type IP struct {
	db *geoip2.Reader
}

func NewGeoIP(path string) (*IP, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open geoip database: %w", err)
	}
	return &IP{db: db}, nil
}

func (g *IP) Country(ip string) string {
	record, err := g.db.Country(net.ParseIP(ip))
	if err != nil || record.Country.IsoCode == "" {
		return "Unknown"
	}
	return record.Country.IsoCode
}

func (g *IP) Close() error {
	return g.db.Close()
}
