package worker

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"github.com/projectdiscovery/nuclei-platform/internal/config"
	"github.com/rs/zerolog"
)

type Worker struct {
	id          string
	scanner     *Scanner
	nc          *nats.Conn
	cfg         *config.WorkerAppConfig
	log         *zerolog.Logger
	cancelFuncs sync.Map // task_id -> context.CancelFunc
}

type taskMessage struct {
	TaskID          string          `json:"task_id"`
	Targets         []string        `json:"targets"`
	TemplateFilters json.RawMessage `json:"template_filters,omitempty"`
	Concurrency     json.RawMessage `json:"concurrency,omitempty"`
	RateLimit       int             `json:"rate_limit,omitempty"`
	Headers         []string        `json:"headers,omitempty"`
}

type resultMessage struct {
	TaskID           string  `json:"task_id"`
	WorkerID         string  `json:"worker_id"`
	Status           string  `json:"status,omitempty"`
	ResultCount      int     `json:"result_count,omitempty"`
	ErrorMessage     string  `json:"error_message,omitempty"`
	CompletedAt      string  `json:"completed_at,omitempty"`
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

func New(cfg *config.WorkerAppConfig, log *zerolog.Logger) (*Worker, error) {
	workerID := cfg.WorkerID
	if workerID == "" {
		workerID = loadOrCreateWorkerID("/data/worker-id")
	}

	// Connect to NATS
	nc, err := nats.Connect(cfg.NATS.URL,
		nats.Name("nuclei-worker-"+workerID[:8]),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, err
	}

	// Build nuclei options
	nucleiOpts := []nuclei.NucleiSDKOptions{}
	if cfg.Nuclei.TemplatesDir != "" {
		nucleiOpts = append(nucleiOpts, nuclei.WithTemplatesOrWorkflows(nuclei.TemplateSources{
			Templates: []string{cfg.Nuclei.TemplatesDir},
		}))
	}
	if len(cfg.Nuclei.ExcludeTags) > 0 {
		nucleiOpts = append(nucleiOpts, nuclei.WithTemplateFilters(nuclei.TemplateFilters{
			ExcludeTags: cfg.Nuclei.ExcludeTags,
		}))
	}

	scanner, err := NewScanner(context.Background(), log, nucleiOpts...)
	if err != nil {
		nc.Close()
		return nil, err
	}

	return &Worker{
		id:      workerID,
		scanner: scanner,
		nc:      nc,
		cfg:     cfg,
		log:     log,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	w.log.Info().Str("worker_id", w.id).Msg("worker starting")

	// Start heartbeat
	StartHeartbeat(ctx, w.nc, w.cfg.NATS.HeartbeatSubject, w.id, w.cfg.Heartbeat, w.log)

	// Subscribe to cancel messages
	_, err := w.nc.Subscribe(w.cfg.NATS.CancelSubject, func(msg *nats.Msg) {
		var cancelMsg struct {
			TaskID string `json:"task_id"`
		}
		if err := json.Unmarshal(msg.Data, &cancelMsg); err != nil {
			return
		}
		if cancel, ok := w.cancelFuncs.Load(cancelMsg.TaskID); ok {
			w.log.Info().Str("task_id", cancelMsg.TaskID).Msg("cancelling task")
			cancel.(context.CancelFunc)()
		}
	})
	if err != nil {
		return err
	}

	// Subscribe to task messages with queue group
	_, err = w.nc.QueueSubscribe(w.cfg.NATS.TaskSubject, "nuclei-workers", func(msg *nats.Msg) {
		w.handleTask(ctx, msg.Data)
	})
	if err != nil {
		return err
	}

	w.log.Info().Msg("worker ready, waiting for tasks")
	<-ctx.Done()
	w.log.Info().Msg("worker shutting down")
	return nil
}

func (w *Worker) handleTask(parentCtx context.Context, data []byte) {
	var task taskMessage
	if err := json.Unmarshal(data, &task); err != nil {
		w.log.Error().Err(err).Msg("failed to unmarshal task")
		return
	}

	w.log.Info().Str("task_id", task.TaskID).Strs("targets", task.Targets).Msg("received task")

	taskCtx, cancel := context.WithCancel(parentCtx)
	w.cancelFuncs.Store(task.TaskID, cancel)
	defer func() {
		cancel()
		w.cancelFuncs.Delete(task.TaskID)
	}()

	// Set up result callback for this task
	resultCount := 0
	w.scanner.SetResultCallback(func(event *output.ResultEvent) {
		resultCount++
		rm := resultMessage{
			TaskID:           task.TaskID,
			WorkerID:         w.id,
			TemplateID:       event.TemplateID,
			TemplateName:     event.Info.Name,
			Host:             event.Host,
			MatchedAt:        event.Matched,
			Severity:         string(event.Info.SeverityHolder.Severity),
			IP:               event.IP,
			Port:             event.Port,
			Scheme:           event.Scheme,
			URL:              event.URL,
			Request:          event.Request,
			Response:         event.Response,
			CurlCommand:      event.CURLCommand,
			MatcherName:      event.MatcherName,
			Type:             string(event.Type),
			Timestamp:        event.Timestamp.Format(time.RFC3339),
		}
		if len(event.ExtractedResults) > 0 {
			b, _ := json.Marshal(event.ExtractedResults)
			rm.ExtractedResults = string(b)
		}
		if event.Info.Classification != nil {
			rm.CVEID = event.Info.Classification.CVEID.String()
			rm.CVSSScore = event.Info.Classification.CVSSScore
		}
		resultData, _ := json.Marshal(rm)
		if err := w.nc.Publish(w.cfg.NATS.ResultSubject, resultData); err != nil {
			w.log.Error().Err(err).Str("task_id", task.TaskID).Msg("failed to publish result")
		}
	})

	// Build nuclei execution options
	var execOpts []nuclei.NucleiSDKOptions
	if len(task.TemplateFilters) > 0 {
		var tf nuclei.TemplateFilters
		if err := json.Unmarshal(task.TemplateFilters, &tf); err == nil {
			execOpts = append(execOpts, nuclei.WithTemplateFilters(tf))
		}
	}
	if task.RateLimit > 0 {
		execOpts = append(execOpts, nuclei.WithGlobalRateLimitCtx(taskCtx, task.RateLimit, time.Second))
	}
	if len(task.Headers) > 0 {
		execOpts = append(execOpts, nuclei.WithHeaders(task.Headers))
	}

	// Execute scan
	err := w.scanner.Execute(taskCtx, task.Targets, execOpts...)

	// Send completion signal
	completion := resultMessage{
		TaskID:      task.TaskID,
		WorkerID:    w.id,
		ResultCount: resultCount,
		CompletedAt: time.Now().Format(time.RFC3339),
	}
	if err != nil {
		completion.Status = "failed"
		completion.ErrorMessage = err.Error()
	} else {
		completion.Status = "completed"
	}
	completionData, _ := json.Marshal(completion)
	if pubErr := w.nc.Publish(w.cfg.NATS.ResultSubject, completionData); pubErr != nil {
		w.log.Error().Err(pubErr).Msg("failed to publish completion")
	}

	w.log.Info().Str("task_id", task.TaskID).Int("results", resultCount).Err(err).Msg("task finished")
}

func (w *Worker) Close() {
	w.scanner.Close()
	w.nc.Drain()
}

func loadOrCreateWorkerID(path string) string {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)

	data, err := os.ReadFile(path)
	if err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id
		}
	}
	id := uuid.New().String()
	_ = os.WriteFile(path, []byte(id), 0644)
	return id
}
