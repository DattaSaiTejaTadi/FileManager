package files

import (
	"fm/models"
	"fm/store"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

type service struct {
	fileStore   store.File
	folderStore store.Folder
	bucket      store.Bucket
}

func New(fileStore store.File, folderStore store.Folder, bucket store.Bucket) *service {
	return &service{fileStore: fileStore, folderStore: folderStore, bucket: bucket}
}

func (s *service) Create(ctx fiber.Ctx, file *models.File) (*models.File, *httperrors.Error) {
	var fullPath string

	if file.FolderId != uuid.Nil {
		parentfolder, err := s.folderStore.GetById(ctx, &file.FolderId)
		if err != nil {
			return nil, err
		}
		fullPath += parentfolder.FullPath + "/" + file.Name
	} else {
		fullPath = file.Name
	}
	fileObjectDetails, err := s.bucket.GeneratePresignedUploadURL(fullPath)
	if err != nil {
		return nil, err
	}
	file.FullPath = fileObjectDetails.S3Key
	file.UploadURL = fileObjectDetails.URL
	if file.Id == uuid.Nil {
		file.Id = uuid.New()
	}
	file.Id = uuid.New()
	return s.fileStore.Create(ctx, file)
}

func (s *service) GetById(ctx fiber.Ctx, id *uuid.UUID) (*models.File, *httperrors.Error) {
	return s.fileStore.GetById(ctx, *id)
}

func (s *service) GetFiles(ctx fiber.Ctx, parentFolderId uuid.UUID) ([]*models.File, *httperrors.Error) {
	return s.fileStore.GetFiles(ctx, parentFolderId)
}
