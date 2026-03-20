package xianyu

import (
	"encoding/json"
	"errors"
	"fmt"
	_myRedis "planA/planB/db/redis"
	"planA/planB/golabl"
	"planA/planB/modules/logs"
	xianYuDll "planA/planB/modules/xianYu"
	"planA/planB/tool"
	_type "planA/planB/type"
	"strconv"
	"time"
)

type XianYu struct {
}

// GoodsAdd 商品新增请求结构体
// 用于向各电商平台（闲鱼、拼多多、淘宝等）提交商品上架的相关信息
type GoodsAdd struct {
	AppId        int64      `json:"appId"`        // 应用 id
	AppSecret    string     `json:"appSecret"`    // 应用密钥[选填，有些平台需要]
	Token        string     `json:"token"`        // token[选填，有些平台需要]
	ApiShopId    int        `json:"apiShopId"`    // API使用的店铺ID[选填，有些平台需要]
	TypePlatform int        `json:"typePlatform"` // 平台类型 0-预留  1-拼多多  2-淘宝  3-京东  4-闲鱼  105-孔夫子
	ShopId       int64      `json:"shopId"`       // 店铺 ID
	ShopToken    string     `json:"shopToken"`    // 店铺 Token
	ShopName     string     `json:"shopName"`     // 店铺名称
	Province     int        `json:"province"`     // 发货省，格式为省级行政区划代码（如210000代表辽宁省）
	City         int        `json:"city"`         // 发货市，格式为市级行政区划代码（如210100代表沈阳市）
	District     int        `json:"district"`     // 发货区，格式为区级行政区划代码（如210101代表和平区）
	TypeClass    string     `json:"typeClass"`    // 分类类型
	TypeGoods    string     `json:"typeGoods"`    // 商品类型
	CatIds       string     `json:"catIds"`       // 类目 ID
	SkuMsgs      []SkuMsg   `json:"skuMsgs"`      // 商品 SKU信息列表[选填]
	Shop         []ShopInfo `json:"shop"`         // 闲鱼用店铺信息
	StuffStatus  int64      `json:"stuffStatus"`  // 成色，90代表对应成色等级
	BookData     []BookInfo `json:"bookData"`     // 图书类商品专属信息列表
	ItemKey      string     `json:"itemKey"`      // 闲鱼批次商品 KEY
}

// ShopInfo 闲鱼店铺信息结构体
// 包含闲鱼平台商品上架所需的店铺及商品基础信息
type ShopInfo struct {
	UserName    string   `json:"userName"`    // 闲鱼会员名（必填）
	Province    int      `json:"province"`    // 发货省（必填），行政区划代码格式
	City        int      `json:"city"`        // 发货市（必填），行政区划代码格式
	District    int      `json:"district"`    // 发货区（必填），行政区划代码格式
	Title       string   `json:"title"`       // 商品标题（必填）
	Content     string   `json:"content"`     // 商品描述（必填）
	MainImgs    []string `json:"mainImgs"`    // 商品主图（必填），图片URL列表
	ContentImgs []string `json:"contentImgs"` // 商品内容图（选填），图片URL列表
}

// BookInfo 图书类商品信息结构体
// 包含图书类商品上架所需的专属信息，适用于闲鱼等平台的图书品类
type BookInfo struct {
	ISBN        string  `json:"ISBN"`        // ISBN编号（必填），图书唯一标识
	Title       string  `json:"Title"`       // 书名（必填）
	Author      string  `json:"Author"`      // 作者（选填）
	Publisher   string  `json:"Publisher"`   // 出版社（选填）
	ItemBizType int     `json:"itemBizType"` // 闲鱼商品类型（枚举），2：普通商品（必填）
	SpBizType   int     `json:"spBizType"`   // 闲鱼行业类型（枚举），24：图书（必填）
	Prices      []int64 `json:"prices"`      // 商品价格（必填），格式为[商品原价，商品售价]，单位为分
	Stock       int64   `json:"stock"`       // 库存（必填），商品可售数量
	CatIds      string  `json:"catIds"`      // 商品类目ID（必填）
}

