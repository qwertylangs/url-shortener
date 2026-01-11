package getUrls

import (
	"context"
	"errors"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"url-shortener/internal/http-server/middleware/auth"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/models"
	"url-shortener/internal/storage"
)


type Response struct {
	URLs []models.URL `json:"urls"`
}

type URLsGetter interface {
	GetUserURLs(ctx context.Context, userID int64) ([]models.URL, error)
}

func New(log *slog.Logger, urlsGetter URLsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.getUrls.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userID, ok := r.Context().Value(auth.UserIDContextKey).(int64)
		if !ok {
			log.Error("user_id not found in context")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("user_id not found in token"))
			return
		}


		urls, err := urlsGetter.GetUserURLs(r.Context(), userID)
		if err != nil {
			if errors.Is(err, storage.ErrUserURLsNotFound) {
				log.Info("user urls not found")
				render.Status(r, http.StatusNoContent)
				render.JSON(w, r, Response{
					URLs: []models.URL{},
				})
				return
			}
			log.Error("failed to get user urls", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get user urls"))
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			URLs: urls,
		})
	}
}
