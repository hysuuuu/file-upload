package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type postgresUploadRepository struct {
	db *sqlx.DB
}

// NewPostgresUploadRepository creates a new instance of the Postgres implementation.
func NewPostgresUploadRepository(db *sqlx.DB) UploadRepository {
	return &postgresUploadRepository{db: db}
}

func (r *postgresUploadRepository) CreateSession(ctx context.Context, session *UploadSession) error {
	query := `
		INSERT INTO upload_sessions (file_name, file_size, file_id, status, s3_upload_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		session.FileName,
		session.FileSize,
		session.FileID,
		session.Status,
		session.S3UploadID,
	).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)

	return err
}

func (r *postgresUploadRepository) GetSessionByFileID(ctx context.Context, fileID string) (*UploadSession, error) {
	var session UploadSession
	query := `SELECT * FROM upload_sessions WHERE file_id = $1`

	err := r.db.GetContext(ctx, &session, query, fileID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *postgresUploadRepository) UpdateSessionStatus(ctx context.Context, status UploadStatus, id uuid.UUID) error {
	query := `
		UPDATE upload_sessions 
		SET status = $1, updated_at = NOW()
		WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *postgresUploadRepository) UpsertChunkState(ctx context.Context, chunk *ChunkState) error {
	query := `
		INSERT INTO chunk_states (session_id, part_number, etag, checksum, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (session_id, part_number)
		DO UPDATE SET
			etag = EXCLUDED.etag,
			checksum = EXCLUDED.checksum,
			status = EXCLUDED.status,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		chunk.SessionID,
		chunk.PartNumber,
		chunk.ETag,
		chunk.Checksum,
		chunk.Status,
	).Scan(&chunk.ID, &chunk.CreatedAt, &chunk.UpdatedAt)

	return err
}

func (r *postgresUploadRepository) GetChunksBySessionID(ctx context.Context, sessionID uuid.UUID) ([]ChunkState, error) {
	var chunks []ChunkState
	query := `
		SELECT * FROM chunk_states 
		WHERE session_id = $1 
		ORDER BY part_number ASC`

	err := r.db.SelectContext(ctx, &chunks, query, sessionID)
	if err != nil {
		return nil, err
	}

	return chunks, nil
}
