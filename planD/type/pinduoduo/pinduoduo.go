package pinduoduo

// DeleteGoodsCommit 删除商品
type DeleteGoodsCommit struct {
	GoodsIds []int64 `json:"goods_ids"`
}

// DeleteGoodsCommitResponse 删除商品响应结构
type DeleteGoodsCommitResponse struct {
	OpenAPIResponse bool   `json:"open_api_response"`
	RequestID       string `json:"request_id"`
}
