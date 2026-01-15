package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/atlanhq/atlan-go/config"

	"github.com/atlanhq/atlan-go/atlan/logger"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

// AtlanClient defines the Atlan API client structure.
type AtlanClient struct {
	Session        *http.Client
	host           string
	ApiKey         string
	requestParams  map[string]interface{}
	logger         logger.Logger
	RoleClient     *RoleClient
	GroupClient    *GroupClient
	UserClient     *UserClient
	TokenClient    *TokenClient
	WorkflowClient *WorkflowClient
	SearchAssets
}

// DefaultAtlanClient represents the default AtlanClient instance.
var (
	DefaultAtlanClient   *AtlanClient
	DefaultAtlanTagCache *AtlanTagCache
	contentType          string
)

// Init initializes the default AtlanClient.
func Init() error {
	apiKey, baseURL := retrieveAPIConfig()

	// Normalize the baseURL
	baseURL = normalizeURL(baseURL)

	// Configure client and logger
	client, logger := configureClient()

	// Initialize default AtlanClient
	DefaultAtlanClient = &AtlanClient{
		Session:       client,
		host:          baseURL,
		ApiKey:        apiKey,
		requestParams: defaultRequestParams(apiKey),
		logger:        *logger,
		SearchAssets:  newDefaultSearchAssets(),
	}

	return nil
}

// Context creates a new AtlanClient with provided API key and base URL.
func Context(baseURL, apiKey string) (*AtlanClient, error) {
	// Normalize the baseURL
	baseURL = normalizeURL(baseURL)

	// Configure client and logger
	client, logger := configureClient()

	atlanClient := &AtlanClient{
		Session:       client,
		host:          baseURL,
		ApiKey:        apiKey,
		requestParams: defaultRequestParams(apiKey),
		logger:        *logger,
		SearchAssets:  newDefaultSearchAssets(),
	}

	// Set as default AtlanClient
	DefaultAtlanClient = atlanClient

	return atlanClient, nil
}

// NewContext initializes a new AtlanClient instance.
func NewContext() *AtlanClient {
	if err := Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize AtlanClient: %v", err))
	}

	return DefaultAtlanClient
}

// configureClient configures HTTP client and logger.
func configureClient() (*http.Client, *logger.Logger) {
	client := &http.Client{}

	var loggerInstance *logger.Logger

	// Check if the logger is already set by the user
	if DefaultAtlanClient != nil && DefaultAtlanClient.logger.Log != nil {
		loggerInstance = &DefaultAtlanClient.logger
	} else {
		// Configure logger with default values
		loggerCfg := &logger.LoggerConfig{Level: "info", Enabled: true}
		newLogger := logger.NewLogger(loggerCfg)
		loggerInstance = &newLogger
	}

	return client, loggerInstance
}

// defaultRequestParams returns default request parameters.
func defaultRequestParams(apiKey string) map[string]interface{} {
	headers := map[string]string{
		"x-atlan-agent":         "sdk",
		"x-atlan-agent-id":      "go",
		"x-atlan-client-origin": "product_sdk",
		"User-Agent":            fmt.Sprintf("Atlan-GOSDK/%s", config.Version()),
	}
	headers["Authorization"] = "Bearer " + apiKey
	headers["Accept"] = "application/json"
	headers["Content-type"] = "application/json"
	return map[string]interface{}{
		"headers": headers,
	}
}

// newDefaultSearchAssets initializes default SearchAssets for AtlanClient.
func newDefaultSearchAssets() SearchAssets {
	return SearchAssets{
		Glossary:         NewSearchGlossary(),
		Table:            NewSearchTable(),
		Column:           NewSearchColumn(),
		Connection:       NewSearchConnection(),
		MaterialisedView: NewSearchMaterialisedView(),
		View:             NewSearchView(),
		Persona:          NewPersonaFields(),
		AccessControl:    NewAccessControlFields(),
		AuthPolicy:       NewAuthPolicyFields(),
		Purpose:          NewPurposeFields(),
	}
}

// retrieveAPIConfig retrieves API configuration from environment variables.
func retrieveAPIConfig() (apiKey, baseURL string) {
	apiKey = os.Getenv("ATLAN_API_KEY")
	if apiKey == "" {
		logger.Log.Error("ATLAN_API_KEY not provided in environmental variables")
		panic("ATLAN_API_KEY not provided in environmental variables")
	}

	baseURL = os.Getenv("ATLAN_BASE_URL")
	if baseURL == "" {
		logger.Log.Error("ATLAN_BASE_URL not provided in environmental variables")
		panic("ATLAN_BASE_URL not provided in environmental variables")
	}

	return apiKey, baseURL
}

