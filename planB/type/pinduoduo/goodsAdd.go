package pinduoduo

type GoodsAdd struct {
	GoodsName           string            `json:"goods_name"`            // 商品名称
	CarouselGallery     []string          `json:"carousel_gallery"`      // 轮播图
	CatId               int64             `json:"cat_id"`                // 商品分类
	GoodsType           int64             `json:"goods_type"`            // 商品类型 1-国内普通商品，2-一般贸易，3-保税仓BBC直供，4-海外BC直邮 ,5-流量 ,6-话费 ,7-优惠券 ,8-QQ充值 ,9-加油卡，15-商家卡券，18-海外CC行邮 19-平台卡券
	MarketPrice         int64             `json:"market_price"`          // 参考价格，单位为分
	DetailGallery       []string          `json:"detail_gallery"`        // 详情图
	OutGoodsId          string            `json:"out_goods_id"`          // 商品ID
	SkuList             []Sku             `json:"sku_list"`              // SKU列表
	IsFolt              bool              `json:"is_folt"`               // 是否支持假一赔十，false-不支持，true-支持
	IsPreSale           bool              `json:"is_pre_sale"`           // 是否预售,true-预售商品，false-非预售商品
	IsRefundable        bool              `json:"is_refundable"`         // 是否7天无理由退换货，true-支持，false-不支持
	SecondHand          bool              `json:"second_hand"`           // 是否二手商品， true -二手商品 ，false-全新商品
	CostTemplateId      int64             `json:"cost_template_id"`      // 物流运费模板ID
	CountryId           int64             `json:"country_id"`            // 国家ID
	ShipmentLimitSecond int64             `json:"shipment_limit_second"` // 承诺发货时间（秒），普通、进口商品可选48小时或24小时；直邮商品（goods_type=4）只可入参120小时，直供商品（goods_type=3）只可入参96小时；is_pre_sale为true时不必传
	TwoPiecesDiscount   int64             `json:"two_pieces_discount"`   // 满2件折扣，可选范围0-100, 0表示取消，95表示95折，设置需先查询规则接口获取实际可填范围
	GoodsProperties     []GoodsProperties `json:"goods_properties"`      //商品属性列表
}
type GoodsProperties struct {
	RefPid    int64  `json:"ref_pid"`    // 属性名称
	Value     string `json:"value"`      // 属性值
	ValueUnit string `json:"value_unit"` // 属性单位
	Vid       int64  `json:"vid"`        // 属性值 id
}

type Sku struct {
	IsOnsale      int64         `json:"is_onsale"`      // sku上架状态，0-已下架，1-上架中
	LimitQuantity int64         `json:"limit_quantity"` // sku购买限制，只入参999
	MultiPrice    int64         `json:"multi_price"`    // 商品团购价格，单位为分
	Price         int64         `json:"price"`          // 商品单买价格，单位为分
	SkuProperties []SkuProperty `json:"sku_properties"` // sku属性列表
	Quantity      int64         `json:"quantity"`       // 商品sku库存初始数量，后续库存update只使用stocks.update接口进行调用
	ThumbUrl      string        `json:"thumb_url"`      // sku 缩略图
	SpecIdList    string        `json:"spec_id_list"`   // 商品规格列表，根据pdd.goods.spec.id.get生成的规格属性id，例如：颜色规格下商家新增白色和黑色，大小规格下商家新增L和XL，则由4种spec组合，入参一种组合即可，在skulist中需要有4个spec组合的sku，示例：[20,5]
	Weight        int64         `json:"weight"`         // 重量，单位为g
	OutSkuSn      string        `json:"out_sku_sn"`     // 商品 sku编号
}

type SkuProperty struct {
	Punit  string `json:"punit"`   // 属性单位
	RefPid int64  `json:"ref_pid"` // 属性id
	Value  string `json:"value"`   // 属性值
	Vid    int64  `json:"vid"`     // 属性值id
}

// PriceMod 价格处理
type PriceMod struct {
	Min         int64 `json:"min"`          // 价格区间最小值
	Max         int64 `json:"max"`          // 价格区间最大值
	MarkupRate  int64 `json:"markup_rate"`  // 加价比例
	MarkupValue int64 `json:"markup_value"` // 价格区间加价值
}