// SkuMsg 商品SKU信息结构体
// 补充定义原需求中提到的skuMsgs字段对应的结构体，保证结构体完整性
type SkuMsg struct {
	Key         string   `json:"key"`         // 主键（必填）
	Value       string   `json:"value"`       // 值（必填）
	Title       string   `json:"title"`       // 商品标题（必填）
	CatIds      string   `json:"cat_ids"`     // 商品类目（必填）
	MainImgs    []string `json:"mainImgs"`    // 商品主图（必填），图片URL列表
	ContentImgs []string `json:"contentImgs"` // 商品内容图（选填），图片URL列表
	ItemBizType int      `json:"itemBizType"` // 闲鱼商品类型（枚举），2：普通商品（必填）
	SpBizType   int      `json:"spBizType"`   // 闲鱼行业类型（枚举），24：图书（必填）
	Prices      []int    `json:"prices"`      // 商品价格（必填），[商品售价，商品原价]，单位为分
	Stock       int      `json:"stock"`       // 库存（必填）
	Content     string   `json:"content"`     // 商品描述（必填）
	UserName    string   `json:"userName"`    // 闲鱼会员名（必填）
}

// Token 闲鱼店铺token传递的是json串
type Token struct {
	AppId     int64  `json:"app_id"`
	AppSecret string `json:"app_secret"`
	Username  string `json:"username"`
}

// Product 上架商品结构体
type Product struct {
	AppId              int64    `json:"appId"`                // 应用 id
	AppSecret          string   `json:"appSecret"`            // 应用密钥[选填，有些平台需要]
	Token              string   `json:"token"`                // token[选填，有些平台需要]
	ProductID          int64    `json:"product_id"`           // 商品 ID
	UserName           []string `json:"user_name"`            // 会员名
	SpecifyPublishTime string   `json:"specify_publish_time"` // 指定发布时间
	NotifyURL          string   `json:"notify_url"`           // 回调地址
}

// NewXianYu 创建闲鱼平台
func NewXianYu() *XianYu {
	return &XianYu{}
}

