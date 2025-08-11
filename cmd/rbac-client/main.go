package main

import (
	"context"
	"grpc-rbac-backend/config"
	"log"
	"time"

	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"grpc-rbac-backend/api"

	"google.golang.org/grpc"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 用 IPv4 地址连接，避免使用 localhost 或 [::1]
	conn, err := grpc.NewClient("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接服务: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	client := api.NewRBACServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// 1. 登录获取 token
	loginResp, err := client.Login(ctx, &api.LoginRequest{
		Username: cfg.AdminUsername,
		Password: cfg.AdminPassword,
	})
	if err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	token := loginResp.Token
	log.Printf("登录成功，token: %s", token)

	// 2. 构造带 token 的 Metadata 上下文
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	// 3. 调用 GetUserRoles
	rolesResp, err := client.GetUserRoles(ctxWithToken, &api.GetUserRolesRequest{UserId: cfg.AdminUsername})
	if err != nil {
		log.Fatalf("调用 GetUserRoles 失败: %v", err)
	}
	log.Printf("用户角色: %v", rolesResp.Roles)

	// 4. 调用 CheckPermission
	permResp, err := client.CheckPermission(ctxWithToken, &api.CheckPermissionRequest{
		UserId:     "1",
		Permission: "write",
	})
	if err != nil {
		log.Fatalf("调用 CheckPermission 失败: %v", err)
	}
	log.Printf("是否有权限写入: %v", permResp.Allowed)
}
