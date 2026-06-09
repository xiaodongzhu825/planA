package kfz

// GetGoodsListReq 获取商品列表请求结构体
type GetGoodsListReq struct {
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
	ItemId        int64   `json:"itemId"`
	AddTime       int64   `json:"addTime"`
	ItemName      string  `json:"itemName"`
	Price         float64 `json:"price"`
	Number        int     `json:"number"`
	Quality       int64   `json:"quality"`
	QualityDesc   string  `json:"qualityDesc"`
	ImgUrl        string  `json:"imgUrl"`
	Images        string  `json:"images"`
	CatId         uint64  `json:"catId"`
	MyCatId       int     `json:"myCatId"`
	ItemSn        string  `json:"itemSn"`
	BearShipping  string  `json:"bearShipping"`
	MouldId       int     `json:"mouldId"`
	Weight        float64 `json:"weight"`
	WeightPiece   float64 `json:"weightPiece"`
	ItemDesc      string  `json:"itemDesc"`
	Isbn          string  `json:"isbn"`
	Author        string  `json:"author"`
	Press         string  `json:"press"`
	PubDate       string  `json:"pubDate"`
	OriPrice      float64 `json:"oriPrice"`
	Binding       string  `json:"binding"`
	PageSize      string  `json:"pageSize"`
	PageNum       int     `json:"pageNum"`
	Tpl           int     `json:"tpl"`
	ImportantDesc string  `json:"importantDesc"`
	BeginSaleTime uint64  `json:"beginSaleTime"`
	EndSaleTime   uint64  `json:"endSaleTime"`
	BizType       int     `json:"bizType"`
	BooklibId     uint64  `json:"booklibId"`
	CertifyStatus string  `json:"certifyStatus"`
	Discount      int     `json:"discount"`
	IsDelete      int     `json:"isDelete"`
	IsDraft       int     `json:"isDraft"`
	IsNewBook     int     `json:"isNewBook"`
	IsOnSale      int     `json:"isOnSale"`
	ProductArea   uint64  `json:"productArea"`
	UpdateTime    string  `json:"updateTime"`
	UserId        uint64  `json:"userId"`
	Years         uint64  `json:"years"`
}

// ProductRet 上架/下架/改价/改库存/删除通用返回结构体
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
			Price         string `json:"price"`
			Number        string `json:"number"`
		} `json:"item"`
	} `json:"successResponse"`
}

// ErrorResponse 通用错误响应
type ErrorResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Msg     string      `json:"msg"`
	SubCode string      `json:"subCode"`
	SubMsg  string      `json:"subMsg"`
}
