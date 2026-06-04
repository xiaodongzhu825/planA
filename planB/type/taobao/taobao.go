package taobao

import (
	"net/http"
)

// Tushu 图书商品管理工具主结构体。
type Tushu struct {
	Client *http.Client
}
