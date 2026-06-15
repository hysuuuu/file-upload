-- Custom ENUM types for status tracking
CREATE TYPE upload_status AS ENUM ('pending', 'completed', 'aborted');
CREATE TYPE chunk_status AS ENUM ('uploading', 'success', 'failed');

-- upload_sessions: Tracks the overall state of a file upload session.
-- file_id is a unique identifier to handle concurrency and race conditions.
-- s3_upload_id stores the Multipart Upload ID returned by the storage layer (S3/MinIO).
CREATE TABLE IF NOT EXISTS upload_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_name TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    file_id TEXT NOT NULL,
    status upload_status NOT NULL DEFAULT 'pending',
    s3_upload_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- chunk_states: Tracks the state and integrity of individual file chunks/parts.
-- checksum stores the client-provided MD5/SHA256 for dual-layer validation.
-- etag is the identifier returned by S3 after a successful part upload.
CREATE TABLE IF NOT EXISTS chunk_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES upload_sessions(id) ON DELETE CASCADE,
    part_number INT NOT NULL,
    etag TEXT,
    checksum TEXT,
    status chunk_status NOT NULL DEFAULT 'uploading',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, part_number)
);

-- Utility function to automatically update the 'updated_at' column on row changes.
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to maintain 'updated_at' for both tables.
CREATE TRIGGER update_upload_sessions_updated_at BEFORE UPDATE ON upload_sessions FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_chunk_states_updated_at BEFORE UPDATE ON chunk_states FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
