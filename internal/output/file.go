package output

import (
	"os"
)

// FileOutput writes output to a file
type FileOutput struct {
	path string
}

// NewFileOutput creates a new file output
func NewFileOutput(path string) *FileOutput {
	return &FileOutput{path: path}
}

// Write writes data to the file
func (f *FileOutput) Write(data []byte) error {
	return os.WriteFile(f.path, data, 0644)
}
