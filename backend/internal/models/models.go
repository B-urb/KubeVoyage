package models

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Email    string `gorm:"uniqueIndex"`
	Password string
	Role     string
}

type Site struct {
	ID  uint   `gorm:"primaryKey"`
	URL string `gorm:"uniqueIndex"`
}

type UserSite struct {
	UserID uint `gorm:"primaryKey"`
	SiteID uint `gorm:"primaryKey"`
	State  State
}

type UserSiteResponse struct {
	User  string `json:"user"`
	Site  string `json:"site"`
	State string `json:"state"`
}

type State string

const (
	Requested  State = "requested"
	Authorized State = "authorized"
	Declined   State = "declined"
)

func (s State) IsValid() bool {
	switch s {
	case Requested, Authorized, Declined:
		return true
	}
	return false
}
