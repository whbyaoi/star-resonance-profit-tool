package good

import (
	"fmt"
	"testing"
)

func TestPrice(t *testing.T) {
	item := Name2Item["辉耀石"]
	_, d, err := GetItemBestProfitByFocus(item.Name, 400, map[string]float64{
		"阿兹特精矿石": 200,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("p: %v\n", d.Comment())
}
