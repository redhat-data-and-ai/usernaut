package assets

import (
	"fmt"
	"strings"
	"sync"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model"
)

/*
   Lazily-loaded cache for translating between Atlan-internal ID strings and human-readable names
   for custom metadata (including attributes).
*/

type CustomMetadataCache struct {
	AtlanClient     *AtlanClient
	CacheByID       map[string]model.CustomMetadataDef
	AttrCacheByID   map[string]model.AttributeDef
	MapIDToName     map[string]string
	MapNameToID     map[string]string
	MapAttrIDToName map[string]map[string]string
	MapAttrNameToID map[string]map[string]string
	archivedAttrIds map[string]string
	mutex           sync.RWMutex
}

// NewCustomMetadataCache creates a new CustomMetadataCache instance.
func NewCustomMetadataCache(atlanClient *AtlanClient) *CustomMetadataCache {
	return &CustomMetadataCache{
		AtlanClient:     DefaultAtlanClient,
		CacheByID:       make(map[string]model.CustomMetadataDef),
		AttrCacheByID:   make(map[string]model.AttributeDef),
		MapIDToName:     make(map[string]string),
		MapNameToID:     make(map[string]string),
		MapAttrIDToName: make(map[string]map[string]string),
		MapAttrNameToID: make(map[string]map[string]string),
		archivedAttrIds: make(map[string]string),
	}
}

var customMetadataCaches = make(map[string]*CustomMetadataCache)

func RefreshCustomMetadataCache() {
	GetCustomMetadataCache().RefreshCache()
}

func GetAttributeDef(attrID string) model.AttributeDef {
	attrdef, _ := GetCustomMetadataCache().GetAttributeDef(attrID)
	return attrdef
}

func GetCustomMetadataIDforName(name string) (string, error) {
	return GetCustomMetadataCache().GetIDForName(name)
}

// GetCustomMetadataCache returns the CustomMetadataCache for the default AtlanClient.
func GetCustomMetadataCache() *CustomMetadataCache {
	client := DefaultAtlanClient
	cacheKey := generateCacheKey(client.host, client.ApiKey)

	mu.Lock()
	defer mu.Unlock()

	if _, ok := customMetadataCaches[cacheKey]; !ok {
		customMetadataCaches[cacheKey] = &CustomMetadataCache{
			AtlanClient:     client,
			CacheByID:       make(map[string]model.CustomMetadataDef),
			AttrCacheByID:   make(map[string]model.AttributeDef),
			MapIDToName:     make(map[string]string),
			MapNameToID:     make(map[string]string),
			MapAttrIDToName: make(map[string]map[string]string),
			MapAttrNameToID: make(map[string]map[string]string),
			archivedAttrIds: make(map[string]string),
		}
	}
	return customMetadataCaches[cacheKey]
}

/*
RefreshCache refreshes the cache of custom metadata structures by requesting the full set of custom metadata structures from Atlan.

:raises LogicError: if duplicate custom attributes are detected
*/
func (c *CustomMetadataCache) RefreshCache() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	response, err := Get(atlan.AtlanTypeCategoryBusinessMetadata)
	if err != nil {
		return err
	}

	if response == nil || len(response.CustomMetadataDefs) == 0 {
		return ThrowAtlanError(err, EXPIRED_API_TOKEN, nil)
	}

	// Clear existing cache data
	c.CacheByID = make(map[string]model.CustomMetadataDef)
	c.MapIDToName = make(map[string]string)
	c.MapNameToID = make(map[string]string)
	c.MapAttrIDToName = make(map[string]map[string]string)
	c.MapAttrNameToID = make(map[string]map[string]string)
	c.archivedAttrIds = make(map[string]string)

	// Populate the cache with the fetched custom metadata structures and their attributes
	for _, cmDef := range response.CustomMetadataDefs {
		typeID := cmDef.Name
		typeName := cmDef.DisplayName

		c.CacheByID[typeID] = cmDef
		c.MapIDToName[typeID] = *typeName
		c.MapNameToID[*typeName] = typeID
		c.MapAttrIDToName[typeID] = make(map[string]string)
		c.MapAttrNameToID[typeID] = make(map[string]string)

		for _, attr := range cmDef.AttributeDefs {
			attrID := *attr.Name
			attrName := *attr.DisplayName
			_, existsInIDToName := c.MapAttrIDToName[typeID][attrID]
			_, existsInNameToID := c.MapAttrNameToID[typeID][attrName]

			// Check for duplicate attributes.
			if existsInIDToName || existsInNameToID {
				return LogicError{AtlanError{ErrorCode: errorCodes[DUPLICATE_CUSTOM_ATTRIBUTES]}}
			}

			if attr.Options == nil || attr.Options.IsArchived {
				c.archivedAttrIds[attrID] = attrName
				continue // Skip adding archived attributes to the active caches.
			}

			c.MapAttrIDToName[typeID][attrID] = attrName
			c.MapAttrNameToID[typeID][attrName] = attrID
			c.AttrCacheByID[attrID] = attr
		}
	}
	return nil
}

