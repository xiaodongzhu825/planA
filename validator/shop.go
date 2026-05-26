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
