package login

import (
	"log/slog"
	"net/http"
	"url-shortener/internal/config"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
)

type Response struct {
	Email string `json:"email"`
	UserId string `json:"user_id"`
}

func GetLogin(log *slog.Logger, cfg *config.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.GetLogin.New"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		cookie, err := r.Cookie("auth_token")
		if err != nil {
			log.Error("failed to get auth token from cookie", sl.Err(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		_, err = jwt.Parse(cookie.Value, func(cookie *jwt.Token) (interface{}, error) {
			return []byte(cfg.AppSecret), nil
		})
		if err != nil {
			log.Error("failed to parse auth token", sl.Err(err))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}