package good

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"xhgm_price_tool/good/templates"
)

var ErrNoDrop error = errors.New("no drop")
var ErrNoRecipe error = errors.New("no recipe")
var ErrNoPrice error = errors.New("no price")

const JSON_DATA_FILE = "xhgm_items.json"

var HomeDir string

func init() {
	var err error
	HomeDir, err = os.UserHomeDir()
	if err != nil {
		panic("获取用户根目录失败")
	}

	file, err := os.Open(filepath.Join(HomeDir, JSON_DATA_FILE))
	if err != nil {
		fmt.Printf("目录 %s 下不存在文件 %s\n", HomeDir, JSON_DATA_FILE)
		fsFile, err := templates.Templates.Open("items.json")
		if err != nil {
			panic(err)
		}
		b, err := io.ReadAll(fsFile)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(b, &ItemArr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("使用预设数据: %v\n", len(ItemArr))
	} else {
		ItemArr, err = loadDiskData(file)
		if err != nil {
			// 如果加载本地数据失败，则使用初始数据
			fmt.Printf("加载目录 %s 下文件 %s 失败: %s\n", HomeDir, JSON_DATA_FILE, err.Error())
			fsFile, err := templates.Templates.Open("items.json")
			if err != nil {
				panic(err)
			}
			b, err := io.ReadAll(fsFile)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(b, &ItemArr)
			if err != nil {
				panic(err)
			}
			fmt.Printf("使用预设数据: %v\n", len(ItemArr))
		} else {
			fmt.Println("加载本地数据成功")
		}
	}

	// 构建分类映射
	itemsMap := map[string][]*Good{}
	for i := range ItemArr {
		if len(ItemArr[i].PriceRange) >= 2 {
			item := ItemArr[i]
			item.PriceRange = []int{
				item.PriceRange[0] * 9 / 10,
				item.PriceRange[1] * 11 / 10,
			}
		}
		key := fmt.Sprintf("%d_%d", ItemArr[i].Sidebar, ItemArr[i].Tab)
		itemsMap[key] = append(itemsMap[key], ItemArr[i])
		Name2Item[ItemArr[i].Name] = ItemArr[i]
	}
	cols := 6
	for k := range itemsMap {
		for i := range itemsMap[k] {
			if itemsMap[k][i].Private {
				continue
			}
			var ok bool
			itemsMap[k][i].Category1, ok = GetSiderbarName(itemsMap[k][i].Sidebar)
			if !ok {
				panic(fmt.Sprintf("sidebar %d not found", itemsMap[k][i].Sidebar))
			}
			itemsMap[k][i].Category2, ok = GetTabName(itemsMap[k][i].Sidebar, itemsMap[k][i].Tab)
			if !ok {
				panic(fmt.Sprintf("sidebar %d tab %d not found", itemsMap[k][i].Sidebar, itemsMap[k][i].Tab))
			}
			Pos2Item[fmt.Sprintf("%d_%d_%d_%d",
				itemsMap[k][i].Sidebar, itemsMap[k][i].Tab, i%cols, i/cols)] = itemsMap[k][i]
		}
	}
}

var ItemArr = []*Good{}

var Name2Item = map[string]*Good{}

var Pos2Item = map[string]*Good{}

type Good struct {
	Sidebar    int    `json:"sidebar,omitempty"`
	Tab        int    `json:"tab,omitempty"`
	Name       string `json:"name"`
	PriceRange []int  `json:"range"`
	Price      int    `json:"price"`
	Category1  string `json:"category1,omitempty"`
	Category2  string `json:"category2,omitempty"`

	Private bool         `json:"private"`
	Recipe  []ItemRecipe `json:"recipe"`
}

type ItemRecipe struct {
	Drop      []RecipeDrop     `json:"drop"`
	Materials []RecipeMaterial `json:"materials"`
	Focus     int              `json:"focus"`
}

type RecipeDrop struct {
	Cnt  int    `json:"cnt"`
	Prob int    `json:"prob"`
	Name string `json:"name"` // 如果不为空，则表示副产品
}

type RecipeMaterial struct {
	Name string  `json:"name"`
	Cnt  float64 `json:"cnt"`
}

// GetProductionByFocus 计算只使用专注配方获得的产出
func (ir *ItemRecipe) GetProductionByFocus(focus float64, name string) (production ProductionDetail, err error) {
	production, err = ir.GetFocusProducingX(1, name)
	if err != nil {
		return production, err
	}
	return production.MultiPower(float64(focus) / production.Focus), nil
}

// GetFocusProducingX 按照配方获取x个产出所需要多少专注
func (ir ItemRecipe) GetFocusProducingX(x float64, name string) (production ProductionDetail, err error) {
	recipe := ir
	unit := 0.0
	for _, drop := range recipe.Drop {
		if drop.Name != "" && drop.Name != name {
			// 添加副产品
			production.ByProducts = append(production.ByProducts, ByProduct{
				Name: drop.Name,
				Cnt:  float64(drop.Prob) * float64(drop.Cnt) / 100,
			})
		} else {
			unit += float64(drop.Prob) * float64(drop.Cnt) / 100
		}
	}
	//
	if unit == 0.0 {
		return production, fmt.Errorf("%s %w", name, ErrNoDrop)
	}
	// 合并副产品
	name2byproduct := map[string]ByProduct{}
	for _, bp := range production.ByProducts {
		if _, ok := name2byproduct[bp.Name]; !ok {
			name2byproduct[bp.Name] = bp
		} else {
			name2byproduct[bp.Name] = ByProduct{
				Name: bp.Name,
				Cnt:  name2byproduct[bp.Name].Cnt + bp.Cnt,
			}
		}
	}
	production.ByProducts = []ByProduct{}
	for _, bp := range name2byproduct {
		bp.Price = getPrice(bp.Name, nil)
		production.ByProducts = append(production.ByProducts, bp)
	}

	// 先计算生产unit个需要多少专注，然后计算x个需要多少专注
	production.Cnt = unit
	production.Focus += float64(recipe.Focus)
	for _, material := range recipe.Materials {
		if item, ok := Name2Item[material.Name]; ok {
			// 获取所需材料所有配方的最低专注
			p, err := item.GetMinFocusProducingX(material.Cnt)
			if err != nil {
				return production, err
			}
			// 计算需要材料需要多少专注
			p.ItemName = material.Name
			p.Price = getPrice(p.ItemName, nil)
			production.Focus += float64(p.Focus)
			production.Details = append(production.Details, p)
		} else {
			// 如果不在表内，则产出不需要专注
		}
	}
	production.Price = getPrice(name, nil)
	return production.MultiPower(float64(x) / production.Cnt), nil
}

type ByProduct struct {
	Name  string  `json:"name"`
	Cnt   float64 `json:"cnt"`
	Price float64 `json:"price"`
}

type ProductionDetail struct {
	ItemName   string             `json:"item_name"`         // 物品名称
	Cnt        float64            `json:"cnt"`               // 物品产出数量
	Price      float64            `json:"price"`             // 参考价格
	Focus      float64            `json:"focus"`             // 生产Cnt个物品总专注消耗
	Cost       float64            `json:"cost"`              // 购买Cnt个物品的总金币花费
	Details    []ProductionDetail `json:"details,omitempty"` // 物品产出所需材料消耗情况
	ByProducts []ByProduct        `json:"by_products"`       // 物品的副产品
}

func (pd *ProductionDetail) GetAllMetarialNames() (rs []string) {
	for _, detail := range pd.Details {
		rs = append(rs, detail.ItemName)
		tmp := detail.GetAllMetarialNames()
		rs = append(rs, tmp...)
	}
	return rs
}

// GetTotalMaterialsCoinCost 获取生产情况消耗材料的金币花费情况
func (pd *ProductionDetail) GetTotalMaterialsCoinCost() (total float64) {
	for _, detail := range pd.Details {
		total += detail.Cost + detail.GetTotalMaterialsCoinCost()
	}
	return total
}

// TrBestProduction 计算纯专注产出情况替换为购买方案的最佳产出情况
func (pd *ProductionDetail) TrBestProduction(prices map[string]float64) (result *ProductionDetail, err error) {
	result = pd.Copy()
	if len(pd.Details) == 0 {
		return
	}
	for i := range pd.Details {
		err = pd.Details[i].trBestProduction(result, prices)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// TrProfit 将产出转换为利润
func (pd *ProductionDetail) TrProfit(prices map[string]float64) (profit Profit) {
	profit.Name = pd.ItemName
	profit.Cnt = pd.Cnt
	profit.Cost = pd.Cost
	profit.Price = getPrice(profit.Name, prices)
	for _, bp := range pd.ByProducts {
		tmp := getPrice(bp.Name, prices)
		profit.ByProducts = append(profit.ByProducts, Profit{
			Name:  bp.Name,
			Cnt:   bp.Cnt,
			Cost:  bp.Cnt,
			Price: tmp,
		})
	}
	return profit
}

// 获取pd在生产output时，最佳的产出result
func (pd *ProductionDetail) trBestProduction(output *ProductionDetail, prices map[string]float64) (err error) {
	for _, detail := range output.Details {
		if detail.ItemName == pd.ItemName {
			*pd = detail
			break
		}
	}
	avgPrice := getPrice(pd.ItemName, prices)
	if avgPrice == 0 {
		return nil
	}
	p := output.TrProfit(prices)
	outProfit := p.Value(prices)
	// for _, bp := range output.ByProducts {
	// 	price := getPrice(bp.Name, prices)
	// 	if price == 0 {
	// 		continue
	// 	}
	// 	outProfit += bp.Cnt * price * 0.95
	// }
	case1 := pd.Copy()
	// pd全部市场购买的情况
	ratio := output.Focus / (output.Focus - pd.Focus)
	newProfit := ratio * (outProfit - pd.Cnt*avgPrice)
	case1.Focus = 0
	case1.Cnt = pd.Cnt
	case1.Cost = pd.Cnt * avgPrice
	case1.Details = nil
	case1.ByProducts = nil
	// fmt.Printf("直接购买: %v\n", case1)
	// fmt.Printf("直接购买比例: %v\n", ratio)

	// 对pd的每种材料获取其最佳产出，然后手动制造
	case2 := pd.Copy()
	for i := range case2.Details {
		err = case2.Details[i].trBestProduction(case2, prices)
		if err != nil {
			return err
		}
	}
	ratio2 := output.Focus / (output.Focus - (pd.Focus - case2.Focus))
	newProfit2 := ratio2 * (outProfit - case2.Cost)
	// fmt.Printf("材料购买: %v\n", case2)
	// fmt.Printf("材料购买比例: %v\n", ratio2)

	// 比较哪种价格利润最高
	// fmt.Printf("原利润: %v\n", outProfit)
	// fmt.Printf("直接购买利润: %v\n", newProfit)
	// fmt.Printf("材料购买利润: %v\n", newProfit2)
	if outProfit > max(newProfit, newProfit2) {
		return nil
	} else if newProfit > max(outProfit, newProfit2) {
		for i := range output.Details {
			if output.Details[i].ItemName == pd.ItemName {
				output.Details[i] = *case1
				output.Cost += case1.Cost
				output.Focus -= pd.Focus
				*output = output.MultiPower(ratio)
				break
			}
		}
	} else if newProfit2 > max(outProfit, newProfit) {
		for i := range output.Details {
			if output.Details[i].ItemName == pd.ItemName {
				output.Details[i] = *case2
				output.Cost += case2.Cost
				output.Focus -= case2.Focus
				*output = output.MultiPower(ratio2)
				break
			}
		}
	}
	// fmt.Printf("新生产: %v\n\n\n", output)
	return nil
}

func (pd *ProductionDetail) GetAllProductions() (pds []ProductionDetail, err error) {
	return
}

func (pd *ProductionDetail) AdjustByPlan(plan map[string]bool) {
	cp := pd.Copy()
	var adjust func(parent, detail *ProductionDetail) (subFocus, addCost float64)
	adjust = func(parent, detail *ProductionDetail) (subFocus, addCost float64) {
		if plan[detail.ItemName] {
			subFocus, addCost = detail.Focus, detail.Cnt*detail.Price
			parent.Focus -= subFocus
			parent.Cost += addCost
			detail.Details = nil
			detail.Focus = 0
			detail.Cost = addCost
			return
		}
		sum1, sum2 := 0.0, 0.0
		for i := range detail.Details {
			tmp1, tmp2 := adjust(detail, &detail.Details[i])
			sum1 += tmp1
			sum2 += tmp2
		}
		parent.Focus -= sum1
		parent.Cost += sum2
		return sum1, sum2
	}
	for i := range cp.Details {
		adjust(cp, &cp.Details[i])
	}
	// fmt.Printf("cp.Comment(): %v\n", cp.Comment())
	*pd = cp.MultiPower(pd.Focus / cp.Focus)
}

func (pd *ProductionDetail) Copy() (copy *ProductionDetail) {
	b, _ := json.Marshal(pd)
	_ = json.Unmarshal(b, &copy)
	return
}

func (pd *ProductionDetail) Comment() string {
	var s string
	if pd.Focus > 0 {
		s = fmt.Sprintf("生产[%s]: %.2f个, 专注总消耗: %.2f, 金币总花费: %.2f\n", pd.ItemName, pd.Cnt, pd.Focus, pd.Cost)
	} else {
		s = fmt.Sprintf("购买[%s]: %.2f个, 金币总花费: %.2f, 参考价格: %.2f\n", pd.ItemName, pd.Cnt, pd.Cost, pd.Cost/pd.Cnt)
	}
	for _, by := range pd.ByProducts {
		s += fmt.Sprintf("副产品[%s]: %.2f个，单价: %.2f\n", by.Name, by.Cnt, getPrice(by.Name, nil))
	}
	if len(pd.Details) > 0 {
		s += "材料情况:\n"
	}

	for _, detail := range pd.Details {
		tmp := detail.Comment()
		arr := strings.Split(strings.TrimSpace(tmp), "\n")
		for i := range arr {
			arr[i] = "  " + arr[i]
		}
		arr = append(arr, "------------------------------------------")
		tmp = strings.Join(arr, "\n")
		s += tmp + "\n"
	}
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "-")
	return string(s)
}

func (pd *ProductionDetail) Equation() (e string) {
	equalize := func(tmp *ProductionDetail) string {
		if tmp.Focus == 0 {
			return fmt.Sprintf("%s(购买，单价%.2f)", tmp.ItemName, tmp.Cost/tmp.Cnt)
		} else {
			return fmt.Sprintf("%s(手搓)", tmp.ItemName)
		}
	}
	lefts := []string{equalize(pd)}
	details := pd.Details
	for _, bp := range pd.ByProducts {
		lefts = append(lefts, fmt.Sprintf("%s(产量%.2f, 专注%.2f, 收入%.2f)", bp.Name, bp.Cnt, pd.Focus, pd.Cost))
	}
	e = strings.Join(lefts, "+")
	for {
		tmp := []ProductionDetail{}
		rights := []string{}
		for _, detail := range details {
			rights = append(rights, equalize(&detail))
			tmp = append(tmp, detail.Details...)
		}
		e += "=" + strings.Join(rights, "+")
		if len(tmp) == 0 {
			break
		}
		details = tmp
	}
	e = strings.TrimRight(e, "=")
	return e
}

func (pd *ProductionDetail) MultiPower(power float64) ProductionDetail {
	pd.Cnt *= power
	pd.Focus *= power
	pd.Cost *= power
	for i := range pd.ByProducts {
		pd.ByProducts[i].Cnt *= power
	}
	if len(pd.Details) == 0 {
		return *pd
	}
	for i := range pd.Details {
		pd.Details[i] = pd.Details[i].MultiPower(power)
	}
	return *pd
}

func (pd *ProductionDetail) AddCost(diff float64) ProductionDetail {
	pd.Cost += diff
	if len(pd.Details) == 0 {
		return *pd
	}
	for i := range pd.Details {
		pd.Details[i] = pd.Details[i].AddCost(diff)
	}
	return *pd
}

func (i *Good) GetMinFocusProducingX(x float64) (production ProductionDetail, err error) {
	if len(i.Recipe) == 0 {
		return production, fmt.Errorf("%s %w", i.Name, ErrNoRecipe)
	}
	productions := make([]ProductionDetail, 0, len(i.Recipe))
	for _, recipe := range i.Recipe {
		p, err := recipe.GetFocusProducingX(x, i.Name)
		if err != nil {
			return production, err
		}
		p.ItemName = i.Name
		productions = append(productions, p)
	}
	return slices.MinFunc(productions, func(a, b ProductionDetail) int {
		return cmp.Compare(a.Focus, b.Focus)
	}), nil
}

func (i *Good) GetMaxCntByFocus(focus float64) (production ProductionDetail, err error) {
	if len(i.Recipe) == 0 {
		return production, fmt.Errorf("%s %w", i.Name, ErrNoRecipe)
	}
	productions := make([]ProductionDetail, 0, len(i.Recipe))
	for _, recipe := range i.Recipe {
		p, err := recipe.GetProductionByFocus(focus, i.Name)
		if err != nil {
			return production, err
		}
		p.ItemName = i.Name
		p.Price = float64(i.Price)
		productions = append(productions, p)
	}

	return slices.MaxFunc(productions, func(a, b ProductionDetail) int {
		return cmp.Compare(a.Cnt, b.Cnt)
	}), nil
}

// GenerateAllFlagCombinations 生成所有满足条件的flag组合
// 1. 根节点的flag固定为false
// 2. 当一个节点的flag为true时，以它为根节点的子树的所有非根节点的flag均为false
func (i *Good) GenerateAllFlagCombinations() []map[string]bool {
	// 结果集合，每个map代表一种可能的flag组合
	result := []map[string]bool{}

	// 如果没有配方，则返回空结果
	if len(i.Recipe) == 0 {
		return result
	}

	// 递归函数，用于生成所有可能的flag组合
	var generateCombinations func(node *Good, currentCombination map[string]bool, parentFlag bool) []map[string]bool
	generateCombinations = func(node *Good, currentCombination map[string]bool, parentFlag bool) []map[string]bool {
		// 如果父节点的flag为true，则当前节点的flag必须为false
		if parentFlag {
			currentCombination[node.Name] = false
			return []map[string]bool{currentCombination}
		}

		// 当父节点flag为false时，当前节点有两种可能：true或false
		combinations := []map[string]bool{}

		// 情况1：当前节点flag为false
		falseCombination := make(map[string]bool)
		for k, v := range currentCombination {
			falseCombination[k] = v
		}
		falseCombination[node.Name] = false

		// 情况2：当前节点flag为true
		trueCombination := make(map[string]bool)
		for k, v := range currentCombination {
			trueCombination[k] = v
		}
		trueCombination[node.Name] = true

		// 如果没有子节点，直接返回当前两种组合
		if len(node.Recipe) == 0 || len(node.Recipe[0].Materials) == 0 {
			combinations = append(combinations, falseCombination, trueCombination)
			return combinations
		}

		// 处理子节点
		// 对于flag=false的情况，递归处理所有子节点
		childCombinations := []map[string]bool{falseCombination}
		for _, material := range node.Recipe[0].Materials {
			childNode, ok := Name2Item[material.Name]
			if !ok {
				continue
			}

			var newCombinations []map[string]bool
			for _, combo := range childCombinations {
				// 为每个现有组合生成子节点的所有可能组合
				childResults := generateCombinations(childNode, combo, false)
				newCombinations = append(newCombinations, childResults...)
			}
			childCombinations = newCombinations
		}
		combinations = append(combinations, childCombinations...)

		// 对于flag=true的情况，所有子节点的flag必须为false
		childCombinations = []map[string]bool{trueCombination}
		for _, material := range node.Recipe[0].Materials {
			childNode, ok := Name2Item[material.Name]
			if !ok {
				continue
			}

			var newCombinations []map[string]bool
			for _, combo := range childCombinations {
				// 为每个现有组合生成子节点的所有可能组合（子节点flag必须为false）
				childResults := generateCombinations(childNode, combo, true)
				newCombinations = append(newCombinations, childResults...)
			}
			childCombinations = newCombinations
		}
		combinations = append(combinations, childCombinations...)

		return combinations
	}

	// 初始组合，根节点的flag固定为false
	initialCombination := map[string]bool{i.Name: false}

	// 从根节点开始生成所有组合
	allCombinations := []map[string]bool{initialCombination}

	// 处理根节点的所有子节点
	for _, material := range i.Recipe[0].Materials {
		childNode, ok := Name2Item[material.Name]
		if !ok {
			continue
		}

		var newCombinations []map[string]bool
		for _, combo := range allCombinations {
			// 为每个现有组合生成子节点的所有可能组合
			childResults := generateCombinations(childNode, combo, false)
			newCombinations = append(newCombinations, childResults...)
		}
		allCombinations = newCombinations
	}

	// 去重：移除重复的组合
	uniqueCombinations := []map[string]bool{}
	combinationExists := make(map[string]bool)

	for _, combination := range allCombinations {
		// 将组合转换为字符串以便比较
		combinationKey := ""
		keys := make([]string, 0, len(combination))
		for k := range combination {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		for _, k := range keys {
			if combination[k] {
				combinationKey += k + ":true,"
			} else {
				combinationKey += k + ":false,"
			}
		}

		// 如果这个组合还没有出现过，则添加到结果中
		if !combinationExists[combinationKey] {
			combinationExists[combinationKey] = true
			uniqueCombinations = append(uniqueCombinations, combination)
		}
	}

	return uniqueCombinations
}

func loadDiskData(reader io.Reader) (goods []*Good, err error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &goods)
	if err != nil {
		return
	}
	if len(goods) > 0 {
		return
	} else {
		return nil, fmt.Errorf("目录 %s 下文件 %s 数据为空", HomeDir, JSON_DATA_FILE)
	}
}
