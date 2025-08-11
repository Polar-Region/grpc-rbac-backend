package rbac

import (
	"context"
	"errors"
	"grpc-rbac-backend/api"
	_ "grpc-rbac-backend/internal/middleware"
	"grpc-rbac-backend/internal/model"
	"grpc-rbac-backend/internal/utils"

	"gorm.io/gorm"
)

type Service struct {
	api.UnimplementedRBACServiceServer
}

func NewRBACService() *Service {
	return &Service{}
}

// Login 登录校验
func (s *Service) Login(_ context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	var user model.User
	if err := model.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名不存在")
		}
		return nil, err
	}
	if user.Password != req.Password {
		return nil, errors.New("密码错误")
	}

	// 获取角色列表
	var roles []model.Role
	if err := model.DB.Model(&user).Association("Roles").Find(&roles); err != nil {
		return nil, err
	}
	roleNames := make([]string, 0)
	for _, r := range roles {
		roleNames = append(roleNames, r.Name)
	}

	token, err := utils.GenerateJWT(user.Username, roleNames)
	if err != nil {
		return nil, err
	}
	return &api.LoginResponse{Token: token}, nil
}

// GetUserRoles 查询角色
func (s *Service) GetUserRoles(_ context.Context, req *api.GetUserRolesRequest) (*api.GetUserRolesResponse, error) {
	var user model.User
	if err := model.DB.Where("username = ?", req.UserId).Preload("Roles").First(&user).Error; err != nil {
		return nil, err
	}
	roleNames := make([]string, 0)
	for _, r := range user.Roles {
		roleNames = append(roleNames, r.Name)
	}
	return &api.GetUserRolesResponse{Roles: roleNames}, nil
}

// CheckPermission 校验权限
func (s *Service) CheckPermission(_ context.Context, req *api.CheckPermissionRequest) (*api.CheckPermissionResponse, error) {
	var user model.User
	if err := model.DB.Where("ID = ?", req.UserId).Preload("Roles.Permissions").First(&user).Error; err != nil {
		return nil, err
	}

	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Name == req.Permission {
				return &api.CheckPermissionResponse{Allowed: true}, nil
			}
		}
	}
	return &api.CheckPermissionResponse{Allowed: false}, nil
}

// Register 注册
func (s *Service) Register(ctx context.Context, req *api.RegisterRequest) (*api.RegisterResponse, error) {
	// 1. 检查用户是否已存在
	var count int64
	if err := model.DB.Model(&model.User{}).
		Where("username = ?", req.Username).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 2. 查找默认角色（user）
	var userRole model.Role
	if err := model.DB.Where("name = ?", "user").First(&userRole).Error; err != nil {
		return nil, errors.New("默认角色不存在，请初始化数据库")
	}

	// 3. 创建用户并关联角色
	user := model.User{
		Username: req.Username,
		Password: req.Password, // 实际生产应 bcrypt 加密
		Roles:    []model.Role{userRole},
	}
	if err := model.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &api.RegisterResponse{
		Message: "注册成功",
	}, nil
}

func (s *Service) ListUsers(ctx context.Context, req *api.ListUsersRequest) (*api.ListUsersResponse, error) {
	var users []model.User
	if err := model.DB.Preload("Roles").Find(&users).Error; err != nil {
		return nil, err
	}

	var userInfos []*api.UserInfo
	for _, user := range users {
		var roleNames []string
		for _, r := range user.Roles {
			roleNames = append(roleNames, r.Name)
		}
		userInfos = append(userInfos, &api.UserInfo{
			Username: user.Username,
			Roles:    roleNames,
		})
	}

	return &api.ListUsersResponse{Users: userInfos}, nil
}

