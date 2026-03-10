package _type

// 店铺结构体

// ShopInfo 完整的店铺信息
type ShopInfo struct {
	Shop          *Shop          `json:"shop"`
	ShopDetail    *ShopDetail    `json:"shop_detail"`
	ShopContext   *ShopContext   `json:"shop_context"`
	Spec          *Spec          `json:"spec"`
	PriceTemplate *PriceTemplate `json:"price_template"`
}

// Shop 店铺基本信息
type Shop struct {
	ID             int64       `json:"id"`
	ShopKey        string      `json:"shop_key"`
	ShopName       string      `json:"shop_name"`
	ShopAliasName  string      `json:"shop_alias_name"`
	ShopType       string      `json:"shop_type"`
	ShopAuthorize  string      `json:"shop_authorize"`
	Status         string      `json:"status"`
	MallID         int64       `json:"mall_id"`
	Token          string      `json:"token"`
	RefreshToken   interface{} `json:"refresh_token"`
	CreateTime     string      `json:"create_time"`
	UpdateTime     string      `json:"update_time"`
	ExpirationTime string      `json:"expiration_time"`
	PublishType    string      `json:"publish_type"`
	SkuSpec        string      `json:"sku_spec"`
	StartUpdatedAt int64       `json:"start_updated_at"`
	TenantID       string      `json:"tenant_id"`
	CreateBy       int64       `json:"create_by" db:"create_by"` // 创建人
}

// ShopDetail 店铺详情
type ShopDetail struct {
	ID                  int64  `json:"id"`
	ShopID              int64  `json:"shop_id"`               //店铺 ID
	SaleTemplateID      int64  `json:"sale_template_id"`      //运费模版 ID
	LowPrice            int    `json:"low_price"`             //最低价格
	HighPrice           int    `json:"high_price"`            //最高价格
	StockDeff           int    `json:"stock_deff"`            //库存
	TemplateId          int    `json:"template_id"`           //物流运费模版 ID
	TitlePrefix         string `json:"title_prefix"`          //标题前缀
	TitleSuffix         string `json:"title_suffix"`          //标题后缀
	TitleConsistOf      string `json:"title_consist_of"`      //标题包含信息
	SpaceCharacter      string `json:"space_character"`       //是否使用空格
	SevenDays           string `json:"seven_days"`            //是否支持7天无理由退换货
	Presale             string `json:"presale"`               //是否预售
	Fake                string `json:"fake"`                  //是否支持假一赔十，false-不支持，true-支持
	IsPreSale           bool   `json:"is_pre_sale"`           //是否预售,true-预售商品，false-非预售商品
	IsRefundable        bool   `json:"is_refundable"`         //是否7天无理由退换货，true-支持，false-不支持
	IsSecondHand        string `json:"is_second_hand"`        //是否二手 1 -二手商品 ，0-全新商品
	ShipmentLimitSecond string `json:"shipment_limit_second"` //承诺发货时间（秒）
	CostTemplateId      int64  `json:"cost_template_id"`      //物流运费模板 ID
	TowDiscount         int64  `json:"two_discount"`          //两件折扣
	WatermarkImgUrl     string `json:"watermark_img_url"`     //水印图片链接
	DistrictId          int64  `json:"district_id"`           //地区类型 0 指定区县 1 指定省 2 全国
	DistrictType        string `json:"district_type"`         //地区 ID 【district_type=0 区县ID district_type=1 省ID district_type=2 全国（空值）】
}

// ShopContext 店铺上下文
type ShopContext struct {
	ID         int64  `json:"id"`
	ShopID     int64  `json:"shop_id"`
	Context    string `json:"context"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

// Spec 规格信息
type Spec struct {
	ID              int64  `json:"id"`
	ShopID          int64  `json:"shop_id"`
	SpecName        string `json:"spec_name"`
	SpecTypeID      string `json:"spec_type_id"`
	SpecTypeName    string `json:"spec_type_name"`
	SpecCompose     string `json:"spec_compose"`
	SpecCodeCompose string `json:"spec_code_compose"`
	CreateTime      string `json:"create_time"`
	UpdateTime      string `json:"update_time"`
}

// PriceTemplate 价格模板
type PriceTemplate struct {
	AddAmount    int    `json:"add_amount"`
	DelFlag      string `json:"del_flag"`
	HighPrice    int64  `json:"high_price"`
	ID           int64  `json:"id"`
	LowPrice     int64  `json:"low_price"`
	PriceType    string `json:"price_type"`
	Proportion   int    `json:"proportion"`
	RangePrice   string `json:"range_price"` // 解析后的价格区间
	Status       string `json:"status"`
	TemplateName string `json:"template_name"`
}
