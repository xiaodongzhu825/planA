package kfz

// GoodsAdd 商品添加结构体
type GoodsAdd struct {
	Tpl           string `json:"tpl"`                     // 模板编号，取值范围：1～17，必须
	CatId         string `json:"catId"`                   // 分类编号，必须
	MyCatId       string `json:"myCatId,omitempty"`       // 本店分类，可选
	ItemName      string `json:"itemName"`                // 商品名称，长度限制200字符，必须
	ImportantDesc string `json:"importantDesc,omitempty"` // 推荐语，长度限制200字符，可选
	Price         string `json:"price"`                   // 售价，0.01～99999999.99，必须
	Number        string `json:"number"`                  // 库存，1~9999，必须
	Quality       string `json:"quality"`                 // 品相，可以是编号(int)或文字(string)，取值：10,20,30,40,50,60,65,70,75,80,85,90,95,100，必须
	QualityDesc   string `json:"qualityDesc,omitempty"`   // 品相描述，长度限制400字符，可选
	ItemSn        string `json:"itemSn,omitempty"`        // 货号，长度限制20字符，可选
	ImgUrl        string `json:"imgUrl"`                  // 商品主图，必须
	Images        string `json:"images,omitempty"`        // 多个商品图片路径，用英文分号隔开，最多30张，可选
	ItemDesc      string `json:"itemDesc,omitempty"`      // 商品描述，长度限制10000字符，可选
	BearShipping  string `json:"bearShipping"`            // 运费设置：seller(卖家包邮)，buyer(买家承担运费)，必须
	MouldId       string `json:"mouldId,omitempty"`       // 运费模板编号，bearShipping=buyer时必填
	Weight        string `json:"weight,omitempty"`        // 商品重量(千克)，选择按重量模板时必填，0.01～9999.99
	WeightPiece   string `json:"weightPiece,omitempty"`   // 商品标准本数，选择按标准本模板时必填，0.01～9999.99
}

// GoodsAdd17 商品添加结构体（分类17专用）
type GoodsAdd17 struct {
	GoodsAdd
	Isbn             string      `json:"isbn"`                       // ISBN号，必须
	Author           string      `json:"author"`                     // 作者，长度限制120字符，必须
	Press            string      `json:"press"`                      // 出版社，长度限制120字符，必须
	PubDate          string      `json:"pubDate"`                    // 出版日期，格式：yyyy-mm，必须
	Edition          string      `json:"edition,omitempty"`          // 版次，取值范围：1～9999，可选
	PrintingTime     string      `json:"printingTime,omitempty"`     // 印刷时间，格式：yyyy-mm，填写printTimes时必填，不能早于出版时间，可选
	PrintTimes       string      `json:"printTimes,omitempty"`       // 印次，取值范围：1～9999，可选
	PrintingNum      string      `json:"printingNum,omitempty"`      // 印数，单位:千册，取值范围：0.001～99999.999，可选
	Binding          interface{} `json:"binding"`                    // 装帧，可以是编号(int)或文字(string)，必须
	PageSize         string      `json:"pageSize,omitempty"`         // 开本，可自己填写，可选
	Paper            interface{} `json:"paper,omitempty"`            // 纸张，可以是编号(int)或文字(string)，可选
	PageNum          string      `json:"pageNum,omitempty"`          // 页数，取值范围：1～99999，可选
	WordNum          string      `json:"wordNum,omitempty"`          // 字数，单位:千字，取值范围：0～99999.999，可选
	OriPrice         string      `json:"oriPrice,omitempty"`         // 图书定价，取值范围：0.01～99999999.99，可选
	ForeignName      string      `json:"foreignName,omitempty"`      // 原版书名，长度限制200字符，可选
	UnifiedIsbn      string      `json:"unifiedIsbn,omitempty"`      // 统一书号，长度限制20字符，可选
	PublishedIn      string      `json:"publishedIn,omitempty"`      // 出版地，长度限制100字符，可选
	Language         string      `json:"language,omitempty"`         // 语种，长度限制30字符，可选
	OriginalLanguage string      `json:"originalLanguage,omitempty"` // 原版语种，长度限制30字符，可选
	Series           string      `json:"series,omitempty"`           // 丛书系列，长度限制60字符，可选
	ContentIntro     string      `json:"contentIntro,omitempty"`     // 内容介绍，长度限制10000字符，可选
	AuthorIntro      string      `json:"authorIntro,omitempty"`      // 作者介绍，长度限制10000字符，可选
	Directory        string      `json:"directory,omitempty"`        // 目录，长度限制5000字符，可选
}

