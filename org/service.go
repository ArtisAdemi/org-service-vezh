package org

import (
	"fmt"
	usersvc "org-service/users"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

type orgApi struct {
	db *gorm.DB
	logger log.AllLogger
}

type OrgAPI interface {
	AddOrg(req *AddOrgRequest) (res *OrgResponse, err error)
	GetOrgs(req *IDRequest) (res *GetOrgsResponse, err error)
}

func NewOrgService(db *gorm.DB, logger log.AllLogger) OrgAPI {
	return &orgApi{db: db, logger: logger}
}

func (s *orgApi) AddOrg(req *AddOrgRequest) (res *OrgResponse, err error) {
	if req.UserID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Size == "" {
		return nil, fmt.Errorf("size is required")
	}

	var user usersvc.User

	if err := s.db.First(&user, req.UserID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var org Org
	orgSlug := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(req.Name, "")
	orgSlug = strings.ToLower(orgSlug)
	orgSlug = strings.ReplaceAll(strings.TrimSpace(orgSlug), " ", "-")
	s.db.Where("slug = ?", orgSlug).First(&org)
	if org.ID != 0 {
		return nil, fmt.Errorf("org slug already exists")
	}

	s.db.Where("name = ?", req.Name).First(&org)
	if org.ID != 0 {
		return nil, fmt.Errorf("org name already exists")
	}

	newOrg := &Org{
		Name: req.Name,
		Size: req.Size,
		Slug: orgSlug,
		UserID: user.ID,
	}

	result := s.db.Create(&newOrg)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create org: %w", result.Error)
	}

	return &OrgResponse{
		ID:   newOrg.ID,
		Name: newOrg.Name,
		Size: newOrg.Size,
		Slug: newOrg.Slug,
	}, nil
}

func (s *orgApi) GetOrgs(req *IDRequest) (res *GetOrgsResponse, err error) {
	if req.UserID == 0 {
		return nil, fmt.Errorf("user id is required")
	}

	orgs := []Org{}

	result := s.db.Find(&orgs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get orgs: %w", result.Error)
	}

	orgResponses := []OrgResponse{}
	for _, org := range orgs {
		orgResponses = append(orgResponses, OrgResponse{
			ID:   org.ID,
			Name: org.Name,
			Size: org.Size,
			Slug: org.Slug,
		})
	}

	return &GetOrgsResponse{
		Orgs: orgResponses,
	}, nil
}
