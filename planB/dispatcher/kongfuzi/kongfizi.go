package kongfuzi

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/planB/initialization/golabl"
	"planA/planB/modules/logs"
	"planA/planB/service"
	"planA/planB/tool"
	planBTypeKfz "planA/planB/type/kfz"
	planAType "planA/type"
	"strconv"
	"strings"
	"time"
)

type KongFuZi struct {
}

// NewKongFuZi 创建孔夫子平台
func NewKongFuZi() *KongFuZi {
	return &KongFuZi{}
}

// AddGoodsTask 添加商品
// @param taskMsg 任务内容
// @return string body 信息
// @return error 错误
func (kongFuZi *KongFuZi) AddGoodsTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}

	// 价格不能小于0
	if taskMsg.Detail.Price <= 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("价格不能小于等于0"))
	}

	//获取出版社信息并解析1
	if getPublishingErr := service.GetPublishingVid(&taskMsg); getPublishingErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("获取出版社信息失败-原因来自:%v", getPublishingErr))
	}

	//违规词处理1
	if golabl.Config.Server.Filter == 1 {
		//开启违规词处理
		if taskMsgErr := tool.FilterWord(&taskMsg); taskMsgErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, taskMsgErr)
		}
	}

	// 价格处理
	price := tool.BuildPrice(golabl.Task.Header.PriceMod, taskMsg.Detail.Price)
	if price == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("不在价格区间内 isbn %v  原始价格 %v  当前价格 %v 价格模版 %v", taskMsg.BookInfo.Isbn, taskMsg.Detail.Price, price, golabl.Task.Header.PriceMod))
	}
	taskMsg.Detail.Price = price

	var goodsAdd planBTypeKfz.GoodsAdd17

	// *********************构建参数 开始******************************** //

	//模板编号 默认 17
	goodsAdd.Tpl = "17"

	//分类编号
	if value, exists := golabl.KfzGetCommonCategory[string(taskMsg.BookInfo.CatIdObject.KongFuZiCatId)]; exists {
		goodsAdd.CatId = strconv.FormatInt(value, 10)
	} else {
		goodsAdd.CatId = "43000000000000000"
	}

	//构建商品名称
	goodsAdd.ItemName = tool.BuildGoodsName(
		golabl.Task.Header.ShopMsg.GoodsNamePrefix, // 商品名称前缀
		golabl.Task.Header.ShopMsg.GoodsNameSuffix, // 商品名称后缀
		golabl.Task.Header.ShopMsg.TitleConsistOf,  // 标题组成
		golabl.Task.Header.ShopMsg.SpaceCharacter,  // 间隔符
		taskMsg.BookInfo) // 图书信息
	taskMsg.Detail.GoodsName = goodsAdd.ItemName

	// 售价 + 运费
	priceTol := taskMsg.Detail.Price
	if golabl.Task.Header.PriceType != "0" {
		priceTol = priceTol + taskMsg.Detail.ShippingCost
	}

	//售价
	goodsAdd.Price = tool.FenToYuan(priceTol)

	//库存
	if taskMsg.Detail.Stock == 0 && (golabl.Task.Header.TaskType == 1 || golabl.Task.Header.TaskType == 2 || golabl.Task.Header.TaskType == 6) {
		//如果库存为0 则给默认库存
		taskMsg.Detail.Stock = golabl.Task.Header.ShopMsg.DefStock
	}
	goodsAdd.Number = strconv.FormatInt(taskMsg.Detail.Stock, 10)

	//品相
	goodsAdd.Quality = strconv.FormatInt(taskMsg.Detail.Condition, 10)
	if goodsAdd.Quality == "" || goodsAdd.Quality == "0" {
		goodsAdd.Quality = strconv.FormatInt(golabl.Task.Header.ShopMsg.ConditionDef, 10)
	}

	if len(taskMsg.BookInfo.ImageObject.CarouselUrlArray) == 0 {
		// 无图片信息 isbn计次
		setNoImgCountErr := service.SetNoImgCount(taskMsg.BookInfo.Isbn)
		if setNoImgCountErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("无图片信息isbn计次错误 isbn %v %v", taskMsg.BookInfo.Isbn, setNoImgCountErr.Error()))
		}
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("缺少官图"))
	}

	// 将 轮播图不是 孔夫子图片空间的图片替换成孔夫子图片空间图片
	for k, v := range taskMsg.BookInfo.ImageObject.CarouselUrlArray {
		if !strings.Contains(v, "kfzimg.com") {
			kfz, uploadImageToKfzErr := UploadImageToKfz(v)
			if uploadImageToKfzErr != nil {
				return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, uploadImageToKfzErr)
			}
			// 添加到轮播图
			taskMsg.BookInfo.ImageObject.CarouselUrlArray[k] = kfz
		}
	}
	//将 最后的图片不是 孔夫子图片空间的图片替换成孔夫子图片空间图片
	for k, v := range golabl.Task.Header.ShopMsg.CarouseLastImgUrlArray {
		if !strings.Contains(v, "kfzimg.com") {
			kfz, uploadImageToKfzErr := UploadImageToKfz(v)
			if uploadImageToKfzErr != nil {
				return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, uploadImageToKfzErr)
			}
			// 添加到轮播图
			golabl.Task.Header.ShopMsg.CarouseLastImgUrlArray[k] = kfz
		}
	}
	//处理图片
	oldCarouselUrlArray := append([]string{}, taskMsg.BookInfo.ImageObject.CarouselUrlArray...) //原始轮播图，用于后续处理，不会被打上水印
	// 存在水印图片，则打水印
	if golabl.Task.Header.ShopMsg.WatermarkImgUrl != "" {
		//获取水印图片
		watermarkImgUrl, watermarkImgErr := tool.GetWatermarkImg()
		if watermarkImgErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("获取水印图片失败 %v", watermarkImgErr))
		}
		//打水印
		watermarkFromURLExsBase64Arr, watermarkFromURLExsErr := tool.AddWatermarkFromURLExs(taskMsg.BookInfo.ImageObject.CarouselUrlArray, watermarkImgUrl, golabl.Task.Header.ShopMsg.WatermarkPosition)
		if watermarkFromURLExsErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("图片打水印失败 %v", watermarkFromURLExsErr))
		}
		//图片上传到孔夫子
		toPdd, uploadImageToPddErr := tool.UploadImageToKfz(watermarkFromURLExsBase64Arr)
		if uploadImageToPddErr != nil {
			return "", fmt.Errorf("图片上传到拼多多失败 %v", uploadImageToPddErr)
		}
		//将上传的图片替换到商品轮播图中
		for i := 0; i < len(toPdd); i++ {
			taskMsg.BookInfo.ImageObject.CarouselUrlArray[i] = toPdd[i]
		}
	}

	imgUrlArr := tool.BuildCarouselGallery(golabl.Task.Header.ShopMsg.CarouseLastImgUrlArray, oldCarouselUrlArray, taskMsg.BookInfo.ImageObject.CarouselUrlArray, golabl.Task.Header.ShopMsg.WatermarkPosition)

	//商品主图
	goodsAdd.ImgUrl = imgUrlArr[0]

	//多个商品图片
	goodsAdd.Images = strings.Join(imgUrlArr, ";")

	//运费设置
	bearShipping := "seller"
	if golabl.Task.Header.ShopMsg.IsParcel == "0" {
		bearShipping = "buyer"
		// 设置运费模板编号
		goodsAdd.MouldId = "1"
	}
	goodsAdd.BearShipping = bearShipping

	//运费模板编号
	//将 golabl.Task.Header.ShopMsg.CostTemplateId 转为int
	costTemplateId, atoiErr := strconv.Atoi(golabl.Task.Header.ShopMsg.CostTemplateId)
	if atoiErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("运费模板编号转换错误 %v", atoiErr))
	}
	goodsAdd.MouldId = strconv.Itoa(costTemplateId)

	//重量
	goodsAdd.Weight = fmt.Sprintf("%v", float64(golabl.Task.Header.ShopMsg.BookWeight/100))

	//商品标准本数
	goodsAdd.WeightPiece = fmt.Sprintf("%v", float64(golabl.Task.Header.ShopMsg.StandardNumber/100))

	// *********************构建模板参数 开始******************************** //

	// ISBN
	goodsAdd.Isbn = taskMsg.BookInfo.Isbn

	// 作者
	goodsAdd.Author = taskMsg.BookInfo.Author

	// 出版社
	goodsAdd.Press = taskMsg.Publishing.Value

	// 出版社日期
	goodsAdd.PubDate = taskMsg.BookInfo.PublicationDate

	// 装帧
	goodsAdd.Binding = taskMsg.BookInfo.Binding

	// 开本
	goodsAdd.PageSize = strconv.FormatInt(taskMsg.BookInfo.Format, 10)

	// 页数
	goodsAdd.PageNum = strconv.FormatInt(taskMsg.BookInfo.PagesCount, 10)

	// 字数
	goodsAdd.WordNum = fmt.Sprintf("%v", float64(taskMsg.BookInfo.WordsCount/1000))

	// 图书定价
	goodsAdd.OriPrice = fmt.Sprintf("%v", float64(tool.BuildGoodsPrice(priceTol)/100))

	url := "http://127.0.0.1:8095"
	tool.HttpGetRequest(url)

	// 新增商品
	goods, _, addGoodsErr := addGoods(logUuid, goodsAdd)
	if addGoodsErr != nil {
		return "", addGoodsErr
	}

	// 构建商品编码
	outGoodsId := ""
	if taskMsg.Detail.OutGoodsId != "" {
		outGoodsId = taskMsg.Detail.OutGoodsId
	} else {
		outGoodsId = taskMsg.BookInfo.Isbn
	}

	//判断错误
	if goods.ErrorResponse != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("新增商品失败 %v", goods.ErrorResponse.SubMsg))
	}
	taskMsg.Detail.GoodsId = goods.SuccessResponse.Item.ItemId
	taskMsg.Detail.OutGoodsId = outGoodsId
	taskMsg.Detail.Img = goodsAdd.ImgUrl
	return tool.ReturnSuccess(taskMsg)
}

