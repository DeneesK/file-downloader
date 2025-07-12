package main

import (
	"context"

	"github.com/DeneesK/file-downloader/internal/app/conf"
	"github.com/DeneesK/file-downloader/internal/app/logger"
	"github.com/DeneesK/file-downloader/internal/app/repository"
	"github.com/DeneesK/file-downloader/internal/app/storage/memorystorage"
)

func main() {
	config := conf.MustLoad()
	log := logger.NewLogger(config.Env)
	defer log.Sync()

	storage := memorystorage.NewMemoryStorage()
	rep, err := repository.NewRepository(storage)
	if err != nil {
		log.Fatalf("failed to initialized repository: %s", err)
	}
	ctx, close := context.WithCancel(context.Background())
	defer close()
	defer rep.Close(ctx)

}
