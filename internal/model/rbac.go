package model

import "gorm.io/gorm"

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex;size:64"`
	Password string `gorm:"size:128"`
	Roles    []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
	ID          uint         `gorm:"primaryKey"`
	Name        string       `gorm:"uniqueIndex;size:64"`
	Description string       `gorm:"size:256"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}
type Permission struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	CreatedAt   int64          `json:"created_at"`
	UpdatedAt   int64          `json:"updated_at"`
	Roles       []*Role        `gorm:"many2many:role_permissions;" json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func DeleteUserWithRelations(tx *gorm.DB, userID uint) error {
	var user User
	if err := tx.Preload("Roles").First(&user, userID).Error; err != nil {
		return err
	}
	if err := tx.Model(&user).Association("Roles").Clear(); err != nil {
		return err
	}
	return tx.Delete(&user).Error
}
