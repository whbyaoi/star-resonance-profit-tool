package match

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"sort"
	"xhgm_price_tool/good/templates"
)

// Config 配置结构体
type Config struct {
	IgnoreTransparent bool    // 是否忽略透明像素
	Threshold         float64 // 检测阈值
	StepSize          int     // 搜索步长
}

// TemplateStats 模板统计信息
type TemplateStats struct {
	mean        float64
	stdDev      float64
	validPixels int
	pixelValues []float64
	positions   []struct{ x, y int }
}

var defaulThreshold = 0.9

var thresholdMap = map[int]float64{
	1: 0.92,
	0: 0.94,
	2: 0.89,
	4: 0.89,
	5: 0.89,
	3: 0.89,
}

var subMap = map[int]int{
	3: 2,
	5: 2,
	0: 2,
	6: 2,
	7: 2,
	8: 3,
	9: 2,
}

func GetPrice(priceClip image.Image, verbose bool) (price int) {
	config := Config{
		IgnoreTransparent: true, // 忽略透明像素
		Threshold:         defaulThreshold,
		StepSize:          1, // 搜索步长
	}
	arr := [][3]int{} // x, number, score
	for i := range 10 {
		if thre, ok := thresholdMap[i]; ok {
			config.Threshold = thre
		} else {
			config.Threshold = defaulThreshold
		}
		if subMap[i] > 0 {
			results := []struct {
				x, y  int
				score float64
			}{}
			for sub := range subMap[i] {
				if verbose {
					fmt.Printf("static/templates/%d_%d.png\n", i, sub)
				}
				templateImg, err := templates.GetEmbedTemplate(fmt.Sprintf("%d_%d", i, sub)) // 假设模板可能有透明通道
				if err != nil {
					panic(err)
				}
				// robotgo.Save(templateImg, fmt.Sprintf("%d_%d", i, sub)+".png")

				// 预计算模板的统计信息（考虑要忽略的像素）
				templateStats := CalculateTemplateStats(templateImg, config)

				result, _ := FindBestMatch(priceClip, templateImg, templateStats, config)
				if verbose {
					fmt.Printf("result: %+v\n", result)
					// if len(result) > 0 {
					// 	robotgo.Save(outputImage, fmt.Sprintf("%d_tmp.png", i))
					// }
				}
				results = append(results, result...)
			}
			// 对结果进行聚合，误差在2像素内的聚合为一类
			tmp := aggResult(results)
			for _, item := range tmp {
				item[1] = i
				arr = append(arr, item)
			}
			if verbose {
				fmt.Printf("arr: %v\n", arr)
			}
		} else {
			if verbose {
				fmt.Printf("static/templates/%d.png\n", i)
			}
			templateImg, err := templates.GetEmbedTemplate(fmt.Sprint(i)) // 假设模板可能有透明通道
			if err != nil {
				panic(err)
			}
			// 预计算模板的统计信息（考虑要忽略的像素）
			templateStats := CalculateTemplateStats(templateImg, config)

			result, _ := FindBestMatch(priceClip, templateImg, templateStats, config)
			if verbose {
				fmt.Printf("result: %+v\n", result)
				// if len(result) > 0 {
				// 	robotgo.Save(outputImage, fmt.Sprintf("%d_tmp.png", i))
				// }
			}
			// 对结果进行聚合，误差在2像素内的聚合为一类
			tmp := aggResult(result)
			for _, item := range tmp {
				item[1] = i
				arr = append(arr, item)
			}
			if verbose {
				fmt.Printf("arr: %v\n", arr)
			}
		}
	}
	hits := map[int][3]int{}
	find := func(i int) bool {
		return hits[i][0] > 0
	}
	for _, item := range arr {
		if verbose {
			fmt.Printf("item: %v\n", item)
		}
		var existItem [3]int
		var index int = -1
		for offset := range 3 {
			if verbose {
				fmt.Printf("hits: %+v\n", hits)
				fmt.Printf("(item[0] + offset): %v\n", (item[0] + offset))
				fmt.Printf("(item[0] - offset): %v\n", (item[0] - offset))
			}
			if find(item[0] + offset) {
				existItem = hits[item[0]+offset]
				index = item[0] + offset
				break
			} else if find(item[1] - offset) {
				existItem = hits[item[0]-offset]
				index = item[0] - offset
				break
			}
		}
		if index > 0 && item[2] > existItem[2] {
			hits[index] = item
		} else if index == -1 {
			hits[item[0]] = item
		}
	}
	if verbose {
		fmt.Printf("hits: %+v\n", hits)
	}
	arr = [][3]int{}
	for _, item := range hits {
		arr = append(arr, item)
	}
	sort.Slice(arr, func(i, j int) bool {
		return arr[i][0] < arr[j][0]
	})
	for _, tmp := range arr {
		price = price*10 + tmp[1]
	}
	return price
}

