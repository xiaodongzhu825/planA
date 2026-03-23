package tool

import (
	"encoding/json"
	"fmt"
	planAType "planA/type"
	"strconv"
	"strings"
)

// ToPtr 使用泛型将任何类型的值转换为指针
func ToPtr[T any](v T) *T {
	return &v
}

// BuildPrice 价格处理
// @param priceMods 价格处理列表
// @param price 价格
// @return int64 处理后的价格
func BuildPrice(priceMods []planAType.PriceMod, price int64) int64 {
	for _, mod := range priceMods {
		if price >= mod.Min && price <= mod.Max {
			newPrice := price * (100 + mod.MarkupRate) / 100
			newPrice += mod.MarkupValue
			return newPrice
		}
	}
	return 0 // 没有匹配到价格模版，直接返回0
}

// ReturnErr 接口返回错误处理
func ReturnErr(logUuid string, taskMsg planAType.TaskBody, typeStr string, err error) (string, error) {
	dataRetBaty, marshalErr := json.Marshal(taskMsg)
	if marshalErr != nil {
		return string(dataRetBaty), fmt.Errorf("[%s] json.Marshal错误: %v", logUuid, marshalErr)
	}
	return string(dataRetBaty), fmt.Errorf("[%s] %v错误: %v", logUuid, typeStr, err)
}

// BuildGoodsName 构建商品名称
// @param goodsNamePrefix 商品名称前缀
// @param goodsNameSuffix 商品名称后缀
// @param titleConsistOf 标题组成
// @param spaceCharacter 间隔符 1=空格
// @param bookInfo 图书信息
// @return string 商品名称
func BuildGoodsName(goodsNamePrefix string, goodsNameSuffix string, titleConsistOf string, spaceCharacter string, bookInfo planAType.BookInfo) string {
	// 解析标题组成
	if titleConsistOf == "" {
		titleConsistOf = "1:true" // 默认使用书名
	}

	// 解析标题组成
	titleOfArr := strings.Split(titleConsistOf, ",")

	// 间隔符
	separator := ""
	if spaceCharacter == "1" {
		separator = " "
	}

	// 构建标题
	title := goodsNamePrefix + separator

	// 遍历标题组成
	for _, item := range titleOfArr {
		// 解析标题组成
		parts := strings.Split(item, ":")
		// 判断是否需要添加标题
		if len(parts) == 2 && parts[1] == "true" {
			switch parts[0] {
			case "0": // ISBN
				title += separator + bookInfo.Isbn
			case "1": // 书名
				title += separator + bookInfo.BookName
			case "2": // 作者
				title += separator + bookInfo.Author
			case "3": // 出版社
				title += separator + bookInfo.Publishing
			case "4": // 出版时间
				title += separator + bookInfo.PublicationDate
			case "5": // 装帧
				title += separator + bookInfo.Binding
			case "6": // 开本
				title += separator + strconv.FormatInt(bookInfo.Format, 10)
			}
		}
	}

	// 添加后缀
	title += separator + goodsNameSuffix

	// 如果标题超过60个字符，截取前60个字符
	if StringLength(title) > 60 {
		title = SubstringByWidth(title, 60)
	}
	//去掉首尾双引号
	title = strings.Trim(title, "\"")
	return title
}

// BuildCarouselGallery 构建轮播图
// @param watermarkImgUrl 水印图片
// @param carouseLastImgUrlArray 最后一张图
// @param carouselUrlArray 轮播图组
// @return []string 轮播图组
func BuildCarouselGallery(watermarkImgUrl string, carouseLastImgUrlArray []string, carouselUrlArray []string) []string {
	// 查看轮播图组长度
	if len(carouselUrlArray)+len(carouseLastImgUrlArray) < 10 {
		length := 10 - (len(carouselUrlArray) + len(carouseLastImgUrlArray))
		// 向轮播图组中添加图片 添加最后一张图片
		if len(carouselUrlArray) > 0 {
			for i := 0; i < length; i++ {
				if carouselUrlArray[len(carouselUrlArray)-1] != "" {
					carouselUrlArray = append(carouselUrlArray, carouselUrlArray[len(carouselUrlArray)-1])
				}
			}
		}
	}
	// 合并数组
	carouselUrlArray = append(carouselUrlArray, carouseLastImgUrlArray...)
	// 循环轮播图组给图片打水印
	for i := 0; i < len(carouselUrlArray); i++ {
		//TODO
		// 给图片打水印
	}
	return carouselUrlArray
}

// BuildDetailGallery 构建详情图
// @param watermarkImgUrl 水印图片
// @param goodsDetailFirstImgUrlArray 商详头图
// @param goodsDetailLastImgUrlArray 商详尾图
// @param detailUrlObject 商详图片
// @return []string 详情图组
func BuildDetailGallery(watermarkImgUrl string, goodsDetailFirstImgUrlArray []string, goodsDetailLastImgUrlArray []string, detailUrlObject planAType.DetailImageObject) []string {
	// 合并数组 简介图+目录图
	imgArr := append(detailUrlObject.IntroductionUrl, detailUrlObject.CatalogueUrl...)
	// 合并数组 简介图+目录图+实拍图
	imgArr = append(imgArr, detailUrlObject.LiveShootingUrl...)
	// 合并数组 简介图+目录图+实拍图+其他图
	imgArr = append(imgArr, detailUrlObject.OtherUrl...)
	// 合并数组 商详头图+简介图+目录图+实拍图+其他图
	imgArr = append(goodsDetailFirstImgUrlArray, imgArr...)
	// 合并数组 商详头图+简介图+目录图+实拍图+其他图+商详尾图
	imgArr = append(imgArr, goodsDetailLastImgUrlArray...)
	// 循环轮播图组给图片打水印
	for i := 0; i < len(imgArr); i++ {
		//TODO
		// 给图片打水印
	}
	return imgArr
}

// BuildGoodsPrice 构建商品价格
// @param bookInfoPrice 图书价格
// @return int64 商品价格
func BuildGoodsPrice(price int64) int64 {
	return price * 4
}

// GoodsAddReturnSuccess 添加商品返回成功处理
func GoodsAddReturnSuccess(taskMsg planAType.TaskBody) (string, error) {
	dataRetBaty, marshalErr := json.Marshal(taskMsg)
	if marshalErr != nil {
		return string(dataRetBaty), fmt.Errorf("json.Marshal错误: %w", marshalErr)
	}
	return string(dataRetBaty), nil
}

// StringLength 计算字符串显示长度
// @param s 字符串
// @return int 字符串显示长度
func StringLength(s string) int {
	length := 0
	for _, r := range s {
		if r > 255 { // 非ASCII字符（如中文）
			length += 2
		} else { // ASCII字符（如英文、数字）
			length += 1
		}
	}
	return length
}

// SubstringByWidth 按显示宽度截取字符串
func SubstringByWidth(s string, maxWidth int) string {
	width := 0
	for i, r := range s {
		if r > 255 {
			width += 2
		} else {
			width += 1
		}

		if width > maxWidth {
			return s[:i] // 返回截取的部分
		}
	}
	return s // 如果整个字符串都不超过maxWidth，返回原字符串
}
