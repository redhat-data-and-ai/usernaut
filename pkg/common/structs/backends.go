package structs

type BackendParams struct {
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	GroupParams TeamParams `json:"group_params,omitempty"`
}

func (b *BackendParams) GetName() string {
	return b.Name
}

func (b *BackendParams) GetType() string {
	return b.Type
}

func (b *BackendParams) GetGroupParams() TeamParams {
	return b.GroupParams
}
