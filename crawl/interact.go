package crawl

import (
	"time"

	"github.com/go-vgo/robotgo"
)

func SwitchSidebar(index int) {
	tmp := make([]int, 4)
	copy(tmp, sidebar_start_box)
	tmp[1] += sidebar_vertical_space * index
	tmp[3] += sidebar_vertical_space * index
	MoveToBox(tmp)
	time.Sleep(time.Second)
}

func SwitchTab(index int) {
	tmp := make([]int, 4)
	copy(tmp, tab_start_box)
	if index > 5 {
		index -= 3
		tmp[0] += tab_horizontal_space * index
		tmp[2] += tab_horizontal_space * index
		robotgo.MoveClick(tmp[2]+x_offset, (tmp[1]+tmp[3])/2+y_offset)
	} else {
		tmp[0] += tab_horizontal_space * index
		tmp[2] += tab_horizontal_space * index
		MoveToBox(tmp)
	}
	time.Sleep(2*time.Second)
}

func SlideTab() {
	robotgo.Move(
		(tab_start_box[0]+tab_start_box[2])/2+5*tab_horizontal_space+x_offset,
		(tab_start_box[1]+tab_start_box[3])/2+y_offset,
	)
	robotgo.DragSmooth(
		tab_start_box[0]+2*tab_horizontal_space+x_offset,
		(tab_start_box[1]+tab_start_box[3])/2+y_offset,
		1.0,
		1.1,
	)
}

func MoveToBox(box []int) {
	// x := (box[0]+box[2])/2 + x_offset
	// y := (box[1]+box[3])/2 + y_offset
	// fmt.Println(x, y)
	robotgo.MoveClick((box[0]+box[2])/2+x_offset, (box[1]+box[3])/2+y_offset)
}

var hscroll = map[string][]int{
	"1_3": {3, 3, 3, 3, 1},
	"1_4": {3, 3, 3, 2},
	"1_6": {3},
}

func SlideNextPage(rowCnt int) {
	var offset int
	switch rowCnt {
	case 1:
		offset = -280
	case 2:
		offset = -530
	case 3:
		offset = -754
	default:
		panic("unknow scroll unit")
	}
	// fmt.Printf("offset: %v\n", offset)
	robotgo.Move(1810+x_offset, 900+y_offset)
	robotgo.DragSmooth(1810+x_offset, 900+y_offset+offset, 1.0, 1.1)
	time.Sleep(2 * time.Second)
}

func SlideToTop() {
	robotgo.Move(1820+x_offset, 900+y_offset)
	robotgo.DragSmooth(1820+x_offset, 200+y_offset)
}
