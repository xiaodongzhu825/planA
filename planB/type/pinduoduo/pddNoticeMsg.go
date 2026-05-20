package pinduoduo

type PddNoticeMsg struct {
	MallId  string `json:"mallId"`
	GoodsId string `json:"goodsId"`
	Type    string `json:"type"`
}
