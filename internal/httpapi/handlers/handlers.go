/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	v1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/redhat-data-and-ai/usernaut/pkg/types"
)

type Handlers struct {
	config *config.AppConfig
	cache  cache.Cache
}

func NewHandlers(cfg *config.AppConfig, c cache.Cache) *Handlers {
	return &Handlers{
		config: cfg,
		cache:  c,
	}
}

func (h *Handlers) GetBackends(c *gin.Context) {
	response := make([]v1alpha1.Backend, 0, len(h.config.Backends))

	for _, backend := range h.config.Backends {
		if backend.Enabled {
			response = append(response, v1alpha1.Backend{
				Name: backend.Name,
				Type: backend.Type,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) GetUser(c *gin.Context) {
	email := strings.ToLower(c.Param("userEmail"))

	value, err := h.cache.Get(c.Request.Context(), email)
	if err != nil {

		if idx := strings.Index(email, "@"); idx > 0 {
			username := email[:idx]
			value, err = h.cache.Get(c.Request.Context(), username)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
	}

	rawValue, ok := value.(string)

	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data format in cache"})
		return
	}

	var cached types.CachedUser

	if err := json.Unmarshal([]byte(rawValue), &cached); err != nil || cached.Groups == nil {
		// if it fails try to unmarshall as simple map format (reconciliation)
		var simpleMap map[string]string
		if mapErr := json.Unmarshal([]byte(rawValue), &simpleMap); mapErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode user data"})
			return
		}

		cached = types.CachedUser{Groups: map[string][]types.BackendUser{}}
	}

	type GroupResponse struct {
		Group    string              `json:"group"`
		Backends []types.BackendUser `json:"backends"`
	}

	response := []GroupResponse{}

	for groupName, backendUsers := range cached.Groups {
		backends := make([]types.BackendUser, 0, len(backendUsers))
		for _, bu := range backendUsers {
			backends = append(backends, types.BackendUser{
				Name: bu.Name,
				Type: bu.Type,
				ID:   bu.ID,
			})
		}
		response = append(response, GroupResponse{
			Group:    groupName,
			Backends: backends,
		})
	}

	c.JSON(http.StatusOK, response)
}
