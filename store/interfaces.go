package store

import (
	"fm/models"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

type Store interface {
	GetFolderDetails(ctx fiber.Ctx, folderId uuid.UUID) (models.Folder, *httperrors.Error)
	CreateFolder(ctx fiber.Ctx, folder models.Folder) (models.Folder, *httperrors.Error)
	GetAllFolders(ctx fiber.Ctx) ([]models.Folder, *httperrors.Error)
}
