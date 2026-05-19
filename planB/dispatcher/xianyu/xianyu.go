package xianyu

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/planB/initialization/golabl"
	"planA/planB/modules/logs"
	"planA/planB/service"
	"planA/planB/tool"
	planBType "planA/planB/type"
	planBTypeXianyu "planA/planB/type/xianyu"
	planAType "planA/type"
	"strconv"
	"time"
)

type XianYu struct {
}

// NewXianYu 创建闲鱼平台
func NewXianYu() *XianYu {
	return &XianYu{}
}

// AddGoodsTask 添加商品
// @param taskMsg 任务内容
// @return string body 信息
// @return error 错误
func (xianYu *XianYu) AddGoodsTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}

	// 初始化 imageDll
	//imageDll, imageDllErr := image.InitImageDll()
	//if imageDllErr != nil {
	//	return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("初始化图片DLL失败 %v", imageDllErr))
	//}

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

	// 构建参数
	var goodsAdd planBTypeXianyu.GoodsAdd

	// 解析应用 id与应用秘钥
	var token planBTypeXianyu.Token
	unmarshalErr := json.Unmarshal([]byte(golabl.Task.Header.ShopMsg.Token), &token)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("解析应用id与应用秘钥 taskHeader.ShopMsg.Token = %v %w", golabl.Task.Header.ShopMsg.Token, unmarshalErr))
	}

	// 应用 ID
	goodsAdd.AppId = token.AppId

	//  应用密钥
	goodsAdd.AppSecret = token.AppSecret

	//  token
	goodsAdd.Token = ""

	// API 使用的店铺ID
	goodsAdd.ApiShopId = 0

	// 平台类型
	goodsAdd.TypePlatform = 4

	// 店铺 ID
	goodsAdd.ShopId = 0

	// 店铺 Token
	goodsAdd.ShopToken = ""

	// 店铺名称
	goodsAdd.ShopName = ""

	// 发货省，格式为省级行政区划代码（如210000代表辽宁省）
	provinceCode, cityCode, districtCode, getProvinceCityDistrictErr := getProvinceCityDistrict(0, 20)
	if getProvinceCityDistrictErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("获取省、市、区信息失败: %v", getProvinceCityDistrictErr))
	}
	goodsAdd.Province = provinceCode

	// 发货市，格式为市级行政区划代码（如210100代表沈阳市）
	goodsAdd.City = cityCode

	// 发货区，格式为区级行政区划代码（如210101代表和平区）
	goodsAdd.District = districtCode

	// 商品类型
	goodsAdd.TypeGoods = ""

	// 分类类型
	goodsAdd.TypeClass = ""

	// 类目 ID
	spBizType := 24                                      //默认 图书类目
	goodsAdd.CatIds = "c3c6e8d1d63c0618b108d382c4e6ea42" //默认类目ID（文学，小说）
	if golabl.Task.Header.ShopMsg.PublishType == "1" {
		spBizType = 99                                          //其他
		goodsAdd.CatIds = golabl.Task.Header.ShopMsg.CategoryId //根据用户选择
	}

	//if goodsAdd.CatIds == "" {
	//	//如果类目ID为空，则使用默认类目ID（文学，小说）
	//	goodsAdd.CatIds = "c3c6e8d1d63c0618b108d382c4e6ea42"
	//	spBizType = 99 //（其他）
	//}

	if len(taskMsg.BookInfo.ImageObject.CarouselUrlArray) == 0 {
		// 无图片信息 isbn计次
		setNoImgCountErr := service.SetNoImgCount(taskMsg.BookInfo.Isbn)
		if setNoImgCountErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("无图片信息isbn计次错误 isbn %v %v", taskMsg.BookInfo.Isbn, setNoImgCountErr.Error()))
		}
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("缺少官图"))
	}
	// 构建详情图
	contentImgs := tool.BuildDetailGallery(golabl.Task.Header.ShopMsg.GoodsDetailFirstImgUrlArray, golabl.Task.Header.ShopMsg.GoodsDetailLastImgUrlArray, taskMsg.BookInfo.ImageObject.DetailUrlObject, taskMsg.BookInfo.ImageObject.CarouselUrlArray[0])

	//oldCarouselUrlArray := append([]string{}, taskMsg.BookInfo.ImageObject.CarouselUrlArray...) //原始轮播图，用于后续处理，不会被打上水印

	// 存在水印图片，则打水印
	//if golabl.Task.Header.ShopMsg.WatermarkImgUrl != "" {
	//	//打水印
	//	watermarkFromURLExsBase64Arr, watermarkFromURLExsErr := tool.AddWatermarkFromURLExs(imageDll, taskMsg.BookInfo.ImageObject.CarouselUrlArray, golabl.Task.Header.ShopMsg.WatermarkImgUrl, golabl.Task.Header.ShopMsg.WatermarkPosition)
	//	if watermarkFromURLExsErr != nil {
	//		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("图片打水印失败 %v", watermarkFromURLExsErr))
	//	}
	//
	//	//图片上传到图片空间
	//	toMinIo, uploadToMinIoErr := tool.UploadToMinIo(watermarkFromURLExsBase64Arr)
	//	if uploadToMinIoErr != nil {
	//		return "", uploadToMinIoErr
	//	}
	//
	//	//将上传的图片替换到商品轮播图中
	//	for i := 0; i < len(toMinIo); i++ {
	//		taskMsg.BookInfo.ImageObject.CarouselUrlArray[i] = toMinIo[i]
	//	}
	//}
	// 构建主图（轮播图）
	refactorCarouselGallery := tool.BuildCarouselGalleryOld(golabl.Task.Header.ShopMsg.CarouseLastImgUrlArray, taskMsg.BookInfo.ImageObject.CarouselUrlArray)
	//refactorCarouselGallery := tool.BuildCarouselGallery(golabl.Task.Header.ShopMsg.CarouseLastImgUrlArray, oldCarouselUrlArray, taskMsg.BookInfo.ImageObject.CarouselUrlArray, "1")

	// 如果轮播图没有图片，并且是优先官图，则使用默认图片
	if len(refactorCarouselGallery) == 0 && golabl.Task.Header.ImgType == 3 && taskMsg.BookInfo.ImageObject.DefaultImageUrl != "" {
		refactorCarouselGallery = append(refactorCarouselGallery, taskMsg.BookInfo.ImageObject.DefaultImageUrl)
	}

	if len(taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl) == 0 && len(refactorCarouselGallery) > 0 {
		taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl = []string{refactorCarouselGallery[0]}
	}
	if len(refactorCarouselGallery) == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("缺少构造轮播图图片-未提交 isbn %v", taskMsg.BookInfo.Isbn))
	}
	//构建商品名称
	title := tool.BuildGoodsName(
		golabl.Task.Header.ShopMsg.GoodsNamePrefix, // 商品名称前缀
		golabl.Task.Header.ShopMsg.GoodsNameSuffix, // 商品名称后缀
		golabl.Task.Header.ShopMsg.TitleConsistOf,  // 标题组成
		golabl.Task.Header.ShopMsg.SpaceCharacter,  // 间隔符
		taskMsg.BookInfo) // 图书信息
	taskMsg.Detail.GoodsName = title

	// 构建商品信息
	content := taskMsg.BookInfo.BookName + " " + taskMsg.BookInfo.Isbn + " " + taskMsg.BookInfo.Author + " " + taskMsg.BookInfo.Publishing
	content = content + "\n" + golabl.Task.Header.ShopMsg.ShopContext

	// 店铺信息
	goodsAdd.Shop = []planBTypeXianyu.ShopInfo{
		{
			UserName:    token.Username,
			Province:    provinceCode,
			City:        cityCode,
			District:    districtCode,
			Title:       title,
			Content:     content,
			MainImgs:    refactorCarouselGallery,
			ContentImgs: contentImgs,
		},
	}

	// 成色
	goodsAdd.StuffStatus = taskMsg.Detail.Condition
	if goodsAdd.StuffStatus == 0 {
		goodsAdd.StuffStatus = 90
	}

	//库存
	if taskMsg.Detail.Stock == 0 && (golabl.Task.Header.TaskType == 1 || golabl.Task.Header.TaskType == 2 || golabl.Task.Header.TaskType == 6) {
		//如果库存为0 则给默认库存
		taskMsg.Detail.Stock = golabl.Task.Header.ShopMsg.DefStock
	}

	url := "http://127.0.0.1:8095"
	tool.HttpGetRequest(url)

	//构建参考价格
	price := tool.BuildPrice(golabl.Task.Header.PriceMod, taskMsg.Detail.Price)
	if price == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("不在价格区间内 isbn:%v", taskMsg.BookInfo.Isbn))
	}

	//价格 + 运费
	price = price + taskMsg.Detail.ShippingCost

	taskMsg.Detail.Price = price

	//构建定价
	taskMsgBookInfoPrice := tool.BuildGoodsPrice(price)

	// 图书类商品信息
	goodsAdd.BookData = []planBTypeXianyu.BookInfo{
		{
			ISBN:        taskMsg.BookInfo.Isbn,
			Title:       title,
			Author:      taskMsg.BookInfo.Author,
			Publisher:   taskMsg.Publishing.Value,
			ItemBizType: 2,
			SpBizType:   spBizType,
			Prices:      []int64{price, taskMsgBookInfoPrice},
			Stock:       taskMsg.Detail.Stock,
			CatIds:      string(taskMsg.BookInfo.CatIdObject.XianYuCatId),
		},
	}

	// 闲鱼批次商品 KEY
	goodsAdd.ItemKey = strconv.FormatInt(time.Now().Unix(), 10)
	// 新增商品
	goodsAddRet, goodsAddStr, err := addGoods(logUuid, goodsAdd)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("商品提交 %v", err))
	}

	if len(goodsAddRet.Data.Success) <= 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("新增商品失败 %v 闲鱼返回信息 %v", err, goodsAddStr))
	}
	// 上架商品
	launchGoodsInfo := planBTypeXianyu.Product{
		AppId:              token.AppId,
		AppSecret:          token.AppSecret,
		Token:              "",
		NotifyURL:          "",
		ProductID:          goodsAddRet.Data.Success[0].ProductID,
		SpecifyPublishTime: "",
		UserName:           []string{token.Username},
	}
	//延迟1分钟
	time.Sleep(time.Minute)
	//商品上架
	if taskMsg.Detail.IsOnsale == 0 {
		_, _, err = launchGoods(logUuid, launchGoodsInfo)
		if err != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("商品提交 %v", err))
		}
	}
	// 构建商品编码
	outGoodsId := ""
	if taskMsg.Detail.OutGoodsId != "" {
		outGoodsId = taskMsg.Detail.OutGoodsId
	} else {
		outGoodsId = taskMsg.BookInfo.Isbn
	}
	taskMsg.Detail.GoodsId = goodsAddRet.Data.Success[0].ProductID
	taskMsg.Detail.OutGoodsId = outGoodsId
	taskMsg.Detail.Img = refactorCarouselGallery[0]

	return tool.ReturnSuccess(taskMsg)
}

