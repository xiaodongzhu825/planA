package xianyu

import (
	"encoding/json"
	"errors"
	"fmt"
	"planA/planB/initialization/golabl"
	"planA/planB/modules/logs"
	xianYuDll "planA/planB/modules/xianYu"
	"planA/planB/service"
	"planA/planB/tool"
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

func (xianYu *XianYu) AddGoodsTask(taskMsg planAType.TaskBody) (string, error) {
	//生成唯一请求标识（用于出错精准查询日志）
	logUuid, generateUUIDErr := tool.GenerateUUID()
	if generateUUIDErr != nil {
		return "", fmt.Errorf("生成唯一请求标识失败: %v", generateUUIDErr)
	}
	//TODO
	// 构建参数
	var goodsAdd planBTypeXianyu.GoodsAdd
	xianYuDlls, err := xianYuDll.InitXianYuDll()
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("初始化拼多多DLL失败 %v", err))
	}

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
	goodsAdd.CatIds = string(taskMsg.BookInfo.CatIdObject.XianYuCatId)
	if goodsAdd.CatIds == "" {
		//如果类目ID为空，则使用默认类目ID（文学/小说）
		goodsAdd.CatIds = "c3c6e8d1d63c0618b108d382c4e6ea42"
	}
	// 构建详情图
	contentImgs := tool.BuildDetailGallery(golabl.Task.Header.ShopMsg.WatermarkImgUrl, golabl.Task.Header.ShopMsg.GoodsDetailFirstImgUrlArray, golabl.Task.Header.ShopMsg.GoodsDetailLastImgUrlArray, taskMsg.BookInfo.ImageObject.DetailUrlObject)

	// 构建主图（轮播图）
	mainImgs := tool.BuildCarouselGallery(golabl.Task.Header.ShopMsg.WatermarkImgUrl, golabl.Task.Header.ShopMsg.CarouseLastImgUrlArray, taskMsg.BookInfo.ImageObject.CarouselUrlArray)
	if len(mainImgs) == 0 && golabl.Task.Header.ImgType == 3 && taskMsg.BookInfo.ImageObject.DefaultImageUrl != "" {
		mainImgs = append(mainImgs, taskMsg.BookInfo.ImageObject.DefaultImageUrl)
	}
	if len(taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl) == 0 && len(mainImgs) > 0 {
		taskMsg.BookInfo.ImageObject.DetailUrlObject.LiveShootingUrl = []string{mainImgs[0]}
	}
	if len(mainImgs) == 0 {
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
	if taskMsg.Detail.Stock == 0 && golabl.Task.Header.TaskType == 1 {
		//如果库存为0 则给默认库存
		taskMsg.Detail.Stock = golabl.Task.Header.ShopMsg.DefStock
	}

	//生成一个2秒的延迟
	url := "http://127.0.0.1:8095"
	tool.HttpGetRequest(url)

	//构建参考价格
	price := tool.BuildPrice(golabl.Task.Header.PriceMod, taskMsg.Detail.Price)
	if price == 0 {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("不在价格区间内 isbn:%v", taskMsg.BookInfo.Isbn))
	}
	taskMsg.Detail.Price = price

	//构建售价
	taskMsgBookInfoPrice := tool.BuildGoodsPrice(price)

	// 图书类商品信息
	goodsAdd.BookData = []planBTypeXianyu.BookInfo{
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
	goodsAddRet, _, err := addGoods(xianYuDlls, logUuid, goodsAdd)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("商品提交 %v", err))
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
	_, _, err = launchGoods(xianYuDlls, logUuid, launchGoodsInfo)
	if err != nil {
		return tool.ReturnErr(logUuid, taskMsg, golabl.TaskType, fmt.Errorf("商品提交 %v", err))
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
// @param xianYuDLL xianYuDLL对象
// @param token 授权令牌
// @param logUuid 日志ID
// @param goodsInfo 添加商品信息
// @return XianYuAddGoodsResponse 商品新增结果
// @return string 添加商品结果json
// @return error 错误信息
func addGoods(xianYuDLL *xianYuDll.XianYuDLL, logUuid string, goodsInfo planBTypeXianyu.GoodsAdd) (planBTypeXianyu.XianYuAddGoodsResponse, string, error) {
	var goodsAdd planBTypeXianyu.XianYuAddGoodsResponse
	goodsInfoStr, marshalErr := json.Marshal(goodsInfo)
	if marshalErr != nil {
		return goodsAdd, "", marshalErr
	}
	goodsAddStr, xianYuGoodsAddErr := xianYuDLL.XianYuGoodsAdd(string(goodsInfoStr), golabl.Config.FileUrl.XianYuDll)
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

// 商品上架
func launchGoods(xianYuDLL *xianYuDll.XianYuDLL, logUuid string, launchGoodsInfo planBTypeXianyu.Product) (planBTypeXianyu.XianYuAddGoodsResponse, string, error) {
	var launchGoods planBTypeXianyu.XianYuAddGoodsResponse
	launchGoodsInfoStr, marshalErr := json.Marshal(launchGoodsInfo)
	if marshalErr != nil {
		return launchGoods, "", marshalErr
	}
	launchGoodsStr, xianYuLaunchGoodsAddErr := xianYuDLL.XianYuLaunchGoods(string(launchGoodsInfoStr), golabl.Config.FileUrl.XianYuDll)
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
