package taobao

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"planA/planB/initialization/golabl"
	"planA/planB/modules/logs"
	"planA/planB/service"
	"planA/planB/tool"
	planAType "planA/type"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Taobao 淘宝平台结构体
type Taobao struct {
	client *http.Client
}

// NewTaobao 创建淘宝平台
// @return *Taobao
func NewTaobao() *Taobao {
	return &Taobao{
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// AddGoodsTask 添加商品
// @param taskMsg 任务内容
// @return string body 信息
// @return error 错误
func (t *Taobao) AddGoodsTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}
	taskMsg, publishGoodsErr := t.publishGoods(logUuid, taskMsg)
	if publishGoodsErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, publishGoodsErr)
	}
	return tool.ReturnSuccess(taskMsg)
}

// GetGoodsTask 获取商品
// @return string body 信息
// @return error 错误
func (t *Taobao) GetGoodsTask() (string, error) {
	// 生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}

	const pageSize = 200
	const maxPage = 50

	var totalFetched int
	var duplicateCount int
	var uniqueCount int

	// 第一阶段：只拉取任务数据，更新总数，不写入wait
	firstTimeGoodsErr := t.phaseOneGoodsOnlyCount(pageSize, maxPage)
	if firstTimeGoodsErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, firstTimeGoodsErr)
	}

	// 查询body_wait是否存在，确定第二阶段的开始时间
	exist, isTaskBodyWaitExistErr := service.IsTaskBodyWaitExist()
	if isTaskBodyWaitExistErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, isTaskBodyWaitExistErr)
	}

	var startPage int
	if exist {
		// 获取body_wait数量，计算应该从第几页开始
		bodyWaitCount, getCountErr := service.GetTaskBodyWaitCount()
		if getCountErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, getCountErr)
		}
		startPage = int(bodyWaitCount/pageSize) + 1
		fmt.Printf("body_wait已存在 %d 条数据，从第 %d 页开始拉取\n", bodyWaitCount, startPage)
	} else {
		startPage = 1
	}

	// 第二阶段：获取商品（写入wait）
	phaseTwoGoodsErr := t.phaseTwoGoods(startPage, maxPage, pageSize, &totalFetched)
	if phaseTwoGoodsErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, phaseTwoGoodsErr)
	}

	// 更新状态为推送中
	updateTaskStatusErr := service.UpdateTaskStatus(planAType.TaskStatusPushTaskStatus)
	if updateTaskStatusErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, updateTaskStatusErr)
	}

	// 重新设置任务进度
	if updateTaskHeaderErr := service.SetTaskCount(strconv.FormatInt(golabl.Task.Footer.TaskCountTrue, 10)); updateTaskHeaderErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, updateTaskHeaderErr)
	}

	// 去重复与保存
	deduplicateToBodyOverErr := t.deduplicateToBodyOver(&duplicateCount, &uniqueCount)
	if deduplicateToBodyOverErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, deduplicateToBodyOverErr)
	}

	// 输出统计信息
	statsLogMsg := fmt.Sprintf(`
════════════════════════════════════════════════════════════════
【淘宝店铺拉取】
请求ID：%s
时间: %s
店铺ID：%v
店铺名称：%v
总共获取商品数（含重复）: %d
不重复商品数: %d
重复商品数: %d
重复率: %.2f%%
════════════════════════════════════════════════════════════════`,
		logUuid,
		time.Now().Format("2006-01-02 15:04:05.000"),
		golabl.Task.TaskId,
		golabl.Task.Header.ShopName,
		totalFetched,
		uniqueCount,
		duplicateCount,
		float64(duplicateCount)/float64(totalFetched)*100)
	fmt.Println(statsLogMsg)
	tool.LoggingMiddleware(logs.LOG_LEVEL_INFO, statsLogMsg)

	return tool.ReturnSuccess(planAType.TaskBody{})
}

// OperationGoodsTask 操作商品
// @param taskMsg 任务内容
// @return string body 信息
// @return error 错误
func (t *Taobao) OperationGoodsTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}

	//暂停 2 秒
	time.Sleep(2 * time.Second)

	switch taskMsg.Detail.Status {
	case 1:
		return t.upOne(logUuid, taskMsg) //上架
	case 2:
		return t.downOne(logUuid, taskMsg) //下架
	case 4:
		return t.updateStock(logUuid, taskMsg) //修改库存
	case 5:
		return t.updatePrice(logUuid, taskMsg) //修改价格
	case 6:
		//发布商品
		taskMsg, publishGoodsErr := t.publishGoods(logUuid, taskMsg)
		if publishGoodsErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, publishGoodsErr)
		}
		return tool.ReturnSuccess(taskMsg)
	case 7:
		//下架
		_, downShelfErr := t.downOne(logUuid, taskMsg)
		if downShelfErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, downShelfErr)
		}
		//发布商品
		taskMsg, publishGoodsErr := t.publishGoods(logUuid, taskMsg)
		if publishGoodsErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, publishGoodsErr)
		}
		return tool.ReturnSuccess(taskMsg)
	default:
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("未知操作类型 %v", taskMsg.Detail.Status))
	}
}

