package services

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/DeneesK/file-downloader/internal/app/model"
	"github.com/DeneesK/file-downloader/pkg/downloader"
	"github.com/google/uuid"
)

var ErrTooManyTasks = errors.New("server busy: too many active tasks")

type TaskStorage interface {
	Store(ctx context.Context, task *model.Task) error
	Get(ctx context.Context, id string) (model.Task, error)
	Update(ctx context.Context, task model.Task) error
	Close(ctx context.Context) error
	Ping(ctx context.Context) error
}

type ZipService interface {
	createZipArchive(files []string) (string, error)
}

type Logger interface {
	Infoln(args ...interface{})
	Fatalf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Error(args ...interface{})
}

type taskService struct {
	active     int
	tasksLimit int
	linksLimit int
	m          sync.RWMutex
	taskStore  TaskStorage
	zip        ZipService
	log        Logger
}

func NewTaskService(store TaskStorage, log Logger, tasksLimit, linksLimit int, zip ZipService) *taskService {
	return &taskService{
		active:     0,
		taskStore:  store,
		log:        log,
		tasksLimit: tasksLimit,
		linksLimit: linksLimit,
		zip:        zip,
	}
}

func (s *taskService) CreateTask(ctx context.Context) (string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.active >= s.tasksLimit {
		return "", ErrTooManyTasks
	}

	id := uuid.NewString()

	task := &model.Task{ID: id, Status: model.StatusCreated, Links: make([]string, 0, 3), FailedLinks: make(map[string]string, 0)}
	err := s.taskStore.Store(ctx, task)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *taskService) AddLinks(ctx context.Context, taskID string, links []string) error {
	if len(links) > s.linksLimit {
		return fmt.Errorf("too many links, links limit per task is %d", s.linksLimit)
	}

	task, err := s.taskStore.Get(ctx, taskID)
	if err != nil {
		return err
	}

	if len(task.Links) >= s.linksLimit {
		return fmt.Errorf("too many links in task with ID %s", taskID)
	}

	task.Links = append(task.Links, links...)

	s.taskStore.Update(ctx, task)

	if len(task.Links) == s.linksLimit {
		go s.processTask(ctx, taskID)
	}

	return nil
}

func (s *taskService) SetStatus(ctx context.Context, taskID string, status string) error {
	task, err := s.taskStore.Get(ctx, taskID)
	if err != nil {
		return err
	}
	task.Status = status

	return s.taskStore.Update(ctx, task)
}

func (s *taskService) GetStatus(ctx context.Context, taskID string) (string, error) {
	task, err := s.taskStore.Get(ctx, taskID)
	if err != nil {
		return "", err
	}
	return task.Status, nil
}

func (s *taskService) GetNumberActiveTasks() int {
	s.m.RLock()
	n := s.active
	s.m.Unlock()
	return n
}

func (s *taskService) incrementActiveTasks() {
	s.m.Lock()
	s.active++
	s.m.Unlock()
}

func (s *taskService) decrementActiveTasks() {
	s.m.Lock()
	s.active--
	s.m.Unlock()
}

func (s *taskService) processTask(ctx context.Context, taskID string) {
	s.incrementActiveTasks()
	defer s.decrementActiveTasks()

	task, err := s.taskStore.Get(ctx, taskID)
	if err != nil {
		s.log.Errorf("failed to find task with ID: %s", err)
		return
	}

	if len(task.Links) < s.linksLimit {
		return
	}

	s.log.Infoln("started task with ID ", taskID)

	task.Status = model.StatusRunning
	s.taskStore.Update(ctx, task)

	downloadedFiles := make([]string, 0, 3)
	for _, link := range task.Links {
		path, err := downloader.DownloadFile(link)
		if err != nil {
			e := fmt.Sprintf("failed to download %v", err)
			task.FailedLinks[link] = e
			s.taskStore.Update(ctx, task)
			continue
		}
		downloadedFiles = append(downloadedFiles, path)
	}

	if len(downloadedFiles) == 0 {
		task.Status = model.StatusFailed
		s.taskStore.Update(ctx, task)
		return
	}

	archivePath, err := s.zip.createZipArchive(downloadedFiles)
	if err != nil {
		task.Status = model.StatusFailed
		s.taskStore.Update(ctx, task)
		return
	}

	task.Status = model.StatusDone
	task.Archive = archivePath
	s.taskStore.Update(ctx, task)
}
