package handler

import (
	"fm/models"

	"github.com/gofiber/fiber/v3"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

func (H *handler) CreateFolder(ctx fiber.Ctx) error {
	var folderReq models.CreateFolderRequest
	if err := ctx.Bind().JSON(&folderReq); err != nil {
		statuscode, errresp := httperrors.BodyValidationError().ErrorResponse()
		ctx.Status(statuscode).JSON(errresp)
		return nil
	}
	if folderReq.Name == "" {
		statuscode, errresp := httperrors.BodyValidationError().ErrorResponse()
		ctx.Status(statuscode).JSON(errresp)
		return nil
	}
	folder, err := H.service.CreateFolder(ctx, folderReq)
	if err != nil {
		statuscode, errresp := err.ErrorResponse()
		ctx.Status(statuscode).JSON(errresp)
		return nil
	}
	ctx.Status(200).JSON(map[string]any{
		"code":    200,
		"message": "Successfully create folder",
		"data":    folder,
	})
	return nil
}

func (H *handler) GetAllFolders(ctx fiber.Ctx) error {
	folders, err := H.service.GetAllFolders(ctx)
	if err != nil {
		statuscode, errresp := err.ErrorResponse()
		ctx.Status(statuscode).JSON(errresp)
		return nil
	}
	ctx.Status(200).JSON(map[string]any{
		"code":    200,
		"message": "Successfully retrieved folders",
		"data":    folders,
	})
	return nil
}