func aggResult(result []struct {
	x, y  int
	score float64
}) (arr [][3]int) {
	agg := map[int]struct {
		x, y  int
		score float64
	}{}
	offsetLimit := 3
	for _, row := range result {
		existIndex := -1
		for xOffset := range offsetLimit {
			if _, ok := agg[row.x+xOffset]; ok {
				existIndex = row.x + xOffset
				break
			}
			if _, ok := agg[row.x-xOffset]; ok {
				existIndex = row.x - xOffset
				break
			}
		}
		if existIndex > 0 && row.score > agg[existIndex].score {
			agg[existIndex] = row
		} else if existIndex == -1 {
			agg[row.x] = row
		}
	}
	for key, item := range agg {
		arr = append(arr, [3]int{key, -1, int(item.score * 10000)})
	}
	return arr
}

// CalculateTemplateStats 计算模板统计信息，跳过指定像素
func CalculateTemplateStats(template image.Image, config Config) TemplateStats {
	width, height := template.Bounds().Dx(), template.Bounds().Dy()

	var validPixels []float64
	var validPositions []struct{ x, y int }
	var sum float64

	// 第一次遍历：收集有效像素
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if shouldSkipPixel(template.At(x, y), config) {
				continue
			}

			gray := getGrayValue(template.At(x, y))
			validPixels = append(validPixels, gray)
			validPositions = append(validPositions, struct{ x, y int }{x, y})
			sum += gray
		}
	}

	validCount := len(validPixels)
	if validCount == 0 {
		return TemplateStats{}
	}

	mean := sum / float64(validCount)

	// 计算标准差
	var sumSq float64
	for _, pixel := range validPixels {
		diff := pixel - mean
		sumSq += diff * diff
	}
	stdDev := math.Sqrt(sumSq / float64(validCount))

	return TemplateStats{
		mean:        mean,
		stdDev:      stdDev,
		validPixels: validCount,
		pixelValues: validPixels,
		positions:   validPositions,
	}
}

// shouldSkipPixel 判断是否应该跳过这个像素
func shouldSkipPixel(c color.Color, config Config) bool {
	// 检查透明像素
	if config.IgnoreTransparent {
		_, _, _, a := c.RGBA()
		if a < 0xFFFF { // 不是完全不透明
			return true
		}
	}

	// 检查特定颜色
	// if config.IgnoreColor != nil {
	// 	r1, g1, b1, a1 := c.RGBA()
	// 	r2, g2, b2, a2 := config.IgnoreColor.RGBA()
	// 	if r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2 {
	// 		return true
	// 	}
	// }

	return false
}

// FindBestMatch 查找最佳匹配位置
func FindBestMatch(main, template image.Image, stats TemplateStats, config Config) (result []struct {
	x, y  int
	score float64
}, outputImage image.Image) {
	mW, mH := main.Bounds().Dx(), main.Bounds().Dy()
	// fmt.Println(mW, mH)
	tW, tH := template.Bounds().Dx(), template.Bounds().Dy()

	// 创建输出图像的副本
	outputImage = image.NewRGBA(main.Bounds())
	draw.Draw(outputImage.(*image.RGBA), main.Bounds(), main, image.Point{}, draw.Src)

	// 定义红色边框颜色
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	for y := 0; y <= mH-tH; y += config.StepSize {
		for x := 0; x <= mW-tW; x += config.StepSize {
			score := calculateNCC(main, x, y, stats)
			if score > config.Threshold {
				result = append(result, struct {
					x, y  int
					score float64
				}{x, y, score})

				// 在匹配位置绘制红色边框
				drawRedBorder(outputImage.(*image.RGBA), x, y, tW, tH, red)
			}
		}
	}

	return result, outputImage
}

// drawRedBorder 在指定位置绘制1px红色边框
func drawRedBorder(img *image.RGBA, x, y, width, height int, borderColor color.RGBA) {
	// 绘制上边框
	for i := x; i < x+width; i++ {
		if i >= 0 && i < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(i, y, borderColor)
		}
	}

	// 绘制下边框
	for i := x; i < x+width; i++ {
		if i >= 0 && i < img.Bounds().Dx() && y+height-1 >= 0 && y+height-1 < img.Bounds().Dy() {
			img.Set(i, y+height-1, borderColor)
		}
	}

	// 绘制左边框
	for j := y; j < y+height; j++ {
		if x >= 0 && x < img.Bounds().Dx() && j >= 0 && j < img.Bounds().Dy() {
			img.Set(x, j, borderColor)
		}
	}

	// 绘制右边框
	for j := y; j < y+height; j++ {
		if x+width-1 >= 0 && x+width-1 < img.Bounds().Dx() && j >= 0 && j < img.Bounds().Dy() {
			img.Set(x+width-1, j, borderColor)
		}
	}
}

