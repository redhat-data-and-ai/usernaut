package assets

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/atlanhq/atlan-go/atlan"

	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

// SnowflakeMiner represents a Snowflake miner package
type SnowflakeMiner struct {
	*AbstractMiner
	AdvancedConfig bool
	Name           string
	PackageName    string
	PackagePrefix  string
	ConnectorType  string
	PackageIcon    string
	PackageLogo    string
}

// NewSnowflakeMiner initializes a new Snowflake miner.
//
// Param:
//   - connectionQualifiedName: the qualified name of the connection to use for the miner
//
// Returns:
//   - SnowflakeMiner instance, initialized with the provided connection qualified name and default values
func NewSnowflakeMiner(connectionQualifiedName string) *SnowflakeMiner {
	return &SnowflakeMiner{
		AbstractMiner: NewAbstractMiner(
			connectionQualifiedName,
			"@atlan/snowflake-miner",
			"atlan-snowflake-miner",
		),
		AdvancedConfig: false,
		Name:           "snowflake",
		PackageName:    "@atlan/snowflake-miner",
		PackagePrefix:  atlan.WorkflowPackageSnowflakeMiner.Name,
		ConnectorType:  "snowflake",
		PackageIcon:    "https://docs.snowflake.com/en/_images/logo-snowflake-sans-text.png",
		PackageLogo:    "https://1amiydhcmj36tz3733v94f15-wpengine.netdna-ssl.com/wp-content/themes/snowflake/assets/img/logo-blue.svg",
	}
}

// Direct sets up the miner to extract directly from Snowflake using the specified start epoch and database/schema.
//
// Param:
//   - startEpoch: the epoch time from which to start mining
//   - database: the database name to extract from (can be empty for the default database)
//   - schema: the schema name to extract from (can be empty for the default schema)
//
// Returns:
//   - SnowflakeMiner instance, set up for direct extraction from Snowflake
func (s *SnowflakeMiner) Direct(startEpoch int64, database, schema string) *SnowflakeMiner {
	if database == "" && schema == "" {
		s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "snowflake-database", Value: "default"})
	} else {
		s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "database-name", Value: database})
		s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "schema-name", Value: schema})
	}

	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "extraction-method", Value: "query_history"})
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "miner-start-time-epoch", Value: strconv.FormatInt(startEpoch, 10)})

	return s
}

// S3 sets up the miner to extract from S3 using JSON line-separated files.
//
// Parameters:
//   - s3Bucket: S3 bucket where the JSON line-separated files are located
//   - s3Prefix: Prefix within the S3 bucket where the JSON files are stored
//   - sqlQueryKey: JSON key containing the query definition
//   - defaultDatabaseKey: JSON key containing the default database name
//   - defaultSchemaKey: JSON key containing the default schema name
//   - sessionIDKey: JSON key containing the session ID of the SQL query
//   - s3BucketRegion: (Optional) Region of the S3 bucket if applicable
//
// Returns:
//   - SnowflakeMiner instance, set up to extract from S3
func (s *SnowflakeMiner) S3(
	s3Bucket string,
	s3Prefix string,
	sqlQueryKey string,
	defaultDatabaseKey string,
	defaultSchemaKey string,
	sessionIDKey string,
	s3BucketRegion *string,
) *SnowflakeMiner {
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "extraction-method", Value: "s3"})
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "extraction-s3-bucket", Value: s3Bucket})
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "extraction-s3-prefix", Value: s3Prefix})
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "sql-json-key", Value: sqlQueryKey})
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "catalog-json-key", Value: defaultDatabaseKey})
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "schema-json-key", Value: defaultSchemaKey})
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "session-json-key", Value: sessionIDKey})

	// Add S3 bucket region only if provided
	if s3BucketRegion != nil {
		s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "extraction-s3-region", Value: *s3BucketRegion})
	}

	return s
}

// ExcludeUsers excludes certain users from being considered in the usage metrics calculation for assets (e.g., system users).
//
// Param:
//   - users: a list of user names to exclude from the usage metrics
//
// Returns:
//   - SnowflakeMiner instance, updated with the specified users to exclude
func (s *SnowflakeMiner) ExcludeUsers(users []string) *SnowflakeMiner {
	userJSON, _ := json.Marshal(users)
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "popularity-exclude-user-config", Value: string(userJSON)})
	return s
}

