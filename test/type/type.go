package _type

// ReturnData 返回数据结构体
type ReturnData struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// BodyErr 数据返回结构体
type BodyErr struct {
	Isbn   string `json:"isbn"`
	ErrMsg string `json:"err_msg"`
}

// DelShopIdAddIsbn 删除指定店铺的isbn结构体
type DelShopIdAddIsbn struct {
	ISBN    string `json:"isbn"`
	Message string `json:"message"`
	ShopID  string `json:"shopId"`
	Success bool   `json:"success"`
}
