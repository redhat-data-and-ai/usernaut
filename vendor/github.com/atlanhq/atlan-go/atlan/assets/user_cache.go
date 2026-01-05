package assets

import (
	"errors"
	"sync"
)

// UserCache provides a lazily-loaded cache for translating Atlan-internal users into their IDs and names.
type UserCache struct {
	userClient   *UserClient
	tokenClient  *TokenClient
	mapIDToName  map[string]string
	mapNameToID  map[string]string
	mapEmailToID map[string]string
	mutex        sync.Mutex
}

var (
	userCaches = make(map[string]*UserCache)
	userMutex  sync.Mutex
)

// GetUserCache retrieves the UserCache for the default Atlan client.
func GetUserCache() (*UserCache, error) {
	userMutex.Lock()
	defer userMutex.Unlock()

	client := DefaultAtlanClient
	cacheKey := generateCacheKey(client.host, client.ApiKey)

	if userCaches[cacheKey] == nil {
		userCaches[cacheKey] = &UserCache{
			userClient:   client.UserClient,
			tokenClient:  client.TokenClient,
			mapIDToName:  make(map[string]string),
			mapNameToID:  make(map[string]string),
			mapEmailToID: make(map[string]string),
		}
	}
	return userCaches[cacheKey], nil
}

// GetUserIDForName translates the provided human-readable username to its GUID.
func GetUserIDForName(name string) (string, error) {
	cache, err := GetUserCache()
	if err != nil {
		return "", err
	}
	return cache.getIDForName(name)
}

// GetUserIDForEmail translates the provided email to its GUID.
func GetUserIDForEmail(email string) (string, error) {
	cache, err := GetUserCache()
	if err != nil {
		return "", err
	}
	return cache.getIDForEmail(email)
}

// GetUserNameForID translates the provided user GUID to the human-readable username.
func GetUserNameForID(id string) (string, error) {
	cache, err := GetUserCache()
	if err != nil {
		return "", err
	}
	return cache.getNameForID(id)
}

// ValidateUserNames validates that the given human-readable usernames are valid.
func ValidateUserNames(names []string) error {
	cache, err := GetUserCache()
	if err != nil {
		return err
	}
	return cache.validateNames(names)
}

func (uc *UserCache) refreshCache() error {
	uc.mutex.Lock()
	defer uc.mutex.Unlock()

	users, err := uc.userClient.GetAll(20, 0, "")
	if err != nil {
		return err
	}

	uc.mapIDToName = make(map[string]string)
	uc.mapNameToID = make(map[string]string)
	uc.mapEmailToID = make(map[string]string)

	for _, user := range users {
		userID := user.ID
		userName := user.Username
		userEmail := user.Email

		uc.mapIDToName[userID] = *userName
		uc.mapNameToID[*userName] = userID
		uc.mapEmailToID[userEmail] = userID
	}

	return nil
}

func (uc *UserCache) getIDForName(name string) (string, error) {
	if id, exists := uc.mapNameToID[name]; exists {
		return id, nil
	}
	// If the name is an API token, try fetching it directly
	if isServiceAccount(name) {
		token, err := uc.tokenClient.GetByID(name)
		if err != nil {
			return "", err
		}
		if token != nil && token.GUID != nil {
			uc.mapNameToID[name] = *token.GUID
			return *token.GUID, nil
		}
		return "", errors.New("API token not found by name")
	}
	uc.refreshCache()
	return uc.mapNameToID[name], nil
}

func (uc *UserCache) getIDForEmail(email string) (string, error) {
	if id, exists := uc.mapEmailToID[email]; exists {
		return id, nil
	}
	uc.refreshCache()
	return uc.mapEmailToID[email], nil
}

func (uc *UserCache) getNameForID(id string) (string, error) {
	if name, exists := uc.mapIDToName[id]; exists {
		return name, nil
	}
	// If the ID is an API token, try fetching it directly
	token, err := uc.tokenClient.GetByGUID(id)
	if err != nil {
		return "", err
	}
	if token != nil && token.ClientID != nil {
		return *token.ClientID, nil
	}
	uc.refreshCache()
	return uc.mapIDToName[id], nil
}

func (uc *UserCache) validateNames(names []string) error {
	for _, name := range names {
		if _, err := uc.getIDForName(name); err != nil {
			return err
		}
	}
	return nil
}

// Helper function to check if a name is a service account
func isServiceAccount(name string) bool {
	return len(name) > len("service-account-") && name[:len("service-account-")] == "service-account-"
}
