package controller

import (
	"net/http"
	"planA/service"
	"planA/tool"
	toolPdd "planA/tool/pdd"
	"planA/validator"
)

// GetShopInfo 查询店铺信息
func GetShopInfo(httpMsg http.ResponseWriter, data *http.Request) {
	// 验证表单
	dataVal, createTaskValidatorErr := validator.GetShopInfoValidator(data)
	if createTaskValidatorErr != nil {
		tool.Error(httpMsg, createTaskValidatorErr.Error(), http.StatusInternalServerError)
		return
	}

	// 查询店铺数据
	shopDataStr, err := service.GetTaskShop(dataVal.ShopId)
	if err != nil {
		errMsg := "获取店铺数据失败: shopId " + dataVal.ShopId + " " + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	// 解析 json数据
	shopData, err := toolPdd.ParseShopData(shopDataStr)
	if err != nil {
		errMsg := "解析店铺数据失败:" + err.Error()
		tool.Error(httpMsg, errMsg, http.StatusInternalServerError)
		return
	}
	tool.Success(httpMsg, shopData)

}
