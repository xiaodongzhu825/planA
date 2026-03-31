package tool

import (
	"encoding/json"
	"fmt"
	"planA/planB/initialization/golabl"
	planBTypePinduoduo "planA/planB/type/pinduoduo"
	"planA/tool"
	"strconv"
)

// GetPddGoodsList 获取商品列表
// @param params 查询参数
// @return planBTypePinduoduo.GoodsListResponse 商品列表
// @return error 错误信息
func GetPddGoodsList(params map[string]string) (planBTypePinduoduo.GoodsListResponse, error) {
	var goodsListt planBTypePinduoduo.GoodsListResponse
	url := golabl.Config.FileUrl.PddGetGoodsUrl
	// 如果是详情拉取，则使用详情拉取接口
	if golabl.Task.Header.TaskType == 4 {
		url = golabl.Config.FileUrl.PddGetGoodsDetailUrl
	}
	withParams, buildURLWithParamsErr := BuildURLWithParams(url, params)
	if buildURLWithParamsErr != nil {
		return goodsListt, buildURLWithParamsErr
	}
	_, resStr, httpGetRequestErr := HttpGetRequest(withParams)

	if httpGetRequestErr != nil {
		return goodsListt, httpGetRequestErr
	}
	unmarshalErr := json.Unmarshal([]byte(resStr), &goodsListt)
	if unmarshalErr != nil {
		return goodsListt, unmarshalErr
	}
	return goodsListt, nil
}

// WritePddGoodsData 写入商品数据
// @param goodsListStr 商品列表
// @return error 错误信息
func WritePddGoodsData(goodsListStr []planBTypePinduoduo.GoodsItem, page int, pageTotal int64) (planBTypePinduoduo.AsyncTaskResponse, error) {
	var ret planBTypePinduoduo.AsyncTaskResponse
	marshal, marshalErr := json.Marshal(goodsListStr)
	if marshalErr != nil {
		return ret, marshalErr
	}
	params := map[string]string{
		"taskId":       golabl.Task.TaskId,
		"shopId":       strconv.FormatInt(golabl.Task.Header.ShopId, 10),
		"goodsListStr": string(marshal),
		"allNum":       strconv.FormatInt(pageTotal, 10),
		"num":          strconv.Itoa(page),
	}
	retStr, submitFormDataErr := tool.SubmitFormData(golabl.Config.FileUrl.PddAddGoodsUrl, params)
	fmt.Println("--------------------retStr----------------------")
	fmt.Println(retStr)
	fmt.Println("--------------------retStr----------------------")
	if submitFormDataErr != nil {
		return ret, submitFormDataErr
	}
	unmarshalErr := json.Unmarshal([]byte(retStr), &ret)
	if unmarshalErr != nil {
		return ret, unmarshalErr
	}
	return ret, nil
}
