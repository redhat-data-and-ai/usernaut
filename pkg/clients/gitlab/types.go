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

package gitlab

import (
	"github.com/gojek/heimdall/v7"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

var (
	ldapProvider = "ldapmain"
)

type GitlabClient struct {
	gitlabClient    *gitlab.Client
	gitlabConfig    *GitlabConfig
	ldapSync        bool
	dependantExists bool
	cn              string
	httpClient      heimdall.Doer
}

type GitlabConfig struct {
	URL           string `json:"url"`
	Token         string `json:"token"`
	ParentGroupId int    `json:"parent_group_id"`
}
