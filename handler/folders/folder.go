package folders

import (
	"fm/models"
	"fm/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/syntaxLabz/errors/pkg/codes"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

type handlers struct {
	svc service.Folder
}

func New(s service.Folder) *handlers {
	return &handlers{svc: s}
}

func (h *handlers) Create(ctx fiber.Ctx) error {
	var folder models.Folder
	err := ctx.Bind().JSON(&folder)
	if err != nil {
		validationError := httperrors.BodyValidationError()
		statuscode, errResp := validationError.ErrorResponse()
		ctx.Status(statuscode).JSON(errResp)
		return nil
	}
	folderResp, serviceError := h.svc.Create(ctx, &folder)

	if serviceError != nil {
		statusCode, errResp := serviceError.ErrorResponse()
		ctx.Status(statusCode).JSON(errResp)
		return nil
	}

	ctx.Status(fiber.StatusCreated).JSON(models.Response{
		Message: "Folder created successfully",
		Data:    folderResp,
	})
	return nil
}

func (h *handlers) GetALL(ctx fiber.Ctx) error {
	folders, serviceError := h.svc.GetALL(ctx)

	if serviceError != nil {
		statusCode, errResp := serviceError.ErrorResponse()
		ctx.Status(statusCode).JSON(errResp)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(models.Response{
		Message: "Folders retrieved successfully",
		Data:    folders,
	})
	return nil
}

func (h *handlers) GetById(ctx fiber.Ctx) error {
	id := ctx.Params("id")
	folderId, err := uuid.Parse(id)
	if err != nil {
		statusCode, errResp := httperrors.New(codes.BadRequest, "Invalid folder ID").ErrorResponse()
		ctx.Status(statusCode).JSON(errResp)
		return nil
	}
	folder, serviceError := h.svc.GetById(ctx, &folderId)

	if serviceError != nil {
		statusCode, errResp := serviceError.ErrorResponse()
		ctx.Status(statusCode).JSON(errResp)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(models.Response{
		Message: "Folder retrieved successfully",
		Data:    folder,
	})
	return nil
}

func (h *handlers) GetSubFolders(ctx fiber.Ctx) error {
	id := ctx.Params("id")
	folderId, err := uuid.Parse(id)
	if err != nil {
		statusCode, errResp := httperrors.New(codes.BadRequest, "Invalid folder ID").ErrorResponse()
		ctx.Status(statusCode).JSON(errResp)
		return nil
	}
	folders, serviceError := h.svc.GetSubFolders(ctx, &folderId)

	if serviceError != nil {
		statusCode, errResp := serviceError.ErrorResponse()
		ctx.Status(statusCode).JSON(errResp)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(models.Response{
		Message: "Subfolders retrieved successfully",
		Data:    folders,
	})
	return nil
}
