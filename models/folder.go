package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Folder struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" db:"id"`
	Name      string         `json:"name" gorm:"not null" db:"name" validate:"required,max=255"`
	ParentID  *uuid.UUID     `json:"parent_id" gorm:"type:uuid" db:"parent_id"`
	OwnerID   uuid.UUID      `json:"owner_id" gorm:"type:uuid;not null" db:"owner_id" validate:"required"`
	FullPath  string         `json:"full_path" gorm:"not null" db:"full_path" validate:"required"`
	CreatedAt time.Time      `json:"created_at" gorm:"default:now()" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"default:now()" db:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index" db:"deleted_at"`

	// Relationships
	Parent   *Folder  `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Children []Folder `json:"children,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Files    []File   `json:"files,omitempty" gorm:"foreignKey:FolderID;references:ID"`
}

type File struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()" db:"id"`
	Name       string         `json:"name" gorm:"not null" db:"name" validate:"required,max=255"`
	FolderID   uuid.UUID      `json:"folder_id" gorm:"type:uuid;not null" db:"folder_id" validate:"required"`
	FullPath   string         `json:"full_path" gorm:"not null" db:"full_path" validate:"required"`
	S3Key      string         `json:"s3_key" gorm:"not null" db:"s3_key" validate:"required"`
	Size       *int64         `json:"size" db:"size"`
	MimeType   *string        `json:"mime_type" db:"mime_type"`
	UploadedBy uuid.UUID      `json:"uploaded_by" gorm:"type:uuid;not null" db:"uploaded_by" validate:"required"`
	CreatedAt  time.Time      `json:"created_at" gorm:"default:now()" db:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"default:now()" db:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index" db:"deleted_at"`

	// Relationships
	Folder *Folder `json:"folder,omitempty" gorm:"foreignKey:FolderID;references:ID"`
}

// Request DTOs
type CreateFolderRequest struct {
	Name     string    `json:"name" validate:"required,max=255"`
	ParentID uuid.UUID `json:"parent_id"`
}

type UpdateFolderRequest struct {
	Name string `json:"name" validate:"required,max=255"`
}

type UploadFileRequest struct {
	Name     string    `json:"name" validate:"required,max=255"`
	FolderID uuid.UUID `json:"folder_id" validate:"required"`
	Size     *int64    `json:"size"`
	MimeType *string   `json:"mime_type"`
}

// Response DTOs
type FolderResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id"`
	FullPath  string     `json:"full_path"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type FileResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	FolderID  uuid.UUID `json:"folder_id"`
	FullPath  string    `json:"full_path"`
	Size      *int64    `json:"size"`
	MimeType  *string   `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
