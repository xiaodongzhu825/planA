package taobao

// TbGoodsListResponse 淘宝商品列表响应
type TbGoodsListResponse struct {
	Items []TbGoodsItemWrap `json:"items"` // 商品列表
}

// TbGoodsItemWrap 淘宝商品包装结构（items 数组中每项为 {"item": {...}}）
type TbGoodsItemWrap struct {
	Item TbGoodsItem `json:"item"`
}

// TbGoodsItem 淘宝商品信息
type TbGoodsItem struct {
	NumIid        string `json:"numIid"`        // 商品数字ID
	Title         string `json:"title"`         // 商品标题
	Price         string `json:"price"`         // 商品价格（字符串，如 "99.00"）
	PicUrl        string `json:"picUrl"`        // 主图URL
	Desc          string `json:"desc"`          // 商品描述HTML
	Num           int64  `json:"num"`           // 库存数量
	ApproveStatus string `json:"approveStatus"` // 商品状态: onsale/instock
	Modified      string `json:"modified"`      // 修改时间（字符串）
	OuterId       string `json:"outerId"`       // 商家编码（ISBN等）
	Nick          string `json:"nick"`          // 掌柜昵称
	Type          string `json:"type"`          // 商品类型
	Cid           int64  `json:"cid"`           // 类目ID
	BigImg        string `json:"bigImg"`        // 大图URL（兼容拼多多字段命名）
}

// TbSkuProperty 淘宝SKU属性（兼容字段）
type TbSkuProperty struct {
	Punit  string `json:"punit"`
	RefPid int    `json:"ref_pid"`
	Value  string `json:"value"`
	Vid    int    `json:"vid"`
}