// calculateNCC 改进的 NCC 计算，只使用有效像素
func calculateNCC(main image.Image, startX, startY int, stats TemplateStats) float64 {
	if stats.validPixels == 0 {
		return 0.0
	}

	// 计算主图像区域的有效像素均值
	regionMean := calculateRegionMean(main, startX, startY, stats.positions)
	// fmt.Printf("regionMean: %v\n", regionMean)

	// 计算协方差和主图像区域的标准差
	var covSum, mainSumSq float64

	for i, pos := range stats.positions {
		templatePixel := stats.pixelValues[i]
		mainPixel := getGrayValue(main.At(startX+pos.x, startY+pos.y))

		// 协方差项
		cov := (templatePixel - stats.mean) * (mainPixel - regionMean)
		covSum += cov

		// 主图像区域的方差项
		diff := mainPixel - regionMean
		mainSumSq += diff * diff
	}

	// 计算主图像区域的标准差
	regionStdDev := math.Sqrt(mainSumSq / float64(stats.validPixels))

	if stats.stdDev == 0 || regionStdDev == 0 {
		return 0.0
	}

	// 计算 NCC
	ncc := covSum / (float64(stats.validPixels) * stats.stdDev * regionStdDev)
	return ncc
}

// calculateRegionMean 计算主图像特定区域的有效像素均值
func calculateRegionMean(main image.Image, startX, startY int, positions []struct{ x, y int }) float64 {
	var sum float64
	for _, pos := range positions {
		sum += getGrayValue(main.At(startX+pos.x, startY+pos.y))
	}
	return sum / float64(len(positions))
}

// getGrayValue 将颜色转换为灰度值
func getGrayValue(c color.Color) float64 {
	// 处理不同类型的颜色模型s
	switch col := c.(type) {
	case color.RGBA:
		return 0.299*float64(col.R) + 0.587*float64(col.G) + 0.114*float64(col.B)
	case color.RGBA64:
		return 0.299*float64(col.R>>8) + 0.587*float64(col.G>>8) + 0.114*float64(col.B>>8)
	case color.NRGBA:
		return 0.299*float64(col.R) + 0.587*float64(col.G) + 0.114*float64(col.B)
	case color.NRGBA64:
		return 0.299*float64(col.R>>8) + 0.587*float64(col.G>>8) + 0.114*float64(col.B>>8)
	case color.Gray:
		return float64(col.Y)
	case color.Gray16:
		return float64(col.Y >> 8)
	default:
		// 通用处理
		r, g, b, a := c.RGBA()
		if a == 0 {
			return 0 // 完全透明
		}
		// 转换为 0-255 范围并计算灰度
		return 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
	}
}

func loadImageViaBytes(b []byte) (img image.Image, err error) {
	reader := bytes.NewReader(b)
	img, err = png.Decode(reader)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// func loadImage(path string) image.Image {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		panic("无法打开图像: " + path)
// 	}
// 	defer file.Close()

// 	// 尝试解码为 JPEG 或 PNG
// 	img, err := jpeg.Decode(file)
// 	if err != nil {
// 		file.Seek(0, 0) // 重置文件指针
// 		img, err = png.Decode(file)
// 		if err != nil {
// 			panic("无法解码图像: " + path)
// 		}
// 	}
// 	return ConvertDarkPixelsToBlack(img, 80)
// }

// ConvertDarkPixelsToBlack 将比指定阈值深的像素转换为黑色
func ConvertDarkPixelsToBlack(img image.Image, threshold uint8) image.Image {
	// 创建新的图像，保持原图尺寸和格式
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// 遍历每个像素
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 获取原图像素颜色
			originalColor := img.At(x, y)
			r, g, b, a := originalColor.RGBA()

			// 将32位颜色值转换为8位
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			// 检查像素是否比阈值深
			if isDarkerThanThreshold(r8, g8, b8, threshold) {
				// 转换为黑色，保持原有透明度
				result.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: a8})
			} else {
				// 保持原色
				result.Set(x, y, originalColor)
			}
		}
	}

	return result
}

// isDarkerThanThreshold 检查颜色是否比阈值深
// 这里使用RGB平均值来判断亮度
func isDarkerThanThreshold(r, g, b, threshold uint8) bool {
	// 计算RGB平均值作为亮度
	brightness := (uint32(r) + uint32(g) + uint32(b)) / 3
	return brightness < uint32(threshold)
}
