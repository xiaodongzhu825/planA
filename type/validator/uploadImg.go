package validator

// ImgUploadToPdd 上传图片到拼多多结构体
type ImgUploadToPdd struct {
	ImgUrl string `form:"img_url" validate:"required"` // 必填
	ShopId string `form:"shop_id" validate:"required"` // 必填
}
