package core

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
)

type HotReloader struct {
	watcher  *fsnotify.Watcher
	registry *ModuleRegistry
	logger   *log.Logger
}

func NewHotReloader(registry *ModuleRegistry, logger *log.Logger) (*HotReloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	hr := &HotReloader{
		watcher:  watcher,
		registry: registry,
		logger:   logger,
	}

	go hr.watchLoop()
	return hr, nil
}

func (h *HotReloader) handleChange(event fsnotify.Event) error {
	if event.Op != fsnotify.Write {
		return nil
	}

	// Get module name from file path
	moduleName := filepath.Base(filepath.Dir(event.Name))

	// Load updated module
	newModule, err := h.registry.Loader.Load(event.Name)
	if err != nil {
		return fmt.Errorf("failed to load updated module: %w", err)
	}

	// Stop old module
	if oldModule, exists := h.registry.Get(moduleName); exists {
		if err := oldModule.Terminate(); err != nil {
			return fmt.Errorf("failed to terminate old module: %w", err)
		}
	}

	// Initialize and register new module
	if err := newModule.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize new module: %w", err)
	}

	h.registry.mu.Lock()
	h.registry.modules[moduleName] = newModule
	h.registry.mu.Unlock()

	h.logger.Printf("Module %s reloaded successfully", moduleName)
	return nil
}

func (h *HotReloader) watchLoop() {
	for {
		select {
		case event := <-h.watcher.Events:
			if err := h.handleChange(event); err != nil {
				h.logger.Printf("Hot reload error: %v", err)
			}
		case err := <-h.watcher.Errors:
			h.logger.Printf("Watcher error: %v", err)
		}
	}
}

func (h *HotReloader) WatchRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return h.watcher.Add(path)
		}
		return nil
	})
}
