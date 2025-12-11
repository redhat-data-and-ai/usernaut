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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/redhat-data-and-ai/usernaut/pkg/store"
)

type Handlers struct {
	config *config.AppConfig
	store  *store.Store
}

func NewHandlers(cfg *config.AppConfig, dataStore *store.Store) *Handlers {
	return &Handlers{
		config: cfg,
		store:  dataStore,
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

// UserGroupsResponse represents the response for user groups endpoint
type UserGroupsResponse struct {
	Email  string          `json:"email"`
	Groups []GroupResponse `json:"groups"`
}

// GroupResponse represents a group with its backends
type GroupResponse struct {
	Name     string            `json:"name"`
	Backends []BackendResponse `json:"backends"`
}

// BackendResponse represents backend info in the response
type BackendResponse struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetUserGroups returns the groups a user belongs to along with backend information
func (h *Handlers) GetUserGroups(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email parameter is required"})
		return
	}

	ctx := c.Request.Context()

	// Get groups for the user from the reverse index
	groups, err := h.store.UserGroups.GetGroups(ctx, email)
	if err != nil {
		logrus.WithField("email", email).WithError(err).Error("failed to fetch user groups")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user groups"})
		return
	}

	// Build response with backend info for each group
	groupResponses := make([]GroupResponse, 0, len(groups))
	for _, groupName := range groups {
		backends, err := h.store.Group.GetBackends(ctx, groupName)
		if err != nil {
			// Skip groups that have errors fetching backends
			logrus.WithField("group", groupName).WithError(err).Warn("failed to fetch backends for group, skipping")
			continue
		}

		backendResponses := make([]BackendResponse, 0, len(backends))
		for _, backendInfo := range backends {
			backendResponses = append(backendResponses, BackendResponse{
				Name: backendInfo.Name,
				Type: backendInfo.Type,
			})
		}

		groupResponses = append(groupResponses, GroupResponse{
			Name:     groupName,
			Backends: backendResponses,
		})
	}

	response := UserGroupsResponse{
		Email:  email,
		Groups: groupResponses,
	}

	c.JSON(http.StatusOK, response)
}
