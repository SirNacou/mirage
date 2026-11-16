-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE IF NOT EXISTS global_sync_seq START 1;

CREATE TABLE chunks (
  hash TEXT PRIMARY KEY,
  size BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
);

CREATE TABLE files (
  root_id TEXT NOT NULL,
  relative_path TEXT NOT NULL,
  size BIGINT NOT NULL,
  mtime_unix BIGINT NOT NULL,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  chunks JSONB NOT NULL DEFAULT '[]'::jsonb,
  sync_seq BIGINT NOT NULL DEFAULT nextval('global_sync_seq'),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (root_id, relative_path)
);

CREATE INDEX idx_files_root_seq ON files (root_id, sync_seq);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS chunks;
DROP SEQUENCE IF EXISTS global_sync_seq;

-- +goose StatementEnd