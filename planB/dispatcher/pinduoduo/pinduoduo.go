package pinduoduo

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"planA/planB/initialization/golabl"
	"planA/planB/logic"
	"planA/planB/modules/logs"
	"planA/planB/service"
	"planA/planB/tool"
	planBTypeModules "planA/planB/type/modules"
	planBTypePinduoduo "planA/planB/type/pinduoduo"
	planAType "planA/type"
	"strconv"
	"strings"
	"time"
)

type PinDuoDuo struct {
}

// NewPinDuoDuo 创建拼多多平台
// @return *PinDuoDuo
func NewPinDuoDuo() *PinDuoDuo {
	return &PinDuoDuo{}
}

// AddGoodsTask 添加商品
// @param taskMsg 任务内容
// @return string body 信息
// @return error 错误
func (pinDuoDuo *PinDuoDuo) AddGoodsTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}
	taskMsg, publishGoodsErr := publishGoods(logUuid, taskMsg)
	if publishGoodsErr != nil {
		return "", publishGoodsErr
	}
	return tool.ReturnSuccess(taskMsg)
}

// GetGoodsTask 获取商品
// @return string body 信息
// @return error 错误
func (pinDuoDuo *PinDuoDuo) GetGoodsTask() (string, error) {
	// 生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}

	const pageSize = 100
	const maxPage = 100
	const maxRecordsPerRange = 10000 // 每个时间范围最多获取10000条

	var lastCreatedAt int64 = 0

	// 统计变量
	totalFetched := 0   // 总共获取到的商品数（包括重复）
	duplicateCount := 0 // 重复商品数量
	uniqueCount := 0    // 不重复商品数量

	//查询body_wait是否存在，如果存在则证明不是第一次执行要获取body_wait中最后一条数据的创建时间作为查询的开始时间
	exist, isTaskBodyWaitExistErr := service.IsTaskBodyWaitExist()
	if isTaskBodyWaitExistErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, isTaskBodyWaitExistErr)
	}
	if exist {
		// 获取最后一条数据的创建时间
		lastBodyWaitDataJson, getLastGoodsCreateTimeErr := service.GetTaskBodyWaitLast()
		if getLastGoodsCreateTimeErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, getLastGoodsCreateTimeErr)
		}
		// 解析 lastBodyWaitData 到结构体
		var lastBodyWaitData planAType.TaskBody
		unmarshalErr := json.Unmarshal([]byte(lastBodyWaitDataJson), &lastBodyWaitData)
		if unmarshalErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, unmarshalErr)
		}
		//将数据的创建时间给到 lastCreatedAt
		lastCreatedAt = lastBodyWaitData.BookInfo.Price
		//第二阶段 获取商品
		phaseTwoGoodsErr := PhaseTwoGoods(pageSize, &totalFetched, &lastCreatedAt, maxRecordsPerRange)
		if phaseTwoGoodsErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, phaseTwoGoodsErr)
		}
	} else {
		//第一阶段 拉取任务数据
		firstTimeGoodsErr := phaseOneGoods(pageSize, maxPage, &totalFetched, &lastCreatedAt)
		if firstTimeGoodsErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, firstTimeGoodsErr)
		}
		//第二阶段 获取商品
		phaseTwoGoodsErr := PhaseTwoGoods(pageSize, &totalFetched, &lastCreatedAt, maxRecordsPerRange)
		if phaseTwoGoodsErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, phaseTwoGoodsErr)
		}
	}

	//更新状态为推送中
	updateTaskStatusErr := service.UpdateTaskStatus(planAType.TaskStatusPushTaskStatus)
	if updateTaskStatusErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, updateTaskStatusErr)
	}

	//重新设置任务进度
	if updateTaskHeaderErr := service.SetTaskCount(strconv.FormatInt(golabl.Task.Footer.TaskCountTrue, 10)); updateTaskHeaderErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, updateTaskHeaderErr)
	}

	//去重复与保存
	deduplicateToBodyOverErr := deduplicateToBodyOver(&duplicateCount, &uniqueCount)
	if deduplicateToBodyOverErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, deduplicateToBodyOverErr)
	}

	// 输出统计信息
	statsLogMsg := fmt.Sprintf(`
════════════════════════════════════════════════════════════════
【拼多多店铺拉取】
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
func (pinDuoDuo *PinDuoDuo) OperationGoodsTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}
	switch taskMsg.Detail.Status {
	case 1, 2:
		return setSaleStatusGoodsTask(logUuid, taskMsg) //设置商品上下架状态 status=1 上架 status=2 下架  {"book_info":{"isbn":"9787543982888"},"detail":{"goods_id":936170582125,"status":2}}
	case 4:
		return updateGoodsQuantity(logUuid, taskMsg, 1, 0) //修改商品库存 {"book_info":{"isbn":"9787532080786"},"detail":{"goods_id":935177284615,"status":4,"stock":2,"sku_id":1882660479308}}
	case 5:
		return updateSkuPrice(logUuid, taskMsg) //修改商品价格 {"book_info":{"isbn":"9787543982888"},"detail":{"goods_id":939229985495,"status":5,"price":5000,"sku_id":1886207421871}}
	case 6:
		//发布商品
		taskMsg, publishGoodsErr := publishGoods(logUuid, taskMsg)
		if publishGoodsErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, publishGoodsErr)
		}
		return tool.ReturnSuccess(taskMsg)
	case 7:
		//下架
		taskMsg.Detail.Status = 2
		_, setSaleStatusGoodsTaskErr := setSaleStatusGoodsTask(logUuid, taskMsg)
		if setSaleStatusGoodsTaskErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, setSaleStatusGoodsTaskErr)
		}
		//删除商品
		logic.DelTask(taskMsg)
		//发布商品
		taskMsg, publishGoodsErr := publishGoods(logUuid, taskMsg)
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
func (pinDuoDuo *PinDuoDuo) IncStockTask(taskMsg planAType.TaskBody) (string, error) {
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
		task, addGoodsTaskErr := publishGoods(logUuid, taskMsg)
		if addGoodsTaskErr != nil {
			return "", addGoodsTaskErr
		}

		//根据回调查询商品回到信息直到成功
		//found := false
		//startTime := time.Now()
		//maxDuration := 3 * time.Minute // 最大查询时间 3分钟

		//for !found {
		//	// 检查是否超时
		//	if time.Since(startTime) > maxDuration {
		//		fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~超时了~~~~~~~~~~~~~~~~~~~~~~~~~~")
		//		found = true
		//		break // 跳出内层循环
		//	}
		list, getPddNoticeMsgErr := service.GetPddNoticeMsg(golabl.Task.TaskId)
		if getPddNoticeMsgErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, getPddNoticeMsgErr)
		}
		for _, v := range list {
			var pddNoticeMsg planBTypePinduoduo.PddNoticeMsg
			unmarshalErr := json.Unmarshal([]byte(v), &pddNoticeMsg)
			if unmarshalErr != nil {
				return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, unmarshalErr)
			}
			//必须是发布上架
			if pddNoticeMsg.Type == "pdd_goods_GoodsOnShelf" {
				// task.Detail.GoodsId 转为字符串
				goodsId := strconv.FormatInt(task.Detail.GoodsId, 10)
				if pddNoticeMsg.GoodsId == goodsId {
					fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~找到了~~~~~~~~~~~~~~~~~~~~~~~~~~")
					//跳出最外层循环
					//found = true
					break // 跳出内层循环
				}
			}
		}
		//// 避免过于频繁的查询，可以添加短暂延迟
		//if !found {
		//	time.Sleep(1 * time.Second)
		//}
		//}
		task.Detail.Error = "发布成功！"
		return tool.ReturnSuccess(task)
	} else {
		// 当前任务的价格
		taskPrice := taskMsg.Detail.Price // 单位：分

		// 价格 + 运费（如果 PriceType != "0"）
		if golabl.Task.Header.PriceType != "0" {
			taskPrice = taskPrice + taskMsg.Detail.ShippingCost
		}

		// 价格模板计算
		taskPrice = tool.BuildPrice(golabl.Task.Header.PriceMod, taskPrice)
		if taskPrice == 0 {
			taskMsg.Detail.Error = "任务价格不在价格模板区间内！"
			return tool.ReturnSuccess(taskMsg)
		}

		// 1元 = 100分，价格相差1元以上即 >= 100分
		const priceDiffThreshold = 100 // 1元

		// 收集所有匹配条件的商品（价格差<1元）
		var matchedItems []struct {
			TrilateralId string
			SkuId        string
			Stock        int64
			Price        int64
			PriceDiff    int64 // 价格差绝对值
		}

		// 记录不匹配原因
		var firstMismatchReason string

		for _, item := range getGoodsByShopIdAndIsbn.Data {
			// 解析价格（单位：分）
			itemPrice, _ := strconv.ParseInt(item.TotalPrice, 10, 64)
			// 解析库存
			itemStock, _ := strconv.ParseInt(item.Stock, 10, 64)

			// 计算价格差（绝对值）
			priceDiff := abs(itemPrice - taskPrice)

			// 价格相差1元以上
			if priceDiff >= priceDiffThreshold {
				if firstMismatchReason == "" {
					firstMismatchReason = fmt.Sprintf("商品[%s]价格相差超过1元: 任务价格=%d分, 商品价格=%d分, 差价=%d分", item.TrilateralId, taskPrice, itemPrice, priceDiff)
				}
				continue
			}

			// 价格相差小于1元 → 加入候选列表
			matchedItems = append(matchedItems, struct {
				TrilateralId string
				SkuId        string
				Stock        int64
				Price        int64
				PriceDiff    int64
			}{
				TrilateralId: item.TrilateralId,
				SkuId:        item.SkuId,
				Stock:        itemStock,
				Price:        itemPrice,
				PriceDiff:    priceDiff,
			})
		}

		// 逻辑：
		// 1. 所有价格相差1元以上 → 重新发布
		// 2. 否则 → 找到价格相差最小的增加库存，如果多个最小差价一样则对第一条增加库存
		if len(matchedItems) == 0 {
			// 所有商品价格相差≥1元 → 重新发布
			fmt.Printf("[重新发布] %s\n", firstMismatchReason)

			task, addGoodsTaskErr := publishGoods(logUuid, taskMsg)
			if addGoodsTaskErr != nil {
				return "", addGoodsTaskErr
			}
			task.Detail.Error = "所有商品价格相差超过1元，重新发布成功！"
			return tool.ReturnSuccess(task)
		}

		// 找到价格相差最小的商品，如果多个最小差价一样则对第一条增加库存
		minDiff := int64(999999999)
		var targetItem struct {
			TrilateralId string
			SkuId        string
			Stock        int64
		}

		for _, item := range matchedItems {
			// 找到更小的差价，或者差价相等但为第一条
			if item.PriceDiff < minDiff || (item.PriceDiff == minDiff && targetItem.TrilateralId == "") {
				minDiff = item.PriceDiff
				targetItem.TrilateralId = item.TrilateralId
				targetItem.SkuId = item.SkuId
				targetItem.Stock = item.Stock
			}
		}

		// 将 targetItem.SkuId 转为 int64
		skuId, skuIdParseIntErr := strconv.ParseInt(targetItem.SkuId, 10, 64)
		if skuIdParseIntErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, skuIdParseIntErr)
		}

		// 将 targetItem.TrilateralId 转为 int64
		trilateralId, trilateralIdParseIntErr := strconv.ParseInt(targetItem.TrilateralId, 10, 64)
		if trilateralIdParseIntErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, trilateralIdParseIntErr)
		}

		//增量修改库存
		taskMsg.Detail.GoodsId = trilateralId
		taskMsg.Detail.SkuId = skuId
		quantity, updateGoodsQuantityErr := updateGoodsQuantity(logUuid, taskMsg, 2, targetItem.Stock)
		if updateGoodsQuantityErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, updateGoodsQuantityErr)
		}

		return quantity, nil
	}
}

// abs 返回绝对值
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func (pinDuoDuo *PinDuoDuo) SetGoodsTask() string {
	return ""
}

// *******************************私有方法************************************ //

// 构建商品属性列表
// @param isbn
// @param bookName 书名
// @param pageCount 页数
// @param price 价格
// @param publishingVid 出版社Vid
// @param author 作者
// @param format 开本
// @param binding 装帧
// @param wordsCount 字数
// @param publicationDate 出版时间
// @return []GoodsProperties 商品属性列表
func buildGoodsPropertiesList(isbn, bookName string, pageCount, price int64, publishingVid int64, author string, format int64, binding string, wordsCount int64, publicationDate string) []planBTypePinduoduo.GoodsProperties {
	var goodsPropertiesArr []planBTypePinduoduo.GoodsProperties
	//isbn
	if isbn != "" {
		goodsPropertiesIsbn := planBTypePinduoduo.GoodsProperties{
			RefPid: 425,
			Value:  isbn,
		}
		goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesIsbn)
	}

	//书名
	if bookName != "" {
		goodsPropertiesBookName := planBTypePinduoduo.GoodsProperties{
			RefPid: 876,
			Value:  bookName,
		}
		goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesBookName)
	}

	//页数
	if pageCount == 0 {
		pageCount = 200
	}
	goodsPropertiesPageNum := planBTypePinduoduo.GoodsProperties{
		RefPid:    692,
		Value:     strconv.FormatInt(pageCount, 10),
		ValueUnit: "页",
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesPageNum)

	//定价
	goodsPropertiesPrice := planBTypePinduoduo.GoodsProperties{
		RefPid:    879,
		Value:     strconv.FormatInt(price/100, 10),
		ValueUnit: "元",
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesPrice)

	//出版社
	if publishingVid != 0 {
		goodsPropertiesPublishing := planBTypePinduoduo.GoodsProperties{
			RefPid: 880,
			Vid:    publishingVid,
		}
		goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesPublishing)
	}

	//作者
	if author != "" {
		goodsPropertiesAuthor := planBTypePinduoduo.GoodsProperties{
			RefPid: 882,
			Value:  author,
		}
		goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesAuthor)
	}

	//开本
	if format != 0 {
		goodsPropertiesFormat := planBTypePinduoduo.GoodsProperties{
			RefPid: 890,
			Value:  strconv.FormatInt(format, 10),
		}
		goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesFormat)
	}

	//装帧类型
	if binding != "" {
		goodsPropertiesBinding := planBTypePinduoduo.GoodsProperties{
			RefPid: 888,
			Value:  binding,
		}
		goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesBinding)
	}

	//字数
	if wordsCount != 0 {
		goodsWordsCountBinding := planBTypePinduoduo.GoodsProperties{
			RefPid: 887,
			Value:  strconv.FormatInt(wordsCount, 10),
		}
		goodsPropertiesArr = append(goodsPropertiesArr, goodsWordsCountBinding)
	}

	//出版时间（部分数据出版时间是1970-01，视为没有出版时间）
	if publicationDate != "" && publicationDate != "1970-01" {
		goodsPublicationDateBinding := planBTypePinduoduo.GoodsProperties{
			RefPid: 881,
			Value:  publicationDate,
		}
		goodsPropertiesArr = append(goodsPropertiesArr, goodsPublicationDateBinding)
	}
	return goodsPropertiesArr
}

// sku规格生成
// @param price 价格
// @param thumbUrl 缩略图
// @param stock 库存
// @param outSkuSn 商品编码
// @param specName 规格名称
// @param isOnsale 上架状态
// @return Sku sku规格
// @return error 错误信息
func buildSkuList(price int64, thumbUrl string, stock int64, outSkuSn string, specChildName string, isOnsale int64) (planBTypePinduoduo.Sku, error) {
	//构建变量
	specId := golabl.Task.Header.ShopMsg.SpecId
	specName := golabl.Task.Header.ShopMsg.SpecName
	// 构建 Spec列表
	var sku planBTypePinduoduo.Sku
	goodsSpec, buildPddGoodsSpecIdErr := buildPddGoodsSpecId(specId, specChildName)
	if buildPddGoodsSpecIdErr != nil {
		return sku, buildPddGoodsSpecIdErr
	}

	// 构建SKU_Properties列表
	skuProperty := planBTypePinduoduo.SkuProperty{
		Punit:  specName,                        // 属性单位
		RefPid: specId,                          // 属性id
		Value:  goodsSpec.DllGoodsSpec.SpecName, // 属性值
		Vid:    goodsSpec.DllGoodsSpec.SpecID,   // 属性值id
	}
	skuProperties := []planBTypePinduoduo.SkuProperty{skuProperty}

	specIdList := "[" + strconv.FormatInt(skuProperty.Vid, 10) + "]"
	// 构建 SKU列表
	var onsale int64
	if isOnsale == 0 {
		onsale = 1
	} else {
		onsale = 0
	}
	sku = planBTypePinduoduo.Sku{
		IsOnsale:      onsale,        //上架状态，0-已下架，1-上架中
		LimitQuantity: 999,           //sku购买限制，只入参999
		MultiPrice:    price,         //团购价格，单位为分
		Price:         price + 100,   //单买价格，单位为分
		SkuProperties: skuProperties, //sku属性列表
		ThumbUrl:      thumbUrl,      //缩略图
		SpecIdList:    specIdList,    //商品规格列表
		Quantity:      stock,         //商品库存初始数量
		Weight:        250,           //重量单位g
		OutSkuSn:      outSkuSn,      //商品编码
	}
	return sku, nil
}

// buildPddGoodsSpecId 根据名称获取规格信息
// @param specId 商品规格id
// @param specName 规格名称
// @return DllGoodsSpec 规格信息
// @return error 错误信息
func buildPddGoodsSpecId(id int64, name string) (planAType.DllGoodsSpec, error) {
	var spec planAType.DllGoodsSpec
	specStr, err := golabl.PddDll.PddGoodsSpecIdGet(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, strconv.FormatInt(id, 10), name)
	if err != nil {
		return spec, err
	}
	// 解析JSON字符串
	err = json.Unmarshal([]byte(specStr), &spec)
	if err != nil {
		return spec, fmt.Errorf("解析拼多多 PddGoodsSpecIdGet 接口返回json失败: %v [拼多多数据：%v]", err, specStr)
	}
	return spec, nil
}

// 商品新增
// @param logUuid 日志ID
// @param goodsInfo 商品信息
// @return GoodsAddResponseWrapper 商品新增结果
// @return string 商品新增结果json
// @return error 错误信息
func addGoods(logUuid string, goodsInfo planBTypePinduoduo.GoodsAdd) (planBTypePinduoduo.GoodsAddResponseWrapper, string, error) {
	var goodsAdd planBTypePinduoduo.GoodsAddResponseWrapper
	goodsInfoStr, jsonMarshalErr := json.Marshal(goodsInfo)
	if jsonMarshalErr != nil {
		return goodsAdd, "", jsonMarshalErr
	}
	//发送请求
	goodsAddStr, pddGoodsAddErr := golabl.PddDll.PddGoodsAdd(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, string(goodsInfoStr))
	//判断是否成功
	if strings.Contains(goodsAddStr, "请求失败") || strings.Contains(goodsAddStr, "错误码") {
		//记录请求日志
		// 记录请求日志
		addGoodsReqMsg := fmt.Sprintf(`
════════════════════════════════════════════════════════════════
【拼多多商品添加请求】
请求ID: %s
时间: %s
参数: %s
════════════════════════════════════════════════════════════════`,
			logUuid,
			time.Now().Format("2006-01-02 15:04:05.000"),
			string(goodsInfoStr))

		tool.LoggingMiddleware(logs.LOG_LEVEL_INFO, addGoodsReqMsg)
		return goodsAdd, goodsAddStr, errors.New("拼多多 PddGoodsAdd 错误:" + goodsAddStr)
	}
	if pddGoodsAddErr != nil {
		return goodsAdd, "", pddGoodsAddErr
	}
	jsonUnmarshal := json.Unmarshal([]byte(goodsAddStr), &goodsAdd)
	if jsonUnmarshal != nil {
		return goodsAdd, "", fmt.Errorf("解析拼多多 PddGoodsAdd 接口返回json失败: %v", jsonUnmarshal)
	}
	return goodsAdd, goodsAddStr, nil
}

// 获取商品提交的商品详情
// @param goodsCommitId 商品提交ID
// @param goodsId 商品ID
// @return GoodsCommitDetailResponse 商品提交详情
// @return error 错误信息
func getGoodsCommitDetail(goodsCommitId int64, goodsId int64) (planBTypePinduoduo.GoodsCommitDetailResponse, string, error) {
	var goodsCommitDetail planBTypePinduoduo.GoodsCommitDetailResponse
	goodsCommitDetailStr, pddGoodsCommitDetailGetErr := golabl.PddDll.PddGoodsCommitDetailGet(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, strconv.FormatInt(goodsCommitId, 10), strconv.FormatInt(goodsId, 10))
	if pddGoodsCommitDetailGetErr != nil {
		return goodsCommitDetail, "", pddGoodsCommitDetailGetErr
	}
	unmarshalErr := json.Unmarshal([]byte(goodsCommitDetailStr), &goodsCommitDetail)
	if unmarshalErr != nil {
		return goodsCommitDetail, "", fmt.Errorf("解析拼多多 PddGoodsCommitDetailGet 接口返回json失败: %v [拼多多数据：%v]", unmarshalErr, goodsCommitDetailStr)
	}
	return goodsCommitDetail, goodsCommitDetailStr, nil
}

// 第一阶段拉取商品信息
// @param maxPage 最大页数
// @param pageSize 每页数量
// @param totalFetched 获取到的商品总数
// @param lastCreatedAt 最后一条数据的创建时间
// @return error 错误信息
func phaseOneGoods(maxPage int, pageSize int, totalFetched *int, lastCreatedAt *int64) error {
	// 第一阶段：获取第1页到第100页，不传入时间参数
	for page := 1; page <= maxPage; page++ {
		// 定义参数
		params := map[string]string{
			"accessToken": golabl.Task.Header.ShopMsg.Token,
			"page":        strconv.Itoa(page),
			"pageSize":    strconv.Itoa(pageSize),
		}

		goodsList, goodsListStr, getGoodsListErr := tool.GetPddGoodsList(params)
		if getGoodsListErr != nil {
			return fmt.Errorf("获取商品列表失败，页码: %d, 错误: %v", page, getGoodsListErr)
		}
		if goodsListStr == "{}" {
			//如果读取不到数据，重试一次
			fmt.Println("通过容器获取获取商品列表数据失败，重试一次")
			goodsList, goodsListStr, getGoodsListErr = tool.GetPddGoodsList(params)
			if getGoodsListErr != nil {
				return fmt.Errorf("获取商品列表失败，页码: %d, 错误: %v", page, getGoodsListErr)
			}
		}
		if goodsListStr == "{}" {
			fmt.Println("---------------------------------------错误！！！！goodsListStr----------------------------------")
			fmt.Println(goodsListStr)
			fmt.Println("---------------------------------------错误！！！！goodsListStr----------------------------------")
			return fmt.Errorf("容器返回数据为空")
		}

		//更新 header进度总数
		if page == 1 {
			// 更新进度
			fmt.Println("总数  ", strconv.Itoa(goodsList.TotalCount))
			if updateTaskHeaderErr := service.SetTaskCount(strconv.Itoa(goodsList.TotalCount)); updateTaskHeaderErr != nil {
				return updateTaskHeaderErr
			}
		}

		//如果是需要拉取详情的商品
		if golabl.Task.Header.TaskType == 4 {
			// 获取原始商品列表
			originalGoodsList := goodsList.GoodsList
			totalCount := len(originalGoodsList)

			if totalCount == 0 {
				return nil // 或继续后续处理
			}

			// 存储所有获取到的商品详情
			allGoodsDetailList := make([]planBTypePinduoduo.GoodsItem, 0, totalCount)

			// 每100条调用一次
			batchSize := 100
			n := 0
			for i := 0; i < totalCount; i += batchSize {
				// 计算当前批次的起始和结束位置
				end := i + batchSize
				if end > totalCount {
					end = totalCount
				}
				batch := originalGoodsList[i:end]
				n++
				// 调用接口获取商品详情
				fmt.Printf("第 %v 页  第 %v 次 \n", page, n)
				goodsDetailList, goodsDetailListStr, getPddGoodsDetailErr := tool.GetPddGoodsDetail(batch)
				if getPddGoodsDetailErr != nil {
					fmt.Println("----------------------------错误！！！！goodsDetailList-------------------------------")
					fmt.Printf("batch start %d end %d, batch size %d, total %d\n", i, end, len(batch), totalCount)
					fmt.Println(goodsDetailListStr)
					fmt.Println("----------------------------错误！！！！goodsDetailList-------------------------------")
					return getPddGoodsDetailErr
				}

				// 将当前批次的结果添加到总结果中
				allGoodsDetailList = append(allGoodsDetailList, goodsDetailList...)
			}

			// 赋值回原变量
			goodsList.GoodsList = allGoodsDetailList
		}
		// 收集商品数据并统计
		for _, goods := range goodsList.GoodsList {
			*totalFetched++
			//写入到数据库中
			//将goods转为json
			jsonData, jsonMarshalErr := json.Marshal(goods)
			if jsonMarshalErr != nil {
				return fmt.Errorf("将商品转为json失败: %v\n", jsonMarshalErr)
			}
			//写入到数据库中
			if len(goods.SkuList) <= 0 {
				return fmt.Errorf("商品sku列表为空 goodsId %v", goods.GoodsId)
			}
			bodyWait := planAType.TaskBody{
				BookInfo: planAType.BookInfo{
					Isbn:            goods.SkuList[0].OuterId,
					BookName:        goods.GoodsName,
					Author:          "",
					Publishing:      "",
					PublicationDate: "",
					Binding:         "",
					PagesCount:      0,
					WordsCount:      0,
					Format:          0,
					Price:           goods.CreatedAt,
				},
				Detail: planAType.TaskDetail{
					Status:  1,
					Error:   string(jsonData),
					GoodsId: goods.GoodsId,
					Stock:   int64(goods.SkuList[0].ReserveQuantity),
				},
			}
			// 将bodyWait 转为json
			bodyWaitJson, jsonMarshalErr := json.Marshal(bodyWait)
			if jsonMarshalErr != nil {
				return fmt.Errorf("将bodyWait转为json失败: %v\n", jsonMarshalErr)
			}
			//写入 body_wait
			addTaskToBodyWaitErr := service.AddTaskToBodyWait(string(bodyWaitJson))
			if addTaskToBodyWaitErr != nil {
				return addTaskToBodyWaitErr
			}
		}

		// 记录最后一页的最后一条数据的创建时间
		if page == maxPage && len(goodsList.GoodsList) > 0 {
			*lastCreatedAt = goodsList.GoodsList[len(goodsList.GoodsList)-1].CreatedAt
			fmt.Printf("最后一页，最后一条数据的创建时间: %v", *lastCreatedAt)
		}

		// 如果没有更多数据，提前退出
		if len(goodsList.GoodsList) == 0 {
			fmt.Println("没有更多数据，退出循环 ")
			fmt.Println(goodsListStr)
			break
		}

		// 获取 footer信息
		if getTaskFooterErr := service.GetTaskFooter(); getTaskFooterErr != nil {
			return getTaskFooterErr
		}
		con := int64(len(goodsList.GoodsList))
		if con >= golabl.Task.Footer.TaskCountTrue {
			con = golabl.Task.Footer.TaskCountTrue - con
		}
		// 更新 进度
		if updateTaskProgressErr := tool.UpdateTaskProgress(con); updateTaskProgressErr != nil {
			return updateTaskProgressErr
		}

		// 可选：添加延迟，避免请求过快
		// 暂停200豪秒
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("第一阶段 - 总数：%v 当前已取出：%v \n", goodsList.TotalCount, *totalFetched)
	}
	return nil
}

// PhaseTwoGoods 第二阶段拉取商品信息
// @param pageSize 每页数量
// @param totalFetched 获取到的商品总数
// @param lastCreatedAt 最后一条数据的创建时间
// @param maxRecordsPerRange 每次请求的时间范围最多获取的记录数
// @return error 错误信息
func PhaseTwoGoods(pageSize int, totalFetched *int, lastCreatedAt *int64, maxRecordsPerRange int) error {
	// 第二阶段：使用时间范围分批次获取数据，每批最多获取10000条
	// 设置结束时间为开始时间+30天
	endTime := *lastCreatedAt + 30*24*60*60
	fmt.Printf("第二阶段开始，结束时间设置为: %d (%s)\n", endTime, time.Unix(endTime, 0).Format("2006-01-02 15:04:05"))

	if *lastCreatedAt > 0 {
		currentCreatedAtFrom := *lastCreatedAt
		maxLoopCount := 100 // 最大循环次数保护
		loopCount := 0
		var lastPageGoodsList []planBTypePinduoduo.GoodsItem // 记录上一页的商品列表

		for loopCount < maxLoopCount {
			loopCount++

			// 检查开始时间是否已超过当前时间
			if currentCreatedAtFrom > time.Now().Unix() {
				fmt.Printf("开始时间 %d 已超过当前时间，停止获取\n", currentCreatedAtFrom)
				break
			}

			// 每次循环都重新设置结束时间为开始时间+30天
			currentCreatedAtEnd := currentCreatedAtFrom + 30*24*60*60

			fmt.Printf("开始获取时间范围: %d 到 %d\n", currentCreatedAtFrom, currentCreatedAtEnd)

			currentPage := 1
			batchGoodsCount := 0
			lastItemCreatedAt := int64(0)
			hasDataInRange := false
			lastPageGoodsList = nil // 重置上一页商品列表

			// 在当前时间范围内分页获取数据
			for {
				params := map[string]string{
					"accessToken":   golabl.Task.Header.ShopMsg.Token,
					"page":          strconv.Itoa(currentPage),
					"pageSize":      strconv.Itoa(pageSize),
					"createdAtFrom": strconv.FormatInt(currentCreatedAtFrom, 10),
					"createdAtEnd":  strconv.FormatInt(currentCreatedAtEnd, 10),
				}

				goodsList, goodsListStr, getGoodsListErr := tool.GetPddGoodsList(params)
				if getGoodsListErr != nil {
					return fmt.Errorf("获取商品列表失败（时间范围），页码: %d, 错误: %v", currentPage, getGoodsListErr)
				}
				if goodsListStr == "{}" {
					fmt.Println("通过容器获取获取商品列表数据失败，重试一次")
					goodsList, goodsListStr, getGoodsListErr = tool.GetPddGoodsList(params)
					if getGoodsListErr != nil {
						return fmt.Errorf("获取商品列表失败（时间范围），页码: %d, 错误: %v", currentPage, getGoodsListErr)
					}
				}
				if goodsListStr == "{}" {
					return fmt.Errorf("容器返回数据为空")
				}

				//如果是需要拉取详情的商品
				if golabl.Task.Header.TaskType == 4 {
					// 获取原始商品列表
					originalGoodsList := goodsList.GoodsList
					totalCount := len(originalGoodsList)

					if totalCount == 0 {
						return nil // 或继续后续处理
					}

					// 存储所有获取到的商品详情
					allGoodsDetailList := make([]planBTypePinduoduo.GoodsItem, 0, totalCount)

					// 每100条调用一次
					batchSize := 100
					n := 0
					for i := 0; i < totalCount; i += batchSize {
						// 计算当前批次的起始和结束位置
						end := i + batchSize
						if end > totalCount {
							end = totalCount
						}
						batch := originalGoodsList[i:end]
						n++
						// 调用接口获取商品详情
						fmt.Printf(" 第 %v 次 \n", n)
						goodsDetailList, goodsDetailListStr, getPddGoodsDetailErr := tool.GetPddGoodsDetail(batch)
						if getPddGoodsDetailErr != nil {
							fmt.Println("----------------------------错误！！！！goodsDetailList-------------------------------")
							fmt.Printf("batch start %d end %d, batch size %d, total %d\n", i, end, len(batch), totalCount)
							fmt.Println(goodsDetailListStr)
							fmt.Println("----------------------------错误！！！！goodsDetailList-------------------------------")
							return getPddGoodsDetailErr
						}

						// 将当前批次的结果添加到总结果中
						allGoodsDetailList = append(allGoodsDetailList, goodsDetailList...)
					}

					// 赋值回原变量
					goodsList.GoodsList = allGoodsDetailList
				}
				// 如果当前页没有数据
				if len(goodsList.GoodsList) == 0 {
					// 如果当前页是第一页且没有数据
					if currentPage == 1 {
						// 整个时间范围都没有数据，直接推进到结束时间
						currentCreatedAtFrom = currentCreatedAtEnd
						fmt.Printf("时间范围 %d - %d 内无数据，推进开始时间到: %d\n", currentCreatedAtFrom-30*24*60*60, currentCreatedAtEnd, currentCreatedAtFrom)
						break
					}

					// 当前页没有数据，但上一页有数据
					// 取上一页最后一条数据的创建时间和GoodsId作为新的开始位置
					if len(lastPageGoodsList) > 0 {
						lastItemOfLastPage := lastPageGoodsList[len(lastPageGoodsList)-1]
						newStartTime := lastItemOfLastPage.CreatedAt
						lastGoodsId := lastItemOfLastPage.GoodsId

						// 使用基于 GoodsId的定位策略
						if newStartTime > currentCreatedAtFrom {
							currentCreatedAtFrom = newStartTime
							fmt.Printf("当前页无数据，使用上一页最后一条商品时间作为新开始时间: %d, 最后商品ID: %d\n", currentCreatedAtFrom, lastGoodsId)
						} else {
							// 如果时间相同，需要基于GoodsId来推进，这里简单地将时间加1秒
							currentCreatedAtFrom = newStartTime + 1
							fmt.Printf("当前页无数据，时间相同，将时间加1秒推进: %d\n", currentCreatedAtFrom)
						}
					} else {
						// 理论上不会走到这里，但为了安全，将开始时间推进到结束时间
						currentCreatedAtFrom = currentCreatedAtEnd
						fmt.Printf("当前页无数据且无上一页数据，将开始时间推进到结束时间: %d\n", currentCreatedAtFrom)
					}

					hasDataInRange = false
					break
				}

				// 有数据，记录上一页的商品列表
				lastPageGoodsList = goodsList.GoodsList
				hasDataInRange = true

				// 收集商品数据并统计
				for _, goods := range goodsList.GoodsList {
					*totalFetched++
					// 写入到数据库中
					// 将goods转为json
					jsonData, jsonMarshalErr := json.Marshal(goods)
					if jsonMarshalErr != nil {
						return fmt.Errorf("将商品转为json失败: %v\n", jsonMarshalErr)
					}
					// 写入到数据库中
					if len(goods.SkuList) <= 0 {
						return fmt.Errorf("商品sku列表为空 goodsId %v", goods.GoodsId)
					}
					bodyWait := planAType.TaskBody{
						BookInfo: planAType.BookInfo{
							Isbn:            goods.SkuList[0].OuterId,
							BookName:        goods.GoodsName,
							Author:          "",
							Publishing:      "",
							PublicationDate: "",
							Binding:         "",
							PagesCount:      0,
							WordsCount:      0,
							Format:          0,
							Price:           goods.CreatedAt,
						},
						Detail: planAType.TaskDetail{
							Error:   string(jsonData),
							GoodsId: goods.GoodsId,
							Stock:   int64(goods.SkuList[0].ReserveQuantity),
						},
					}
					// 将bodyWait 转为json
					bodyWaitJson, jsonMarshalErr := json.Marshal(bodyWait)
					if jsonMarshalErr != nil {
						return fmt.Errorf("将bodyWait转为json失败: %v\n", jsonMarshalErr)
					}
					//写入 body_wait
					addTaskToBodyWaitErr := service.AddTaskToBodyWait(string(bodyWaitJson))
					if addTaskToBodyWaitErr != nil {
						return addTaskToBodyWaitErr
					}
				}

				batchGoodsCount += len(goodsList.GoodsList)

				// 记录最后一条商品的创建时间
				lastItem := goodsList.GoodsList[len(goodsList.GoodsList)-1]
				lastItemCreatedAt = lastItem.CreatedAt

				fmt.Printf("第二阶段 - 当前时间范围已获取: %d 条，累计总数: %d，当前页码: %d，最后商品时间: %d\n",
					batchGoodsCount, *totalFetched, currentPage, lastItemCreatedAt)

				// 判断是否需要结束当前时间范围
				// 1. 如果当前批次已经达到或超过 maxRecordsPerRange
				// 2. 或者返回的数据少于 pageSize（说明没有下一页了）
				if batchGoodsCount >= maxRecordsPerRange || len(goodsList.GoodsList) < pageSize {
					// 关键修复：使用最后一条商品的时间作为新的开始时间
					// 如果最后一条商品时间等于当前开始时间，则加1秒避免死循环
					if lastItemCreatedAt == currentCreatedAtFrom {
						currentCreatedAtFrom = lastItemCreatedAt + 1
						fmt.Printf("最后商品时间与开始时间相同，推进1秒: %d -> %d\n", lastItemCreatedAt, currentCreatedAtFrom)
					} else {
						currentCreatedAtFrom = lastItemCreatedAt
					}
					fmt.Printf("当前时间范围已获取 %d 条数据，准备进入下一时间范围 更新开始时间为: %d \n", batchGoodsCount, currentCreatedAtFrom)
					break
				}

				currentPage++

				// 获取 footer信息
				if getTaskFooterErr := service.GetTaskFooter(); getTaskFooterErr != nil {
					return getTaskFooterErr
				}
				con := int64(len(goodsList.GoodsList))
				if con >= golabl.Task.Footer.TaskCountTrue {
					con = golabl.Task.Footer.TaskCountTrue - con
				}
				// 更新 进度
				if updateTaskProgressErr := tool.UpdateTaskProgress(con); updateTaskProgressErr != nil {
					return updateTaskProgressErr
				}

				// 暂停200豪秒
				time.Sleep(200 * time.Millisecond)
			}

			// 判断是否需要继续循环
			// 情况1：当前时间范围内没有获取到任何数据
			if !hasDataInRange {
				// 检查新的开始时间是否已超过当前时间
				if currentCreatedAtFrom > time.Now().Unix() {
					fmt.Printf("开始时间 %d 已超过当前时间，停止获取\n", currentCreatedAtFrom)
					break
				}
				// 否则继续下一轮循环
				fmt.Printf("继续下一轮查询，新起始时间: %d\n", currentCreatedAtFrom)
				continue
			}

			// 情况2：当前批次获取的数据少于 maxRecordsPerRange 并且开始时间大于当前时间，说明已经没有更多数据了
			if batchGoodsCount < maxRecordsPerRange && currentCreatedAtFrom > time.Now().Unix() {
				fmt.Printf("当前批次获取 %d 条数据，少于 %d，且开始时间已超过当前时间，已完成所有数据获取\n", batchGoodsCount, maxRecordsPerRange)
				break
			}

			// 暂停200豪秒
			time.Sleep(200 * time.Millisecond)
		}

		if loopCount >= maxLoopCount {
			fmt.Printf("警告：已达到最大循环次数 %d，强制退出\n", maxLoopCount)
		}
	}
	return nil
}

// 拉取任务 读取body_wait去重复后写入到body_over中
// @param duplicateCount int64 重复商品数量
// @param uniqueCount int64 不重复商品数量
// @return error 错误信息
func deduplicateToBodyOver(duplicateCount *int, uniqueCount *int) error {
	page := 1
	pageSize := 1000
	var dataList []planBTypePinduoduo.GoodsItem
	// 在循环外维护一个已处理的商品 ID集合
	processedGoodsIds := make(map[int64]bool)
	//在循环前删除 body_over与body_backup，避免重复写入
	deleteTaskBodyOverErr := service.DeleteTaskBodyOver()
	if deleteTaskBodyOverErr != nil {
		return deleteTaskBodyOverErr
	}
	deleteTaskBodyBackupErr := service.DeleteTaskBodyBackup()
	if deleteTaskBodyBackupErr != nil {
		return deleteTaskBodyBackupErr
	}
	num := 0
	//获取body_wait总数量
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
			// 没有数据，结束循环
			break
		}
		for _, v := range list {
			// 解析v 到结构体
			goods := planAType.TaskBody{}
			jsonUnmarshalErr := json.Unmarshal([]byte(v), &goods)
			if jsonUnmarshalErr != nil {
				return fmt.Errorf("将json转为结构体失败: %v\n", jsonUnmarshalErr)
			}
			if !processedGoodsIds[goods.Detail.GoodsId] {
				//写入到去重复集合
				processedGoodsIds[goods.Detail.GoodsId] = true
				//不重复数据 计次
				*uniqueCount++
				// goods.Detail.Error(原始json到结构体)
				var GoodsItem planBTypePinduoduo.GoodsItem
				jsonUnmarshalErr = json.Unmarshal([]byte(goods.Detail.Error), &GoodsItem)
				if jsonUnmarshalErr != nil {
					return fmt.Errorf("将json转为结构体失败: %v\n", jsonUnmarshalErr)
				}
				//记录到data数组中，之后推送到写入店铺商品数据接口
				dataList = append(dataList, GoodsItem)
				//写入到body_over
				goods.Detail.Status = 1
				addTaskToBodyOverErr := service.AddTaskToBodyOver(goods, []string{"body_over", "body_backup"})
				if addTaskToBodyOverErr != nil {
					return addTaskToBodyOverErr
				}

				//将指定店铺信息记录到本地
				isFileShopId, isShopIDExistsErr := tool.IsShopIDExists(golabl.Task.Header.ShopId)
				if isShopIDExistsErr != nil {
					return isShopIDExistsErr
				}
				if isFileShopId {
					text := goods.BookInfo.Isbn + " " + GoodsItem.BigImg
					txtUrl := golabl.Config.FileUrl.PddGoodsDetailsUrl
					fileName := golabl.Task.Header.TaskId
					// 构建完整的文件路径
					filePath := filepath.Join(txtUrl, fileName+".txt")
					// 写入文件
					if err := tool.AppendTextToFile(filePath, text); err != nil {
						fmt.Println("保存详情信息到文本 失败", err.Error())
					}
				}
			} else {
				//重复数据 计次
				*duplicateCount++
			}
		}
		// 将获取的数据推送写入店铺商品数据接口
		ret, retStr, writePddGoodsDataErr := tool.WritePddGoodsData(dataList, page, pageTotal)
		if writePddGoodsDataErr != nil {
			return writePddGoodsDataErr
		}
		if ret.Code != "200" {
			return fmt.Errorf("添加商品失败 %v", retStr)
		}
		num = num + len(dataList)
		fmt.Printf("开始添加商品信息到系统店铺中 当前页 %v 总页数 %v 当前数据量 %v 总数据量 %v \n", page, pageTotal, len(dataList), num)
		page++

		// 获取 footer信息
		if getTaskFooterErr := service.GetTaskFooter(); getTaskFooterErr != nil {
			return getTaskFooterErr
		}
		con := int64(len(list))
		if con >= golabl.Task.Footer.TaskCountTrue {
			con = golabl.Task.Footer.TaskCountTrue - con
		}
		// 更新 进度
		if updateTaskProgressErr := tool.UpdateTaskProgress(con); updateTaskProgressErr != nil {
			return updateTaskProgressErr
		}

		//清空 dataStr
		dataList = []planBTypePinduoduo.GoodsItem{}
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

// setSaleStatusGoodsTask 设置商品上下架状态
// @param taskMsg 任务内容
// @return string body 信息
func setSaleStatusGoodsTask(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	// 拼多多商品 Id不能为空
	if taskMsg.Detail.GoodsId == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("拼多多商品 Id不能为空"))
	}

	var reqDataInfo planBTypePinduoduo.SetSaleStatusGoodsTaskReq
	reqDataInfo.GoodsId = taskMsg.Detail.GoodsId
	if taskMsg.Detail.Status == 1 {
		reqDataInfo.IsOnsale = 1 //上架
	} else if taskMsg.Detail.Status == 2 {
		reqDataInfo.IsOnsale = 0 //下架
	} else {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("任务类型错误"))
	}
	setSoleStatusGoodsRet, _, err := setSoleStatusGoods(logUuid, reqDataInfo)
	if err != nil {
		return "", err
	}
	if !setSoleStatusGoodsRet.GoodsSaleStatusSetResponse.IsSuccess {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("设置商品上下架状态失败"))
	}
	return tool.ReturnSuccess(taskMsg)
}

// setSoleStatusGoods 商品上下架
// @param logUuid 日志ID
// @param reqDataInfo 请求信息
// @return SetSaleStatusGoodsTaskResponse 结果
// @return string 结果json
// @return error 错误信息
func setSoleStatusGoods(logUuid string, reqDataInfo planBTypePinduoduo.SetSaleStatusGoodsTaskReq) (planBTypePinduoduo.SetSaleStatusGoodsTaskResponse, string, error) {
	var setSoleStatusGoods planBTypePinduoduo.SetSaleStatusGoodsTaskResponse
	goodsInfoStr, jsonMarshalErr := json.Marshal(reqDataInfo)
	if jsonMarshalErr != nil {
		return setSoleStatusGoods, "", jsonMarshalErr
	}
	//发送请求
	goodsSoleStatusStr, pddGoodsSoleStatusErr := golabl.PddDll.PddGoodsSaleStatusSet(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, string(goodsInfoStr))
	//判断是否成功
	if strings.Contains(goodsSoleStatusStr, "请求失败") || strings.Contains(goodsSoleStatusStr, "错误码") {
		//记录请求日志
		reqMsg := fmt.Sprintf(`
════════════════════════════════════════════════════════════════
【拼多多上下架请求】
请求ID: %s
时间: %s
参数: %s
════════════════════════════════════════════════════════════════`,
			logUuid,
			time.Now().Format("2006-01-02 15:04:05.000"),
			string(goodsInfoStr))

		tool.LoggingMiddleware(logs.LOG_LEVEL_INFO, reqMsg)
		return setSoleStatusGoods, goodsSoleStatusStr, errors.New("拼多多 PddGoodsSaleStatusSet 错误:" + goodsSoleStatusStr)
	}
	if pddGoodsSoleStatusErr != nil {
		return setSoleStatusGoods, "", pddGoodsSoleStatusErr
	}
	jsonUnmarshal := json.Unmarshal([]byte(goodsSoleStatusStr), &setSoleStatusGoods)
	if jsonUnmarshal != nil {
		return setSoleStatusGoods, "", fmt.Errorf("解析拼多多 PddGoodsAdd 接口返回json失败: %v", jsonUnmarshal)
	}
	return setSoleStatusGoods, goodsSoleStatusStr, nil
}

// updateGoodsQuantity 修改库存
func updateGoodsQuantity(logUuid string, taskMsg planAType.TaskBody, UpdateType int, stock int64) (string, error) {
	// 拼多多商品 Id不能为空
	if taskMsg.Detail.GoodsId == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("拼多多商品 Id不能为空"))
	}
	// 拼多多商品 SkuId不能为空
	if taskMsg.Detail.SkuId == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("拼多多商品 SkuId不能为空"))
	}
	reqDataInfo := planBTypePinduoduo.UpdateGoodsQuantity{
		ForceUpdate: true,
		GoodsId:     taskMsg.Detail.GoodsId,
		SkuId:       taskMsg.Detail.SkuId,
		Quantity:    taskMsg.Detail.Stock,
		UpdateType:  UpdateType,
	}
	delGoodsRet, _, err := quantityGoods(logUuid, reqDataInfo)
	if err != nil {
		return "", err
	}
	if !delGoodsRet.GoodsQuantityUpdateResponse.IsSuccess {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("修改商品库存失败"))
	}
	if UpdateType == 2 {
		taskMsg.Detail.Stock = taskMsg.Detail.Stock + stock
	}
	taskMsg.Detail.Error = "增加库存成功！"
	return tool.ReturnSuccess(taskMsg)
}

// quantityGoods 修改库存
// @param logUuid 日志ID
// @param reqDataInfo 请求信息
// @return GoodsAddResponseWrapper 结果
// @return string 结果json
// @return error 错误信息
func quantityGoods(logUuid string, reqDataInfo planBTypePinduoduo.UpdateGoodsQuantity) (planBTypePinduoduo.UpdateGoodsQuantityResponse, string, error) {
	var quantityGoods planBTypePinduoduo.UpdateGoodsQuantityResponse
	goodsInfoStr, jsonMarshalErr := json.Marshal(reqDataInfo)
	if jsonMarshalErr != nil {
		return quantityGoods, "", jsonMarshalErr
	}
	//发送请求
	quantityGoodsStr, pddGoodsQuantityErr := golabl.PddDll.PddGoodsQuantityUpdate(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, string(goodsInfoStr))
	//判断是否成功
	if strings.Contains(quantityGoodsStr, "请求失败") || strings.Contains(quantityGoodsStr, "错误码") {
		//记录请求日志
		reqMsg := fmt.Sprintf(`
	════════════════════════════════════════════════════════════════
	【拼多多修改商品库存请求】
	请求ID: %s
	时间: %s
	参数: %s
	════════════════════════════════════════════════════════════════`,
			logUuid,
			time.Now().Format("2006-01-02 15:04:05.000"),
			string(goodsInfoStr))

		tool.LoggingMiddleware(logs.LOG_LEVEL_INFO, reqMsg)
		return quantityGoods, quantityGoodsStr, errors.New("拼多多 quantityGoods 错误:" + quantityGoodsStr)
	}
	if pddGoodsQuantityErr != nil {
		return quantityGoods, "", pddGoodsQuantityErr
	}
	jsonUnmarshal := json.Unmarshal([]byte(quantityGoodsStr), &quantityGoods)
	if jsonUnmarshal != nil {
		return quantityGoods, "", fmt.Errorf("解析拼多多 quantityGoods 接口返回json失败: %v", jsonUnmarshal)
	}
	return quantityGoods, quantityGoodsStr, nil
}

// updateSkuPrice 修改商品价格
func updateSkuPrice(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	// 拼多多商品 Id不能为空
	if taskMsg.Detail.GoodsId == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("拼多多商品 Id不能为空"))
	}
	// 拼多多商品 SkuId不能为空
	if taskMsg.Detail.SkuId == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("拼多多商品 SkuId不能为空"))
	}
	// 价格0 不能发布
	if taskMsg.Detail.Price == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("拼多多商品 价格不能为0"))
	}
	fen := tool.BuildGoodsPrice(taskMsg.Detail.Price)
	yuan := tool.FenToYuan(fen)
	reqDataInfo := planBTypePinduoduo.UpdateSkuPrice{
		GoodsId:           taskMsg.Detail.GoodsId,
		MarketPrice:       fen,
		MarketPriceInYuan: yuan,
		SkuPriceList: []planBTypePinduoduo.SkuPriceItem{
			{
				GroupPrice:  taskMsg.Detail.Price,
				SinglePrice: taskMsg.Detail.Price + 100,
				SkuId:       taskMsg.Detail.SkuId,
			},
		},
	}
	delGoodsRet, _, err := skuPrice(logUuid, reqDataInfo)
	if err != nil {
		return "", err
	}
	if !delGoodsRet.GoodsUpdateSkuPriceResponse.IsSuccess {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("修改商品库存失败"))
	}
	return tool.ReturnSuccess(taskMsg)
}

// skuPrice 修改价格
// @param logUuid 日志ID
// @param reqDataInfo 请求信息
// @return GoodsAddResponseWrapper 结果
// @return string 结果json
// @return error 错误信息
func skuPrice(logUuid string, reqDataInfo planBTypePinduoduo.UpdateSkuPrice) (planBTypePinduoduo.UpdateGoodsSkuPriceResponse, string, error) {
	var skuPriceGoods planBTypePinduoduo.UpdateGoodsSkuPriceResponse
	goodsInfoStr, jsonMarshalErr := json.Marshal(reqDataInfo)
	if jsonMarshalErr != nil {
		return skuPriceGoods, "", jsonMarshalErr
	}
	//发送请求
	skuPriceGoodsStr, pddGoodsSkuPriceErr := golabl.PddDll.PddGoodsSkuPriceUpdate(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, string(goodsInfoStr))
	//判断是否成功
	if strings.Contains(skuPriceGoodsStr, "请求失败") || strings.Contains(skuPriceGoodsStr, "错误码") {
		//记录请求日志
		reqMsg := fmt.Sprintf(`
	════════════════════════════════════════════════════════════════
	【拼多多修改商品价格请求】
	请求ID: %s
	时间: %s
	参数: %s
	════════════════════════════════════════════════════════════════`,
			logUuid,
			time.Now().Format("2006-01-02 15:04:05.000"),
			string(goodsInfoStr))

		tool.LoggingMiddleware(logs.LOG_LEVEL_INFO, reqMsg)
		return skuPriceGoods, skuPriceGoodsStr, errors.New("拼多多 skuPrice 错误:" + skuPriceGoodsStr)
	}
	if pddGoodsSkuPriceErr != nil {
		return skuPriceGoods, "", pddGoodsSkuPriceErr
	}
	jsonUnmarshal := json.Unmarshal([]byte(skuPriceGoodsStr), &skuPriceGoods)
	if jsonUnmarshal != nil {
		return skuPriceGoods, "", fmt.Errorf("解析拼多多 skuPrice 接口返回json失败: %v", jsonUnmarshal)
	}
	return skuPriceGoods, skuPriceGoodsStr, nil
}

// 拼多多发布
func publishGoods(logUuid string, taskMsg planAType.TaskBody) (planAType.TaskBody, error) {
	// 价格不能小于0
	if taskMsg.Detail.Price <= 0 {
		return taskMsg, fmt.Errorf("价格不能小于等于0")
	}

	//获取出版社信息并解析1
	if getPublishingErr := service.GetPublishingVid(&taskMsg); getPublishingErr != nil {
		return taskMsg, fmt.Errorf("获取出版社信息失败-原因来自:%v", getPublishingErr)
	}

	//违规词处理1
	if golabl.Config.Server.Filter == 1 {
		//开启违规词处理
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

	var goodsAdd planBTypePinduoduo.GoodsAdd

	// *********************构建参数 开始******************************** //

	//构建商品名称
	goodsAdd.GoodsName = tool.BuildGoodsName(
		golabl.Task.Header.ShopMsg.GoodsNamePrefix, // 商品名称前缀
		golabl.Task.Header.ShopMsg.GoodsNameSuffix, // 商品名称后缀
		golabl.Task.Header.ShopMsg.TitleConsistOf,  // 标题组成
		golabl.Task.Header.ShopMsg.SpaceCharacter,  // 间隔符
		taskMsg.BookInfo) // 图书信息
	taskMsg.Detail.GoodsName = goodsAdd.GoodsName

	// 构建轮播图
	//if taskHeader.ShopMsg.WatermarkImgUrl == "" && len(taskHeader.ShopMsg.CarouseLastImgUrlArray) == 0 && len(taskMsg.BookInfo.ImageObject.CarouselUrlArray) == 0 && taskMsg.BookInfo.ImageObject.DefaultImageUrl == "" {
	//	return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd,fmt.Errorf("缺少构造轮播图图片-未提交 isbn %v", taskMsg.BookInfo.Isbn))
	//}
	if len(taskMsg.BookInfo.ImageObject.CarouselUrlArray) == 0 {
		// 无图片信息 isbn计次
		setNoImgCountErr := service.SetNoImgCount(taskMsg.BookInfo.Isbn)
		if setNoImgCountErr != nil {
			return taskMsg, fmt.Errorf("无图片信息isbn计次错误 isbn %v %v", taskMsg.BookInfo.Isbn, setNoImgCountErr.Error())
		}
		return taskMsg, fmt.Errorf("缺少轮播图")
	}

	//判断是否拼多多的图片，如果非拼多多图片则上传到拼多多的图片空间
	for k, v := range taskMsg.BookInfo.ImageObject.CarouselUrlArray {
		// 判断v 是否包含字符串 img.pddpic.com
		if !strings.Contains(v, "img.pddpic.com") {
			tempUrl, saveBase64ImageByDateErr := tool.SaveBase64ImageByDate(v, golabl.Config.FileUrl.PddImgTempUrl)
			if saveBase64ImageByDateErr != nil {
				return taskMsg, saveBase64ImageByDateErr
			}
			//转为base64
			imgBase64, format, processImageErr := tool.ProcessImage(tempUrl)
			if processImageErr != nil {
				return taskMsg, processImageErr
			}

			// 将base64字符串包装成 ImageResult 类型
			imageResult := []planBTypeModules.ImageResult{
				{
					Success: true,      // 假设图片处理成功
					Format:  format,    // 或者根据实际情况获取图片格式
					Data:    imgBase64, // base64数据
				},
			}

			//上传到拼多多图片空间
			toPdd, uploadImageToPddErr := tool.UploadImageToPdd(imageResult)
			if uploadImageToPddErr != nil {
				return taskMsg, fmt.Errorf("上传图片到拼多多失败 %v", uploadImageToPddErr)
			}
			taskMsg.BookInfo.ImageObject.CarouselUrlArray[k] = toPdd[0]
		}
	}

	oldCarouselUrlArray := append([]string{}, taskMsg.BookInfo.ImageObject.CarouselUrlArray...) //原始轮播图，用于后续处理，不会被打上水印

	// 存在水印图片，则打水印
	if golabl.Task.Header.ShopMsg.WatermarkImgUrl != "" {
		//获取水印图片
		watermarkImgUrl, watermarkImgErr := tool.GetWatermarkImg()
		if watermarkImgErr != nil {
			return taskMsg, fmt.Errorf("获取水印图片失败 %v", watermarkImgErr)
		}
		//打水印
		watermarkFromURLExsBase64Arr, watermarkFromURLExsErr := tool.AddWatermarkFromURLExs(taskMsg.BookInfo.ImageObject.CarouselUrlArray, watermarkImgUrl, golabl.Task.Header.ShopMsg.WatermarkPosition)
		if watermarkFromURLExsErr != nil {
			return taskMsg, fmt.Errorf("图片打水印失败 %v", watermarkFromURLExsErr)
		}
		//图片上传到拼多多
		toPdd, uploadImageToPddErr := tool.UploadImageToPdd(watermarkFromURLExsBase64Arr)
		if uploadImageToPddErr != nil {
			return taskMsg, fmt.Errorf("图片上传到拼多多失败 %v", uploadImageToPddErr)
		}
		//将上传的图片替换到商品轮播图中
		for i := 0; i < len(toPdd); i++ {
			taskMsg.BookInfo.ImageObject.CarouselUrlArray[i] = toPdd[i]
		}
	}

	goodsAdd.CarouselGallery = tool.BuildCarouselGallery(golabl.Task.Header.ShopMsg.CarouseLastImgUrlArray, oldCarouselUrlArray, taskMsg.BookInfo.ImageObject.CarouselUrlArray, golabl.Task.Header.ShopMsg.WatermarkPosition)

	if len(taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl) == 0 && len(oldCarouselUrlArray) > 0 {
		taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl = []string{oldCarouselUrlArray[0]}
	}
	if len(goodsAdd.CarouselGallery) == 0 {
		return taskMsg, fmt.Errorf("缺少构造轮播图图片-未提交 isbn %v", taskMsg.BookInfo.Isbn)
	}
	// 构建详情图
	goodsAdd.DetailGallery = tool.BuildDetailGallery(golabl.Task.Header.ShopMsg.GoodsDetailFirstImgUrlArray, golabl.Task.Header.ShopMsg.GoodsDetailLastImgUrlArray, taskMsg.BookInfo.ImageObject.DetailUrlObject, oldCarouselUrlArray[0])
	if len(goodsAdd.DetailGallery) == 0 {
		return taskMsg, fmt.Errorf("缺少构造详情图-未提交 isbn %v", taskMsg.BookInfo.Isbn)
	}

	// 构建 catId
	var catID int64

	if taskMsg.BookInfo.CatIdObject.PinDuoDuoCatId == "" {
		// 调用拼多多 SDK 取类目信息
		pddCalbackStr, pddGoodsOuterCatMappingGetErr := golabl.PddDll.PddGoodsOuterCatMappingGet(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, "15543", "书籍/杂志/报纸", "书籍 "+taskMsg.BookInfo.BookName)
		if pddGoodsOuterCatMappingGetErr != nil {
			return taskMsg, fmt.Errorf("调用DLL类目映射失败 %w", pddGoodsOuterCatMappingGetErr)
		}

		// 解析返回的 JSON 字符串
		var response planBTypePinduoduo.PddSuccessResponse
		if unmarshalErr := json.Unmarshal([]byte(pddCalbackStr), &response); unmarshalErr != nil {
			return taskMsg, fmt.Errorf("json.Unmarshal错误 %w %v", unmarshalErr, pddCalbackStr)
		}

		// 判断 catID4 是否为0
		if response.OuterCatMappingGetResponse.CatID4 != 0 {
			catID = response.OuterCatMappingGetResponse.CatID4
		} else {
			catID = response.OuterCatMappingGetResponse.CatID3
		}
	} else {
		// 数据库原本存储的为字符串 转成int64再使用
		retCatID, toInt64Err := taskMsg.BookInfo.CatIdObject.PinDuoDuoCatId.ToInt64()
		if toInt64Err != nil {
			return taskMsg, fmt.Errorf("转换catId错误 %w", toInt64Err)
		}
		catID = retCatID
	}

	// 设置 catId
	goodsAdd.CatId = catID

	// 构建商品类型
	goodsAdd.GoodsType = 1

	// 构建参考价格
	goodsAdd.MarketPrice = tool.BuildGoodsPrice(price)

	// 构建商品编码
	if taskMsg.Detail.OutGoodsId != "" {
		goodsAdd.OutGoodsId = taskMsg.Detail.OutGoodsId
	} else {
		goodsAdd.OutGoodsId = taskMsg.BookInfo.Isbn
	}

	// 是否支持假一赔十
	goodsAdd.IsFolt = golabl.Task.Header.ShopMsg.IsFolt

	// 是否预售
	goodsAdd.IsPreSale = golabl.Task.Header.ShopMsg.IsPreSale

	// 是否支持7天无理由退换货
	goodsAdd.IsRefundable = golabl.Task.Header.ShopMsg.IsRefundable

	// 构建是否是二手商品
	goodsAdd.SecondHand = golabl.Task.Header.ShopMsg.IsSecondHand

	// 构建物流运费模板 ID
	goodsAdd.CostTemplateId = golabl.Task.Header.ShopMsg.CostTemplateId

	// 构建承诺发货时间
	goodsAdd.ShipmentLimitSecond = 48 * 60 * 60

	//满两件折扣
	goodsAdd.TwoPiecesDiscount = golabl.Task.Header.ShopMsg.TwoDiscount

	//构建
	taskMsgBookInfoPrice := taskMsg.BookInfo.Price
	if taskMsgBookInfoPrice < 10000 {
		taskMsgBookInfoPrice = 10000
	}

	goodsAdd.GoodsProperties = buildGoodsPropertiesList(
		taskMsg.BookInfo.Isbn,            // ISBN
		goodsAdd.GoodsName,               // 商品名称
		taskMsg.BookInfo.PagesCount,      // 页数
		taskMsgBookInfoPrice,             // 价格
		taskMsg.Publishing.Vid,           // 出版社 Vid
		taskMsg.BookInfo.Author,          // 作者
		taskMsg.BookInfo.Format,          // 开本
		taskMsg.BookInfo.Binding,         // 装帧
		taskMsg.BookInfo.WordsCount,      // 字数
		taskMsg.BookInfo.PublicationDate, // 出版时间
	)

	//库存
	if taskMsg.Detail.Stock == 0 && (golabl.Task.Header.TaskType == 1 || golabl.Task.Header.TaskType == 2 || golabl.Task.Header.TaskType == 6) {
		//如果库存为0 则给默认库存
		taskMsg.Detail.Stock = golabl.Task.Header.ShopMsg.DefStock
	} else {
		if taskMsg.Detail.Stock == 0 && golabl.Task.Header.TaskType == 8 {
			return taskMsg, fmt.Errorf("库存不能为0")
		}
	}

	//生成一个2秒的延迟
	url := "http://127.0.0.1:8095"
	tool.HttpGetRequest(url)

	// 规格编号
	if taskMsg.Detail.SkuCode == "" {
		taskMsg.Detail.SkuCode = goodsAdd.OutGoodsId
	}

	//构建 sku信息
	skuThumbnail := oldCarouselUrlArray[0]
	if golabl.Task.Header.ShopMsg.SkuWatermarkImgUrl != "" {
		//获取水印图片
		skuWatermarkImgUrl, skuWatermarkImgErr := tool.GetSkuWatermarkImg()
		if skuWatermarkImgErr != nil {
			return taskMsg, fmt.Errorf("获取水印图片失败 %v", skuWatermarkImgErr)
		}
		//sku 打水印
		skuThumbnailArr := []string{skuThumbnail}
		skuWatermarkFromURLExsBase64Arr, skuWatermarkFromURLExsErr := tool.AddWatermarkFromURLExs(skuThumbnailArr, skuWatermarkImgUrl, "1")
		if skuWatermarkFromURLExsErr != nil {
			return taskMsg, fmt.Errorf("sku图片打水印失败 %v", skuWatermarkFromURLExsErr)
		}
		//sku 图片上传到拼多多
		skuToPdd, uploadImageToPddErr := tool.UploadImageToPdd(skuWatermarkFromURLExsBase64Arr)
		if uploadImageToPddErr != nil {
			return taskMsg, fmt.Errorf("图片上传到拼多多失败 %v", uploadImageToPddErr)
		}
		//将上传的图片替换到商品轮播图中
		skuThumbnail = skuToPdd[0]
	}

	//构建规格名称
	var specChildName string
	if golabl.Task.Header.ShopMsg.SpecCompose == "0" {
		specChildName = golabl.Task.Header.ShopMsg.SpecChildName
	} else if golabl.Task.Header.ShopMsg.SpecCompose == "1" {
		// 前缀+isbn+后缀
		specChildName = golabl.Task.Header.ShopMsg.SpecPrefix + taskMsg.BookInfo.Isbn + golabl.Task.Header.ShopMsg.SpecSuffix
	} else if golabl.Task.Header.ShopMsg.SpecCompose == "2" {
		// 前缀+书名+后缀
		specChildName = golabl.Task.Header.ShopMsg.SpecPrefix + taskMsg.BookInfo.BookName + golabl.Task.Header.ShopMsg.SpecSuffix
	} else if golabl.Task.Header.ShopMsg.SpecCompose == "3" {
		// 前缀+货号+后缀
		specChildName = golabl.Task.Header.ShopMsg.SpecPrefix + strconv.FormatInt(taskMsg.Detail.SkuId, 10) + golabl.Task.Header.ShopMsg.SpecSuffix
	}
	// 如果规格名称超过30个字符，截取前30个字符
	if tool.StringLength(specChildName) > 30 {
		specChildName = tool.SubstringByWidth(specChildName, 30)
	}

	// 根据配置重新构建skuCode
	if golabl.Task.Header.ShopMsg.SpecCodeCompose == "1" {
		taskMsg.Detail.SkuCode = taskMsg.BookInfo.Isbn
	}

	//构建 sku信息
	sku, err := buildSkuList(price, skuThumbnail, taskMsg.Detail.Stock, taskMsg.Detail.SkuCode, specChildName, taskMsg.Detail.IsOnsale)
	if err != nil {
		return taskMsg, err
	}
	goodsAdd.SkuList = []planBTypePinduoduo.Sku{sku}

	// *********************构建参数 结束******************************** //

	// 发送请求
	goodsAddRet, _, err := addGoods(logUuid, goodsAdd)
	if err != nil {
		return taskMsg, fmt.Errorf("商品提交 %v", err)
	}

	// 获取商品提交的商品详情
	goodsCommitDetail, _, getGoodsCommitDetailErr := getGoodsCommitDetail(goodsAddRet.Response.GoodsCommitID, goodsAddRet.Response.GoodsID)
	if getGoodsCommitDetailErr != nil {
		return taskMsg, fmt.Errorf("获取商品提交的商品详情失败 %w", getGoodsCommitDetailErr)
	}

	//拼接接口调用成功的返回数据
	if len(goodsCommitDetail.GoodsCommitDetailResponse.SkuList) > 0 {
		//taskMsg.Detail.SkuCode = goodsCommitDetail.GoodsCommitDetailResponse.SkuList[0].OutSkuSn
		taskMsg.Detail.SkuId = goodsCommitDetail.GoodsCommitDetailResponse.SkuList[0].SkuID
	}
	taskMsg.Detail.GoodsId = goodsAddRet.Response.GoodsID
	taskMsg.Detail.ReturnId = goodsAddRet.Response.GoodsCommitID
	taskMsg.Detail.OutGoodsId = goodsAdd.OutGoodsId
	taskMsg.Detail.Img = goodsAdd.CarouselGallery[0]
	return taskMsg, nil
}
