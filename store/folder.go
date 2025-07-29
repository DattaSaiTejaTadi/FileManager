package store

import (
	"database/sql"
	"errors"
	"fm/models"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

func (S *store) GetFolderDetails(ctx fiber.Ctx, folderId uuid.UUID) (models.Folder, *httperrors.Error) {
	var folder models.Folder
	const query = ` SELECT id, name, parent_id, owner_id, full_path, created_at, updated_at 
        FROM folders 
        WHERE id = $1;`
	row := S.db.QueryRowContext(ctx.Context(), query, folderId)
	var parentID sql.NullString
	err := row.Scan(&folder.ID, &folder.Name, &parentID, &folder.OwnerID, &folder.FullPath, &folder.CreatedAt, &folder.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Folder{}, nil
		}
		return models.Folder{}, httperrors.NewDBError()
	}
	if parentID.Valid {
		pid, err := uuid.Parse(parentID.String)
		if err != nil {
			return models.Folder{}, httperrors.NewDBError()
		}
		folder.ParentID = &pid
	} else {
		folder.ParentID = nil
	}

	return folder, nil
}

func (S *store) CreateFolder(ctx fiber.Ctx, folder models.Folder) (models.Folder, *httperrors.Error) {
	var createdFolder models.Folder
	var parentID sql.NullString
	query := `INSERT INTO folders (id, name, parent_id, owner_id, full_path)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, name, parent_id, owner_id, full_path, created_at, updated_at`
	rows := S.db.QueryRowContext(ctx.Context(), query, folder.ID, folder.Name, folder.ParentID, folder.OwnerID, folder.FullPath)
	err := rows.Scan(&createdFolder.ID, &createdFolder.Name, &parentID, &createdFolder.OwnerID, &createdFolder.FullPath, &createdFolder.CreatedAt, &createdFolder.UpdatedAt)
	if err != nil {
		fmt.Print(err.Error())
		return models.Folder{}, httperrors.NewDBError()
	}
	if parentID.Valid {
		pid, err := uuid.Parse(parentID.String)
		if err != nil {
			fmt.Print(err.Error())
			return models.Folder{}, httperrors.NewDBError()
		}
		createdFolder.ParentID = &pid
	} else {
		createdFolder.ParentID = nil
	}
	return createdFolder, nil
}

func (S *store) GetAllFolders(ctx fiber.Ctx) ([]models.Folder, *httperrors.Error) {
	const query = ` SELECT id, name, parent_id, owner_id, full_path, created_at, updated_at 
        FROM folders;`
	rows, err := S.db.QueryContext(ctx.Context(), query)

	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	folders := make([]models.Folder, 0)
	for rows.Next() {
		var folder models.Folder
		err := rows.Scan(&folder.ID, &folder.Name, &folder.ParentID, &folder.OwnerID, &folder.FullPath, &folder.CreatedAt, &folder.UpdatedAt)
		if err != nil {
			fmt.Print(err.Error())
			return nil, httperrors.NewDBError()
		}
		folders = append(folders, folder)
	}
	return folders, nil
}
