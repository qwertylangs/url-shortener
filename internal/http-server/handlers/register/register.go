package register

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	ssoGrpc "url-shortener/internal/clients/sso/grpc"
	"url-shortener/internal/config"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func New(log *slog.Logger, ssoClient *ssoGrpc.Client, cfg *config.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.register.New"
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty")
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("empty request"))
				return
			}
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			render.Status(r, http.StatusBadRequest) 
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		_, err = ssoClient.Register(r.Context(), req.Email, req.Password, cfg.Clients.SSO.AppId)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.AlreadyExists {
				log.Error("user already exists")
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, resp.Error("user already exists"))
				return
			}
			log.Error("failed to register", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to register"))
			return
		}

		token, err := ssoClient.Login(r.Context(), req.Email, req.Password, cfg.Clients.SSO.AppId)
		if err != nil {
			log.Error("failed to login", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to login"))
			return
		}

		// Set token in cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			Path:     "/",
			MaxAge:   3600 * 24, // 24 hours
			HttpOnly: true,
			Secure:   false, // TODO: Set to true in production with HTTPS
			SameSite: http.SameSiteStrictMode,
		})

		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}