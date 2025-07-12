package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/DeneesK/file-downloader/internal/app/model"
	"github.com/DeneesK/file-downloader/pkg/downloader"
	"github.com/google/uuid"
)

var ErrTooManyTasks = errors.New("server busy: too many active tasks")
var ErrNotValidExaction = errors.New("not valid exaction")

var allowedExtensions = map[string]struct{}{
	".pdf":  {},
	".jpeg": {},
	".jpg":  {},
}

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
	activeTasks int
	tasksLimit  int
	linksLimit  int
	taskQueue   chan string
	wg          sync.WaitGroup
	m           sync.RWMutex
	taskStore   TaskStorage
	zip         ZipService
	log         Logger
}

func NewTaskService(store TaskStorage, log Logger, tasksLimit, linksLimit int, zip ZipService) *taskService {
	return &taskService{
		activeTasks: 0,
		taskStore:   store,
		log:         log,
		tasksLimit:  tasksLimit,
		linksLimit:  linksLimit,
		zip:         zip,
		taskQueue:   make(chan string, tasksLimit),
	}
}

func (s *taskService) CreateTask(ctx context.Context) (string, error) {
	s.m.RLock()
	if s.activeTasks >= s.tasksLimit {
		return "", ErrTooManyTasks
	}
	defer s.m.RUnlock()
	s.incrementActiveTasks()

	id := uuid.NewString()

	task := &model.Task{ID: id, Status: model.StatusCreated, Links: make([]string, 0, 3), FailedLinks: make(map[string]string, 0)}
	err := s.taskStore.Store(ctx, task)
	if err != nil {
		return "", err
	}

	go func() {
		s.taskQueue <- task.ID
	}()

	return id, nil
}

func (s *taskService) AddLinks(ctx context.Context, taskID string, links []string) error {
	if !(s.isAllowedExtension(links)) {
		return ErrNotValidExaction
	}

	if len(links) > s.linksLimit {
		return fmt.Errorf("too many files, file limit per task is %d", s.linksLimit)
	}

	task, err := s.taskStore.Get(ctx, taskID)
	if err != nil {
		return err
	}

	if len(task.Links) >= s.linksLimit {
		return fmt.Errorf("too many files in task with ID %s", taskID)
	}

	task.Links = append(task.Links, links...)

	return s.taskStore.Update(ctx, task)
}

func (s *taskService) GetTask(ctx context.Context, taskID string) (*model.Task, error) {
	task, err := s.taskStore.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *taskService) Start() {
	s.log.Infoln("task service started")
	for task := range s.taskQueue {
		go s.processTask(task)
	}
}

func (s *taskService) Shutdown() {
	s.log.Infoln("task service try to gracefully shutdown")
	close(s.taskQueue)
	s.wg.Wait()
	s.log.Infoln("task service gracefully shutdown")
}

func (s *taskService) GetNumberActiveTasks() int {
	return len(s.taskQueue)
}

func (s *taskService) incrementActiveTasks() {
	s.m.Lock()
	s.activeTasks++
	s.m.Unlock()
}

func (s *taskService) decrementActiveTasks() {
	s.m.Lock()
	s.activeTasks--
	s.m.Unlock()
}

func (s *taskService) setStatus(ctx context.Context, taskID string, status string) error {
	task, err := s.taskStore.Get(ctx, taskID)
	if err != nil {
		return err
	}
	task.Status = status

	return s.taskStore.Update(ctx, task)
}

func (s *taskService) processTask(taskID string) {
	defer s.decrementActiveTasks()
	s.log.Infoln("started process of task ID ", taskID)
	s.setStatus(context.Background(), taskID, model.StatusRunning)

	for {
		task, err := s.taskStore.Get(context.Background(), taskID)
		if err != nil {
			s.log.Errorf("during process of task ID ", taskID)
		}

		if len(task.FailedLinks) == s.linksLimit {
			task.Status = model.StatusFailed
			s.taskStore.Update(context.Background(), task)
			return
		} else if len(task.FailedLinks)+len(task.DownloadedFiles) == s.linksLimit {
			task.Status = model.StatusDone
			s.taskStore.Update(context.Background(), task)
			return
		}

		for _, l := range task.Links {
			_, ok := task.FailedLinks[l]
			if !ok {
				filepath, err := downloader.DownloadFile(l)
				if err != nil {
					s.log.Errorf("during process task ID %s failed to download file %s", taskID, err)
					task.FailedLinks[l] = fmt.Sprintf("%s", err)
					s.taskStore.Update(context.Background(), task)
				}
				task.DownloadedFiles = append(task.DownloadedFiles, filepath)
			}

		}
	}
}

func (s *taskService) isAllowedExtension(links []string) bool {
	for _, l := range links {
		u, err := url.Parse(l)
		if err != nil {
			return false
		}
		ext := strings.ToLower(filepath.Ext(u.Path))
		_, ok := allowedExtensions[ext]
		if !ok {
			return false
		}
	}

	return true
}
