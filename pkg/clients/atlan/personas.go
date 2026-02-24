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

package atlan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// AddGroupToPersonas adds a group to the default persona and any additional personas specified
func (ac *AtlanClient) AddGroupToPersonas(ctx context.Context, groupName string, additionalPersonas []string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":   "atlan",
		"groupName": groupName,
	})

	// Build persona list: default + additional
	personas := additionalPersonas
	if ac.defaultPersona != "" {
		personas = append([]string{ac.defaultPersona}, personas...)
	}

	if len(personas) == 0 {
		log.Info("no personas configured, skipping persona assignment")
		return nil
	}

	for _, personaName := range personas {
		if err := ac.addGroupToSinglePersona(ctx, groupName, personaName); err != nil {
			return err
		}
		log.WithField("persona", personaName).Info("added group to persona")
	}

	return nil
}

func (ac *AtlanClient) addGroupToSinglePersona(ctx context.Context, groupName, personaName string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service":     "atlan",
		"groupName":   groupName,
		"personaName": personaName,
	})

	// Search for persona
	persona, err := ac.findPersonaByName(ctx, personaName)
	if err != nil {
		return err
	}
	if persona == nil {
		return fmt.Errorf("persona %s not found", personaName)
	}

	// Check if group already assigned
	if slices.Contains(persona.PersonaGroups, groupName) {
		log.Info("group already assigned to persona")
		return nil
	}

	// Update persona with new group
	return ac.updatePersonaGroups(ctx, persona, append(persona.PersonaGroups, groupName))
}

type personaSearchResponse struct {
	Entities []struct {
		TypeName   string `json:"typeName"`
		Guid       string `json:"guid"`
		Attributes struct {
			Name          string   `json:"name"`
			QualifiedName string   `json:"qualifiedName"`
			PersonaGroups []string `json:"personaGroups"`
		} `json:"attributes"`
	} `json:"entities"`
}

type personaEntity struct {
	Guid          string
	Name          string
	QualifiedName string
	PersonaGroups []string
}

func (ac *AtlanClient) findPersonaByName(ctx context.Context, name string) (*personaEntity, error) {
	url := fmt.Sprintf("%s/api/meta/search/indexsearch", ac.url)

	requestBody := map[string]interface{}{
		"dsl": map[string]interface{}{
			"from": 0,
			"size": 1,
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"filter": []map[string]interface{}{
						{"term": map[string]interface{}{"__state": "ACTIVE"}},
						{"term": map[string]interface{}{"__typeName.keyword": "Persona"}},
						{"term": map[string]interface{}{"name.keyword": name}},
					},
				},
			},
		},
		"attributes": []string{"name", "qualifiedName", "personaGroups"},
	}

	response, err := ac.sendRequest(ctx, url, http.MethodPost, requestBody, "FindPersonaByName")
	if err != nil {
		return nil, err
	}

	var searchResponse personaSearchResponse
	if err := json.Unmarshal(response, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse persona search response: %w", err)
	}

	for _, entity := range searchResponse.Entities {
		if entity.TypeName == "Persona" && entity.Attributes.Name == name {
			return &personaEntity{
				Guid:          entity.Guid,
				Name:          entity.Attributes.Name,
				QualifiedName: entity.Attributes.QualifiedName,
				PersonaGroups: entity.Attributes.PersonaGroups,
			}, nil
		}
	}

	return nil, nil
}

func (ac *AtlanClient) updatePersonaGroups(ctx context.Context, persona *personaEntity, groups []string) error {
	url := fmt.Sprintf("%s/api/meta/entity/bulk", ac.url)

	requestBody := map[string]interface{}{
		"entities": []map[string]interface{}{
			{
				"typeName": "Persona",
				"guid":     persona.Guid,
				"attributes": map[string]interface{}{
					"qualifiedName":          persona.QualifiedName,
					"name":                   persona.Name,
					"personaGroups":          groups,
					"isAccessControlEnabled": true,
				},
			},
		},
	}

	_, err := ac.sendRequest(ctx, url, http.MethodPost, requestBody, "UpdatePersonaGroups")
	return err
}

// GetDefaultPersona returns the default persona name
func (ac *AtlanClient) GetDefaultPersona() string {
	return ac.defaultPersona
}
