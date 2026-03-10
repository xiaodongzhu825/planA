package router

import (
	"net/http"
	"planA/initialization/golabl"
)

// StaticInit 静态文件初始化
func StaticInit() {
	golabl.Router.PathPrefix("/export/").Handler(http.StripPrefix("/export/", http.FileServer(http.Dir("./export"))))
}
