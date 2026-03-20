package pinDuoDuo

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/modules/logs"
	"planA/planB/config"
	"planA/planB/db/redis"
	"planA/planB/golabl"
	"planA/planB/modules/pdd"
	"planA/planB/tool"
	_type "planA/planB/type"
	"strconv"
	"strings"
	"time"
)

type PinDuoDuo struct {
}
type GoodsAdd struct {
	GoodsName           string            `json:"goods_name"`            // 商品名称
	CarouselGallery     []string          `json:"carousel_gallery"`      // 轮播图
	CatId               int64             `json:"cat_id"`                // 商品分类
	GoodsType           int64             `json:"goods_type"`            // 商品类型 1-国内普通商品，2-一般贸易，3-保税仓BBC直供，4-海外BC直邮 ,5-流量 ,6-话费 ,7-优惠券 ,8-QQ充值 ,9-加油卡，15-商家卡券，18-海外CC行邮 19-平台卡券
	MarketPrice         int64             `json:"market_price"`          // 参考价格，单位为分
	DetailGallery       []string          `json:"detail_gallery"`        // 详情图
	OutGoodsId          string            `json:"out_goods_id"`          // 商品ID
	SkuList             []Sku             `json:"sku_list"`              // SKU列表
	IsFolt              bool              `json:"is_folt"`               // 是否支持假一赔十，false-不支持，true-支持
	IsPreSale           bool              `json:"is_pre_sale"`           // 是否预售,true-预售商品，false-非预售商品
	IsRefundable        bool              `json:"is_refundable"`         // 是否7天无理由退换货，true-支持，false-不支持
	SecondHand          bool              `json:"second_hand"`           // 是否二手商品， true -二手商品 ，false-全新商品
	CostTemplateId      int64             `json:"cost_template_id"`      // 物流运费模板ID
	CountryId           int64             `json:"country_id"`            // 国家ID
	ShipmentLimitSecond int64             `json:"shipment_limit_second"` // 承诺发货时间（秒），普通、进口商品可选48小时或24小时；直邮商品（goods_type=4）只可入参120小时，直供商品（goods_type=3）只可入参96小时；is_pre_sale为true时不必传
	TwoPiecesDiscount   int64             `json:"two_pieces_discount"`   // 满2件折扣，可选范围0-100, 0表示取消，95表示95折，设置需先查询规则接口获取实际可填范围
	GoodsProperties     []GoodsProperties `json:"goods_properties"`      //商品属性列表
}
type GoodsProperties struct {
	RefPid    int64  `json:"ref_pid"`    // 属性名称
	Value     string `json:"value"`      // 属性值
	ValueUnit string `json:"value_unit"` // 属性单位
	Vid       int64  `json:"vid"`        // 属性值 id
}

type Sku struct {
	IsOnsale      int64         `json:"is_onsale"`      // sku上架状态，0-已下架，1-上架中
	LimitQuantity int64         `json:"limit_quantity"` // sku购买限制，只入参999
	MultiPrice    int64         `json:"multi_price"`    // 商品团购价格，单位为分
	Price         int64         `json:"price"`          // 商品单买价格，单位为分
	SkuProperties []SkuProperty `json:"sku_properties"` // sku属性列表
	Quantity      int64         `json:"quantity"`       // 商品sku库存初始数量，后续库存update只使用stocks.update接口进行调用
	ThumbUrl      string        `json:"thumb_url"`      // sku 缩略图
	SpecIdList    string        `json:"spec_id_list"`   // 商品规格列表，根据pdd.goods.spec.id.get生成的规格属性id，例如：颜色规格下商家新增白色和黑色，大小规格下商家新增L和XL，则由4种spec组合，入参一种组合即可，在skulist中需要有4个spec组合的sku，示例：[20,5]
	Weight        int64         `json:"weight"`         // 重量，单位为g
	OutSkuSn      string        `json:"out_sku_sn"`     // 商品 sku编号
}

type SkuProperty struct {
	Punit  string `json:"punit"`   // 属性单位
	RefPid int64  `json:"ref_pid"` // 属性id
	Value  string `json:"value"`   // 属性值
	Vid    int64  `json:"vid"`     // 属性值id
}

// PriceMod 价格处理
type PriceMod struct {
	Min         int64 `json:"min"`          // 价格区间最小值
	Max         int64 `json:"max"`          // 价格区间最大值
	MarkupRate  int64 `json:"markup_rate"`  // 加价比例
	MarkupValue int64 `json:"markup_value"` // 价格区间加价值
}

