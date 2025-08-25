package service

import (
	"fm/models"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

type File interface {
	Create(ctx fiber.Ctx, file *models.File) (*models.File, *httperrors.Error)
	GetById(ctx fiber.Ctx, id *uuid.UUID) (*models.File, *httperrors.Error)
	GetFiles(ctx fiber.Ctx, parentFolderId uuid.UUID) ([]*models.File, *httperrors.Error)
}

type Folder interface {
	Create(ctx fiber.Ctx, folder *models.Folder) (*models.Folder, *httperrors.Error)
	GetALL(ctx fiber.Ctx) ([]models.Folder, *httperrors.Error)
	GetById(ctx fiber.Ctx, id *uuid.UUID) (*models.Folder, *httperrors.Error)
	GetSubFolders(ctx fiber.Ctx, id *uuid.UUID) ([]models.Folder, *httperrors.Error)
}

type Bucket interface {
	CreateFolder(fullPath string) (*models.CreateObjectResponse, *httperrors.Error)
}
