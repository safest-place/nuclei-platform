package worker

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
)

type heartbeatPayload struct {
	WorkerID      string `json:"worker_id"`
	Hostname      string `json:"hostname"`
	IP            string `json:"ip"`
	Status        string `json:"status"`
	CurrentTaskID string `json:"current_task_id,omitempty"`
	Capabilities  string `json:"capabilities,omitempty"`
}

func StartHeartbeat(ctx context.Context, nc *nats.Conn, subject, workerID string, interval time.Duration, log *zerolog.Logger) {
	hostname, _ := os.Hostname()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				payload := heartbeatPayload{
					WorkerID: workerID,
					Hostname: hostname,
					IP:       getLocalIP(),
					Status:   "online",
				}
				data, _ := json.Marshal(payload)
				if err := nc.Publish(subject, data); err != nil {
					log.Warn().Err(err).Msg("failed to send heartbeat")
				}
			}
		}
	}()
}

func getLocalIP() string {
	addrs, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return addrs
}
