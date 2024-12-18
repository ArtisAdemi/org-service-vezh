package org

import (
	"fmt"
	"org-service/helper"

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
	FindMyOrgs(req *IDRequest) (res []*OrgWithRole, err error)
	GetOrgMembers(req *OrgRequest) (res *OrgMembersResponse, err error)
}

func NewOrgService(db *gorm.DB, logger log.AllLogger) OrgAPI {
	return &orgApi{db: db, logger: logger}
}



// @Summary      	Add Org
// @Description		Validates user id, org name and org size, checks if org exists in DB by name or slug, if not a new organization with trial subscription will be created and then the created ID will be returned.
// @Tags			Orgs
// @Accept			json
// @Produce			json
// @Param			Authorization					header		string			true	"Authorization Key(e.g Bearer key)"
// @Param			AddOrgRequest					body		AddOrgRequest	true	"AddOrgRequest"
// @Success			200								{object}	OrgResponse
// @Router           /api/orgs                        [POST]	
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

	var user User

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
	}

	result := s.db.Create(&newOrg)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create org: %w", result.Error)
	}

	var ownerRole Role
	if err := s.db.Where("name = ?", helper.OwnerRoleName).First(&ownerRole).Error; err != nil {
		return nil, fmt.Errorf("failed to get owner role: %w", err)
	}

	var userOrgRole UserOrgRole
	userOrgRole.OrgID = newOrg.ID
	userOrgRole.UserID = user.ID
	userOrgRole.RoleID = int(ownerRole.ID)
	userOrgRole.Status = "active"
	result = s.db.Table(UserOrgRoleTableName).Create(&userOrgRole)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user org role: %w", result.Error)
	}

	return &OrgResponse{
		ID:   newOrg.ID,
		Name: newOrg.Name,
		Slug: newOrg.Slug,
	}, nil
}

// @Summary      	FindMyOrgs
// @Description		Validates user is, will query DB the orgs that current user is linked to and then returns them in JSON.
// @Tags			Orgs
// @Produce			json
// @Param			Authorization					header		string			true	"Authorization Key(e.g Bearer key)"
// @Success			200								{array}	OrgWithRole
// @Router			/api/orgs/me			[GET]
func (s *orgApi) FindMyOrgs(req *IDRequest) (res []*OrgWithRole, err error) {
	if req.UserID == 0 {
		return nil, fmt.Errorf("user id is required")
	}

	rows, err := s.db.Table(OrgTableName).
	Select("orgs.id", "orgs.name", "orgs.slug", "user_org_roles.role_id", "user_org_roles.user_id").
	Joins("Left JOIN user_org_roles on user_org_roles.org_id = orgs.id").
	Where("user_org_roles.user_id = ?", req.UserID).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query orgs: %w", err)
	}
	defer rows.Close()

	var orgRoles []*OrgWithRole

	for rows.Next() {
		onr := &OrgWithRole{}
		err := rows.Scan(&onr.OrgID, &onr.Name, &onr.Slug, &onr.RoleID, &onr.UserID)
		if err != nil {
			fmt.Println("error scanning row", err)
		}
		orgRoles = append(orgRoles, onr)
	}

	return orgRoles, nil
}


// @Summary      	GetOrgMembers
// @Description		Validates user is, will query DB the orgs that current user is linked to and then returns them in JSON.
// @Tags			Orgs
// @Produce			json
// @Param			Authorization					header		string			true	"Authorization Key(e.g Bearer key)"
// @Param			orgId							path		int			true	"OrgID"
// @Success			200								{object}	OrgMembersResponse
// @Router			/api/o/{orgId}/members		[GET]
func (s *orgApi) GetOrgMembers(req *OrgRequest) (res *OrgMembersResponse, err error) {
	if req.UserID == 0 {
		return nil, fmt.Errorf("user id is required")
	}

	// Check if user is a member of the org
	var userOrgRole UserOrgRole
	
	if req.OrgID == 0 {
		return nil , fmt.Errorf("org id is required")
	} 
    
	result := s.db.Table(UserOrgRoleTableName).
		Where("user_id = ?", req.UserID).
		Where("org_id = ?", req.OrgID).
		First(&userOrgRole)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user org role: %w", result.Error)
	}

	if userOrgRole.UserID == 0 {
		return nil, fmt.Errorf("user is not a member of the org")
	}


	// Get all users in the org
	var users []User
	result = s.db.Table("users").
	        Joins("LEFT JOIN user_org_roles AS uor on uor.user_id = users.id").
			Where("uor.org_id = ?", userOrgRole.OrgID).
			Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get users: %w", result.Error)
	}

	var userOrgRoles []UserOrgRole
	result = s.db.Table("user_org_roles").
			Where("org_id = ?", userOrgRole.OrgID).
			Find(&userOrgRoles)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user org roles: %w", result.Error)
	}

	var userOrgRoleResponses []UserOrgRoleResponse
	for _, userOrgRole := range userOrgRoles {
		userOrgRoleResponses = append(userOrgRoleResponses, UserOrgRoleResponse{
			UserID: userOrgRole.UserID,
			OrgID:  userOrgRole.OrgID,
			RoleID: userOrgRole.RoleID,
			Status: userOrgRole.Status,
		})
	}

	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: *user.Username,
			FirstName: user.FirstName,
			LastName: user.LastName,
			Status: user.Status,
			AvatarImgKey: user.AvatarImgKey,
			Active: user.Active,
			Phone: user.Phone,
		})
	}

	var orgMembers []OrgMembers
	for i, _ := range users {
		orgMembers = append(orgMembers, OrgMembers{
			UserOrgRole: userOrgRoleResponses[i],
			User:        userResponses[i],
		})
	}

	return &OrgMembersResponse{
		OrgMembers: orgMembers,
	}, nil
}