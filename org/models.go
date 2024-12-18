package org

type OrgResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Size string `json:"size"`
	Slug string `json:"slug"`
}

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

type AddOrgRequest struct {
	Name   string `json:"name"`
	Size   string `json:"size"`
	UserID int    `json:"-"`
}

type IDRequest struct {
	ID     int `json:"id"`
	UserID int `json:"-"`
}

type GetOrgsResponse struct {
	Orgs []OrgResponse `json:"orgs"`
}

type OrgWithRole struct {
	OrgID  int    `json:"orgId"`
	RoleID int    `json:"roleId"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	UserID int    `json:"userId"`
}

type OrgRequest struct {
	UserID int `json:"-"`
	OrgID  int `json:"-"`
}

type UserOrgRoleResponse struct {
	UserID int    `json:"userId"`
	OrgID  int    `json:"orgId"`
	RoleID int    `json:"roleId"`
	Status string `json:"status"`
}

type UserResponse struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Status       string `json:"status"`
	AvatarImgKey string `json:"avatarImgKey"`
	Active       bool   `json:"active"`
	Phone        string `json:"phone"`
}

type OrgMembers struct {
	UserOrgRole UserOrgRoleResponse `json:"userOrgRole"`
	User        UserResponse        `json:"user"`
}

type OrgMembersResponse struct {
	OrgMembers []OrgMembers `json:"orgMembers"`
}
