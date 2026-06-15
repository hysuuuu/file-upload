package repository

import (
	"context"

	"github.com/google/uuid"
)

// UploadRepository defines the data access layer for file uploads.
type UploadRepository interface {
	// CreateSession initializes a new upload record in the database.
	// It should take an UploadSession model and return the created record or an error.
	CreateSession(ctx context.Context, session *UploadSession) error

	// GetSessionByFileID retrieves a session using the unique FileID.
	// This is crucial for checking if an upload for a specific file already exists.
	GetSessionByFileID(ctx context.Context, fileID string) (*UploadSession, error)

	// UpdateSessionStatus updates the status of the overall upload.
	UpdateSessionStatus(ctx context.Context, status UploadStatus, id uuid.UUID) error

	// UpsertChunkState records or updates the progress of a specific part.
	// If a part was previously 'failed', this should allow updating it to 'success'.
	UpsertChunkState(ctx context.Context, chunk *ChunkState) error

	// GetChunksBySessionID retrieves all recorded chunks for a given session.
	// This is used for the "Reconciliation" phase to tell the client what's missing.
	GetChunksBySessionID(ctx context.Context, sessionID uuid.UUID) ([]ChunkState, error)
}
