package ip2region

import (
	"strings"
	"unicode"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// isoCountryCodeList contains the 249 officially assigned ISO 3166-1 alpha-2 codes.
const isoCountryCodeList = `
AD AE AF AG AI AL AM AO AQ AR AS AT AU AW AX AZ
BA BB BD BE BF BG BH BI BJ BL BM BN BO BQ BR BS BT BV BW BY BZ
CA CC CD CF CG CH CI CK CL CM CN CO CR CU CV CW CX CY CZ
DE DJ DK DM DO DZ
EC EE EG EH ER ES ET
FI FJ FK FM FO FR
GA GB GD GE GF GG GH GI GL GM GN GP GQ GR GS GT GU GW GY
HK HM HN HR HT HU
ID IE IL IM IN IO IQ IR IS IT
JE JM JO JP
KE KG KH KI KM KN KP KR KW KY KZ
LA LB LC LI LK LR LS LT LU LV LY
MA MC MD ME MF MG MH MK ML MM MN MO MP MQ MR MS MT MU MV MW MX MY MZ
NA NC NE NF NG NI NL NO NP NR NU NZ
OM
PA PE PF PG PH PK PL PM PN PR PS PT PW PY
QA
RE RO RS RU RW
SA SB SC SD SE SG SH SI SJ SK SL SM SN SO SR SS ST SV SX SY SZ
TC TD TF TG TH TJ TK TL TM TN TO TR TT TV TW TZ
UA UG UM US UY UZ
VA VC VE VG VI VN VU
WF WS
YE YT
ZA ZM ZW
`

// countryCodeAliases maps common non-ISO forms to official codes.
var countryCodeAliases = map[string]string{
	"UK": "GB",
	"EL": "GR",
}

// countryNameAliases adds common short forms, historical names and non-standard spellings found in ip2region data.
var countryNameAliases = map[string]string{
	"UAE":                                    "AE",
	"阿联酋":                                    "AE",
	"KSA":                                    "SA",
	"沙特":                                     "SA",
	"印尼":                                     "ID",
	"Brunei Darussalam":                      "BN",
	"East Timor":                             "TL",
	"Myanmar":                                "MM",
	"Burma":                                  "MM",
	"Palestine":                              "PS",
	"巴勒斯坦":                                   "PS",
	"Hong Kong":                              "HK",
	"香港":                                     "HK",
	"Macao":                                  "MO",
	"Macau":                                  "MO",
	"澳门":                                     "MO",
	"Taiwan":                                 "TW",
	"台湾":                                     "TW",
	"Korea, Republic of":                     "KR",
	"North Korea":                            "KP",
	"Korea, Democratic People's Republic of": "KP",
	"Russian Federation":                     "RU",
	"Viet Nam":                               "VN",
	"United States of America":               "US",
	"USA":                                    "US",
	"Great Britain":                          "GB",
	"Czech Republic":                         "CZ",
	"Türkiye":                                "TR",
	"Turkiye":                                "TR",
	"Cote d'Ivoire":                          "CI",
	"Ivory Coast":                            "CI",
	"Democratic Republic of the Congo":       "CD",
	"Republic of the Congo":                  "CG",
	"Swaziland":                              "SZ",
	"Macedonia":                              "MK",
	"Cape Verde":                             "CV",
	"Holland":                                "NL",
	"Kosovo":                                 "XK",
	"科索沃":                                    "XK",
}

var (
	isoCountryCodes  = strings.Fields(isoCountryCodeList)
	allCountryCodes  = append(append([]string{}, isoCountryCodes...), "XK")
	isoCountrySet    = buildISOCountrySet()
	countryNameIndex = buildCountryNameIndex()
)

func buildISOCountrySet() map[string]struct{} {
	set := make(map[string]struct{}, len(allCountryCodes))
	for _, code := range allCountryCodes {
		set[code] = struct{}{}
	}
	return set
}

func buildCountryNameIndex() map[string]string {
	index := make(map[string]string, len(allCountryCodes)*2)
	ambiguous := make(map[string]struct{})

	english := display.English.Regions()
	chinese := display.SimplifiedChinese.Regions()
	for _, code := range allCountryCodes {
		region, err := language.ParseRegion(code)
		if err != nil {
			continue
		}
		for _, name := range []string{english.Name(region), chinese.Name(region)} {
			key := normalizeCountryName(name)
			if key == "" {
				continue
			}
			if existing, ok := index[key]; ok && existing != code {
				ambiguous[key] = struct{}{}
				continue
			}
			index[key] = code
		}
	}

	for key := range ambiguous {
		delete(index, key)
	}
	for name, code := range countryNameAliases {
		key := normalizeCountryName(name)
		if key != "" {
			index[key] = code
		}
	}
	return index
}

func isISOCountryCode(value string) bool {
	if len(value) != 2 {
		return false
	}
	_, ok := isoCountrySet[strings.ToUpper(value)]
	return ok
}

func normalizeCountryCode(code string) string {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if alias, ok := countryCodeAliases[normalized]; ok {
		return alias
	}
	return normalized
}

func lookupCountryName(name string) string {
	return countryNameIndex[normalizeCountryName(name)]
}

func normalizeCountryName(name string) string {
	var builder strings.Builder
	for _, character := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}