/*
GetIDForName Translate the provided human-readable custom metadata set name to its Atlan-internal ID string.

:param name: human-readable name of the custom metadata set

:returns: Atlan-internal ID string of the custom metadata set

:raises InvalidRequestError: if no name was provided

:raises NotFoundError: if the custom metadata cannot be found
*/
func (c *CustomMetadataCache) GetIDForName(name string) (string, error) {
	if name == "" || strings.TrimSpace(name) == "" {
		return "", ThrowAtlanError(nil, MISSING_CM_NAME, nil)
	}

	c.mutex.RLock()
	cmID, exists := c.MapNameToID[name]
	c.mutex.RUnlock()

	if !exists {
		if err := c.RefreshCache(); err != nil {
			return "", ThrowAtlanError(err, CONNECTION_ERROR, nil)
		}

		c.mutex.RLock()
		cmID, exists = c.MapNameToID[name]
		c.mutex.RUnlock()

		if !exists {
			return "", ThrowAtlanError(nil, CM_NOT_FOUND_BY_NAME, nil, name)
		}
	}

	return cmID, nil
}

/*
GetNameForID Translate the provided Atlan-internal custom metadata ID string to the human-readable custom metadata set name.

:param idstr: Atlan-internal ID string of the custom metadata set

:returns: human-readable name of the custom metadata set

:raises InvalidRequestError: if no ID was provided

:raises NotFoundError: if the custom metadata cannot be found
*/
func (c *CustomMetadataCache) GetNameForID(idstr string) (string, error) {
	if idstr = strings.TrimSpace(idstr); idstr == "" {
		return "", ThrowAtlanError(nil, MISSING_CM_ID, nil)
	}

	c.mutex.RLock()
	cmName, exists := c.MapIDToName[idstr]
	c.mutex.RUnlock()

	if !exists {
		if err := c.RefreshCache(); err != nil {
			return "", err
		}

		c.mutex.RLock()
		cmName, exists = c.MapIDToName[idstr]
		c.mutex.RUnlock()

		if !exists {
			return "", ThrowAtlanError(nil, CM_NOT_FOUND_BY_ID, nil, idstr)
		}
	}

	return cmName, nil
}

/*
GetAllCustomAttributes Retrieve all the custom metadata attributes. The dict will be keyed by custom metadata set
name, and the value will be a listing of all the attributes within that set (with all the details
of each of those attributes).

:param includeDeleted: if True, include the archived (deleted) custom attributes; otherwise only include active custom attributes

:param forceRefresh: if True, will refresh the custom metadata cache; if False, will only refresh the cache if it is empty

:returns: a dict from custom metadata set name to all details about its attributes

:raises NotFoundError: if the custom metadata cannot be found
*/
func (c *CustomMetadataCache) GetAllCustomAttributes(includeDeleted, forceRefresh bool) (map[string][]model.AttributeDef, error) {
	if len(c.CacheByID) == 0 || forceRefresh {
		c.RefreshCache()
	}
	m := make(map[string][]model.AttributeDef)
	for typeID, cm := range c.CacheByID {
		typeName, err := c.GetNameForID(typeID)
		if err != nil {
			return nil, ThrowAtlanError(err, CM_NOT_FOUND_BY_NAME, nil, typeID)
		}
		var toInclude []model.AttributeDef
		if includeDeleted {
			toInclude = cm.AttributeDefs
		} else {
			for _, attr := range cm.AttributeDefs {
				if attr.Options == nil || !attr.Options.IsArchived {
					toInclude = append(toInclude, attr)
				}
			}
		}
		m[typeName] = toInclude
	}
	return m, nil
}

/*
GetAttrIDForName Translate the provided human-readable custom metadata set and attribute names to the Atlan-internal ID string
for the attribute.

:param setName: human-readable name of the custom metadata set

:param attrName: human-readable name of the attribute

:returns: Atlan-internal ID string for the attribute

:raises NotFoundError: if the custom metadata attribute cannot be found
*/
func (c *CustomMetadataCache) GetAttrIDForName(setName, attrName string) (string, error) {
	setID, err := c.GetIDForName(setName)
	if err != nil {
		return "", err
	}
	if subMap, ok := c.MapAttrNameToID[setID]; ok {
		if attrID, ok := subMap[attrName]; ok {
			// If found, return straight away
			return attrID, nil
		}
	}
	// Refresh cache and try again
	c.RefreshCache()
	if subMap, ok := c.MapAttrNameToID[setID]; ok {
		if attrID, ok := subMap[attrName]; ok {
			// If found, return straight away
			return attrID, nil
		}
		return "", ThrowAtlanError(nil, CM_ATTR_NOT_FOUND_BY_NAME, nil, attrName, setName)
	}
	return "", ThrowAtlanError(nil, CM_ATTR_NOT_FOUND_BY_ID, nil, setID)
}