// GoodsAdd13 商品添加结构体（分类13专用）
type GoodsAdd13 struct {
	GoodsAdd
	Author       string      `json:"author"`                 // 作者，长度限制120字符，必须
	Press        string      `json:"press"`                  // 出版社，长度限制120字符，必须
	PubDate      string      `json:"pubDate"`                // 出版日期，格式：yyyy-mm，必须
	Edition      string      `json:"edition,omitempty"`      // 版次，取值范围：1～9999，可选
	PrintingTime string      `json:"printingTime,omitempty"` // 印刷时间，格式：yyyy-mm，填写printTimes时必填，不能早于出版时间，可选
	PrintTimes   string      `json:"printTimes,omitempty"`   // 印次，取值范围：1～9999，可选
	PrintingNum  string      `json:"printingNum,omitempty"`  // 印数，单位:千册，取值范围：0.001～99999.999，可选
	Binding      interface{} `json:"binding"`                // 装帧，可以是编号(int)或文字(string)，必须
	PageSize     string      `json:"pageSize,omitempty"`     // 开本，可自己填写，可选
	Paper        interface{} `json:"paper,omitempty"`        // 纸张，可以是编号(int)或文字(string)，可选
	PageNum      string      `json:"pageNum,omitempty"`      // 页数，取值范围：1～99999，可选
	WordNum      string      `json:"wordNum,omitempty"`      // 字数，单位:千字，取值范围：0～99999.999，可选
	OriPrice     string      `json:"oriPrice,omitempty"`     // 图书定价，取值范围：0.01～99999999.99，可选
}

// UploadImgRet 图片上传返回结构体
type UploadImgRet struct {
	ErrorResponse   *ErrorResponse               `json:"errorResponse"`
	RequestId       string                       `json:"requestId"`
	RequestMethod   string                       `json:"requestMethod"`
	SuccessResponse *UploadImgRetSuccessResponse `json:"successResponse"`
}

type ErrorResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Msg     string      `json:"msg"`
	SubCode string      `json:"subCode"`
	SubMsg  string      `json:"subMsg"`
}

type UploadImgRetSuccessResponse struct {
	Image Image `json:"image"`
}

type Image struct {
	Url string `json:"url"`
}

// KfzCategoryList 获取本店分类返回结构体
type KfzCategoryList struct {
	ErrorResponse   interface{}           `json:"errorResponse"`
	RequestId       string                `json:"requestId"`
	RequestMethod   string                `json:"requestMethod"`
	SuccessResponse []SuccessResponseItem `json:"successResponse"`
}

type SuccessResponseItem struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// KfzCategoryRet 获取公共分类返回结构体
type KfzCategoryRet struct {
	ErrorResponse   interface{}      `json:"errorResponse"`
	RequestId       string           `json:"requestId"`
	RequestMethod   string           `json:"requestMethod"`
	SuccessResponse []CategoryLevel1 `json:"successResponse"`
}

// 一级分类 (图书, 艺术品收藏, 文创与周边)
type CategoryLevel1 struct {
	Children []CategoryLevel2 `json:"children"`
	HasLeaf  int              `json:"hasLeaf"`
	Id       string           `json:"id"`
	Level    int              `json:"level"`
	Name     string           `json:"name"`
	Type     int              `json:"type"`
	Years    interface{}      `json:"years,omitempty"` // 有些没有years字段
}

