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

	"github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
)

type Handlers struct {
	config *config.AppConfig
}

func NewHandlers(cfg *config.AppConfig) *Handlers {
	return &Handlers{
		config: cfg,
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
