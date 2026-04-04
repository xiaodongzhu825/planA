package tool

import (
	"encoding/json"
	"planA/planB/initialization/golabl"
	planBTypePinduoduo "planA/planB/type/pinduoduo"
	"planA/tool"
	"strconv"
)

// GetPddGoodsList 获取商品列表
// @param params 查询参数
// @return planBTypePinduoduo.GoodsListResponse 商品列表
// @return error 错误信息
func GetPddGoodsList(params map[string]string) (planBTypePinduoduo.GoodsListResponse, string, error) {
	var goodsListt planBTypePinduoduo.GoodsListResponse
	url := golabl.Config.FileUrl.PddGetGoodsUrl
	withParams, buildURLWithParamsErr := BuildURLWithParams(url, params)
	if buildURLWithParamsErr != nil {
		return goodsListt, "", buildURLWithParamsErr
	}
	_, resStr, httpGetRequestErr := HttpGetRequest(withParams)
	if httpGetRequestErr != nil {
		return goodsListt, resStr, httpGetRequestErr
	}
	unmarshalErr := json.Unmarshal([]byte(resStr), &goodsListt)
	if unmarshalErr != nil {
		return goodsListt, resStr, unmarshalErr
	}
	return goodsListt, resStr, nil
}

// WritePddGoodsData 写入商品数据
// @param goodsListStr 商品列表
// @return error 错误信息
func WritePddGoodsData(goodsListStr []planBTypePinduoduo.GoodsItem, page int, pageTotal int64) (planBTypePinduoduo.AsyncTaskResponse, string, error) {
	var ret planBTypePinduoduo.AsyncTaskResponse
	marshal, marshalErr := json.Marshal(goodsListStr)
	if marshalErr != nil {
		return ret, "", marshalErr
	}
	params := map[string]string{
		"taskId":       golabl.Task.TaskId,
		"shopId":       strconv.FormatInt(golabl.Task.Header.ShopId, 10),
		"goodsListStr": string(marshal),
		"allNum":       strconv.FormatInt(pageTotal, 10),
		"num":          strconv.Itoa(page),
	}
	retStr, submitFormDataErr := tool.SubmitFormData(golabl.Config.FileUrl.PddAddGoodsUrl, params)
	if submitFormDataErr != nil {
		return ret, retStr, submitFormDataErr
	}
	unmarshalErr := json.Unmarshal([]byte(retStr), &ret)
	if unmarshalErr != nil {
		return ret, retStr, unmarshalErr
	}
	return ret, retStr, nil
}

// GetPddGoodsDetail 获取商品详情
func GetPddGoodsDetail(goodsListStr []planBTypePinduoduo.GoodsItem) ([]planBTypePinduoduo.GoodsItem, string, error) {
	var ret []planBTypePinduoduo.GoodsItem
	marshal, marshalErr := json.Marshal(goodsListStr)
	if marshalErr != nil {
		return ret, "", marshalErr
	}
	params := map[string]string{
		"accessToken":          golabl.Task.Header.ShopMsg.Token,
		"goodsListGetResponse": string(marshal),
	}
	retStr, submitFormDataErr := tool.SubmitFormData(golabl.Config.FileUrl.PddGetGoodsDetailUrl, params)
	if submitFormDataErr != nil {
		return ret, retStr, submitFormDataErr
	}
	unmarshalErr := json.Unmarshal([]byte(retStr), &ret)
	if unmarshalErr != nil {
		return ret, retStr, unmarshalErr
	}
	return ret, retStr, nil
}
