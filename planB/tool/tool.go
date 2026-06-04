package tool

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"planA/planB/initialization/golabl"
	"planA/planB/service"
	planBType "planA/planB/type"
	planBTypeModules "planA/planB/type/modules"
	planAType "planA/type"
	"strconv"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

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
// @param mainImage 主图
// @return []string 详情图组
func BuildDetailGallery(goodsDetailFirstImgUrlArray []string, goodsDetailLastImgUrlArray []string, detailUrlObject planAType.DetailImageObject, mainImage string) []string {
	// 合并数组 简介图+目录图
	imgArr := append(detailUrlObject.IntroductionUrl, detailUrlObject.CatalogueUrl...)
	// 合并数组 简介图+目录图+实拍图
	//imgArr = append(imgArr, detailUrlObject.LiveShootingUrl...)
	// 合并数组 简介图+目录图+实拍图+主图
	imgArr = append(imgArr, mainImage)
	// 合并数组 简介图+目录图+实拍图+主图+其他图
	imgArr = append(imgArr, detailUrlObject.OtherUrl...)
	// 合并数组 商详头图+简介图+目录图+实拍图+主图+其他图
	imgArr = append(goodsDetailFirstImgUrlArray, imgArr...)
	// 合并数组 商详头图+简介图+目录图+实拍图+主图+其他图+商详尾图
	imgArr = append(imgArr, goodsDetailLastImgUrlArray...)
	return imgArr
}

// BuildGoodsPrice 构建商品价格
// @param bookInfoPrice 图书价格
// @return int64 商品价格
func BuildGoodsPrice(price int64) int64 {
	return price * 4
}

