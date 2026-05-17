package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/projectdiscovery/nuclei-platform/internal/model"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type ScanHandler struct {
	db  *gorm.DB
	pub TaskPublisher
	log *zerolog.Logger
}

type TaskPublisher interface {
	PublishTask(taskID string, msg json.RawMessage) error
	PublishCancel(taskID string) error
}

type CreateScanRequest struct {
	Name            string          `json:"name" validate:"required"`
	Targets         []string        `json:"targets" validate:"required,min=1"`
	TemplateFilters json.RawMessage `json:"template_filters,omitempty"`
	Concurrency     json.RawMessage `json:"concurrency,omitempty"`
	RateLimit       int             `json:"rate_limit,omitempty"`
	Headers         []string        `json:"headers,omitempty"`
}

func NewScanHandler(db *gorm.DB, pub TaskPublisher, log *zerolog.Logger) *ScanHandler {
	return &ScanHandler{db: db, pub: pub, log: log}
}

func (h *ScanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Targets) == 0 {
		Error(w, http.StatusBadRequest, "targets is required")
		return
	}

	task := model.Task{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Status:          model.TaskStatusCreated,
		RateLimit:       req.RateLimit,
		TemplateFilters: stringOrDefault(req.TemplateFilters),
		Concurrency:     stringOrDefault(req.Concurrency),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if b, err := json.Marshal(req.Targets); err == nil {
		task.Targets = string(b)
	}
	if len(req.Headers) > 0 {
		if b, err := json.Marshal(req.Headers); err == nil {
			task.Headers = string(b)
		}
	}

	if err := h.db.Create(&task).Error; err != nil {
		h.log.Error().Err(err).Msg("failed to create task in db")
		Error(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	// Build the NATS message with all scan parameters
	natsMsg := buildTaskMessage(task)
	if err := h.pub.PublishTask(task.ID, natsMsg); err != nil {
		h.log.Error().Err(err).Str("task_id", task.ID).Msg("failed to publish task to NATS")
		task.Status = model.TaskStatusFailed
		task.ErrorMessage = "failed to queue: " + err.Error()
		h.db.Save(&task)
		Error(w, http.StatusInternalServerError, "failed to queue task")
		return
	}

	task.Status = model.TaskStatusQueued
	h.db.Model(&task).Update("status", model.TaskStatusQueued)

	Created(w, task)
}

func (h *ScanHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	status := r.URL.Query().Get("status")

	query := h.db.Model(&model.Task{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var tasks []model.Task
	query.Order("created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&tasks)

	Paginated(w, tasks, total, page, perPage)
}

func (h *ScanHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var task model.Task
	if err := h.db.First(&task, "id = ?", id).Error; err != nil {
		Error(w, http.StatusNotFound, "task not found")
		return
	}
	Success(w, task)
}

func (h *ScanHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var task model.Task
	if err := h.db.First(&task, "id = ?", id).Error; err != nil {
		Error(w, http.StatusNotFound, "task not found")
		return
	}
	if task.Status != model.TaskStatusQueued && task.Status != model.TaskStatusRunning {
		Error(w, http.StatusBadRequest, "task cannot be cancelled in current status: "+string(task.Status))
		return
	}

	_ = h.pub.PublishCancel(id)

	now := time.Now()
	h.db.Model(&task).Updates(map[string]interface{}{
		"status":       model.TaskStatusCancelled,
		"completed_at": now,
	})
	task.Status = model.TaskStatusCancelled
	task.CompletedAt = &now
	Success(w, task)
}

func (h *ScanHandler) Retry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var task model.Task
	if err := h.db.First(&task, "id = ?", id).Error; err != nil {
		Error(w, http.StatusNotFound, "task not found")
		return
	}
	if task.Status != model.TaskStatusFailed && task.Status != model.TaskStatusCancelled {
		Error(w, http.StatusBadRequest, "only failed or cancelled tasks can be retried")
		return
	}

	natsMsg := buildTaskMessage(task)
	if err := h.pub.PublishTask(task.ID, natsMsg); err != nil {
		Error(w, http.StatusInternalServerError, "failed to requeue task")
		return
	}

	h.db.Model(&task).Updates(map[string]interface{}{
		"status":          model.TaskStatusQueued,
		"assigned_worker_id": "",
		"error_message":   "",
		"started_at":      nil,
		"completed_at":    nil,
	})
	task.Status = model.TaskStatusQueued
	Success(w, task)
}

func buildTaskMessage(task model.Task) json.RawMessage {
	msg := map[string]interface{}{
		"task_id":   task.ID,
		"targets":   json.RawMessage(task.Targets),
		"created_at": task.CreatedAt,
	}
	if task.TemplateFilters != "" {
		msg["template_filters"] = json.RawMessage(task.TemplateFilters)
	}
	if task.Concurrency != "" {
		msg["concurrency"] = json.RawMessage(task.Concurrency)
	}
	if task.RateLimit > 0 {
		msg["rate_limit"] = task.RateLimit
	}
	if task.Headers != "" {
		msg["headers"] = json.RawMessage(task.Headers)
	}
	b, _ := json.Marshal(msg)
	return b
}

func stringOrDefault(b json.RawMessage) string {
	if len(b) == 0 {
		return ""
	}
	return string(b)
}
