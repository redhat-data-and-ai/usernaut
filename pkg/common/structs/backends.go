package structs

type BackendParams struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (b *BackendParams) GetName() string {
	return b.Name
}

func (b *BackendParams) GetType() string {
	return b.Type
}
