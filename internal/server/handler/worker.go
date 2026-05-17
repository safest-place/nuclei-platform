package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/projectdiscovery/nuclei-platform/internal/model"
	"gorm.io/gorm"
)

type WorkerHandler struct {
	db *gorm.DB
}

func NewWorkerHandler(db *gorm.DB) *WorkerHandler {
	return &WorkerHandler{db: db}
}

func (h *WorkerHandler) List(w http.ResponseWriter, r *http.Request) {
	var workers []model.Worker
	h.db.Order("last_heartbeat DESC").Find(&workers)
	Success(w, workers)
}

func (h *WorkerHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var worker model.Worker
	if err := h.db.First(&worker, "id = ?", id).Error; err != nil {
		Error(w, http.StatusNotFound, "worker not found")
		return
	}
	Success(w, worker)
}

func (h *WorkerHandler) Disable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var worker model.Worker
	if err := h.db.First(&worker, "id = ?", id).Error; err != nil {
		Error(w, http.StatusNotFound, "worker not found")
		return
	}
	h.db.Model(&worker).Update("disabled", true)
	worker.Disabled = true
	Success(w, worker)
}

func (h *WorkerHandler) Enable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var worker model.Worker
	if err := h.db.First(&worker, "id = ?", id).Error; err != nil {
		Error(w, http.StatusNotFound, "worker not found")
		return
	}
	h.db.Model(&worker).Update("disabled", false)
	worker.Disabled = false
	Success(w, worker)
}
