package models

import (
	"time"

	"github.com/google/uuid"
)

type Folder struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	OwnerID   uuid.UUID  `json:"owner_id"`
	FullPath  string     `json:"full_path"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
