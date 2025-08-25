package files

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

func (s *store) Create(ctx fiber.Ctx, file *models.File) (*models.File, *httperrors.Error) {
	query := `INSERT INTO files (id, name, folder_id, full_path, upload_url, s3_key, size, mime_type, created_at, updated_at, uploaded_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	now := time.Now().UTC()
	if file.CreatedAt.IsZero() {
		file.CreatedAt = now
	}
	file.UpdatedAt = now

	if file.Id == uuid.Nil {
		file.Id = uuid.New()
	}

	_, err := s.db.ExecContext(ctx.Context(), query,
		file.Id,
		file.Name,
		file.FolderId,
		file.FullPath,
		file.UploadURL,
		file.S3Key,
		file.Size,
		file.MimeType,
		file.CreatedAt,
		file.UpdatedAt,
		file.UploadedBy,
	)
	if err != nil {
		return nil, httperrors.New(codes.InternalServerError, err.Error())
	}
	return file, nil
}

func (s *store) GetById(ctx fiber.Ctx, id uuid.UUID) (*models.File, *httperrors.Error) {
	query := `SELECT id, name, folder_id, full_path, upload_url, s3_key, size, mime_type, created_at, updated_at, uploaded_by FROM files WHERE id = $1`
	row := s.db.QueryRowContext(ctx.Context(), query, id)
	var file models.File
	err := row.Scan(
		&file.Id,
		&file.Name,
		&file.FolderId,
		&file.FullPath,
		&file.UploadURL,
		&file.S3Key,
		&file.Size,
		&file.MimeType,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.UploadedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, httperrors.New(codes.NotFound, "File not found")
		}
		return nil, httperrors.New(codes.InternalServerError, err.Error())
	}
	return &file, nil
}

func (s *store) GetFiles(ctx fiber.Ctx, parentFolderId uuid.UUID) ([]*models.File, *httperrors.Error) {
	query := `SELECT id, name, folder_id, full_path, upload_url, s3_key, size, mime_type, created_at, updated_at, uploaded_by FROM files WHERE folder_id = $1`
	rows, err := s.db.QueryContext(ctx.Context(), query, parentFolderId)
	if err != nil {
		return nil, httperrors.New(codes.InternalServerError, err.Error())
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.Id,
			&file.Name,
			&file.FolderId,
			&file.FullPath,
			&file.UploadURL,
			&file.S3Key,
			&file.Size,
			&file.MimeType,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.UploadedBy,
		); err != nil {
			return nil, httperrors.New(codes.InternalServerError, err.Error())
		}
		files = append(files, &file)
	}
	if err := rows.Err(); err != nil {
		return nil, httperrors.New(codes.InternalServerError, err.Error())
	}
	return files, nil
}
