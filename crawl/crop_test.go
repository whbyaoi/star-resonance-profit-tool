package crawl

import (
	"fmt"
	"image/png"
	"os"
	"testing"
	"xhgm_price_tool/match"
)

func TestCrop(t *testing.T) {
	InitGamePid()
	target, err := GetGameScreenshot(GAME_PID, true)
	if err != nil {
		panic(err)
	}
	_, err = GetItemImages(target, 0, true)
	if err != nil {
		panic(err)
	}
}

func Test1(_ *testing.T) {
	f, _ := os.Open("../image.png")
	img, _ := png.Decode(f)
	p := match.GetPrice(img, true)
	fmt.Printf("p: %v\n", p)
}
