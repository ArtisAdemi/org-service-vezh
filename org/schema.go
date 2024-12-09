package org

import (
	"time"
)

type Org struct {
	ID            int            `gorm:"primaryKey"`
	Name          string         `gorm:"not null"`
	Slug          string         `gorm:"unique;not null"`
	Size          string         `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     *time.Time
	DeletedAt     *time.Time
}