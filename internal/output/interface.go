package output

import (
	"fmt"

	"github.com/melee-elo-ranking/internal/config"
)

// Output defines the interface for different output methods
type Output interface {
	Write(data []byte) error
}

// New creates an output based on the configuration
func New(outputType string, cfg *config.Config) (Output, error) {
	switch outputType {
	case "file":
		return NewFileOutput(cfg.Paths.Output), nil
	default:
		return nil, fmt.Errorf("unknown output type: %s", outputType)
	}
}
