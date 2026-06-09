package validator

// GetShopInfo 获取店铺信息验证结构体
type GetShopInfo struct {
	ShopId string `form:"shop_id" validate:"required"` //必填
}

// CreateTbShop 创建淘宝店铺数据
type CreateTbShop struct {
	ShopId   string `form:"shop_id" validate:"required"`
	ShopName string `form:"shop_name" validate:"required"`
}
