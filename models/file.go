package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	FolderId   uuid.UUID `json:"folder_id"`
	FullPath   string    `json:"full_path"`
	UploadURL  string    `json:"upload_url"`
	S3Key      string    `json:"s3_key"`
	Size       int       `json:"size"`
	MimeType   string    `json:"mime_type"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	UploadedBy uuid.UUID `json:"uploaded_by"`
}
