package main

import (
	"fmt"
	"os"
	"planA/planE/initialization"
	"planA/planE/logic"
)

func main() {

	imgUrl := os.Args[1]
	token := os.Args[2]

	err := initialization.Init()
	if err != nil {
		fmt.Println("初始化失败:", err)
		return
	}
	logic.Logic(imgUrl, token)
}
