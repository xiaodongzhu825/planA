package pinduoduo

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/planB/initialization/golabl"
	"planA/planB/modules/logs"
	"planA/planB/modules/pdd"
	"planA/planB/server"
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
	goodsAdd.CarouselGallery = tool.BuildCarouselGallery(golabl.Task.Header.ShopMsg.WatermarkImgUrl, golabl.Task.Header.ShopMsg.CarouseLastImgUrlArray, taskMsg.BookInfo.ImageObject.CarouselUrlArray)
	if len(goodsAdd.CarouselGallery) == 0 && golabl.Task.Header.ImgType == 3 && taskMsg.BookInfo.ImageObject.DefaultImageUrl != "" {
		goodsAdd.CarouselGallery = append(goodsAdd.CarouselGallery, taskMsg.BookInfo.ImageObject.DefaultImageUrl)
	}
	if len(taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl) == 0 && len(goodsAdd.CarouselGallery) > 0 {
		taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl = []string{goodsAdd.CarouselGallery[0]}
	}
	if len(goodsAdd.CarouselGallery) == 0 {
		// 无图片信息 isbn计次
		setNoImgCountErr := server.SetNoImgCount(taskMsg.BookInfo.Isbn)
		if setNoImgCountErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("无图片信息isbn计次错误 isbn %v %v", taskMsg.BookInfo.Isbn, setNoImgCountErr.Error()))
		}
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("缺少构造轮播图图片-未提交 isbn %v", taskMsg.BookInfo.Isbn))
	}

	// 构建详情图
	goodsAdd.DetailGallery = tool.BuildDetailGallery(golabl.Task.Header.ShopMsg.WatermarkImgUrl, golabl.Task.Header.ShopMsg.GoodsDetailFirstImgUrlArray, golabl.Task.Header.ShopMsg.GoodsDetailLastImgUrlArray, taskMsg.BookInfo.ImageObject.DetailUrlObject)

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
	sku, err := buildSkuList(pddDll, golabl.Task.Header.ShopMsg.Token, golabl.Task.Header.ShopMsg.SpecId, golabl.Task.Header.ShopMsg.SpecName, golabl.Task.Header.ShopMsg.SpecChildName, price, goodsAdd.CarouselGallery[0], taskMsg.Detail.Stock, taskMsg.Detail.SkuCode)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, err)
	}
	goodsAdd.SkuList = []planBTypePinduoduo.Sku{sku}

	// *********************构建参数 结束******************************** //

	// 发送请求
	goodsAddRet, _, err := addGoods(pddDll, logUuid, golabl.Task.Header.ShopMsg.Token, goodsAdd)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("商品提交 %v", err))
	}

	// 获取商品提交的商品详情
	goodsCommitDetail, _, getGoodsCommitDetailErr := getGoodsCommitDetail(pddDll, golabl.Task.Header.ShopMsg.Token, goodsAddRet.Response.GoodsCommitID, goodsAddRet.Response.GoodsID)
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

func (pinDuoDuo *PinDuoDuo) GetGoodsTask() string {
	return ""
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
// @param token 授权令牌
// @param specId 商品规格id
// @param specName 商品规格名称
// @param specChildName 商品规格子项名称
// @param price 价格
// @param thumbUrl 缩略图
// @param stock 库存
// @param outSkuSn 商品编码
// @return Sku sku规格
// @return error 错误信息
func buildSkuList(pddDll *pdd.PddDLL, token string, specId int64, specName string, specChildName string, price int64, thumbUrl string, stock int64, outSkuSn string) (planBTypePinduoduo.Sku, error) {
	// 构建Spec列表
	var sku planBTypePinduoduo.Sku
	goodsSpec, buildPddGoodsSpecIdErr := buildPddGoodsSpecId(pddDll, token, specId, specChildName)
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
// @param token 授权令牌
// @param specId 商品规格id
// @param specName 规格名称
// @return DllGoodsSpec 规格信息
// @return error 错误信息
func buildPddGoodsSpecId(pddDll *pdd.PddDLL, token string, id int64, name string) (planAType.DllGoodsSpec, error) {
	var spec planAType.DllGoodsSpec
	specStr, err := pddDll.PddGoodsSpecIdGet(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, token, strconv.FormatInt(id, 10), name)
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
// @param token 授权令牌
// @param logUuid 日志ID
// @param goodsInfo 商品信息
// @return GoodsAddResponseWrapper 商品新增结果
// @return string 商品新增结果json
// @return error 错误信息
func addGoods(pddDll *pdd.PddDLL, logUuid string, token string, goodsInfo planBTypePinduoduo.GoodsAdd) (planBTypePinduoduo.GoodsAddResponseWrapper, string, error) {
	var goodsAdd planBTypePinduoduo.GoodsAddResponseWrapper
	goodsInfoStr, jsonMarshalErr := json.Marshal(goodsInfo)
	if jsonMarshalErr != nil {
		return goodsAdd, "", jsonMarshalErr
	}
	//发送请求
	goodsAddStr, pddGoodsAddErr := pddDll.PddGoodsAdd(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, token, string(goodsInfoStr))
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
// @param token 授权令牌
// @param goodsCommitId 商品提交ID
// @param goodsId 商品ID
// @return GoodsCommitDetailResponse 商品提交详情
// @return error 错误信息
func getGoodsCommitDetail(pddDll *pdd.PddDLL, token string, goodsCommitId int64, goodsId int64) (planBTypePinduoduo.GoodsCommitDetailResponse, string, error) {
	var goodsCommitDetail planBTypePinduoduo.GoodsCommitDetailResponse
	goodsCommitDetailStr, pddGoodsCommitDetailGetErr := pddDll.PddGoodsCommitDetailGet(golabl.Config.PddConfig.ClientId, golabl.Config.PddConfig.ClientSecret, token, strconv.FormatInt(goodsCommitId, 10), strconv.FormatInt(goodsId, 10))
	if pddGoodsCommitDetailGetErr != nil {
		return goodsCommitDetail, "", pddGoodsCommitDetailGetErr
	}
	unmarshalErr := json.Unmarshal([]byte(goodsCommitDetailStr), &goodsCommitDetail)
	if unmarshalErr != nil {
		return goodsCommitDetail, "", fmt.Errorf("解析拼多多 PddGoodsCommitDetailGet 接口返回json失败: %v [拼多多数据：%v]", unmarshalErr, goodsCommitDetailStr)
	}
	return goodsCommitDetail, goodsCommitDetailStr, nil
}
