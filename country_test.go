package ip2region

import "testing"

func TestCountryCode(t *testing.T) {
	tests := []struct {
		name   string
		region string
		want   string
	}{
		{name: "中国大陆省份", region: "中国|0|广东省|深圳市|电信", want: "CN"},
		{name: "中国未知区域", region: "中国|0|0|0|0", want: "CN"},
		{name: "中国香港", region: "中国|0|香港特别行政区|0|0", want: "HK"},
		{name: "中国澳门", region: "中国|0|澳门特别行政区|0|0", want: "MO"},
		{name: "中国台湾", region: "中国|0|台湾省|台北市|0", want: "TW"},
		{name: "海外国家", region: "阿联酋|0|迪拜|0|0", want: "AE"},
		{name: "海外别名", region: "印度尼西亚|0|雅加达|0|0", want: "ID"},
		{name: "未知区域", region: "0|0|0|0|0", want: ""},
		{name: "空字符串", region: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := countryCode(test.region); got != test.want {
				t.Fatalf("countryCode(%q) = %q, want %q", test.region, got, test.want)
			}
		})
	}
}
