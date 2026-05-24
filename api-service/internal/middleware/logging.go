package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const loggerKey = contextKey("logger")

func LoggingMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get Request ID from chi middleware (if it exists)
			reqID := middleware.GetReqID(r.Context())

			// Inject request-scoped logger into context
			reqLogger := logger.With(zap.String("request_id", reqID))
			ctx := context.WithValue(r.Context(), loggerKey, reqLogger)
			r = r.WithContext(ctx)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			reqLogger.Info("handled request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

// GetLogger retrieves the logger from the context, or returns the global no-op logger if not found
func GetLogger(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return zap.NewNop()
}
