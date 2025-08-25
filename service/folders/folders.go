package folders

import (
	"fm/models"
	"fm/store"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

type service struct {
	folder store.Folder
	bucket store.Bucket
}

func New(f store.Folder, b store.Bucket) *service {
	return &service{
		folder: f,
		bucket: b,
	}
}

func (s *service) Create(ctx fiber.Ctx, folder *models.Folder) (*models.Folder, *httperrors.Error) {

	var folderPath string

	// if parent exist append the path of parent
	if folder.ParentID != nil {

		parentFolder, err := s.folder.GetById(ctx, folder.ParentID)
		if err != nil {
			return nil, err
		}

		folderPath += parentFolder.FullPath

	}
	// in case no parent folder exist it will create normally and it will be a root folder
	folderPath += "/" + folder.Name

	// from the fullpath create the folder in bucket
	folderObjectDetails, err := s.bucket.CreateFolder(folderPath)
	if err != nil {
		return nil, err
	}

	// Assigne folder values
	folder.ID = folderObjectDetails.Id
	log.Println(folderObjectDetails.Id)
	log.Println(folder.ID)
	folder.FullPath = folderPath
	folder.CreatedAt = time.Now()
	folder.UpdatedAt = time.Now()

	folderResult, err := s.folder.Create(ctx, folder)
	if err != nil {
		return nil, err
	}

	return folderResult, nil
}

func (s *service) GetALL(ctx fiber.Ctx) ([]models.Folder, *httperrors.Error) {
	folders, err := s.folder.GetALL(ctx)
	if err != nil {
		return nil, err
	}
	return folders, nil
}

func (s *service) GetById(ctx fiber.Ctx, id *uuid.UUID) (*models.Folder, *httperrors.Error) {
	folder, err := s.folder.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *service) GetSubFolders(ctx fiber.Ctx, id *uuid.UUID) ([]models.Folder, *httperrors.Error) {
	folders, err := s.folder.GetSubFolders(ctx, id)
	if err != nil {
		return nil, err
	}
	return folders, nil
}