// GetGoodsTask 获取商品
// @return string body 信息
// @return error 错误
func (kongFuZi *KongFuZi) GetGoodsTask() (string, error) {
	// 生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}

	const pageSize = 100
	const maxRecordsPerRange = 10000 // 每个时间范围最多获取10000条
	const timeRangeDays = 30         // 时间范围天数

	var lastAddTime int64 = 0 // 上次拉取的最后一条商品的添加时间（秒级时间戳）

	// 统计变量
	totalFetched := 0   // 总共获取到的商品数（包括重复）
	duplicateCount := 0 // 重复商品数量
	uniqueCount := 0    // 不重复商品数量

	//查询body_wait是否存在，如果存在则证明不是第一次执行
	//要获取body_wait中最后一条数据的添加时间作为查询的开始时间
	exist, isTaskBodyWaitExistErr := service.IsTaskBodyWaitExist()
	if isTaskBodyWaitExistErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, isTaskBodyWaitExistErr)
	}

	if exist {
		// 获取最后一条数据的添加时间
		lastBodyWaitDataJson, getLastGoodsAddTimeErr := service.GetTaskBodyWaitLast()
		if getLastGoodsAddTimeErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, getLastGoodsAddTimeErr)
		}
		// 解析 lastBodyWaitData 到结构体
		var lastBodyWaitData planAType.TaskBody
		unmarshalErr := json.Unmarshal([]byte(lastBodyWaitDataJson), &lastBodyWaitData)
		if unmarshalErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, unmarshalErr)
		}
		// 将数据的添加时间给到 lastAddTime
		lastAddTime = lastBodyWaitData.BookInfo.Price
		// 第二阶段：使用时间范围分批拉取（goroutine + 120秒超时，防止 DLL 永久阻塞）
		phaseTwoResultCh := make(chan error, 1)
		go func() {
			fmt.Println("[DEBUG] PhaseTwoGoods goroutine started (exist=true branch)")
			phaseTwoResultCh <- PhaseTwoGoods(pageSize, &totalFetched, &lastAddTime, maxRecordsPerRange, timeRangeDays)
		}()
		// 等待 PhaseTwoGoods 完成或超时
		select {
		case phaseTwoErr := <-phaseTwoResultCh:
			fmt.Printf("[DEBUG] PhaseTwoGoods completed (exist=true): %v\n", phaseTwoErr)
			if phaseTwoErr != nil {
				return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, phaseTwoErr)
			}
		case <-time.After(120 * time.Second):
			fmt.Println("[ERROR] PhaseTwoGoods 超时（120秒），可能是 DLL 函数 KongfzShopItemList 不可用")
			return "", fmt.Errorf("PhaseTwoGoods 超时（120秒），可能是 DLL 函数 KongfzShopItemList 不可用")
		}
	} else {
		// 第一阶段：获取第一页商品数量作为总数
		firstTimeGoodsErr := phaseOneGoods(pageSize, &totalFetched)
		if firstTimeGoodsErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, firstTimeGoodsErr)
		}

		// 计算起始时间（当前时间往前180天）
		lastAddTime = time.Now().Unix() - 180*24*60*60
		// 第二阶段：使用时间范围分批拉取（goroutine + 120秒超时，防止 DLL 永久阻塞）
		phaseTwoResultCh := make(chan error, 1)
		go func() {
			fmt.Println("[DEBUG] PhaseTwoGoods goroutine started")
			phaseTwoResultCh <- PhaseTwoGoods(pageSize, &totalFetched, &lastAddTime, maxRecordsPerRange, timeRangeDays)
		}()
		// 等待 PhaseTwoGoods 完成或超时
		select {
		case phaseTwoErr := <-phaseTwoResultCh:
			fmt.Printf("[DEBUG] PhaseTwoGoods completed: %v\n", phaseTwoErr)
			if phaseTwoErr != nil {
				return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, phaseTwoErr)
			}
		case <-time.After(120 * time.Second):
			fmt.Println("[ERROR] PhaseTwoGoods 超时（120秒），可能是 DLL 函数 KongfzShopItemList 不可用")
			return "", fmt.Errorf("PhaseTwoGoods 超时（120秒），可能是 DLL 函数 KongfzShopItemList 不可用")
		}
	}
	//更新状态为推送中
	updateTaskStatusErr := service.UpdateTaskStatus(planAType.TaskStatusPushTaskStatus)
	if updateTaskStatusErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, updateTaskStatusErr)
	}

	// 重新设置任务进度
	if updateTaskHeaderErr := service.SetTaskCount(strconv.FormatInt(golabl.Task.Footer.TaskCountTrue, 10)); updateTaskHeaderErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, updateTaskHeaderErr)
	}

	// 去重复与保存
	deduplicateToBodyOverErr := deduplicateToBodyOver(&duplicateCount, &uniqueCount)
	if deduplicateToBodyOverErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, deduplicateToBodyOverErr)
	}

	// 输出统计信息
	statsLogMsg := fmt.Sprintf(`
════════════════════════════════════════════════════════════════
【孔夫子店铺拉取】
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
	tool.LoggingMiddleware(logs.LOG_LEVEL_INFO, statsLogMsg)

	return tool.ReturnSuccess(planAType.TaskBody{})
}

// OperationGoodsTask 操作商品
// @param taskMsg 任务内容
// @return string body 信息
// @return string error 错误
func (kongFuZi *KongFuZi) OperationGoodsTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}
	switch taskMsg.Detail.Status {
	case 1:
		return executeGoodsLaunch(logUuid, taskMsg) //上架 {"book_info":{"isbn":"9787115600387"},"detail":{"goods_id":1562238986012229,"status":1}}
	case 2:
		return executeGoodsDownShelf(logUuid, taskMsg) // 下架 {"book_info":{"isbn":"9787115600387"},"detail":{"goods_id":1562238986012229,"status":2}}
	case 4:
		return executeGoodsUpdateStock(logUuid, taskMsg) //修改商品库存 {"book_info":{"isbn":"9787115600387"},"detail":{"goods_id":1562238986012229,"status":4,"stock":2}}
	case 5:
		return executeGoodsUpdatePrice(logUuid, taskMsg) //修改商品价格 {"book_info":{"isbn":"9787115600387"},"detail":{"goods_id":1562238986012229,"status":5,"price":5000}}
	default:
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("未知操作类型"))
	}
}

// IncStockTask 增量库存
// @param taskMsg 任务内容
// @return string body 信息
// @return error 错误
func (kongFuZi *KongFuZi) IncStockTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}
	// 获取商品id
	getGoodsByShopIdAndIsbn, GetGoodsByShopIdAndIsbnErr := tool.GetGoodsByShopIdAndIsbn(golabl.Task.Header.ShopId, taskMsg.BookInfo.Isbn)
	if GetGoodsByShopIdAndIsbnErr != nil {
		return "", GetGoodsByShopIdAndIsbnErr
	}
	if getGoodsByShopIdAndIsbn.Code != "200" {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("请求ERP获取商品编码与skuid失败: %v", getGoodsByShopIdAndIsbn))
	}
	if len(getGoodsByShopIdAndIsbn.Data) == 0 {
		//新发布
		task, addGoodsTaskErr := kongFuZi.AddGoodsTask(taskMsg)
		if addGoodsTaskErr != nil {
			return "", addGoodsTaskErr
		}
		return task, nil
	} else {
		// 将 getGoodsByShopIdAndIsbn.Data[0].TrilateralId 转为 int64
		trilateralId, trilateralIdParseIntErr := strconv.ParseInt(getGoodsByShopIdAndIsbn.Data[0].TrilateralId, 10, 64)
		if trilateralIdParseIntErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, trilateralIdParseIntErr)
		}

		// 获取商品详情
		getGoodsListReq := planBTypeKfz.GetGoodsListReq{
			ItemId: getGoodsByShopIdAndIsbn.Data[0].TrilateralId,
		}

		getGoodsListReqJson, marshalErr := json.Marshal(getGoodsListReq)
		if marshalErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, marshalErr)
		}

		// 发送请求
		goodsListResp, goodsListRespErr := sendGoodsListRequest(string(getGoodsListReqJson))
		if goodsListRespErr != nil {
			return "", fmt.Errorf("获取商品列表失败 %v", goodsListRespErr)
		}

		//检验商品数量
		if len(goodsListResp.SuccessResponse.List) == 0 {
			return "", fmt.Errorf("在孔夫子中未查询到该商品id  %v", trilateralId)
		}

		//增量修改库存
		taskMsg.Detail.GoodsId = trilateralId
		taskMsg.Detail.Stock = taskMsg.Detail.Stock + int64(goodsListResp.SuccessResponse.List[0].Number)
		quantity, updateGoodsQuantityErr := executeGoodsUpdateStock(logUuid, taskMsg)
		if updateGoodsQuantityErr != nil {
			return "", updateGoodsQuantityErr
		}
		return quantity, nil
	}
}

func (kongFuZi *KongFuZi) SetGoodsTask() string {
	//TODO implement me
	return ""
}

// *******************************私有方法************************************ //

// UploadImageToKfz 将图片上传到孔夫子图片空间
// @param imgUrl 图片地址
// @return string 图片地址
// @return error 错误
func UploadImageToKfz(imgUrl string) (string, error) {

	//将图片保存到本地
	imgTempUrl, saveBase64ImageByDateErr := tool.SaveBase64ImageByDate(imgUrl)
	if saveBase64ImageByDateErr != nil {
		return "", fmt.Errorf("保存图片失败 %v", saveBase64ImageByDateErr)
	}
	//将图片上传到孔夫子
	upload, kfzGoodsImageUploadErr := golabl.KfzDll.KfzGoodsImageUpload(golabl.Config.KfzConfig.AppId, golabl.Config.KfzConfig.AppSecret, golabl.Task.Header.ShopMsg.Token, imgTempUrl)
	if kfzGoodsImageUploadErr != nil {
		return "", fmt.Errorf("上传图片失败 %v", kfzGoodsImageUploadErr)
	}
	// 解析数据
	var uploadData planBTypeKfz.UploadImgRet
	unmarshalErr := json.Unmarshal([]byte(upload), &uploadData)
	if unmarshalErr != nil {
		return "", fmt.Errorf("解析上传图片数据失败 %v", unmarshalErr)
	}
	// 修复：先判断 ErrorResponse 是否为 nil
	if uploadData.ErrorResponse != nil && uploadData.ErrorResponse.Code != 0 {
		fmt.Println("错误码:", uploadData.ErrorResponse.Code)
		return "", fmt.Errorf("上传图片失败 %v", uploadData.ErrorResponse.SubMsg)
	}

	// 修复：判断 SuccessResponse 是否为 nil
	if uploadData.SuccessResponse == nil {
		return "", fmt.Errorf("上传图片成功但返回数据为空")
	}
	return uploadData.SuccessResponse.Image.Url, nil
}

// 商品新增
// @param logUuid 日志ID
// @param goodsInfo 商品信息
// @return GoodsAddResponseWrapper 商品新增结果
// @return string 商品新增结果json
// @return error 错误信息
func addGoods(logUuid string, goodsInfo planBTypeKfz.GoodsAdd17) (planBTypeKfz.AddGoodsRet, string, error) {
	var goodsAdd planBTypeKfz.AddGoodsRet
	goodsInfoStr, jsonMarshalErr := json.Marshal(goodsInfo)
	if jsonMarshalErr != nil {
		return goodsAdd, "", jsonMarshalErr
	}
	//发送请求
	goodsAddStr, publishGoodsErr := golabl.KfzDll.PublishGoods(golabl.Config.KfzConfig.AppId, golabl.Config.KfzConfig.AppSecret, golabl.Task.Header.ShopMsg.Token, string(goodsInfoStr))
	//判断是否成功
	if strings.Contains(goodsAddStr, "失败") || strings.Contains(goodsAddStr, "错误码") {
		//记录请求日志
		addGoodsReqMsg := fmt.Sprintf(`
	════════════════════════════════════════════════════════════════
	【孔夫子商品添加请求】
	请求ID: %s
	时间: %s
	参数: %s
	════════════════════════════════════════════════════════════════`,
			logUuid,
			time.Now().Format("2006-01-02 15:04:05.000"),
			string(goodsInfoStr))

		tool.LoggingMiddleware(logs.LOG_LEVEL_INFO, addGoodsReqMsg)
		return goodsAdd, goodsAddStr, errors.New("孔夫子 GoodsAdd 错误:" + goodsAddStr)
	}
	if publishGoodsErr != nil {
		return goodsAdd, "", publishGoodsErr
	}
	jsonUnmarshal := json.Unmarshal([]byte(goodsAddStr), &goodsAdd)
	if jsonUnmarshal != nil {
		return goodsAdd, "", fmt.Errorf("解析孔夫子 GoodsAdd 接口返回json失败: %v", jsonUnmarshal)
	}
	return goodsAdd, goodsAddStr, nil
}

// phaseOneGoods 第一阶段拉取商品信息（获取总数）
// @param pageSize 每页数量
// @param totalFetched 获取到的商品总数
// @return error 错误信息
func phaseOneGoods(pageSize int, totalFetched *int) error {
	// 第一阶段：获取第一页商品，获取总数
	getGoodsListReq := planBTypeKfz.GetGoodsListReq{
		Type:      "sale", // 第一阶段不传时间范围，获取所有商品的第一页
		PageNum:   1,
		PageSize:  pageSize,
		SortOrder: "addTime",
		SortType:  "ASC",
	}

	getGoodsListReqJson, marshalErr := json.Marshal(getGoodsListReq)
	if marshalErr != nil {
		return fmt.Errorf("参数转换错误 %v", marshalErr)
	}
	// 发送请求
	goodsListResp, goodsListRespErr := sendGoodsListRequest(string(getGoodsListReqJson))
	if goodsListRespErr != nil {
		return fmt.Errorf("获取商品列表失败: %v", goodsListRespErr)
	}

	if goodsListResp.SuccessResponse == nil {
		return fmt.Errorf("孔夫子商品列表返回数据为空")
	}

	// 更新进度总数
	totalCount := goodsListResp.SuccessResponse.Total
	fmt.Println("总数  ", totalCount)
	if updateTaskHeaderErr := service.SetTaskCount(strconv.Itoa(totalCount)); updateTaskHeaderErr != nil {
		return updateTaskHeaderErr
	}

	// 收集商品数据并写入 body_wait
	for _, goods := range goodsListResp.SuccessResponse.List {
		*totalFetched++

		// 将商品转为 JSON 存入 Detail.Error
		jsonData, jsonMarshalErr := json.Marshal(goods)
		if jsonMarshalErr != nil {
			return fmt.Errorf("将商品转为json失败: %v\n", jsonMarshalErr)
		}

		// 解析添加时间转为时间戳（秒）
		addTimeUnix := parseAddTimeToUnix(goods.AddTime)

		// 构建 bodyWait
		bodyWait := planAType.TaskBody{
			BookInfo: planAType.BookInfo{
				Isbn:            goods.Isbn,
				BookName:        goods.ItemName,
				Author:          goods.Author,
				Publishing:      goods.Press,
				PublicationDate: goods.PubDate,
				Binding:         goods.Binding,
				PagesCount:      int64(goods.PageNum),
				Format:          0,
				WordsCount:      0,
				Price:           addTimeUnix, // 使用添加时间作为 Price 字段存储
			},
			Detail: planAType.TaskDetail{
				Error:   string(jsonData),
				GoodsId: goods.ItemId,
				Stock:   int64(goods.Number),
			},
		}

		// 将 bodyWait 转为 json
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

	fmt.Printf("第一阶段 - 总数：%v 当前已取出：%v \n", totalCount, *totalFetched)
	return nil
}

// PhaseTwoGoods 第二阶段拉取商品信息（按时间范围分批拉取）
// @param pageSize 每页数量
// @param totalFetched 获取到的商品总数
// @param lastAddTime 最后一条数据的添加时间（秒级时间戳）
// @param maxRecordsPerRange 每次请求的时间范围最多获取的记录数
// @param timeRangeDays 时间范围天数
// @return error 错误信息
func PhaseTwoGoods(pageSize int, totalFetched *int, lastAddTime *int64, maxRecordsPerRange int, timeRangeDays int) error {
	fmt.Printf("[DEBUG] PhaseTwoGoods called, lastAddTime=%d, currentTime=%d\n", *lastAddTime, time.Now().Unix())
	if *lastAddTime > 0 {
		currentAddTimeFrom := *lastAddTime
		maxLoopCount := 100 // 最大循环次数保护
		loopCount := 0
		var lastPageGoodsList []planBTypeKfz.KfzGoodsItem // 记录上一页的商品列表

		// 去重：加载 body_wait 中已有的 GoodsId，避免 phaseOneGoods 和 PhaseTwoGoods 重复写入同一商品
		bodyWaitCount, _ := service.GetTaskBodyWaitCount()
		existingGoodsIds := make(map[int64]bool)
		if bodyWaitCount > 0 {
			existingPage := 1
			existingPageSize := 1000
			for {
				existingList, _ := service.GetTaskBodyWaitList(existingPage, existingPageSize)
				if len(existingList) == 0 {
					break
				}
				for _, v := range existingList {
					var taskBody planAType.TaskBody
					if err := json.Unmarshal([]byte(v), &taskBody); err == nil {
						existingGoodsIds[taskBody.Detail.GoodsId] = true
					}
				}
				existingPage++
			}
			fmt.Printf("[dedup] PhaseTwoGoods: 加载 body_wait 中已有 %d 个商品ID用于去重\n", len(existingGoodsIds))
		}

		for loopCount < maxLoopCount {
			loopCount++

			// 检查开始时间是否已超过当前时间
			if currentAddTimeFrom > time.Now().Unix() {
				fmt.Printf("开始时间 %d 已超过当前时间，停止获取\n", currentAddTimeFrom)
				break
			}

			// 每次循环都重新设置结束时间为开始时间+timeRangeDays天
			currentAddTimeEnd := currentAddTimeFrom + int64(timeRangeDays)*24*60*60
			fmt.Printf("开始获取时间范围: %d (%s) 到 %d (%s)\n",
				currentAddTimeFrom, time.Unix(currentAddTimeFrom, 0).Format("2006-01-02 15:04:05"),
				currentAddTimeEnd, time.Unix(currentAddTimeEnd, 0).Format("2006-01-02 15:04:05"))

			currentPage := 1
			batchGoodsCount := 0
			lastItemAddTime := int64(0)
			hasDataInRange := false
			lastPageGoodsList = nil // 重置上一页商品列表

			// 在当前时间范围内分页获取数据
			for {
				getGoodsListReq := planBTypeKfz.GetGoodsListReq{
					Type:         "sale",
					PageNum:      currentPage,
					PageSize:     pageSize,
					AddTimeBegin: formatUnixTime(currentAddTimeFrom),
					AddTimeEnd:   formatUnixTime(currentAddTimeEnd),
					SortOrder:    "addTime",
					SortType:     "ASC",
				}

				getGoodsListReqJson, marshalErr := json.Marshal(getGoodsListReq)
				if marshalErr != nil {
					return fmt.Errorf("参数转换错误 %v", marshalErr)
				}

				goodsListResp, goodsListRespErr := sendGoodsListRequest(string(getGoodsListReqJson))
				if goodsListRespErr != nil {
					// DLL 不可用时直接退出第二阶段，跳到去重写入步骤
					fmt.Printf("获取商品列表失败，退出第二阶段: %v\n", goodsListRespErr)
					return nil
				}

				if goodsListResp.SuccessResponse == nil || len(goodsListResp.SuccessResponse.List) == 0 {
					// 如果当前页是第一页且没有数据
					if currentPage == 1 {
						// 整个时间范围都没有数据，直接推进到结束时间
						currentAddTimeFrom = currentAddTimeEnd
						fmt.Printf("时间范围 %d - %d 内无数据，推进开始时间到: %d\n", currentAddTimeFrom-int64(timeRangeDays)*24*60*60, currentAddTimeEnd, currentAddTimeFrom)
						break
					}

					// 当前页没有数据，但上一页有数据
					if len(lastPageGoodsList) > 0 {
						lastItemOfLastPage := lastPageGoodsList[len(lastPageGoodsList)-1]
						newStartTime := parseAddTimeToUnix(lastItemOfLastPage.AddTime)

						if newStartTime > currentAddTimeFrom {
							currentAddTimeFrom = newStartTime
							fmt.Printf("当前页无数据，使用上一页最后一条商品时间作为新开始时间: %d\n", currentAddTimeFrom)
						} else {
							currentAddTimeFrom = newStartTime + 1
							fmt.Printf("当前页无数据，时间相同，将时间加1秒推进: %d\n", currentAddTimeFrom)
						}
					} else {
						currentAddTimeFrom = currentAddTimeEnd
						fmt.Printf("当前页无数据且无上一页数据，将开始时间推进到结束时间: %d\n", currentAddTimeFrom)
					}

					hasDataInRange = false
					break
				}

				// 有数据，记录上一页的商品列表
				lastPageGoodsList = goodsListResp.SuccessResponse.List
				hasDataInRange = true

				// 收集商品数据并统计
				for _, goods := range goodsListResp.SuccessResponse.List {
					*totalFetched++

					// 将商品转为 JSON 存入 Detail.Error
					jsonData, jsonMarshalErr := json.Marshal(goods)
					if jsonMarshalErr != nil {
						return fmt.Errorf("将商品转为json失败: %v\n", jsonMarshalErr)
					}

					// 解析添加时间转为时间戳（秒）
					addTimeUnix := parseAddTimeToUnix(goods.AddTime)

					// 构建 bodyWait
					bodyWait := planAType.TaskBody{
						BookInfo: planAType.BookInfo{
							Isbn:            goods.Isbn,
							BookName:        goods.ItemName,
							Author:          goods.Author,
							Publishing:      goods.Press,
							PublicationDate: goods.PubDate,
							Binding:         goods.Binding,
							PagesCount:      int64(goods.PageNum),
							Format:          0,
							WordsCount:      0,
							Price:           addTimeUnix,
						},
						Detail: planAType.TaskDetail{
							Error:   string(jsonData),
							GoodsId: goods.ItemId,
							Stock:   int64(goods.Number),
						},
					}

					bodyWaitJson, jsonMarshalErr := json.Marshal(bodyWait)
					if jsonMarshalErr != nil {
						return fmt.Errorf("将bodyWait转为json失败: %v\n", jsonMarshalErr)
					}

					// 去重：检查 GoodsId 是否已存在于 body_wait（phaseOneGoods 可能已写过）
					if existingGoodsIds[goods.ItemId] {
						fmt.Printf("[dedup] 跳过重复商品 GoodsId=%d (%s)\n", goods.ItemId, goods.ItemName)
						continue
					}
					existingGoodsIds[goods.ItemId] = true
					addTaskToBodyWaitErr := service.AddTaskToBodyWait(string(bodyWaitJson))
					if addTaskToBodyWaitErr != nil {
						return addTaskToBodyWaitErr
					}
				}

				batchGoodsCount += len(goodsListResp.SuccessResponse.List)

				// 记录最后一条商品的添加时间
				lastItem := goodsListResp.SuccessResponse.List[len(goodsListResp.SuccessResponse.List)-1]
				lastItemAddTime = parseAddTimeToUnix(lastItem.AddTime)

				fmt.Printf("第二阶段 - 当前时间范围已获取: %d 条，累计总数: %d，当前页码: %d，最后商品时间: %d\n",
					batchGoodsCount, *totalFetched, currentPage, lastItemAddTime)

				// 判断是否需要结束当前时间范围
				if batchGoodsCount >= maxRecordsPerRange || len(goodsListResp.SuccessResponse.List) < pageSize {
					if lastItemAddTime == currentAddTimeFrom {
						currentAddTimeFrom = lastItemAddTime + 1
						fmt.Printf("最后商品时间与开始时间相同，推进1秒: %d -> %d\n", lastItemAddTime, currentAddTimeFrom)
					} else {
						currentAddTimeFrom = lastItemAddTime
					}
					fmt.Printf("当前时间范围已获取 %d 条数据，准备进入下一时间范围 更新开始时间为: %d \n", batchGoodsCount, currentAddTimeFrom)
					break
				}

				currentPage++

				// 获取 footer 信息
				if getTaskFooterErr := service.GetTaskFooter(); getTaskFooterErr != nil {
					return getTaskFooterErr
				}
				con := int64(len(goodsListResp.SuccessResponse.List))
				if con >= golabl.Task.Footer.TaskCountTrue {
					con = golabl.Task.Footer.TaskCountTrue - con
				}
				if updateTaskProgressErr := tool.UpdateTaskProgress(con); updateTaskProgressErr != nil {
					return updateTaskProgressErr
				}

				// 暂停200毫秒
				time.Sleep(200 * time.Millisecond)
			}

			// 判断是否需要继续循环
			if !hasDataInRange {
				if currentAddTimeFrom > time.Now().Unix() {
					fmt.Printf("开始时间 %d 已超过当前时间，停止获取\n", currentAddTimeFrom)
					break
				}
				fmt.Printf("继续下一轮查询，新起始时间: %d\n", currentAddTimeFrom)
				continue
			}

			if batchGoodsCount < maxRecordsPerRange && currentAddTimeFrom > time.Now().Unix() {
				fmt.Printf("当前批次获取 %d 条数据，少于 %d，且开始时间已超过当前时间，已完成所有数据获取\n", batchGoodsCount, maxRecordsPerRange)
				break
			}

			// 暂停200毫秒
			time.Sleep(200 * time.Millisecond)
		}

		if loopCount >= maxLoopCount {
			fmt.Printf("警告：已达到最大循环次数 %d，强制退出\n", maxLoopCount)
		}
	}
	return nil
}

// deduplicateToBodyOver 读取body_wait去重复后写入到body_over中
// @param duplicateCount 重复商品数量
// @param uniqueCount 不重复商品数量
// @return error 错误信息
func deduplicateToBodyOver(duplicateCount *int, uniqueCount *int) error {
	page := 1
	pageSize := 1000
	var dataList []planBTypeKfz.KfzGoodsItem

	// 在循环外维护一个已处理的商品 ID集合
	processedGoodsIds := make(map[int64]bool)

	// 在循环前删除 body_over与body_backup，避免重复写入
	deleteTaskBodyOverErr := service.DeleteTaskBodyOver()
	if deleteTaskBodyOverErr != nil {
		return deleteTaskBodyOverErr
	}
	deleteTaskBodyBackupErr := service.DeleteTaskBodyBackup()
	if deleteTaskBodyBackupErr != nil {
		return deleteTaskBodyBackupErr
	}

	num := 0

	// 获取body_wait总数量
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
			var taskBody planAType.TaskBody
			jsonUnmarshalErr := json.Unmarshal([]byte(v), &taskBody)
			if jsonUnmarshalErr != nil {
				return fmt.Errorf("将json转为结构体失败: %v\n", jsonUnmarshalErr)
			}

			if !processedGoodsIds[taskBody.Detail.GoodsId] {
				processedGoodsIds[taskBody.Detail.GoodsId] = true
				*uniqueCount++

				var kfzGoodsItem planBTypeKfz.KfzGoodsItem
				jsonUnmarshalErr = json.Unmarshal([]byte(taskBody.Detail.Error), &kfzGoodsItem)
				if jsonUnmarshalErr != nil {
					return fmt.Errorf("将json转为结构体失败: %v\n", jsonUnmarshalErr)
				}

				dataList = append(dataList, kfzGoodsItem)

				// 写入到body_over
				taskBody.Detail.Status = 1
				addTaskToBodyOverErr := service.AddTaskToBodyOver(taskBody, []string{"body_over", "body_backup"})
				if addTaskToBodyOverErr != nil {
					return addTaskToBodyOverErr
				}
			} else {
				*duplicateCount++
			}
		}

		// 将获取的数据推送写入店铺商品数据接口（待实现）
		if len(dataList) > 0 {
			pageFlag := 0
			//如果是最后一批数据则将 pageFlag 设置为 1
			if page == int(pageTotal) {
				pageFlag = 1
			}
			pushShopGoodsDataRet, pushShopGoodsDataErr := pushShopGoodsData(dataList, pageFlag)
			if pushShopGoodsDataErr != nil {
				return pushShopGoodsDataErr
			}
			var pushShopGoodsDatas planBTypeKfz.AddGoodsToErpRet
			jsonUnmarshalErr := json.Unmarshal([]byte(pushShopGoodsDataRet), &pushShopGoodsDatas)
			if jsonUnmarshalErr != nil {
				return fmt.Errorf("将json转为结构体失败: %v\n", jsonUnmarshalErr)
			}
			if pushShopGoodsDatas.Code != 200 {
				return fmt.Errorf("推送商品数据失败: %v\n", pushShopGoodsDatas.Msg)
			}
			fmt.Printf("当前页 %v 总页数 %v 当前数据量 %v 总数据量 %v \n", page, pageTotal, len(dataList), num)
		}

		num = num + len(dataList)
		page++

		// 获取 footer信息
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

		// 清空 dataList
		dataList = []planBTypeKfz.KfzGoodsItem{}
		// 暂停1秒
		time.Sleep(1 * time.Second)
	}

	// 删除body_wait
	deleteTaskBodyWaitErr := service.DeleteTaskBodyWait()
	if deleteTaskBodyWaitErr != nil {
		return deleteTaskBodyWaitErr
	}
	return nil
}

// sendGoodsListRequest 发送商品列表请求
// @param reqJson 请求JSON
// @return GetGoodsListResp 响应结构体
// @return error 错误
func sendGoodsListRequest(reqJson string) (*planBTypeKfz.GetGoodsListResp, error) {
	// 使用超时包装，防止 DLL 调用永久阻塞
	// 使用容量为1的缓冲通道，避免 goroutine 泄露
	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		fmt.Println("[DEBUG] goroutine started, about to call GetGoodsList...")
		goodsListStr, getGoodsListErr := golabl.KfzDll.GetGoodsList(
			golabl.Config.KfzConfig.AppId,
			golabl.Config.KfzConfig.AppSecret,
			golabl.Task.Header.ShopMsg.Token,
			reqJson)
		fmt.Println("[DEBUG] GetGoodsList returned, err:", getGoodsListErr)
		if getGoodsListErr != nil {
			errCh <- getGoodsListErr
			return
		}
		// 使用 select+default 确保超时后 goroutine 能安全退出
		select {
		case resultCh <- goodsListStr:
		default:
			// channel 已满（超时已触发），直接退出
			fmt.Println("[DEBUG] resultCh full, goroutine exiting")
		}
	}()

	// 等待结果或超时（30秒）
	select {
	case goodsListStr := <-resultCh:
		// 正常返回
		var goodsListResp planBTypeKfz.GetGoodsListResp
		fmt.Println(goodsListStr)
		unmarshalErr := json.Unmarshal([]byte(goodsListStr), &goodsListResp)
		if unmarshalErr != nil {
			fmt.Printf("解析孔夫子商品列表返回json失败: %v [孔夫子数据：%v]", unmarshalErr, goodsListStr)
			return nil, fmt.Errorf("解析孔夫子商品列表返回json失败: %v [孔夫子数据：%v]", unmarshalErr, goodsListStr)
		}
		return &goodsListResp, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("DLL 调用 GetGoodsList 超时（30秒），可能是 DLL 函数不可用或网络问题")
	}
}

// parseAddTimeToUnix 获取商品的添加时间（Unix时间戳，秒）
// @param addTime 添加时间（Unix时间戳，秒级）
// @return int64 Unix时间戳（秒）
func parseAddTimeToUnix(addTime int64) int64 {
	return addTime
}

// formatUnixTime 将Unix时间戳（秒）转为孔夫子时间格式字符串
// @param unixTime Unix时间戳（秒）
// @return string 时间字符串（格式：yyyy-MM-dd HH:mm:ss）
func formatUnixTime(unixTime int64) string {
	if unixTime <= 0 {
		return ""
	}
	return time.Unix(unixTime, 0).Format("2006-01-02 15:04:05")
}

// 上架
func executeGoodsLaunch(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	// 上架商品
	launchGoodsInfo := planBTypeKfz.Product{ItemId: strconv.FormatInt(taskMsg.Detail.GoodsId, 10)}
	//转为json
	jsonData, marshalErr := json.Marshal(launchGoodsInfo)
	if marshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, marshalErr)
	}
	var launchGoods planBTypeKfz.ProductRet
	launchGoodsStr, kfzPutOnSaleErr := golabl.KfzDll.PutOnSale(golabl.Config.KfzConfig.AppId, golabl.Config.KfzConfig.AppSecret, golabl.Task.Header.ShopMsg.Token, string(jsonData))
	if kfzPutOnSaleErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, kfzPutOnSaleErr)
	}
	unmarshalErr := json.Unmarshal([]byte(launchGoodsStr), &launchGoods)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, unmarshalErr)
	}
	if launchGoods.ErrorResponse != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("上架商品失败 %s", launchGoods.ErrorResponse))
	}
	return tool.ReturnSuccess(taskMsg)
}

// 下架
func executeGoodsDownShelf(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	// 下架商品
	launchGoodsInfo := planBTypeKfz.Product{ItemId: strconv.FormatInt(taskMsg.Detail.GoodsId, 10)}
	//转为json
	jsonData, marshalErr := json.Marshal(launchGoodsInfo)
	if marshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, marshalErr)
	}
	var launchGoods planBTypeKfz.ProductRet
	launchGoodsStr, kfzPutOffSaleErr := golabl.KfzDll.PutOffSale(golabl.Config.KfzConfig.AppId, golabl.Config.KfzConfig.AppSecret, golabl.Task.Header.ShopMsg.Token, string(jsonData))
	if kfzPutOffSaleErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, kfzPutOffSaleErr)
	}
	unmarshalErr := json.Unmarshal([]byte(launchGoodsStr), &launchGoods)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, unmarshalErr)
	}
	if launchGoods.ErrorResponse != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("下架商品失败 %s", launchGoods.ErrorResponse))
	}
	return tool.ReturnSuccess(taskMsg)
}

// 改价格
func executeGoodsUpdatePrice(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	//将价格由分转为元
	price := tool.FenToYuanFloat64(taskMsg.Detail.Price)
	// 改价格
	launchGoodsInfo := planBTypeKfz.UpdatePriceReq{ItemId: strconv.FormatInt(taskMsg.Detail.GoodsId, 10), Price: price}
	//转为json
	jsonData, marshalErr := json.Marshal(launchGoodsInfo)
	if marshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, marshalErr)
	}
	var launchGoods planBTypeKfz.UpdatePriceRet
	launchGoodsStr, kfzUpdateGoodsPriceErr := golabl.KfzDll.UpdateGoodsPrice(golabl.Config.KfzConfig.AppId, golabl.Config.KfzConfig.AppSecret, golabl.Task.Header.ShopMsg.Token, string(jsonData))
	if kfzUpdateGoodsPriceErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, kfzUpdateGoodsPriceErr)
	}
	unmarshalErr := json.Unmarshal([]byte(launchGoodsStr), &launchGoods)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, unmarshalErr)
	}
	if launchGoods.ErrorResponse != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("改价格商品失败 %s", launchGoods.ErrorResponse))
	}
	return tool.ReturnSuccess(taskMsg)
}

// 改库存
func executeGoodsUpdateStock(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	// 改库存
	launchGoodsInfo := planBTypeKfz.UpdateStockReq{ItemId: strconv.FormatInt(taskMsg.Detail.GoodsId, 10), Number: taskMsg.Detail.Stock}
	//转为json
	jsonData, marshalErr := json.Marshal(launchGoodsInfo)
	if marshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, marshalErr)
	}
	var launchGoods planBTypeKfz.ProductRet
	launchGoodsStr, kfzUpdateGoodsStockErr := golabl.KfzDll.UpdateGoodsStock(golabl.Config.KfzConfig.AppId, golabl.Config.KfzConfig.AppSecret, golabl.Task.Header.ShopMsg.Token, string(jsonData))
	if kfzUpdateGoodsStockErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, kfzUpdateGoodsStockErr)
	}
	unmarshalErr := json.Unmarshal([]byte(launchGoodsStr), &launchGoods)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, unmarshalErr)
	}
	if launchGoods.ErrorResponse != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("改库存商品失败 %s", launchGoods.ErrorResponse))
	}
	return tool.ReturnSuccess(taskMsg)
}

// 添加商品到ERP接口
func pushShopGoodsData(shopGoodsList []planBTypeKfz.KfzGoodsItem, pageFlag int) (string, error) {

	// 拼接商品列表
	var goodsList []map[string]string
	for _, shopGoods := range shopGoodsList {

		goods := map[string]string{
			"isbn":          shopGoods.Isbn,
			"itemName":      shopGoods.ItemName,
			"price":         fmt.Sprintf("%.2f", shopGoods.Price),
			"quality":       strconv.Itoa(int(shopGoods.Quality)),
			"author":        shopGoods.Author,
			"press":         shopGoods.Press,
			"pubDate":       shopGoods.PubDate,
			"itemId":        strconv.FormatInt(shopGoods.ItemId, 10),
			"addTime":       strconv.FormatInt(shopGoods.AddTime, 10),
			"beginSaleTime": strconv.FormatInt(int64(shopGoods.BeginSaleTime), 10),
			"isDraft":       strconv.Itoa(shopGoods.IsDraft),
			"discount":      strconv.Itoa(shopGoods.Discount),
			"stock":         strconv.Itoa(shopGoods.Number),
			"myCatId":       strconv.Itoa(shopGoods.MyCatId),
			"bearShipping":  shopGoods.BearShipping,
			"weight":        fmt.Sprintf("%.2f", shopGoods.Weight),
			"catId":         strconv.FormatInt(int64(shopGoods.CatId), 10),
			"isNewBook":     strconv.Itoa(shopGoods.IsNewBook),
			"bizType":       strconv.Itoa(shopGoods.BizType),
			"certifyStatus": shopGoods.CertifyStatus,
			"weightPiece":   fmt.Sprintf("%.2f", shopGoods.WeightPiece),
			"mouldId":       strconv.Itoa(shopGoods.MouldId),
			"booklibId":     strconv.Itoa(int(shopGoods.BooklibId)),
			"isOnSale":      strconv.Itoa(shopGoods.IsOnSale),
			"isDelete":      strconv.Itoa(shopGoods.IsDelete),
			"updateTime":    shopGoods.UpdateTime,
			"endSaleTime":   strconv.FormatInt(int64(shopGoods.EndSaleTime), 10),
			"userId":        strconv.Itoa(int(shopGoods.UserId)),
			"imgUrl":        golabl.Config.FileUrl.KfzImgHttpUrl + shopGoods.ImgUrl,
			"oriPrice":      fmt.Sprintf("%.2f", shopGoods.OriPrice),
			"itemSn":        shopGoods.ItemSn,
		}

		goodsList = append(goodsList, goods)

	}

	shopGoodslistData := map[string]interface{}{
		"total": strconv.Itoa(len(goodsList)),
		"list":  goodsList,
	}

	data := map[string]interface{}{
		"shopId":            golabl.Task.Header.ShopId,
		"shopType":          golabl.Task.Header.ShopType,
		"token":             golabl.Task.Header.ShopMsg.Token,
		"sycFlag":           1,
		"taskId":            golabl.Task.TaskId,
		"shopGoodsResponse": shopGoodslistData, // 关键修改
		"pageFlag":          pageFlag,
	}

	//转为json
	jsonData, marshalErr := json.Marshal(data)
	if marshalErr != nil {
		return "", marshalErr
	}

	response, err := tool.PostJSON(golabl.Config.FileUrl.KfzAddGoodsUrl, string(jsonData))
	if err != nil {
		return "", err
	}

	return response, nil
}
