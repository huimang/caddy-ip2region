package ip2region

import (
	"fmt"
	"strconv"
	"strings"
)

// countryISO 是国家名到 ISO 3166-1 二字码的兜底映射。
// 当前 xdb 的最后一个字段通常已经是国家码，仅在缺失该字段时使用此表。
var countryISO = map[string]string{
	"中国":                   "CN",
	"China":                "CN",
	"阿联酋":                  "AE",
	"阿拉伯联合酋长国":             "AE",
	"United Arab Emirates": "AE",
	"沙特阿拉伯":                "SA",
	"沙特":                   "SA",
	"Saudi Arabia":         "SA",
	"卡塔尔":                  "QA",
	"Qatar":                "QA",
	"科威特":                  "KW",
	"Kuwait":               "KW",
	"巴林":                   "BH",
	"Bahrain":              "BH",
	"阿曼":                   "OM",
	"Oman":                 "OM",
	"也门":                   "YE",
	"Yemen":                "YE",
	"伊拉克":                  "IQ",
	"Iraq":                 "IQ",
	"伊朗":                   "IR",
	"Iran":                 "IR",
	"约旦":                   "JO",
	"Jordan":               "JO",
	"黎巴嫩":                  "LB",
	"Lebanon":              "LB",
	"叙利亚":                  "SY",
	"Syria":                "SY",
	"以色列":                  "IL",
	"Israel":               "IL",
	"巴勒斯坦":                 "PS",
	"Palestine":            "PS",
	"土耳其":                  "TR",
	"Turkey":               "TR",
	"埃及":                   "EG",
	"Egypt":                "EG",
	"新加坡":                  "SG",
	"Singapore":            "SG",
	"马来西亚":                 "MY",
	"Malaysia":             "MY",
	"印度尼西亚":                "ID",
	"印尼":                   "ID",
	"Indonesia":            "ID",
	"泰国":                   "TH",
	"Thailand":             "TH",
	"越南":                   "VN",
	"Vietnam":              "VN",
	"菲律宾":                  "PH",
	"Philippines":          "PH",
	"文莱":                   "BN",
	"文莱达鲁萨兰国":              "BN",
	"Brunei":               "BN",
	"柬埔寨":                  "KH",
	"Cambodia":             "KH",
	"老挝":                   "LA",
	"Laos":                 "LA",
	"缅甸":                   "MM",
	"Myanmar":              "MM",
	"东帝汶":                  "TL",
	"Timor-Leste":          "TL",
	"美国":                   "US",
	"United States":        "US",
	"英国":                   "GB",
	"United Kingdom":       "GB",
	"德国":                   "DE",
	"Germany":              "DE",
	"法国":                   "FR",
	"France":               "FR",
	"日本":                   "JP",
	"Japan":                "JP",
	"韩国":                   "KR",
	"South Korea":          "KR",
	"澳大利亚":                 "AU",
	"Australia":            "AU",
	"加拿大":                  "CA",
	"Canada":               "CA",
	"俄罗斯":                  "RU",
	"Russia":               "RU",
	"印度":                   "IN",
	"India":                "IN",
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

	if country == "中国" || strings.EqualFold(country, "China") {
		for _, index := range []int{1, 2} {
			if len(parts) <= index {
				continue
			}
			province := parts[index]
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

	if strings.EqualFold(country, "Reserved") {
		return ""
	}

	last := strings.TrimSpace(parts[len(parts)-1])
	if len(last) == 2 && isASCIILetter(last[0]) && isASCIILetter(last[1]) {
		return strings.ToUpper(last)
	}

	return countryISO[country]
}

func isASCIILetter(character byte) bool {
	return (character >= 'A' && character <= 'Z') || (character >= 'a' && character <= 'z')
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
