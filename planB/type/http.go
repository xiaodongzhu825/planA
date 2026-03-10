package _type

// Response 通用响应结构体
type Response struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"` // omitempty 表示如果为空则忽略
}
