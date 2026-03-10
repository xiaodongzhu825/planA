package _type

// 违禁词结构体

// HttpBannedWordSubstitutionBookInfoRes 违禁词响应结构体
type HttpBannedWordSubstitutionBookInfoRes struct {
	Msg       string      `json:"msg"`       // 查询成功
	Code      string      `json:"code"`      // 状态码 200
	Data      []MatchRule `json:"data"`      // 匹配规则列表
	Success   bool        `json:"success"`   // 是否成功
	Author    string      `json:"author"`    // 作者
	Isbn      string      `json:"isbn"`      // ISBN
	Publisher string      `json:"publisher"` // 出版社
	BookName  string      `json:"bookName"`  // 书名（可能包含***）
}

// MatchRule 匹配规则结构体
type MatchRule struct {
	CreateBy       int64  `json:"createBy"`       // 创建人ID
	MatchType      string `json:"matchType"`      // 匹配类型：ISBN匹配/书名匹配
	AddTxt         string `json:"addTxt"`         // 匹配文本
	ID             int64  `json:"id"`             // 规则ID
	Sort           string `json:"sort"`           // 排序信息 "0,3"
	LimitationType string `json:"limitationType"` // 限制类型 "0"/"1"/"6"
}
