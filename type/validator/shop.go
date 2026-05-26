package validator

// GetShopInfo 获取店铺信息验证结构体
type GetShopInfo struct {
	ShopId string `form:"shop_id" validate:"required"` //必填
}
