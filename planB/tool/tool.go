package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"planA/modules/logs"
	_type "planA/planB/type"
	"strconv"
	"strings"
	"unicode"
)

// CheckContext 检查上下文是否取消
func CheckContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err() // 返回取消原因
	default:
		return nil // 上下文仍然有效
	}
}

// CalculateStringLength 计算字符串长度
// @param text 字符串
// @return int 字符串长度
func CalculateStringLength(text string) int {
	length := 0
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			// 汉字算2个字符
			length += 2
		} else {
			// 其他字符（符号、数字、字母等）算1个字符
			length += 1
		}
	}

	return length
}

// SubstringByCharLength 根据字符长度截取字符串
// @param text 字符串
// @param maxLength 最大字符长度
// @return string 截取后的字符串
// @return int 截取后的字符串实际占用的长度（按上述规则计算）
// @return bool 标识是否发生了截取：true=发生了截取，false=没有发生截取
func SubstringByCharLength(text string, maxLength int) (string, int, bool) {
	if maxLength <= 0 {
		return "", 0, false
	}

	length := 0
	truncated := false
	resultRunes := []rune{}

	for _, r := range text {
		charLength := 1
		if unicode.Is(unicode.Han, r) {
			charLength = 2
		}

		// 检查添加当前字符是否会超过最大长度
		if length+charLength > maxLength {
			truncated = true
			break
		}

		resultRunes = append(resultRunes, r)
		length += charLength
	}

	return string(resultRunes), length, truncated
}

// 判断是否为英文字符（字母、数字、英文符号）
func isEnglishChar(r rune) bool {
	// 检查是否为ASCII字符（英文、数字、基本符号）
	if r < 128 {
		return true
	}

	// 检查是否为英文字母（包括带重音的字母）
	return unicode.Is(unicode.Latin, r)
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

// 计算字符串显示长度
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

// GoodsAddReturnSuccess 添加商品返回成功处理
func GoodsAddReturnSuccess(taskMsg _type.TaskBody) (string, error) {
	dataRetBaty, marshalErr := json.Marshal(taskMsg)
	if marshalErr != nil {
		return string(dataRetBaty), fmt.Errorf("json.Marshal错误: %w", marshalErr)
	}
	return string(dataRetBaty), nil
}

// HttpBannedWordSubstitution 违禁词处理
func HttpBannedWordSubstitution(url string, reqData map[string]string) (_type.HttpBannedWordSubstitutionBookInfoRes, error) {
	var resDta _type.HttpBannedWordSubstitutionBookInfoRes

	// 构建带参数的 URL
	reqUrl, err := BuildURLWithParams(url, reqData)
	if err != nil {
		return resDta, fmt.Errorf("构建URL失败: %v", err)
	}

	// 发送 GET请求
	_, resStr, httpGetRequestErr := HttpGetRequest(reqUrl)

	if httpGetRequestErr != nil {
		return resDta, httpGetRequestErr
	}

	// 将字符串转换为结构体
	jsonErr := json.Unmarshal([]byte(resStr), &resDta)
	if jsonErr != nil {
		return resDta, jsonErr
	}

	if resDta.Code != "200" {
		return resDta, fmt.Errorf("请求违禁词接口错误 错误: url %s %s", reqUrl, resStr)
	}
	// 返回结果
	return resDta, nil
}

// BuildURLWithParams 将map参数拼接到URL后面
func BuildURLWithParams(baseURL string, params map[string]string) (string, error) {
	if len(params) == 0 {
		return baseURL, nil
	}

	// 解析基础URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("解析URL失败: %v", err)
	}

	// 获取现有的查询参数
	query := parsedURL.Query()

	// 添加新的参数
	for key, value := range params {
		query.Set(key, value)
	}
	// 重新编码查询参数
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// ReturnErr 接口返回错误处理
func ReturnErr(logUuid string, taskMsg _type.TaskBody, typeStr _type.GoodsType, err error) (string, error) {
	dataRetBaty, marshalErr := json.Marshal(taskMsg)
	if marshalErr != nil {
		errMsg := fmt.Sprintf("[%s] json.Marshal错误: %v", logUuid, marshalErr)
		logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
		return string(dataRetBaty), fmt.Errorf(errMsg)
	}
	errMsg := fmt.Sprintf("[%s] %v错误: %v", logUuid, typeStr, err)
	logs.LoggingMiddleware(logs.LOG_LEVEL_ERROR, errMsg)
	return string(dataRetBaty), err
}

// BuildDetailGallery 构建详情图
// @param watermarkImgUrl 水印图片
// @param goodsDetailFirstImgUrlArray 商详头图
// @param goodsDetailLastImgUrlArray 商详尾图
// @param detailUrlObject 商详图片
// @return []string 详情图组
func BuildDetailGallery(watermarkImgUrl string, goodsDetailFirstImgUrlArray []string, goodsDetailLastImgUrlArray []string, detailUrlObject _type.DetailImageObject) []string {
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

// BuildGoodsName 构建商品名称
// @param goodsNamePrefix 商品名称前缀
// @param goodsNameSuffix 商品名称后缀
// @param titleConsistOf 标题组成
// @param spaceCharacter 间隔符 1=空格
// @param bookInfo 图书信息
// @return string 商品名称
func BuildGoodsName(goodsNamePrefix string, goodsNameSuffix string, titleConsistOf string, spaceCharacter string, bookInfo _type.BookInfo) string {
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

// PauseTask 暂停B程序运行
// @param url 暂停接口地址
// @param taskId 任务ID
// @return error 错误信息
func PauseTask(url string, taskId string) error {
	_, _, err := HttpGetRequest(url + "/task/pause/" + taskId)
	fmt.Println(err)
	return err
}
