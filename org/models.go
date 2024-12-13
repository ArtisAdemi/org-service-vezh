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
