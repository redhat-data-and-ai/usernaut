package assets

import (
	"encoding/json"
	"fmt"

	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

type TokenClient AtlanClient

// Get retrieves an ApiTokenResponse with a list of API tokens based on the provided parameters.
func (tc *TokenClient) Get(limit *int, postFilter, sort *string, count bool, offset int) (*ApiTokenResponse, error) {
	queryParams := map[string]string{
		"count":  fmt.Sprintf("%v", count),
		"offset": fmt.Sprintf("%d", offset),
	}
	if limit != nil {
		queryParams["limit"] = fmt.Sprintf("%d", *limit)
	}
	if postFilter != nil {
		queryParams["filter"] = *postFilter
	}
	if sort != nil {
		queryParams["sort"] = *sort
	}

	rawJSON, err := DefaultAtlanClient.CallAPI(&GET_API_TOKENS, queryParams, nil)
	if err != nil {
		return nil, err
	}

	var response ApiTokenResponse
	err = json.Unmarshal(rawJSON, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetByName retrieves the API token with a display name.
// returns the API token with the provided display name as structs.ApiToken.
func (tc *TokenClient) GetByName(displayName string) (*structs.ApiToken, error) {
	filter := fmt.Sprintf(`{"displayName":"%s"}`, displayName)
	response, err := tc.Get(nil, &filter, nil, true, 0)
	if err != nil || response == nil || *response.TotalRecord == 0 {
		return nil, err
	}
	return response.Records[0], nil
}

// GetByID retrieves the API token with a client ID.
// returns the API token with the provided client ID as structs.ApiToken.
func (tc *TokenClient) GetByID(clientID string) (*structs.ApiToken, error) {
	if len(clientID) > len(structs.ServiceAccount) && clientID[:len(structs.ServiceAccount)] == structs.ServiceAccount {
		clientID = clientID[len(structs.ServiceAccount):]
	}
	filter := fmt.Sprintf(`{"clientId":"%s"}`, clientID)
	response, err := tc.Get(nil, &filter, nil, true, 0)
	if err != nil || response == nil || len(response.Records) == 0 {
		return nil, err
	}
	return response.Records[0], nil
}

// GetByGUID retrieves the API token with a GUID.
// returns the API token with the provided GUID as structs.ApiToken.
func (tc *TokenClient) GetByGUID(guid string) (*structs.ApiToken, error) {
	filter := fmt.Sprintf(`{"id":"%s"}`, guid)
	sort := "createdAt"
	response, err := tc.Get(nil, &filter, &sort, true, 0)
	if err != nil || response == nil || len(response.Records) == 0 {
		return nil, err
	}
	return response.Records[0], nil
}

// Create creates a new API token.
// displayName: Human-readable name of the token.
// description: Description of the token.
// personas: List of persona qualified names.
// validitySeconds: Validity of the token in seconds
// returns the created API token as structs.ApiToken.
func (tc *TokenClient) Create(displayName, description *string, personas []string, validitySeconds *int) (*structs.ApiToken, error) {
	request := structs.ApiTokenRequest{
		DisplayName:           displayName,
		Description:           " ",
		Personas:              []string{},
		PersonaQualifiedNames: personas,
		ValiditySeconds:       validitySeconds,
	}

	if description != nil {
		request.Description = *description
	}

	if validitySeconds != nil {
		request.ValiditySeconds = validitySeconds
	}

	rawJSON, err := DefaultAtlanClient.CallAPI(&UPSERT_API_TOKEN, nil, request)
	if err != nil {
		return nil, err
	}

	var token structs.ApiToken
	err = json.Unmarshal(rawJSON, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Update updates an existing API token with the provided settings.
// guid: GUID of the token to update.
// displayName: Updated Human-readable name of the token.
// description: Updated Description of the token.
// personas: Updated List of persona qualified names.
// returns the updated API token as structs.ApiToken.
func (tc *TokenClient) Update(guid, displayName, description *string, personas []string) (*structs.ApiToken, error) {
	request := structs.ApiTokenRequest{
		DisplayName: displayName,
	}

	if description != nil {
		request.Description = *description
	}

	if personas != nil {
		request.PersonaQualifiedNames = personas
	}

	api := &UPSERT_API_TOKEN
	api.Path = fmt.Sprintf("apikeys/%s", *guid)
	rawJSON, err := DefaultAtlanClient.CallAPI(api, nil, request)
	if err != nil {
		return nil, err
	}

	var token structs.ApiToken
	err = json.Unmarshal(rawJSON, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Purge deletes the API token with the provided GUID.
// returns error if the API token could not be deleted.
func (tc *TokenClient) Purge(guid string) error {
	api := &DELETE_API_TOKEN
	api.Path = fmt.Sprintf("apikeys/%s", guid)
	_, err := DefaultAtlanClient.CallAPI(api, nil, nil)
	return err
}

// ApiTokenResponse represents the response for API token requests.
type ApiTokenResponse struct {
	TotalRecord  *int                `json:"totalRecord,omitempty"`  // Total number of API tokens.
	FilterRecord *int                `json:"filterRecord,omitempty"` // Number of records matching filters.
	Records      []*structs.ApiToken `json:"records,omitempty"`      // Matching API tokens.
}

// UnmarshalJSON custom unmarshal method for ApiTokenResponse.
func (r *ApiTokenResponse) UnmarshalJSON(data []byte) error {
	type Alias ApiTokenResponse
	aux := &struct {
		Records json.RawMessage `json:"records"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	// Unmarshal top-level fields
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Handle records field
	var tokens []json.RawMessage
	if err := json.Unmarshal(aux.Records, &tokens); err != nil {
		return fmt.Errorf("error unmarshalling records: %w", err)
	}

	// Process each record individually
	for _, tokenData := range tokens {
		var token structs.ApiToken
		if err := json.Unmarshal(tokenData, &token); err != nil {
			return fmt.Errorf("error unmarshalling ApiToken: %w", err)
		}
		r.Records = append(r.Records, &token)
	}

	return nil
}
