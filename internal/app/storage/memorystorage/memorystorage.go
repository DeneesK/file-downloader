package memorystorage

import (
	"context"
	"sync"

	"github.com/DeneesK/file-downloader/internal/app/model"
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

func (s *MemoryStorage) Store(ctx context.Context, id string, value *model.Task) error {
	s.m.Lock()
	defer s.m.Unlock()
	return nil
}

func (s *MemoryStorage) Get(ctx context.Context, id string) (*model.Task, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	return nil, nil
}

func (s *MemoryStorage) Update(ctx context.Context, task *model.Task) error {
	s.m.Lock()
	defer s.m.Unlock()
	s.storage[task.ID] = task
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
