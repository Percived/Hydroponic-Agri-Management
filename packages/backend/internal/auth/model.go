package auth

import "time"

const (
	RoleAdmin    = "ADMIN"
	RoleOperator = "OPERATOR"
	RoleViewer   = "VIEWER"

	UserStatusEnabled  = "ENABLED"
	UserStatusDisabled = "DISABLED"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"size:32;uniqueIndex;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	Nickname     string    `gorm:"size:64"`
	Phone        string    `gorm:"size:32"`
	Email        string    `gorm:"size:64"`
	Status       string    `gorm:"size:16;default:ENABLED"`
	CreatedAt    time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime:milli"`
	Roles        []Role    `gorm:"many2many:user_roles"`
}

func (User) TableName() string { return "users" }

type Role struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	Name        string    `gorm:"size:32;uniqueIndex;not null"`
	Description string    `gorm:"size:64"`
	CreatedAt   time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:milli"`
}

func (Role) TableName() string { return "roles" }

type UserRole struct {
	ID     uint64 `gorm:"primaryKey;autoIncrement"`
	UserID uint64 `gorm:"not null"`
	RoleID uint64 `gorm:"not null"`
}

func (UserRole) TableName() string { return "user_roles" }
