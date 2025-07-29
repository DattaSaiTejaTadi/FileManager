package models

import "time"

type S3Config struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
}
type FileInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
	IsFolder     bool
}

type PresignedURL struct {
	URL        string
	ExpiresAt  time.Time
	HTTPMethod string
	FilePath   string
}

type PresignedUploadOptions struct {
	FilePath      string
	ContentType   string
	MaxFileSize   int64
	ExpiryMinutes int
}

type PresignedDownloadOptions struct {
	FilePath      string
	ExpiryMinutes int
	Filename      string // Optional: Custom filename for download
}
