package users

import (
	"fmt"
	"org-service/helper"
	orgsvc "org-service/org"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
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
	logger *log.AllLogger
}

type UserAPI interface {
	GetUser(req *GetUserRequest) (*GetUserResponse, error)
	GetUsers(req *IDRequest) (*GetUsersResponse, error)
	ChangeUserRole(req *ChangeUserRoleRequest) (*StatusResponse, error)
	ChangeUserStatus(req *ChangeUserStatusRequest) (*StatusResponse, error)
	InviteUser(req *InviteUserRequest) (*StatusResponse, error)
	AcceptInvitation(req *AcceptInvitationRequest) (*AcceptInvitationResponse, error)
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

// @Summary      	InviteUser
// @Description	Validates email, role ID in request, checks in DB if req email exists with req orgId, if not generates a JWT token, send via email a UI app URL containing the token.
// @Tags			Users
// @Accept			json
// @Produce			json
// @Param			Authorization			header		string	true	"Authorization Key(e.g Bearer key)"
// @Param			orgId					path		int		true	"OrgID"
// @Param			email					path		string	true	"Email"
// @Param			roleId					path		int		true	"RoleID"
// @Success			200						{object}		StatusResponse
// @Router			/api/o/{orgId}/users/invite/{email}/{roleId}	[GET]
func (s *userApi) InviteUser(req *InviteUserRequest) (res *StatusResponse, err error) {
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if req.RoleID == 0 {
		return nil, fmt.Errorf("roleId is required")
	}

	if req.OrgID == 0 {
		return nil, fmt.Errorf("orgId is required")
	}

	if req.CurrentUserID == 0 {
		return nil, fmt.Errorf("currentUserId is required")
	}

	if req.CurrentRoleID == 0 {
		return nil, fmt.Errorf("currentRoleId is required")
	}

	req.Email = strings.TrimSpace(req.Email)

	var userOrgCount int64
	result := s.db.Table(orgsvc.UserOrgRoleTableName).
		Joins("LEFT JOIN users AS u ON u.id=user_org_roles.user_id").
		Where("u.email = ? AND user_org_roles.org_id = ? AND user_org_roles.status = ?", req.Email, req.OrgID, UserStatusActive).
		Count(&userOrgCount)
	if result.Error != nil {
		return nil, result.Error
	}

	if userOrgCount > 0 {
		return nil, fmt.Errorf("user already has an active role in this organization")
	}

	var org orgsvc.Org
	result = s.db.Table(orgsvc.OrgTableName).Where("id = ?", req.OrgID).First(&org)
	if result.Error != nil {
		return nil, result.Error
	}

	var cUser User
	result = s.db.Table(UserTableName).Where("id = ?", req.CurrentUserID).First(&cUser)
	if result.Error != nil {
		return nil, result.Error
	}

	status := UserStatusPending
	active := false

	if req.CurrentRoleID == 1 || req.CurrentRoleID == 2 {
		status = UserStatusActive
		active = true
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	claims["email"] = req.Email
	claims["orgId"] = req.OrgID
	claims["roleId"] = req.RoleID
	claims["status"] = status


	firstName := ""
	lastName := ""
	fullName := firstName + " " + lastName
	if strings.TrimSpace(fullName) == "" {
		fullName = org.Name
	}

	var user User
	// Check if user exists
	result = s.db.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		log.Error("user not found")
	}

	currentUserFullName := cUser.FirstName + " " + cUser.LastName
	claims["currentUserFullName"] = currentUserFullName

	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, err
	}

	if user.ID == 0 {
		// if err := handleTotalUsersLimit(s.db, req.OrgID); err != nil {
		// 	return nil, err
		// }

		// if err := handleAdminRoleLimit(s.db, req.OrgID); err != nil {
		// 	return nil, err
		// }

		// if err := handleAdvisorRoleLimit(s.db, req.OrgID); err != nil {
		// 	return nil, err
		// }

		// Generate hash pw
		pwd := helper.RandomString(8)
		pwh, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		user.Email = req.Email
		user.Password = string(pwh)
		user.Active = active
		user.VerifiedEmail = false

		result = s.db.Omit("UpdatedAt").Create(&user)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	// Prevent relationship with status=invited creation if exists
	var userOrgRoleCount int64
	result = s.db.Table("user_org_roles").
		Joins("LEFT JOIN users ON users.id = user_org_roles.user_id").
		Where("users.id = ? AND user_org_roles.org_id = ? AND user_org_roles.status = ?", 
			user.ID, req.OrgID, UserStatusInvited).
    	Count(&userOrgRoleCount)
	if result.Error != nil {
		return nil, result.Error
	}

	if userOrgRoleCount > 0 {
		return nil, fmt.Errorf("user already has already been invited to this organization")
	}

	if userOrgRoleCount == 0 {
		userOrgRole := orgsvc.UserOrgRole{
			UserID: user.ID,
			OrgID: req.OrgID,
			RoleID: req.RoleID,
			Status: UserStatusInvited,
		}

		result = s.db.Create(&userOrgRole)
		if result.Error != nil {
			return nil, result.Error
		}

	}

	m := gomail.NewMessage()	
		m.SetHeader("From", "influxoks@gmail.com")
		m.SetHeader("To", user.Email)
		m.SetHeader("Subject", "Vezhguesi: You're invited to join " + org.Name)
		m.SetBody("text/html", fmt.Sprintf(`You've received an invitation!<br/><br/>

			%s has invited you to join the Organization %s.<br/>
			In order to access this Organization you must click the link below and continue registration: <br/><br/>

			<a href='%s'>Accept Invitation</a><br/><br/>

			Thank you, <br/>
			Vezhguesi Team
		`, fullName, org.Name, fmt.Sprintf(`%s/accept-invitation/%s`, s.uiAppUrl, t)))

		err = s.dialer.DialAndSend(m)
		if err != nil {
			return nil, err
	}

	return &StatusResponse{Status: true}, nil
}

// @Summary      	InviteAccept
// @Description		Validates token, username, firstName, lastName, password and confirmPassword, then check in DB if user with same email and org is already connected if not creates User, Profile and Org Relationship and returns created user ID in response.
// @Tags			Users
// @Accept			json
// @Produce			json
// @Param			token				path		string				true	"Token"
// @Param			AcceptInvitationRequest	body		AcceptInvitationRequest	true	"AcceptInvitationRequest"
// @Success			200					{object}	AcceptInvitationResponse
// @Router			/api/users/invite/accept/{token}	[POST]
func (s *userApi) AcceptInvitation(req *AcceptInvitationRequest) (res *AcceptInvitationResponse, err error) {
	if req.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	email := fmt.Sprintf("%v", claims["email"])
	orgIdStr := fmt.Sprintf("%v", claims["orgId"])
	roleIdStr := fmt.Sprintf("%v", claims["roleId"])
	status := fmt.Sprintf("%v", claims["status"])
	currentUserFullName := fmt.Sprintf("%v", claims["currentUserFullName"])

	roleId, err := strconv.Atoi(roleIdStr)
	if err != nil {
		return nil, err
	}

	orgId, err := strconv.Atoi(orgIdStr)
	if err != nil {
		return nil, err
	}

	var org orgsvc.Org
	result := s.db.Table(orgsvc.OrgTableName).Where("id = ?", orgId).First(&org)
	if result.Error != nil {
		return nil, result.Error
	}

	var user User
	result = s.db.Table(UserTableName).Where("email = ?", email).First(&user)
	if result.Error != nil {
		log.Error("user not found")
	}

	var userOrgRoleCount int64
	result = s.db.Table(orgsvc.UserOrgRoleTableName).
		Joins("LEFT JOIN users as u ON u.id=user_org_roles.user_id").
		Where("u.email = ? AND user_org_roles.org_id = ? AND user_org_roles.status = ?", email, orgId, UserStatusActive).
		Count(&userOrgRoleCount)
	if result.Error != nil {
		return nil, result.Error
	}

	if userOrgRoleCount > 0 {
		return nil, fmt.Errorf("user already has an active role in this organization")
	}

	token := ""
	if roleId == 5 || roleId == 6 || roleId == 7 {
		token = req.Token
	}


	type ownerAdminEmail struct {
		Email string
	}

	rows, err := s.db.Table(UserTableName).
	Select("email").
	Joins("LEFT JOIN user_org_roles AS uor ON uor.user_id = users.id").
		Joins("LEFT JOIN roles on roles.id = uor.role_id").
		Joins("LEFT JOIN orgs on orgs.id = uor.org_id").
		Where("orgs.id = ?", org.ID).
		Where("roles.id = ?", 1).
		Or("roles.id = ?", 2).
		Rows()
	if err != nil {
		return nil, err
	}

	var ownerAdminEmails []ownerAdminEmail
	for rows.Next() {
		oau := ownerAdminEmail{}
		err = rows.Scan(&oau.Email)
		if err != nil {
			return nil, err
		}
		ownerAdminEmails = append(ownerAdminEmails, oau)
	}

	if user.Username != nil {
		// This means user already completed registration so we just redirect

		var usrOrgRole orgsvc.UserOrgRole
		result = s.db.Where("user_id = ? AND org_id = ?", user.ID, org.ID).First(&usrOrgRole)
		if result.Error != nil {
			return nil, result.Error
		}

		result = s.db.Model(&usrOrgRole).Where("user_id = ? AND org_id = ? AND role_id = ?", usrOrgRole.UserID, usrOrgRole.OrgID, usrOrgRole.RoleID).Update("status", status)
		if result.Error != nil {
			return nil, result.Error
		}

		roleStr := helper.RoleTypeStr(usrOrgRole.RoleID)
		inviteeFullName := user.FirstName + " " + user.LastName

		requestedUserType := helper.RoleTypeStr(usrOrgRole.RoleID)
		userProfileLink := s.uiAppUrl + "/o/" + org.Slug + "/users/" + strings.ToLower(*user.Username)

		if status == UserStatusPending {
			for _, ownerAdmin := range ownerAdminEmails {
				m := gomail.NewMessage()	
				m.SetHeader("From", "influxoks@gmail.com")
				m.SetHeader("To", ownerAdmin.Email)
				m.SetHeader("Subject", "Vezhguesi: Request to approve new invited user.")
				m.SetBody("text/html", fmt.Sprintf(`Hello from Vezhguesi!<br/><br/>

					%s has invited %s to join the Organization %s as %s.<br/>
					Please click the link below to review and approve/deny the invitation: <br/><br/>

					<a href='%s'>Approve Invitation</a><br/><br/>

				Thank you, <br/>
				Vezhguesi Team
			`, currentUserFullName, user.Email, requestedUserType, userProfileLink))
				err = s.dialer.DialAndSend(m)
				if err != nil {
					return nil, err
				}
			}
		} else {
			for _, ownerAdmin := range ownerAdminEmails {
				m := gomail.NewMessage()	
				m.SetHeader("From", "influxoks@gmail.com")
				m.SetHeader("To", ownerAdmin.Email)
				m.SetHeader("Subject", "Vezhguesi - New " + roleStr + " has joined " + org.Name)
				m.SetBody("text/html", fmt.Sprintf(`Hello from Vezhguesi!<br/><br/>

					New %s has joined your %s organization!<br/><br/>

					Full Name: %s<br/>
					User Type: %s <br/><br/>

					<a href='%s'>View Profile</a><br/><br/>

					Thank you, <br/>
					Vezhguesi Team
			`, currentUserFullName, user.Email, requestedUserType, inviteeFullName, userProfileLink))
				err = s.dialer.DialAndSend(m)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// If user is not found, we create a new user
	if user.ID == 0 {
		// Generate hash pw
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %v", err)
		}

		user.Password = string(hashedPassword)
		
		// For debugging
		fmt.Printf("New password hash: %s\n", user.Password)
		
		// Verify the hash immediately to ensure it works
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			return nil, fmt.Errorf("password hash verification failed: %v", err)
		}

		user.Email = email
		user.FirstName = req.FirstName
		user.LastName = req.LastName
		user.Username = &req.UserName
		user.Active = true
		user.VerifiedEmail = true

		result = s.db.Omit("UpdatedAt").Create(&user)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &AcceptInvitationResponse{
		InviteAccepted: true,
		OrgSlug: org.Slug,
		Token: token,
		Status: status,
		RoleID: roleId,
	}, nil
}





// Private helper funcs
func handleTotalUsersLimit(db *gorm.DB, orgId int) error {
	var totalUserCount int64
	result := db.Table("user_org_roles").
		Where("org_id = ?", orgId).
		Count(&totalUserCount)
	if result.Error != nil {
		return result.Error
	}

	var totalUserLimitStr string
	result = db.Table("features").Select("value").
		Joins("LEFT JOIN subscriptions AS s ON s.id=features.subscription_id").
		Joins("LEFT JOIN orgs AS o ON o.subscription_id=s.id").
		Where("o.id = ? AND features.key = ?", orgId, "MembersLimit").
		Scan(&totalUserLimitStr)
	if result.Error != nil {
		return result.Error
	}
	totalUserLimit, err := strconv.Atoi(totalUserLimitStr)
	if err != nil {
		return err
	}

	if int(totalUserCount) >= totalUserLimit {
		return fmt.Errorf("user creation has reached limit, consider upgrading your plan.")
	}
	return nil
}

func handleAdminRoleLimit(db *gorm.DB, orgId int) error {
	var adminUserCount int64
	result := db.Table("user_org_roles").
		Where("org_id = ? AND role_id = ?", orgId, 2).
		Count(&adminUserCount)
	if result.Error != nil {
		return result.Error
	}

	var adminRoleLimitStr string
	result = db.Table("features").Select("value").
		Joins("LEFT JOIN subscriptions AS s ON s.id=features.subscription_id").
		Joins("LEFT JOIN orgs AS o ON o.subscription_id=s.id").
		Where("o.id = ? AND features.key = ?", orgId, "AdminRoleLimit").
		Scan(&adminRoleLimitStr)
	if result.Error != nil {
		return result.Error
	}
	adminRoleLimit, err := strconv.Atoi(adminRoleLimitStr)
	if err != nil {
		return err
	}
	if int(adminUserCount) >= adminRoleLimit {
		return fmt.Errorf("User admin roles has reached limit, consider upgrading your plan.")
	}
	return nil
}

func handleAdvisorRoleLimit(db *gorm.DB, orgId int) error {
	var advisorUserCount int64
	result := db.Table("user_org_roles").
		Where("org_id = ? AND role_id = ?", orgId, 4).
		Count(&advisorUserCount)
	if result.Error != nil {
		return result.Error
	}

	var advisorRoleLimitStr string
	result = db.Table("features").Select("value").
		Joins("LEFT JOIN subscriptions AS s ON s.id=features.subscription_id").
		Joins("LEFT JOIN orgs AS o ON o.subscription_id=s.id").
		Where("o.id = ? AND features.key = ?", orgId, "MentorRoleLimit").
		Scan(&advisorRoleLimitStr)
	if result.Error != nil {
		return result.Error
	}
	advisorRoleLimit, err := strconv.Atoi(advisorRoleLimitStr)
	if err != nil {
		return err
	}
	if int(advisorUserCount) >= advisorRoleLimit {
		return fmt.Errorf("User mentor roles has reached limit, consider upgrading your plan.")
	}
	return nil
}
