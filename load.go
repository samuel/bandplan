package bandplan

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FileVersion is the version of the file format LoadFile reads and SaveFile
// writes. It changes only when an existing reader would misread a file.
const FileVersion = 1

// File is the top level of a band sets JSON file. bandplan.schema.json
// describes the format.
type File struct {
	Version int        `json:"version"`
	Sets    []*BandSet `json:"sets"`
}

// Decode parses and validates a band sets file.
func Decode(data []byte) ([]*BandSet, error) {
	var f File
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("parsing band sets file: %w", err)
	}
	if f.Version != FileVersion {
		return nil, fmt.Errorf("band sets file: version %d, want %d", f.Version, FileVersion)
	}
	for i, bs := range f.Sets {
		if bs == nil {
			return nil, fmt.Errorf("band sets file: set %d is null", i)
		}
		if err := bs.Validate(); err != nil {
			return nil, fmt.Errorf("band sets file: %w", err)
		}
	}
	return f.Sets, nil
}

// Encode returns sets as an indented band sets file ending in a newline.
func Encode(sets []*BandSet) ([]byte, error) {
	if sets == nil {
		sets = []*BandSet{}
	}
	data, err := json.MarshalIndent(File{Version: FileVersion, Sets: sets}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling band sets: %w", err)
	}
	return append(data, '\n'), nil
}

// LoadFile loads and validates BandSets from a JSON file. Returns an empty
// slice (not an error) if the file does not exist.
func LoadFile(path string) ([]*BandSet, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading band sets file: %w", err)
	}
	return Decode(data)
}

// SaveFile writes sets to path as JSON, creating parent directories as needed.
func SaveFile(path string, sets []*BandSet) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	data, err := Encode(sets)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing band sets file: %w", err)
	}
	return nil
}

// MergeSets returns a combined slice of builtins and custom sets. Custom sets
// with the same name as a builtin replace the builtin.
func MergeSets(builtins, custom []*BandSet) []*BandSet {
	customByName := make(map[string]bool, len(custom))
	for _, bs := range custom {
		customByName[bs.Name] = true
	}
	result := make([]*BandSet, 0, len(builtins)+len(custom))
	for _, bs := range builtins {
		if !customByName[bs.Name] {
			result = append(result, bs)
		}
	}
	return append(result, custom...)
}