func (xianYu *XianYu) AddGoodsTask(taskHeader _type.TaskHeader, taskMsg _type.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}
	//TODO
	// 构建参数
	var goodsAdd GoodsAdd
	xianYuDll, err := xianYuDll.InitXianYuDll()
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("初始化拼多多DLL失败 %v", err))
	}

	// 解析应用 id与应用秘钥
	var token Token
	unmarshalErr := json.Unmarshal([]byte(taskHeader.ShopMsg.Token), &token)
	if unmarshalErr != nil {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("解析应用id与应用秘钥 taskHeader.ShopMsg.Token = %v %w", taskHeader.ShopMsg.Token, unmarshalErr))
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
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("获取省、市、区信息失败: %v", getProvinceCityDistrictErr))
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
	goodsAdd.CatIds = string(taskMsg.BookInfo.CatIdObject.XianYuCatId)
	if goodsAdd.CatIds == "" {
		//如果类目ID为空，则使用默认类目ID（文学/小说）
		goodsAdd.CatIds = "c3c6e8d1d63c0618b108d382c4e6ea42"
	}

	// 构建详情图
	contentImgs := tool.BuildDetailGallery(taskHeader.ShopMsg.WatermarkImgUrl, taskHeader.ShopMsg.GoodsDetailFirstImgUrlArray, taskHeader.ShopMsg.GoodsDetailLastImgUrlArray, taskMsg.BookInfo.ImageObject.DetailUrlObject)

	// 构建主图（轮播图）
	mainImgs := tool.BuildCarouselGallery(taskHeader.ShopMsg.WatermarkImgUrl, taskHeader.ShopMsg.CarouseLastImgUrlArray, taskMsg.BookInfo.ImageObject.CarouselUrlArray)
	if len(mainImgs) == 0 && taskHeader.ImgType == 3 && taskMsg.BookInfo.ImageObject.DefaultImageUrl != "" {
		mainImgs = append(mainImgs, taskMsg.BookInfo.ImageObject.DefaultImageUrl)
	}
	if len(taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl) == 0 && len(mainImgs) > 0 {
		taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl = []string{mainImgs[0]}
	}
	if len(mainImgs) == 0 {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("缺少构造轮播图图片-未提交 isbn %v", taskMsg.BookInfo.Isbn))
	}

	//构建商品名称
	title := tool.BuildGoodsName(
		taskHeader.ShopMsg.GoodsNamePrefix, // 商品名称前缀
		taskHeader.ShopMsg.GoodsNameSuffix, // 商品名称后缀
		taskHeader.ShopMsg.TitleConsistOf,  // 标题组成
		taskHeader.ShopMsg.SpaceCharacter,  // 间隔符
		taskMsg.BookInfo) // 图书信息
	taskMsg.Detail.GoodsName = title

	// 构建商品信息
	content := taskMsg.BookInfo.BookName + " " + taskMsg.BookInfo.Isbn + " " + taskMsg.BookInfo.Author + " " + taskMsg.BookInfo.Publishing
	content = content + "\n" + taskHeader.ShopMsg.ShopContext

	// 店铺信息
	goodsAdd.Shop = []ShopInfo{
		{
			UserName:    token.Username,
			Province:    provinceCode,
			City:        cityCode,
			District:    districtCode,
			Title:       title,
			Content:     content,
			MainImgs:    mainImgs,
			ContentImgs: contentImgs,
		},
	}

	// 成色
	goodsAdd.StuffStatus = taskMsg.Detail.Condition
	if goodsAdd.StuffStatus == 0 {
		goodsAdd.StuffStatus = 90
	}

	//库存
	if taskMsg.Detail.Stock == 0 && taskHeader.TaskType == 1 {
		//如果库存为0 则给默认库存
		taskMsg.Detail.Stock = taskHeader.ShopMsg.DefStock
	}

	//生成一个2秒的延迟
	url := "http://127.0.0.1:8095"
	tool.HttpGetRequest(url)

	//构建参考价格
	price := tool.BuildPrice(taskHeader.PriceMod, taskMsg.Detail.Price)
	if price == 0 {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("不在价格区间内 isbn:%v", taskMsg.BookInfo.Isbn))
	}
	taskMsg.Detail.Price = price

	//构建售价
	taskMsgBookInfoPrice := tool.BuildGoodsPrice(price)

	// 图书类商品信息
	goodsAdd.BookData = []BookInfo{
		{
			ISBN:        taskMsg.BookInfo.Isbn,
			Title:       title,
			Author:      taskMsg.BookInfo.Author,
			Publisher:   taskMsg.Publishing.Value,
			ItemBizType: 2,
			SpBizType:   24,
			Prices:      []int64{price, taskMsgBookInfoPrice},
			Stock:       taskMsg.Detail.Stock,
			CatIds:      string(taskMsg.BookInfo.CatIdObject.XianYuCatId),
		},
	}

	// 闲鱼批次商品 KEY
	goodsAdd.ItemKey = strconv.FormatInt(time.Now().Unix(), 10)

	// 新增商品
	goodsAddRet, _, err := AddGoods(xianYuDll, logUuid, goodsAdd)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeAdd, fmt.Errorf("商品提交 %v", err))
	}

	// 上架商品
	launchGoodsInfo := Product{
		AppId:              token.AppId,
		AppSecret:          token.AppSecret,
		Token:              "",
		NotifyURL:          "",
		ProductID:          goodsAddRet.Data.Success[0].ProductID,
		SpecifyPublishTime: "",
		UserName:           []string{token.Username},
	}
	_, _, err = LaunchGoods(xianYuDll, logUuid, launchGoodsInfo)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, _type.GoodsTypeLaunch, fmt.Errorf("商品提交 %v", err))
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
	taskMsg.Detail.Img = mainImgs[0]
	return tool.GoodsAddReturnSuccess(taskMsg)
}
func (xianYu *XianYu) SetGoodsTask() string {
	return "闲鱼商品修改任务"

}

func (xianYu *XianYu) GetGoodsTask() string {
	return "闲鱼商品获取任务"
}

func (xianYu *XianYu) DelGoodsTask() string {
	return "闲鱼商品删除任务"
}

