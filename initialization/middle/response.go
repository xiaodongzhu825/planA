package middle

import "net/http"

func Response(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 全局设置 Content-Type 为 application/json
		w.Header().Set("Content-Type", "application/json")
		// 调用下一个处理函数（核心业务逻辑）
		next.ServeHTTP(w, r)
	})
}
