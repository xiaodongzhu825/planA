package validator

import (
	"fmt"
	"net/http"
	"planA/initialization/golabl"
	taskValidator "planA/type/validator"

	"github.com/gorilla/mux"
)

// GetShopInfoValidator 获取店铺信息参数验证
func GetShopInfoValidator(data *http.Request) (taskValidator.GetShopInfo, error) {
	vars := mux.Vars(data)
	shopId := vars["shopId"]

	form := taskValidator.GetShopInfo{
		ShopId: shopId,
	}
	fieldCN := map[string]string{"ShopId": "店铺ID"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}

// CreateTbShopValidator 创建淘宝店铺数据
func CreateTbShopValidator(data *http.Request) (taskValidator.CreateTbShop, error) {
	form := taskValidator.CreateTbShop{
		ShopId:   data.FormValue("shop_id"),
		ShopName: data.FormValue("shop_name"),
	}
	fieldCN := map[string]string{"ShopId": "店铺ID", "ShopName": "店铺名称"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}
