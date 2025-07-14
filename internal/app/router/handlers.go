package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/DeneesK/file-downloader/internal/app/model"
	"github.com/DeneesK/file-downloader/internal/app/services"
	"github.com/DeneesK/file-downloader/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

type Links struct {
	Links []string `json:"links"`
}

func CreateTask(taskService TaskService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, err := taskService.CreateTask(ctx)
		if err == services.ErrTooManyTasks {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		task := model.Task{ID: id, Status: model.StatusCreated}

		err = json.NewEncoder(w).Encode(task)
		if err != nil {
			errorString := fmt.Sprintf("failed to encode task to json: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}
	}
}

func AddLinks(taskService TaskService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")

		links := Links{}

		if err := json.NewDecoder(r.Body).Decode(&links); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		err := taskService.AddLinks(ctx, id, links.Links)
		if errors.Is(err, services.ErrNotValidExaction) || errors.Is(err, services.ErrTooManyFiles) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err == storage.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetTask(taskService TaskService, log Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "id")

		task, err := taskService.GetTask(ctx, id)
		if err == storage.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(task)
		if err != nil {
			errorString := fmt.Sprintf("failed to encode task to json: %s", err.Error())
			log.Error(errorString)
			http.Error(w, errorString, http.StatusBadRequest)
			return
		}
	}
}
