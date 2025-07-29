package service

import (
	"fm/models"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

func (s *service) CreateFolder(ctx fiber.Ctx, folder models.CreateFolderRequest) (models.Folder, *httperrors.Error) {
	var newFolder models.Folder
	parentFolder, err := s.store.GetFolderDetails(ctx, folder.ParentID)
	if err != nil {
		return models.Folder{}, httperrors.NewDBError()
	}
	folderid := uuid.New()
	var folderPath = ""
	if parentFolder.FullPath == "" {
		folderPath = "/" + folderid.String()
	} else {
		newFolder.ParentID = &folder.ParentID
		folderPath = parentFolder.FullPath + "/" + folderid.String()
	}
	newFolder.Name = folder.Name
	newFolder.ID = folderid
	newFolder.FullPath = folderPath
	newFolder.OwnerID = uuid.New()
	fmt.Print(newFolder)
	s3err := s.s3Client.CreateFolder(newFolder.FullPath)
	if s3err != nil {
		return models.Folder{}, httperrors.NewDBError()
	}
	createdFolder, err := s.store.CreateFolder(ctx, newFolder)
	if err != nil {
		return models.Folder{}, err
	}
	return createdFolder, nil
}

func (s *service) GetAllFolders(ctx fiber.Ctx) ([]models.Folder, *httperrors.Error) {
	return s.store.GetAllFolders(ctx)
}
