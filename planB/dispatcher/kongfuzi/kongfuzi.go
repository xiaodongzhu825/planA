package kongfuzi

import _type "planA/planB/type"

type KongFuzi struct {
}

// NewKongfuzi 创建孔子平台
func NewKongfuzi() *KongFuzi {
	return &KongFuzi{}
}
func (kongFuzi *KongFuzi) AddGoodsTask(pddConfig _type.PddConfig, header interface{}, bodyWait _type.TaskBody) (string, error) {
	return "孔夫子商品添加任务", nil
}
func (kongFuzi *KongFuzi) SetGoodsTask() string {
	return "孔夫子商品修改任务"

}

func (kongFuzi *KongFuzi) GetGoodsTask() string {
	return "孔夫子商品获取任务"
}

func (kongFuzi *KongFuzi) DelGoodsTask() string {
	return "孔夫子商品删除任务"
}
