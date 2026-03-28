package main

import (
	"fmt"
	"planA/test/initialization"
	"planA/test/initialization/golabl"
	"planA/test/logic"
)

func main() {
	//fmt.Println("请输入店铺ID：")
	//fmt.Scanln(&golabl.ShopId)
	golabl.ShopId = "2035973762883399682"

	fmt.Println("请输入店铺类型  1 拼多多   2 孔夫子  5 闲鱼：")
	fmt.Scanln(&golabl.ShopType)

	fmt.Println("请输入任务类型  1核价发布  2 表格发布  3：拼多多商品拉取")
	fmt.Scanln(&golabl.TaskType)

	fmt.Println("请输入图片类型 1仅官图 2 实拍图 3 优先官图 4 优先实拍图")
	fmt.Scanln(&golabl.ImgType)

	err := initialization.Init()
	if err != nil {
		fmt.Println(err)
		return
	}

	logicErr := logic.Logic()
	if logicErr != nil {
		fmt.Println(logicErr)
	}

}
