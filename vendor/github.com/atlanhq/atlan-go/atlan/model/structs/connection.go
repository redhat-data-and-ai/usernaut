package structs

import (
	"time"

	"github.com/atlanhq/atlan-go/atlan"
)

type Connection struct {
	Asset
	Category                     *string                      `json:"category,omitempty"`
	SubCategory                  *string                      `json:"subCategory,omitempty"`
	Host                         *string                      `json:"host,omitempty"`
	Port                         *int                         `json:"port,omitempty"`
	AllowQuery                   *bool                        `json:"allowQuery,omitempty"`
	AllowQueryPreview            *bool                        `json:"allowQueryPreview,omitempty"`
	QueryPreviewConfig           *map[string]string           `json:"queryPreviewConfig,omitempty"`
	QueryConfig                  *string                      `json:"queryConfig,omitempty"`
	CredentialStrategy           *string                      `json:"credentialStrategy,omitempty"`
	PreviewCredentialStrategy    *string                      `json:"previewCredentialStrategy,omitempty"`
	PolicyStrategy               *string                      `json:"policyStrategy,omitempty"`
	QueryUsernameStrategy        *atlan.QueryUsernameStrategy `json:"queryUsernameStrategy,omitempty"`
	RowLimit                     *int                         `json:"rowLimit,omitempty"`
	QueryTimeout                 *int                         `json:"queryTimeout,omitempty"`
	DefaultCredentialGuid        *string                      `json:"defaultCredentialGuid,omitempty"`
	ConnectorIcon                *string                      `json:"connectorIcon,omitempty"`
	ConnectorImage               *string                      `json:"connectorImage,omitempty"`
	SourceLogo                   *string                      `json:"sourceLogo,omitempty"`
	IsSampleDataPreviewEnabled   *bool                        `json:"isSampleDataPreviewEnabled,omitempty"`
	PopularityInsightsTimeframe  *int                         `json:"popularityInsightsTimeframe,omitempty"`
	HasPopularityInsights        *bool                        `json:"hasPopularityInsights,omitempty"`
	ConnectionDbtEnvironments    *map[string]bool             `json:"connectionDbtEnvironments,omitempty"`
	ConnectionSSOCredentialGuid  *string                      `json:"connectionSSOCredentialGuid,omitempty"`
	UseObjectStorage             *bool                        `json:"useObjectStorage,omitempty"`
	ObjectStorageUploadThreshold *int                         `json:"objectStorageUploadThreshold,omitempty"`
	VectorEmbeddingsEnabled      *bool                        `json:"vectorEmbeddingsEnabled,omitempty"`
	VectorEmbeddingsUpdatedAt    *time.Time                   `json:"vectorEmbeddingsUpdatedAt,omitempty"`
}
