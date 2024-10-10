package storage

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Storer represents the interface for a file storing service that is used by the Keel runtime
// TODO: all these methods should take context as first arg
type Storer interface {
	// Store will save the given file and return a FileInfo struct for it
	//
	// The input should be a well formed dataURL https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URLs
	// The name of the file can also be passed as a parameter of the mediaType segment; e.g.
	// data:application/pdf;name=MyUploadedFile.pdf;base64,xxxxxx[...]
	Store(dataURL string) (FileInfo, error)

	// GetFileInfo will return the file information for the given unique file key as stored in the database.
	GetFileInfo(key string) (FileInfo, error)

	// GenerateFileResponse will take the given file info and generate a response to be returned from an API.
	//
	// The use of this function is to generate any signed URLs for file downloads.
	GenerateFileResponse(fi *FileInfo) (FileResponse, error)
}

// FileInfo contains important data for the File type as stored in the database
type FileInfo struct {
	Key         string `json:"key"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
}

func (t FileInfo) Value() (driver.Value, error) {
	json, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("marshalling to json: %w", err)
	}

	return string(json), nil
}

// FileResponse is what is returned from our APIs
type FileResponse struct {
	Key         string `json:"key"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
}

func (fi *FileInfo) ToJSON() (string, error) {
	json, err := json.Marshal(fi)
	if err != nil {
		return "", fmt.Errorf("marshalling to json: %w", err)
	}
	return string(json), nil
}

func (fi *FileInfo) ToDbRecord() (string, error) {
	json, err := json.Marshal(fi)
	if err != nil {
		return "", fmt.Errorf("marshalling to json: %w", err)
	}

	return string(json), nil
}
