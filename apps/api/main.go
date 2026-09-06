package main

import (
	"auth/internal/handler"
	"auth/internal/httpx"
	"auth/internal/infra"
	"auth/internal/repository"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
)

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, ""))
	if err != nil {
		return fallback
	}
	return value
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// authn
	httpx.AccessTokenSecret = env("ACCESS_TOKEN_SECRET", "secret")

	// infra
	db := infra.NewSQLDB(
		env("DB_HOST", "localhost"),
		envInt("DB_PORT", 4001),
		env("DB_USER", "forum"),
		env("DB_PASSWORD", ""),
		env("DB_NAME", "forum"),
	)
	defer db.Close()

	// repository
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)
	postRepo := repository.NewPostRepository(db, tagRepo)
	commentRepo := repository.NewCommentRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)

	// handler
	categoryHandler := handler.NewCategoryHandler(categoryRepo, tagRepo)
	postHandler := handler.NewPostHandler(postRepo, favoriteRepo, commentRepo)
	commentHandler := handler.NewCommentHandler(commentRepo)
	externalCommentHandler := handler.NewExternalCommentHandler(commentRepo)
	meHandler := handler.NewMeHandler(postRepo)

	// router
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK\n"))
	})
	router.Route("/v1", func(router chi.Router) {
		router.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{
			Level:         slog.LevelInfo,
			Schema:        httplog.SchemaECS,
			RecoverPanics: true,
		}))
		router.Route("/category", categoryHandler.RegisterRoutes)
		router.Route("/post", postHandler.RegisterRoutes)
		router.Route("/comment", commentHandler.RegisterRoutes)
		router.Route("/external/comment", externalCommentHandler.RegisterRoutes)
		router.Route("/me", meHandler.RegisterRoutes)
		router.Route("/admin", func(router chi.Router) {
			router.Use(httpx.RequireAdmin)
			router.Route("/category", categoryHandler.RegisterAdminRoutes)
			router.Route("/post", postHandler.RegisterAdminRoutes)
			router.Route("/comment", commentHandler.RegisterAdminRoutes)
		})
	})

	// start server
	slog.Info("Listening on localhost:8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		slog.Error("HTTP server failed", "error", err)
	}
}