// IncStockTask 增量库存
// @param taskMsg 任务内容
// @return string body 信息
// @return error 错误
func (t *Taobao) IncStockTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}

	// 获取商品id
	getGoodsByShopIdAndIsbn, GetGoodsByShopIdAndIsbnErr := tool.GetGoodsByShopIdAndIsbn(golabl.Task.Header.ShopId, taskMsg.BookInfo.Isbn)
	if GetGoodsByShopIdAndIsbnErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, GetGoodsByShopIdAndIsbnErr)
	}
	if getGoodsByShopIdAndIsbn.Code != "200" {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("ERP未找到商品"))
	}
	if len(getGoodsByShopIdAndIsbn.Data) == 0 {
		//新发布
		task, addGoodsTaskErr := t.publishGoods(logUuid, taskMsg)
		if addGoodsTaskErr != nil {
			return "", addGoodsTaskErr
		}
		return tool.ReturnSuccess(task)
	} else {
		// 将 getGoodsByShopIdAndIsbn.Data[0].TrilateralId 转为 int64
		trilateralId, trilateralIdParseIntErr := strconv.ParseInt(getGoodsByShopIdAndIsbn.Data[0].TrilateralId, 10, 64)
		if trilateralIdParseIntErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, trilateralIdParseIntErr)
		}

		// 将 getGoodsByShopIdAndIsbn.Data[0].Stock 转为 int64
		stock, stockParseIntErr := strconv.ParseInt(getGoodsByShopIdAndIsbn.Data[0].Stock, 10, 64)
		if stockParseIntErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, stockParseIntErr)
		}

		//增量修改库存
		taskMsg.Detail.GoodsId = trilateralId
		taskMsg.Detail.Stock = taskMsg.Detail.Stock + stock
		return t.updateStock(logUuid, taskMsg)
	}
}

func (t *Taobao) SetGoodsTask() string {
	return ""
}

// *******************************私有方法************************************ //

