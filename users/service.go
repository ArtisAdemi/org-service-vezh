package users

import (
	"fmt"

	"gorm.io/gorm"
)

type userApi struct {
	db *gorm.DB
}

type UserAPI interface {
	GetUser(req *GetUserRequest) (*GetUserResponse, error)
	GetUsers(req *IDRequest) (*GetUsersResponse, error)
}

func NewUserService(db *gorm.DB) UserAPI {
	return &userApi{db: db}
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