// PopularityWindow sets the number of days to consider for calculating popularity metrics for assets.
//
// Param:
//   - days: number of days to use for the popularity window (default is 30)
//
// Returns:
//   - SnowflakeMiner instance, updated with the popularity window configuration
func (s *SnowflakeMiner) PopularityWindow(days int) *SnowflakeMiner {
	s.AdvancedConfig = true
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "popularity-window-days", Value: strconv.Itoa(days)})
	return s
}

// NativeLineage enables or disables the use of Snowflake's native lineage feature for tracking lineage information.
//
// Param:
//   - enabled: if true, native lineage from Snowflake will be enabled
//
// Returns:
//   - SnowflakeMiner instance, updated with the native lineage setting
func (s *SnowflakeMiner) NativeLineage(enabled bool) *SnowflakeMiner {
	s.AdvancedConfig = true
	s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "native-lineage-active", Value: fmt.Sprintf("%t", enabled)})
	return s
}

// CustomConfig sets a custom configuration JSON for the Snowflake miner, allowing experimental feature flags or custom settings.
//
// Param:
//   - config: a map of custom configurations to be applied to the miner
//
// Returns:
//   - SnowflakeMiner instance, updated with the custom configuration
func (s *SnowflakeMiner) CustomConfig(config map[string]interface{}) *SnowflakeMiner {
	if len(config) > 0 {
		configJSON, _ := json.Marshal(config)
		s.Parameters = append(s.Parameters, structs.NameValuePair{Name: "control-config", Value: string(configJSON)})
	}
	s.AdvancedConfig = true
	return s
}

// GetMetadata generates workflow metadata for Snowflake Miner
func (s *SnowflakeMiner) GetMetadata() *structs.WorkflowMetadata {
	return &structs.WorkflowMetadata{
		Name:      structs.StringPtr(fmt.Sprintf("%s-%d", s.PackagePrefix, s.Epoch)),
		Namespace: structs.StringPtr("default"),
		Labels: map[string]string{
			"orchestration.atlan.com/certified":      "true",
			"orchestration.atlan.com/source":         s.Name,
			"orchestration.atlan.com/sourceCategory": "warehouse",
			"orchestration.atlan.com/type":           "miner",
			"orchestration.atlan.com/verified":       "true",
			"package.argoproj.io/installer":          "argopm",
			"package.argoproj.io/name":               fmt.Sprintf("%s-miner", s.Name),
			"package.argoproj.io/registry":           "httpsc-o-l-o-ns-l-a-s-hs-l-a-s-hpackages.atlan.com",
			"orchestration.atlan.com/atlan-ui":       "true",
		},
		Annotations: map[string]string{
			"orchestration.atlan.com/allowSchedule":   "true",
			"orchestration.atlan.com/categories":      "warehouse,miner",
			"orchestration.atlan.com/docsUrl":         "https://ask.atlan.com/hc/en-us/articles/6482067592337",
			"orchestration.atlan.com/emoji":           "\\uD83D\\uDE80",
			"orchestration.atlan.com/icon":            s.PackageIcon,
			"orchestration.atlan.com/logo":            s.PackageLogo,
			"orchestration.atlan.com/marketplaceLink": fmt.Sprintf("https://packages.atlan.com/-/web/detail/%s", s.PackageName),
			"orchestration.atlan.com/name":            "Snowflake Miner",
			"package.argoproj.io/author":              "Atlan",
			"package.argoproj.io/description":         "Package to mine query history data from Snowflake and store it for further processing. The data mined will be used for generating lineage and usage metrics.", //nolint
			"package.argoproj.io/homepage":            fmt.Sprintf("https://packages.atlan.com/-/web/detail/%s", s.PackageName),
			"package.argoproj.io/keywords":            `["snowflake","warehouse","connector","miner"]`,
			"package.argoproj.io/name":                s.PackageName,
			"package.argoproj.io/registry":            "https://packages.atlan.com",
			"package.argoproj.io/repository":          "https://github.com/atlanhq/marketplace-packages.git",
			"package.argoproj.io/support":             "support@atlan.com",
			"orchestration.atlan.com/atlanName":       fmt.Sprintf("%s-%d", s.PackagePrefix, s.Epoch),
		},
	}
}

// ToWorkflow generates a workflow from the miner configuration
func (s *SnowflakeMiner) ToWorkflow() *structs.Workflow {
	workflow := s.AbstractPackage.ToWorkflow()
	workflow.Metadata = s.GetMetadata() // Override metadata
	return workflow
}
