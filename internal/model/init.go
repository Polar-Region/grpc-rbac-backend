package model

import (
	"errors"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dsn string, adminUsername string, adminPassword string) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	DB = db

	// 自动迁移所有模型
	err = db.AutoMigrate(&User{}, &Role{}, &Permission{})
	if err != nil {
		log.Fatalf("❌ 自动迁移失败: %v", err)
	}

	log.Println("✅ 数据库连接成功，模型迁移完成")

	// 初始化数据
	initAdminRoleAndUser(db, adminUsername, adminPassword)
}

func initAdminRoleAndUser(db *gorm.DB, adminUsername string, adminPassword string) {
	// 检查是否已存在 admin 角色
	var adminRole Role
	if err := db.First(&adminRole, "name = ?", "admin").Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有找到 admin 角色，创建一个新的
			adminRole = Role{
				Name:        "admin",
				Description: "Administrator with full access",
			}
			if err := db.Create(&adminRole).Error; err != nil {
				log.Fatalf("❌ 创建 admin 角色失败: %v", err)
			}
			log.Println("✅ 创建 admin 角色成功")
		} else {
			log.Fatalf("❌ 查询 admin 角色失败: %v", err)
		}
	}

	// 检查是否已存在 write 权限
	var writePermission Permission
	if err := db.First(&writePermission, "name = ?", "write").Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有找到 write 权限，创建一个新的
			writePermission = Permission{
				Name:        "write",
				Description: "write blogs",
			}
			if err := db.Create(&writePermission).Error; err != nil {
				log.Fatalf("\"❌ 创建 write 权限失败: %v\"", err)
			}
			log.Println("✅ 创建 write 权限成功")
		} else {
			log.Fatalf("❌ 查询 write 权限失败: %v", err)
		}
	}

	// 检查并为用户分配 admin 角色
	if err := db.Model(&adminRole).Association("Permissions").Append(&writePermission); err != nil {
		log.Fatalf("❌ 为 admin 角色分配 write 权限失败: %v", err)
	}
	log.Println("✅ 为 admin 角色分配 write 权限成功")

	// 检查是否已存在管理员用户
	var adminUser User
	if err := db.First(&adminUser, "username = ?", adminUsername).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有找到管理员用户，创建一个新的
			adminUser = User{
				Username: adminUsername,
				Password: adminPassword, // 使用合适的密码
			}
			if err := db.Create(&adminUser).Error; err != nil {
				log.Fatalf("❌ 创建管理员用户失败: %v", err)
			}
			log.Println("✅ 创建管理员用户成功")
		} else {
			log.Fatalf("❌ 查询管理员用户失败: %v", err)
		}
	}

	// 检查并为用户分配 admin 角色
	if err := db.Model(&adminUser).Association("Roles").Append(&adminRole); err != nil {
		log.Fatalf("❌ 为管理员用户分配 admin 角色失败: %v", err)
	}
	log.Println("✅ 为管理员用户分配 admin 角色成功")
}
