package users

const (
	UserTableName = "users"
)

type User struct {
	ID            int     `gorm:"primaryKey"`
	Email         string  `gorm:"unique"`
	Username      *string `gorm:"unique"`
	OrgID         int
	Password      string
	FirstName     string
	LastName      string
	Status        string
	AvatarImgKey  string
	Active        bool
	Phone         string
	VerifiedEmail bool
	Role          string
}

type IDRequest struct {
	OrgID int `json:"orgId"`
}

type GetUsersResponse struct {
	Users []*User
}

type GetUserRequest struct {
	ID    int    `json:"id"`
	OrgID int    `json:"orgId"`
	Email string `json:"email"`
}

type GetUserResponse struct {
	User *User
}

type ChangeUserRoleRequest struct {
	OrgID     int `json:"-"`
	UserID    int `json:"userId"`
	NewRoleID int `json:"newRoleId"`
}

type StatusResponse struct {
	Status bool `json:"status"`
}

type ChangeUserStatusRequest struct {
	OrgID  int    `json:"-"`
	UserID int    `json:"userId"`
	Status string `json:"status"`
}

type InviteUserRequest struct {
	Email         string `json:"email"`
	RoleID        int    `json:"roleId"`
	OrgID         int    `json:"orgId"`
	CurrentUserID int    `json:"-"`
	CurrentRoleID int    `json:"-"`
}

type AcceptInvitationRequest struct {
	Token           string `json:"token"`
	UserName        string `json:"username"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type AcceptInvitationResponse struct {
	InviteAccepted bool   `json:"inviteAccepted"`
	OrgSlug        string `json:"orgSlug"`
	Token          string `json:"token"`
	Status         string `json:"status"`
	RoleID         int    `json:"roleId"`
}