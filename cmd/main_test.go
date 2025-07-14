package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DeneesK/file-downloader/internal/app/model"
	"github.com/DeneesK/file-downloader/internal/app/router"
	"github.com/DeneesK/file-downloader/internal/app/services"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const (
	jpgLink = `https://upload.wikimedia.org/wikipedia/commons/thumb/5/52/Cat_November_2010-1a_%28cropped_2023%29.jpg/250px-Cat_November_2010-1a_%28cropped_2023%29.jpg`
)

type mockStorage struct {
	tasks map[string]*model.Task
}

func newMockStorage() *mockStorage {
	return &mockStorage{tasks: make(map[string]*model.Task)}
}

func (m *mockStorage) Store(_ context.Context, task *model.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *mockStorage) Get(_ context.Context, id string) (model.Task, error) {
	task, ok := m.tasks[id]
	if !ok {
		return model.Task{}, errors.New("not found")
	}
	return *task, nil
}

func (m *mockStorage) Update(_ context.Context, task model.Task) error {
	m.tasks[task.ID] = &task
	return nil
}

func (m *mockStorage) Close(_ context.Context) error { return nil }
func (m *mockStorage) Ping(_ context.Context) error  { return nil }

type mockZip struct{}

func (z *mockZip) CreateZipArchive(files []string) (string, error) {
	return "/fake/path.zip", nil
}

func TestServiceCreateTask(t *testing.T) {
	store := newMockStorage()
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer log.Sync()
	service := services.NewTaskService(store, log, 3, 3, &mockZip{})

	ctx := context.Background()
	id, err := service.CreateTask(ctx)

	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	task, err := store.Get(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, model.StatusCreated, task.Status)
}

func TestServiceAddLinks_Valid(t *testing.T) {
	store := newMockStorage()
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer log.Sync()
	service := services.NewTaskService(store, log, 3, 3, &mockZip{})
	ctx := context.Background()

	id, _ := service.CreateTask(ctx)
	links := []string{
		"https://example.com/a.pdf",
		"https://example.com/b.jpeg",
	}

	err := service.AddLinks(ctx, id, links)
	assert.NoError(t, err)

	task, _ := store.Get(ctx, id)
	assert.Equal(t, 2, len(task.Links))
}

func TestServiceAddLinks_InvalidExtension(t *testing.T) {
	store := newMockStorage()
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer log.Sync()
	service := services.NewTaskService(store, log, 3, 3, &mockZip{})
	ctx := context.Background()

	id, _ := service.CreateTask(ctx)
	links := []string{"https://example.com/malware.exe"}

	err := service.AddLinks(ctx, id, links)
	assert.Equal(t, services.ErrNotValidExaction, err)
}

func TestHandlerCreateTask_Success(t *testing.T) {
	store := newMockStorage()
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer log.Sync()
	service := services.NewTaskService(store, log, 3, 3, &mockZip{})
	req := httptest.NewRequest(http.MethodPost, "/task", nil)
	w := httptest.NewRecorder()
	handler := router.CreateTask(service, log)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var task model.Task
	err := json.NewDecoder(w.Body).Decode(&task)
	assert.NoError(t, err)

	assert.NotEmpty(t, task.ID)
	assert.Equal(t, model.StatusCreated, task.Status)
}

func TestHandlerCreateTask_TooMany(t *testing.T) {
	store := newMockStorage()
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer log.Sync()
	service := services.NewTaskService(store, log, 0, 3, &mockZip{})
	req := httptest.NewRequest(http.MethodPost, "/task", nil)
	w := httptest.NewRecorder()

	handler := router.CreateTask(service, log)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestHandlerAddLinks_Success(t *testing.T) {
	store := newMockStorage()
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer log.Sync()
	service := services.NewTaskService(store, log, 3, 3, &mockZip{})
	req := httptest.NewRequest(http.MethodPost, "/task", nil)
	w := httptest.NewRecorder()

	handler := router.CreateTask(service, log)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var task model.Task
	err := json.NewDecoder(w.Body).Decode(&task)
	assert.NoError(t, err)

	body := fmt.Sprintf(`{"links": ["%s"]}`, jpgLink)
	reqURL := fmt.Sprintf("/task/%s", task.ID)
	r := httptest.NewRequest(http.MethodPatch, reqURL, strings.NewReader(body))
	chi.URLParam(req, "id")
	req = r.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{task.ID},
		},
	}))

	w = httptest.NewRecorder()

	handler = router.AddLinks(service, log)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerAddLinks_InvalidType(t *testing.T) {
	store := newMockStorage()
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer log.Sync()
	service := services.NewTaskService(store, log, 3, 3, &mockZip{})
	req := httptest.NewRequest(http.MethodPost, "/task", nil)
	w := httptest.NewRecorder()

	handler := router.CreateTask(service, log)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var task model.Task
	err := json.NewDecoder(w.Body).Decode(&task)
	assert.NoError(t, err)

	body := fmt.Sprintf(`{"links": ["%s"]}`, "https://example.com/bad.exe")
	reqURL := fmt.Sprintf("/task/%s", task.ID)
	r := httptest.NewRequest(http.MethodPatch, reqURL, strings.NewReader(body))
	req = r.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
	chi.URLParam(req, "id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{task.ID},
		},
	}))

	w = httptest.NewRecorder()

	handler = router.AddLinks(service, log)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTask_Success(t *testing.T) {
	store := newMockStorage()
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer log.Sync()
	service := services.NewTaskService(store, log, 3, 3, &mockZip{})
	req := httptest.NewRequest(http.MethodPost, "/task", nil)
	w := httptest.NewRecorder()

	handler := router.CreateTask(service, log)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var task model.Task
	err := json.NewDecoder(w.Body).Decode(&task)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	reqURL := fmt.Sprintf("/task/%s", task.ID)
	fmt.Println(task)
	fmt.Println(reqURL)
	r := httptest.NewRequest(http.MethodGet, reqURL, nil)
	req = r.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{"id"},
			Values: []string{task.ID},
		},
	}))
	handler = router.GetTask(service, log)
	handler.ServeHTTP(w, req)

	responseBody := w.Body.String()
	fmt.Println(responseBody)
	assert.Equal(t, http.StatusOK, w.Code)

	var taskReturned model.Task
	err = json.NewDecoder(w.Body).Decode(&taskReturned)
	assert.NoError(t, err)
	assert.Equal(t, taskReturned.ID, task.ID)
}
