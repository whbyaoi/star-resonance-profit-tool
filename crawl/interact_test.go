package crawl

import (
	"testing"
	"time"
)

func TestSwitch(_ *testing.T) {
	InitGamePid()

	// fmt.Printf("x_offset: %v\n", x_offset)
	// fmt.Printf("y_offset: %v\n", y_offset)
	// SwitchSidebar(0)
	// SwitchTab(0)
	// SwitchTab(1)
	// SwitchTab(2)
	// SwitchTab(3)
	// SwitchSidebar(1)
	// SwitchTab(0)
	// SwitchTab(1)
	// SwitchTab(2)
	// SwitchTab(3)
	// SwitchTab(4)
	// SwitchTab(5)
	// SlideTab()
	// SwitchTab(6)
	// SwitchTab(7)
	SlideNextPage(3)
	time.Sleep(2 * time.Second)
	SlideNextPage(3)
	time.Sleep(2 * time.Second)
	SlideNextPage(3)
	time.Sleep(2 * time.Second)
	SlideNextPage(3)
	time.Sleep(2 * time.Second)
	SlideNextPage(1)
	time.Sleep(2 * time.Second)
}
