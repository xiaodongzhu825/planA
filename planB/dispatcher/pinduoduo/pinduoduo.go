package pinduoduo

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/planB/initialization/golabl"
	"planA/planB/modules/image"
	"planA/planB/modules/logs"
	"planA/planB/modules/pdd"
	"planA/planB/service"
	"planA/planB/tool"
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
// @param taskHeader 任务头
// @param taskMsg 任务内容
// @return string 预留
// @return error 错误
func (pinDuoDuo *PinDuoDuo) AddGoodsTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}

	// 价格不能小于0
	if taskMsg.Detail.Price <= 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("价格不能小于等于0"))
	}

	//获取出版社信息并解析
	if getPublishingErr := service.GetPublishingVid(&taskMsg); getPublishingErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("获取出版社信息失败-原因来自:%v", getPublishingErr))
	}

	//违规词处理
	if golabl.Config.Server.Filter == 1 {
		//开启违规词处理
		if taskMsgErr := tool.FilterWord(&taskMsg); taskMsgErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, taskMsgErr)
		}
	}

	// 价格出来
	price := tool.BuildPrice(golabl.Task.Header.PriceMod, taskMsg.Detail.Price)
	if price == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("不在价格区间内 isbn %v  原始价格 %v  当前价格 %v 价格模版 %v", taskMsg.BookInfo.Isbn, taskMsg.Detail.Price, price, golabl.Task.Header.PriceMod))
	}
	taskMsg.Detail.Price = price

	// 初始化 PddDll
	var goodsAdd planBTypePinduoduo.GoodsAdd
	pddDll, err := pdd.InitPddDll()
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("初始化拼多多DLL失败 %v", err))
	}

	// 初始化 imageDll
	imageDll, imageDllErr := image.InitImageDll()
	if imageDllErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("初始化图片DLL失败 %v", imageDllErr))
	}

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
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("无图片信息isbn计次错误 isbn %v %v", taskMsg.BookInfo.Isbn, setNoImgCountErr.Error()))
		}
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("缺少官图"))
	}
	oldCarouselUrlArray := append([]string{}, taskMsg.BookInfo.ImageObject.CarouselUrlArray...) //原始轮播图，用于后续处理，不会被打上水印

	// 存在水印图片，则打水印
	if golabl.Task.Header.ShopMsg.WatermarkImgUrl != "" {
		//打水印
		watermarkFromURLExsBase64Arr, watermarkFromURLExsErr := tool.AddWatermarkFromURLExs(imageDll, taskMsg.BookInfo.ImageObject.CarouselUrlArray, golabl.Task.Header.ShopMsg.WatermarkImgUrl, golabl.Task.Header.ShopMsg.WatermarkPosition)
		if watermarkFromURLExsErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("图片打水印失败 %v", watermarkFromURLExsErr))
		}
		//图片上传到拼多多
		toPdd, uploadImageToPddErr := tool.UploadImageToPdd(pddDll, watermarkFromURLExsBase64Arr)
		if uploadImageToPddErr != nil {
			return "", fmt.Errorf("图片上传到拼多多失败 %v", uploadImageToPddErr)
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
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("缺少构造轮播图图片-未提交 isbn %v", taskMsg.BookInfo.Isbn))
	}

	// 构建详情图
	goodsAdd.DetailGallery = tool.BuildDetailGallery(golabl.Task.Header.ShopMsg.GoodsDetailFirstImgUrlArray, golabl.Task.Header.ShopMsg.GoodsDetailLastImgUrlArray, taskMsg.BookInfo.ImageObject.DetailUrlObject)

	// 构建 catId
	var catID int64

	if taskMsg.BookInfo.CatIdObject.PinDuoDuoCatId == "" {
		// 调用拼多多 SDK 取类目信息
		pddCalbackStr, pddGoodsOuterCatMappingGetErr := pddDll.PddGoodsOuterCatMappingGet(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, "15543", "书籍/杂志/报纸", "书籍 "+taskMsg.BookInfo.BookName)
		if pddGoodsOuterCatMappingGetErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("调用DLL类目映射失败 %w", err))
		}

		// 解析返回的 JSON 字符串
		var response planBTypePinduoduo.PddSuccessResponse
		if unmarshalErr := json.Unmarshal([]byte(pddCalbackStr), &response); unmarshalErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("json.Unmarshal错误 %w %v", unmarshalErr, pddCalbackStr))
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
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("转换catId错误 %w", toInt64Err))
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
		taskMsg.BookInfo.Isbn,       // ISBN
		goodsAdd.GoodsName,          // 商品名称
		taskMsg.BookInfo.PagesCount, // 页数
		taskMsgBookInfoPrice,        // 价格
		taskMsg.Publishing.Vid,      // 出版社 Vid
		taskMsg.BookInfo.Author,     // 作者
		taskMsg.BookInfo.Format,     // 开本
		taskMsg.BookInfo.Binding,    // 装帧
	)

	//库存
	if taskMsg.Detail.Stock == 0 && golabl.Task.Header.TaskType == 1 {
		//如果库存为0 则给默认库存
		taskMsg.Detail.Stock = golabl.Task.Header.ShopMsg.DefStock
	}

	//生成一个2秒的延迟
	url := "http://127.0.0.1:8095"
	tool.HttpGetRequest(url)

	// 规格编号
	if taskMsg.Detail.SkuCode == "" {
		taskMsg.Detail.SkuCode = goodsAdd.OutGoodsId
	}

	//获取 sku
	sku, err := buildSkuList(pddDll, price, goodsAdd.CarouselGallery[0], taskMsg.Detail.Stock, taskMsg.Detail.SkuCode)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, err)
	}
	goodsAdd.SkuList = []planBTypePinduoduo.Sku{sku}

	// *********************构建参数 结束******************************** //

	// 发送请求
	goodsAddRet, _, err := addGoods(pddDll, logUuid, goodsAdd)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("商品提交 %v", err))
	}

	// 获取商品提交的商品详情
	goodsCommitDetail, _, getGoodsCommitDetailErr := getGoodsCommitDetail(pddDll, goodsAddRet.Response.GoodsCommitID, goodsAddRet.Response.GoodsID)
	if getGoodsCommitDetailErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("获取商品提交的商品详情失败 %w", getGoodsCommitDetailErr))
	}

	//拼接接口调用成功的返回数据
	if len(goodsCommitDetail.GoodsCommitDetailResponse.SkuList) > 0 {
		taskMsg.Detail.SkuCode = goodsCommitDetail.GoodsCommitDetailResponse.SkuList[0].OutSkuSn
		taskMsg.Detail.SkuId = goodsCommitDetail.GoodsCommitDetailResponse.SkuList[0].SkuID
	}
	taskMsg.Detail.GoodsId = goodsAddRet.Response.GoodsID
	taskMsg.Detail.ReturnId = goodsAddRet.Response.GoodsCommitID
	taskMsg.Detail.OutGoodsId = goodsAdd.OutGoodsId
	taskMsg.Detail.Img = goodsAdd.CarouselGallery[0]
	taskMsg.Detail.SkuCode = goodsAdd.OutGoodsId

	return tool.GoodsAddReturnSuccess(taskMsg)
}
func (pinDuoDuo *PinDuoDuo) SetGoodsTask() string {
	return ""
}

