package ip2region

import (
	"fmt"
	"strconv"
	"strings"
)

// specialRegions 用于识别数据中可能出现在国家段或省份段的港澳台写法。
var specialRegions = []struct {
	name string
	code string
}{
	{name: "香港", code: "HK"},
	{name: "hongkong", code: "HK"},
	{name: "澳门", code: "MO"},
	{name: "macao", code: "MO"},
	{name: "macau", code: "MO"},
	{name: "台湾", code: "TW"},
	{name: "taiwan", code: "TW"},
}

// countryCode 从 ip2region 的 region 字符串提取国家码。
// 当前数据格式为 国家|省份|城市|ISP|国家码，标准格式为 国家|区域|省份|城市|ISP。
func countryCode(region string) string {
	parts := strings.Split(region, "|")
	if len(parts) == 0 {
		return ""
	}

	country := strings.TrimSpace(parts[0])
	if country == "" || country == "0" {
		return ""
	}

	if code := specialRegionCode(country); code != "" {
		return code
	}

	if country == "中国" || strings.EqualFold(country, "China") {
		for _, index := range []int{1, 2} {
			if len(parts) <= index {
				continue
			}
			if code := specialRegionCode(parts[index]); code != "" {
				return code
			}
		}
		return "CN"
	}

	if strings.EqualFold(country, "Reserved") {
		return ""
	}

	last := strings.TrimSpace(parts[len(parts)-1])
	if isISOCountryCode(last) {
		return strings.ToUpper(last)
	}

	return lookupCountryName(country)
}

func specialRegionCode(value string) string {
	normalized := normalizeCountryName(value)
	if normalized == "" {
		return ""
	}
	for _, region := range specialRegions {
		if strings.Contains(normalized, region.name) {
			return region.code
		}
	}
	return ""
}

func parseSearchers(value string) (int, error) {
	searchers, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if searchers <= 0 {
		return 0, fmt.Errorf("must be greater than 0")
	}
	return searchers, nil
}
