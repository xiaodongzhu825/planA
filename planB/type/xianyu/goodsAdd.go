package xianyu

// GoodsAdd 商品新增请求结构体
// 用于向各电商平台（闲鱼、拼多多、淘宝等）提交商品上架的相关信息
type GoodsAdd struct {
	AppId        int64      `json:"appId"`        // 应用 id
	AppSecret    string     `json:"appSecret"`    // 应用密钥[选填，有些平台需要]
	Token        string     `json:"token"`        // token[选填，有些平台需要]
	ApiShopId    int        `json:"apiShopId"`    // API使用的店铺ID[选填，有些平台需要]
	TypePlatform int        `json:"typePlatform"` // 平台类型 0-预留  1-拼多多  2-淘宝  3-京东  4-闲鱼  105-孔夫子
	ShopId       int64      `json:"shopId"`       // 店铺 ID
	ShopToken    string     `json:"shopToken"`    // 店铺 Token
	ShopName     string     `json:"shopName"`     // 店铺名称
	Province     int        `json:"province"`     // 发货省，格式为省级行政区划代码（如210000代表辽宁省）
	City         int        `json:"city"`         // 发货市，格式为市级行政区划代码（如210100代表沈阳市）
	District     int        `json:"district"`     // 发货区，格式为区级行政区划代码（如210101代表和平区）
	TypeClass    string     `json:"typeClass"`    // 分类类型
	TypeGoods    string     `json:"typeGoods"`    // 商品类型
	CatIds       string     `json:"catIds"`       // 类目 ID
	SkuMsgs      []SkuMsg   `json:"skuMsgs"`      // 商品 SKU信息列表[选填]
	Shop         []ShopInfo `json:"shop"`         // 闲鱼用店铺信息
	StuffStatus  int64      `json:"stuffStatus"`  // 成色，90代表对应成色等级
	BookData     []BookInfo `json:"bookData"`     // 图书类商品专属信息列表
	ItemKey      string     `json:"itemKey"`      // 闲鱼批次商品 KEY
}

// ShopInfo 闲鱼店铺信息结构体
// 包含闲鱼平台商品上架所需的店铺及商品基础信息
type ShopInfo struct {
	UserName    string   `json:"userName"`    // 闲鱼会员名（必填）
	Province    int      `json:"province"`    // 发货省（必填），行政区划代码格式
	City        int      `json:"city"`        // 发货市（必填），行政区划代码格式
	District    int      `json:"district"`    // 发货区（必填），行政区划代码格式
	Title       string   `json:"title"`       // 商品标题（必填）
	Content     string   `json:"content"`     // 商品描述（必填）
	MainImgs    []string `json:"mainImgs"`    // 商品主图（必填），图片URL列表
	ContentImgs []string `json:"contentImgs"` // 商品内容图（选填），图片URL列表
}

// BookInfo 图书类商品信息结构体
// 包含图书类商品上架所需的专属信息，适用于闲鱼等平台的图书品类
type BookInfo struct {
	ISBN        string  `json:"ISBN"`        // ISBN编号（必填），图书唯一标识
	Title       string  `json:"Title"`       // 书名（必填）
	Author      string  `json:"Author"`      // 作者（选填）
	Publisher   string  `json:"Publisher"`   // 出版社（选填）
	ItemBizType int     `json:"itemBizType"` // 闲鱼商品类型（枚举），2：普通商品（必填）
	SpBizType   int     `json:"spBizType"`   // 闲鱼行业类型（枚举），24：图书（必填）
	Prices      []int64 `json:"prices"`      // 商品价格（必填），格式为[商品原价，商品售价]，单位为分
	Stock       int64   `json:"stock"`       // 库存（必填），商品可售数量
	CatIds      string  `json:"catIds"`      // 商品类目ID（必填）
}

// SkuMsg 商品SKU信息结构体
// 补充定义原需求中提到的skuMsgs字段对应的结构体，保证结构体完整性
type SkuMsg struct {
	Key         string   `json:"key"`         // 主键（必填）
	Value       string   `json:"value"`       // 值（必填）
	Title       string   `json:"title"`       // 商品标题（必填）
	CatIds      string   `json:"cat_ids"`     // 商品类目（必填）
	MainImgs    []string `json:"mainImgs"`    // 商品主图（必填），图片URL列表
	ContentImgs []string `json:"contentImgs"` // 商品内容图（选填），图片URL列表
	ItemBizType int      `json:"itemBizType"` // 闲鱼商品类型（枚举），2：普通商品（必填）
	SpBizType   int      `json:"spBizType"`   // 闲鱼行业类型（枚举），24：图书（必填）
	Prices      []int    `json:"prices"`      // 商品价格（必填），[商品售价，商品原价]，单位为分
	Stock       int      `json:"stock"`       // 库存（必填）
	Content     string   `json:"content"`     // 商品描述（必填）
	UserName    string   `json:"userName"`    // 闲鱼会员名（必填）
}

// Token 闲鱼店铺token传递的是json串
type Token struct {
	AppId     int64  `json:"app_id"`
	AppSecret string `json:"app_secret"`
	Username  string `json:"username"`
}

// Product 上架商品结构体
type Product struct {
	AppId              int64    `json:"appId"`                // 应用 id
	AppSecret          string   `json:"appSecret"`            // 应用密钥[选填，有些平台需要]
	Token              string   `json:"token"`                // token[选填，有些平台需要]
	ProductID          int64    `json:"product_id"`           // 商品 ID
	UserName           []string `json:"user_name"`            // 会员名
	SpecifyPublishTime string   `json:"specify_publish_time"` // 指定发布时间
	NotifyURL          string   `json:"notify_url"`           // 回调地址
}

// XianYuAddGoodsResponse 闲鱼商品新增响应结构体
type XianYuAddGoodsResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data Data   `json:"data"`
}

type Data struct {
	Success []SuccessItem `json:"success"`
	Error   []interface{} `json:"error"` // 空数组，使用 interface{} 或定义具体结构
}

type SuccessItem struct {
	ItemKey       string `json:"item_key"`
	ProductID     int64  `json:"product_id"` // 注意：这个数字较大，使用 int64
	ProductStatus int    `json:"product_status"`
}
