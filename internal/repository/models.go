package repository

import (
	"time"

	"github.com/google/uuid"
)

// UploadStatus represents the custom ENUM type in Postgres.
type UploadStatus string

const (
	StatusPending   UploadStatus = "pending"
	StatusCompleted UploadStatus = "completed"
	StatusAborted   UploadStatus = "aborted"
)

// ChunkStatus represents the custom ENUM type in Postgres.
type ChunkStatus string

const (
	ChunkUploading ChunkStatus = "uploading"
	ChunkSuccess   ChunkStatus = "success"
	ChunkFailed    ChunkStatus = "failed"
)

// UploadSession represents the 'upload_sessions' table.
type UploadSession struct {
	ID         uuid.UUID    `db:"id"`
	FileName   string       `db:"file_name"`
	FileSize   int64        `db:"file_size"`
	FileID     string       `db:"file_id"`
	Status     UploadStatus `db:"status"`
	S3UploadID string       `db:"s3_upload_id"`
	CreatedAt  time.Time    `db:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at"`
}

// ChunkState represents the 'chunk_states' table.
type ChunkState struct {
	ID         uuid.UUID   `db:"id"`
	SessionID  uuid.UUID   `db:"session_id"`
	PartNumber int         `db:"part_number"`
	ETag       *string     `db:"etag"`     // Pointer used for nullable column
	Checksum   *string     `db:"checksum"` // Pointer used for nullable column
	Status     ChunkStatus `db:"status"`
	CreatedAt  time.Time   `db:"created_at"`
	UpdatedAt  time.Time   `db:"updated_at"`
}
