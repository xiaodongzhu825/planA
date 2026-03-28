package tool

import (
	"encoding/json"
	"fmt"
	"planA/planB/initialization/golabl"
	"planA/planB/modules/image"
	"planA/planB/modules/pdd"
	planBTypeModules "planA/planB/type/modules"
	planAType "planA/type"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
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
// @param carouseLastImgUrlArray 最后一张图
// @param oldCarouselUrlArray 旧轮播图
// @param carouselUrlArray 轮播图组
// @param watermarkPosition 水印位置 0 全部 1第一张
// @return []string 轮播图组
func BuildCarouselGallery(carouseLastImgUrlArray []string, oldCarouselUrlArray []string, carouselUrlArray []string, watermarkPosition string) []string {
	// 查看轮播图组长度
	if len(carouselUrlArray)+len(carouseLastImgUrlArray) < 10 {
		length := 10 - (len(carouselUrlArray) + len(carouseLastImgUrlArray))
		// 向轮播图组中添加图片 添加最后一张图片
		if len(carouselUrlArray) > 0 {
			for i := 0; i < length; i++ {
				if carouselUrlArray[len(carouselUrlArray)-1] != "" {
					if watermarkPosition == "1" {
						// 使用不打水印的图片补充
						carouselUrlArray = append(carouselUrlArray, oldCarouselUrlArray[len(oldCarouselUrlArray)-1])
					} else {
						// 使用打水印的图片补充
						carouselUrlArray = append(carouselUrlArray, carouselUrlArray[len(carouselUrlArray)-1])
					}
				}
			}
		}
	}
	// 合并数组
	carouselUrlArray = append(carouselUrlArray, carouseLastImgUrlArray...)

	return carouselUrlArray
}

// BuildCarouselGalleryOld 构建轮播图
// @param carouseLastImgUrlArray 最后一张图
// @param carouselUrlArray 轮播图组
// @return []string 轮播图组
func BuildCarouselGalleryOld(carouseLastImgUrlArray []string, carouselUrlArray []string) []string {
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

	return carouselUrlArray
}

// BuildDetailGallery 构建详情图
// @param goodsDetailFirstImgUrlArray 商详头图
// @param goodsDetailLastImgUrlArray 商详尾图
// @param detailUrlObject 商详图片
// @return []string 详情图组
func BuildDetailGallery(goodsDetailFirstImgUrlArray []string, goodsDetailLastImgUrlArray []string, detailUrlObject planAType.DetailImageObject) []string {
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

// FilterWord 违规词处理
// @param taskMsg 任务信息
func FilterWord(taskMsg *planAType.TaskBody) error {
	substitution, httpBannedWordSubstitutionErr := HttpFilterWord(taskMsg.BookInfo.Isbn, taskMsg.BookInfo.BookName, taskMsg.BookInfo.Author, taskMsg.BookInfo.Publishing)
	if httpBannedWordSubstitutionErr != nil {
		return fmt.Errorf("HttpFilterWord 违禁词处理失败-原因来自:%v", httpBannedWordSubstitutionErr)
	}
	if golabl.Config.Server.ReplaceMark == "0" && len(substitution.Data) > 0 {
		errMsg := "违规词命中 "
		for _, v := range substitution.Data {
			errMsg = errMsg + " " + v.AddTxt + "(" + v.MatchType + ") "
		}
		return fmt.Errorf(errMsg)
	}
	if golabl.Config.Server.ReplaceMark == "1" && len(substitution.Data) > 0 {
		//替换违禁词
		taskMsg.BookInfo.BookName = substitution.BookName
		taskMsg.BookInfo.Author = substitution.Author
		taskMsg.BookInfo.Publishing = substitution.Publisher
		taskMsg.BookInfo.Isbn = substitution.Isbn
	}

	return nil
}

// AddWatermarkFromURLExs 打水印
// @param imageDll 图片处理DLL
// @param carouselUrlArray 轮播图组
// @param watermarkImgUrl 水印图片
// @param watermarkPosition 水印位置 0 全部 1第一张
// @return []string 轮播图组
// @return error 错误信息
func AddWatermarkFromURLExs(imageDll *image.ImageDLL, imgUrl []string, watermarkImgUrl string, watermarkPosition string) ([]planBTypeModules.ImageResult, error) {
	var watermarkFromURLExsBase64Arr []planBTypeModules.ImageResult
	// 循环轮播图组给图片打水印
	for i := 0; i < len(imgUrl); i++ {
		var newImgJson string
		var addWatermarkFromURLExsErr error

		// 给图片打水印，带重试机制，最大重试次数为3
		maxRetries := 3
		for retryCount := 0; retryCount <= maxRetries; retryCount++ {
			newImgJson, addWatermarkFromURLExsErr = imageDll.AddWatermarkFromURLExs(imgUrl[i], watermarkImgUrl)

			// 判断是否包含超时错误
			if addWatermarkFromURLExsErr != nil && strings.Contains(addWatermarkFromURLExsErr.Error(), "dialing to the given TCP address timed out") {
				if retryCount < maxRetries {
					// 重试前等待一段时间（可选）
					time.Sleep(time.Duration(retryCount+1) * time.Second)
					continue
				}
			}
			// 如果没有错误或者不是超时错误，跳出重试循环
			break
		}

		if addWatermarkFromURLExsErr != nil {
			return watermarkFromURLExsBase64Arr, fmt.Errorf("给图片打水印错误 %w", addWatermarkFromURLExsErr)
		}

		// 将 newImg 转为结构体
		var newImg planBTypeModules.ImageResult
		unmarshalErr := json.Unmarshal([]byte(newImgJson), &newImg)
		if unmarshalErr != nil {
			return nil, fmt.Errorf("解析失败 %w 原始数据 %v", unmarshalErr, newImgJson)
		}
		watermarkFromURLExsBase64Arr = append(watermarkFromURLExsBase64Arr, newImg)

		if watermarkPosition == "1" {
			break
		}
	}
	return watermarkFromURLExsBase64Arr, nil
}

// UploadImageToPdd 将图片上传到拼多多
// @param pddDll pddDLL对象
// @param watermarkFromURLExsBase64Arr 待上传的base64图片列表
// @return []string 图片列表
// @return error 错误信息
func UploadImageToPdd(pddDll *pdd.PddDLL, watermarkFromURLExsBase64Arr []planBTypeModules.ImageResult) ([]string, error) {
	var imageUrlArr []string
	for _, watermarkFromURLExsBase64 := range watermarkFromURLExsBase64Arr {
		var pddImg planBTypeModules.GoodsImageUploadResponse
		imageUrl, pddGoodsImageUploadErr := pddDll.PddGoodsImageUpload(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, watermarkFromURLExsBase64.Data)
		if pddGoodsImageUploadErr != nil {
			return imageUrlArr, pddGoodsImageUploadErr
		}
		// 解析 JSON字符串
		unmarshalErr := json.Unmarshal([]byte(imageUrl), &pddImg)
		if unmarshalErr != nil {
			return imageUrlArr, fmt.Errorf("解析拼多多 PddGoodsImageUpload 错误: %v [拼多多数据：%v]", unmarshalErr, imageUrl)
		}
		imageUrlArr = append(imageUrlArr, pddImg.GoodsImageUploadResponse.ImageURL)
	}
	return imageUrlArr, nil
}

// SetConsoleTitle 设置窗口标题
// @param title 标题
func SetConsoleTitle(title string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleTitle := kernel32.NewProc("SetConsoleTitleW")
	// 将字符串转换为UTF-16指针
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	procSetConsoleTitle.Call(uintptr(unsafe.Pointer(titlePtr)))
}
