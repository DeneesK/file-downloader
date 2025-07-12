package main

import (
	"github.com/DeneesK/file-downloader/internal/app/conf"
	"github.com/DeneesK/file-downloader/internal/app/logger"
)

func main() {
	config := conf.MustLoad()
	log := logger.NewLogger(config.Env)
	defer log.Sync()
}
