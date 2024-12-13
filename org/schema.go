package org

import (
	"time"

	"gorm.io/gorm"
)

const (
	OrgTableName = "orgs"
	UserOrgRoleTableName = "user_org_roles"
)

type Org struct {
	ID             int           `gorm:"primaryKey"`
	Name           string        `gorm:"not null"`
	Size           string        `gorm:"not null"`
	Slug           string        `gorm:"unique;not null"`
	UserOrgRole    []UserOrgRole `gorm:"foreignKey:OrgID"`
	// SubscriptionID int           `gorm:"foreignKey:ID"`
	// Subscription   Subscription
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
}

type UserOrgRole struct {
	UserID int `gorm:"foreignKey:ID"`
	User   User
	OrgID  int `gorm:"foreignKey:ID"`
	Org    Org
	RoleID int `gorm:"foreignKey:ID"`
	Role   Role
	Status string
}


type Role struct {
	gorm.Model
}
