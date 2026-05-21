package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aric/garden/internal/memory"
)

type Store interface {
	Init() error
	Load() (memory.Document, error)
	Save(memory.Document) error
}

type JSONStore struct {
	root string
}

func NewJSONStore(root string) *JSONStore {
	if root == "" {
		root = "."
	}
	return &JSONStore{root: root}
}

func (s *JSONStore) Init() error {
	if err := os.MkdirAll(s.gardenDir(), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(s.path()); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return s.Save(memory.NewDocument())
}

func (s *JSONStore) Load() (memory.Document, error) {
	data, err := os.ReadFile(s.path())
	if errors.Is(err, os.ErrNotExist) {
		return memory.Document{}, fmt.Errorf("garden is not initialized; run garden init")
	}
	if err != nil {
		return memory.Document{}, err
	}

	var doc memory.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return memory.Document{}, fmt.Errorf("read memories: %w", err)
	}
	if doc.Version != memory.Version {
		return memory.Document{}, fmt.Errorf("unsupported memories version %d", doc.Version)
	}
	return memory.NormalizeDocument(doc), nil
}

func (s *JSONStore) Save(doc memory.Document) error {
	doc = memory.NormalizeDocument(doc)
	if err := os.MkdirAll(s.gardenDir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(s.gardenDir(), "memories-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.path())
}

func (s *JSONStore) gardenDir() string {
	return filepath.Join(s.root, ".garden")
}

func (s *JSONStore) path() string {
	return filepath.Join(s.gardenDir(), "memories.json")
}
