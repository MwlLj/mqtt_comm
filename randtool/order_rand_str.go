package randtool

import (
	"strings"
)

func GetOrderRandStr(customStr string) string {
	// ns time stamp
	stamp := GetNSTimeStamp()
	// 32bit rand string
	randString := GetRandomString(32)
	// append
	return strings.Join([]string{stamp, randString, customStr}, "_")
}