// 获取省市区 信息
func getProvinceCityDistrict(types int64, id int) (int, int, int, error) {
	if types == 0 { // 直接指定区域的省市区
		//根据区id 获取省、市、区code
		provinceCode, cityCode, districtCode, getRegionIdErr := _myRedis.GetRegionId(golabl.RedisClientC, strconv.Itoa(id))
		if getRegionIdErr != nil {
			return 0, 0, 0, getRegionIdErr
		}
		return provinceCode, cityCode, districtCode, nil
	} else if types == 1 { // 返回指定省下的随机区
		region, getRandomDistrictInProvinceErr := _myRedis.GetRandomDistrictInProvince(golabl.RedisClientC, id)
		if getRandomDistrictInProvinceErr != nil {
			return 0, 0, 0, getRandomDistrictInProvinceErr
		}
		//根据区id 获取省、市、区code
		provinceCode, cityCode, districtCode, getRegionIdErr := _myRedis.GetRegionId(golabl.RedisClientC, region["id"])
		if getRegionIdErr != nil {
			return 0, 0, 0, getRegionIdErr
		}
		return provinceCode, cityCode, districtCode, nil

	} else if types == 2 { //在全国返回随机省下的随机区
		region, getRandomDistrictErr := _myRedis.GetRandomDistrict(golabl.RedisClientC)
		if getRandomDistrictErr != nil {
			return 0, 0, 0, getRandomDistrictErr
		}
		//根据区id 获取省、市、区code
		provinceCode, cityCode, districtCode, getRegionIdErr := _myRedis.GetRegionId(golabl.RedisClientC, region["id"])
		if getRegionIdErr != nil {
			return 0, 0, 0, getRegionIdErr
		}
		return provinceCode, cityCode, districtCode, nil
	}
	return 0, 0, 0, fmt.Errorf("参数错误")
}

// AddGoods 商品新增
// @param xianYuDLL xianYuDLL对象
// @param token 授权令牌
// @param logUuid 日志ID
// @param goodsInfo 添加商品信息
// @return XianYuAddGoodsResponse 商品新增结果
// @return string 添加商品结果json
// @return error 错误信息
func AddGoods(xianYuDLL *xianYuDll.XianYuDLL, logUuid string, goodsInfo GoodsAdd) (_type.XianYuAddGoodsResponse, string, error) {
	var goodsAdd _type.XianYuAddGoodsResponse
	goodsInfoStr, marshalErr := json.Marshal(goodsInfo)
	if marshalErr != nil {
		return goodsAdd, "", marshalErr
	}
	goodsAddStr, xianYuGoodsAddErr := xianYuDLL.XianYuGoodsAdd(string(goodsInfoStr), golabl.MainConfig.FileUrl.XianYuDll)
	if xianYuGoodsAddErr != nil {
		return goodsAdd, "", xianYuGoodsAddErr
	}
	unmarshalErr := json.Unmarshal([]byte(goodsAddStr), &goodsAdd)
	if unmarshalErr != nil {
		return goodsAdd, "", unmarshalErr
	}
	if goodsAdd.Code != 0 {
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
		logs.LoggingMiddleware(logs.LOG_LEVEL_INFO, addGoodsReqMsg)
		return goodsAdd, goodsAddStr, errors.New("闲鱼 XianYuGoodsAdd 错误:" + goodsAddStr)
	}
	return goodsAdd, goodsAddStr, nil
}

// LaunchGoods 商品上架
func LaunchGoods(xianYuDLL *xianYuDll.XianYuDLL, logUuid string, launchGoodsInfo Product) (_type.XianYuAddGoodsResponse, string, error) {
	var launchGoods _type.XianYuAddGoodsResponse
	launchGoodsInfoStr, marshalErr := json.Marshal(launchGoodsInfo)
	if marshalErr != nil {
		return launchGoods, "", marshalErr
	}
	launchGoodsStr, xianYuLaunchGoodsAddErr := xianYuDLL.XianYuLaunchGoods(string(launchGoodsInfoStr), golabl.MainConfig.FileUrl.XianYuDll)
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
		logs.LoggingMiddleware(logs.LOG_LEVEL_INFO, addGoodsReqMsg)
		return launchGoods, launchGoodsStr, errors.New("闲鱼 XianYuLaunchGoods 错误:" + launchGoodsStr)
	}
	return launchGoods, launchGoodsStr, nil
}
