package tool

import (
	"encoding/json"
	planBTypePinduoduo "planA/planB/type/pinduoduo"
	"planA/planD/initialization/golabl"
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
