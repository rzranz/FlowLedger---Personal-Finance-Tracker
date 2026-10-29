package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"flowledger/internal/adapter/sqlite"
	"flowledger/internal/domain"
	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
)

// GetRequestID returns the request ID from the context if present.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// RequestID middleware injects or propagates an X-Request-ID header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// statusWriter wraps http.ResponseWriter to capture the HTTP status code.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(status int) {
	if !w.wroteHeader {
		w.status = status
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// Logger middleware logs request details using log/slog.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			duration := time.Since(start)
			reqID := GetRequestID(r.Context())

			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.Int64("duration_ms", duration.Milliseconds()),
				slog.String("request_id", reqID),
			)
		})
	}
}

// Recoverer catches panics, logs the stack trace, and returns an RFC 9457 500 error.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					reqID := GetRequestID(r.Context())
					logger.Error("panic recovered in http handler",
						slog.Any("error", rec),
						slog.String("stack", string(debug.Stack())),
						slog.String("request_id", reqID),
					)
					RespondError(w, r, domain.ErrInvalidInput)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORS middleware handles preflight OPTIONS requests and CORS headers.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, Content-Disposition")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// IdempotencyStore defines storage contract for idempotency caching.
type IdempotencyStore interface {
	Get(ctx context.Context, key string) (*sqlite.IdempotencyRecord, error)
	Save(ctx context.Context, record *sqlite.IdempotencyRecord) error
}

// responseRecorder captures response status and payload for idempotency storage.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// IdempotencyMiddleware ensures idempotent execution for write requests carrying an Idempotency-Key.
func IdempotencyMiddleware(store IdempotencyStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Read and hash the request body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				RespondError(w, r, domain.ErrInvalidInput)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			hash := sha256.Sum256(bodyBytes)
			hashHex := hex.EncodeToString(hash[:])

			// Check for existing record
			record, err := store.Get(r.Context(), key)
			if err == nil && record != nil {
				if record.RequestHash != hashHex {
					RespondError(w, r, domain.ErrIdempotencyConflict)
					return
				}

				// Replay cached response
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Header().Set("X-Cache-Lookup", "HIT")
				w.WriteHeader(record.StatusCode)
				_, _ = w.Write(record.ResponseBody)
				return
			}

			// Wrap writer to record the output
			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(rec, r)

			// Persist only successful responses (2xx)
			if rec.statusCode >= 200 && rec.statusCode < 300 {
				_ = store.Save(r.Context(), &sqlite.IdempotencyRecord{
					Key:          key,
					RequestHash:  hashHex,
					StatusCode:   rec.statusCode,
					ResponseBody: rec.body.Bytes(),
				})
			}
		})
	}
}

