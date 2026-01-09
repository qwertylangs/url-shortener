package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"url-shortener/internal/config"
	"url-shortener/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDContextKey = "user_id"
const UserEmailContextKey = "user_email"


func New(log *slog.Logger, cfg *config.AppConfig) func(next http.Handler) http.Handler { 
	return func(next http.Handler) http.Handler {
		op := "middleware.auth.New"
		
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			
			tokenParsed, err := jwt.Parse(cookie.Value, func(cookie *jwt.Token) (interface{}, error) {
				return []byte(cfg.AppSecret), nil
			})
			if err != nil {
				log.Error("failed to parse auth token", sl.Err(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			tokenClaims, _ := tokenParsed.Claims.(jwt.MapClaims)

			var userId string
			switch v := tokenClaims["uid"].(type) {
			case float64:
				userId = fmt.Sprintf("%.0f", v) // Преобразуем float64 в строку без десятичной части
			case string:
				userId = v
			default:
				log.Error("unexpected uid type", slog.Any("type", v))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			userEmail := tokenClaims["email"].(string)
			log.Info("token claims", slog.String("user_id", userId), slog.String("user_email", userEmail))

			r = r.WithContext(context.WithValue(r.Context(), UserIDContextKey, userId))
			r = r.WithContext(context.WithValue(r.Context(), UserEmailContextKey, userEmail))
			next.ServeHTTP(w, r)
		})
	}
}