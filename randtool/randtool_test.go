package randtool

import (
	"fmt"
	"testing"
)

func TestGetOrderRandStr(t *testing.T) {
	result := GetOrderRandStr("test")
	fmt.Println(result)
}
