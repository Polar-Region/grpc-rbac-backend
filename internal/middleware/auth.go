package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"grpc-rbac-backend/internal/utils"
)

type contextKey string

const ContextUserKey = contextKey("user")

func AuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// 登录、注册和健康接口不校验token
	if info.FullMethod == "/rbac.RBACService/Login" ||
		info.FullMethod == "/rbac.RBACService/Register" ||
		info.FullMethod == "/grpc.health.v1.Health/Check" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")

	claims, err := utils.ParseJWT(tokenStr)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// 把解析出的用户信息放到 context，业务接口可以取出来用
	newCtx := context.WithValue(ctx, ContextUserKey, claims)
	return handler(newCtx, req)
}
