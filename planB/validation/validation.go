package validation

import (
	"fmt"
	"os"
)

func Validation() (string, error) {
	taskId := os.Args[1]
	if taskId == "" {
		return "", fmt.Errorf("任务Id 不能为空")
	}
	return taskId, nil
}
