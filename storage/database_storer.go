package storage

import (
	"context"
	"fmt"

	"github.com/teamkeel/keel/db"
	"github.com/vincent-petithory/dataurl"
)

type DbStore struct {
	db db.Database
}

// Make sure DbStore implements the storage.Storer interface. Any missing methods will cause a compile error
var _ Storer = &DbStore{}

const dbTable = "keel_storage" // table where files will be stored

type fileData struct {
	Filename    string
	ContentType string
	Data        []byte
}

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

// decodeDataURL will take a dataURL and return it as a FileData struct
func decodeDataURL(dataURL string) (fileData, error) {
	durl, err := dataurl.DecodeString(dataURL)
	if err != nil {
		return fileData{}, fmt.Errorf("decoding data url: %w", err)
	}

	return fileData{
		ContentType: durl.ContentType(),
		Filename:    durl.Params["name"],
		Data:        durl.Data,
	}, nil
}

func encodeDataURL(data fileData) string {
	durl := dataurl.New(data.Data, data.ContentType, "name", data.Filename)
	return durl.String()
}

func (s *DbStore) Store(url string) (FileInfo, error) {
	fd, err := decodeDataURL(url)
	if err != nil {
		return FileInfo{}, fmt.Errorf("decoding data URL: %w", err)
	}

	sql := `INSERT INTO ` + dbTable + ` (filename, content_type, data) VALUES (?, ?, ?)  
	 	RETURNING 
			id AS key, 
			filename,
			content_type,
			octet_length(data) as size`

	var fi FileInfo
	db := s.db.GetDB().Raw(sql, fd.Filename, fd.ContentType, fd.Data).Scan(&fi)
	if db.Error != nil {
		return FileInfo{}, fmt.Errorf("saving file in db: %w", db.Error)
	}

	return fi, nil
}

func (s *DbStore) GetFileInfo(key string) (FileInfo, error) {
	sql := `SELECT
			id AS KEY,
			filename,
			content_type,
			octet_length(data) AS size
		FROM ` + dbTable + ` WHERE id = ?`

	var fi FileInfo

	db := s.db.GetDB().Raw(sql, key).Scan(&fi)
	if db.Error != nil {
		return FileInfo{}, fmt.Errorf("retrieving file info: %w", db.Error)
	}

	return fi, nil
}

func (s *DbStore) HydrateFileInfo(fi *FileInfo) (FileInfo, error) {
	sql := `SELECT
			filename,
			content_type,
			data
		FROM ` + dbTable + ` WHERE id = ?`

	var fd fileData

	db := s.db.GetDB().Raw(sql, fi.Key).Scan(&fd)
	if db.Error != nil {
		return FileInfo{}, fmt.Errorf("retrieving file data: %w", db.Error)
	}
	dataURL := encodeDataURL(fd)

	return FileInfo{
		Key:         fi.Key,
		Filename:    fi.Filename,
		ContentType: fi.ContentType,
		Size:        fi.Size,
		URL:         &dataURL,
	}, nil
}
