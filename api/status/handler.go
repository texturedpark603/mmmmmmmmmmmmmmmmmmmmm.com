package status

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var startTime = time.Now()

type statusResp struct {
	Up      bool   `json:"up"`
	Uptime  string `json:"uptime"`  // human-readable
	UptimeS int64  `json:"uptime_s"` // seconds
	Now     string `json:"now"`      // RFC3339 UTC
}

func humanDuration(d time.Duration) string {
	total := int(d.Seconds())
	days := total / 86400
	hours := (total % 86400) / 3600
	mins := (total % 3600) / 60
	secs := total % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)
	out := statusResp{
		Up:      true,
		Uptime:  humanDuration(uptime),
		UptimeS: int64(uptime.Seconds()),
		Now:     time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(out)
}
