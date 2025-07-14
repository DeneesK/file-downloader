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
var ErrTooManyFiles = errors.New("too many files per task")

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
	CreateZipArchive(files []string) (string, error)
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
	s.m.Lock()
	if s.activeTasks >= s.tasksLimit {
		return "", ErrTooManyTasks
	}
	s.activeTasks++
	s.m.Unlock()

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
	if len(links) > s.linksLimit {
		return ErrTooManyFiles
	}
	if !(s.isAllowedExtension(links)) {
		return ErrNotValidExaction
	}

	task, err := s.taskStore.Get(ctx, taskID)
	if err != nil {
		return err
	}

	if task.LinksNumber >= s.linksLimit {
		return ErrTooManyFiles
	}

	task.Links = append(task.Links, links...)
	task.LinksNumber++

	return s.taskStore.Update(ctx, task)
}

func (s *taskService) GetTask(ctx context.Context, taskID string) (*model.Task, error) {
	task, err := s.taskStore.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *taskService) Start(ctx context.Context) {
	s.log.Infoln("task service started")

	for {
		select {
		case <-ctx.Done():
			s.log.Infoln("task service try to gracefully shutdown")

			s.wg.Wait()

			s.log.Infoln("task service gracefully shutdown")
			return
		case task := <-s.taskQueue:
			s.wg.Add(1)
			s.processTask(ctx, task)
		}
	}
}

func (s *taskService) GetNumberActiveTasks() int {
	s.m.RLock()
	c := s.activeTasks
	s.m.RUnlock()
	return c
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

func (s *taskService) processTask(ctx context.Context, taskID string) {
	defer s.decrementActiveTasks()
	defer s.wg.Done()
	s.log.Infoln("started process of task ID ", taskID)
	s.setStatus(ctx, taskID, model.StatusRunning)

	for {
		select {
		case <-ctx.Done():
			s.log.Infoln("task canceled:", taskID)
			return
		default:
			task, err := s.taskStore.Get(ctx, taskID)
			if err != nil {
				s.log.Errorf("during process of task ID %s error %v", taskID, err)
			}

			if len(task.FailedLinks) == s.linksLimit {
				task.Status = model.StatusFailed
				s.taskStore.Update(ctx, task)
				return
			} else if len(task.FailedLinks)+len(task.DownloadedFiles) == s.linksLimit {
				archive, err := s.zip.CreateZipArchive(task.DownloadedFiles)
				if err != nil {
					s.log.Errorf("during process of task ID %s error %v", taskID, err)
					task.Status = model.StatusFailed
					s.taskStore.Update(ctx, task)
					return
				}
				task.Archive = archive
				task.Status = model.StatusDone
				s.taskStore.Update(ctx, task)
				return
			}

			for _, l := range task.Links {
				filepath, err := downloader.DownloadFile(l)
				if err != nil {
					s.log.Errorf("during process task ID %s failed to download file %v", taskID, err)
					task.FailedLinks[l] = fmt.Sprintf("%s", err)
					s.taskStore.Update(context.Background(), task)
				}
				task.DownloadedFiles = append(task.DownloadedFiles, filepath)
				task.Links = task.Links[:0]
				err = s.taskStore.Update(context.Background(), task)
				if err != nil {
					s.log.Errorf("during process task ID %s error %s", taskID, err)
					s.setStatus(context.Background(), taskID, model.StatusFailed)
					return
				}
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