func (xianYu *XianYu) GetGoodsTask() (string, error) {
	// 生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}

	const pageSize = 100
	const maxPage = 100
	const maxRecordsPerRange = 10000 // 每个时间范围最多获取10000条
	var lastUpdateTime int64 = 0

	// 统计变量
	totalFetched := 0   // 总共获取到的商品数（包括重复）
	duplicateCount := 0 // 重复商品数量
	uniqueCount := 0    // 不重复商品数量

	// 解析应用 id与应用秘钥
	var token planBTypeXianyu.Token
	unmarshalErr := json.Unmarshal([]byte(golabl.Task.Header.ShopMsg.Token), &token)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, fmt.Errorf("解析应用id与应用秘钥失败: %v", unmarshalErr))
	}

	// 第一阶段：只拉取任务数据，更新总数，不写入wait
	firstTimeGoodsErr := xianYu.phaseOneGoodsOnlyCount(token, pageSize, maxPage)
	if firstTimeGoodsErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, firstTimeGoodsErr)
	}

	// 查询body_wait是否存在，确定第二阶段的开始时间
	exist, isTaskBodyWaitExistErr := service.IsTaskBodyWaitExist()
	if isTaskBodyWaitExistErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, isTaskBodyWaitExistErr)
	}

	if exist {
		// 获取最后一条数据的更新时间作为开始时间
		lastBodyWaitDataJson, getLastGoodsUpdateTimeErr := service.GetTaskBodyWaitLast()
		if getLastGoodsUpdateTimeErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, getLastGoodsUpdateTimeErr)
		}
		// 解析 lastBodyWaitData 到结构体
		var lastBodyWaitData planAType.TaskBody
		unmarshalErr := json.Unmarshal([]byte(lastBodyWaitDataJson), &lastBodyWaitData)
		if unmarshalErr != nil {
			return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, unmarshalErr)
		}
		lastUpdateTime = lastBodyWaitData.BookInfo.Price - (86400 * 30) //wait中最后一条数据的更新时间-30天作为下次的开始时间
		// 将数据的更新时间给到 lastUpdateTime
		fmt.Println("使用wait中最后一条数据的时间作为开始时间: ", lastUpdateTime)
	} else {
		// 如果没有wait数据，使用当前时间180天前的时间戳作为开始时间
		lastUpdateTime = time.Now().Unix() - 180*24*60*60
		fmt.Println("没有wait数据，使用180天前的时间作为开始时间: ", lastUpdateTime, time.Unix(lastUpdateTime, 0).Format("2006-01-02 15:04:05"))
	}

	// 第二阶段：获取商品（写入wait）
	phaseTwoGoodsErr := xianYu.phaseTwoGoods(token, pageSize, &totalFetched, &lastUpdateTime, maxRecordsPerRange)
	if phaseTwoGoodsErr != nil {
		fmt.Println(phaseTwoGoodsErr.Error())
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
	deduplicateToBodyOverErr := xianYu.deduplicateToBodyOver(&duplicateCount, &uniqueCount)
	if deduplicateToBodyOverErr != nil {
		return tool.ReturnErr(logUuid, planAType.TaskBody{}, golabl.TaskType, deduplicateToBodyOverErr)
	}

	// 输出统计信息
	statsLogMsg := fmt.Sprintf(`
	════════════════════════════════════════════════════════════════
	【闲鱼店铺拉取】
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
// @return string error 错误
func (xianYu *XianYu) OperationGoodsTask(taskMsg planAType.TaskBody) (string, error) {
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
// @return string error 错误
func (xianYu *XianYu) IncStockTask(taskMsg planAType.TaskBody) (string, error) {
	return tool.ReturnSuccess(planAType.TaskBody{})
}

func (xianYu *XianYu) SetGoodsTask() string {
	return "闲鱼商品修改任务"

}

// *******************************私有方法************************************ //

// 获取省市区 信息
func getProvinceCityDistrict(types int64, id int) (int, int, int, error) {
	if types == 0 { // 直接指定区域的省市区
		//根据区id 获取省、市、区code
		provinceCode, cityCode, districtCode, getRegionIdErr := service.GetRegionId(strconv.Itoa(id))
		if getRegionIdErr != nil {
			return 0, 0, 0, getRegionIdErr
		}
		return provinceCode, cityCode, districtCode, nil
	} else if types == 1 { // 返回指定省下的随机区
		region, getRandomDistrictInProvinceErr := service.GetRandomDistrictInProvince(id)
		if getRandomDistrictInProvinceErr != nil {
			return 0, 0, 0, getRandomDistrictInProvinceErr
		}
		//根据区id 获取省、市、区code
		provinceCode, cityCode, districtCode, getRegionIdErr := service.GetRegionId(region["id"])
		if getRegionIdErr != nil {
			return 0, 0, 0, getRegionIdErr
		}
		return provinceCode, cityCode, districtCode, nil

	} else if types == 2 { //在全国返回随机省下的随机区
		region, getRandomDistrictErr := service.GetRandomDistrict()
		if getRandomDistrictErr != nil {
			return 0, 0, 0, getRandomDistrictErr
		}
		//根据区id 获取省、市、区code
		provinceCode, cityCode, districtCode, getRegionIdErr := service.GetRegionId(region["id"])
		if getRegionIdErr != nil {
			return 0, 0, 0, getRegionIdErr
		}
		return provinceCode, cityCode, districtCode, nil
	}
	return 0, 0, 0, fmt.Errorf("参数错误")
}

// 商品新增
// @param token 授权令牌
// @param logUuid 日志ID
// @param goodsInfo 添加商品信息
// @return XianYuAddGoodsResponse 商品新增结果
// @return string 添加商品结果json
// @return error 错误信息
func addGoods(logUuid string, goodsInfo planBTypeXianyu.GoodsAdd) (planBTypeXianyu.XianYuAddGoodsResponse, string, error) {
	var goodsAdd planBTypeXianyu.XianYuAddGoodsResponse
	goodsInfoStr, marshalErr := json.Marshal(goodsInfo)
	if marshalErr != nil {
		return goodsAdd, "", marshalErr
	}
	goodsAddStr, xianYuGoodsAddErr := golabl.XianYuDll.XianYuGoodsAdd(string(goodsInfoStr), golabl.Config.FileUrl.XianYuDll)
	if xianYuGoodsAddErr != nil {
		return goodsAdd, "", xianYuGoodsAddErr
	}
	unmarshalErr := json.Unmarshal([]byte(goodsAddStr), &goodsAdd)
	if unmarshalErr != nil {
		return goodsAdd, "", unmarshalErr
	}
	if goodsAdd.Code != 0 || goodsAdd.Msg != "OK" {
		//记录请求日志
		addGoodsReqMsg := fmt.Sprintf(`
════════════════════════════════════════════════════════════════
【闲鱼商品添加请求】
请求ID: %s
时间: %s
参数: %s
════════════════════════════════════════════════════════════════`,
			logUuid,
			time.Now().Format("2006-01-02 15:04:05.000"),
			string(goodsInfoStr))
		tool.LoggingMiddleware(logs.LOG_LEVEL_INFO, addGoodsReqMsg)
		return goodsAdd, goodsAddStr, errors.New("闲鱼 XianYuGoodsAdd 错误:" + goodsAddStr)
	}
	return goodsAdd, goodsAddStr, nil
}

// 商品上架
func launchGoods(logUuid string, launchGoodsInfo planBTypeXianyu.Product) (planBTypeXianyu.XianYuAddGoodsResponse, string, error) {
	var launchGoods planBTypeXianyu.XianYuAddGoodsResponse
	launchGoodsInfoStr, marshalErr := json.Marshal(launchGoodsInfo)
	if marshalErr != nil {
		return launchGoods, "", marshalErr
	}
	launchGoodsStr, xianYuLaunchGoodsAddErr := golabl.XianYuDll.XianYuLaunchGoods(string(launchGoodsInfoStr), golabl.Config.FileUrl.XianYuDll)
	if xianYuLaunchGoodsAddErr != nil {
		return launchGoods, "", xianYuLaunchGoodsAddErr
	}
	unmarshalErr := json.Unmarshal([]byte(launchGoodsStr), &launchGoods)
	if unmarshalErr != nil {
		return launchGoods, "", unmarshalErr
	}
	if launchGoods.Code != 0 {
		//记录请求日志
		addGoodsReqMsg := fmt.Sprintf(`
	════════════════════════════════════════════════════════════════
	【闲鱼上架商品请求】
	请求ID: %s
	时间: %s
	参数: %s
	════════════════════════════════════════════════════════════════`,
			logUuid,
			time.Now().Format("2006-01-02 15:04:05.000"),
			string(launchGoodsInfoStr))
		tool.LoggingMiddleware(logs.LOG_LEVEL_INFO, addGoodsReqMsg)
		return launchGoods, launchGoodsStr, errors.New("闲鱼 XianYuLaunchGoods 错误:" + launchGoodsStr)
	}
	return launchGoods, launchGoodsStr, nil
}

// phaseOneGoodsOnlyCount 第一阶段只获取商品总数，不写入wait队列
func (xianYu *XianYu) phaseOneGoodsOnlyCount(token planBTypeXianyu.Token, pageSize int, maxPage int) error {
	for page := 1; page <= maxPage; page++ {
		xianYuListReq := planBTypeXianyu.GoodsListReq{
			AppId:         token.AppId,
			AppSecret:     token.AppSecret,
			UpdateTime:    nil, // 不传入时间，获取所有商品
			ProductStatus: 22,
			PageNo:        page,
			PageSize:      pageSize,
		}

		listJson, err := xianYu.sendGoodsListRequest(xianYuListReq)
		if err != nil {
			return fmt.Errorf("获取商品列表失败，页码: %d, 错误: %v", page, err)
		}

		var list planBTypeXianyu.GoodsListRet
		err = json.Unmarshal([]byte(listJson), &list)
		if err != nil {
			return fmt.Errorf("解析响应失败，页码: %d, 错误: %v", page, err)
		}

		if list.Code != 0 {
			return fmt.Errorf("获取商品列表失败 code=%d, msg=%s", list.Code, list.Msg)
		}

		// 更新header进度总数（第一页）
		if page == 1 {
			fmt.Println("总数: ", strconv.Itoa(list.Data.Count))
			if updateTaskHeaderErr := service.SetTaskCount(strconv.Itoa(list.Data.Count)); updateTaskHeaderErr != nil {
				return updateTaskHeaderErr
			}
			// 获取到总数后即可退出，不需要继续拉取
			fmt.Println("第一阶段完成，已获取商品总数，不写入wait队列")
			break
		}
	}
	return nil
}

// phaseTwoGoods 第二阶段拉取商品信息（按时间范围分批）
// 修改：结束时间固定为当前时间，开始时间为当前时间往前推30天
func (xianYu *XianYu) phaseTwoGoods(token planBTypeXianyu.Token, pageSize int, totalFetched *int, lastUpdateTime *int64, maxRecordsPerRange int) error {
	// 第二阶段：以当前时间为结束时间，开始时间为当前时间往前推30天
	// 注意：lastUpdateTime 参数在此版本中不再使用，改为固定从"当前时间-30天"开始
	now := time.Now().Unix()
	endTime := now                             // 结束时间固定为当前时间
	currentUpdateTimeFrom := now - 30*24*60*60 // 开始时间 = 当前时间 - 30天

	fmt.Printf("第二阶段开始，开始时间: %d (%s), 结束时间: %d (%s) [当前时间]",
		currentUpdateTimeFrom, time.Unix(currentUpdateTimeFrom, 0).Format("2006-01-02 15:04:05"),
		endTime, time.Unix(endTime, 0).Format("2006-01-02 15:04:05"))

	if currentUpdateTimeFrom > 0 {
		currentUpdateTimeFrom := *lastUpdateTime
		maxLoopCount := 100 // 最大循环次数保护
		loopCount := 0

		var currentUpdateTimeEnd int64 // 声明在循环外部，避免每次迭代重置

		for loopCount < maxLoopCount {
			loopCount++

			// 检查开始时间是否已超过当前时间
			if currentUpdateTimeFrom > time.Now().Unix() {
				fmt.Printf("开始时间 %d 已超过当前时间，停止获取\n", currentUpdateTimeFrom)
				break
			}

			// 设置结束时间：首次循环用当前时间，后续循环由pageExceededThreshold或>=10000条件块设置
			if loopCount == 1 {
				currentUpdateTimeEnd = endTime // 首次循环赋值
			}

			// 检查结束时间是否已超过半年（180天），超过则停止获取
			halfYearAgo := time.Now().Unix() - 180*24*60*60
			fmt.Printf("[调试] loopCount=%d, currentUpdateTimeEnd=%d (%s), halfYearAgo=%d (%s), <halfYear=%v\n",
				loopCount, currentUpdateTimeEnd, time.Unix(currentUpdateTimeEnd, 0).Format("2006-01-02"),
				halfYearAgo, time.Unix(halfYearAgo, 0).Format("2006-01-02"),
				currentUpdateTimeEnd < halfYearAgo)
			if currentUpdateTimeEnd < halfYearAgo {
				fmt.Printf("结束时间 %d (%s) 已小于半年前时间 %d (%s)，停止获取\n",
					currentUpdateTimeEnd, time.Unix(currentUpdateTimeEnd, 0).Format("2006-01-02 15:04:05"),
					halfYearAgo, time.Unix(halfYearAgo, 0).Format("2006-01-02 15:04:05"))
				break
			}

			if loopCount >= maxLoopCount {
				fmt.Printf("达到最大循环次数 %d，强制退出\n", maxLoopCount)
				break
			}

			fmt.Printf("开始获取时间范围: %d (%s) 到 %d (%s)\n",
				currentUpdateTimeFrom, time.Unix(currentUpdateTimeFrom, 0).Format("2006-01-02 15:04:05"),
				currentUpdateTimeEnd, time.Unix(currentUpdateTimeEnd, 0).Format("2006-01-02 15:04:05"))

			currentPage := 1
			batchGoodsCount := 0
			lastItemUpdateTime := int64(0)
			hasDataInRange := false
			pageExceededThreshold := false // 标记是否页码超过阈值

			// 在当前时间范围内分页获取数据
			for {
				// 检查当前页码是否超过100
				if currentPage > 100 {
					fmt.Printf("警告：当前页码 %d 已超过100，将使用上一次的结束时间作为新的开始时间\n", currentPage)
					pageExceededThreshold = true
					break
				}

				xianYuListReq := planBTypeXianyu.GoodsListReq{
					AppId:         token.AppId,
					AppSecret:     token.AppSecret,
					UpdateTime:    []int64{currentUpdateTimeFrom, currentUpdateTimeEnd},
					ProductStatus: 22,
					PageNo:        currentPage,
					PageSize:      pageSize,
				}

				listJson, err := xianYu.sendGoodsListRequest(xianYuListReq)
				if err != nil {
					return fmt.Errorf("获取商品列表失败（时间范围），页码: %d, 错误: %v", currentPage, err)
				}

				var list planBTypeXianyu.GoodsListRet
				err = json.Unmarshal([]byte(listJson), &list)
				if err != nil {
					return fmt.Errorf("解析响应失败（时间范围），页码: %d, 错误: %v", currentPage, err)
				}

				if list.Code != 0 {
					return fmt.Errorf("获取商品列表失败 code=%d, msg=%s", list.Code, list.Msg)
				}

				// 如果当前页没有数据
				if len(list.Data.List) == 0 {
					// 如果当前页是第一页且没有数据，说明整个时间范围都没有数据
					if currentPage == 1 {
						fmt.Printf("时间范围 %d - %d 内无数据\n", currentUpdateTimeFrom, currentUpdateTimeEnd)
						hasDataInRange = false
						break
					}
					// 当前页没有数据，但前面有数据，说明当前时间范围的数据已取完
					fmt.Printf("当前时间范围数据已取完，共获取 %d 条数据\n", batchGoodsCount)
					hasDataInRange = true
					break
				}

				// 有数据
				hasDataInRange = true

				// 收集商品数据并统计
				for _, goods := range list.Data.List {
					*totalFetched++
					// 获取商品详情并写入数据库
					err = xianYu.processGoodsDetail(goods, token)
					if err != nil {
						fmt.Printf("处理商品 %s 失败: %v\n", goods.ProductID, err)
						continue
					}
					//拉取后暂停0.01秒
					time.Sleep(10 * time.Millisecond)
				}

				batchGoodsCount += len(list.Data.List)

				// 记录最后一条商品的更新时间
				lastItem := list.Data.List[len(list.Data.List)-1]
				lastItemUpdateTime = lastItem.UpdateTime

				fmt.Printf("第二阶段 - 当前时间范围已获取: %d 条，累计总数: %d，当前页码: %d，最后商品时间: %d (%s)\n",
					batchGoodsCount, *totalFetched, currentPage,
					lastItemUpdateTime, time.Unix(lastItemUpdateTime, 0).Format("2006-01-02 15:04:05"))

				// 更新进度
				if getTaskFooterErr := service.GetTaskFooter(); getTaskFooterErr != nil {
					return getTaskFooterErr
				}
				processed := int64(len(list.Data.List))
				if updateTaskProgressErr := tool.UpdateTaskProgress(processed); updateTaskProgressErr != nil {
					return updateTaskProgressErr
				}

				// 判断是否继续下一页
				// 如果返回的数据少于 pageSize，说明没有下一页了
				if len(list.Data.List) < pageSize {
					fmt.Printf("当前页数据不足 %d 条，当前时间范围数据已取完\n", pageSize)
					break
				}

				currentPage++
			}

			// 处理页码超过阈值的情况
			if pageExceededThreshold {
				// 页码超过100时，使用本次最后一条商品的更新时间作为下一时间阶段的结束时间
				currentUpdateTimeEnd = lastItemUpdateTime
				// 开始时间 = 当前时间往前推30天
				currentUpdateTimeFrom = time.Now().Unix() - 30*24*60*60
				if currentUpdateTimeFrom > currentUpdateTimeEnd {
					// 如果开始时间超过结束时间，说明时间范围无效，直接使用最后商品时间+1秒
					currentUpdateTimeFrom = lastItemUpdateTime + 1
				}
				fmt.Printf("页码超过100，使用最后商品时间 %d (%s) 作为下一时间阶段结束时间，开始时间: %d (%s) [当前时间-30天]\n",
					currentUpdateTimeEnd, time.Unix(currentUpdateTimeEnd, 0).Format("2006-01-02 15:04:05"),
					currentUpdateTimeFrom, time.Unix(currentUpdateTimeFrom, 0).Format("2006-01-02 15:04:05"))
				// 继续下一次循环
				continue
			}

			// 新增逻辑：当获取数量大于等于10000条时，使用最后一条商品的更新时间作为下一时间阶段的结束时间，开始时间则从当前时间往前推30天
			const maxRecordsThreshold = 10000 // 定义阈值常量
			if hasDataInRange && batchGoodsCount >= maxRecordsThreshold {
				// 大于等于10000条：使用最后一条商品的更新时间作为下一时间阶段的结束时间
				currentUpdateTimeEnd = lastItemUpdateTime
				// 开始时间 = 当前时间往前推30天
				currentUpdateTimeFrom = time.Now().Unix() - 30*24*60*60
				if currentUpdateTimeFrom > currentUpdateTimeEnd {
					// 如果开始时间超过结束时间，说明时间范围无效，直接使用最后商品时间+1秒
					currentUpdateTimeFrom = lastItemUpdateTime + 1
				}
				fmt.Printf("获取到 %d 条数据（达到阈值 %d），下一时间阶段结束时间: %d (%s)，开始时间: %d (%s) [当前时间-30天]\n",
					batchGoodsCount, maxRecordsThreshold,
					currentUpdateTimeEnd, time.Unix(currentUpdateTimeEnd, 0).Format("2006-01-02 15:04:05"),
					currentUpdateTimeFrom, time.Unix(currentUpdateTimeFrom, 0).Format("2006-01-02 15:04:05"))
			} else if !hasDataInRange {
				// 情况1：当前时间范围完全没有数据
				// 时间窗口往前移动30天：结束时间 = 开始时间，开始时间 = 开始时间 - 30天
				currentUpdateTimeEnd = currentUpdateTimeFrom
				currentUpdateTimeFrom = currentUpdateTimeFrom - 30*24*60*60
				fmt.Printf("时间范围内无数据，时间窗口往前移动30天\n")
				fmt.Printf("下一时间范围: 开始时间 %d (%s), 结束时间 %d (%s)\n",
					currentUpdateTimeFrom, time.Unix(currentUpdateTimeFrom, 0).Format("2006-01-02 15:04:05"),
					currentUpdateTimeEnd, time.Unix(currentUpdateTimeEnd, 0).Format("2006-01-02 15:04:05"))
			} else if batchGoodsCount < maxRecordsPerRange {
				// 情况2：当前时间范围有数据，但数量少于100条
				// 上一次的开始时间作为下一时间阶段的结束时间，结束时间-30天作为开始时间
				currentUpdateTimeEnd = currentUpdateTimeFrom
				currentUpdateTimeFrom = currentUpdateTimeEnd - 30*24*60*60
				fmt.Printf("当前时间范围获取数据 %d 条，少于 %d 条，上一次开始时间作为下一阶段结束时间，结束时间-30天作为开始时间\n",
					batchGoodsCount, maxRecordsPerRange)
				fmt.Printf("下一时间范围: 开始时间 %d (%s), 结束时间 %d (%s)\n",
					currentUpdateTimeFrom, time.Unix(currentUpdateTimeFrom, 0).Format("2006-01-02 15:04:05"),
					currentUpdateTimeEnd, time.Unix(currentUpdateTimeEnd, 0).Format("2006-01-02 15:04:05"))
			} else {
				// 情况3：当前时间范围获取了100条或以上（但不足10000条），可能还有更多数据
				// 使用最后一条商品的更新时间作为新的开始时间，继续在当前时间范围附近查询
				if lastItemUpdateTime == currentUpdateTimeFrom {
					currentUpdateTimeFrom = lastItemUpdateTime
					fmt.Printf("最后商品时间与开始时间相同，推进1秒: %d -> %d\n", lastItemUpdateTime, currentUpdateTimeFrom)
				} else {
					currentUpdateTimeFrom = lastItemUpdateTime
				}
				fmt.Printf("当前时间范围获取 %d 条数据（达到阈值 %d），继续查询，新开始时间: %d (%s)\n",
					batchGoodsCount, maxRecordsPerRange, currentUpdateTimeFrom, time.Unix(currentUpdateTimeFrom, 0).Format("2006-01-02 15:04:05"))
			}

			// 检查新的开始时间是否已超过当前时间
			if currentUpdateTimeFrom > time.Now().Unix() {
				fmt.Printf("开始时间 %d (%s) 已超过当前时间 %d (%s)，停止获取\n",
					currentUpdateTimeFrom, time.Unix(currentUpdateTimeFrom, 0).Format("2006-01-02 15:04:05"),
					time.Now().Unix(), time.Now().Format("2006-01-02 15:04:05"))
				break
			}
		}
		if loopCount >= maxLoopCount {
			fmt.Printf("警告：已达到最大循环次数 %d，强制退出\n", maxLoopCount)
		}
	}
	return nil
}

// deduplicateToBodyOver 拉取任务读取body_wait去重复后写入到body_over中
func (xianYu *XianYu) deduplicateToBodyOver(duplicateCount *int, uniqueCount *int) error {
	page := 1
	pageSize := 100

	// 按店铺存储去重后的商品数据
	shopGoodsMap := make(map[string][]planBTypeXianyu.GoodsDetailRet)

	// 修改：使用复合key（店铺名+商品ID）进行去重，避免同一店铺重复相同商品
	processedKeys := make(map[string]bool)

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

	// 调试计数器
	debugCount := 0

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
			// 解析v到结构体
			goods := planAType.TaskBody{}
			jsonUnmarshalErr := json.Unmarshal([]byte(v), &goods)
			if jsonUnmarshalErr != nil {
				return fmt.Errorf("将json转为结构体失败: %v\n", jsonUnmarshalErr)
			}

			// 解析商品详情获取真实的商品ID
			var goodsItem planBTypeXianyu.GoodsDetailRet
			jsonUnmarshalErr = json.Unmarshal([]byte(goods.Detail.Error), &goodsItem)
			if jsonUnmarshalErr != nil {
				return fmt.Errorf("将json转为结构体失败: %v\n", jsonUnmarshalErr)
			}

			//提取 Isbn
			if goodsItem.BookData.ISBN == "" {
				goodsItem.BookData.ISBN = tool.ExtractISBN978(goodsItem.Title)
			}
			//Isbn 为空则跳过
			if goodsItem.BookData.ISBN == "" {
				fmt.Println("####################商品无法获取 Isbn，跳过：", goodsItem.Title)
				continue
			}

			// 使用 ProductID 作为唯一标识（int64类型）
			goodsId := goodsItem.ProductID

			// 获取闲鱼会员名
			username := ""
			if len(goodsItem.PublishShop) > 0 {
				username = goodsItem.PublishShop[0].UserName
			}

			// 修改：使用店铺名+商品ID作为复合key进行去重
			uniqueKey := fmt.Sprintf("%s_%d", username, goodsId)

			// 调试：打印前10条数据的ID
			if debugCount < 10 {
				fmt.Printf("[去重调试] 第%d条 - 商品ID: %d, 店铺: %s, 复合Key: %s, Title: %s\n",
					debugCount+1, goodsId, username, uniqueKey, goodsItem.Title)
				debugCount++
			}

			if !processedKeys[uniqueKey] {
				// 标记为已处理
				processedKeys[uniqueKey] = true
				*uniqueCount++

				// 按店铺暂存数据
				shopGoodsMap[username] = append(shopGoodsMap[username], goodsItem)

				// 写入到body_over
				goods.Detail.Status = 1
				addTaskToBodyOverErr := service.AddTaskToBodyOver(goods, []string{"body_over", "body_backup"})
				if addTaskToBodyOverErr != nil {
					return addTaskToBodyOverErr
				}

				// 检查每个店铺的商品数量，达到batchSize则推送
				if len(shopGoodsMap[username]) >= pageSize {
					fmt.Println("推送 username ", username, " 长度 ", len(shopGoodsMap[username]))
					if err := pushShopGoodsData(username, shopGoodsMap[username], int64(page), pageTotal, &num); err != nil {
						return err
					}
					// 清空该店铺的数据
					shopGoodsMap[username] = []planBTypeXianyu.GoodsDetailRet{}
				}
			} else {
				// 重复数据 计次
				*duplicateCount++
				// 调试：打印前10条重复数据
				if *duplicateCount <= 10 {
					fmt.Printf("[去重调试] 发现重复商品: 店铺=%s, 商品ID=%d, 复合Key=%s\n", username, goodsId, uniqueKey)
				}
			}
		}

		page++

		// 更新进度
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

		// 暂停1秒
		time.Sleep(1 * time.Second)
	}

	// 循环结束后，推送所有店铺剩余的数据（不足batchSize的部分）
	for shopId, goodsList := range shopGoodsMap {
		if len(goodsList) > 0 {
			fmt.Println("最后一次 推送 username ", shopId, " 长度 ", len(goodsList))
			if err := pushShopGoodsData(shopId, goodsList, pageTotal, pageTotal, &num); err != nil {
				return err
			}

			//打印每个店铺的商品数量
			fmt.Println("推送 username ", shopId, " 长度 ", len(goodsList))
		}
	}

	// 删除body_wait
	deleteTaskBodyWaitErr := service.DeleteTaskBodyWait()
	if deleteTaskBodyWaitErr != nil {
		return deleteTaskBodyWaitErr
	}

	fmt.Printf("[去重完成] 总处理: %d, 唯一: %d, 重复: %d\n",
		*uniqueCount+*duplicateCount, *uniqueCount, *duplicateCount)

	return nil
}
func (xianYu *XianYu) deduplicateToBodyOver1(duplicateCount *int, uniqueCount *int) error {
	page := 1
	pageSize := 100

	// 按店铺存储去重后的商品数据
	shopGoodsMap := make(map[string][]planBTypeXianyu.GoodsDetailRet)
	// 已处理的商品ID集合（用于全局去重）
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

	// 调试计数器
	debugCount := 0

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
			// 解析v到结构体
			goods := planAType.TaskBody{}
			jsonUnmarshalErr := json.Unmarshal([]byte(v), &goods)
			if jsonUnmarshalErr != nil {
				return fmt.Errorf("将json转为结构体失败: %v\n", jsonUnmarshalErr)
			}

			// 解析商品详情获取真实的商品ID
			var goodsItem planBTypeXianyu.GoodsDetailRet
			jsonUnmarshalErr = json.Unmarshal([]byte(goods.Detail.Error), &goodsItem)
			if jsonUnmarshalErr != nil {
				return fmt.Errorf("将json转为结构体失败: %v\n", jsonUnmarshalErr)
			}

			//提取 Isbn
			if goodsItem.BookData.ISBN == "" {
				goodsItem.BookData.ISBN = tool.ExtractISBN978(goodsItem.Title)
			}
			//Isbn 为空则跳过
			if goodsItem.BookData.ISBN == "" {
				fmt.Println("####################商品无法获取 Isbn，跳过：", goodsItem.Title)
				continue
			}

			// 使用 ProductID 作为唯一标识（int64类型）
			goodsId := goodsItem.ProductID

			// 调试：打印前10条数据的ID
			if debugCount < 10 {
				fmt.Printf("[去重调试] 第%d条 - 商品ID: %d, ProductID: %d, Title: %s\n",
					debugCount+1, goodsId, goodsItem.ProductID, goodsItem.Title)
				debugCount++
			}

			if !processedGoodsIds[goodsId] {
				// 标记为已处理
				processedGoodsIds[goodsId] = true
				*uniqueCount++

				// 获取闲鱼会员名
				username := ""
				if len(goodsItem.PublishShop) > 0 {
					username = goodsItem.PublishShop[0].UserName
				}

				// 按店铺暂存数据
				shopGoodsMap[username] = append(shopGoodsMap[username], goodsItem)

				// 写入到body_over
				goods.Detail.Status = 1
				addTaskToBodyOverErr := service.AddTaskToBodyOver(goods, []string{"body_over", "body_backup"})
				if addTaskToBodyOverErr != nil {
					return addTaskToBodyOverErr
				}

				// 检查每个店铺的商品数量，达到batchSize则推送
				if len(shopGoodsMap[username]) >= pageSize {
					fmt.Println("推送 username ", username, " 长度 ", len(shopGoodsMap[username]))
					if err := pushShopGoodsData(username, shopGoodsMap[username], int64(page), pageTotal, &num); err != nil {
						return err
					}
					// 清空该店铺的数据
					shopGoodsMap[username] = []planBTypeXianyu.GoodsDetailRet{}
				}
			} else {
				// 重复数据 计次
				*duplicateCount++
				// 调试：打印前10条重复数据
				if *duplicateCount <= 10 {
					fmt.Printf("[去重调试] 发现重复商品: %d\n", goodsId)
				}
			}
		}

		page++

		// 更新进度
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

		// 暂停1秒
		time.Sleep(1 * time.Second)
	}

	// 循环结束后，推送所有店铺剩余的数据（不足batchSize的部分）
	for shopId, goodsList := range shopGoodsMap {
		if len(goodsList) > 0 {
			fmt.Println("最后一次 推送 username ", shopId, " 长度 ", len(goodsList))
			if err := pushShopGoodsData(shopId, goodsList, pageTotal, pageTotal, &num); err != nil {
				return err
			}

			//打印每个店铺的商品数量
			fmt.Println("推送 username ", shopId, " 长度 ", len(goodsList))
		}
	}

	// 删除body_wait
	deleteTaskBodyWaitErr := service.DeleteTaskBodyWait()
	if deleteTaskBodyWaitErr != nil {
		return deleteTaskBodyWaitErr
	}

	fmt.Printf("[去重完成] 总处理: %d, 唯一: %d, 重复: %d\n",
		*uniqueCount+*duplicateCount, *uniqueCount, *duplicateCount)

	return nil
}

// pushShopGoodsData 推送单个店铺的商品数据到接口
func pushShopGoodsData(username string, goodsList []planBTypeXianyu.GoodsDetailRet, currentPage int64, totalPage int64, totalCount *int) error {
	if len(goodsList) == 0 {
		return nil
	}

	// 将获取的数据推送写入店铺商品数据接口
	ret, retStr, writeXianyuGoodsDataErr := writeXianyuGoodsData(goodsList, username, int(currentPage), totalPage)
	if writeXianyuGoodsDataErr != nil {
		return writeXianyuGoodsDataErr
	}
	if ret.Code != "200" {
		return fmt.Errorf("添加商品失败 %v", retStr)
	}

	*totalCount = *totalCount + len(goodsList)
	fmt.Printf("推送店铺 %s 的商品数据，当前批次数量 %v 总数据量 %v \n", username, len(goodsList), *totalCount)

	return nil
}

// processGoodsDetail 处理单个商品的详情并写入数据库（修改版）
func (xianYu *XianYu) processGoodsDetail(goodsItem planBTypeXianyu.GoodsListRetProduct, token planBTypeXianyu.Token) error {
	// 获取商品详情
	xianYuDetailReq := planBTypeXianyu.GoodsDetailReq{
		AppId:     token.AppId,
		AppSecret: token.AppSecret,
		ProductId: goodsItem.ProductID,
	}

	detailJson, err := xianYu.getGoodsDetail(xianYuDetailReq)
	if err != nil {
		return err
	}

	var goodDetailRet planBTypeXianyu.GoodDetailRet
	err = json.Unmarshal([]byte(detailJson), &goodDetailRet)
	if err != nil {
		return err
	}

	// 检查返回码
	if goodDetailRet.Code != 0 {
		return fmt.Errorf("获取商品详情失败 code=%d, msg=%s", goodDetailRet.Code, goodDetailRet.Msg)
	}

	// 将内容转为 json
	detailDataJson, marshalErr := json.Marshal(goodDetailRet.Data)
	if marshalErr != nil {
		return marshalErr
	}

	// 调试：打印商品ID信息
	//fmt.Printf("[商品处理] ProductID: %v, 详情ProductID: %d, Title: %s\n",
	//	goodsItem.ProductID, goodDetailRet.Data.ProductID, goodDetailRet.Data.Title)

	// 构建任务数据
	bodyWait := planAType.TaskBody{
		BookInfo: planAType.BookInfo{
			Isbn:            goodDetailRet.Data.BookData.ISBN,
			BookName:        goodDetailRet.Data.Title,
			Author:          goodDetailRet.Data.BookData.Author,
			Publishing:      goodDetailRet.Data.BookData.Publisher,
			PublicationDate: "",
			Binding:         "",
			PagesCount:      0,
			WordsCount:      0,
			Format:          0,
			Price:           goodsItem.UpdateTime, // 使用UpdateTime作为时间戳
		},
		Detail: planAType.TaskDetail{
			Error:   string(detailDataJson),
			GoodsId: goodDetailRet.Data.ProductID, // 使用详情中的ProductID
			Stock:   goodDetailRet.Data.Stock,
		},
	}

	// 验证商品 ID不为空
	if bodyWait.Detail.GoodsId == 0 {
		fmt.Printf("[警告] 商品 %s 的GoodsId为0\n", goodsItem.ProductID)
	}

	// 写入数据库
	bodyWaitJson, err := json.Marshal(bodyWait)
	if err != nil {
		return fmt.Errorf("将bodyWait转为json失败: %v", err)
	}

	return service.AddTaskToBodyWait(string(bodyWaitJson))
}

// 发送商品列表请求
func (xianYu *XianYu) sendGoodsListRequest(req planBTypeXianyu.GoodsListReq) (string, error) {
	reqJson, marshalErr := json.Marshal(req)
	if marshalErr != nil {
		return "", marshalErr
	}

	listJson, err := golabl.XianYuDll.XianYuGetGoodsList(string(reqJson), golabl.Config.FileUrl.XianYuDll)
	if err != nil {
		return "", err
	}
	return listJson, nil
}

// 获取商品详情
func (xianYu *XianYu) getGoodsDetail(req planBTypeXianyu.GoodsDetailReq) (string, error) {
	reqJson, marshalErr := json.Marshal(req)
	if marshalErr != nil {
		return "", marshalErr
	}

	detailJson, err := golabl.XianYuDll.XianYuGetGoodsDetail(string(reqJson), golabl.Config.FileUrl.XianYuDll)
	if err != nil {
		return "", err
	}
	return detailJson, nil
}

// writeXianyuGoodsData 写入商品数据
// @param goodsListStr 商品列表
// @param username 闲鱼会员名
// @param page 当前页
// @param pageTotal 总页数
// @return error 错误信息
func writeXianyuGoodsData(goodsListStr []planBTypeXianyu.GoodsDetailRet, username string, page int, pageTotal int64) (planBType.AsyncTaskResponse, string, error) {
	var ret planBType.AsyncTaskResponse
	marshal, marshalErr := json.Marshal(goodsListStr)
	if marshalErr != nil {
		return ret, "", marshalErr
	}
	params := map[string]string{
		"taskId":       golabl.Task.TaskId,
		"shopId":       username,
		"goodsListStr": string(marshal),
		"allNum":       strconv.FormatInt(pageTotal, 10),
		"num":          strconv.Itoa(page),
	}
	retStr, submitFormDataErr := tool.SubmitFormData(golabl.Config.FileUrl.XianYuAddGoodsUrl, params)
	if submitFormDataErr != nil {
		return ret, retStr, submitFormDataErr
	}
	unmarshalErr := json.Unmarshal([]byte(retStr), &ret)
	if unmarshalErr != nil {
		return ret, retStr, unmarshalErr
	}
	return ret, retStr, nil
}

// executeGoodsLaunch 上架商品
// @param logUuid 日志ID
// @param taskMsg 任务内容
// @return error 错误信息
func executeGoodsLaunch(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	// 解析应用 id与应用秘钥
	var token planBTypeXianyu.Token
	unmarshalErr := json.Unmarshal([]byte(golabl.Task.Header.ShopMsg.Token), &token)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("解析应用id与应用秘钥 taskHeader.ShopMsg.Token = %v %w", golabl.Task.Header.ShopMsg.Token, unmarshalErr))
	}

	// 上架商品
	launchGoodsInfo := planBTypeXianyu.Product{
		AppId:              token.AppId,
		AppSecret:          token.AppSecret,
		ProductID:          taskMsg.Detail.GoodsId,
		SpecifyPublishTime: "",
		UserName:           []string{token.Username},
	}
	//转为json
	jsonData, marshalErr := json.Marshal(launchGoodsInfo)
	if marshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, marshalErr)
	}
	var launchGoods planBTypeXianyu.XianYuAddGoodsResponse
	launchGoodsStr, xianYuLaunchGoodsAddErr := golabl.XianYuDll.XianYuLaunchGoods(string(jsonData), golabl.Config.FileUrl.XianYuDll)
	if xianYuLaunchGoodsAddErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, xianYuLaunchGoodsAddErr)
	}
	unmarshalErr = json.Unmarshal([]byte(launchGoodsStr), &launchGoods)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, unmarshalErr)
	}
	if launchGoods.Code != 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("上架商品失败 %s", launchGoods.Msg))
	}
	return tool.ReturnSuccess(taskMsg)
}

// executeGoodsDownShelf 下架商品
// @param logUuid 日志ID
// @param taskMsg 任务内容
// @return error 错误信息
func executeGoodsDownShelf(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	var downShelf planBTypeXianyu.DownShelf

	// 解析应用 id与应用秘钥
	var token planBTypeXianyu.Token
	unmarshalErr := json.Unmarshal([]byte(golabl.Task.Header.ShopMsg.Token), &token)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("解析应用id与应用秘钥 taskHeader.ShopMsg.Token = %v %w", golabl.Task.Header.ShopMsg.Token, unmarshalErr))
	}
	downShelf.AppId = token.AppId
	downShelf.AppSecret = token.AppSecret
	downShelf.ProductID = taskMsg.Detail.GoodsId
	//转为json
	jsonData, marshalErr := json.Marshal(downShelf)
	if marshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, marshalErr)
	}
	shelf, xianYuExecuteGoodsDownShelfErr := golabl.XianYuDll.XianYuExecuteGoodsDownShelf(string(jsonData), golabl.Config.FileUrl.XianYuDll)
	if xianYuExecuteGoodsDownShelfErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, xianYuExecuteGoodsDownShelfErr)
	}
	var downShelfRes planBTypeXianyu.XianYuAddGoodsResponse
	unmarshalErr = json.Unmarshal([]byte(shelf), &downShelfRes)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, unmarshalErr)
	}
	if downShelfRes.Code != 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("下架商品失败 %s", downShelfRes.Msg))
	}
	return tool.ReturnSuccess(taskMsg)
}

// executeGoodsUpdateStock 修改库存
func executeGoodsUpdateStock(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	var updateStock planBTypeXianyu.UpdateStock

	// 解析应用 id与应用秘钥
	var token planBTypeXianyu.Token
	unmarshalErr := json.Unmarshal([]byte(golabl.Task.Header.ShopMsg.Token), &token)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("解析应用id与应用秘钥 taskHeader.ShopMsg.Token = %v %w", golabl.Task.Header.ShopMsg.Token, unmarshalErr))
	}
	updateStock.AppId = token.AppId
	updateStock.AppSecret = token.AppSecret
	updateStock.ProductID = taskMsg.Detail.GoodsId
	updateStock.Stock = taskMsg.Detail.Stock
	//转为json
	jsonData, marshalErr := json.Marshal(updateStock)
	if marshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, marshalErr)
	}
	updateStockStr, xianYuExecuteGoodsUpdateStockErr := golabl.XianYuDll.XianYuExecuteGoodsUpdateStock(string(jsonData), golabl.Config.FileUrl.XianYuDll)
	if xianYuExecuteGoodsUpdateStockErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, xianYuExecuteGoodsUpdateStockErr)
	}
	var updateStockRes planBTypeXianyu.XianYuAddGoodsResponse
	unmarshalErr = json.Unmarshal([]byte(updateStockStr), &updateStockRes)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, unmarshalErr)
	}
	if updateStockRes.Code != 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("修改商库存品失败 %s", updateStockRes.Msg))
	}
	return tool.ReturnSuccess(taskMsg)
}

// 修改价格
func executeGoodsUpdatePrice(logUuid string, taskMsg planAType.TaskBody) (string, error) {
	var updatePrice planBTypeXianyu.UpdatePrice

	// 解析应用 id与应用秘钥
	var token planBTypeXianyu.Token
	unmarshalErr := json.Unmarshal([]byte(golabl.Task.Header.ShopMsg.Token), &token)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("解析应用id与应用秘钥 taskHeader.ShopMsg.Token = %v %w", golabl.Task.Header.ShopMsg.Token, unmarshalErr))
	}
	updatePrice.AppId = token.AppId
	updatePrice.AppSecret = token.AppSecret
	updatePrice.ProductID = taskMsg.Detail.GoodsId
	updatePrice.Price = taskMsg.Detail.Price
	updatePrice.OriginalPrice = tool.BuildGoodsPrice(taskMsg.Detail.Price)
	//转为json
	jsonData, marshalErr := json.Marshal(updatePrice)
	if marshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, marshalErr)
	}
	updatePriceStr, xianYuExecuteGoodsUpdatePrice := golabl.XianYuDll.XianYuExecuteGoodsUpdatePrice(string(jsonData), golabl.Config.FileUrl.XianYuDll)
	if xianYuExecuteGoodsUpdatePrice != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, xianYuExecuteGoodsUpdatePrice)
	}
	var updatePriceRes planBTypeXianyu.XianYuAddGoodsResponse
	unmarshalErr = json.Unmarshal([]byte(updatePriceStr), &updatePriceRes)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, unmarshalErr)
	}
	if updatePriceRes.Code != 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("修改商品价格失败 %s", updatePriceRes.Msg))
	}
	return tool.ReturnSuccess(taskMsg)
}
