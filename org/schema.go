package org

import (
	usersvc "org-service/users"
	"time"
)

type Org struct {
	ID            int            `gorm:"primaryKey"`
	Name          string         `gorm:"not null"`
	Slug          string         `gorm:"unique;not null"`
	Size          string         `gorm:"not null"`
	UserID        int            `gorm:"not null"`
	User          usersvc.User   `gorm:"foreignKey:UserID"`
	CreatedAt     time.Time
	UpdatedAt     *time.Time
	DeletedAt     *time.Time
}