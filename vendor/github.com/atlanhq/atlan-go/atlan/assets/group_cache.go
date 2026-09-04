package assets

import (
	"errors"
	"sync"
)

// GroupCache provides a lazily-loaded cache for translating Atlan-internal groups into their IDs and names.
type GroupCache struct {
	groupClient  *GroupClient
	cacheByID    map[string]AtlanGroup
	mapIDToName  map[string]string
	mapNameToID  map[string]string
	mapAliasToID map[string]string
	mutex        sync.Mutex
}

var (
	groupCaches = make(map[string]*GroupCache)
	groupMutex  sync.Mutex
)

// GetGroupCache retrieves the GroupCache for the default Atlan client.
func GetGroupCache() (*GroupCache, error) {
	groupMutex.Lock()
	defer groupMutex.Unlock()

	client := DefaultAtlanClient
	cacheKey := generateCacheKey(client.host, client.ApiKey)

	if groupCaches[cacheKey] == nil {
		groupCaches[cacheKey] = &GroupCache{
			groupClient:  client.GroupClient,
			cacheByID:    make(map[string]AtlanGroup),
			mapIDToName:  make(map[string]string),
			mapNameToID:  make(map[string]string),
			mapAliasToID: make(map[string]string),
		}
	}
	return groupCaches[cacheKey], nil
}

// GetGroupIDForGroupName translates the provided group name to its GUID.
func GetGroupIDForGroupName(name string) (string, error) {
	cache, err := GetGroupCache()
	if err != nil {
		return "", err
	}
	return cache.getIDForName(name), nil
}

// GetGroupIDForAlias translates the provided group alias to its GUID.
func GetGroupIDForAlias(alias string) (string, error) {
	cache, err := GetGroupCache()
	if err != nil {
		return "", err
	}
	return cache.getIDForAlias(alias), nil
}

// GetGroupNameForGroupID translates the provided group GUID to its name.
func GetGroupNameForGroupID(id string) (string, error) {
	cache, err := GetGroupCache()
	if err != nil {
		return "", err
	}
	return cache.getNameForID(id), nil
}

// ValidateGroupAliases validates that the given group aliases are valid.
func ValidateGroupAliases(aliases []string) error {
	cache, err := GetGroupCache()
	if err != nil {
		return err
	}
	return cache.validateAliases(aliases)
}

func (gc *GroupCache) refreshCache() error {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	groups, err := gc.groupClient.GetAll(20, 0, "")
	if err != nil {
		return err
	}

	gc.cacheByID = make(map[string]AtlanGroup)
	gc.mapIDToName = make(map[string]string)
	gc.mapNameToID = make(map[string]string)
	gc.mapAliasToID = make(map[string]string)

	for _, group := range groups {
		groupID := *group.ID
		groupName := *group.Name
		groupAlias := *group.Alias

		gc.cacheByID[groupID] = *group
		gc.mapIDToName[groupID] = groupName
		gc.mapNameToID[groupName] = groupID
		gc.mapAliasToID[groupAlias] = groupID
	}

	return nil
}

func (gc *GroupCache) getIDForName(name string) string {
	if id, exists := gc.mapNameToID[name]; exists {
		return id
	}
	gc.refreshCache()
	return gc.mapNameToID[name]
}

func (gc *GroupCache) getIDForAlias(alias string) string {
	if id, exists := gc.mapAliasToID[alias]; exists {
		return id
	}
	gc.refreshCache()
	return gc.mapAliasToID[alias]
}

func (gc *GroupCache) getNameForID(id string) string {
	if name, exists := gc.mapIDToName[id]; exists {
		return name
	}
	gc.refreshCache()
	return gc.mapIDToName[id]
}

func (gc *GroupCache) validateAliases(aliases []string) error {
	for _, alias := range aliases {
		if _, exists := gc.mapAliasToID[alias]; !exists {
			gc.refreshCache()
			if _, exists := gc.mapAliasToID[alias]; !exists {
				return errors.New("provided group alias not found in Atlan")
			}
		}
	}
	return nil
}