// normalizeURL ensures the URL starts with "https://" and truncates after the domain.
func normalizeURL(rawURL string) string {
	// Ensure URL starts with "http://" or "https://"
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("Invalid URL: %v", err))
	}

	// Truncate the URL after the host
	return parsedURL.Scheme + "://" + parsedURL.Host
}

// SetLogger enables or disables logging and sets the log level.
func (ac *AtlanClient) SetLogger(enabled bool, level string) {
	// Create a new logger configuration
	loggerCfg := &logger.LoggerConfig{
		Level:   level,
		Enabled: enabled,
	}
	// Create a new logger instance
	ac.logger = logger.NewLogger(loggerCfg)
}

// EnableLogging enables logging with the specified log level.
func (ac *AtlanClient) EnableLogging(level string) {
	ac.SetLogger(true, level)
}

// DisableLogging disables logging.
func (ac *AtlanClient) DisableLogging() {
	ac.SetLogger(false, "")
}

// Removes authorization from header when using
// s3PresignedUrlFileUpload/Download and returns the removed value.
func (ac *AtlanClient) removeAuthorization() (string, error) {
	if headers, ok := ac.requestParams["headers"].(map[string]string); ok {
		auth, exists := headers["Authorization"]
		if exists {
			delete(headers, "Authorization")
			return auth, nil
		}
		return "", nil
	} else {
		return "", InvalidRequestError{
			AtlanError{
				ErrorCode: errorCodes[UNABLE_TO_PERFORM_OPERATION_ON_AUTHORIZATION],
				Args:      []interface{}{"remove", "from"},
			},
		}
	}
}

// Restores the authorization to the header when using s3PresignedUrlFileUpload/Download .
func (ac *AtlanClient) restoreAuthorization(auth string) error {
	if headers, ok := ac.requestParams["headers"].(map[string]string); ok {
		headers["Authorization"] = auth
	} else {
		return InvalidRequestError{
			AtlanError{
				ErrorCode: errorCodes[UNABLE_TO_PERFORM_OPERATION_ON_AUTHORIZATION],
				Args:      []interface{}{"restore", "to"},
			},
		}
	}
	return nil
}

