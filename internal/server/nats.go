package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/projectdiscovery/nuclei-platform/internal/config"
	"github.com/projectdiscovery/nuclei-platform/internal/model"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type NATSManager struct {
	nc   *nats.Conn
	cfg  config.NATSConfig
	db   *gorm.DB
	log  *zerolog.Logger
}

func NewNATSManager(cfg config.NATSConfig, db *gorm.DB, log *zerolog.Logger) (*NATSManager, error) {
	nc, err := nats.Connect(cfg.URL,
		nats.Name("nuclei-platform-api"),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, err
	}
	log.Info().Str("url", cfg.URL).Msg("connected to NATS")
	return &NATSManager{nc: nc, cfg: cfg, db: db, log: log}, nil
}

func (m *NATSManager) PublishTask(taskID string, msg json.RawMessage) error {
	return m.nc.Publish(m.cfg.TaskSubject, msg)
}

func (m *NATSManager) PublishCancel(taskID string) error {
	msg, _ := json.Marshal(map[string]string{"task_id": taskID})
	return m.nc.Publish(m.cfg.CancelSubject, msg)
}

type resultMessage struct {
	TaskID        string  `json:"task_id"`
	WorkerID      string  `json:"worker_id"`
	Status        string  `json:"status,omitempty"`
	ResultCount   int     `json:"result_count,omitempty"`
	ErrorMessage  string  `json:"error_message,omitempty"`
	CompletedAt   string  `json:"completed_at,omitempty"`
	// Result fields
	TemplateID       string  `json:"template_id,omitempty"`
	TemplateName     string  `json:"template_name,omitempty"`
	Host             string  `json:"host,omitempty"`
	MatchedAt        string  `json:"matched_at,omitempty"`
	Severity         string  `json:"severity,omitempty"`
	IP               string  `json:"ip,omitempty"`
	Port             string  `json:"port,omitempty"`
	Scheme           string  `json:"scheme,omitempty"`
	URL              string  `json:"url,omitempty"`
	Request          string  `json:"request,omitempty"`
	Response         string  `json:"response,omitempty"`
	CurlCommand      string  `json:"curl_command,omitempty"`
	ExtractedResults string  `json:"extracted_results,omitempty"`
	MatcherName      string  `json:"matcher_name,omitempty"`
	Type             string  `json:"type,omitempty"`
	CVEID            string  `json:"cve_id,omitempty"`
	CVSSScore        float64 `json:"cvss_score,omitempty"`
	Timestamp        string  `json:"timestamp,omitempty"`
}

type heartbeatMessage struct {
	WorkerID      string `json:"worker_id"`
	Hostname      string `json:"hostname"`
	IP            string `json:"ip"`
	Status        string `json:"status"`
	CurrentTaskID string `json:"current_task_id,omitempty"`
	Capabilities  string `json:"capabilities,omitempty"`
}

