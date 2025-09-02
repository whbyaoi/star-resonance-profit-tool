package good

var SidebarMap = map[int]string{
	0: "养成道具",
	1: "生活职业",
	// 2: "模组",
	// 3: "外观",
}

// sidebar_tab -> name
var TabMap = map[int]map[int]string{
	0: {
		0: "能力养成",
		1: "装备养成",
		2: "意志",
		3: "幻想材料",
	},
	1: {
		0: "植生学",
		1: "矿物学",
		2: "晶石学",
		3: "烹饪",
		4: "炼金",
		5: "铸造",
		6: "工艺",
		7: "木作",
		8: "染料",
	},
}

func GetSiderbarName(index int) (string, bool) {
	name, ok := SidebarMap[index]
	return name, ok
}

func GetTabName(sidebarIndex, tabIndex int) (string, bool) {
	tmp, ok := TabMap[sidebarIndex]
	if !ok {
		return "", ok
	}
	name, ok := tmp[tabIndex]
	return name, ok
}
