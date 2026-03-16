package _type

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
)

// TaskStatus 任务状态
type TaskStatus int64

const (
	TaskStatusRunning TaskStatus = 1 // 运行中
	TaskStatusPaused  TaskStatus = 2 // 已暂停
	TaskStatusStopped TaskStatus = 3 // 已停止
	TaskStatusOver    TaskStatus = 4 // 已完成
)

// GoodsType 接口类型
type GoodsType string

const (
	GoodsTypeAdd GoodsType = "新增商品" // 新增商品
	GoodsTypeSet GoodsType = "设置商品" // 设置商品
	GoodsTypeGet GoodsType = "获取商品" // 获取商品
	GoodsTypeDel GoodsType = "删除商品" // 删除商品
)

// Task 任务结构
type Task struct {
	Header   TaskHeader `json:"header"`    // 任务头
	BodyWait TaskBody   `json:"body_wait"` // 任务队列
	BodyOver TaskBody   `json:"body_over"` // 已完成任务队列
	Footer   TaskFooter `json:"footer"`    // 任务尾
}

// TaskHeader 任务头结构

type TaskHeader struct {
	TaskId           string     `json:"task_id"`            // 任务 ID
	TaskType         int64      `json:"task_type"`          // 任务类型
	ShopId           int64      `json:"shop_id"`            // 店铺 ID
	ShopName         string     `json:"shop_name"`          // 店铺名称
	ShopType         string     `json:"shop_type"`          // 店铺类型
	ShopMsg          ShopMsg    `json:"shop_msg"`           // 店铺信息
	PriceMod         []PriceMod `json:"price_mod"`          // 价格模版
	ShipPriceMod     string     `json:"ship_price_mod"`     // 运费模版
	TaskCount        int64      `json:"task_count"`         // 任务数量
	TaskCountTrue    int64      `json:"task_count_true"`    // 任务数量（真实）
	TaskCountWait    int64      `json:"task_count_wait"`    // 任务数量（等待）
	TaskCountOver    int64      `json:"task_count_over"`    // 任务数量（结束）
	TaskCountSuccess int64      `json:"task_count_success"` // 任务数量（成功）
	TaskCountError   int64      `json:"task_count_error"`   // 任务数量（错误）
	Status           TaskStatus `json:"status"`             // 任务状态
	TaskQpm          int64      `json:"task_qpm"`           // 任务 QPM
	TaskCreateAt     int64      `json:"task_create_at"`     // 任务创建时间
	TaskOverAt       int64      `json:"task_over_at"`       // 任务结束时间
	LastIndex        int64      `json:"last_index"`         // 最后任务索引
	ImgType          int64      `json:"img_type"`           //图片类型 1仅观图 2 实拍图 3 优先观图 4 优先实拍图
}

// TaskBody 任务主体结构
type TaskBody struct {
	BookInfo   BookInfo   `json:"book_info"`  //书籍信息
	Detail     TaskDetail `json:"detail"`     //书籍详情
	Publishing Publishing `json:"publishing"` //出版社信息
}

// TaskFooter 任务项结构
type TaskFooter struct {
	TaskCount        int64        `json:"task_count"`         // 任务数量
	TaskCountTrue    int64        `json:"task_count_true"`    // 任务数量（真实）
	TaskCountWait    atomic.Int64 `json:"task_count_wait"`    // 任务数量（等待）
	TaskCountOver    atomic.Int64 `json:"task_count_over"`    // 任务数量（结束）
	TaskCountSuccess atomic.Int64 `json:"task_count_success"` // 任务数量（成功）
	TaskCountError   atomic.Int64 `json:"task_count_error"`   // 任务数量（错误）
	TaskQpm          int64        `json:"task_qpm"`           // 任务 QPM
	LastIndex        int64        `json:"last_index"`         // 最后任务索引
}

// ShopMsg 店铺信息结构体
type ShopMsg struct {
	ID                          int64       `json:"id"`
	ShopAliasName               string      `json:"shop_alias_name"`
	ShopName                    string      `json:"shop_name"`
	Token                       string      `json:"token"`                            //店铺 token
	XianYuAppId                 int64       `json:"xian_yu_app_id"`                   //闲鱼 appId
	GoodsNamePrefix             string      `json:"goods_name_prefix"`                //店铺名称前缀
	GoodsNameSuffix             string      `json:"goods_name_suffix"`                //店铺名称后缀
	TitleConsistOf              string      `json:"title_consist_of"`                 //标题包含信息 如：作者、出版社等等
	SpaceCharacter              string      `json:"space_character"`                  //是否使用空格 1为使用
	WatermarkImgUrl             string      `json:"watermark_img_url"`                //水印图片链接
	WatermarkPosition           string      `json:"watermark_position"`               //水印位置 0全部  1第一张
	CarouseLastImgUrlArray      []string    `json:"carouse_last_img_url_array"`       //轮播图最后图片
	GoodsDetailFirstImgUrlArray []string    `json:"goods_detail_first_img_url_array"` //商品详情首图 URL 数组
	GoodsDetailLastImgUrlArray  []string    `json:"goods_detail_last_img_url_array"`  //商品详情最后图片 URL 数组
	SpecName                    string      `json:"spec_name"`                        //规格名称
	SpecId                      int64       `json:"spec_id"`                          //规格 ID
	SpecChildName               string      `json:"spec_child_name"`                  //子规格名称
	IsFolt                      bool        `json:"is_fotl"`                          //是否支持假一赔十，false-不支持，true-支持
	IsPreSale                   bool        `json:"is_pre_sale"`                      //是否预售,true-预售商品，false-非预售商品
	IsRefundable                bool        `json:"is_refundable"`                    //是否7天无理由退换货，true-支持，false-不支持
	IsSecondHand                bool        `json:"is_second_hand"`                   //是否二手 true -二手商品 ，false-全新商品
	ShipmentLimitSecond         int64       `json:"shipment_limit_second"`            //承诺发货时间（秒）
	CostTemplateId              int64       `json:"cost_template_id"`                 //物流运费模板 ID
	DefStock                    int64       `json:"def_stock"`                        // 默认库存
	TwoDiscount                 int64       `json:"two_discount"`                     // 两件折扣
	DistrictMsg                 DistrictMsg `json:"district_msg"`                     //地区信息【限闲鱼使用】
}

