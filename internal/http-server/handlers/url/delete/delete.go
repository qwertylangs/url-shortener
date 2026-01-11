package delete

import (
	"context"
	"errors"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	ssoGrpc "url-shortener/internal/clients/sso/grpc"
	"url-shortener/internal/http-server/middleware/auth"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"
)

type Response struct {
	resp.Response
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLDeleter
type URLDeleter interface {
	DeleteURL(ctx context.Context, alias string, userID int64, isAdmin bool) error
}

func New(log *slog.Logger, urlDeleter URLDeleter, ssoClient *ssoGrpc.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		userID, ok := r.Context().Value(auth.UserIDContextKey).(int64)
		if !ok {
			log.Error("user_id not found in context")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("user_id not found in token"))
			return
		}

		isAdmin, err := ssoClient.IsAdmin(r.Context(), userID)
		if err != nil {
			log.Error("failed to check if user is admin", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to check if user is admin"))
			return
		}

		err = urlDeleter.DeleteURL(r.Context(), alias, userID, isAdmin)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Error("url not found", sl.Err(err))
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("url not found"))
				return
			}
			if errors.Is(err, storage.ErrURLNotOwned) {
				log.Error("url not owned", sl.Err(err))
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, resp.Error("url not owned"))
				return
			}
			log.Error("failed to delete url", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to delete url"))
			return
		}

		render.JSON(w, r, Response{
			Response: resp.OK(),
		})
	}
}
