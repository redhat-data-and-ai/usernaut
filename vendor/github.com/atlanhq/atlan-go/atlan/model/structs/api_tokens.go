package structs

import (
	"encoding/json"
	"fmt"
	"math"
)

const (
	ServiceAccount = "SERVICE_ACCOUNT_"
	// The value was previously set to 13 years (409968000 secs).
	// It has been reverted to 5 years due to an integer overflow issue in Keycloak.
	// https://github.com/keycloak/keycloak/issues/19671
	MaxValidity = 157680000 // 5 years in seconds
)

// ApiTokenPersona represents a linked persona in the API token model.
type ApiTokenPersona struct {
	GUID                 *string `json:"id,omitempty"`                   // Unique identifier (GUID) of the linked persona.
	Persona              *string `json:"persona,omitempty"`              // Unique name of the linked persona.
	PersonaQualifiedName *string `json:"personaQualifiedName,omitempty"` // Unique qualified name of the persona.
}

// ApiTokenAttributes contains detailed characteristics of an API token.
type ApiTokenAttributes struct {
	AccessTokenLifespan  *string            `json:"access.token.lifespan,omitempty"` // Time in seconds after which the token will expire.
	AccessToken          *string            `json:"accessToken,omitempty"`           // The actual API token.
	ClientID             *string            `json:"clientId,omitempty"`              // Unique client identifier of the API token.
	CreatedAt            *string            `json:"createdAt,omitempty"`             // Epoch time when the token was created.
	CreatedBy            *string            `json:"createdBy,omitempty"`             // User who created the token.
	Description          *string            `json:"description,omitempty"`           // Explanation of the token.
	DisplayName          *string            `json:"displayName,omitempty"`           // Human-readable name of the token.
	Personas             []string           `json:"personas,omitempty"`              // personas associated with the token.
	PersonaQualifiedName []*ApiTokenPersona `json:"personaQualifiedName,omitempty"`  // Associated personas with the token.
	Purposes             interface{}        `json:"purposes,omitempty"`              // Placeholder for purposes associated with the token.
	WorkspacePermissions []string           `json:"workspacePermissions,omitempty"`  // Permissions associated with the token.
}

/*
  We sometimes gets `Personas`, `Workspace Permissions` and `Persona Qualified Name` as string(in case of retrieval) or
  array of strings (in case of update). This is a custom UnmarshalJSON function to handle this by converting string to
  array.
*/

// UnmarshalJSON for ApiTokenAttributes to handle nested fields.
func (a *ApiTokenAttributes) UnmarshalJSON(data []byte) error {
	type Alias ApiTokenAttributes
	aux := &struct {
		Personas             json.RawMessage `json:"personas"`
		WorkspacePermissions json.RawMessage `json:"workspacePermissions"`
		PersonaQualifiedName json.RawMessage `json:"personaQualifiedName"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	// Unmarshal into aux
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Handle personas: check if it's a string or an array of strings
	if len(aux.Personas) > 0 {
		var personasAsString string
		if err := json.Unmarshal(aux.Personas, &personasAsString); err == nil {
			a.Personas = []string{personasAsString} // Convert string to []string
		} else {
			var personasAsArray []string
			if err := json.Unmarshal(aux.Personas, &personasAsArray); err == nil {
				a.Personas = personasAsArray
			} else {
				return fmt.Errorf("error unmarshaling personas: %w", err)
			}
		}
	}

	// Handle workspacePermissions: check if it's a string or an array of strings
	if len(aux.WorkspacePermissions) > 0 {
		var permissionsAsString string
		if err := json.Unmarshal(aux.WorkspacePermissions, &permissionsAsString); err == nil {
			a.WorkspacePermissions = []string{permissionsAsString} // Convert string to []string
		} else {
			var permissionsAsArray []string
			if err := json.Unmarshal(aux.WorkspacePermissions, &permissionsAsArray); err == nil {
				a.WorkspacePermissions = permissionsAsArray
			} else {
				return fmt.Errorf("error unmarshaling workspacePermissions: %w", err)
			}
		}
	}

	// Handle personaQualifiedName: check if it's a string or an array of objects
	if len(aux.PersonaQualifiedName) > 0 {
		var personaAsString string
		if err := json.Unmarshal(aux.PersonaQualifiedName, &personaAsString); err == nil {
			// Convert string to []*ApiTokenPersona
			a.PersonaQualifiedName = []*ApiTokenPersona{
				{PersonaQualifiedName: &personaAsString},
			}
		} else {
			var personaAsArray []*ApiTokenPersona
			if err := json.Unmarshal(aux.PersonaQualifiedName, &personaAsArray); err == nil {
				a.PersonaQualifiedName = personaAsArray
			} else {
				return fmt.Errorf("error unmarshaling personaQualifiedName: %w", err)
			}
		}
	}

	return nil
}

// ApiToken represents an API token object.
type ApiToken struct {
	GUID        *string             `json:"id,omitempty"`          // Unique identifier of the API token.
	ClientID    *string             `json:"clientId,omitempty"`    // Unique client identifier of the token.
	DisplayName *string             `json:"displayName,omitempty"` // Human-readable name of the token.
	Attributes  *ApiTokenAttributes `json:"attributes,omitempty"`  // Detailed characteristics of the token.
}

// Username generates a service account username based on the token's client ID.
func (a *ApiToken) Username() string {
	if a.ClientID != nil {
		return ServiceAccount + *a.ClientID
	}
	if a.Attributes != nil && a.Attributes.ClientID != nil {
		return ServiceAccount + *a.Attributes.ClientID
	}
	return ""
}

// ApiTokenRequest represents the request body for creating or updating an API token.
type ApiTokenRequest struct {
	DisplayName           *string  `json:"displayName,omitempty"`           // Human-readable name of the token.
	Description           string   `json:"description,omitempty"`           // Explanation of the token.
	Personas              []string `json:"personas,omitempty"`              // GUIDs of associated personas.
	PersonaQualifiedNames []string `json:"personaQualifiedNames,omitempty"` // Qualified names of associated personas.
	ValiditySeconds       *int     `json:"validitySeconds,omitempty"`       // Validity of the token in seconds.
}

// SetMaxValidity ensures validity seconds are within allowed range.
func (r *ApiTokenRequest) SetMaxValidity() {
	if r.ValiditySeconds != nil {
		if *r.ValiditySeconds < 0 {
			*r.ValiditySeconds = MaxValidity // Treat negative numbers as "infinite" (never expire)
		} else if *r.ValiditySeconds > MaxValidity {
			// Otherwise use "infinite" as the ceiling for values
			*r.ValiditySeconds = int(math.Min(float64(*r.ValiditySeconds), MaxValidity))
		}
	}
	if r.Personas == nil {
		r.Personas = []string{}
	}
}

// UnmarshalJSON for ApiToken to handle attributes.
func (t *ApiToken) UnmarshalJSON(data []byte) error {
	type Alias ApiToken
	aux := &struct {
		Attributes json.RawMessage `json:"attributes"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	// Unmarshal the base fields of ApiToken
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Unmarshal the nested Attributes field
	if len(aux.Attributes) > 0 {
		var attributes ApiTokenAttributes
		if err := json.Unmarshal(aux.Attributes, &attributes); err != nil {
			return fmt.Errorf("error unmarshalling attributes: %w", err)
		}
		t.Attributes = &attributes
	}

	return nil
}