// BookInfo 书籍信息结构
type BookInfo struct {
	Isbn            string       `json:"isbn"`             // ISBN
	BookName        string       `json:"book_name"`        // 书名
	Author          string       `json:"author"`           // 作者
	Publishing      string       `json:"publishing"`       // 出版社
	PublicationDate string       `json:"publication_date"` // 出版时间
	Binding         string       `json:"binding"`          // 装帧
	PagesCount      int64        `json:"pages_count"`      // 页数
	WordsCount      int64        `json:"words_count"`      // 字数
	Format          int64        `json:"format"`           // 开本
	ImageObject     *ImageObject `json:"image_object"`     // 图片
	Price           int64        `json:"price"`            // 售价
	CatIdObject     CatIdObject  `json:"cat_id"`           // 分类
}

// ImageObject 图片对象结构
type ImageObject struct {
	CarouselUrlArray   []string          `json:"carousel_url_array"`   // 轮播图
	WhiteBackgroundUrl string            `json:"white_background_url"` // 白底图
	DetailUrlObject    DetailImageObject `json:"detail_url_object"`    // 详情对象
	DefaultImageUrl    string            `json:"default_image_url"`    // 默认图
}

// DetailImageObject 详情图片对象结构
type DetailImageObject struct {
	IntroductionUrl []string `json:"introduction_url"`  // 简介图
	CatalogueUrl    []string `json:"catalogue_url"`     // 目录图
	LiveShootingUrl []string `json:"live_shooting_url"` // 实拍图
	OtherUrl        []string `json:"other_url"`         // 其他图
}

// TaskDetail 详情结构
type TaskDetail struct {
	Condition  int64  `json:"condition"`    // 品相
	Price      int64  `json:"price"`        // 价格
	Stock      int64  `json:"stock"`        // 库存
	Status     int64  `json:"status"`       // 状态 0=失败 1=成功
	Error      string `json:"error"`        // 错误信息
	GoodsId    int64  `json:"goods_id"`     // 商品 ID
	ReturnId   int64  `json:"return_id"`    // 拼多多返回 ID
	SkuCode    string `json:"sku_code"`     // 规格编码（sku维度）
	SkuId      int64  `json:"sku_id"`       // sku编码
	Img        string `json:"img"`          // 图片
	OutGoodsId string `json:"out_goods_id"` // 商家编码
	GoodsName  string `json:"goods_name"`   // 商品名称
}

// PriceMod 价格处理
type PriceMod struct {
	Min         int64 `json:"min"`          // 价格区间最小值
	Max         int64 `json:"max"`          // 价格区间最大值
	MarkupRate  int64 `json:"markup_rate"`  // 加价比例
	MarkupValue int64 `json:"markup_value"` // 价格区间加价值
}

type CatIdObject struct {
	PinDuoDuoCatId FlexibleStr `json:"pin_duo_duo_cat_id"` // 拼多多分类 ID
	KongFuZiCatId  FlexibleStr `json:"kong_fu_zi_cat_id"`  // 孔夫子分类 ID
	XianYuCatId    FlexibleStr `json:"xian_yu_cat_id"`     // 闲鱼分类 ID
}

// FlexibleInt64 ====================== 临时 ======================
type FlexibleStr string

// UnmarshalJSON 反序列化：接受数字、布尔值、字符串等任意类型，都转换为字符串
func (fi *FlexibleStr) UnmarshalJSON(data []byte) error {
	// 1. 尝试直接解析为字符串
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fi = FlexibleStr(s)
		return nil
	}

	// 2. 尝试解析为数字
	var num json.Number
	if err := json.Unmarshal(data, &num); err == nil {
		*fi = FlexibleStr(num.String())
		return nil
	}

	// 3. 尝试解析为布尔值
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*fi = FlexibleStr(strconv.FormatBool(b))
		return nil
	}

	// 4. 其他任意类型
	var any interface{}
	if err := json.Unmarshal(data, &any); err != nil {
		return err
	}

	// 将任意类型转为字符串
	*fi = FlexibleStr(fmt.Sprintf("%v", any))
	return nil
}

