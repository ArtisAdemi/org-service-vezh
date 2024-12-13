package middleware

import "gorm.io/gorm"

// Note: Do not change here
// Reference: communihub/app/orgss/orgs/schema.go
type UserOrgRole struct {
	UserID int `gorm:"foreignKey:ID"`
	User   User
	OrgID  int
	RoleID int `gorm:"foreignKey:ID"`
	Role   Role
	Status string
}

// Note: Do not change here
// Reference: communihub/core/users/schema.go
type User struct {
	gorm.Model
}

// Note: Do not change here
// Reference: communihub/core/authorization/roles/schema.go
type Role struct {
	gorm.Model
}

// Note: Do not change here
// Reference: communihub/app/orgs/orgs/schema.go
type Org struct {
	gorm.Model
}
