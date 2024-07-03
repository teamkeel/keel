package storage

import (
	"encoding/json"
	"fmt"

	"github.com/vincent-petithory/dataurl"
)

// Storer represents the interface for a file storing service that is used by the Keel runtime
type Storer interface {
	// Store will save the given file and return a FileData struct for it
	//
	// The input should be a well formed dataURL https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URLs
	// The name of the file can also be passed as a parameter of the mediaType segment; e.g.
	// data:application/pdf;name=MyUploadedFile.pdf;base64,xxxxxx[...]
	Store(dataURL string) (FileInfo, error)

	// GetFileInfo will return the file information for the given unique file key.
	//
	// The File info returned can contain a URL where the file can be downloaded from, if applicable; i.e. for database
	// storage, at the moment files cannot be retrieved via URLs.
	GetFileInfo(key string) (FileInfo, error)

	// HydrateFileInfo will take the given file info and hydrate it with the most up to date information.
	//
	// The use of this function is to generate any signed URLs for file downlaods
	HydrateFileInfo(fi *FileInfo) (FileInfo, error)
}

type FileInfo struct {
	Key         string  `json:"key"`
	Filename    string  `json:"filename"`
	ContentType string  `json:"contentType"`
	Size        int     `json:"size"`
	URL         *string `json:"url,omitempty"`
	Public      bool    `json:"public"`
}
type fileData struct {
	Filename    string
	ContentType string
	Data        []byte
}

func (fi *FileInfo) ToJSON() (string, error) {
	json, err := json.Marshal(fi)
	if err != nil {
		return "", fmt.Errorf("marshalling to json: %w", err)
	}
	return string(json), nil
}

// DecodeDataURL will take a dataURL and return it as a FileData struct
func DecodeDataURL(dataURL string) (fileData, error) {
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