func (m *NATSManager) SubscribeResults(ctx context.Context) error {
	_, err := m.nc.QueueSubscribe(m.cfg.ResultSubject, "nuclei-api", func(msg *nats.Msg) {
		var rm resultMessage
		if err := json.Unmarshal(msg.Data, &rm); err != nil {
			m.log.Error().Err(err).Msg("failed to unmarshal result message")
			return
		}

		// Completion signal
		if rm.Status != "" {
			now := time.Now()
			updates := map[string]interface{}{
				"status":        model.TaskStatus(rm.Status),
				"result_count":  rm.ResultCount,
				"completed_at":  now,
				"assigned_worker_id": rm.WorkerID,
			}
			if rm.ErrorMessage != "" {
				updates["error_message"] = rm.ErrorMessage
			}
			m.db.Model(&model.Task{}).Where("id = ?", rm.TaskID).Updates(updates)
			m.log.Info().Str("task_id", rm.TaskID).Str("status", rm.Status).Int("results", rm.ResultCount).Msg("task completed")
			return
		}

		// Individual result
		var ts time.Time
		if rm.Timestamp != "" {
			ts, _ = time.Parse(time.RFC3339, rm.Timestamp)
		}
		if ts.IsZero() {
			ts = time.Now()
		}

		result := model.Result{
			ID:               generateID(),
			TaskID:           rm.TaskID,
			TemplateID:       rm.TemplateID,
			TemplateName:     rm.TemplateName,
			Host:             rm.Host,
			MatchedAt:        rm.MatchedAt,
			Severity:         rm.Severity,
			IP:               rm.IP,
			Port:             rm.Port,
			Scheme:           rm.Scheme,
			URL:              rm.URL,
			Request:          truncate(rm.Request, 65536),
			Response:         truncate(rm.Response, 65536),
			CurlCommand:      rm.CurlCommand,
			ExtractedResults: rm.ExtractedResults,
			MatcherName:      rm.MatcherName,
			Type:             rm.Type,
			CVEID:            rm.CVEID,
			CVSSScore:        rm.CVSSScore,
			Timestamp:        ts,
			CreatedAt:        time.Now(),
		}

		if err := m.db.Create(&result).Error; err != nil {
			m.log.Error().Err(err).Str("task_id", rm.TaskID).Msg("failed to save result")
			return
		}

		m.db.Model(&model.Task{}).Where("id = ?", rm.TaskID).
			UpdateColumn("result_count", gorm.Expr("result_count + 1"))

		// Mark task as running if it's the first result
		m.db.Model(&model.Task{}).
			Where("id = ? AND status = ?", rm.TaskID, model.TaskStatusQueued).
			Updates(map[string]interface{}{
				"status":            model.TaskStatusRunning,
				"assigned_worker_id": rm.WorkerID,
			})
	})
	if err != nil {
		return err
	}
	m.log.Info().Str("subject", m.cfg.ResultSubject).Msg("subscribed to results")
	return nil
}

func (m *NATSManager) SubscribeHeartbeats(ctx context.Context) error {
	_, err := m.nc.Subscribe(m.cfg.HeartbeatSubject, func(msg *nats.Msg) {
		var hb heartbeatMessage
		if err := json.Unmarshal(msg.Data, &hb); err != nil {
			m.log.Error().Err(err).Msg("failed to unmarshal heartbeat")
			return
		}

		worker := model.Worker{
			ID:            hb.WorkerID,
			Hostname:      hb.Hostname,
			IP:            hb.IP,
			Status:        hb.Status,
			Capabilities:  hb.Capabilities,
			LastHeartbeat: time.Now(),
		}
		m.db.Where("id = ?", hb.WorkerID).Assign(map[string]interface{}{
			"hostname":       hb.Hostname,
			"ip":             hb.IP,
			"status":         hb.Status,
			"capabilities":   hb.Capabilities,
			"last_heartbeat": time.Now(),
		}).FirstOrCreate(&worker)
	})
	if err != nil {
		return err
	}
	m.log.Info().Str("subject", m.cfg.HeartbeatSubject).Msg("subscribed to heartbeats")
	return nil
}

func (m *NATSManager) StartStaleWorkerDetector(ctx context.Context, interval, threshold time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().Add(-threshold)
				result := m.db.Model(&model.Worker{}).
					Where("status != ? AND last_heartbeat < ?", "offline", cutoff).
					Update("status", "offline")
				if result.RowsAffected > 0 {
					m.log.Info().Int64("count", result.RowsAffected).Msg("marked stale workers offline")
				}
			}
		}
	}()
}

func (m *NATSManager) StartTaskTimeoutDetector(ctx context.Context, interval, maxDuration time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cutoff := time.Now().Add(-maxDuration)
				result := m.db.Model(&model.Task{}).
					Where("status = ? AND started_at < ?", model.TaskStatusRunning, cutoff).
					Updates(map[string]interface{}{
						"status":        model.TaskStatusFailed,
						"error_message": "task timed out",
						"completed_at":  time.Now(),
					})
				if result.RowsAffected > 0 {
					m.log.Warn().Int64("count", result.RowsAffected).Msg("timed out stuck tasks")
				}
			}
		}
	}()
}

func (m *NATSManager) Close() {
	if m.nc != nil {
		m.nc.Drain()
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}
