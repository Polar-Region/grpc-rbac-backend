# gRPC RBAC Backend

一个基于 gRPC 和 HTTP Gateway 的现代化角色权限管理系统（RBAC - Role-Based Access Control），提供完整的用户认证、角色管理和权限控制功能。

## 🚀 功能特性

### 核心功能
- **用户管理**: 用户注册、登录、CRUD 操作
- **角色管理**: 角色创建、权限分配、角色查询
- **权限管理**: 权限创建、权限列表、权限验证
- **认证授权**: JWT Token 认证，基于角色的权限控制
- **服务发现**: 集成 Consul 服务注册与发现

### 技术特性
- **gRPC 服务**: 高性能的 RPC 通信
- **HTTP Gateway**: 自动生成 RESTful API
- **数据库**: MySQL + GORM ORM
- **服务健康检查**: gRPC 健康检查机制
- **API 文档**: 自动生成 Swagger 文档
- **配置管理**: 环境变量配置支持

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │    │   gRPC Client   │    │   Consul UI     │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  HTTP Gateway   │    │   gRPC Server   │    │   Consul Agent  │
│   (Port: 8080)  │    │   (Port: 50051) │    │                 │
└─────────┬───────┘    └─────────┬───────┘    └─────────────────┘
          │                      │
          └──────────┬───────────┘
                     ▼
            ┌─────────────────┐
            │   RBAC Service  │
            └─────────┬───────┘
                      ▼
            ┌─────────────────┐
            │   MySQL DB      │
            └─────────────────┘
```

## 📁 项目结构

```
grpc-rbac-backend/
├── api/                    # 自动生成的 gRPC 和 HTTP 代码
│   ├── rbac_grpc.pb.go    # gRPC 服务定义
│   ├── rbac.pb.go         # Protocol Buffers 消息定义
│   ├── rbac.pb.gw.go      # HTTP Gateway 代码
│   └── swagger/           # Swagger API 文档
├── cmd/                   # 应用程序入口
│   ├── gateway/           # HTTP Gateway 服务
│   ├── rbac-client/       # gRPC 客户端示例
│   └── rbac-server/       # gRPC 服务器
├── config/                # 配置管理
├── internal/              # 内部包
│   ├── middleware/        # 中间件（认证、JWT）
│   ├── model/             # 数据模型
│   ├── rbac/              # RBAC 业务逻辑
│   └── utils/             # 工具函数
├── proto/                 # Protocol Buffers 定义
├── buf.yaml              # Buf 配置
├── buf.gen.yaml          # Buf 代码生成配置
└── go.mod                # Go 模块依赖
```

## 🛠️ 技术栈

- **语言**: Go 1.24+
- **gRPC**: google.golang.org/grpc
- **HTTP Gateway**: grpc-ecosystem/grpc-gateway
- **数据库**: MySQL + GORM
- **认证**: JWT (golang-jwt/jwt)
- **服务发现**: Consul
- **API 文档**: Swagger/OpenAPI
- **配置**: godotenv
- **代码生成**: Buf

## 📋 环境要求

- Go 1.24 或更高版本
- MySQL 5.7+ 或 MySQL 8.0+
- Consul (可选，用于服务发现)
- Buf CLI (用于代码生成)

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone github.com/Polar-Region/gRPC-rbac-backend
cd grpc-rbac-backend
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 安装 Buf CLI

```bash
# macOS
brew install bufbuild/buf/buf

# Windows
scoop install buf

# Linux
curl -sSL \
  "https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-$(uname -s)-$(uname -m)" \
  -o "$(go env GOPATH)/bin/buf" && \
  chmod +x "$(go env GOPATH)/bin/buf"
