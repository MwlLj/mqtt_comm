package randtool

import (
	"strconv"
	"time"
)

func GetNSTimeStamp() string {
	return strconv.FormatInt(int64(time.Now().UnixNano()), 10)
}