// MarshalJSON 序列化：总是输出为字符串
func (fi FlexibleStr) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(fi))
}

// String 实现 Stringer 接口
func (fi FlexibleStr) String() string {
	return string(fi)
}

// ToInt64 如果需要转换为 int64
func (fi FlexibleStr) ToInt64() (int64, error) {
	return strconv.ParseInt(string(fi), 10, 64)
}

// ToFloat64 如果需要转换为 float64
func (fi FlexibleStr) ToFloat64() (float64, error) {
	return strconv.ParseFloat(string(fi), 64)
}

// ToBool 如果需要转换为 bool
func (fi FlexibleStr) ToBool() (bool, error) {
	return strconv.ParseBool(string(fi))
}

// FlexibleInt64 ====================== 临时 ======================

// DllGoodsSpec 拼多多接口 PddGoodsSpecIdGet 返回结构体
type DllGoodsSpec struct {
	DllGoodsSpec GoodsSpec `json:"goods_spec_id_get_response"`
}
type GoodsSpec struct {
	ParentSpecID int64  `json:"parent_spec_id"`
	RequestID    string `json:"request_id"`
	SpecID       int64  `json:"spec_id"`
	SpecName     string `json:"spec_name"`
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

// Publishing Redis中存储的出版社信息结构体
type Publishing struct {
	Value string `json:"value"`
	Vid   int64  `json:"vid"`
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
type PddErrorResponse struct {
	ErrorCode int64   `json:"error_code"` // 错误码
	ErrorMsg  string  `json:"error_msg"`  // 错误信息
	SubCode   *string `json:"sub_code"`   // 子错误码
	SubMsg    string  `json:"sub_msg"`    // 子错误信息
	RequestID string  `json:"request_id"` // 请求 ID
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

// SkuProperty 定义SKU属性
type SkuProperty struct {
	Punit  string `json:"punit"`
	RefPid int    `json:"ref_pid"`
	Value  string `json:"value"`
	Vid    int    `json:"vid"`
}

// Spec 定义规格信息
type Spec struct {
	ParentID   int         `json:"parent_id"`
	ParentName string      `json:"parent_name"`
	SpecID     int64       `json:"spec_id"`
	SpecName   string      `json:"spec_name"`
	SpecNote   interface{} `json:"spec_note"`
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

// VideoInfo 定义视频信息
type VideoInfo struct {
	FileID   interface{} `json:"file_id"`
	VideoURL interface{} `json:"video_url"`
}

// HttpBannedWordSubstitutionBookInfoReq 请求校验违禁词图书信息结构体
type HttpBannedWordSubstitutionBookInfoReq struct {
	Isbn        string `json:"isbn"`        // ISBN
	BookName    string `json:"bookName"`    // 书名
	Author      string `json:"author"`      // 作者
	Publisher   string `json:"publisher"`   // 出版社
	ShopId      string `json:"shopId"`      // 店铺ID
	ReplaceMark string `json:"replaceMark"` // 标题违规词是否替换 0不替换 1替换
}

// HttpBannedWordSubstitutionBookInfoRes 书籍查询响应结构体
type HttpBannedWordSubstitutionBookInfoRes struct {
	Msg       string      `json:"msg"`       // 查询成功
	Code      string      `json:"code"`      // 状态码 200
	Data      []MatchRule `json:"data"`      // 匹配规则列表
	Success   bool        `json:"success"`   // 是否成功
	Author    string      `json:"author"`    // 作者
	Isbn      string      `json:"isbn"`      // ISBN
	Publisher string      `json:"publisher"` // 出版社
	BookName  string      `json:"bookName"`  // 书名（可能包含***）
}

// MatchRule 匹配规则结构体
type MatchRule struct {
	CreateBy       int64  `json:"createBy"`       // 创建人ID
	MatchType      string `json:"matchType"`      // 匹配类型：ISBN匹配/书名匹配
	AddTxt         string `json:"addTxt"`         // 匹配文本
	ID             int64  `json:"id"`             // 规则ID
	Sort           string `json:"sort"`           // 排序信息 "0,3"
	LimitationType string `json:"limitationType"` // 限制类型 "0"/"1"/"6"
}

// DistrictMsg 地区信息
type DistrictMsg struct {
	DistrictId   int64  `json:"district_id"`
	DistrictType string `json:"district_type"`
}

// Returns DLL水印响应结构体
type Returns struct {
	Success bool   `json:"success"`
	Format  string `json:"format"`
	Data    string `json:"data"` // Base64编码的图片数据
	Size    int    `json:"size"`
}

// GoodsImageUploadResponse 商品图片上传响应
type GoodsImageUploadResponse struct {
	GoodsImageUploadResponse struct { // 注意：字段名与JSON中的key一致
		ImageURL  string `json:"image_url"`  // 图片URL地址
		RequestID string `json:"request_id"` // 请求ID
	} `json:"goods_image_upload_response"` // 外层字段名
}
