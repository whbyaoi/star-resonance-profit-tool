package crawl

import (
	"fmt"
	"testing"

	"github.com/go-vgo/robotgo"
)

func TestRun(t *testing.T) {
	InitGamePid()
	Run(true)
}

func TestGet(t *testing.T) {
	ids, _ := robotgo.FindIds("Star")
	fmt.Printf("ids: %v\n", ids)
	fmt.Println(robotgo.FindName(4140))
	fmt.Println(robotgo.GetBounds(4140))
}
