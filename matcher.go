// Package ip2region provides a Caddy HTTP request matcher that uses ip2region
// xdb data to match requests by the client IP's country.
package ip2region

import (
	"fmt"
	"net"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/lionsoul2014/ip2region/binding/golang/service"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

const defaultSearchers = 20

func init() {
	caddy.RegisterModule(Matcher{})
}

// Matcher matches requests by the country of the client IP using ip2region xdb data.
type Matcher struct {
	// DBPath is the required path to the IPv4 xdb file.
	DBPath string `json:"db_path,omitempty"`

	// DBPathV6 is the optional path to the IPv6 xdb file.
	DBPathV6 string `json:"db_path_v6,omitempty"`

	// AllowCountries is the list of allowed ISO 3166-1 alpha-2 country codes.
	AllowCountries []string `json:"allow_countries,omitempty"`

	// DenyCountries is the list of denied ISO 3166-1 alpha-2 country codes.
	DenyCountries []string `json:"deny_countries,omitempty"`

	// Searchers is the searcher pool size. The default is 20.
	Searchers int `json:"searchers,omitempty"`

	region *service.Ip2Region
}

// CaddyModule returns the Caddy module information.
func (Matcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.ip2region",
		New: func() caddy.Module { return new(Matcher) },
	}
}

// Provision loads the xdb files and creates a concurrency-safe search service.
func (m *Matcher) Provision(_ caddy.Context) error {
	if m.DBPath == "" {
		return fmt.Errorf("db_path is required")
	}
	if m.Searchers == 0 {
		m.Searchers = defaultSearchers
	}
	if err := xdb.VerifyFromFile(m.DBPath); err != nil {
		return fmt.Errorf("verify ipv4 xdb %s: %w", m.DBPath, err)
	}
	v4Config, err := service.NewV4Config(service.VIndexCache, m.DBPath, m.Searchers)
	if err != nil {
		return fmt.Errorf("create ipv4 config: %w", err)
	}

	var v6Config *service.Config
	if m.DBPathV6 != "" {
		if err := xdb.VerifyFromFile(m.DBPathV6); err != nil {
			return fmt.Errorf("verify ipv6 xdb %s: %w", m.DBPathV6, err)
		}
		v6Config, err = service.NewV6Config(service.VIndexCache, m.DBPathV6, m.Searchers)
		if err != nil {
			return fmt.Errorf("create ipv6 config: %w", err)
		}
	}

	region, err := service.NewIp2Region(v4Config, v6Config)
	if err != nil {
		return fmt.Errorf("create ip2region service: %w", err)
	}
	m.region = region
	m.AllowCountries = normalizeCountryCodes(m.AllowCountries)
	m.DenyCountries = normalizeCountryCodes(m.DenyCountries)
	return nil
}

// Validate checks required options and country codes.
func (m *Matcher) Validate() error {
	if m.DBPath == "" {
		return fmt.Errorf("db_path is required")
	}
	if len(m.AllowCountries) == 0 && len(m.DenyCountries) == 0 {
		return fmt.Errorf("allow_countries or deny_countries is required")
	}
	if m.Searchers < 0 {
		return fmt.Errorf("searchers must not be negative")
	}
	if err := validateCountryCodes(m.AllowCountries); err != nil {
		return fmt.Errorf("allow_countries: %w", err)
	}
	if err := validateCountryCodes(m.DenyCountries); err != nil {
		return fmt.Errorf("deny_countries: %w", err)
	}
	return nil
}

// Match reports whether the client IP belongs to the configured country lists.
func (m *Matcher) Match(request *http.Request) bool {
	if m.region == nil {
		return false
	}
	ip := clientIP(request)
	if ip == "" {
		return false
	}
	region, err := m.region.Search(ip)
	if err != nil || region == "" {
		return false
	}

	code := countryCode(region)
	if code == "" {
		return false
	}
	if len(m.AllowCountries) > 0 && !containsCountry(m.AllowCountries, code) {
		return false
	}
	if containsCountry(m.DenyCountries, code) {
		return false
	}
	return true
}

// MatchWithError implements caddyhttp.RequestMatcherWithError.
func (m *Matcher) MatchWithError(request *http.Request) (bool, error) {
	return m.Match(request), nil
}

// Cleanup releases the search service.
func (m *Matcher) Cleanup() error {
	if m.region != nil {
		m.region.Close()
		m.region = nil
	}
	return nil
}

// UnmarshalCaddyfile parses the ip2region matcher configuration from a Caddyfile.
func (m *Matcher) UnmarshalCaddyfile(dispenser *caddyfile.Dispenser) error {
	for dispenser.Next() {
		for nesting := dispenser.Nesting(); dispenser.NextBlock(nesting); {
			switch dispenser.Val() {
			case "db_path":
				if !dispenser.NextArg() {
					return dispenser.ArgErr()
				}
				m.DBPath = dispenser.Val()
			case "db_path_v6":
				if !dispenser.NextArg() {
					return dispenser.ArgErr()
				}
				m.DBPathV6 = dispenser.Val()
			case "allow_countries":
				args := dispenser.RemainingArgs()
				if len(args) == 0 {
					return dispenser.ArgErr()
				}
				m.AllowCountries = append(m.AllowCountries, args...)
			case "deny_countries":
				args := dispenser.RemainingArgs()
				if len(args) == 0 {
					return dispenser.ArgErr()
				}
				m.DenyCountries = append(m.DenyCountries, args...)
			case "searchers":
				if !dispenser.NextArg() {
					return dispenser.ArgErr()
				}
				searchers, err := parseSearchers(dispenser.Val())
				if err != nil {
					return dispenser.Errf("invalid searchers: %v", err)
				}
				m.Searchers = searchers
			default:
				return dispenser.Errf("unrecognized subdirective: %s", dispenser.Val())
			}
		}
	}
	return nil
}

func clientIP(request *http.Request) string {
	if value, ok := caddyhttp.GetVar(request.Context(), caddyhttp.ClientIPVarKey).(string); ok && value != "" {
		return value
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return request.RemoteAddr
	}
	return host
}

func normalizeCountryCodes(codes []string) []string {
	result := make([]string, 0, len(codes))
	for _, code := range codes {
		normalized := normalizeCountryCode(code)
		if normalized != "" {
			result = append(result, normalized)
		}
	}
	return result
}

func validateCountryCodes(codes []string) error {
	for _, code := range codes {
		if !isISOCountryCode(code) {
			return fmt.Errorf("invalid ISO 3166-1 alpha-2 country code %q", code)
		}
	}
	return nil
}

func containsCountry(codes []string, code string) bool {
	for _, candidate := range codes {
		if candidate == code {
			return true
		}
	}
	return false
}

var (
	_ caddy.Module                      = (*Matcher)(nil)
	_ caddy.Provisioner                 = (*Matcher)(nil)
	_ caddy.Validator                   = (*Matcher)(nil)
	_ caddy.CleanerUpper                = (*Matcher)(nil)
	_ caddyhttp.RequestMatcher          = (*Matcher)(nil)
	_ caddyhttp.RequestMatcherWithError = (*Matcher)(nil)
	_ caddyfile.Unmarshaler             = (*Matcher)(nil)
)
