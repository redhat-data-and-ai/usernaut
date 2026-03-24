package model

// Method represents the HTTP methods for the presigned URL request.
type Method string

const (
	GET Method = "GET"
	PUT Method = "PUT"
)

// PresignedURLRequest represents a request to generate a presigned URL.
type PresignedURLRequest struct {
	Key    string `json:"key"`
	Expiry string `json:"expiry"`
	Method Method `json:"method"`
}

type PresignedURLResponse struct {
	URL string `json:"url"`
}

// CloudStorageIdentifier represents cloud storage identifiers.
type CloudStorageIdentifier string

const (
	S3        CloudStorageIdentifier = "amazonaws.com"
	GCS       CloudStorageIdentifier = "storage.googleapis.com"
	AzureBlob CloudStorageIdentifier = "blob.core.windows.net"
)
