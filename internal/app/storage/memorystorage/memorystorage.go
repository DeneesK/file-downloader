package memorystorage

import (
	"context"
	"sync"

	"github.com/DeneesK/file-downloader/internal/app/model"
	"github.com/DeneesK/file-downloader/internal/app/storage"
)

type MemoryStorage struct {
	m       sync.RWMutex
	storage map[string]*model.Task
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		storage: make(map[string]*model.Task),
	}
}

func (s *MemoryStorage) Store(ctx context.Context, value *model.Task) error {
	s.m.Lock()
	defer s.m.Unlock()
	if s.isExists(value.ID) {
		return storage.ErrNotUniqueVallation
	}
	s.storage[value.ID] = value
	return nil
}

func (s *MemoryStorage) Get(ctx context.Context, id string) (model.Task, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	if !(s.isExists(id)) {
		return model.Task{}, storage.ErrNotFound
	}
	task := *s.storage[id]
	return task, nil
}

func (s *MemoryStorage) Update(ctx context.Context, task model.Task) error {
	s.m.Lock()
	defer s.m.Unlock()
	s.storage[task.ID] = &task
	return nil
}

func (s *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) Close(ctx context.Context) error {
	return nil
}

func (s *MemoryStorage) isExists(id string) bool {
	_, ok := s.storage[id]
	return ok
}
