package assets

import (
	"errors"
	"sync"

	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

// RoleCache provides a lazily-loaded cache for translating Atlan-internal roles into their IDs and names.
type RoleCache struct {
	roleClient  *RoleClient
	cacheByID   map[string]structs.AtlanRole
	mapIDToName map[string]string
	mapNameToID map[string]string
	mutex       sync.Mutex
}

var (
	roleCaches = make(map[string]*RoleCache)
	cacheMutex sync.Mutex
)

// GetCache retrieves the RoleCache for the default Atlan client.
func GetCache() (*RoleCache, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	client := DefaultAtlanClient
	cacheKey := generateCacheKey(client.host, client.ApiKey)

	if roleCaches[cacheKey] == nil {
		roleCaches[cacheKey] = &RoleCache{
			roleClient:  client.RoleClient,
			cacheByID:   make(map[string]structs.AtlanRole),
			mapIDToName: make(map[string]string),
			mapNameToID: make(map[string]string),
		}
	}
	return roleCaches[cacheKey], nil
}

// GetRoleIDForRoleName  translates the provided role name to its GUID.
func GetRoleIDForRoleName(name string) (string, error) {
	cache, err := GetCache()
	if err != nil {
		return "", err
	}
	return cache.getIDForName(name), nil
}

// GetRoleNameForRoleID translates the provided role GUID to its human-readable name.
func GetRoleNameForRoleID(id string) (string, error) {
	cache, err := GetCache()
	if err != nil {
		return "", err
	}
	return cache.getNameForID(id), nil
}

// ValidateIDStrings validates that the given role GUIDs are valid.
func ValidateIDStrings(ids []string) error {
	cache, err := GetCache()
	if err != nil {
		return err
	}
	return cache.validateIDStrings(ids)
}

func (rc *RoleCache) refreshCache() error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	response, err := rc.roleClient.Get(100, `{"name":{"$ilike":"$%"}}`, "", false, 0)
	if err != nil {
		return err
	}

	rc.cacheByID = make(map[string]structs.AtlanRole)
	rc.mapIDToName = make(map[string]string)
	rc.mapNameToID = make(map[string]string)

	for _, role := range *response.Records {
		rc.cacheByID[*role.ID] = role
		rc.mapIDToName[*role.ID] = *role.Name
		rc.mapNameToID[*role.Name] = *role.ID
	}

	return nil
}

func (rc *RoleCache) getIDForName(name string) string {
	if id, exists := rc.mapNameToID[name]; exists {
		return id
	}
	rc.refreshCache()
	return rc.mapNameToID[name]
}

func (rc *RoleCache) getNameForID(id string) string {
	if name, exists := rc.mapIDToName[id]; exists {
		return name
	}
	rc.refreshCache()
	return rc.mapIDToName[id]
}

func (rc *RoleCache) validateIDStrings(ids []string) error {
	for _, id := range ids {
		if _, exists := rc.mapIDToName[id]; !exists {
			rc.refreshCache()
			if _, exists := rc.mapIDToName[id]; !exists {
				return errors.New("provided role ID not found in Atlan")
			}
		}
	}
	return nil
}
