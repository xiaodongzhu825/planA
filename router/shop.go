package router

import (
	"planA/controller"
	"planA/initialization/golabl"
)

// ShopInit 店铺信息
func ShopInit() {
	taskRouter := golabl.Router.PathPrefix("/shop").Subrouter()
	taskRouter.HandleFunc("/get/{shopId}", controller.GetShopInfo).Methods("GET") // 获取店铺信息
}
