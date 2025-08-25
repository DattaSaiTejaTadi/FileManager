package models

import "github.com/google/uuid"

type CreateObjectResponse struct {
	Key string    `json:"Key"`
	Id  uuid.UUID `json:"Id"`
}

type UploadSignedURLResponse struct {
	URL   string `json:"url"`
	Token string `json:"token"`
	S3Key string `json:"s3Key"`
}