```

### 4. 生成代码

```bash
buf generate
```

### 5. 配置环境变量

创建 `.env` 文件：

```env
MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/rbac_db?charset=utf8mb4&parseTime=True&loc=Local
ADMIN_USERNAME=admin
ADMIN_PASSWORD=123456
JWT_SECRET=your-secret-key
```

### 6. 启动服务

#### 启动 gRPC 服务器

```bash
go run cmd/rbac-server/main.go
```

#### 启动 HTTP Gateway

```bash
go run cmd/gateway/main.go
```

### 7. 验证服务

- gRPC 服务: `localhost:50051`
- HTTP Gateway: `http://localhost:8080`
- Swagger 文档: `http://localhost:8080/swagger-ui/`

## 📚 API 文档

### 认证相关

#### 用户注册
```http
POST /v1/register
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

#### 用户登录
```http
POST /v1/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

### 用户管理

#### 创建用户
```http
POST /v1/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "newuser",
  "password": "password123"
}
```

#### 获取用户列表
```http
GET /v1/users
Authorization: Bearer <token>
```

#### 获取用户信息
```http
GET /v1/users/{userId}
Authorization: Bearer <token>
```

#### 更新用户
```http
PUT /v1/users/{userId}
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "updateduser",
  "password": "newpassword"
}
```

#### 删除用户
```http
DELETE /v1/users/{userId}
Authorization: Bearer <token>
```

### 角色管理

#### 创建角色
```http
POST /v1/roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "editor",
  "description": "内容编辑角色"
}
```

#### 分配权限给角色
```http
POST /v1/roles/{roleId}/permissions
Authorization: Bearer <token>
Content-Type: application/json

{
  "permissionIds": [1, 2, 3]
}
```

#### 获取角色权限
```http
GET /v1/roles/{roleId}/permissions
Authorization: Bearer <token>
```

### 权限管理

#### 创建权限
```http
POST /v1/permissions
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "read:articles",
  "description": "读取文章权限"
}
```

#### 获取权限列表
```http
GET /v1/permissions
Authorization: Bearer <token>
```

#### 检查用户权限
```http
GET /v1/users/{userId}/permissions/{permission}
Authorization: Bearer <token>
```

## 🔧 开发指南

### 代码生成

当修改 `proto/rbac.proto` 文件后，需要重新生成代码：

```bash
buf generate
```

### 数据库迁移

项目使用 GORM 自动迁移，启动时会自动创建表结构。

### 添加新的 API

1. 在 `proto/rbac.proto` 中定义新的消息和服务
2. 运行 `buf generate` 生成代码
3. 在 `internal/rbac/service.go` 中实现业务逻辑
4. 更新中间件和权限控制

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/rbac
```

## 🐳 Docker 部署

### 构建镜像

```bash
# 构建 gRPC 服务器镜像
docker build -t rbac-server -f Dockerfile.server .

# 构建 Gateway 镜像
docker build -t rbac-gateway -f Dockerfile.gateway .
```

### 使用 Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: rbac_db
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  consul:
    image: consul:latest
    ports:
      - "8500:8500"
    command: consul agent -server -bootstrap-expect=1 -ui -client=0.0.0.0

  rbac-server:
    build:
      context: .
      dockerfile: Dockerfile.server
    environment:
      MYSQL_DSN: root:password@tcp(mysql:3306)/rbac_db?charset=utf8mb4&parseTime=True&loc=Local
    depends_on:
      - mysql
      - consul

  rbac-gateway:
    build:
      context: .
      dockerfile: Dockerfile.gateway
    ports:
      - "8080:8080"
    depends_on:
      - rbac-server

volumes:
  mysql_data:
```

启动服务：

```bash
docker-compose up -d
```

## 🔒 安全考虑

- 使用 JWT 进行身份认证
- 密码加密存储
- 基于角色的权限控制
- 输入验证和清理
- HTTPS 部署（生产环境）

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

**注意**: 这是一个示例项目，生产环境部署前请确保：
- 修改默认密码
- 配置安全的 JWT 密钥
- 启用 HTTPS
- 配置适当的数据库权限
- 设置防火墙规则
