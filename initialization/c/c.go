package c

import (
	"fmt"
	"planA/initialization/golabl"
	"planA/tool/process"
)

// RunC 运行C代码
func RunC() {
	if golabl.Config.Server.IsC {
		//启动 C程序
		if err := process.RunCProgram(); err != nil {
			fmt.Println("启动C程序失败:", err)
			return
		}
	}
}
