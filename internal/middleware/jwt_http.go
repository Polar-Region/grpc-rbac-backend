package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte("secret123") // 和 gRPC 中的一致

// JWTClaims 用于解码的结构
type JWTClaims struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTAuthMiddleware 用于 REST API 的中间件
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 白名单：登录、注册接口放行
		if r.URL.Path == "/v1/login" || r.URL.Path == "/v1/register" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Token 验证成功，注入上下文（可选）
		claims := token.Claims.(*JWTClaims)
		ctx := context.WithValue(r.Context(), "user", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