// 二级分类
type CategoryLevel2 struct {
	Children []CategoryLevel3 `json:"children"`
	HasLeaf  int              `json:"hasLeaf"`
	Id       string           `json:"id"`
	Level    int              `json:"level"`
	Name     string           `json:"name"`
	Type     int              `json:"type"`
	Tpl      int              `json:"tpl,omitempty"`   // 有些有tpl字段
	Years    interface{}      `json:"years,omitempty"` // 有些有years字段
}

// 三级分类
type CategoryLevel3 struct {
	Children []CategoryLevel4 `json:"children,omitempty"`
	HasLeaf  int              `json:"hasLeaf"`
	Id       string           `json:"id"`
	Level    int              `json:"level"`
	Name     string           `json:"name"`
	Tpl      int              `json:"tpl,omitempty"`
	Type     int              `json:"type,omitempty"`
	Years    interface{}      `json:"years,omitempty"`
	EndYears interface{}      `json:"endYears,omitempty"` // 有些有endYears字段
}

// 四级分类
type CategoryLevel4 struct {
	Children []CategoryLevel5 `json:"children,omitempty"`
	HasLeaf  int              `json:"hasLeaf"`
	Id       string           `json:"id"`
	Level    int              `json:"level"`
	Name     string           `json:"name"`
	Tpl      int              `json:"tpl,omitempty"`
	Years    interface{}      `json:"years,omitempty"`
	EndYears interface{}      `json:"endYears,omitempty"`
}

// 五级分类
type CategoryLevel5 struct {
	Children []CategoryLevel6 `json:"children,omitempty"`
	HasLeaf  int              `json:"hasLeaf"`
	Id       string           `json:"id"`
	Level    int              `json:"level"`
	Name     string           `json:"name"`
	Tpl      int              `json:"tpl,omitempty"`
	Years    interface{}      `json:"years,omitempty"`
	EndYears interface{}      `json:"endYears,omitempty"`
}

// 六级分类 (最深层级)
type CategoryLevel6 struct {
	HasLeaf  int         `json:"hasLeaf"`
	Id       string      `json:"id"`
	Level    int         `json:"level"`
	Name     string      `json:"name"`
	Tpl      int         `json:"tpl,omitempty"`
	Years    interface{} `json:"years,omitempty"`
	EndYears interface{} `json:"endYears,omitempty"`
}

// AddGoodsRet 通用响应结构体
type AddGoodsRet struct {
	ErrorResponse   *AddGoodsError              `json:"errorResponse"`
	RequestId       string                      `json:"requestId"`
	RequestMethod   string                      `json:"requestMethod"`
	SuccessResponse *AddGoodsRetSuccessResponse `json:"successResponse"`
}

// AddGoodsError 错误响应详情
type AddGoodsError struct {
	Code    int               `json:"code"`
	Data    map[string]string `json:"data"` // 字段错误时使用 map，也可用具体结构体
	Msg     string            `json:"msg"`
	SubCode string            `json:"subCode"`
	SubMsg  string            `json:"subMsg"`
}

// AddGoodsRetSuccessResponse 成功响应详情
type AddGoodsRetSuccessResponse struct {
	Item Item `json:"item"`
}

// Item 商品详情
type Item struct {
	AddTime string `json:"addTime"`
	ItemId  int64  `json:"itemId"`
}

// GetGoodsListReq 获取商品列表请求结构体
type GetGoodsListReq struct {
	ItemId       string `json:"itemId"`
	Type         string `json:"type"`
	PageNum      int    `json:"pageNum"`
	PageSize     int    `json:"pageSize"`
	AddTimeBegin string `json:"addTimeBegin"`
	AddTimeEnd   string `json:"addTimeEnd"`
	SortOrder    string `json:"sortOrder"`
	SortType     string `json:"sortType"`
}

