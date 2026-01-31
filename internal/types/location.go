package types

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
)

// GeoPoint represents a PostGIS geography Point (WGS84) and is the single source of truth for location.
// JSON and form use "latitude" and "longitude" for API compatibility.
type GeoPoint struct {
	Lat decimal.Decimal `json:"latitude" form:"latitude"`
	Lng decimal.Decimal `json:"longitude" form:"longitude"`
}

// WKT point pattern: POINT(lng lat) with optional SRID prefix and whitespace.
var wktPointRegex = regexp.MustCompile(`(?i)POINT\s*\(\s*([-\d.eE+]+)\s+([-\d.eE+]+)\s*\)`)

// Value implements driver.Valuer. Returns PostGIS WKT format: "POINT(lng lat)".
func (p GeoPoint) Value() (driver.Value, error) {
	if p.Lat.IsZero() && p.Lng.IsZero() {
		return nil, nil
	}
	return fmt.Sprintf("POINT(%f %f)", p.Lng.InexactFloat64(), p.Lat.InexactFloat64()), nil
}

// Scan implements sql.Scanner. Handles PostGIS output: WKT string, EWKB binary, or hex-encoded EWKB.
func (p *GeoPoint) Scan(src interface{}) error {
	if src == nil {
		p.Lng = decimal.Zero
		p.Lat = decimal.Zero
		return nil
	}
	switch v := src.(type) {
	case string:
		if err := p.scanWKT(v); err == nil {
			return nil
		}
		// Try hex-encoded EWKB (e.g. from pgx).
		decoded, err := hex.DecodeString(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("failed to parse GeoPoint (not WKT or hex EWKB): %w", err)
		}
		return p.scanEWKB(decoded)
	case []byte:
		if len(v) == 0 {
			p.Lng = decimal.Zero
			p.Lat = decimal.Zero
			return nil
		}
		if err := p.scanWKT(string(v)); err == nil {
			return nil
		}
		// Driver may return hex-encoded EWKB as []byte (e.g. lib/pq, pgx).
		decoded, err := hex.DecodeString(strings.TrimSpace(string(v)))
		if err == nil {
			return p.scanEWKB(decoded)
		}
		return p.scanEWKB(v)
	default:
		return fmt.Errorf("unsupported type for GeoPoint: %T", src)
	}
}

func (p *GeoPoint) scanWKT(wkt string) error {
	if idx := strings.Index(wkt, ";"); idx != -1 {
		wkt = wkt[idx+1:]
	}
	matches := wktPointRegex.FindStringSubmatch(strings.TrimSpace(wkt))
	if len(matches) != 3 {
		return fmt.Errorf("input does not match format")
	}
	var lng, lat float64
	if _, err := fmt.Sscanf(matches[1], "%f", &lng); err != nil {
		return err
	}
	if _, err := fmt.Sscanf(matches[2], "%f", &lat); err != nil {
		return err
	}
	p.Lng = decimal.NewFromFloat(lng)
	p.Lat = decimal.NewFromFloat(lat)
	return nil
}

// scanEWKB parses EWKB (Extended Well-Known Binary) for a Point.
// Layout: 1 byte order, 4 bytes type, optional 4 bytes SRID, 8 bytes X (lng), 8 bytes Y (lat).
func (p *GeoPoint) scanEWKB(b []byte) error {
	const (
		wkbXDR       = 0
		wkbNDR       = 1
		wkbPoint     = 1
		wkbSRIDFlag  = 0x20000000
		minPointSize = 1 + 4 + 8 + 8 // order + type + X + Y
	)
	if len(b) < minPointSize {
		return fmt.Errorf("EWKB too short for Point")
	}
	order := b[0]
	var bo binary.ByteOrder = binary.LittleEndian
	if order == wkbXDR {
		bo = binary.BigEndian
	}
	typ := bo.Uint32(b[1:5])
	if typ&0x1f != wkbPoint {
		return fmt.Errorf("EWKB type is not Point: %d", typ)
	}
	off := 5
	if typ&wkbSRIDFlag != 0 {
		if len(b) < 5+4+8+8 {
			return fmt.Errorf("EWKB too short for Point with SRID")
		}
		off += 4 // skip SRID
	}
	if len(b) < off+8+8 {
		return fmt.Errorf("EWKB too short for coordinates")
	}
	lng := math.Float64frombits(bo.Uint64(b[off : off+8]))
	lat := math.Float64frombits(bo.Uint64(b[off+8 : off+16]))
	p.Lng = decimal.NewFromFloat(lng)
	p.Lat = decimal.NewFromFloat(lat)
	return nil
}

// Validate validates the GeoPoint coordinates.
func (p GeoPoint) Validate() error {
	return ValidateCoordinates(p.Lat, p.Lng)
}

// IsZero returns true if both coordinates are zero.
func (p GeoPoint) IsZero() bool {
	return p.Lat.IsZero() && p.Lng.IsZero()
}

// NewGeoPoint creates a GeoPoint from latitude and longitude.
func NewGeoPoint(lat, lng decimal.Decimal) GeoPoint {
	return GeoPoint{Lat: lat, Lng: lng}
}
