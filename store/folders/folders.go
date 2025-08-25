package folders

import (
	"database/sql"
	"fm/models"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/syntaxLabz/errors/pkg/codes"
	"github.com/syntaxLabz/errors/pkg/httperrors"
)

type store struct {
	db *sql.DB
}

func New(db *sql.DB) *store {
	return &store{db: db}
}

func (s *store) Create(ctx fiber.Ctx, folder *models.Folder) (*models.Folder, *httperrors.Error) {
	query := `INSERT INTO folders (id, name, parent_id, owner_id, full_path, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	// Set timestamps
	now := time.Now().UTC()
	if folder.CreatedAt.IsZero() {
		folder.CreatedAt = now
	}
	folder.UpdatedAt = now

	_, err := s.db.ExecContext(ctx.Context(), query,
		folder.ID,
		folder.Name,
		folder.ParentID,
		folder.OwnerID,
		folder.FullPath,
		folder.CreatedAt,
		folder.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation (Postgres and SQLite)
		if err.Error() != "" && ( // SQLite
		err.Error() == "UNIQUE constraint failed: folders.name" ||
			// Postgres
			err.Error() == "pq: duplicate key value violates unique constraint \"folders_pkey\"") {
			return nil, httperrors.New(codes.Conflict, "Folder name already exists")
		}
		return nil, httperrors.New(codes.InternalServerError, err.Error())
	}
	return folder, nil
}

func (s *store) GetById(ctx fiber.Ctx, id *uuid.UUID) (*models.Folder, *httperrors.Error) {
	query := `SELECT id, name, parent_id, owner_id, full_path, created_at, updated_at FROM folders WHERE id = $1`
	row := s.db.QueryRowContext(ctx.Context(), query, &id)
	var folder models.Folder
	err := row.Scan(
		&folder.ID,
		&folder.Name,
		&folder.ParentID,
		&folder.OwnerID,
		&folder.FullPath,
		&folder.CreatedAt,
		&folder.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperrors.New(codes.NotFound, "Folder not found")
		}
		return nil, httperrors.New(codes.InternalServerError, err.Error())
	}
	return &folder, nil
}

func (s *store) GetALL(ctx fiber.Ctx) ([]models.Folder, *httperrors.Error) {
	query := `SELECT id, name, parent_id, owner_id, full_path, created_at, updated_at FROM folders`
	rows, err := s.db.QueryContext(ctx.Context(), query)
	if err != nil {
		return nil, httperrors.New(codes.InternalServerError, err.Error())
	}
	defer rows.Close()

	var folders []models.Folder
	for rows.Next() {
		var folder models.Folder
		if err := rows.Scan(
			&folder.ID,
			&folder.Name,
			&folder.ParentID,
			&folder.OwnerID,
			&folder.FullPath,
			&folder.CreatedAt,
			&folder.UpdatedAt,
		); err != nil {
			return nil, httperrors.New("internal_error", err.Error())
		}
		folders = append(folders, folder)
	}
	if err := rows.Err(); err != nil {
		return nil, httperrors.New("internal_error", err.Error())
	}
	return folders, nil
}

func (s *store) GetSubFolders(ctx fiber.Ctx, id *uuid.UUID) ([]models.Folder, *httperrors.Error) {
	query := `SELECT id, name, parent_id, owner_id, full_path, created_at, updated_at FROM folders WHERE parent_id = $1`
	rows, err := s.db.QueryContext(ctx.Context(), query, id)
	if err != nil {
		return nil, httperrors.New(codes.InternalServerError, err.Error())
	}
	defer rows.Close()

	var folders []models.Folder
	for rows.Next() {
		var folder models.Folder
		if err := rows.Scan(
			&folder.ID,
			&folder.Name,
			&folder.ParentID,
			&folder.OwnerID,
			&folder.FullPath,
			&folder.CreatedAt,
			&folder.UpdatedAt,
		); err != nil {
			return nil, httperrors.New("internal_error", err.Error())
		}
		folders = append(folders, folder)
	}
	if err := rows.Err(); err != nil {
		return nil, httperrors.New("internal_error", err.Error())
	}
	return folders, nil
}