/*
GetAttrNameForID Given the Atlan-internal ID string for the set and the Atlan-internal ID for the attribute return the
human-readable custom metadata name for the attribute.

:param setId: Atlan-internal ID string for the custom metadata set
:param attrId: Atlan-internal ID string for the attribute
:returns: human-readable name of the attribute
:raises NotFoundError: if the custom metadata attribute cannot be found
*/
func (c *CustomMetadataCache) GetAttrNameForID(setID, attrID string) (string, error) {
	if subMap, ok := c.MapAttrIDToName[setID]; ok {
		if attrName, ok := subMap[attrID]; ok {
			return attrName, nil
		}
	}
	// Refresh cache and try again
	c.RefreshCache()
	if subMap, ok := c.MapAttrIDToName[setID]; ok {
		if attrName, ok := subMap[attrID]; ok {
			return attrName, nil
		}
	}
	return "", ThrowAtlanError(nil, CM_ATTR_NOT_FOUND_BY_ID, nil, setID)
}

func (c *CustomMetadataCache) GetAttributesForSearchResults(setID string) []string {
	if subMap, ok := c.MapAttrIDToName[setID]; ok {
		var dotNames []string
		for idstr := range subMap {
			dotNames = append(dotNames, fmt.Sprintf("%s.%s", setID, idstr))
		}
		return dotNames
	}
	return nil
}

/*
GetAttributeForSearchResults Retrieve a single custom attribute name to include on search results.

:param set_name: human-readable name of the custom metadata set for which to retrieve the custom metadata attribute name

:param attr_name: human-readable name of the attribute

:returns: the attribute name, strictly useful for inclusion in search results
*/
func (c *CustomMetadataCache) GetAttributeForSearchResults(setID, attrName string) string {
	if subMap, ok := c.MapAttrNameToID[setID]; ok {
		if attrID, ok := subMap[attrName]; ok {
			return attrID
		}
	}
	return ""
}

/*
GetAttributesForSearchResultsByName Retrieve the full set of custom attributes to include on search results.

:param set_name: human-readable name of the custom metadata set for which to retrieve attribute names

:returns: a list of the attribute names, strictly useful for inclusion in search results
*/
func (c *CustomMetadataCache) GetAttributesForSearchResultsByName(setName string) []string {
	setID, _ := c.GetIDForName(setName)
	if setID == "" {
		return nil
	}
	return c.GetAttributesForSearchResults(setID)
}

/*
GetCustomMetadataDef Retrieve the full custom metadata structure definition.

:param name: human-readable name of the custom metadata set

:returns: the full custom metadata structure definition for that set

:raises InvalidRequestError: if no name was provided

:raises NotFoundError: if the custom metadata cannot be found
*/
func (c *CustomMetadataCache) GetCustomMetadataDef(name string) (model.CustomMetadataDef, error) {
	baID, _ := c.GetIDForName(name)
	if baID == "" {
		return model.CustomMetadataDef{}, ThrowAtlanError(nil, CM_NOT_FOUND_BY_NAME, nil, name)
	}
	if typedef, ok := c.CacheByID[baID]; ok {
		return typedef, nil
	}
	return model.CustomMetadataDef{}, ThrowAtlanError(nil, CM_NOT_FOUND_BY_NAME, nil, name)
}

/*
GetAttributeDef Retrieve a specific custom metadata attribute definition by its unique Atlan-internal ID string.

:param attr_id: Atlan-internal ID string for the custom metadata attribute

:returns: attribute definition for the custom metadata attribute

:raises InvalidRequestError: if no attribute ID was provided

:raises NotFoundError: if the custom metadata attribute cannot be found
*/
func (c *CustomMetadataCache) GetAttributeDef(attrID string) (model.AttributeDef, error) {
	if attrID == "" {
		return model.AttributeDef{}, ThrowAtlanError(nil, MISSING_CM_ATTR_ID, nil, attrID)
	}
	if c.AttrCacheByID == nil {
		c.RefreshCache()
	}
	if attrDef, ok := c.AttrCacheByID[attrID]; ok {
		return attrDef, nil
	}
	return model.AttributeDef{}, ThrowAtlanError(nil, CM_ATTR_NOT_FOUND_BY_ID, nil, attrID)
}
