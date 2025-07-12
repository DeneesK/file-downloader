package main

import (
	"context"

	"github.com/DeneesK/file-downloader/internal/app"
	"github.com/DeneesK/file-downloader/internal/app/conf"
	"github.com/DeneesK/file-downloader/internal/app/logger"
	"github.com/DeneesK/file-downloader/internal/app/services"
	"github.com/DeneesK/file-downloader/internal/app/storage/memorystorage"
)

func main() {
	config := conf.MustLoad()
	log := logger.NewLogger(config.Env)
	defer log.Sync()

	storage := memorystorage.NewMemoryStorage()

	ctx, close := context.WithCancel(context.Background())
	defer close()
	defer storage.Close(ctx) // в memory storage ctx не нужен, но на будущее если поменяем реализацию и заменим на ДБ

	zipService := services.NewZipService(config.ArchiveDir)
	taskService := services.NewTaskService(storage, log, config.MaxActiveTasks, config.MaxLinksPerTask, zipService)

	app := app.NewApp(config.ServerAddr, log, taskService)
	app.Run()
}
