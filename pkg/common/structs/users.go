package structs

type User struct {
	ID          string `json:"id,omitempty"`
	UserName    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

func (u *User) GetID() string {
	return u.ID
}

func (u *User) GetUserName() string {
	return u.UserName
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetFirstName() string {
	return u.FirstName
}

func (u *User) GetLastName() string {
	return u.LastName
}

func (u *User) GetDisplayName() string {
	return u.DisplayName
}
