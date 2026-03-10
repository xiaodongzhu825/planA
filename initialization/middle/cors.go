package middle

import (
	"net/http"
)

// Cors 跨域中间件
func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		// 处理OPTIONS请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent) // 204
			return
		}

		// 处理正常请求
		next.ServeHTTP(w, r)
	})
}
