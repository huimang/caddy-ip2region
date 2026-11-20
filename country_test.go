package ip2region

import "testing"

func TestCountryCode(t *testing.T) {
	tests := []struct {
		name   string
		region string
		want   string
	}{
		{name: "中国大陆省份", region: "中国|浙江省|杭州市|阿里|CN", want: "CN"},
		{name: "中国未知区域", region: "中国|0|0|0|0", want: "CN"},
		{name: "中国香港", region: "中国|香港特别行政区|0|电讯盈科|CN", want: "HK"},
		{name: "中国澳门", region: "中国|澳门特别行政区|0|澳门电讯|CN", want: "MO"},
		{name: "中国台湾", region: "中国|台湾省|0|中华电信|CN", want: "TW"},
		{name: "标准格式中国香港", region: "中国|0|香港特别行政区|0|0", want: "HK"},
		{name: "海外国家码字段", region: "United Arab Emirates|Dubai|0|0|AE", want: "AE"},
		{name: "海外标准格式", region: "Singapore|Singapore|0|SingNet|SG", want: "SG"},
		{name: "无国家码字段", region: "美国|0|加利福尼亚州|0|0", want: "US"},
		{name: "意大利标准格式", region: "意大利|0|拉齐奥|罗马|0", want: "IT"},
		{name: "南非英文名称", region: "South Africa|0|Gauteng|0|0", want: "ZA"},
		{name: "阿联酋简称", region: "阿联酋|Dubai|0|0|0", want: "AE"},
		{name: "香港英文名称", region: "Hong Kong|0|0|0|0", want: "HK"},
		{name: "中国香港国家段", region: "中国香港|0|0|0|0", want: "HK"},
		{name: "澳门英文名称", region: "Macao|0|0|0|0", want: "MO"},
		{name: "台湾英文名称", region: "Taiwan|0|0|0|0", want: "TW"},
		{name: "旧称斯威士兰", region: "Swaziland|0|0|0|0", want: "SZ"},
		{name: "保留地址", region: "Reserved|Reserved|Reserved|0|0", want: ""},
		{name: "未知区域", region: "0|0|0|0|0", want: ""},
		{name: "未知国家", region: "Atlantis|0|0|0|0", want: ""},
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
