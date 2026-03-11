package _type

import "sync/atomic"

// 任务结构体

// Task 关键数据结构
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
	BookInfo BookInfo   `json:"book_info"`
	Detail   TaskDetail `json:"detail"`
}

// BookInfo 书籍信息结构
type BookInfo struct {
	Isbn            string      `json:"isbn"`             // ISBN
	BookName        string      `json:"book_name"`        // 书名
	Author          string      `json:"author"`           // 作者
	Publishing      string      `json:"publishing"`       // 出版社
	PublicationDate string      `json:"publication_date"` // 出版时间
	Binding         string      `json:"binding"`          // 装帧
	PagesCount      int64       `json:"pages_count"`      // 页数
	WordsCount      int64       `json:"words_count"`      // 字数
	Format          int64       `json:"format"`           // 开本
	ImageObject     ImageObject `json:"image_object"`     // 图片
	Price           int64       `json:"price"`            // 售价
	CatIdObject     CatIdObject `json:"cat_id"`           // 分类
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
	Token                       string      `json:"token"`                            //店铺 token 店铺类型=拼多多店铺，此token则是常规token 店铺类型=咸鱼店铺，此token则是【应用Id:应用密钥】
	GoodsNamePrefix             string      `json:"goods_name_prefix"`                //店铺名称前缀
	GoodsNameSuffix             string      `json:"goods_name_suffix"`                //店铺名称后缀
	TitleConsistOf              string      `json:"title_consist_of"`                 //标题包含信息 如：作者、出版社等等
	SpaceCharacter              string      `json:"space_character"`                  //是否使用空格 1为使用
	WatermarkImgUrl             string      `json:"watermark_img_url"`                //水印图片链接
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

// PriceMod 价格模版结构体
type PriceMod struct {
	Min         int64 `json:"min"`          // 价格区间最小值
	Max         int64 `json:"max"`          // 价格区间最大值
	MarkupRate  int64 `json:"markup_rate"`  // 加价比例
	MarkupValue int64 `json:"markup_value"` // 价格区间加价值
}

// TaskStatus 任务状态
type TaskStatus int64

const (
	TaskStatusRunning TaskStatus = 1 // 运行中
	TaskStatusPaused  TaskStatus = 2 // 已暂停
	TaskStatusStopped TaskStatus = 3 // 已停止
	TaskStatusOver    TaskStatus = 4 // 已完成
)

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
	OutGoodsId string `json:"out_goods_id"` // 商品编码
	GoodsName  string `json:"goods_name"`   // 商品名称
}

// ImageObject 图片对象结构
type ImageObject struct {
	CarouselUrlArray   []string          `json:"carousel_url_array"`   // 轮播图
	WhiteBackgroundUrl string            `json:"white_background_url"` // 白底图
	DetailUrlObject    DetailImageObject `json:"detail_url_object"`    // 详情对象
	DefaultImageUrl    string            `json:"default_image_url"`    // 默认图
}

// CatIdObject 平台分类结构
type CatIdObject struct {
	PinDuoDuoCatId int64 `json:"pin_duo_duo_cat_id"` // 拼多多分类 ID
	KongFuZiCatId  int64 `json:"kong_fu_zi_cat_id"`  // 孔夫子分类 ID
	XianYuCatId    int64 `json:"xian_yu_cat_id"`     // 闲鱼分类 ID
}

// DetailImageObject 详情图片对象结构
type DetailImageObject struct {
	IntroductionUrl []string `json:"introduction_url"`  // 简介图
	CatalogueUrl    []string `json:"catalogue_url"`     // 目录图
	LiveShootingUrl []string `json:"live_shooting_url"` // 实拍图
	OtherUrl        []string `json:"other_url"`         // 其他图
}

// PriceRange 价格区间
type PriceRange struct {
	MinPrice      int64       `json:"minPrice"`
	MaxPrice      int64       `json:"maxPrice"`
	AdjustPercent interface{} `json:"adjustPercent"` // 可能是 int 或 string
	AdjustAmount  int64       `json:"adjustAmount"`
}

// DistrictMsg 地区信息
type DistrictMsg struct {
	DistrictId   int64  `json:"district_id"`
	DistrictType string `json:"district_type"`
}
