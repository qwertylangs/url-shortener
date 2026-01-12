package health

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// TODO проверять SSO
func New(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.health.New"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		log.Info("health check")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}