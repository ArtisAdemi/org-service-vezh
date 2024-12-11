package org

type OrgResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Size string `json:"size"`
	Slug string `json:"slug"`
}

type AddOrgRequest struct {
	Name string `json:"name"`
	Size string `json:"size"`
	UserID int `json:"-"`
}

type IDRequest struct {
	ID int `json:"id"`
	UserID int `json:"-"`
}

type GetOrgsResponse struct {
	Orgs []OrgResponse `json:"orgs"`
}
