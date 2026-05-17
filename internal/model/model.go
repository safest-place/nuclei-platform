package model

import "time"

type TaskStatus string

const (
	TaskStatusCreated   TaskStatus = "created"
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type Task struct {
	ID               string     `gorm:"primaryKey;type:text" json:"id"`
	Name             string     `json:"name"`
	Targets          string     `gorm:"type:text" json:"targets,omitempty"`
	TemplateFilters  string     `gorm:"type:text" json:"template_filters,omitempty"`
	Concurrency      string     `gorm:"type:text" json:"concurrency,omitempty"`
	RateLimit        int        `json:"rate_limit,omitempty"`
	Headers          string     `gorm:"type:text" json:"headers,omitempty"`
	Status           TaskStatus `gorm:"type:text;index" json:"status"`
	AssignedWorkerID string     `gorm:"type:text" json:"assigned_worker_id,omitempty"`
	ErrorMessage     string     `gorm:"type:text" json:"error_message,omitempty"`
	ResultCount      int        `json:"result_count"`
	CreatedAt        time.Time  `json:"created_at"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Result struct {
	ID               string    `gorm:"primaryKey;type:text" json:"id"`
	TaskID           string    `gorm:"type:text;index" json:"task_id"`
	TemplateID       string    `json:"template_id"`
	TemplateName     string    `json:"template_name"`
	Host             string    `json:"host"`
	MatchedAt        string    `json:"matched_at"`
	Severity         string    `gorm:"type:text;index" json:"severity"`
	IP               string    `json:"ip"`
	Port             string    `json:"port"`
	Scheme           string    `json:"scheme"`
	URL              string    `json:"url"`
	Request          string    `gorm:"type:text" json:"request,omitempty"`
	Response         string    `gorm:"type:text" json:"response,omitempty"`
	CurlCommand      string    `json:"curl_command,omitempty"`
	ExtractedResults string    `gorm:"type:text" json:"extracted_results,omitempty"`
	MatcherName      string    `json:"matcher_name,omitempty"`
	Type             string    `json:"type"`
	CVEID            string    `json:"cve_id,omitempty"`
	CVSSScore        float64   `json:"cvss_score,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
	CreatedAt        time.Time `json:"created_at"`
}

type Worker struct {
	ID            string    `gorm:"primaryKey;type:text" json:"id"`
	Hostname      string    `json:"hostname"`
	IP            string    `json:"ip"`
	Status        string    `gorm:"type:text;index" json:"status"`
	Capabilities  string    `gorm:"type:text" json:"capabilities,omitempty"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Disabled      bool      `gorm:"default:false" json:"disabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