// Initialize the file progress bar using default configuration settings
func initFileProgressBar(fileSize int64, description string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions(int(fileSize),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionOnCompletion(func() {
			fmt.Printf("\n")
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[blue]=[reset]",
			SaucerHead:    "[blue]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	return bar
}

func (ac *AtlanClient) s3PresignedUrlFileUpload(api *API, uploadFile *os.File, uploadFileSize int64) error {
	// Remove authorization and returns the auth value
	auth, err := ac.removeAuthorization()
	if err != nil {
		return err
	}

	// Call the API with upload file options
	uploadProgressBarDescription := "Uploading file to the object store:"
	uploadProgressBar := initFileProgressBar(uploadFileSize, uploadProgressBarDescription)
	options := map[string]interface{}{
		"use_presigned_url": true,
		"file_size":         uploadFileSize,
		"progress_bar":      uploadProgressBar,
	}
	_, err = ac.CallAPI(api, nil, uploadFile, options)
	if err != nil {
		return err
	}

	// Restore authorization after API call
	err = ac.restoreAuthorization(auth)
	if err != nil {
		ac.logger.Errorf("failed to restore authorization: %s", err)
		return err
	}

	return nil
}

func (ac *AtlanClient) azureBlobPresignedUrlFileUpload(api *API, uploadFile *os.File, uploadFileSize int64) error {
	// Remove authorization and returns the auth value
	auth, err := ac.removeAuthorization()
	if err != nil {
		return err
	}

	// Add mandatory headers for Azure Blob storage
	if headers, ok := ac.requestParams["headers"].(map[string]string); ok {
		headers["x-ms-blob-type"] = "BlockBlob"
	}

	// Call the API with upload file options
	uploadProgressBarDescription := "Uploading file to the object store:"
	uploadProgressBar := initFileProgressBar(uploadFileSize, uploadProgressBarDescription)
	options := map[string]interface{}{
		"use_presigned_url": true,
		"file_size":         uploadFileSize,
		"progress_bar":      uploadProgressBar,
	}
	_, err = ac.CallAPI(api, nil, uploadFile, options)
	if err != nil {
		return err
	}

	// Restore authorization after API call
	err = ac.restoreAuthorization(auth)
	if err != nil {
		ac.logger.Errorf("failed to restore authorization: %s", err)
		return err
	}

	return nil
}

func (ac *AtlanClient) gcsPresignedUrlFileUpload(api *API, uploadFile *os.File, uploadFileSize int64) error {
	// Remove authorization and returns the auth value
	auth, err := ac.removeAuthorization()
	if err != nil {
		return err
	}

	// Call the API with upload file options
	uploadProgressBarDescription := "Uploading file to the object store:"
	uploadProgressBar := initFileProgressBar(uploadFileSize, uploadProgressBarDescription)
	options := map[string]interface{}{
		"use_presigned_url": true,
		"file_size":         uploadFileSize,
		"progress_bar":      uploadProgressBar,
	}
	_, err = ac.CallAPI(api, nil, uploadFile, options)
	if err != nil {
		return err
	}

	// Restore authorization after API call
	err = ac.restoreAuthorization(auth)
	if err != nil {
		ac.logger.Errorf("failed to restore authorization: %s", err)
		return err
	}

	return nil
}

func (ac *AtlanClient) s3PresignedUrlFileDownload(api *API, downloadFilePath string) error {
	// Remove authorization and returns the auth value
	auth, err := ac.removeAuthorization()
	if err != nil {
		return err
	}

	// Call the API with download file options
	downloadProgressBarDescription := "Downloading file from the object store:"
	downloadProgressBar := initFileProgressBar(0, downloadProgressBarDescription)
	options := map[string]interface{}{
		"use_presigned_url": true,
		"save_file":         true,
		"file_path":         downloadFilePath,
		"progress_bar":      downloadProgressBar,
	}
	_, err = ac.CallAPI(api, nil, nil, options)
	if err != nil {
		return err
	}

	// Restore authorization after API call
	err = ac.restoreAuthorization(auth)
	if err != nil {
		ac.logger.Errorf("failed to restore authorization: %s", err)
		return err
	}

	return nil
}

// CallAPI makes a generic API call.
func (ac *AtlanClient) CallAPI(api *API, queryParams interface{}, requestObj interface{}, options ...interface{}) ([]byte, error) {
	var saveFile bool
	var filePath string
	var fileProgressBar *progressbar.ProgressBar
	params := deepCopy(ac.requestParams)
	path := ac.host + api.Endpoint.Atlas + api.Path

	query := url.Values{}
	switch v := queryParams.(type) {
	case map[string]string:
		for key, value := range v {
			if key == "attr:qualifiedName" {
				path += fmt.Sprintf("?%s=%s", key, url.QueryEscape(value))
			} else {
				query.Add(key, value)
			}
		}
	case map[string][]string:
		for key, values := range v {
			for _, value := range values {
				query.Add(key, value)
			}
		}
	case map[string]interface{}:
		// When queryParams is map[string]interface{}, process accordingly
		for key, value := range v {
			switch v := value.(type) {
			case string:
				query.Add(key, v)
			case []string:
				for _, val := range v {
					query.Add(key, val)
				}
			default:
				// For unsupported types, you can log or handle differently
				params[key] = value
			}
		}
	default:
		params["params"] = queryParams
	}

	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	// Check for extra any API call options
	if len(options) > 0 {
		if optMap, ok := options[0].(map[string]interface{}); ok {
			if _, ok := optMap["save_file"].(bool); ok {
				saveFile = ok
			}
			if path, ok := optMap["file_path"].(string); ok {
				filePath = path
			}
			if fs, ok := optMap["file_size"].(int64); ok {
				params["content_length"] = fs
			}
			if _, ok := optMap["use_presigned_url"].(bool); ok {
				path = api.Path
			}
			if bar, ok := optMap["progress_bar"].(*progressbar.ProgressBar); ok {
				fileProgressBar = bar
			}
		}
	}

	if requestObj != nil {
		switch reqObj := requestObj.(type) {
		// In case of file upload/download
		case *os.File:
			if fileProgressBar != nil {
				params["progress_bar"] = fileProgressBar
				params["data"] = progressbar.NewReader(reqObj, fileProgressBar)
			}
			params["content_type"] = "application/octet-stream"
		default:
			// Otherwise just use `json.Marshal()`
			requestJSON, err := json.Marshal(requestObj)
			ac.logger.Debugf("Request JSON: %s", string(requestJSON))
			if err != nil {
				ac.logger.Errorf("error marshaling request object: %v", err)
				return nil, fmt.Errorf("error marshaling request object: %v", err)
			}
			params["data"] = bytes.NewBuffer(requestJSON)
		}
	}
	// Send the request
	response, err := ac.makeRequest(api.Method, path, params)
	if err != nil {
		return nil, handleApiError(response, err)
	}

	ac.logHTTPStatus(response)

	// Handle API error based on response status code
	if response.StatusCode != api.Status {
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			fmt.Printf("Error reading response body: %v\n", readErr)
			return nil, handleApiError(response, fmt.Errorf("error reading response body: %v", readErr))
		}

		// Create a descriptive error if `err` is nil
		var errorMessage string
		if err == nil {
			errorMessage = fmt.Sprintf("API returned status code %d: %s", response.StatusCode, string(body))
			err = fmt.Errorf("%s", errorMessage)
			//	fmt.Printf("Constructed error: %s\n", errorMessage) // Optional for debugging
		}
		return nil, handleApiError(response, err)
	}

	// Handle file download
	if saveFile {
		file, err := os.Create(filePath)
		if err != nil {
			return nil, AtlanError{
				ErrorCode: errorCodes[UNABLE_TO_PREPARE_DOWNLOAD_FILE],
				Args:      []interface{}{err.Error()},
			}
		}
		defer file.Close()

		// Set the progress bar size based on the response content-length
		fileProgressBar.ChangeMax64(response.ContentLength)
		_, err = io.Copy(io.MultiWriter(file, fileProgressBar), response.Body)
		if err != nil {
			return nil, AtlanError{
				ErrorCode: errorCodes[UNABLE_TO_COPY_DOWNLOAD_FILE_CONTENTS],
				Args:      []interface{}{err.Error()},
			}
		}

		ac.logger.Debugf("Successfully downloaded file: %s", file.Name())
		return []byte{}, nil
	}

	// Handle JSON response
	responseJSON, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Finally, close the request body
	response.Body.Close()

	ac.logResponse(responseJSON)

	return responseJSON, nil
}

// makeRequest makes an HTTP request.
func (ac *AtlanClient) makeRequest(method, path string, params map[string]interface{}) (*http.Response, error) {
	var req *http.Request
	var err error

	switch method {
	case http.MethodGet:
		req, err = http.NewRequest(method, path, nil)
		if err != nil {
			return nil, ThrowAtlanError(err, CONNECTION_ERROR, nil)
		}
	case http.MethodPost, http.MethodPut:
		var body io.Reader
		data, ok := params["data"]
		if !ok {
			return nil, fmt.Errorf("missing 'data' parameter for POST/PUT request")
		}
		switch requestData := data.(type) {
		case progressbar.Reader:
			// File data upload with progressbar reader
			body = &requestData
		case io.Reader:
			// JSON payload
			body = requestData
		default:
			return nil, fmt.Errorf("invalid 'data' parameter type for POST/PUT request")
		}
		req, err = http.NewRequest(method, path, body)
		if err != nil {
			return nil, ThrowAtlanError(err, CONNECTION_ERROR, nil)
		}
	case http.MethodDelete:
		// DELETE requests may not always have a body.
		var body io.Reader
		if data, ok := params["data"]; ok {
			body, ok = data.(io.Reader)
			if !ok {
				return nil, fmt.Errorf("invalid 'data' parameter for DELETE request")
			}
		}
		// Create a new http request
		req, err = http.NewRequest(method, path, body)
		if err != nil {
			return nil, ThrowAtlanError(err, CONNECTION_ERROR, nil)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Set request headers
	headers, ok := params["headers"].(map[string]string)
	if ok {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// Set content-type
	if ct, ok := params["content_type"].(string); ok {
		contentType = ct
	} else {
		// Default content type
		contentType = "application/json"
	}
	req.Header.Set("Content-Type", contentType)

	// Set content-length
	if contentLength, ok := params["content_length"].(int64); ok {
		req.ContentLength = contentLength
	}

	// Set query parameters
	queryParams, ok := params["params"].(map[string]string)
	if ok {
		// This implementation can be improved, not doing so since requires significant changes to codebase
		var query string
		for key, value := range queryParams {
			// Check if the key is "guid" and value contains commas
			if key == "guid" && strings.Contains(value, ",") {
				// Split the value by commas
				guids := strings.Split(value, ",")
				for _, guid := range guids {
					// Append each guid to the query string with &guid=
					query += "&guid=" + guid
				}
			} else {
				// For other keys, add them normally
				query += "&" + key + "=" + value
			}
		}
		// Remove the leading "&" from the query string
		if len(query) > 0 {
			query = query[1:]
		}
		req.URL.RawQuery = query
	}

	ac.logAPICall(req.Method, path, req)
	// Finally, execute the request
	return ac.Session.Do(req)
}

func (ac *AtlanClient) logAPICall(method, path string, request *http.Request) {
	ac.logger.Debugf("------------------------------------------------------")
	ac.logger.Debugf("Call         : %s %s", method, path)
	ac.logger.Debugf("Content-type : %s", request.Header.Get("Content-Type"))
	ac.logger.Debugf("Accept       : %s", request.Header.Get("Accept"))
}

func (ac *AtlanClient) logHTTPStatus(response *http.Response) {
	ac.logger.Debugf("HTTP Status: %s", response.Status)
}

func (ac *AtlanClient) logResponse(responseJSON []byte) {
	ac.logger.Debugf("<== __call_api %s", string(responseJSON))
}

func deepCopy(original map[string]interface{}) map[string]interface{} {
	dcopy := make(map[string]interface{})
	for key, value := range original {
		dcopy[key] = value
	}
	return dcopy
}