// publishGoods 发布商品核心逻辑
func (t *Taobao) publishGoods(logUuid string, taskMsg planAType.TaskBody) (planAType.TaskBody, error) {
	// 价格不能小于0
	if taskMsg.Detail.Price <= 0 {
		return taskMsg, fmt.Errorf("价格不能小于等于0")
	}

	//获取出版社信息并解析
	if getPublishingErr := service.GetPublishingVid(&taskMsg); getPublishingErr != nil {
		return taskMsg, fmt.Errorf("获取出版社信息失败-原因来自:%v", getPublishingErr)
	}

	//违规词处理
	if golabl.Config.Server.Filter == 1 {
		if taskMsgErr := tool.FilterWord(&taskMsg); taskMsgErr != nil {
			return taskMsg, taskMsgErr
		}
	}

	//价格 + 运费
	if golabl.Task.Header.PriceType != "0" {
		taskMsg.Detail.Price = taskMsg.Detail.Price + taskMsg.Detail.ShippingCost
	}

	// 价格处理
	price := tool.BuildPrice(golabl.Task.Header.PriceMod, taskMsg.Detail.Price)
	if price == 0 {
		return taskMsg, fmt.Errorf("不在价格区间内 isbn %v  原始价格 %v  当前价格 %v 价格模版 %v", taskMsg.BookInfo.Isbn, taskMsg.Detail.Price, price, golabl.Task.Header.PriceMod)
	}
	taskMsg.Detail.Price = price

	//构建商品名称
	goodsName := tool.BuildGoodsName(
		golabl.Task.Header.ShopMsg.GoodsNamePrefix,
		golabl.Task.Header.ShopMsg.GoodsNameSuffix,
		golabl.Task.Header.ShopMsg.TitleConsistOf,
		golabl.Task.Header.ShopMsg.SpaceCharacter,
		taskMsg.BookInfo)
	taskMsg.Detail.GoodsName = goodsName

	// 检查轮播图
	if len(taskMsg.BookInfo.ImageObject.CarouselUrlArray) == 0 {
		setNoImgCountErr := service.SetNoImgCount(taskMsg.BookInfo.Isbn)
		if setNoImgCountErr != nil {
			return taskMsg, fmt.Errorf("无图片信息isbn计次错误 isbn %v %v", taskMsg.BookInfo.Isbn, setNoImgCountErr.Error())
		}
		return taskMsg, fmt.Errorf("缺少轮播图")
	}

	oldCarouselUrlArray := append([]string{}, taskMsg.BookInfo.ImageObject.CarouselUrlArray...)

	// 存在水印图片，则打水印
	var mainImgLocalPath string
	if golabl.Task.Header.ShopMsg.WatermarkImgUrl != "" {
		watermarkImgUrl, watermarkImgErr := tool.GetWatermarkImg()
		if watermarkImgErr != nil {
			return taskMsg, fmt.Errorf("获取水印图片失败 %v", watermarkImgErr)
		}
		watermarkFromURLExsBase64Arr, watermarkFromURLExsErr := tool.AddWatermarkFromURLExs(taskMsg.BookInfo.ImageObject.CarouselUrlArray, watermarkImgUrl, golabl.Task.Header.ShopMsg.WatermarkPosition)
		if watermarkFromURLExsErr != nil {
			return taskMsg, fmt.Errorf("图片打水印失败 %v", watermarkFromURLExsErr)
		}
		// 第一张作为主图，保存到本地
		if len(watermarkFromURLExsBase64Arr) > 0 {
			// 从 ImageResult 中提取 base64 数据
			base64Data := watermarkFromURLExsBase64Arr[0].Data
			// 保存到本地
			savedPath, saveErr := tool.SaveBase64ImageByDate(base64Data, golabl.Config.TaobaoConfig.LocalImgDir)
			if saveErr != nil {
				return taskMsg, fmt.Errorf("保存水印图片到本地失败 %v", saveErr)
			}
			mainImgLocalPath = savedPath
		}
	} else {
		// 没有水印，直接下载原图到本地
		savedPath, saveErr := tool.SaveBase64ImageByDate(oldCarouselUrlArray[0], golabl.Config.TaobaoConfig.LocalImgDir)
		if saveErr != nil {
			return taskMsg, fmt.Errorf("保存原图到本地失败 %v", saveErr)
		}
		mainImgLocalPath = savedPath
	}

	// 上传到淘宝图片空间
	mainImg, uploadErr := t.uploadImageToTaobao(mainImgLocalPath, taskMsg.BookInfo.BookName)
	if uploadErr != nil {
		return taskMsg, fmt.Errorf("上传图片到淘宝图片空间失败 %v", uploadErr)
	}

	// 库存处理
	if taskMsg.Detail.Stock == 0 && (golabl.Task.Header.TaskType == 1 || golabl.Task.Header.TaskType == 2 || golabl.Task.Header.TaskType == 6) {
		taskMsg.Detail.Stock = golabl.Task.Header.ShopMsg.DefStock
	}

	// 调用淘宝API发布商品
	ret, err := t.tushuAdd(goodsName, int(taskMsg.Detail.Stock), float64(price)/100, mainImg, taskMsg)
	if err != nil {
		return taskMsg, fmt.Errorf("淘宝商品发布失败: %v", err)
	}

	// 解析返回结果，获取商品ID
	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(ret, &result); unmarshalErr != nil {
		return taskMsg, fmt.Errorf("解析淘宝返回结果失败: %v", unmarshalErr)
	}

	// 提取商品ID（根据实际返回结构调整）
	if data, ok := result["data"].(map[string]interface{}); ok {
		if numIid, ok := data["num_iid"].(string); ok {
			taskMsg.Detail.GoodsId, _ = strconv.ParseInt(numIid, 10, 64)
		}
	}

	taskMsg.Detail.OutGoodsId = taskMsg.BookInfo.Isbn
	taskMsg.Detail.Img = mainImg

	return taskMsg, nil
}

