package types

type BackendUser struct {
	Name string `json:"name"`
	Type string `json:"type"`
	ID   string `json:"id"`
}

type User struct {
	Email    string        `json:"email"`
	Groups   []string      `json:"groups"`
	Backends []BackendUser `json:"backends"`
}

type CachedUser struct {
	Groups map[string][]BackendUser `json:"groups"`
}