// GoodsCommitDetail 获取商品提交的商品详情
type GoodsCommitDetail struct {
	GoodsCommitId int64 `json:"goods_commit_id"` // 商品提交id
	GoodsId       int64 `json:"goods_id"`        // 商品id
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
func (pinDuoDuo *PinDuoDuo) AddGoodsTask(taskHeader _type.TaskHeader, taskMsg _type.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}
	//拼接返回结果
	price := tool.BuildPrice(taskHeader.PriceMod, taskMsg.Detail.Price)
	if price == 0 {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("不在价格区间内 isbn:%v", taskMsg.BookInfo.Isbn))
	}
	taskMsg.Detail.Price = price
	//TODO
	// 构建参数
	var goodsAdd GoodsAdd
	pddDll, err := pdd.InitPddDll()
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("初始化拼多多DLL失败 %v", err))
	}
	//构建商品名称
	goodsAdd.GoodsName = tool.BuildGoodsName(
		taskHeader.ShopMsg.GoodsNamePrefix, // 商品名称前缀
		taskHeader.ShopMsg.GoodsNameSuffix, // 商品名称后缀
		taskHeader.ShopMsg.TitleConsistOf,  // 标题组成
		taskHeader.ShopMsg.SpaceCharacter,  // 间隔符
		taskMsg.BookInfo)                   // 图书信息
	taskMsg.Detail.GoodsName = goodsAdd.GoodsName
	// 构建轮播图
	//if taskHeader.ShopMsg.WatermarkImgUrl == "" && len(taskHeader.ShopMsg.CarouseLastImgUrlArray) == 0 && len(taskMsg.BookInfo.ImageObject.CarouselUrlArray) == 0 && taskMsg.BookInfo.ImageObject.DefaultImageUrl == "" {
	//	return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd,fmt.Errorf("缺少构造轮播图图片-未提交 isbn %v", taskMsg.BookInfo.Isbn))
	//}
	goodsAdd.CarouselGallery = tool.BuildCarouselGallery(taskHeader.ShopMsg.WatermarkImgUrl, taskHeader.ShopMsg.CarouseLastImgUrlArray, taskMsg.BookInfo.ImageObject.CarouselUrlArray)
	if len(goodsAdd.CarouselGallery) == 0 && taskHeader.ImgType == 3 && taskMsg.BookInfo.ImageObject.DefaultImageUrl != "" {
		goodsAdd.CarouselGallery = append(goodsAdd.CarouselGallery, taskMsg.BookInfo.ImageObject.DefaultImageUrl)
	}
	if len(taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl) == 0 && len(goodsAdd.CarouselGallery) > 0 {
		taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl = []string{goodsAdd.CarouselGallery[0]}
	}
	if len(goodsAdd.CarouselGallery) == 0 {
		// 无图片信息 isbn计次
		setNoImgCountErr := redis.SetNoImgCount(golabl.RedisClientD, taskMsg.BookInfo.Isbn)
		if setNoImgCountErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("无图片信息isbn计次错误 isbn %v %v", taskMsg.BookInfo.Isbn, setNoImgCountErr.Error()))
		}
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("缺少构造轮播图图片-未提交 isbn %v", taskMsg.BookInfo.Isbn))
	}
	// 构建详情图
	goodsAdd.DetailGallery = tool.BuildDetailGallery(taskHeader.ShopMsg.WatermarkImgUrl, taskHeader.ShopMsg.GoodsDetailFirstImgUrlArray, taskHeader.ShopMsg.GoodsDetailLastImgUrlArray, taskMsg.BookInfo.ImageObject.DetailUrlObject)

	// 构建 catId
	var catID int64

	if taskMsg.BookInfo.CatIdObject.PinDuoDuoCatId == "" {
		// 获取拼多多配置
		pddConfig, getPddClientErr := config.GetPddClient()
		if getPddClientErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("获取拼多多配置失败 %w", getPddClientErr))
		}
		// 调用拼多多 SDK 取类目信息
		pddCalbackStr, pddGoodsOuterCatMappingGetErr := pddDll.PddGoodsOuterCatMappingGet(pddConfig.ClientId, pddConfig.ClientSecret, taskHeader.ShopMsg.Token, "15543", "书籍/杂志/报纸", "书籍 "+taskMsg.BookInfo.BookName)
		if pddGoodsOuterCatMappingGetErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("调用DLL类目映射失败 %w", err))
		}

		// 解析返回的 JSON 字符串
		var response _type.PddSuccessResponse
		if unmarshalErr := json.Unmarshal([]byte(pddCalbackStr), &response); unmarshalErr != nil {
			return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("json.Unmarshal错误 %w %v", unmarshalErr, pddCalbackStr))
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
			return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("转换catId错误 %w", toInt64Err))
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
	goodsAdd.IsFolt = taskHeader.ShopMsg.IsFolt

	// 是否预售
	goodsAdd.IsPreSale = taskHeader.ShopMsg.IsPreSale

	// 是否支持7天无理由退换货
	goodsAdd.IsRefundable = taskHeader.ShopMsg.IsRefundable

	// 构建是否是二手商品
	goodsAdd.SecondHand = taskHeader.ShopMsg.IsSecondHand

	// 构建物流运费模板 ID
	goodsAdd.CostTemplateId = taskHeader.ShopMsg.CostTemplateId

	// 构建承诺发货时间
	goodsAdd.ShipmentLimitSecond = 48 * 60 * 60

	//满两件折扣
	goodsAdd.TwoPiecesDiscount = taskHeader.ShopMsg.TwoDiscount

	//构建
	taskMsgBookInfoPrice := taskMsg.BookInfo.Price
	if taskMsgBookInfoPrice < 10000 {
		taskMsgBookInfoPrice = 10000
	}

	goodsAdd.GoodsProperties = BuildGoodsPropertiesList(
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
	if taskMsg.Detail.Stock == 0 && taskHeader.TaskType == 1 {
		//如果库存为0 则给默认库存
		taskMsg.Detail.Stock = taskHeader.ShopMsg.DefStock
	}

	//生成一个2秒的延迟
	url := "http://127.0.0.1:8095"
	tool.HttpGetRequest(url)

	// 规格编号
	if taskMsg.Detail.SkuCode == "" {
		taskMsg.Detail.SkuCode = goodsAdd.OutGoodsId
	}

	//获取 sku
	sku, err := BuildSkuList(pddDll, taskHeader.ShopMsg.Token, taskHeader.ShopMsg.SpecId, taskHeader.ShopMsg.SpecName, taskHeader.ShopMsg.SpecChildName, price, goodsAdd.CarouselGallery[0], taskMsg.Detail.Stock, taskMsg.Detail.SkuCode)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, err)
	}
	goodsAdd.SkuList = []Sku{sku}

	// 发送请求
	goodsAddRet, _, err := AddGoods(pddDll, logUuid, taskHeader.ShopMsg.Token, goodsAdd)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("商品提交 %v", err))
	}

	// 获取商品提交的商品详情
	goodsCommitDetail, _, getGoodsCommitDetailErr := GetGoodsCommitDetail(pddDll, taskHeader.ShopMsg.Token, goodsAddRet.Response.GoodsCommitID, goodsAddRet.Response.GoodsID)
	if getGoodsCommitDetailErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("获取商品提交的商品详情失败 %w", getGoodsCommitDetailErr))
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

