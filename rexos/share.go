package rexos

// User is a container for user details used for project sharing
type User struct {
	UserName  string `json:"userName"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserID    string `json:"userID"`
}

// UserShare describes the type of sharing of a project with a user
type UserShare struct {
	User  User `json:"user"`
	Write bool `json:"write"`
	Read  bool `json:"read"`
}

// Share contains all sharing information for a project
type Share struct {
	PublicShare bool        `json:"publicShare"`
	UserShares  []UserShare `json:"userShares"`
}
