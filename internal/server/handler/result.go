package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/projectdiscovery/nuclei-platform/internal/model"
	"gorm.io/gorm"
)

type ResultHandler struct {
	db *gorm.DB
}

func NewResultHandler(db *gorm.DB) *ResultHandler {
	return &ResultHandler{db: db}
}

func (h *ResultHandler) ListByTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	page, perPage := pagination(r)
	severity := r.URL.Query().Get("severity")

	query := h.db.Model(&model.Result{}).Where("task_id = ?", taskID)
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	var total int64
	query.Count(&total)

	var results []model.Result
	query.Order("created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&results)

	Paginated(w, results, total, page, perPage)
}

func (h *ResultHandler) ListGlobal(w http.ResponseWriter, r *http.Request) {
	page, perPage := pagination(r)
	severity := r.URL.Query().Get("severity")
	host := r.URL.Query().Get("host")
	templateID := r.URL.Query().Get("template_id")

	query := h.db.Model(&model.Result{})
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if host != "" {
		query = query.Where("host LIKE ?", "%"+host+"%")
	}
	if templateID != "" {
		query = query.Where("template_id = ?", templateID)
	}

	var total int64
	query.Count(&total)

	var results []model.Result
	query.Order("created_at DESC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&results)

	Paginated(w, results, total, page, perPage)
}

type SeverityStats struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}

type HostStats struct {
	Host  string `json:"host"`
	Count int64  `json:"count"`
}

func (h *ResultHandler) Stats(w http.ResponseWriter, r *http.Request) {
	var bySeverity []SeverityStats
	h.db.Model(&model.Result{}).
		Select("severity, count(*) as count").
		Group("severity").
		Order("count DESC").
		Find(&bySeverity)

	var byHost []HostStats
	h.db.Model(&model.Result{}).
		Select("host, count(*) as count").
		Group("host").
		Order("count DESC").
		Limit(20).
		Find(&byHost)

	Success(w, map[string]interface{}{
		"by_severity": bySeverity,
		"by_host":     byHost,
	})
}

func pagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return page, perPage
}
