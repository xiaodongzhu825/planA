package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
)

func main() {
	r := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 0})
	ctx := context.Background()

	// 检查最新上下架任务的 body_over
	tid := "2047906786013282306"
	cnt, _ := r.LLen(ctx, tid+":body_over").Result()
	fmt.Printf("body_over count: %d\n", cnt)

	for i := int64(0); i < cnt; i++ {
		val, err := r.LIndex(ctx, tid+":body_over", i).Result()
		if err != nil {
			fmt.Printf("  [%d] error: %v\n", i, err)
			continue
		}
		var item map[string]interface{}
		json.Unmarshal([]byte(val), &item)

		bi, _ := item["book_info"].(map[string]interface{})
		isbn, _ := bi["isbn"].(string)
		detail, _ := item["detail"].(map[string]interface{})

		errMsg, _ := detail["error"].(string)
		status, _ := detail["status"].(float64)
		price, _ := detail["price"].(float64)
		skuID, _ := detail["sku_id"].(float64)
		goodsID, _ := detail["goods_id"].(float64)

		fmt.Printf("  [%d] isbn=%s status=%.0f price=%.0f sku_id=%.0f goods_id=%.0f error=%s\n",
			i, isbn, status, price, skuID, goodsID, truncate(errMsg, 200))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
