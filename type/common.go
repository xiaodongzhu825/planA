package _type

//通用结构体

// CreateTaskResponse 创建任务响应结构体
type CreateTaskResponse struct {
	Msg    string `json:"msg"`
	Code   int    `json:"code"`
	TaskID string `json:"taskId"`
}

type Page struct {
	PageNum  int // 页码，从1开始
	PageSize int // 每页条数
}

// APIResponse API响应结构体
// 用于统一API接口的返回格式
type APIResponse struct {
	Success bool        `json:"success"`         // 请求是否成功
	Message string      `json:"message"`         // 响应消息
	Data    interface{} `json:"data,omitempty"`  // 响应数据（可选）
	Error   string      `json:"error,omitempty"` // 错误信息（可选）
}