// GoodsCommitDetail 获取商品提交的商品详情
type GoodsCommitDetail struct {
	GoodsCommitId int64 `json:"goods_commit_id"` // 商品提交id
	GoodsId       int64 `json:"goods_id"`        // 商品id
}

// PddSuccessResponse 拼多多接口 PddGoodsOuterCatMappingGet 返回结构体
type PddSuccessResponse struct {
	OuterCatMappingGetResponse PddCategoryMappingResponse `json:"outer_cat_mapping_get_response"`
}
type PddCategoryMappingResponse struct {
	CatID1    int64  `json:"cat_id1"`    // 一级类目 ID
	CatID2    int64  `json:"cat_id2"`    // 二级类目 ID
	CatID3    int64  `json:"cat_id3"`    // 三级类目 ID
	CatID4    int64  `json:"cat_id4"`    // 四级类目 ID
	RequestID string `json:"request_id"` // 请求 ID
}

// GoodsAddResponseWrapper 拼多多接口 PddGoodsAdd 返回结构体
type GoodsAddResponseWrapper struct {
	Response GoodsAddData `json:"goods_add_response"`
}
type GoodsAddData struct {
	GoodsCommitID int64  `json:"goods_commit_id"`
	GoodsID       int64  `json:"goods_id"`
	MatchedSpuID  *int64 `json:"matched_spu_id"` // null值需要特殊处理
	RequestID     string `json:"request_id"`
}