// uploadImageToTaobao 上传图片到淘宝图片空间
// 参数:
//   - localPath: 本地图片路径
//   - title: 商品标题（用于生成文件名）
//
// 返回:
//   - 淘宝图片空间URL
//   - 错误信息
func (t *Taobao) uploadImageToTaobao(localPath string, title string) (string, error) {
	taobaoConfig := golabl.Config.TaobaoConfig

	// 生成安全文件名
	imgTitle := sanitizeFileName(title)
	if imgTitle == "" {
		imgTitle = fmt.Sprintf("img_%d", time.Now().Unix())
	}

	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("打开本地图片文件失败: %w", err)
	}
	defer file.Close()

	// 构造 multipart/form-data 请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", imgTitle+".jpg")
	if err != nil {
		return "", fmt.Errorf("创建 multipart 文件部分失败: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("复制文件内容失败: %w", err)
	}
	writer.WriteField("token", taobaoConfig.Token)
	writer.Close()

	uploadURL := taobaoConfig.BaseURL + "/GetJsNamesForUrl.do?loadImg="
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return "", fmt.Errorf("创建上传请求失败: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:145.0) Gecko/20100101 Firefox/145.0")
	req.Header.Set("Origin", "https://fxzsweb.yulinkai.com")
	req.Header.Set("Referer", "https://fxzsweb.yulinkai.com/")

	client := &http.Client{Timeout: 50 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("上传图片请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// 解析 JSON 响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析上传响应 JSON 失败: %w", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("上传响应缺少 data 字段: %s", string(respBody))
	}

	tbImagePath, ok := data["path"].(string)
	if !ok || tbImagePath == "" {
		return "", fmt.Errorf("上传响应缺少 path 字段: %s", string(respBody))
	}

	return tbImagePath, nil
}

// sanitizeFileName 清理文件名，移除非法字符
func sanitizeFileName(name string) string {
	// 移除或替换 Windows 和 Unix 不允许的字符
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	// 限制长度
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

// tushuAdd 调用淘宝API发布图书商品
func (t *Taobao) tushuAdd(title string, num int, price float64, picUrl string, taskMsg planAType.TaskBody) ([]byte, error) {
	timestamp := time.Now().UnixMilli()

	// 解析店铺Token获取配置信息
	taobaoConfig := golabl.Config.TaobaoConfig

	appInfo := map[string]interface{}{
		"app_key":     taobaoConfig.AppKey,
		"app_secret":  taobaoConfig.AppSecret,
		"format":      "json",
		"ipaddress":   "127.0.0.1",
		"method":      "ylk.base.add.book",
		"partner_id":  "sdk-1.0",
		"sign_method": "hmac",
		"timestamp":   timestamp,
		"token":       golabl.Task.Header.ShopMsg.Token,
		"v":           "2.0",
	}

	// bookInfo（商品信息）
	bookInfo := map[string]interface{}{
		"title":            title,
		"outerId":          taskMsg.BookInfo.Isbn,
		"price":            price,
		"cid":              50010485, // 图书类目
		"postageId":        golabl.Task.Header.ShopMsg.CostTemplateId,
		"num":              num,
		"desc":             title,
		"sellerCids":       "",
		"picUrl":           picUrl,
		"approveStatus":    "instock", // 仓库中（不立即上架）
		"bookName":         taskMsg.BookInfo.BookName,
		"deliveryTimeType": 0,
		"newFlag":          1,
	}

	bookInfoJSON, _ := json.Marshal(bookInfo)

	formData := map[string]interface{}{
		"user_id":          taobaoConfig.UserID,
		"company_id":       taobaoConfig.CompanyID,
		"author_shop_name": golabl.Task.Header.ShopName,
		"book_info":        string(bookInfoJSON),
	}

	sign := t.sign(appInfo, formData)

	// 拼接完整 URL
	URL := fmt.Sprintf("%s/router/ylk/base/add/book?"+
		"app_key=%d&format=json&ipaddress=127.0.0.1&partner_id=sdk-1.0&sign_method=hmac&timestamp=%d&v=2.0&"+
		"token=%s&method=ylk.base.add.book&sign=%s&ati=%s",
		taobaoConfig.BaseURL, taobaoConfig.AppKey, timestamp, golabl.Task.Header.ShopMsg.Token, sign, taobaoConfig.Ati)

	// 构造 form data（URL 编码）
	form := url.Values{}
	form.Set("user_id", taobaoConfig.UserID)
	form.Set("company_id", taobaoConfig.CompanyID)
	form.Set("author_shop_name", golabl.Task.Header.ShopName)
	form.Set("book_info", url.QueryEscape(string(bookInfoJSON)))

	resp, err := t.client.PostForm(URL, form)
	if err != nil {
		return nil, fmt.Errorf("发布商品请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, readAllErr := io.ReadAll(resp.Body)
	if readAllErr != nil {
		return nil, fmt.Errorf("读取响应失败: %w", readAllErr)
	}
	fmt.Println("----------------------------------------------body---------------------------------------------")
	fmt.Println(string(body))
	fmt.Println("----------------------------------------------body---------------------------------------------")

	// 发布后等待 5 秒（平台限流）
	time.Sleep(5 * time.Second)

	return nil, nil
}

// sign 计算 API 请求签名（HMAC-MD5，结果为大写十六进制字符串）
func (t *Taobao) sign(appInfo, formData map[string]interface{}) string {
	data := make(map[string]interface{})
	for k, v := range appInfo {
		data[k] = v
	}
	for k, v := range formData {
		data[k] = v
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(fmt.Sprintf("%v", data[k]))
	}
	concatStr := sb.String()

	taobaoConfig := golabl.Config.TaobaoConfig
	mac := hmac.New(md5.New, []byte(taobaoConfig.AppSecret))
	mac.Write([]byte(concatStr))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

// phaseOneGoodsOnlyCount 第一阶段只获取总数
func (t *Taobao) phaseOneGoodsOnlyCount(pageSize int, maxPage int) error {
	for page := 1; page <= maxPage; page++ {
		items, err := t.tushuGetList(page, pageSize)
		if err != nil {
			return fmt.Errorf("获取商品列表失败，页码: %d, 错误: %v", page, err)
		}
		if items == nil || len(items) == 0 {
			break
		}

		// 第一页更新总数
		if page == 1 {
			// 淘宝API不直接返回总数，这里用估算或从header中获取
			fmt.Printf("第一阶段完成，当前页获取 %d 条商品\n", len(items))
		}
	}
	return nil
}

// phaseTwoGoods 第二阶段拉取商品
func (t *Taobao) phaseTwoGoods(startPage int, maxPage int, pageSize int, totalFetched *int) error {
	for page := startPage; page <= maxPage; page++ {
		items, err := t.tushuGetList(page, pageSize)
		if err != nil {
			return fmt.Errorf("获取商品列表失败，页码: %d, 错误: %v", page, err)
		}
		if items == nil || len(items) == 0 {
			fmt.Printf("第 %d 页无数据，结束拉取\n", page)
			break
		}

		// 处理商品数据
		for _, item := range items {
			*totalFetched++

			itemMap, ok := item["item"].(map[string]interface{})
			if !ok {
				continue
			}

			numIid, _ := itemMap["numIid"].(string)
			title, _ := itemMap["title"].(string)
			outerId, _ := itemMap["outerId"].(string)
			priceStr, _ := itemMap["price"].(string)
			num, _ := itemMap["num"].(float64)

			price, _ := strconv.ParseFloat(priceStr, 64)

			// 转换为TaskBody
			bodyWait := planAType.TaskBody{
				BookInfo: planAType.BookInfo{
					Isbn:     outerId,
					BookName: title,
					Price:    int64(price * 100),
				},
				Detail: planAType.TaskDetail{
					Status:  1,
					GoodsId: 0,
					Stock:   int64(num),
					Error:   numIid, // 临时存储numIid
				},
			}

			// 转为JSON
			bodyWaitJson, jsonMarshalErr := json.Marshal(bodyWait)
			if jsonMarshalErr != nil {
				return fmt.Errorf("将bodyWait转为json失败: %v\n", jsonMarshalErr)
			}

			// 写入 body_wait
			addTaskToBodyWaitErr := service.AddTaskToBodyWait(string(bodyWaitJson))
			if addTaskToBodyWaitErr != nil {
				return addTaskToBodyWaitErr
			}
		}

		// 更新进度
		if getTaskFooterErr := service.GetTaskFooter(); getTaskFooterErr != nil {
			return getTaskFooterErr
		}
		con := int64(len(items))
		if con >= golabl.Task.Footer.TaskCountTrue {
			con = golabl.Task.Footer.TaskCountTrue - con
		}
		if updateTaskProgressErr := tool.UpdateTaskProgress(con); updateTaskProgressErr != nil {
			return updateTaskProgressErr
		}

		fmt.Printf("第二阶段 - 页码: %d, 本页获取: %d 条，累计: %d\n", page, len(items), *totalFetched)
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

// tushuGetList 获取淘宝商品列表
func (t *Taobao) tushuGetList(page int, pageSize int) ([]map[string]interface{}, error) {
	//timestamp := time.Now().UnixMilli()
	//taobaoConfig := golabl.Config.TaobaoConfig
	//
	//appInfo := map[string]interface{}{
	//	"app_key":     taobaoConfig.AppKey,
	//	"app_secret":  taobaoConfig.AppSecret,
	//	"format":      "json",
	//	"ipaddress":   "127.0.0.1",
	//	"method":      "ylk.item.tb.api.list",
	//	"partner_id":  "sdk-1.0",
	//	"sign_method": "hmac",
	//	"timestamp":   timestamp,
	//	"token":       golabl.Task.Header.ShopMsg.Token,
	//	"v":           "2.0",
	//}
	//
	//formData := map[string]interface{}{
	//	"user_id":        taobaoConfig.UserId,
	//	"company_id":     taobaoConfig.CompanyId,
	//	"approve_status": "onsale",
	//	"shop_name":      golabl.Task.Header.ShopName,
	//	"diagnose_flag":  1,
	//	"page_no":        page,
	//	"page_size":      pageSize,
	//	"order_by":       "modified:asc",
	//}
	//
	//sign := t.sign(appInfo, formData)
	//
	//URL := fmt.Sprintf("%s/router/ylk/item/tb/api/list?"+
	//	"app_key=%d&format=json&ipaddress=127.0.0.1&partner_id=sdk-1.0&sign_method=hmac&timestamp=%d&v=2.0&"+
	//	"token=%s&method=ylk.item.tb.api.list&sign=%s&ati=%s",
	//	taobaoConfig.BaseUrl, taobaoConfig.AppKey, timestamp, golabl.Task.Header.ShopMsg.Token, sign, taobaoConfig.Ati)
	//
	//form := url.Values{}
	//for k, v := range formData {
	//	form.Set(k, fmt.Sprintf("%v", v))
	//}
	//
	//resp, err := t.client.PostForm(URL, form)
	//if err != nil {
	//	return nil, fmt.Errorf("获取商品列表请求失败: %w", err)
	//}
	//defer resp.Body.Close()
	//
	//body, _ := io.ReadAll(resp.Body)
	//
	//var result map[string]interface{}
	//if err := json.Unmarshal(body, &result); err != nil {
	//	return nil, fmt.Errorf("解析响应 JSON 失败: %w", err)
	//}
	//
	//items, ok := result["items"].([]interface{})
	//if !ok || items == nil {
	//	return nil, nil
	//}
	//
	//resultItems := make([]map[string]interface{}, 0, len(items))
	//for _, itm := range items {
	//	if m, ok := itm.(map[string]interface{}); ok {
	//		resultItems = append(resultItems, m)
	//	}
	//}

	return nil, nil
}

// deduplicateToBodyOver 去重并写入body_over
func (t *Taobao) deduplicateToBodyOver(duplicateCount *int, uniqueCount *int) error {
	page := 1
	pageSize := 1000

	processedGoodsIds := make(map[string]bool)

	deleteTaskBodyOverErr := service.DeleteTaskBodyOver()
	if deleteTaskBodyOverErr != nil {
		return deleteTaskBodyOverErr
	}
	deleteTaskBodyBackupErr := service.DeleteTaskBodyBackup()
	if deleteTaskBodyBackupErr != nil {
		return deleteTaskBodyBackupErr
	}

	bodyWaitCount, getTaskBodyWaitCountErr := service.GetTaskBodyWaitCount()
	if getTaskBodyWaitCountErr != nil {
		return getTaskBodyWaitCountErr
	}
	pageTotal := (bodyWaitCount + int64(pageSize) - 1) / int64(pageSize)

	for {
		list, getTaskBodyOverListErr := service.GetTaskBodyWaitList(page, pageSize)
		if getTaskBodyOverListErr != nil {
			return getTaskBodyOverListErr
		}
		if len(list) <= 0 {
			break
		}

		for _, v := range list {
			goods := planAType.TaskBody{}
			jsonUnmarshalErr := json.Unmarshal([]byte(v), &goods)
			if jsonUnmarshalErr != nil {
				return fmt.Errorf("将json转为结构体失败: %v\n", jsonUnmarshalErr)
			}

			// 使用outerId(ISBN)作为去重key
			goodsKey := goods.BookInfo.Isbn
			if goodsKey == "" {
				goodsKey = goods.Detail.Error // 使用numIid作为备选
			}

			if !processedGoodsIds[goodsKey] {
				processedGoodsIds[goodsKey] = true
				*uniqueCount++

				goods.Detail.Status = 1
				addTaskToBodyOverErr := service.AddTaskToBodyOver(goods, []string{"body_over", "body_backup"})
				if addTaskToBodyOverErr != nil {
					return addTaskToBodyOverErr
				}
			} else {
				*duplicateCount++
			}
		}

		fmt.Printf("去重处理中 - 当前页 %v 总页数 %v\n", page, pageTotal)
		page++

		if getTaskFooterErr := service.GetTaskFooter(); getTaskFooterErr != nil {
			return getTaskFooterErr
		}
		con := int64(len(list))
		if con >= golabl.Task.Footer.TaskCountTrue {
			con = golabl.Task.Footer.TaskCountTrue - con
		}
		if updateTaskProgressErr := tool.UpdateTaskProgress(con); updateTaskProgressErr != nil {
			return updateTaskProgressErr
		}

		time.Sleep(1 * time.Second)
	}

	deleteTaskBodyWaitErr := service.DeleteTaskBodyWait()
	if deleteTaskBodyWaitErr != nil {
		return deleteTaskBodyWaitErr
	}
	return nil
}

// upOne 上架商品
func (t *Taobao) upOne(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	//if taskMsg.Detail.GoodsId == 0 {
	//	return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("淘宝商品 Id不能为空"))
	//}
	//
	//timestamp := time.Now().UnixMilli()
	//taobaoConfig := golabl.Config.TaobaoConfig
	//
	//appInfo := map[string]interface{}{
	//	"app_key":     taobaoConfig.AppKey,
	//	"app_secret":  taobaoConfig.AppSecret,
	//	"format":      "json",
	//	"ipaddress":   "127.0.0.1",
	//	"method":      "ylk.item.listing",
	//	"partner_id":  "sdk-1.0",
	//	"sign_method": "hmac",
	//	"timestamp":   timestamp,
	//	"token":       golabl.Task.Header.ShopMsg.Token,
	//	"v":           "2.0",
	//}
	//
	//uid := strconv.FormatInt(time.Now().UnixMilli(), 10)
	//formData := map[string]interface{}{
	//	"user_id":    uid,
	//	"company_id": uid,
	//	"num_iid":    strconv.FormatInt(taskMsg.Detail.GoodsId, 10),
	//	"nick":       golabl.Task.Header.ShopName,
	//	"num":        "2",
	//}
	//
	//sign := t.sign(appInfo, formData)
	//
	//URL := fmt.Sprintf("%s/router/ylk/item/listing?"+
	//	"app_key=%d&format=json&ipaddress=127.0.0.1&partner_id=sdk-1.0&sign_method=hmac&timestamp=%d&v=2.0&"+
	//	"token=%s&method=ylk.item.listing&sign=%s&ati=%s",
	//	taobaoConfig.BaseUrl, taobaoConfig.AppKey, timestamp, golabl.Task.Header.ShopMsg.Token, sign, taobaoConfig.Ati)
	//
	//form := url.Values{}
	//for k, v := range formData {
	//	form.Set(k, fmt.Sprintf("%v", v))
	//}
	//
	//resp, err := t.client.PostForm(URL, form)
	//if err != nil {
	//	return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("上架请求失败: %v", err))
	//}
	//defer resp.Body.Close()

	return tool.ReturnSuccess(taskMsg)
}

// downOne 下架商品
func (t *Taobao) downOne(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	//if taskMsg.Detail.GoodsId == 0 {
	//	return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("淘宝商品 Id不能为空"))
	//}
	//
	//timestamp := time.Now().UnixMilli()
	//taobaoConfig := golabl.Config.TaobaoConfig
	//
	//appInfo := map[string]interface{}{
	//	"app_key":     taobaoConfig.AppKey,
	//	"app_secret":  taobaoConfig.AppSecret,
	//	"format":      "json",
	//	"ipaddress":   "127.0.0.1",
	//	"method":      "ylk.item.delisting",
	//	"partner_id":  "sdk-1.0",
	//	"sign_method": "hmac",
	//	"timestamp":   timestamp,
	//	"token":       golabl.Task.Header.ShopMsg.Token,
	//	"v":           "2.0",
	//}
	//
	//uid := strconv.FormatInt(time.Now().UnixMilli(), 10)
	//formData := map[string]interface{}{
	//	"user_id":    uid,
	//	"company_id": uid,
	//	"num_iid":    strconv.FormatInt(taskMsg.Detail.GoodsId, 10),
	//	"nick":       golabl.Task.Header.ShopName,
	//}
	//
	//sign := t.sign(appInfo, formData)
	//
	//URL := fmt.Sprintf("%s/router/ylk/item/delisting?"+
	//	"app_key=%d&format=json&ipaddress=127.0.0.1&partner_id=sdk-1.0&sign_method=hmac&timestamp=%d&v=2.0&"+
	//	"token=%s&method=ylk.item.delisting&sign=%s&ati=%s",
	//	taobaoConfig.BaseUrl, taobaoConfig.AppKey, timestamp, golabl.Task.Header.ShopMsg.Token, sign, taobaoConfig.Ati)
	//
	//form := url.Values{}
	//for k, v := range formData {
	//	form.Set(k, fmt.Sprintf("%v", v))
	//}
	//
	//resp, err := t.client.PostForm(URL, form)
	//if err != nil {
	//	return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("下架请求失败: %v", err))
	//}
	//defer resp.Body.Close()

	return tool.ReturnSuccess(taskMsg)
}

// updateStock 修改库存
func (t *Taobao) updateStock(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	//if taskMsg.Detail.GoodsId == 0 {
	//	return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("淘宝商品 Id不能为空"))
	//}
	//
	//timestamp := time.Now().UnixMilli()
	//taobaoConfig := golabl.Config.TaobaoConfig
	//
	//appInfo := map[string]interface{}{
	//	"app_key":     taobaoConfig.AppKey,
	//	"app_secret":  taobaoConfig.AppSecret,
	//	"format":      "json",
	//	"ipaddress":   "127.0.0.1",
	//	"method":      "ylk.item.operate.update",
	//	"partner_id":  "sdk-1.0",
	//	"sign_method": "hmac",
	//	"timestamp":   timestamp,
	//	"token":       golabl.Task.Header.ShopMsg.Token,
	//	"v":           "2.0",
	//}
	//
	//itemUpdateObj := map[string]interface{}{
	//	"num":     taskMsg.Detail.Stock,
	//	"nick":    golabl.Task.Header.ShopName,
	//	"outerId": taskMsg.BookInfo.Isbn,
	//	"listSku": []interface{}{},
	//}
	//
	//itemUpdateJSON, _ := json.Marshal(itemUpdateObj)
	//itemUpdateEncoded := url.QueryEscape(string(itemUpdateJSON))
	//
	//formData := map[string]interface{}{
	//	"company_id": taobaoConfig.CompanyId,
	//	"itemUpdate": string(itemUpdateJSON),
	//	"num_iid":    strconv.FormatInt(taskMsg.Detail.GoodsId, 10),
	//	"user_id":    taobaoConfig.UserId,
	//}
	//
	//sign := t.sign(appInfo, formData)
	//
	//URL := fmt.Sprintf("%s/router/ylk/item/operate/update?"+
	//	"app_key=%d&format=json&ipaddress=127.0.0.1&partner_id=sdk-1.0&sign_method=hmac&timestamp=%d&v=2.0&"+
	//	"token=%s&method=ylk.item.operate.update&sign=%s&ati=%s",
	//	taobaoConfig.BaseUrl, taobaoConfig.AppKey, timestamp, golabl.Task.Header.ShopMsg.Token, sign, taobaoConfig.Ati)
	//
	//form := url.Values{}
	//form.Set("company_id", taobaoConfig.CompanyId)
	//form.Set("itemUpdate", itemUpdateEncoded)
	//form.Set("num_iid", strconv.FormatInt(taskMsg.Detail.GoodsId, 10))
	//form.Set("user_id", taobaoConfig.UserId)
	//
	//resp, err := t.client.PostForm(URL, form)
	//if err != nil {
	//	return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("修改库存请求失败: %v", err))
	//}
	//defer resp.Body.Close()

	return tool.ReturnSuccess(taskMsg)
}

// updatePrice 修改价格
func (t *Taobao) updatePrice(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	//if taskMsg.Detail.GoodsId == 0 {
	//	return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("淘宝商品 Id不能为空"))
	//}
	//if taskMsg.Detail.Price == 0 {
	//	return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("淘宝商品价格不能为0"))
	//}
	//
	//timestamp := time.Now().UnixMilli()
	//taobaoConfig := golabl.Config.TaobaoConfig
	//
	//appInfo := map[string]interface{}{
	//	"app_key":     taobaoConfig.AppKey,
	//	"app_secret":  taobaoConfig.AppSecret,
	//	"format":      "json",
	//	"ipaddress":   "127.0.0.1",
	//	"method":      "ylk.item.operate.update",
	//	"partner_id":  "sdk-1.0",
	//	"sign_method": "hmac",
	//	"timestamp":   timestamp,
	//	"token":       golabl.Task.Header.ShopMsg.Token,
	//	"v":           "2.0",
	//}
	//
	//fen := tool.BuildGoodsPrice(taskMsg.Detail.Price)
	//yuan := tool.FenToYuan(fen)
	//
	//itemUpdateObj := map[string]interface{}{
	//	"price":   yuan,
	//	"nick":    golabl.Task.Header.ShopName,
	//	"outerId": taskMsg.BookInfo.Isbn,
	//	"listSku": []interface{}{},
	//}
	//
	//itemUpdateJSON, _ := json.Marshal(itemUpdateObj)
	//itemUpdateEncoded := url.QueryEscape(string(itemUpdateJSON))
	//
	//formData := map[string]interface{}{
	//	"company_id": taobaoConfig.CompanyId,
	//	"itemUpdate": string(itemUpdateJSON),
	//	"num_iid":    strconv.FormatInt(taskMsg.Detail.GoodsId, 10),
	//	"user_id":    taobaoConfig.UserId,
	//}
	//
	//sign := t.sign(appInfo, formData)
	//
	//URL := fmt.Sprintf("%s/router/ylk/item/operate/update?"+
	//	"app_key=%d&format=json&ipaddress=127.0.0.1&partner_id=sdk-1.0&sign_method=hmac&timestamp=%d&v=2.0&"+
	//	"token=%s&method=ylk.item.operate.update&sign=%s&ati=%s",
	//	taobaoConfig.BaseUrl, taobaoConfig.AppKey, timestamp, golabl.Task.Header.ShopMsg.Token, sign, taobaoConfig.Ati)
	//
	//form := url.Values{}
	//form.Set("company_id", taobaoConfig.CompanyId)
	//form.Set("itemUpdate", itemUpdateEncoded)
	//form.Set("num_iid", strconv.FormatInt(taskMsg.Detail.GoodsId, 10))
	//form.Set("user_id", taobaoConfig.UserId)
	//
	//resp, err := t.client.PostForm(URL, form)
	//if err != nil {
	//	return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("修改价格请求失败: %v", err))
	//}
	//defer resp.Body.Close()

	return tool.ReturnSuccess(taskMsg)
}
