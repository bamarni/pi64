package pi64

import (
	"encoding/json"
	"os"
)

var metadataPath = "/boot/pi64.json"

type Metadata struct {
	Version       string `json:"version"`
	KernelVersion string `json:"kernel-version"`
}

func ReadMetadata() (Metadata, error) {
	metadata := Metadata{}
	file, err := os.Open(metadataPath)
	if err != nil {
		return metadata, err
	}
	err = json.NewDecoder(file).Decode(&metadata)
	file.Close()
	return metadata, err
}

func WriteMetadata(metadata Metadata) error {
	file, err := os.OpenFile(metadataPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 644)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(metadata)
}
