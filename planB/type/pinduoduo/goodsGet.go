package pinduoduo

// GoodsQueryParams 商品查询参数
type GoodsQueryParams struct {
	Page     int `json:"page" form:"page"`           // 返回页码，默认1
	PageSize int `json:"page_size" form:"page_size"` // 返回数量，默认100，最大100
}

// GoodsListResponse 响应结构体
type GoodsListResponse struct {
	GoodsList  []GoodsItem `json:"goods_list"`
	TotalCount int         `json:"total_count"`
}

type GoodsItem struct {
	CreatedAt            int64             `json:"created_at"`
	GoodsId              int64             `json:"goods_id"`
	GoodsName            string            `json:"goods_name"`
	GoodsQuantity        int               `json:"goods_quantity"`
	GoodsReserveQuantity int               `json:"goods_reserve_quantity"`
	ImageUrl             string            `json:"image_url"`
	IsMoreSku            int               `json:"is_more_sku"`
	IsOnsale             int               `json:"is_onsale"`
	SkuList              []SkuItem         `json:"sku_list"`
	ThumbUrl             string            `json:"thumb_url"`
	Price                int64             `json:"price"`
	SkuCode              string            `json:"skuCode"`
	SkuPropertyList      []SkuPropertyList `json:"skuPropertyList"`
	BigImg               string            `json:"bigImg"`
}
type SkuPropertyList struct {
	Punit  string `json:"punit"`
	RefPid int    `json:"ref_pid"`
	Value  string `json:"value"`
	Vid    int    `json:"vid"`
}

type SkuItem struct {
	IsSkuOnsale     int    `json:"is_sku_onsale"`
	OuterGoodsId    string `json:"outer_goods_id"`
	OuterId         string `json:"outer_id"`
	ReserveQuantity int    `json:"reserve_quantity"`
	SkuId           int64  `json:"sku_id"`
	SkuQuantity     int    `json:"sku_quantity"`
	Spec            string `json:"spec"`
}
