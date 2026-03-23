package task

import (
	"fmt"
	"planA/planB/server"
)

// GetTaskHeaderAndFooterSetToG 获取任务头和尾并保存到全局变量中
// @return error 错误信息
func GetTaskHeaderAndFooterSetToG() error {
	// 获取任务头
	if err := server.GetTaskHeader(); err != nil {
		return fmt.Errorf("获取任务头失败 %v", err)
	}
	// 获取任务尾
	if err := server.GetTaskFooter(); err != nil {
		return fmt.Errorf("获取任务尾失败 %v", err)
	}
	return nil
}
