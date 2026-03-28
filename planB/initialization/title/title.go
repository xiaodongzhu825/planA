package title

import (
	"fmt"
	"planA/planB/initialization/golabl"
	"planA/planB/tool"
	"time"
)

// SetWinTitle 设置窗口标题
func SetWinTitle() {
	title := ""

	//平台
	switch golabl.Task.Header.ShopType {
	//case 2:
	//	return kongFuZi.NewKongfuzi(), nil
	case "1":
		title = title + "【拼多多】"
	case "5":
		title = title + "【闲鱼】"
	default:
		title = title + "【其他平台 " + golabl.Task.Header.ShopType + "】"
	}

	//任务类型
	switch golabl.Task.Header.TaskType {
	case 1:
		title = title + "【核价发布】"
	case 2:
		title = title + "【表格发布】"
	case 3:
		title = title + "【商品发布】"
	default:
		title = title + "【其他任务类型 " + fmt.Sprint(golabl.Task.Header.TaskType) + "】"
	}

	//图片类型
	switch golabl.Task.Header.ImgType {
	case 1:
		title = title + "【仅官图】"
	case 2:
		title = title + "【实拍图】"
	case 3:
		title = title + "【优先官图】"
	case 4:
		title = title + "【优先实拍图】"
	default:
		title = title + "【其他图片类型 " + fmt.Sprint(golabl.Task.Header.ImgType) + "】"
	}

	//创建时间
	createTime := time.Unix(golabl.Task.Header.TaskCreateAt, 0)
	timeStr := createTime.Format("2006-01-02 15:04:05")
	title = title + "【创建时间 " + timeStr + "】"

	//店铺名称
	title = title + "【" + golabl.Task.Header.ShopName + "】"

	//任务 id
	title = title + golabl.Task.Header.TaskId
	tool.SetConsoleTitle(title)
}