// ReturnSuccess 添加商品返回成功处理
func ReturnSuccess(taskMsg planAType.TaskBody) (string, error) {
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
// @param imgUrl 轮播图组
// @param watermarkImgUrl 水印图片
// @param watermarkPosition 水印位置 0 全部 1第一张
// @return []string 轮播图组
// @return error 错误信息
func AddWatermarkFromURLExs(imgUrl []string, watermarkImgUrl string, watermarkPosition string) ([]planBTypeModules.ImageResult, error) {
	var watermarkFromURLExsBase64Arr []planBTypeModules.ImageResult
	// 循环轮播图组给图片打水印
	for i := 0; i < len(imgUrl); i++ {
		var newImgJson string
		var addWatermarkFromURLExsErr error

		// 给图片打水印，带重试机制，最大重试次数为3
		maxRetries := 3
		for retryCount := 0; retryCount <= maxRetries; retryCount++ {

			newImgJson, addWatermarkFromURLExsErr = golabl.ImageDll.AddWatermarkFromURLExs(imgUrl[i], watermarkImgUrl)

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
// @param watermarkFromURLExsBase64Arr 待上传的base64图片列表
// @return []string 图片列表
// @return error 错误信息
func UploadImageToPdd(watermarkFromURLExsBase64Arr []planBTypeModules.ImageResult) ([]string, error) {
	var imageUrlArr []string
	for _, watermarkFromURLExsBase64 := range watermarkFromURLExsBase64Arr {
		var pddImg planBTypeModules.GoodsImageUploadResponse
		imageUrl, pddGoodsImageUploadErr := golabl.PddDll.PddGoodsImageUpload(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, watermarkFromURLExsBase64.Data)
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

// GetPlatformName 获取平台名称
func GetPlatformName() string {
	title := ""
	switch golabl.Task.Header.ShopType {
	//case 2:
	//	return kongFuZi.NewKongfuzi(), nil
	case "1":
		title = title + "拼多多"
	case "5":
		title = title + "闲鱼"
	default:
		title = title + "其他平台 " + golabl.Task.Header.ShopType
	}
	return title
}

// GetTaskType 获取店铺类型
func GetTaskType() string {
	switch golabl.Task.Header.TaskType {
	case 1: //核价发布
		return "核价发布"
	case 2: //表格发布
		return "表格发布"
	case 3: //获取商品
		return "获取商品"
	default:
		return "错误！"
	}
}

// GetWatermarkImg 获取水印图片
// @return string 水印图片 base64
func GetWatermarkImg() (string, error) {
	// 1. 获取日期
	t := time.Unix(golabl.Task.Header.TaskCreateAt, 0)
	yearMonthDay := t.Format("2006-01-02")

	// 2. 获取文件后缀（去除前面的.）
	extWithDot := path.Ext(golabl.Task.Header.ShopMsg.WatermarkImgUrl)
	// 拼接本地存储路径
	imgUrl := "img/watermark/" + yearMonthDay + "/" + golabl.Task.TaskId + extWithDot

	// 3. 判断本地文件是否存在
	if _, err := os.Stat(imgUrl); err == nil {
		// 文件存在 → 直接读取并转base64返回
		imgBytes, err := os.ReadFile(imgUrl)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(imgBytes), nil
	}

	// 4. 文件不存在 → 创建目录、下载图片、保存本地、转base64
	// 创建多级目录
	dirPath := filepath.Dir(imgUrl)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	// 下载远程图片
	resp, err := http.Get(golabl.Task.Header.ShopMsg.WatermarkImgUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 判断下载响应是否成功
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("下载水印图片响应异常: %s\n", resp.Status)
		return "", fmt.Errorf("下载水印图片失败: %s", resp.Status)
	}

	// 读取图片二进制数据
	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取下载的图片数据失败: %v\n", err)
		return "", err
	}

	// 保存图片到本地
	if err := os.WriteFile(imgUrl, imgBytes, 0644); err != nil {
		fmt.Printf("保存水印图片到本地失败: %v\n", err)
		return "", err
	}

	// 5. 转base64返回
	return base64.StdEncoding.EncodeToString(imgBytes), nil
}

// GetSkuWatermarkImg 获取sku水印图片
// @return string 水印图片 base64
func GetSkuWatermarkImg() (string, error) {
	// 1. 获取日期
	t := time.Unix(golabl.Task.Header.TaskCreateAt, 0)
	yearMonthDay := t.Format("2006-01-02")

	// 2. 获取文件后缀（去除前面的.）
	extWithDot := path.Ext(golabl.Task.Header.ShopMsg.SkuWatermarkImgUrl)
	// 拼接本地存储路径
	imgUrl := "img/skuwatermark/" + yearMonthDay + "/" + golabl.Task.TaskId + extWithDot

	// 3. 判断本地文件是否存在
	if _, err := os.Stat(imgUrl); err == nil {
		// 文件存在 → 直接读取并转base64返回
		imgBytes, err := os.ReadFile(imgUrl)
		if err != nil {
			return "", err
		}
		return base64.StdEncoding.EncodeToString(imgBytes), nil
	}

	// 4. 文件不存在 → 创建目录、下载图片、保存本地、转base64
	// 创建多级目录
	dirPath := filepath.Dir(imgUrl)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	// 下载远程图片
	resp, err := http.Get(golabl.Task.Header.ShopMsg.SkuWatermarkImgUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 判断下载响应是否成功
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("下载水印图片响应异常: %s\n", resp.Status)
		return "", fmt.Errorf("下载水印图片失败: %s", resp.Status)
	}

	// 读取图片二进制数据
	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取下载的图片数据失败: %v\n", err)
		return "", err
	}

	// 保存图片到本地
	if err := os.WriteFile(imgUrl, imgBytes, 0644); err != nil {
		fmt.Printf("保存水印图片到本地失败: %v\n", err)
		return "", err
	}

	// 5. 转base64返回
	return base64.StdEncoding.EncodeToString(imgBytes), nil
}

// UpdateTaskHeader 更新头部信息
// @return error 错误信息
func UpdateTaskHeader() error {
	//通过 footer 来更新 header 的计数
	golabl.Task.Header.TaskCountWait = golabl.Task.Footer.TaskCountWait.Load()
	golabl.Task.Header.TaskCountOver = golabl.Task.Footer.TaskCountOver.Load()
	golabl.Task.Header.TaskCountSuccess = golabl.Task.Footer.TaskCountSuccess.Load()
	golabl.Task.Header.TaskCountError = golabl.Task.Footer.TaskCountError.Load()
	golabl.Task.Header.LastIndex = golabl.Logic.LastIndex
	return service.UpdateTaskHeaderCount()
}

// UpdateTaskProgress 更新拉取商品进度
// @param con int64 更新进度数
// @return error 错误信息
func UpdateTaskProgress(con int64) error {
	// 更新 进度
	if updateTaskFooterErr := service.UpdateTaskFooter(1, con); updateTaskFooterErr != nil {
		return updateTaskFooterErr
	}
	// 重新获取 footer信息
	if getTaskFooterErr := service.GetTaskFooter(); getTaskFooterErr != nil {
		return getTaskFooterErr
	}
	if updateTaskHeaderCountErr := UpdateTaskHeader(); updateTaskHeaderCountErr != nil {
		return updateTaskHeaderCountErr
	}
	return nil
}

// FenToYuan 将金额从分转换为元
// 参数：fen - 金额（分），int64类型
// 返回值：金额（元），string类型
func FenToYuan(fen int64) string {
	yuan := float64(fen) / 100.0
	return fmt.Sprintf("%.2f", yuan)
}

// FenToYuanFloat64 将金额从分转换为元
// 参数：fen - 金额（分），int64类型
// 返回值：金额（元），float64类型
func FenToYuanFloat64(fen int64) float64 {
	return float64(fen) / 100.0
}

// IsShopIDExists 判断店铺ID是否存在于shopId.txt文件中
func IsShopIDExists(targetShopID string) (bool, error) {
	// 打开文件
	file, err := os.Open("shopId.txt")
	if err != nil {
		return false, fmt.Errorf("无法打开文件: %w", err)
	}
	defer file.Close()

	// 逐行扫描
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text()) // 去除空格和换行符
		// 忽略空行
		if line == "" {
			continue
		}
		if line == targetShopID {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("读取文件出错: %w", err)
	}

	return false, nil
}

// AppendTextToFile 在文件中追加文本
// @param filePath 文件路径
// @param text 要追加的文本
// @return error 错误信息
func AppendTextToFile(filePath string, text string) error {
	// 获取文件所在的目录
	dir := filepath.Dir(filePath)

	// 创建目录（如果不存在）
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 以追加模式打开文件，如果文件不存在则创建
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 写入文本并添加换行符
	if _, err := file.WriteString(text + "\n"); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// UploadImageToKfz 将图片上传到孔夫子
// @param watermarkFromURLExsBase64Arr 待上传的base64图片列表
// @return []string 图片列表
// @return error 错误信息
func UploadImageToKfz(watermarkFromURLExsBase64Arr []planBTypeModules.ImageResult) ([]string, error) {

	var imageUrlArr []string
	for _, watermarkFromURLExsBase64 := range watermarkFromURLExsBase64Arr {
		//将图片保存到本地
		imgTempUrl, saveBase64ImageByDateErr := SaveBase64ImageByDate(watermarkFromURLExsBase64.Data, golabl.Config.FileUrl.KfzImgTempUrl)
		if saveBase64ImageByDateErr != nil {
			return nil, saveBase64ImageByDateErr
		}
		//将图片上传到孔夫子
		_, kfzGoodsImageUploadErr := golabl.KfzDll.KfzGoodsImageUpload(golabl.Config.KfzConfig.AppId, golabl.Config.KfzConfig.AppSecret, golabl.Task.Header.ShopMsg.Token, imgTempUrl)
		if kfzGoodsImageUploadErr != nil {
			return nil, kfzGoodsImageUploadErr
		}
	}
	return imageUrlArr, nil
}

// SaveBase64ImageByDate 保存base64图片或下载URL图片到按年月日组织的文件夹中
// 参数1: input - base64编码的图片字符串 或 网络图片地址
// 参数2: basePath - 基础路径（如 D:\\file\\kfzImg）
// 返回: 保存的完整图片地址和错误信息
func SaveBase64ImageByDate(input string, url string) (string, error) {
	// 去除首尾空格
	input = strings.TrimSpace(input)

	var imageData []byte
	var err error

	// 判断是否为URL（以http://或https://开头）
	if strings.HasPrefix(strings.ToLower(input), "http://") ||
		strings.HasPrefix(strings.ToLower(input), "https://") {
		// 处理URL情况：下载图片
		imageData, err = downloadImage(input)
		if err != nil {
			return "", fmt.Errorf("下载图片失败: %v", err)
		}
	} else {
		// 处理base64情况
		imageData, err = decodeBase64Image(input)
		if err != nil {
			return "", fmt.Errorf("base64解码失败: %v", err)
		}
	}

	// 生成按年月日的文件夹路径
	now := time.Now()
	dateFolder := now.Format("2006-01-02") // 格式: 2026-05-11
	saveDir := filepath.Join(url, dateFolder)

	// 创建目录（如果不存在）
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成唯一文件名（使用时间戳+随机数避免重名）
	timestamp := now.UnixNano()
	filename := fmt.Sprintf("%d.png", timestamp)
	savePath := filepath.Join(saveDir, filename)

	// 写入文件
	err = os.WriteFile(savePath, imageData, 0644)
	if err != nil {
		return "", fmt.Errorf("保存图片失败: %v", err)
	}

	return savePath, nil
}

// decodeBase64Image 解码base64图片
func decodeBase64Image(base64Str string) ([]byte, error) {
	// 去除可能存在的base64头部信息
	if idx := strings.Index(base64Str, ","); idx != -1 {
		base64Str = base64Str[idx+1:]
	}

	// 解码base64字符串
	imageData, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}

// downloadImage 下载网络图片
func downloadImage(url string) ([]byte, error) {
	// 创建HTTP客户端（设置超时时间）
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发送GET请求
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码异常: %d", resp.StatusCode)
	}

	// 检查Content-Type是否为图片
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("非图片资源: Content-Type=%s", contentType)
	}

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取图片数据失败: %v", err)
	}

	return imageData, nil
}

// ProcessImage 处理图片
func ProcessImage(imageURL string, saveBase64 bool, saveImage bool) (string, error) {
	img, format, err := GetImageFromURL(imageURL)
	if err != nil {
		return "", err
	}

	bounds := img.Bounds()
	fmt.Printf("原始尺寸: %dx%d\n", bounds.Dx(), bounds.Dy())

	var processedImg image.Image
	if bounds.Dx() == 800 && bounds.Dy() == 800 {
		fmt.Println("已经是800x800，无需缩放")
		processedImg = img
	} else {
		fmt.Println("缩放到800x800")
		processedImg = ResizeImageHighQuality(img, 800, 800)
	}

	// 转换为base64
	base64Str, err := ImageToBase64(processedImg, format)
	if err != nil {
		return "", fmt.Errorf("转换为base64失败: %v", err)
	}

	return base64Str, nil
}

// ResizeImageHighQuality 高质量缩放图片
func ResizeImageHighQuality(img image.Image, width, height uint) image.Image {
	return resize.Resize(width, height, img, resize.Lanczos3)
}

// GetImageFromURL 从URL获取图片
func GetImageFromURL(url string) (image.Image, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// 尝试解码PNG
	img, err := png.Decode(strings.NewReader(string(data)))
	if err == nil {
		return img, "png", nil
	}

	// 尝试解码JPEG
	img, err = jpeg.Decode(strings.NewReader(string(data)))
	if err == nil {
		return img, "jpg", nil
	}

	return nil, "", fmt.Errorf("不支持的图片格式")
}

// ImageToBase64 将图片转换为base64
func ImageToBase64(img image.Image, format string) (string, error) {
	buf := new(strings.Builder)

	if format == "png" {
		err := png.Encode(buf, img)
		if err != nil {
			return "", err
		}
	} else {
		err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 95})
		if err != nil {
			return "", err
		}
	}

	base64Str := base64.StdEncoding.EncodeToString([]byte(buf.String()))
	return fmt.Sprintf("data:image/%s;base64,%s", format, base64Str), nil
}

// GetGoodsByShopIdAndIsbn 根据店铺id与isbn获取商品
func GetGoodsByShopIdAndIsbn(shopId, isbn string) (planBType.GetShopGoodsByShopIdAndIsbn, error) {

	var getShopGoodsByShopIdAndIsbn planBType.GetShopGoodsByShopIdAndIsbn

	params := map[string]string{
		"shopId": shopId,
		"isbn":   isbn,
	}
	withParams, buildURLWithParamsErr := BuildURLWithParams(golabl.Config.FileUrl.GetPddGoodsShopIdIsbnUrl, params)
	if buildURLWithParamsErr != nil {
		return getShopGoodsByShopIdAndIsbn, buildURLWithParamsErr
	}
	_, resStr, httpGetRequestErr := HttpGetRequest(withParams)
	if httpGetRequestErr != nil {
		return getShopGoodsByShopIdAndIsbn, httpGetRequestErr
	}
	unmarshalErr := json.Unmarshal([]byte(resStr), &getShopGoodsByShopIdAndIsbn)
	if unmarshalErr != nil {
		return getShopGoodsByShopIdAndIsbn, unmarshalErr
	}
	return getShopGoodsByShopIdAndIsbn, nil
}
