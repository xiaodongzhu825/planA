package main

import (
	"fmt"
	"planA/planB/initialization"
	"planA/planB/logic"
	"planA/planB/validation"
	"syscall"
	"time"
	"unsafe"
)

func main() {

	//校验参数
	taskId, validationErr := validation.Validation()
	if validationErr != nil {
		fmt.Println(validationErr)
		return
	}

	// 是否测试模式
	if taskId == "111" {
		test()
		return
	}

	//设置窗口标题
	setConsoleTitle(taskId)

	// 初始化配置
	err := initialization.Init(taskId)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 执行任务
	logic.Logic()

}

// 测试模式
func test() {
	//循环1000次
	for i := 0; i < 1000; i++ {
		//每秒打印 i
		fmt.Printf("i:%v\n", i)
		time.Sleep(time.Second)
	}
}

// 设置窗口标题
func setConsoleTitle(title string) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleTitle := kernel32.NewProc("SetConsoleTitleW")
	// 将字符串转换为UTF-16指针
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	procSetConsoleTitle.Call(uintptr(unsafe.Pointer(titlePtr)))
}
