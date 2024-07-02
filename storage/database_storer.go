package storage

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/db"
)

type DbStore struct {
	db db.Database
}

// Make sure DbStore implements the storage.Storer interface. Any missing methods will cause a compile error
var _ Storer = &DbStore{}

const dbTable = "keel_storage" // table where files will be stored

// NewDbStore will return a Storage service for files that is db based
func NewDbStore(ctx context.Context, db db.Database) (*DbStore, error) {
	svc := &DbStore{
		db: db,
	}
	if err := svc.setupDB(ctx); err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *DbStore) setupDB(ctx context.Context) error {
	if _, err := s.db.ExecuteStatement(ctx, `
	CREATE TABLE IF NOT EXISTS `+dbTable+` (
		"id" text NOT NULL DEFAULT ksuid(),
		"filename" text NOT NULL,
		"content_type" text NOT NULL,
		"data" bytea NOT NULL,
		"created_at" timestamptz NOT NULL DEFAULT now(),
		PRIMARY KEY ("id")
	);`); err != nil {
		return fmt.Errorf("failed to initialise DB file storage: %w", err)
	}

	return nil
}

func (s *DbStore) Store(url string) (*FileData, error) {
	fd, err := DecodeDataURL(url)
	if err != nil {
		return nil, fmt.Errorf("decoding data URL: %w", err)
	}

	sql := `INSERT INTO ` + dbTable + ` (filename, content_type, data) VALUES (?, ?, ?)`

	db := s.db.GetDB().Exec(sql, fd.Filename, fd.ContentType, fd.Data)
	if db.Error != nil {
		return nil, fmt.Errorf("saving file in db: %w", db.Error)
	}

	return &fd, nil
}

func (s *DbStore) GetFile() error {
	return nil
}
