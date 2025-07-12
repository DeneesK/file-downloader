package repository

import (
	"context"

	"github.com/DeneesK/file-downloader/internal/app/model"
)

type TaskStorage interface {
	Store(ctx context.Context, id string, task *model.Task) error
	Get(context.Context, string) (*model.Task, error)
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
}

type Repository struct {
	storage TaskStorage
}

func NewRepository(storage TaskStorage) (*Repository, error) {
	rep := &Repository{
		storage: storage,
	}
	return rep, nil
}

func (rep *Repository) SaveTask(ctx context.Context, id string, task *model.Task) (string, error) {
	return "", nil
}

func (rep *Repository) GetTask(ctx context.Context, id string) (*model.Task, error) {
	return rep.storage.Get(ctx, id)
}

func (rep *Repository) PingStorage(ctx context.Context) error {
	return rep.storage.Ping(ctx)
}

func (rep *Repository) Close(ctx context.Context) error {
	return rep.storage.Close(ctx)
}