// GetGoodsTask 获取商品
// @return string 预留
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

	var allGoodsList []planBTypePinduoduo.GoodsItem
	var lastCreatedAt int64 = 0

	// 统计变量
	totalFetched := 0   // 总共获取到的商品数（包括重复）
	duplicateCount := 0 // 重复商品数量
	uniqueCount := 0    // 不重复商品数量

	// 在循环外维护一个已处理的商品 ID集合
	processedGoodsIds := make(map[int64]bool)

	// 查询body_wait中最后一条商品的创建时间

	// 第一阶段：获取第1页到第100页，不传入时间参数
	for page := 1; page <= maxPage; page++ {
		// 定义参数
		params := map[string]string{
			"accessToken": golabl.Task.Header.ShopMsg.Token,
			"page":        strconv.Itoa(page),
			"pageSize":    strconv.Itoa(pageSize),
		}

		goodsList, getGoodsListErr := tool.GetGoodsList(params)
		if getGoodsListErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType,
				fmt.Errorf("获取商品列表失败，页码: %d, 错误: %v", page, getGoodsListErr))
		}

		// 收集商品数据并统计
		for _, goods := range goodsList.GoodsList {
			totalFetched++
			if !processedGoodsIds[goods.GoodsId] {
				processedGoodsIds[goods.GoodsId] = true
				allGoodsList = append(allGoodsList, goods)
				uniqueCount++
			} else {
				duplicateCount++
			}
		}

		// 记录最后一页的最后一条数据的创建时间
		if page == maxPage && len(goodsList.GoodsList) > 0 {
			lastCreatedAt = goodsList.GoodsList[len(goodsList.GoodsList)-1].CreatedAt
		}

		// 如果没有更多数据，提前退出
		if len(goodsList.GoodsList) == 0 {
			break
		}

		// 可选：添加延迟，避免请求过快
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("第一阶段 - 总数：%v 当前已取出：%v \n", goodsList.TotalCount, len(allGoodsList))
	}

	// 第二阶段：使用时间范围分批次获取数据，每批最多获取10000条
	// 设置结束时间为当前时间+24小时
	endTime := time.Now().Add(24 * time.Hour).Unix()
	fmt.Printf("第二阶段开始，结束时间设置为: %d (%s)\n", endTime, time.Unix(endTime, 0).Format("2006-01-02 15:04:05"))

	if lastCreatedAt > 0 {
		currentCreatedAtFrom := lastCreatedAt
		maxLoopCount := 100 // 最大循环次数保护
		loopCount := 0
		var lastPageGoodsList []planBTypePinduoduo.GoodsItem // 记录上一页的商品列表

		for loopCount < maxLoopCount {
			loopCount++

			// 每次循环都重新设置结束时间为当前时间+24小时
			currentCreatedAtEnd := time.Now().Add(24 * time.Hour).Unix()

			// 检查起始时间是否已超过结束时间
			if currentCreatedAtFrom >= currentCreatedAtEnd {
				fmt.Printf("起始时间 %d 已大于等于结束时间 %d，停止获取\n", currentCreatedAtFrom, currentCreatedAtEnd)
				break
			}

			fmt.Printf("开始获取时间范围: %d 到 %d\n", currentCreatedAtFrom, currentCreatedAtEnd)

			currentPage := 1
			batchGoodsCount := 0
			lastItemCreatedAt := int64(0)
			lastItemGoodsId := int64(0) // 记录最后一条商品的 GoodsId
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

				goodsList, getGoodsListErr := tool.GetGoodsList(params)
				if getGoodsListErr != nil {
					return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType,
						fmt.Errorf("获取商品列表失败（时间范围），页码: %d, 错误: %v", currentPage, getGoodsListErr))
				}

				// 如果当前页没有数据
				if len(goodsList.GoodsList) == 0 {
					// 如果当前页是第一页且没有数据
					if currentPage == 1 {
						fmt.Printf("时间范围 %d - %d 内无数据\n", currentCreatedAtFrom, currentCreatedAtEnd)
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
							// 记录这个开始时间对应的最后一个GoodsId，用于下一轮去重
							fmt.Printf("当前页无数据，使用上一页最后一条商品时间作为新开始时间: %d, 最后商品ID: %d\n", currentCreatedAtFrom, lastGoodsId)
						} else {
							// 如果时间相同，需要基于GoodsId来推进
							fmt.Printf("当前页无数据，但时间相同，需要基于GoodsId推进\n")
							// 设置一个标志，表示下一轮需要基于GoodsId定位
							// 这里不直接修改currentCreatedAtFrom，而是在下一轮开始时通过查询来定位
						}
					} else {
						// 理论上不会走到这里，但为了安全，将开始时间推进到结束时间
						currentCreatedAtFrom = currentCreatedAtEnd
						fmt.Printf("当前页无数据且无上一页数据，将开始时间推进到结束时间: %d\n", currentCreatedAtFrom)
					}

					// 标记无数据，跳出当前时间范围的循环
					hasDataInRange = false
					break
				}

				// 有数据，记录上一页的商品列表
				lastPageGoodsList = goodsList.GoodsList
				hasDataInRange = true

				// 收集商品数据并统计
				for _, goods := range goodsList.GoodsList {
					totalFetched++
					if !processedGoodsIds[goods.GoodsId] {
						processedGoodsIds[goods.GoodsId] = true
						allGoodsList = append(allGoodsList, goods)
						uniqueCount++
					} else {
						duplicateCount++
					}
				}

				batchGoodsCount += len(goodsList.GoodsList)

				// 记录最后一条商品的创建时间和 GoodsId
				lastItemCreatedAt = goodsList.GoodsList[len(goodsList.GoodsList)-1].CreatedAt
				lastItemGoodsId = goodsList.GoodsList[len(goodsList.GoodsList)-1].GoodsId

				fmt.Printf("第二阶段 - 当前时间范围已获取: %d 条，累计总数: %d，当前页码: %d\n",
					batchGoodsCount, len(allGoodsList), currentPage)

				// 判断是否需要结束当前时间范围
				// 1. 如果当前批次已经达到或超过 maxRecordsPerRange
				// 2. 或者返回的数据少于 pageSize（说明没有下一页了）
				if batchGoodsCount >= maxRecordsPerRange || len(goodsList.GoodsList) < pageSize {
					fmt.Printf("当前时间范围已获取 %d 条数据，准备进入下一时间范围\n", batchGoodsCount)

					// 如果是因为数据量达到上限而结束，更新开始位置
					if batchGoodsCount >= maxRecordsPerRange && lastItemCreatedAt > 0 {
						if lastItemCreatedAt > currentCreatedAtFrom {
							currentCreatedAtFrom = lastItemCreatedAt
							fmt.Printf("达到批次上限，更新开始时间为: %d, 最后商品ID: %d\n", currentCreatedAtFrom, lastItemGoodsId)
						} else {
							// 时间相同，需要通过查询来定位下一个商品
							fmt.Printf("达到批次上限但时间相同，需要通过查询定位下一个商品\n")
						}
					}
					break
				}

				currentPage++

				// 可选：添加延迟，避免请求过快
				time.Sleep(100 * time.Millisecond)
			}

			// 判断是否需要继续循环
			// 情况1：当前时间范围内没有获取到任何数据
			if !hasDataInRange {
				// 如果起始时间已经 >= 结束时间，退出循环
				if currentCreatedAtFrom >= currentCreatedAtEnd {
					fmt.Printf("起始时间 %d 已大于等于结束时间 %d，且无数据，停止获取\n", currentCreatedAtFrom, currentCreatedAtEnd)
					break
				}
				// 否则继续下一轮循环
				fmt.Printf("继续下一轮查询，新起始时间: %d\n", currentCreatedAtFrom)
				continue
			}

			// 情况2：当前批次获取的数据少于 maxRecordsPerRange，说明已经没有更多数据了
			if batchGoodsCount < maxRecordsPerRange {
				fmt.Printf("当前批次获取 %d 条数据，少于 %d，已完成所有数据获取\n", batchGoodsCount, maxRecordsPerRange)
				break
			}

			// 可选：在批次之间添加延迟，避免请求过快
			time.Sleep(200 * time.Millisecond)
		}

		if loopCount >= maxLoopCount {
			fmt.Printf("警告：已达到最大循环次数 %d，强制退出\n", maxLoopCount)
		}
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
最终保存商品数: %d
════════════════════════════════════════════════════════════════`,
		logUuid,
		time.Now().Format("2006-01-02 15:04:05.000"),
		golabl.Task.TaskId,
		golabl.Task.Header.ShopName,
		totalFetched,
		uniqueCount,
		duplicateCount,
		float64(duplicateCount)/float64(totalFetched)*100,
		len(allGoodsList))
	fmt.Println(statsLogMsg)
	logs.LoggingMiddleware(logs.LOG_LEVEL_INFO, statsLogMsg)

	//延迟3分钟,并且循环打印每秒倒计时
	totalSeconds := 180 // 3分钟 = 180秒
	for i := totalSeconds; i >= 0; i-- {
		minutes := i / 60
		seconds := i % 60
		fmt.Printf("\r剩余时间: %02d:%02d", minutes, seconds)
		if i > 0 {
			time.Sleep(1 * time.Second)
		}
	}

	return tool.GoodsAddReturnSuccess(planAType.TaskBody{})
}

func (pinDuoDuo *PinDuoDuo) DelGoodsTask() string {
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
// @return []GoodsProperties 商品属性列表
func buildGoodsPropertiesList(isbn, bookName string, pageCount, price int64, publishingVid int64, author string, format int64, binding string) []planBTypePinduoduo.GoodsProperties {
	var goodsPropertiesArr []planBTypePinduoduo.GoodsProperties
	//isbn
	goodsPropertiesIsbn := planBTypePinduoduo.GoodsProperties{
		RefPid: 425,
		Value:  isbn,
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesIsbn)

	//书名
	goodsPropertiesBookName := planBTypePinduoduo.GoodsProperties{
		RefPid: 876,
		Value:  bookName,
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesBookName)

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
		Value:     strconv.FormatInt(price/10000, 10),
		ValueUnit: "元",
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesPrice)

	//出版社
	goodsPropertiesPublishing := planBTypePinduoduo.GoodsProperties{
		RefPid: 880,
		Vid:    publishingVid,
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesPublishing)

	//作者
	goodsPropertiesAuthor := planBTypePinduoduo.GoodsProperties{
		RefPid: 882,
		Value:  author,
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesAuthor)

	//开本
	goodsPropertiesFormat := planBTypePinduoduo.GoodsProperties{
		RefPid: 890,
		Value:  strconv.FormatInt(format, 10),
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesFormat)

	//装帧
	goodsPropertiesBinding := planBTypePinduoduo.GoodsProperties{
		RefPid: 891,
		Value:  binding,
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesBinding)
	return goodsPropertiesArr
}

// sku规格生成
// @param pddDll pddDLL对象
// @param price 价格
// @param thumbUrl 缩略图
// @param stock 库存
// @param outSkuSn 商品编码
// @return Sku sku规格
// @return error 错误信息
func buildSkuList(pddDll *pdd.PddDLL, price int64, thumbUrl string, stock int64, outSkuSn string) (planBTypePinduoduo.Sku, error) {
	//构建变量
	specId := golabl.Task.Header.ShopMsg.SpecId
	specName := golabl.Task.Header.ShopMsg.SpecName
	specChildName := golabl.Task.Header.ShopMsg.SpecChildName
	// 构建 Spec列表
	var sku planBTypePinduoduo.Sku
	goodsSpec, buildPddGoodsSpecIdErr := buildPddGoodsSpecId(pddDll, specId, specChildName)
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
	sku = planBTypePinduoduo.Sku{
		IsOnsale:      1,             //上架状态，0-已下架，1-上架中
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
// @param pddDll pddDLL对象
// @param specId 商品规格id
// @param specName 规格名称
// @return DllGoodsSpec 规格信息
// @return error 错误信息
func buildPddGoodsSpecId(pddDll *pdd.PddDLL, id int64, name string) (planAType.DllGoodsSpec, error) {
	var spec planAType.DllGoodsSpec
	specStr, err := pddDll.PddGoodsSpecIdGet(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, strconv.FormatInt(id, 10), name)
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
// @param pddDll pddDLL对象
// @param logUuid 日志ID
// @param goodsInfo 商品信息
// @return GoodsAddResponseWrapper 商品新增结果
// @return string 商品新增结果json
// @return error 错误信息
func addGoods(pddDll *pdd.PddDLL, logUuid string, goodsInfo planBTypePinduoduo.GoodsAdd) (planBTypePinduoduo.GoodsAddResponseWrapper, string, error) {
	var goodsAdd planBTypePinduoduo.GoodsAddResponseWrapper
	goodsInfoStr, jsonMarshalErr := json.Marshal(goodsInfo)
	if jsonMarshalErr != nil {
		return goodsAdd, "", jsonMarshalErr
	}
	//发送请求
	goodsAddStr, pddGoodsAddErr := pddDll.PddGoodsAdd(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, string(goodsInfoStr))
	//判断是否成功
	if strings.Contains(goodsAddStr, "请求失败") || strings.Contains(goodsAddStr, "错误码") {
		//记录请求日志
		// 记录请求日志
		addGoodsReqMsg := fmt.Sprintf(`
