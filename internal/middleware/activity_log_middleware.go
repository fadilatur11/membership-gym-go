package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxActivityBodySize = 64 * 1024

var sensitiveActivityFields = map[string]struct{}{
	"access_token":   {},
	"authorization":  {},
	"client_secret":  {},
	"code":           {},
	"google_sub":     {},
	"owner_password": {},
	"password":       {},
	"password_hash":  {},
	"qr_token":       {},
	"refresh_token":  {},
	"secret":         {},
	"state":          {},
	"token":          {},
}

type activityResponseWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

type ActivityLogRepository interface {
	Create(
		ctx context.Context,
		gymID any,
		userID any,
		requestPayload any,
		response any,
		curl string,
		statusCode int,
		status string,
		responseTime int64,
	) error
}

func (w *activityResponseWriter) Write(data []byte) (int, error) {
	w.capture(data)
	return w.ResponseWriter.Write(data)
}

func (w *activityResponseWriter) WriteString(data string) (int, error) {
	w.capture([]byte(data))
	return w.ResponseWriter.WriteString(data)
}

func (w *activityResponseWriter) capture(data []byte) {
	remaining := maxActivityBodySize - w.body.Len()
	if remaining <= 0 {
		return
	}
	if len(data) > remaining {
		data = data[:remaining]
	}
	_, _ = w.body.Write(data)
}

func ActivityLogMiddleware(repository ActivityLogRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			ctx.Next()
			return
		}

		startedAt := time.Now()
		requestBody := readAndRestoreRequestBody(ctx.Request)
		writer := &activityResponseWriter{ResponseWriter: ctx.Writer}
		ctx.Writer = writer

		ctx.Next()

		statusCode := ctx.Writer.Status()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		status := "failed"
		if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
			status = "success"
		}

		auth := GetAuthUser(ctx)
		var gymID any
		var userID any
		if auth.GymID > 0 {
			gymID = auth.GymID
		}
		if auth.UserID > 0 {
			userID = auth.UserID
		}

		requestJSON := sanitizeBody(requestBody)
		responseJSON := sanitizeBody(writer.body.Bytes())
		curlCommand := buildCurl(ctx.Request, requestJSON)
		responseTime := time.Since(startedAt).Milliseconds()

		logContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := repository.Create(
			logContext,
			gymID,
			userID,
			nullableJSON(requestJSON),
			nullableJSON(responseJSON),
			curlCommand,
			statusCode,
			status,
			responseTime,
		); err != nil {
			log.Printf("activity log insert failed: %v", err)
		}
	}
}

func readAndRestoreRequestBody(request *http.Request) []byte {
	if request.Body == nil {
		return nil
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil
	}
	_ = request.Body.Close()
	request.Body = io.NopCloser(bytes.NewReader(body))
	return body
}

func sanitizeBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if len(body) > maxActivityBodySize {
		body = body[:maxActivityBodySize]
	}

	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		raw := string(body)
		encoded, _ := json.Marshal(map[string]any{"raw": raw, "truncated": len(body) == maxActivityBodySize})
		return string(encoded)
	}

	sanitizeActivityValue(value)
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func sanitizeActivityValue(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			if isSensitiveActivityField(key) {
				typed[key] = "[REDACTED]"
				continue
			}
			sanitizeActivityValue(item)
		}
	case []any:
		for _, item := range typed {
			sanitizeActivityValue(item)
		}
	}
}

func isSensitiveActivityField(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if _, found := sensitiveActivityFields[normalized]; found {
		return true
	}
	return strings.Contains(normalized, "password") ||
		strings.Contains(normalized, "secret") ||
		strings.HasSuffix(normalized, "_token")
}

func nullableJSON(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func buildCurl(request *http.Request, sanitizedBody string) string {
	parts := []string{"curl", "-X", shellQuote(request.Method), shellQuote(sanitizedRequestURL(request))}

	headerNames := make([]string, 0, len(request.Header))
	for name := range request.Header {
		if strings.EqualFold(name, "Authorization") ||
			strings.EqualFold(name, "Content-Type") ||
			strings.EqualFold(name, "Accept") {
			headerNames = append(headerNames, name)
		}
	}
	sort.Strings(headerNames)
	for _, name := range headerNames {
		value := request.Header.Get(name)
		if strings.EqualFold(name, "Authorization") {
			value = "Bearer [REDACTED]"
		}
		parts = append(parts, "-H", shellQuote(fmt.Sprintf("%s: %s", name, value)))
	}
	if sanitizedBody != "" {
		parts = append(parts, "--data-raw", shellQuote(sanitizedBody))
	}
	return strings.Join(parts, " ")
}

func sanitizedRequestURL(request *http.Request) string {
	copyURL := *request.URL
	query := copyURL.Query()
	for key := range query {
		if isSensitiveActivityField(key) {
			query.Set(key, "[REDACTED]")
		}
	}
	copyURL.RawQuery = query.Encode()
	if copyURL.IsAbs() {
		return copyURL.String()
	}

	scheme := request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
		if request.TLS != nil {
			scheme = "https"
		}
	}
	host := request.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = request.Host
	}
	if host == "" {
		return copyURL.RequestURI()
	}
	copyURL.Scheme = scheme
	copyURL.Host = host
	return copyURL.String()
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