// GoodsCommitDetailResponse 拼多多接口 PddGoodsCommitDetail 响应结构体
type GoodsCommitDetailResponse struct {
	GoodsCommitDetailResponse struct {
		BadFruitClaim            int           `json:"bad_fruit_claim"`
		BuyLimit                 int           `json:"buy_limit"`
		CarouselGalleryList      []string      `json:"carousel_gallery_list"`
		CatID                    int           `json:"cat_id"`
		CommitMessage            interface{}   `json:"commit_message"`
		CostTemplateID           int64         `json:"cost_template_id"`
		CountryID                int           `json:"country_id"`
		CustomerNum              int           `json:"customer_num"`
		Customs                  string        `json:"customs"`
		Deleted                  int           `json:"deleted"`
		DeliveryOneDay           interface{}   `json:"delivery_one_day"`
		DeliveryType             interface{}   `json:"delivery_type"`
		DetailGalleryList        []string      `json:"detail_gallery_list"`
		ElecGoodsAttributes      interface{}   `json:"elec_goods_attributes"`
		EndProductionDate        interface{}   `json:"end_production_date"`
		FabricContentID          interface{}   `json:"fabric_content_id"`
		FabricID                 interface{}   `json:"fabric_id"`
		GoodsCommitID            int64         `json:"goods_commit_id"`
		GoodsDesc                string        `json:"goods_desc"`
		GoodsID                  int64         `json:"goods_id"`
		GoodsName                string        `json:"goods_name"`
		GoodsPattern             int           `json:"goods_pattern"`
		GoodsPropertyList        []interface{} `json:"goods_property_list"`
		GoodsStatus              int           `json:"goods_status"`
		GoodsTradeAttr           interface{}   `json:"goods_trade_attr"`
		GoodsTravelAttr          interface{}   `json:"goods_travel_attr"`
		GoodsType                int           `json:"goods_type"`
		HdThumbURL               string        `json:"hd_thumb_url"`
		ImageURL                 string        `json:"image_url"`
		InvoiceStatus            int           `json:"invoice_status"`
		IsCustoms                int           `json:"is_customs"`
		IsFolt                   int           `json:"is_folt"`
		IsGroupPreSale           interface{}   `json:"is_group_pre_sale"`
		IsPreSale                int           `json:"is_pre_sale"`
		IsRefundable             int           `json:"is_refundable"`
		IsSkuPreSale             int           `json:"is_sku_pre_sale"`
		LackOfWeightClaim        interface{}   `json:"lack_of_weight_claim"`
		LocalServiceIDList       interface{}   `json:"local_service_id_list"`
		MaiJiaZiTi               interface{}   `json:"mai_jia_zi_ti"`
		MarketPrice              int           `json:"market_price"`
		OrderLimit               int           `json:"order_limit"`
		OriginCountryID          int           `json:"origin_country_id"`
		OutSourceGoodsID         interface{}   `json:"out_source_goods_id"`
		OutSourceType            interface{}   `json:"out_source_type"`
		OuterGoodsID             string        `json:"outer_goods_id"`
		OverseaGoods             interface{}   `json:"oversea_goods"`
		OverseaType              int           `json:"oversea_type"`
		PaperLength              interface{}   `json:"paper_length"`
		PaperNetWeight           interface{}   `json:"paper_net_weight"`
		PaperPliesNum            interface{}   `json:"paper_plies_num"`
		PaperWidth               interface{}   `json:"paper_width"`
		PreSaleTime              int           `json:"pre_sale_time"`
		PrivacyDelivery          int           `json:"privacy_delivery"`
		ProductionLicense        interface{}   `json:"production_license"`
		ProductionStandardNumber interface{}   `json:"production_standard_number"`
		QuanGuoLianBao           int           `json:"quan_guo_lian_bao"`
		RequestID                string        `json:"request_id"`
		SecondHand               int           `json:"second_hand"`
		ShangMenAnZhuang         interface{}   `json:"shang_men_an_zhuang"`
		ShelfLife                interface{}   `json:"shelf_life"`
		ShipmentLimitSecond      int           `json:"shipment_limit_second"`
		ShopGroupID              interface{}   `json:"shop_group_id"`
		ShopGroupName            interface{}   `json:"shop_group_name"`
		SizeSpecID               interface{}   `json:"size_spec_id"`
		SkuList                  []SkuInfo     `json:"sku_list"`
		SkuType                  interface{}   `json:"sku_type"`
		SongHuoAnZhuang          interface{}   `json:"song_huo_an_zhuang"`
		SongHuoRuHu              interface{}   `json:"song_huo_ru_hu"`
		StartProductionDate      interface{}   `json:"start_production_date"`
		ThumbURL                 string        `json:"thumb_url"`
		TinyName                 string        `json:"tiny_name"`
		TwoPiecesDiscount        int           `json:"two_pieces_discount"`
		VideoGallery             []VideoInfo   `json:"video_gallery"`
		Warehouse                string        `json:"warehouse"`
		WarmTips                 string        `json:"warm_tips"`
		ZhiHuanBuXiu             int           `json:"zhi_huan_bu_xiu"`
	} `json:"goods_commit_detail_response"`
}

// SkuInfo 定义SKU信息
type SkuInfo struct {
	IsOnsale        int           `json:"is_onsale"`
	Length          interface{}   `json:"length"`
	LimitQuantity   int           `json:"limit_quantity"`
	MultiPrice      int           `json:"multi_price"`
	OutSkuSn        string        `json:"out_sku_sn"`
	OutSourceSkuID  interface{}   `json:"out_source_sku_id"`
	OverseaSku      interface{}   `json:"oversea_sku"`
	Price           int           `json:"price"`
	Quantity        int           `json:"quantity"`
	ReserveQuantity int           `json:"reserve_quantity"`
	SkuID           int64         `json:"sku_id"`
	SkuPreSaleTime  int           `json:"sku_pre_sale_time"`
	SkuPropertyList []SkuProperty `json:"sku_property_list"`
	Spec            []Spec        `json:"spec"`
	ThumbURL        string        `json:"thumb_url"`
	Weight          int           `json:"weight"`
}

// Spec 定义规格信息
type Spec struct {
	ParentID   int         `json:"parent_id"`
	ParentName string      `json:"parent_name"`
	SpecID     int64       `json:"spec_id"`
	SpecName   string      `json:"spec_name"`
	SpecNote   interface{} `json:"spec_note"`
}

// VideoInfo 定义视频信息
type VideoInfo struct {
	FileID   interface{} `json:"file_id"`
	VideoURL interface{} `json:"video_url"`
}
