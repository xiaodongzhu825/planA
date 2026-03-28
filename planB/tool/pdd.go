package tool

import (
	"encoding/json"
	"planA/planB/initialization/golabl"
	planBTypePinduoduo "planA/planB/type/pinduoduo"
)

// GetGoodsList 获取商品列表
// @param params 查询参数
// @return planBTypePinduoduo.GoodsListResponse 商品列表
// @return error 错误信息
func GetGoodsList(params map[string]string) (planBTypePinduoduo.GoodsListResponse, error) {
	var goodsListt planBTypePinduoduo.GoodsListResponse
	withParams, buildURLWithParamsErr := BuildURLWithParams(golabl.Config.FileUrl.PddGetGoodsUrl, params)
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
