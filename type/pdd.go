package _type

//拼多多结构体

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
