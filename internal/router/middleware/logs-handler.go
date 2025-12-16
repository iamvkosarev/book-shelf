package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

func LogProcessesEdges(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			requestID := strconv.FormatInt(time.Now().UnixMilli(), 10)
			path := r.URL.Path[1:]
			slogAttrID := slog.String("request_id", requestID)
			startTime := time.Now()
			slog.Debug(fmt.Sprintf("start request: %s %s", r.Method, path), slogAttrID)
			next.ServeHTTP(w, r)
			slogAttrDuration := slog.String("duration", time.Since(startTime).String())
			slog.Debug(fmt.Sprintf("finish request: %s %s", r.Method, path), slogAttrID, slogAttrDuration)
		},
	)
}