// BuildSkuList sku规格生成
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
func BuildSkuList(pddDll *pdd.PddDLL, token string, specId int64, specName string, specChildName string, price int64, thumbUrl string, stock int64, outSkuSn string) (Sku, error) {
	// 构建Spec列表
	var sku Sku
	goodsSpec, buildPddGoodsSpecIdErr := buildPddGoodsSpecId(pddDll, token, specId, specChildName)
	if buildPddGoodsSpecIdErr != nil {
		return sku, buildPddGoodsSpecIdErr
	}

	// 构建SKU_Properties列表
	skuProperty := SkuProperty{
		Punit:  specName,                        // 属性单位
		RefPid: specId,                          // 属性id
		Value:  goodsSpec.DllGoodsSpec.SpecName, // 属性值
		Vid:    goodsSpec.DllGoodsSpec.SpecID,   // 属性值id
	}
	skuProperties := []SkuProperty{skuProperty}

	specIdList := "[" + strconv.FormatInt(skuProperty.Vid, 10) + "]"
	// 构建 SKU列表
	sku = Sku{
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
func buildPddGoodsSpecId(pddDll *pdd.PddDLL, token string, id int64, name string) (_type.DllGoodsSpec, error) {

	var spec _type.DllGoodsSpec
	client, err := config.GetPddClient()
	if err != nil {
		return spec, err
	}
	//发送请求 生成商家自定义的规格
	clientId := client.ClientId
	clientSecret := client.ClientSecret
	specStr, err := pddDll.PddGoodsSpecIdGet(clientId, clientSecret, token, strconv.FormatInt(id, 10), name)
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

// BuildGoodsPropertiesList 构建商品属性列表
// @param isbn
// @param bookName 书名
// @param pageCount 页数
// @param price 价格
// @param publishingVid 出版社Vid
// @param author 作者
// @param format 开本
// @param binding 装帧
// @return []GoodsProperties 商品属性列表
func BuildGoodsPropertiesList(isbn, bookName string, pageCount, price int64, publishingVid int64, author string, format int64, binding string) []GoodsProperties {
	var goodsPropertiesArr []GoodsProperties
	//isbn
	goodsPropertiesIsbn := GoodsProperties{
		RefPid: 425,
		Value:  isbn,
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesIsbn)

	//书名
	goodsPropertiesBookName := GoodsProperties{
		RefPid: 876,
		Value:  bookName,
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesBookName)

	//页数
	if pageCount == 0 {
		pageCount = 200
	}
	goodsPropertiesPageNum := GoodsProperties{
		RefPid:    692,
		Value:     strconv.FormatInt(pageCount, 10),
		ValueUnit: "页",
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesPageNum)

	//定价
	goodsPropertiesPrice := GoodsProperties{
		RefPid:    879,
		Value:     strconv.FormatInt(price/10000, 10),
		ValueUnit: "元",
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesPrice)

	//出版社
	goodsPropertiesPublishing := GoodsProperties{
		RefPid: 880,
		Vid:    publishingVid,
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesPublishing)

	//作者
	goodsPropertiesAuthor := GoodsProperties{
		RefPid: 882,
		Value:  author,
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesAuthor)

	//开本
	goodsPropertiesFormat := GoodsProperties{
		RefPid: 890,
		Value:  strconv.FormatInt(format, 10),
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesFormat)

	//装帧
	goodsPropertiesBinding := GoodsProperties{
		RefPid: 891,
		Value:  binding,
	}
	goodsPropertiesArr = append(goodsPropertiesArr, goodsPropertiesBinding)
	return goodsPropertiesArr
}

// AddGoods 商品新增
// @param pddDll pddDLL对象
// @param token 授权令牌
// @param logUuid 日志ID
// @param goodsInfo 商品信息
// @return GoodsAddResponseWrapper 商品新增结果
// @return string 商品新增结果json
// @return error 错误信息
func AddGoods(pddDll *pdd.PddDLL, logUuid string, token string, goodsInfo GoodsAdd) (_type.GoodsAddResponseWrapper, string, error) {
	var goodsAdd _type.GoodsAddResponseWrapper
	client, getPddClientErr := config.GetPddClient()
	if getPddClientErr != nil {
		return goodsAdd, "", getPddClientErr
	}
	goodsInfoStr, jsonMarshalErr := json.Marshal(goodsInfo)
	if jsonMarshalErr != nil {
		return goodsAdd, "", jsonMarshalErr
	}
	clientId := client.ClientId
	clientSecret := client.ClientSecret
	//发送请求
	goodsAddStr, pddGoodsAddErr := pddDll.PddGoodsAdd(clientId, clientSecret, token, string(goodsInfoStr))
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

// GetGoodsCommitDetail 获取商品提交的商品详情
// @param pddDll pddDLL对象
// @param token 授权令牌
// @param goodsCommitId 商品提交ID
// @param goodsId 商品ID
// @return GoodsCommitDetailResponse 商品提交详情
// @return error 错误信息
func GetGoodsCommitDetail(pddDll *pdd.PddDLL, token string, goodsCommitId int64, goodsId int64) (_type.GoodsCommitDetailResponse, string, error) {
	var goodsCommitDetail _type.GoodsCommitDetailResponse
	client, err := config.GetPddClient()
	if err != nil {
		return goodsCommitDetail, "", err
	}
	clientId := client.ClientId
	clientSecret := client.ClientSecret
	goodsCommitDetailStr, pddGoodsCommitDetailGetErr := pddDll.PddGoodsCommitDetailGet(clientId, clientSecret, token, strconv.FormatInt(goodsCommitId, 10), strconv.FormatInt(goodsId, 10))
	if pddGoodsCommitDetailGetErr != nil {
		return goodsCommitDetail, "", pddGoodsCommitDetailGetErr
	}
	unmarshalErr := json.Unmarshal([]byte(goodsCommitDetailStr), &goodsCommitDetail)
	if unmarshalErr != nil {
		return goodsCommitDetail, "", fmt.Errorf("解析拼多多 PddGoodsCommitDetailGet 接口返回json失败: %v [拼多多数据：%v]", unmarshalErr, goodsCommitDetailStr)
	}
	return goodsCommitDetail, goodsCommitDetailStr, nil
}
