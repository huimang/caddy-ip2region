package ip2region

import (
	"fmt"
	"strconv"
	"strings"
)

// countryISO 是 ip2region 中文国家名到 ISO 3166-1 二字码的映射。
// ip2region 对同一国家可能返回不同写法，因此按实测结果保留别名。
var countryISO = map[string]string{
	"中国":       "CN",
	"阿联酋":      "AE",
	"阿拉伯联合酋长国": "AE",
	"沙特阿拉伯":    "SA",
	"沙特":       "SA",
	"卡塔尔":      "QA",
	"科威特":      "KW",
	"巴林":       "BH",
	"阿曼":       "OM",
	"也门":       "YE",
	"伊拉克":      "IQ",
	"伊朗":       "IR",
	"约旦":       "JO",
	"黎巴嫩":      "LB",
	"叙利亚":      "SY",
	"以色列":      "IL",
	"巴勒斯坦":     "PS",
	"土耳其":      "TR",
	"埃及":       "EG",
	"新加坡":      "SG",
	"马来西亚":     "MY",
	"印度尼西亚":    "ID",
	"印尼":       "ID",
	"泰国":       "TH",
	"越南":       "VN",
	"菲律宾":      "PH",
	"文莱":       "BN",
	"文莱达鲁萨兰国":  "BN",
	"柬埔寨":      "KH",
	"老挝":       "LA",
	"缅甸":       "MM",
	"东帝汶":      "TL",
	"美国":       "US",
	"英国":       "GB",
	"德国":       "DE",
	"法国":       "FR",
	"日本":       "JP",
	"韩国":       "KR",
	"澳大利亚":     "AU",
	"加拿大":      "CA",
	"俄罗斯":      "RU",
	"印度":       "IN",
}

// countryCode 从 ip2region 的 region 字符串提取国家码。
// region 格式为 国家|区域|省份|城市|ISP。
func countryCode(region string) string {
	parts := strings.Split(region, "|")
	if len(parts) == 0 {
		return ""
	}

	country := strings.TrimSpace(parts[0])
	if country == "" || country == "0" {
		return ""
	}

	switch {
	case strings.Contains(country, "香港"):
		return "HK"
	case strings.Contains(country, "澳门"):
		return "MO"
	case strings.Contains(country, "台湾"):
		return "TW"
	}

	if country != "中国" {
		return countryISO[country]
	}
	if len(parts) > 2 {
		province := strings.TrimSpace(parts[2])
		switch {
		case strings.Contains(province, "香港"):
			return "HK"
		case strings.Contains(province, "澳门"):
			return "MO"
		case strings.Contains(province, "台湾"):
			return "TW"
		}
	}
	return "CN"
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
