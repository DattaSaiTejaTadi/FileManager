package service

import (
	"fm/models"

	"github.com/gofiber/fiber/v3"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

type Service interface {
	CreateFolder(ctx fiber.Ctx, folder models.CreateFolderRequest) (models.Folder, *httperrors.Error)
	GetAllFolders(ctx fiber.Ctx) ([]models.Folder, *httperrors.Error) 
}
