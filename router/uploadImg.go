package router

import (
	"net/http"
	"planA/controller"
	"planA/initialization/golabl"
	"planA/initialization/middle"
)

// UploadImgInit 任务初始化
func UploadImgInit() {
	uploadImgRouter := golabl.Router.PathPrefix("/uploadImg").Subrouter()
	// ====================== 【需要验签】的接口 ======================
	uploadImgRouter.Handle("/ImgUploadToPdd", middle.Sign(http.HandlerFunc(controller.ImgUploadToPdd))).Methods("POST") // 上传图片到拼多多
	// ====================== 【不需要验签】的接口 ======================
	//uploadImgRouter.HandleFunc("/ImgUploadToPdd", controller.ImgUploadToPdd).Methods("POST") // 创建新任务
}
