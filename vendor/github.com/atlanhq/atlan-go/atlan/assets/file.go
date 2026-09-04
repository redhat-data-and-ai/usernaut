package assets

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/atlanhq/atlan-go/atlan/model"
)

// A client for operating on Atlan's tenant object storage.
type FileClient struct {
	*AtlanClient
}

// NewFileClient creates a new instance of FileClient.
func NewFileClient(client *AtlanClient) *FileClient {
	return &FileClient{client}
}

func handleFileUpload(filePath string) (*os.File, fs.FileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error while opening file: %v", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("error while getting file info: %v", err)
	}
	return file, fileInfo, nil
}

// Generates a presigned URL based on Atlan's tenant object store.
func (client *FileClient) GeneratePresignedURL(request *model.PresignedURLRequest) (string, error) {
	rawJSON, err := client.CallAPI(&PRESIGNED_URL, nil, request)
	if err != nil {
		return "", err
	}

	// Now unmarshal `rawJSON` to the `PresignedURLResponse`
	var response model.PresignedURLResponse
	err = json.Unmarshal(rawJSON, &response)
	if err != nil {
		return "", InvalidRequestError{
			AtlanError{
				ErrorCode: errorCodes[UNABLE_TO_PREPARE_UPLOAD_FILE],
				Args:      []interface{}{err.Error()},
			},
		}
	}
	return response.URL, nil
}

// Uploads a file to Atlan's object storage.
func (client *FileClient) UploadFile(presignedUrl string, filePath string) error {
	PRESIGNED_URL_UPLOAD_S3 := API{
		Path:     presignedUrl,
		Method:   http.MethodPut,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}
	PRESIGNED_URL_UPLOAD_AZURE_BLOB := API{
		Path:     presignedUrl,
		Method:   http.MethodPut,
		Status:   http.StatusCreated,
		Endpoint: HeraclesEndpoint,
	}
	PRESIGNED_URL_UPLOAD_GCS := API{
		Path:     presignedUrl,
		Method:   http.MethodPut,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	file, fileInfo, err := handleFileUpload(filePath)
	if err != nil {
		return InvalidRequestError{
			AtlanError{
				ErrorCode: errorCodes[UNABLE_TO_PREPARE_UPLOAD_FILE],
				Args:      []interface{}{err.Error()},
			},
		}
	}
	defer file.Close()

	// Currently supported upload methods for different cloud storage providers
	switch {
	case strings.Contains(presignedUrl, string(model.S3)):
		err = client.s3PresignedUrlFileUpload(&PRESIGNED_URL_UPLOAD_S3, file, fileInfo.Size())
	case strings.Contains(presignedUrl, string(model.AzureBlob)):
		err = client.azureBlobPresignedUrlFileUpload(&PRESIGNED_URL_UPLOAD_AZURE_BLOB, file, fileInfo.Size())
	case strings.Contains(presignedUrl, string(model.GCS)):
		err = client.gcsPresignedUrlFileUpload(&PRESIGNED_URL_UPLOAD_GCS, file, fileInfo.Size())
	default:
		return InvalidRequestError{AtlanError{ErrorCode: errorCodes[UNSUPPORTED_PRESIGNED_URL]}}
	}

	if err != nil {
		return err
	}
	return nil
}

// Downloads a file from Atlan's tenant object storage.
func (client *FileClient) DownloadFile(presignedUrl string, filePath string) error {
	PRESIGNED_URL_DOWNLOAD := API{
		Path:     presignedUrl,
		Method:   http.MethodGet,
		Status:   http.StatusOK,
		Endpoint: HeraclesEndpoint,
	}

	err := client.s3PresignedUrlFileDownload(&PRESIGNED_URL_DOWNLOAD, filePath)
	if err != nil {
		return err
	}
	return nil
}
