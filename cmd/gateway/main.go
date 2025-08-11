package main

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"grpc-rbac-backend/api"
	"grpc-rbac-backend/internal/middleware" // 导入 JWT 中间件
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 创建 gRPC-Gateway 的 mux
	gwMux := runtime.NewServeMux()

	// gRPC 连接配置
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// 注册 gRPC 服务到 Gateway
	err := api.RegisterRBACServiceHandlerFromEndpoint(ctx, gwMux, "127.0.0.1:50051", opts)
	if err != nil {
		log.Fatalf("❌ 注册 gRPC Gateway 失败: %v", err)
	}

	// 包裹 JWT 中间件
	handlerWithMiddleware := middleware.JWTAuthMiddleware(gwMux)

	log.Println("🚀 HTTP 网关启动成功，监听 http://localhost:8080")
	if err := http.ListenAndServe(":8080", handlerWithMiddleware); err != nil {
		log.Fatalf("❌ HTTP 服务启动失败: %v", err)
	}
}
