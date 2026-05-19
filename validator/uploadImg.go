package validator

import (
	"fmt"
	"net/http"
	"planA/initialization/golabl"
	taskValidator "planA/type/validator"
)

// ImgUploadToPddValidator 上传图片到拼多多验证
func ImgUploadToPddValidator(data *http.Request) (taskValidator.ImgUploadToPdd, error) {
	form := taskValidator.ImgUploadToPdd{
		ImgUrl: data.FormValue("img_url"),
		ShopId: data.FormValue("shop_id"),
	}
	fieldCN := map[string]string{"ImgUrl": "图片url", "ShopId": "店铺ID"}
	if err := golabl.Validator.Struct(form); err != nil {
		errMsg := ValidatorRule(err, fieldCN)
		return form, fmt.Errorf("参数错误：%s", errMsg)
	}
	return form, nil
}
