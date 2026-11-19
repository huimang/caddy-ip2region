package ip2region

import (
	"net/http/httptest"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestUnmarshalCaddyfile(t *testing.T) {
	dispenser := caddyfile.NewTestDispenser(`ip2region {
		db_path /opt/ip2region/ip2region_v4.xdb
		db_path_v6 /opt/ip2region/ip2region_v6.xdb
		allow_countries cn hk mo tw
		deny_countries 0
		searchers 5
	}`)

	var matcher Matcher
	if err := matcher.UnmarshalCaddyfile(dispenser); err != nil {
		t.Fatalf("UnmarshalCaddyfile() error = %v", err)
	}
	if matcher.DBPath != "/opt/ip2region/ip2region_v4.xdb" {
		t.Fatalf("DBPath = %q", matcher.DBPath)
	}
	if matcher.DBPathV6 != "/opt/ip2region/ip2region_v6.xdb" {
		t.Fatalf("DBPathV6 = %q", matcher.DBPathV6)
	}
	if len(matcher.AllowCountries) != 4 {
		t.Fatalf("AllowCountries = %v", matcher.AllowCountries)
	}
	if matcher.Searchers != 5 {
		t.Fatalf("Searchers = %d", matcher.Searchers)
	}
}

func TestValidateRejectsMissingConfig(t *testing.T) {
	tests := []struct {
		name    string
		matcher Matcher
	}{
		{name: "缺少数据库路径", matcher: Matcher{AllowCountries: []string{"CN"}}},
		{name: "缺少国家列表", matcher: Matcher{DBPath: "/tmp/ip2region_v4.xdb"}},
		{name: "非法国家码", matcher: Matcher{DBPath: "/tmp/ip2region_v4.xdb", AllowCountries: []string{"CHN"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.matcher.Validate(); err == nil {
				t.Fatal("Validate() = nil, want error")
			}
		})
	}
}

func TestClientIPFallback(t *testing.T) {
	request := httptest.NewRequest("GET", "https://gymlin.com/", nil)
	request.RemoteAddr = "203.0.113.10:54321"
	if got := clientIP(request); got != "203.0.113.10" {
		t.Fatalf("clientIP() = %q, want %q", got, "203.0.113.10")
	}
}
