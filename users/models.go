package users

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
