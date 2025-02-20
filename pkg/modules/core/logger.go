package core

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type ModuleLogger struct {
	Outputs map[string]*os.File
	mu      sync.RWMutex
}

func (ml *ModuleLogger) Log(module string, level string, msg string) error {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	output := ml.Outputs[module]
	_, err := fmt.Fprintf(output, "[%s] %s: %s\n", level, time.Now(), msg)
	return err
}
