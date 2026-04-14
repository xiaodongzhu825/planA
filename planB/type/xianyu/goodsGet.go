package xianyu

// GoodsListReq 获取列表请求结构体
type GoodsListReq struct {
	AppId         int64   `json:"appId"`       // 应用 id
	AppSecret     string  `json:"appSecret"`   // 应用密钥[选填，有些平台需要]
	UpdateTime    []int64 `json:"update_time"` // 可传空
	ProductStatus int     `json:"product_status"`
	PageNo        int     `json:"page_no"`   // 页码
	PageSize      int     `json:"page_size"` // 页大小
}

// GoodsListRet 获取列表返回结构体
type GoodsListRet struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data GoodsListRetData `json:"data"`
}

// GoodsListRetData 结构
type GoodsListRetData struct {
	List     []GoodsListRetProduct `json:"list"`
	Count    int                   `json:"count"`
	PageNo   int                   `json:"page_no"`
	PageSize int                   `json:"page_size"`
}

// GoodsListRetProduct 商品结构
type GoodsListRetProduct struct {
	ProductID          int64  `json:"product_id"`
	ProductStatus      int    `json:"product_status"`
	ItemBizType        int    `json:"item_biz_type"`
	SpBizType          int    `json:"sp_biz_type"`
	ChannelCatID       string `json:"channel_cat_id"`
	OriginalPrice      int    `json:"original_price"`
	Price              int    `json:"price"`
	Stock              int    `json:"stock"`
	Sold               int    `json:"sold"`
	Title              string `json:"title"`
	DistrictID         int    `json:"district_id"`
	OuterID            string `json:"outer_id"`
	StuffStatus        int    `json:"stuff_status"`
	ExpressFee         int    `json:"express_fee"`
	SpecType           int    `json:"spec_type"`
	Source             int    `json:"source"`
	SpecifyPublishTime int64  `json:"specify_publish_time"`
	OnlineTime         int64  `json:"online_time"`
	OfflineTime        int64  `json:"offline_time"`
	SoldTime           int64  `json:"sold_time"`
	UpdateTime         int64  `json:"update_time"`
	CreateTime         int64  `json:"create_time"`
}

// GoodsDetailReq 请求详商品情结构体
type GoodsDetailReq struct {
	AppId     int64  `json:"appId"`      // 应用 id
	AppSecret string `json:"appSecret"`  // 应用密钥[选填，有些平台需要]
	ProductId int64  `json:"product_id"` // 管家商品id
}

// GoodDetailRet 获取列表返回结构体
type GoodDetailRet struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data GoodsDetailRet `json:"data"`
}

// GoodsDetailRet 获取商品详情返回结构体
type GoodsDetailRet struct {
	ProductID          int64         `json:"product_id"`
	ProductStatus      int           `json:"product_status"`
	PublishStatus      int           `json:"publish_status"`
	ItemBizType        int           `json:"item_biz_type"`
	SpBizType          int           `json:"sp_biz_type"`
	FlashSaleType      int           `json:"flash_sale_type"`
	ChannelCatID       string        `json:"channel_cat_id"`
	Title              string        `json:"title"`
	Price              int           `json:"price"`
	OriginalPrice      int           `json:"original_price"`
	ExpressFee         int           `json:"express_fee"`
	Stock              int64         `json:"stock"`
	Sold               int           `json:"sold"`
	OuterID            string        `json:"outer_id"`
	StuffStatus        int           `json:"stuff_status"`
	SpecifyPublishTime string        `json:"specify_publish_time"`
	PublishShop        []PublishShop `json:"publish_shop"`
	SpecType           int           `json:"spec_type"`
	BookData           BookData      `json:"book_data"`
	OnlineTime         int64         `json:"online_time"`
	OfflineTime        int64         `json:"offline_time"`
	SoldTime           int64         `json:"sold_time"`
	UpdateTime         int64         `json:"update_time"`
	CreateTime         int64         `json:"create_time"`
	IsTaxIncluded      bool          `json:"is_tax_included"`
}

type PublishShop struct {
	UserName       string   `json:"user_name"`
	ItemID         int64    `json:"item_id"`
	Province       int      `json:"province"`
	City           int      `json:"city"`
	District       int      `json:"district"`
	Title          string   `json:"title"`
	Content        string   `json:"content"`
	Images         []string `json:"images"`
	Status         int      `json:"status"`
	WhiteImages    string   `json:"white_images"`
	ServiceSupport string   `json:"service_support"`
}

type BookData struct {
	ISBN      string `json:"isbn"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Publisher string `json:"publisher"`
}
