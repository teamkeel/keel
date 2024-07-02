package storage

import (
	"encoding/binary"
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
	Store(dataURL string) (*FileData, error)
	GetFile() error
}

type FileData struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Data        []byte `json:"-"`
	Size        int    `json:"size"`
}

func (fd *FileData) ToJSON() (string, error) {
	json, err := json.Marshal(fd)
	if err != nil {
		return "", fmt.Errorf("marshalling to json: %w", err)
	}
	return string(json), nil
}

// DecodeDataURL will take a dataURL and return it as a FileData struct
func DecodeDataURL(dataURL string) (FileData, error) {
	durl, err := dataurl.DecodeString(dataURL)
	if err != nil {
		return FileData{}, fmt.Errorf("decoding data url: %w", err)
	}

	return FileData{
		ContentType: durl.ContentType(),
		Filename:    durl.Params["name"],
		Data:        durl.Data,
		Size:        binary.Size(durl.Data),
	}, nil
}
