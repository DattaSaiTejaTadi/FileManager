package files

import (
	"fm/models"
	"fm/service"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

type handler struct {
	svc service.File
}

func New(s service.File) *handler {
	return &handler{svc: s}
}

func (h *handler) Create(ctx fiber.Ctx) error {
	var file models.File
	err := ctx.Bind().JSON(&file)
	if err != nil {
		validationError := httperrors.BodyValidationError()
		statuscode, errResp := validationError.ErrorResponse()
		ctx.Status(statuscode).JSON(errResp)
		return nil
	}
	fileResp, serviceError := h.svc.Create(ctx, &file)

	if serviceError != nil {
		statusCode, errResp := serviceError.ErrorResponse()
		ctx.Status(statusCode).JSON(errResp)
		return nil
	}

	ctx.Status(fiber.StatusCreated).JSON(models.Response{
		Message: "File created successfully",
		Data:    fileResp,
	})
	return nil
}

func (h *handler) GetById(ctx fiber.Ctx) error {
	id := ctx.Params("id")
	folderId, err := uuid.Parse(id)
	if err != nil {
		validationError := httperrors.BodyValidationError()
		statuscode, errResp := validationError.ErrorResponse()
		ctx.Status(statuscode).JSON(errResp)
		return nil
	}

	fileResp, serviceError := h.svc.GetById(ctx, &folderId)
	if serviceError != nil {
		statusCode, errResp := serviceError.ErrorResponse()
		ctx.Status(statusCode).JSON(errResp)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(models.Response{
		Message: "File retrieved successfully",
		Data:    fileResp,
	})
	return nil
}

func (h *handler) GetFiles(ctx fiber.Ctx) error {
	parentFolderId := ctx.Params("folderId")
	folderId, err := uuid.Parse(parentFolderId)
	if err != nil {
		validationError := httperrors.BodyValidationError()
		statuscode, errResp := validationError.ErrorResponse()
		ctx.Status(statuscode).JSON(errResp)
		return nil
	}

	fileResp, serviceError := h.svc.GetFiles(ctx, folderId)
	if serviceError != nil {
		statusCode, errResp := serviceError.ErrorResponse()
		ctx.Status(statusCode).JSON(errResp)
		return nil
	}

	ctx.Status(fiber.StatusOK).JSON(models.Response{
		Message: "Files retrieved successfully",
		Data:    fileResp,
	})
	return nil
}
