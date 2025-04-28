package structs

type Team struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Role        string `json:"role,omitempty"`
}

func (t *Team) GetID() string {
	return t.ID
}

func (t *Team) GetName() string {
	return t.Name
}

func (t *Team) GetDescription() string {
	return t.Description
}

func (t *Team) GetRole() string {
	return t.Role
}
