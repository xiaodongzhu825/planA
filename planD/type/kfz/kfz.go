package kfz

// DeleteGoodsCommit 删除商品
type DeleteGoodsCommit struct {
	ItemId string `json:"itemId"`
}

// DeleteGoodsCommitResponse 删除商品返回结构
type DeleteGoodsCommitResponse struct {
	ErrorResponse   interface{}      `json:"errorResponse"` // 根据实际类型可替换为具体结构体
	RequestId       string           `json:"requestId"`
	RequestMethod   string           `json:"requestMethod"`
	SuccessResponse *SuccessResponse `json:"successResponse"`
}

type SuccessResponse struct {
	Item *Item `json:"item"`
}

type Item struct {
	IsDelete   string `json:"isDelete"`
	ItemId     int64  `json:"itemId"`
	UpdateTime string `json:"updateTime"`
}
