package ed2ksrv

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/monkeyWie/goed2k/protocol"
)

const dbOperationTimeout = 10 * time.Second

// sqlCatalogStore persists static catalog files in MySQL or PostgreSQL.
type sqlCatalogStore struct {
	db        *sql.DB
	backend   string
	tableName string
}

func (s *sqlCatalogStore) init() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancel()
	if err := s.db.PingContext(ctx); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, s.schemaSQL())
	return err
}

func (s *sqlCatalogStore) Load() ([]FileRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancel()
	query := fmt.Sprintf("SELECT hash, name, size, file_type, extension, media_codec, media_length, media_bitrate, sources, complete_sources, endpoints_json FROM %s ORDER BY name ASC", s.tableName)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	files := make([]FileRecord, 0)
	for rows.Next() {
		var record FileRecord
		var hashText string
		var endpointsJSON string
		if err := rows.Scan(&hashText, &record.Name, &record.Size, &record.FileType, &record.Extension, &record.MediaCodec, &record.MediaLength, &record.MediaBitrate, &record.Sources, &record.CompleteSources, &endpointsJSON); err != nil {
			return nil, err
		}
		record.Hash, err = protocol.HashFromString(hashText)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(endpointsJSON) != "" {
			if err := json.Unmarshal([]byte(endpointsJSON), &record.Endpoints); err != nil {
				return nil, err
			}
		}
		files = append(files, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

func (s *sqlCatalogStore) Save(files []FileRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbOperationTimeout)
	defer cancel()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s", s.tableName)); err != nil {
		return err
	}
	insertSQL := s.insertSQL()
	for _, file := range files {
		endpointsJSON, marshalErr := json.Marshal(file.Endpoints)
		if marshalErr != nil {
			return marshalErr
		}
		if _, err = tx.ExecContext(ctx, insertSQL,
			file.Hash.String(),
			file.Name,
			file.Size,
			file.FileType,
			file.Extension,
			file.MediaCodec,
			file.MediaLength,
			file.MediaBitrate,
			file.Sources,
			file.CompleteSources,
			string(endpointsJSON),
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *sqlCatalogStore) Description() string {
	return fmt.Sprintf("%s:%s", s.backend, s.tableName)
}

func (s *sqlCatalogStore) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *sqlCatalogStore) schemaSQL() string {
	if s.backend == storageBackendMySQL {
		return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  hash CHAR(32) PRIMARY KEY,
  name VARCHAR(1024) NOT NULL,
  size BIGINT NOT NULL,
  file_type VARCHAR(128) NOT NULL DEFAULT '',
  extension VARCHAR(64) NOT NULL DEFAULT '',
  media_codec VARCHAR(128) NOT NULL DEFAULT '',
  media_length INT NOT NULL DEFAULT 0,
  media_bitrate INT NOT NULL DEFAULT 0,
  sources INT NOT NULL DEFAULT 0,
  complete_sources INT NOT NULL DEFAULT 0,
  endpoints_json JSON NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
)`, s.tableName)
	}
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
  hash CHAR(32) PRIMARY KEY,
  name VARCHAR(1024) NOT NULL,
  size BIGINT NOT NULL,
  file_type VARCHAR(128) NOT NULL DEFAULT '',
  extension VARCHAR(64) NOT NULL DEFAULT '',
  media_codec VARCHAR(128) NOT NULL DEFAULT '',
  media_length INTEGER NOT NULL DEFAULT 0,
  media_bitrate INTEGER NOT NULL DEFAULT 0,
  sources INTEGER NOT NULL DEFAULT 0,
  complete_sources INTEGER NOT NULL DEFAULT 0,
  endpoints_json JSONB NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`, s.tableName)
}

func (s *sqlCatalogStore) insertSQL() string {
	if s.backend == storageBackendPgSQL {
		return fmt.Sprintf("INSERT INTO %s (hash, name, size, file_type, extension, media_codec, media_length, media_bitrate, sources, complete_sources, endpoints_json) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)", s.tableName)
	}
	return fmt.Sprintf("INSERT INTO %s (hash, name, size, file_type, extension, media_codec, media_length, media_bitrate, sources, complete_sources, endpoints_json) VALUES (?,?,?,?,?,?,?,?,?,?,?)", s.tableName)
}
