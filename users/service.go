package users

import (
	"fmt"
	orgsvc "org-service/org"

	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

const (
	UserStatusActive   = string("active")
	UserStatusInactive = string("inactive")
	UserStatusInvited  = string("invited")
	UserStatusPending  = string("pending")
	UserStatusReject   = string("rejected")
)

type userApi struct {
	db *gorm.DB
	dialer *gomail.Dialer
	uiAppUrl string
}

type UserAPI interface {
	GetUser(req *GetUserRequest) (*GetUserResponse, error)
	GetUsers(req *IDRequest) (*GetUsersResponse, error)
	ChangeUserRole(req *ChangeUserRoleRequest) (*StatusResponse, error)
	ChangeUserStatus(req *ChangeUserStatusRequest) (*StatusResponse, error)
}

func NewUserService(db *gorm.DB, dialer *gomail.Dialer, uiAppUrl string) UserAPI {
	return &userApi{
		db: db,
		dialer: dialer,
		uiAppUrl: uiAppUrl,
	}
}

func (s *userApi) GetUsers(req *IDRequest) (*GetUsersResponse, error) {
	if req.OrgID == 0 {
		return nil, fmt.Errorf("orgId is required")
	}

	users := []*User{}
	if err := s.db.Where("org_id = ?", req.OrgID).Find(&users).Error; err != nil {
		return nil, err
	}

	return &GetUsersResponse{Users: users}, nil
}

func (s *userApi) GetUser(req *GetUserRequest) (*GetUserResponse, error) {
	if req.OrgID == 0 {
		return nil, fmt.Errorf("orgId is required")
	}

	user := &User{}
	if err := s.db.Where("org_id = ? AND email = ?", req.OrgID, req.Email).First(&user).Error; err != nil {
		return nil, err
	}

	return &GetUserResponse{User: user}, nil
}

// NOTE: This API Endpoint FOR NOW will be used only to change the admin to owner and vice-versa!
// @Summary      	ChangeUserRole
// @Description	Validates org id and user id, and new role id, will query DB in users for user by user id, then tries to change the role from admin to owner or vice-versa.
// @Tags			Users
// @Produce			json
// @Param			Authorization					header		string			true	"Authorization Key(e.g Bearer key)"
// @Param			orgId							path		int				true	"Org ID"
// @Param			ChangeUserRoleRequest	body		ChangeUserRoleRequest	true	"ChangeUserRoleRequest"
// @Success		200								{object}	StatusResponse
// @Router			/o/{orgId}/users/change-user-role	[PUT]
func (s *userApi) ChangeUserRole(req *ChangeUserRoleRequest) (res *StatusResponse, err error) {
	if req.OrgID == 0 {
		return nil, fmt.Errorf("orgId is required")
	}

	if req.UserID == 0 {
		return nil, fmt.Errorf("userId is required")
	}

	if req.NewRoleID == 0 {
		return nil, fmt.Errorf("newRoleId is required")
	}

	if req.NewRoleID != 1 && req.NewRoleID != 2 {
		return nil, fmt.Errorf("invalid roleId")
	}

	var user User 
	result := s.db.Table(UserTableName).Where("id = ?", req.UserID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	var userOrgRole orgsvc.UserOrgRole
	result = s.db.Where("user_id = ? AND org_id = ?", req.UserID, req.OrgID).First(&userOrgRole)
	if result.Error != nil {
		fmt.Errorf("userOrgRole not found")
	}
	if userOrgRole.RoleID != 1 && userOrgRole.RoleID != 2 {
		return nil, fmt.Errorf("user role is not valid")
	}

	if userOrgRole.RoleID == req.NewRoleID {
		return nil, fmt.Errorf("user role is already set to the new role")
	}

	userOrgRole.RoleID = req.NewRoleID
	result = s.db.Model(&userOrgRole).Where("org_id = ? AND user_id = ?", userOrgRole.OrgID, userOrgRole.UserID).Updates(&userOrgRole)
	if result.Error != nil {
		return nil, result.Error
	}

	return &StatusResponse{Status: true}, nil
}

// @Summary      	ChangeUserStatus
// @Description	Validates org id and user id, and status, will try to find user by user id, then tries to change the status.
// @Tags			Users
// @Produce			json
// @Param			Authorization						header		string			true	"Authorization Key(e.g Bearer key)"
// @Param			orgId								path		int				true	"Org ID"
// @Param			ChangeUserStatusRequest	body		ChangeUserStatusRequest	true	"ChangeUserStatusRequest"
// @Success			200									{object}	StatusResponse
// @Router			/o/{orgId}/users/change-user-status	[PUT]
func (s *userApi) ChangeUserStatus(req *ChangeUserStatusRequest) (res *StatusResponse, err error) {
	if req.OrgID == 0 {
		return nil, fmt.Errorf("orgId is required")
	}

	if req.UserID == 0 {
		return nil, fmt.Errorf("userId is required")
	}

	if req.Status != UserStatusActive && req.Status != UserStatusInactive && req.Status != UserStatusReject {
		return nil, fmt.Errorf("invalid status")
	}

	var user User
	result := s.db.Table(UserTableName).Where("id = ?", req.UserID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	var org orgsvc.Org
	result = s.db.Table(orgsvc.OrgTableName).Where("id = ?", req.OrgID).First(&org)
	if result.Error != nil {
		return nil, result.Error
	}

	var userOrgRole orgsvc.UserOrgRole
	result = s.db.Where("org_id = ? AND user_id = ?", req.OrgID, req.UserID).First(&userOrgRole)
	if result.Error != nil {
		return nil, result.Error
	}

	sendApprovedUserEmail := false
	sendRejectUserEmail := false
	if userOrgRole.Status == UserStatusPending && req.Status == UserStatusActive {
		sendApprovedUserEmail = true
	}
	if userOrgRole.Status == UserStatusPending && req.Status == UserStatusReject {
		sendRejectUserEmail = true
	}

	if userOrgRole.Status == req.Status {
		return nil, fmt.Errorf("user has already this status")
	}

	userActive := user.Active
	userActive = true
	if req.Status == UserStatusInactive || req.Status == UserStatusReject {
		userActive = false
	}

	user.Active = userActive
	result = s.db.Table(UserTableName).Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	userOrgRole.Status = req.Status
	result = s.db.Model(&userOrgRole).Where("org_id = ? AND user_id = ?", userOrgRole.OrgID, userOrgRole.UserID).Updates(&userOrgRole)
	if result.Error != nil {
		return nil, result.Error
	}

	if sendApprovedUserEmail {
		orgLink := s.uiAppUrl + "/o/" + org.Slug

		m := gomail.NewMessage()	
		m.SetHeader("From", "influxoks@gmail.com")
		m.SetHeader("To", user.Email)
		m.SetHeader("Subject", "Your account has been approved")
		m.SetBody("text/html", fmt.Sprintf(`Hello from Vezhguesi!<br/><br/>
			Congratulations! You have now been approved by the Organization administrator to join %s!<br/><br/>

			<a href='%s'>Explore Organization</a><br/><br/>

			Thank you, <br/>
			Vezhguesi Team
		`, org.Name, orgLink))

		err = s.dialer.DialAndSend(m)
		if err != nil {
			return nil, err
		}
	}

	if sendRejectUserEmail {
		m := gomail.NewMessage()	
		m.SetHeader("From", "influxoks@gmail.com")
		m.SetHeader("To", user.Email)
		m.SetHeader("Subject", "Your account has been rejected")
		m.SetBody("text/html", fmt.Sprintf(`Hello from Vezhguesi!<br/><br/>
			Unfortunately, your account has been rejected by the Organization administrator in %s.<br/><br/>
		`, org.Name))

		if err != nil {
			return nil, err
		}
	}

	return &StatusResponse{Status: true}, nil
}

