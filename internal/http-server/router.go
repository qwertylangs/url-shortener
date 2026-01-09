package httpserver

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"

	ssoGrpc "url-shortener/internal/clients/sso/grpc"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/login"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/register"
	"url-shortener/internal/http-server/handlers/url/getUrls"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/middleware/auth"
	mwLogger "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/storage"
)

// NewRouter creates and configures a new chi router with all routes and middleware
func NewRouter(
	log *slog.Logger,
	urlStorage storage.Storage,
	ssoClient *ssoGrpc.Client,
	cfg *config.AppConfig,
) *chi.Mux {
	router := chi.NewRouter()

	c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"}, 
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
        AllowCredentials: true,
        MaxAge:           300, 
    })

    router.Use(c.Handler)

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	
	// Auth routes
	router.Post("/login", login.New(log, ssoClient, cfg))
	router.Post("/register", register.New(log, ssoClient, cfg))
	router.Get("/login", login.GetLogin(log, cfg))

	// Protected routes
	router.Route("/url", func(r chi.Router) {
		authMiddleware := auth.New(log, cfg)
		r.Use(authMiddleware)
		r.Post("/", save.New(log, urlStorage))
		r.Get("/", getUrls.New(log, urlStorage))
		// TODO: add DELETE /url/{id}
	})

	// Public routes
	router.Get("/{alias}", redirect.New(log, urlStorage))

	return router
}

