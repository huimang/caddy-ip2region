package ip2region

import (
	"strings"
	"testing"
)

func TestISOCountryCodeList(t *testing.T) {
	if len(isoCountryCodes) != 249 {
		t.Fatalf("isoCountryCodes has %d entries, want 249", len(isoCountryCodes))
	}

	seen := make(map[string]struct{}, len(isoCountryCodes))
	for _, code := range isoCountryCodes {
		if len(code) != 2 || code != strings.ToUpper(code) {
			t.Fatalf("invalid ISO country code %q", code)
		}
		if _, exists := seen[code]; exists {
			t.Fatalf("duplicate ISO country code %q", code)
		}
		seen[code] = struct{}{}
		if !isISOCountryCode(code) {
			t.Fatalf("code %q missing from ISO country set", code)
		}
	}

	if !isISOCountryCode("XK") {
		t.Fatal("XK should be accepted as a user-assigned code")
	}
	if isISOCountryCode("XX") {
		t.Fatal("XX must not be accepted as an ISO country code")
	}
}

func TestEveryISOCodeFromCountryField(t *testing.T) {
	for _, code := range isoCountryCodes {
		region := "Unknown|0|0|0|" + code
		if got := countryCode(region); got != code {
			t.Fatalf("countryCode(%q) = %q, want %q", region, got, code)
		}
	}
}

func TestCountryNameFallback(t *testing.T) {
	tests := []struct {
		name   string
		region string
		want   string
	}{
		{name: "意大利中文名称", region: "意大利|0|拉齐奥|罗马|0", want: "IT"},
		{name: "巴西中文名称", region: "巴西|0|圣保罗|圣保罗|0", want: "BR"},
		{name: "南非英文名称", region: "South Africa|Gauteng|0|0|0", want: "ZA"},
		{name: "科特迪瓦英文名称", region: "Côte d’Ivoire|0|0|0|0", want: "CI"},
		{name: "科特迪瓦无重音名称", region: "Cote d'Ivoire|0|0|0|0", want: "CI"},
		{name: "缅甸英文名称", region: "Myanmar|0|0|0|0", want: "MM"},
		{name: "缅甸旧称", region: "Burma|0|0|0|0", want: "MM"},
		{name: "巴勒斯坦英文名称", region: "Palestine|0|0|0|0", want: "PS"},
		{name: "科索沃英文名称", region: "Kosovo|0|0|0|0", want: "XK"},
		{name: "科索沃中文名称", region: "科索沃|0|0|0|0", want: "XK"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := countryCode(test.region); got != test.want {
				t.Fatalf("countryCode(%q) = %q, want %q", test.region, got, test.want)
			}
		})
	}
}

func TestCountryCodeNormalization(t *testing.T) {
	if got := normalizeCountryCode(" us "); got != "US" {
		t.Fatalf("normalizeCountryCode(\" us \") = %q, want US", got)
	}
	if got := normalizeCountryCode("uk"); got != "GB" {
		t.Fatalf("normalizeCountryCode(\"uk\") = %q, want GB", got)
	}
	if got := normalizeCountryCode("el"); got != "GR" {
		t.Fatalf("normalizeCountryCode(\"el\") = %q, want GR", got)
	}
	if err := validateCountryCodes([]string{"US", "XK"}); err != nil {
		t.Fatalf("validateCountryCodes() error = %v", err)
	}
	if err := validateCountryCodes([]string{"XX"}); err == nil {
		t.Fatal("validateCountryCodes() = nil for invalid code, want error")
	}
}
