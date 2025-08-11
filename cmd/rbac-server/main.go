package main

import (
	"fmt"
	"grpc-rbac-backend/config"
	"grpc-rbac-backend/internal/model"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"grpc-rbac-backend/api"
	"grpc-rbac-backend/internal/middleware"
	"grpc-rbac-backend/internal/rbac"
)

func registerService(consulClient *consulapi.Client, serviceID, serviceName string, port int) error {
	check := &consulapi.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("127.0.0.1:%d/%s", port, serviceName),
		Interval:                       "10s",
		DeregisterCriticalServiceAfter: "1m",
	}
	reg := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: "127.0.0.1",
		Port:    port,
		Check:   check,
	}

	err := consulClient.Agent().ServiceRegister(reg)
	if err != nil {
		log.Printf("❌ 注册服务失败: %v", err)
	} else {
		log.Println("✅ 注册服务成功")
	}
	return err
}

func deregisterService(consulClient *consulapi.Client, serviceID string) {
	if err := consulClient.Agent().ServiceDeregister(serviceID); err != nil {
		log.Printf("❌ 注销服务失败: %v", err)
	} else {
		log.Println("✅ 服务已从 Consul 注销")
	}
}

func main() {
	// 加载配置
	cfg := config.Load()

	// ✅ 初始化数据库连接 + 自动建表 + 自动创建admin
	model.InitDB(cfg.MysqlDsn, cfg.AdminUsername, cfg.AdminPassword)

	const (
		port        = 50051
		serviceID   = "rbac-service-1"
		serviceName = "rbac-service"
	)

	// 初始化 Consul 客户端
	consulClient, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		log.Fatalf("❌ 创建 Consul 客户端失败: %v", err)
	}

	// 监听 TCP 地址
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Fatalf("❌ 监听失败: %v", err)
	}
	log.Printf("✅ 服务监听地址: %s", lis.Addr().String())

	// 创建 gRPC Server，带认证中间件
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.AuthInterceptor),
	)

	// 注册 RBAC 业务服务
	rbacService := rbac.NewRBACService()
	api.RegisterRBACServiceServer(grpcServer, rbacService)

	// 注册健康检查服务
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus(serviceName, healthpb.HealthCheckResponse_SERVING)

	// 开启 gRPC 反射服务，方便调试
	reflection.Register(grpcServer)

	// 注册服务到 Consul
	if err := registerService(consulClient, serviceID, serviceName, port); err != nil {
		log.Fatalf("❌ 服务注册失败: %v", err)
	}

	// 启动 gRPC 服务
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("❌ 服务启动失败: %v", err)
		}
	}()
	log.Printf("🚀 RBAC gRPC 服务启动成功，监听端口: %d", port)

	// 等待系统信号优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("⚠️ 捕获退出信号：%v，准备关闭服务", sig)

	// 注销服务
	deregisterService(consulClient, serviceID)

	// 优雅停止 gRPC 服务
	grpcServer.GracefulStop()
	log.Println("✅ 服务已优雅停止")
}
