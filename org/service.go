package org

import (
	"fmt"

	"gorm.io/gorm"
)

type OrgService struct {
	db *gorm.DB
}

func NewOrgService(db *gorm.DB) *OrgService {
	return &OrgService{db: db}
}

func (s *OrgService) AddOrg(name, size string) (*Org, error) {
	org := &Org{Name: name, Size: size}
	if err := s.db.Create(org).Error; err != nil {
		return nil, fmt.Errorf("failed to create org: %w", err)
	}
	return org, nil
}