// GetGoodsListResp 获取商品列表响应结构体
type GetGoodsListResp struct {
	ErrorResponse   interface{}              `json:"errorResponse"`
	RequestId       string                   `json:"requestId"`
	RequestMethod   string                   `json:"requestMethod"`
	SuccessResponse *GetGoodsListSuccessResp `json:"successResponse"`
}

// GetGoodsListSuccessResp 成功响应详情
type GetGoodsListSuccessResp struct {
	List     []KfzGoodsItem `json:"list"`
	PageNum  int            `json:"pageNum"`
	PageSize int            `json:"pageSize"`
	Pages    int            `json:"pages"`
	Size     int            `json:"size"`
	Total    int            `json:"total"`
}

// KfzGoodsItem 孔夫子商品项
type KfzGoodsItem struct {
	ItemId        int64   `json:"itemId"`        // 商品编号
	AddTime       int64   `json:"addTime"`       // 添加时间（Unix时间戳，秒级）
	ItemName      string  `json:"itemName"`      // 商品名称
	Price         float64 `json:"price"`         // 售价（单位：元）
	Number        int     `json:"number"`        // 库存数量
	Quality       int64   `json:"quality"`       // 品相（如：85、90）
	QualityDesc   string  `json:"qualityDesc"`   // 品相描述
	ImgUrl        string  `json:"imgUrl"`        // 商品主图URL
	Images        string  `json:"images"`        // 商品图片列表
	CatId         uint64  `json:"catId"`         // 商品分类编号（可能超过int64范围，使用uint64）
	MyCatId       int     `json:"myCatId"`       // 本店分类编号
	ItemSn        string  `json:"itemSn"`        // 货号
	BearShipping  string  `json:"bearShipping"`  // 运费设置（seller/buyer）
	MouldId       int     `json:"mouldId"`       // 运费模板编号
	Weight        float64 `json:"weight"`        // 商品重量（千克）
	WeightPiece   float64 `json:"weightPiece"`   // 商品标准本数
	ItemDesc      string  `json:"itemDesc"`      // 商品描述
	Isbn          string  `json:"isbn"`          // ISBN号
	Author        string  `json:"author"`        // 作者
	Press         string  `json:"press"`         // 出版社
	PubDate       string  `json:"pubDate"`       // 出版日期
	OriPrice      float64 `json:"oriPrice"`      // 图书定价
	Binding       string  `json:"binding"`       // 装帧
	PageSize      string  `json:"pageSize"`      // 开本
	PageNum       int     `json:"pageNum"`       // 页数
	Tpl           int     `json:"tpl"`           // 模板编号
	ImportantDesc string  `json:"importantDesc"` // 推荐语
	// 以下为实际API返回的额外字段
	BeginSaleTime uint64 `json:"beginSaleTime"` // 上架时间（Unix时间戳）
	EndSaleTime   uint64 `json:"endSaleTime"`   // 下架时间（Unix时间戳）
	BizType       int    `json:"bizType"`       // 业务类型
	BooklibId     uint64 `json:"booklibId"`     // 书库ID
	CertifyStatus string `json:"certifyStatus"` // 认证状态
	Discount      int    `json:"discount"`      // 折扣
	IsDelete      int    `json:"isDelete"`      // 是否删除
	IsDraft       int    `json:"isDraft"`       // 是否草稿
	IsNewBook     int    `json:"isNewBook"`     // 是否新书
	IsOnSale      int    `json:"isOnSale"`      // 是否上架
	ProductArea   uint64 `json:"productArea"`   // 产地
	UpdateTime    string `json:"updateTime"`    // 更新时间（yyyy-MM-dd HH:mm:ss）
	UserId        uint64 `json:"userId"`        // 用户ID
	Years         uint64 `json:"years"`         // 年份
}

// Product 上架与下架请求结构体
type Product struct {
	ItemId string `json:"itemId"`
}