════════════════════════════════════════════════════════════════
【拼多多商品添加请求】
请求ID: %s
店铺ID：%v
店铺名称：%v
时间: %s
参数: %s
════════════════════════════════════════════════════════════════`,
			logUuid,
			time.Now().Format("2006-01-02 15:04:05.000"),
			golabl.Task.TaskId,
			golabl.Task.Header.ShopName,
			string(goodsInfoStr))

		logs.LoggingMiddleware(logs.LOG_LEVEL_INFO, addGoodsReqMsg)
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
// @param pddDll pddDLL对象
// @param goodsCommitId 商品提交ID
// @param goodsId 商品ID
// @return GoodsCommitDetailResponse 商品提交详情
// @return error 错误信息
func getGoodsCommitDetail(pddDll *pdd.PddDLL, goodsCommitId int64, goodsId int64) (planBTypePinduoduo.GoodsCommitDetailResponse, string, error) {
	var goodsCommitDetail planBTypePinduoduo.GoodsCommitDetailResponse
	goodsCommitDetailStr, pddGoodsCommitDetailGetErr := pddDll.PddGoodsCommitDetailGet(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, golabl.Task.Header.ShopMsg.Token, strconv.FormatInt(goodsCommitId, 10), strconv.FormatInt(goodsId, 10))
	if pddGoodsCommitDetailGetErr != nil {
		return goodsCommitDetail, "", pddGoodsCommitDetailGetErr
	}
	unmarshalErr := json.Unmarshal([]byte(goodsCommitDetailStr), &goodsCommitDetail)
	if unmarshalErr != nil {
		return goodsCommitDetail, "", fmt.Errorf("解析拼多多 PddGoodsCommitDetailGet 接口返回json失败: %v [拼多多数据：%v]", unmarshalErr, goodsCommitDetailStr)
	}
	return goodsCommitDetail, goodsCommitDetailStr, nil
}
