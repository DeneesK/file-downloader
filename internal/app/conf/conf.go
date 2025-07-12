package conf

import (
	"flag"
	"os"
)

type ServerConf struct {
	MaxLinksPerTask int
	MaxActiveTasks  int
	ServerAddr      string
	Env             string
	ArchiveDir      string
}

var cfg ServerConf

func init() {
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.Env, "env", "dev", "environment 'dev' or 'prod'")
	flag.StringVar(&cfg.Env, "dir", "static/archives", "dir for created archives")
	flag.IntVar(&cfg.MaxActiveTasks, "tasks", 3, "limit of tasks per user")
	flag.IntVar(&cfg.MaxLinksPerTask, "links", 3, "limit of links per task")
}

func MustLoad() *ServerConf {
	flag.Parse()

	if serverAddr, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddr = serverAddr
	}

	if env, ok := os.LookupEnv("Env"); ok {
		cfg.Env = env
	}

	return &cfg
}