func (s *Service) CreatePermission(ctx context.Context, req *api.CreatePermissionRequest) (*api.CreatePermissionResponse, error) {
	p := model.Permission{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := model.DB.Create(&p).Error; err != nil {
		return nil, err
	}
	return &api.CreatePermissionResponse{Id: uint32(p.ID)}, nil
}

func (s *Service) ListPermissions(ctx context.Context, req *api.ListPermissionsRequest) (*api.ListPermissionsResponse, error) {
	var perms []model.Permission
	if err := model.DB.Find(&perms).Error; err != nil {
		return nil, err
	}
	var permInfos []*api.PermissionInfo
	for _, p := range perms {
		permInfos = append(permInfos, &api.PermissionInfo{
			Id:          uint32(p.ID),
			Name:        p.Name,
			Description: p.Description,
		})
	}
	return &api.ListPermissionsResponse{Permissions: permInfos}, nil
}

func (s *Service) CreateRole(ctx context.Context, req *api.CreateRoleRequest) (*api.CreateRoleResponse, error) {
	role := model.Role{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := model.DB.Create(&role).Error; err != nil {
		return nil, err
	}
	return &api.CreateRoleResponse{
		Message: "角色创建成功",
		RoleId:  uint32(role.ID),
	}, nil
}

func (s *Service) AssignPermissions(ctx context.Context, req *api.AssignPermissionsRequest) (*api.AssignPermissionsResponse, error) {
	var role model.Role
	if err := model.DB.Preload("Permissions").First(&role, req.RoleId).Error; err != nil {
		return nil, err
	}

	var permissions []model.Permission
	if err := model.DB.Where("id IN ?", req.PermissionIds).Find(&permissions).Error; err != nil {
		return nil, err
	}

	if err := model.DB.Model(&role).Association("Permissions").Replace(&permissions); err != nil {
		return nil, err
	}

	return &api.AssignPermissionsResponse{Message: "权限分配成功"}, nil
}

func (s *Service) GetRolePermissions(ctx context.Context, req *api.GetRolePermissionsRequest) (*api.GetRolePermissionsResponse, error) {
	var role model.Role
	if err := model.DB.Preload("Permissions").First(&role, req.RoleId).Error; err != nil {
		return nil, err
	}

	permissions := make([]*api.PermissionInfo, 0, len(role.Permissions))
	for _, perm := range role.Permissions {
		permissions = append(permissions, &api.PermissionInfo{
			Id:   uint32(perm.ID),
			Name: perm.Name,
		})
	}

	return &api.GetRolePermissionsResponse{Permissions: permissions}, nil
}

func (s *Service) CreateUser(ctx context.Context, req *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	user := model.User{
		Username: req.Username,
		Password: req.Password, // 生产环境需加密
	}
	if err := model.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &api.CreateUserResponse{
		Message: "用户创建成功",
		UserId:  uint32(user.ID),
	}, nil
}

func (s *Service) UpdateUser(ctx context.Context, req *api.UpdateUserRequest) (*api.UpdateUserResponse, error) {
	var user model.User
	if err := model.DB.First(&user, req.UserId).Error; err != nil {
		return nil, err
	}
	user.Username = req.Username
	if req.Password != "" {
		user.Password = req.Password
	}
	if err := model.DB.Save(&user).Error; err != nil {
		return nil, err
	}
	return &api.UpdateUserResponse{Message: "用户更新成功"}, nil
}

func (s *Service) DeleteUser(ctx context.Context, req *api.DeleteUserRequest) (*api.DeleteUserResponse, error) {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		return model.DeleteUserWithRelations(tx, uint(req.UserId))
	})
	if err != nil {
		return nil, err
	}
	return &api.DeleteUserResponse{Message: "用户删除成功"}, nil
}

func (s *Service) GetUser(ctx context.Context, req *api.GetUserRequest) (*api.GetUserResponse, error) {
	var user model.User
	if err := model.DB.Preload("Roles").First(&user, req.UserId).Error; err != nil {
		return nil, err
	}
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Name
	}
	return &api.GetUserResponse{
		Username: user.Username,
		Roles:    roles,
	}, nil
}