// ProductRet 上架与下架返回结构体
type ProductRet struct {
	ErrorResponse   interface{} `json:"errorResponse"`
	RequestId       string      `json:"requestId"`
	RequestMethod   string      `json:"requestMethod"`
	SuccessResponse struct {
		Item struct {
			BeginSaleTime string `json:"beginSaleTime"`
			CertifyStatus string `json:"certifyStatus"`
			EndSaleTime   string `json:"endSaleTime"`
			IsOnSale      string `json:"isOnSale"`
			ItemId        int    `json:"itemId"`
			UpdateTime    string `json:"updateTime"`
		} `json:"item"`
	} `json:"successResponse"`
}

// UpdatePriceReq 改价格请求结构体
type UpdatePriceReq struct {
	ItemId string  `json:"itemId"`
	Price  float64 `json:"price"`
}

// UpdatePriceRet 改价格返回结构体
type UpdatePriceRet struct {
	ErrorResponse   interface{} `json:"errorResponse"`
	RequestId       string      `json:"requestId"`
	RequestMethod   string      `json:"requestMethod"`
	SuccessResponse struct {
		Item struct {
			UpdateTime string `json:"updateTime"`
			ItemId     int    `json:"itemId"`
			Price      string `json:"price"`
		} `json:"item"`
	} `json:"successResponse"`
}

// UpdateStockReq 改库存请求结构体
type UpdateStockReq struct {
	ItemId string `json:"itemId"`
	Number int64  `json:"number"`
}

// UpdateStockRet 改库存返回结构体
type UpdateStockRet struct {
	ErrorResponse   interface{} `json:"errorResponse"`
	RequestId       string      `json:"requestId"`
	RequestMethod   string      `json:"requestMethod"`
	SuccessResponse struct {
		Item struct {
			UpdateTime string `json:"updateTime"`
			ItemId     int    `json:"itemId"`
			Number     string `json:"number"`
		} `json:"item"`
	} `json:"successResponse"`
}

// AddGoodsToErp 获取商品后请求ERP结构体
type AddGoodsToErp struct {
	ShopId    string `json:"shopId"`
	ShopType  string `json:"shopType"`
	Token     string `json:"token"`
	SycFlag   int    `json:"sycFlag"`
	TaskId    string `json:"taskId"`
	PageFlag  int    `json:"pageFlag"`
	GoodsList []GoodsList
}
type GoodsList struct {
	ISBN          string `json:"isbn"`
	ItemName      string `json:"itemName"`
	Price         string `json:"price"`
	Quality       string `json:"quality"`
	Author        string `json:"author"`
	Press         string `json:"press"`
	PubDate       string `json:"pubDate"`
	ItemId        string `json:"itemId"`
	AddTime       string `json:"addTime"`
	BeginSaleTime string `json:"beginSaleTime"`
	IsDraft       string `json:"isDraft"`
	Discount      string `json:"discount"`
	Stock         string `json:"stock"`
	MyCatId       string `json:"myCatId"`
	BearShipping  string `json:"bearShipping"`
	Weight        string `json:"weight"`
	CatId         string `json:"catId"`
	IsNewBook     string `json:"isNewBook"`
	BizType       string `json:"bizType"`
	CertifyStatus string `json:"certifyStatus"`
	WeightPiece   string `json:"weightPiece"`
	MouldId       string `json:"mouldId"`
	BooklibId     string `json:"booklibId"`
	IsOnSale      string `json:"isOnSale"`
	IsDelete      string `json:"isDelete"`
	UpdateTime    string `json:"updateTime"`
	EndSaleTime   string `json:"endSaleTime"`
	UserId        string `json:"userId"`
	ImgUrl        string `json:"imgUrl"`
	OriPrice      string `json:"oriPrice"`
	ItemSn        string `json:"itemSn"`
}

// AddGoodsToErpRet 获取商品后请求ERP返回结构体
type AddGoodsToErpRet struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data bool   `json:"data"`
}
