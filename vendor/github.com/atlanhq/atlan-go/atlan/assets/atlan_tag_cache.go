package assets

import (
	"fmt"
	"sync"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model"
)

// AtlanTagCache represents a  cache for translating between Atlan-internal ID strings
// and human-readable names for Atlan tags.
type AtlanTagCache struct {
	atlanClient  *AtlanClient
	cacheByID    map[string]model.AtlanTagDef
	mapIDToName  map[string]string
	mapNameToID  map[string]string
	deletedIDs   map[string]struct{}
	deletedNames map[string]struct{}
	mutex        sync.RWMutex
}

// NewAtlanTagCache creates a new AtlanTagCache instance.
func NewAtlanTagCache(atlanClient *AtlanClient) *AtlanTagCache {
	return &AtlanTagCache{
		atlanClient:  DefaultAtlanClient,
		cacheByID:    make(map[string]model.AtlanTagDef),
		mapIDToName:  make(map[string]string),
		mapNameToID:  make(map[string]string),
		deletedIDs:   make(map[string]struct{}),
		deletedNames: make(map[string]struct{}),
	}
}

var (
	caches = make(map[string]*AtlanTagCache)
	mu     sync.Mutex
)

// GetCache returns the AtlanTagCache for the default AtlanClient.
func GetAtlanTagCache() *AtlanTagCache {
	client := DefaultAtlanClient
	cacheKey := generateCacheKey(client.host, client.ApiKey)

	mu.Lock()
	defer mu.Unlock()

	if _, ok := caches[cacheKey]; !ok {
		caches[cacheKey] = &AtlanTagCache{
			atlanClient:  client,
			cacheByID:    make(map[string]model.AtlanTagDef),
			mapIDToName:  make(map[string]string),
			mapNameToID:  make(map[string]string),
			deletedIDs:   make(map[string]struct{}),
			deletedNames: make(map[string]struct{}),
		}
	}
	return caches[cacheKey]
}

func RefreshCache() {
	GetAtlanTagCache().RefreshCache()
}

func GetAtlanTagIDForName(name string) (string, error) {
	return GetAtlanTagCache().GetIDForName(name)
}

func GetAtlanTagNameForID(idstr string) (string, error) {
	return GetAtlanTagCache().GetNameForID(idstr)
}

// RefreshCache ref	reshes the cache of Atlan tags by requesting the full set of Atlan tags from Atlan.
// RefreshCache updates the AtlanTagCache with the latest data from Atlan.
func (c *AtlanTagCache) RefreshCache() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	response, err := Get(atlan.AtlanTypeCategoryClassification)
	if err != nil {
		fmt.Printf("Error making API call: %v", err)
		return err
	}

	c.cacheByID = make(map[string]model.AtlanTagDef)
	c.mapIDToName = make(map[string]string)
	c.mapNameToID = make(map[string]string)

	for _, atlanTag := range response.AtlanTagDefs {
		c.cacheByID[atlanTag.TypeDefBase.GUID] = atlanTag
		c.mapIDToName[atlanTag.Name] = atlanTag.DisplayName
		c.mapNameToID[atlanTag.DisplayName] = atlanTag.Name
	}

	return nil
}

// GetIDForName translates the provided human-readable Atlan tag name to its Atlan-internal ID string.
func (c *AtlanTagCache) GetIDForName(name string) (string, error) {
	clsID, found := c.mapNameToID[name]

	if !found && name != "" {
		// If not found, refresh the cache and look again (could be stale)
		if err := c.RefreshCache(); err != nil {
			return "", err
		}
		clsID, found = c.mapNameToID[name]
		if !found {
			// If still not found after refresh, mark it as deleted (could be
			// an entry in an audit log that refers to a classification that
			// no longer exists)
			c.deletedNames[name] = struct{}{}
		}
	}

	return clsID, nil
}

// GetNameForID translates the provided Atlan-internal classification ID string to the human-readable Atlan tag name.
func (c *AtlanTagCache) GetNameForID(idstr string) (string, error) {
	clsName, found := c.mapIDToName[idstr]

	if !found && idstr != "" {
		// If not found, refresh the cache and look again (could be stale)
		if err := c.RefreshCache(); err != nil {
			return "", err
		}
		clsName, found = c.mapIDToName[idstr]

		if !found {
			// If still not found after refresh, mark it as deleted (could be
			// an entry in an audit log that refers to a classification that
			// no longer exists)
			c.deletedIDs[idstr] = struct{}{}
		}
	}

	return clsName, nil
}
